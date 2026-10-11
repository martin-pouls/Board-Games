package gate

import (
	"crypto/sha256"
	"crypto/subtle"
	"html/template"
	"io/fs"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/martin-pouls/shared/backend/httpkit"
)

// The login page (see ui.go).
var loginTemplate, uiAssets = func() (*template.Template, fs.FS) {
	t, assets, err := loadUI(uiFiles)
	if err != nil {
		panic("gate: login page: " + err.Error())
	}
	return t, assets
}()

const (
	maxBodyBytes = 4 << 10
	csp          = "default-src 'none'; style-src 'self' 'unsafe-inline'; font-src 'self'; script-src 'unsafe-inline'; img-src 'self' data:; form-action 'self'; frame-ancestors 'none'; base-uri 'none'"

	msgWrongPassword = "Wrong password."
	msgTooMany       = "Too many attempts. Try again in a minute."
)

// LoginRate is the limit on password attempts per client.
var LoginRate = httpkit.Rate{Burst: 10, Every: 30 * time.Second}

// Handler serves the gate's routes.
type Handler struct {
	cfg     Config
	tokens  *Tokens
	logger  *slog.Logger
	limiter *httpkit.Limiter
	digests [][sha256.Size]byte // one per cfg.Passwords
	mux     *http.ServeMux
}

// NewHandler wires the routes. now is the clock (nil: time.Now).
func NewHandler(cfg Config, logger *slog.Logger, now func() time.Time) *Handler {
	if now == nil {
		now = time.Now
	}
	h := &Handler{
		cfg:     cfg,
		tokens:  NewTokens(cfg.Secret, cfg.Passwords, now),
		logger:  logger,
		limiter: httpkit.NewLimiter(LoginRate, now),
	}
	for _, p := range cfg.Passwords {
		h.digests = append(h.digests, sha256.Sum256([]byte(p.Password)))
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /verify", h.verify)
	mux.HandleFunc("GET /login", h.loginPage)
	mux.HandleFunc("POST /login", h.loginSubmit)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, h.cfg.BasePath+"/login", http.StatusFound)
	})
	// The login page's CSS, fonts and favicon.
	assets := serveAssets(uiAssets)
	mux.Handle("GET /_app/", assets)
	mux.Handle("GET /favicon.svg", assets)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})
	h.mux = mux
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.mux.ServeHTTP(w, r) }

func (h *Handler) validCookie(r *http.Request) bool {
	c, err := r.Cookie(CookieName)
	if err != nil {
		return false
	}
	_, ok := h.tokens.Verify(c.Value)
	return ok
}

func (h *Handler) verify(w http.ResponseWriter, r *http.Request) {
	if h.validCookie(r) {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if isPageLoad(r) {
		next := r.Header.Get("X-Forwarded-Uri")
		http.Redirect(w, r, h.cfg.BasePath+"/login?next="+url.QueryEscape(next), http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":{"code":"login_required","message":"login required"}}`))
}

// isPageLoad tells a browser navigation (worth a redirect to the login page) from fetches,
// WebSockets and other requests (which get a 401 the frontends understand).
func isPageLoad(r *http.Request) bool {
	method := r.Header.Get("X-Forwarded-Method")
	if method == "" {
		method = http.MethodGet
	}
	if method != http.MethodGet && method != http.MethodHead {
		return false
	}
	if mode := r.Header.Get("Sec-Fetch-Mode"); mode != "" {
		return mode == "navigate"
	}
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

type loginView struct {
	Action  string
	Next    string
	Message string
}

func (h *Handler) render(w http.ResponseWriter, status int, next, message string) {
	hd := w.Header()
	hd.Set("Content-Type", "text/html; charset=utf-8")
	hd.Set("Cache-Control", "no-store")
	hd.Set("X-Content-Type-Options", "nosniff")
	hd.Set("Referrer-Policy", "same-origin")
	hd.Set("X-Frame-Options", "DENY")
	hd.Set("Content-Security-Policy", csp)
	w.WriteHeader(status)
	if err := loginTemplate.Execute(w, loginView{Action: h.cfg.BasePath + "/login", Next: next, Message: message}); err != nil {
		h.logger.Error("render login page", "err", err)
	}
}

func (h *Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	next := h.sanitizeNext(r.URL.Query().Get("next"))
	if h.validCookie(r) {
		w.Header().Set("Cache-Control", "no-store")
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	h.render(w, http.StatusOK, next, "")
}

func (h *Handler) loginSubmit(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := r.ParseForm(); err != nil {
		h.render(w, http.StatusBadRequest, h.cfg.DefaultNext, "Bad request.")
		return
	}
	next := h.sanitizeNext(r.PostForm.Get("next"))
	if wait := h.limiter.Take(r); wait > 0 {
		h.logger.Warn("login rate limited")
		w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
		h.render(w, http.StatusTooManyRequests, next, msgTooMany)
		return
	}
	label, ok := h.match(r.PostForm.Get("password"))
	if !ok {
		h.logger.Info("login failed")
		h.render(w, http.StatusUnauthorized, next, msgWrongPassword)
		return
	}
	value, _ := h.tokens.Issue(label)
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(MaxAge / time.Second),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	h.logger.Info("login", "label", label)
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, next, http.StatusSeeOther)
}

// match compares the submitted password with every configured one (digests, constant time, no
// early exit) and returns the label of the first match.
func (h *Handler) match(password string) (string, bool) {
	got := sha256.Sum256([]byte(password))
	found := -1
	for i := range h.digests {
		if subtle.ConstantTimeCompare(got[:], h.digests[i][:]) == 1 && found < 0 {
			found = i
		}
	}
	if found < 0 {
		return "", false
	}
	return h.cfg.Passwords[found].Label, true
}

// sanitizeNext allows only same-origin paths, so the login cannot redirect to another site.
func (h *Handler) sanitizeNext(next string) string {
	if !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") || strings.HasPrefix(next, `/\`) {
		return h.cfg.DefaultNext
	}
	for _, c := range next {
		if c < 0x20 || c == 0x7f || c == '\\' {
			return h.cfg.DefaultNext
		}
	}
	return next
}
