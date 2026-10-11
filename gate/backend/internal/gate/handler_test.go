package gate

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func newTestHandler(t *testing.T) (*Handler, *clock, *strings.Builder) {
	t.Helper()
	c := &clock{t0}
	logs := &strings.Builder{}
	cfg := Config{Secret: testSecret, Passwords: testPws, BasePath: "/gate", DefaultNext: "/board-games/"}
	return NewHandler(cfg, slog.New(slog.NewTextHandler(logs, nil)), c.now), c, logs
}

func do(h http.Handler, method, target string, hdr map[string]string, body string) *httptest.ResponseRecorder {
	var rd io.Reader
	if body != "" {
		rd = strings.NewReader(body)
	}
	r := httptest.NewRequest(method, target, rd)
	r.RemoteAddr = "127.0.0.1:1234"
	if body != "" {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range hdr {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func loginBody(pw, next string) string {
	return url.Values{"password": {pw}, "next": {next}}.Encode()
}

func TestVerify(t *testing.T) {
	h, _, _ := newTestHandler(t)
	cookie, _ := h.tokens.Issue("family")
	bad := cookie[:len(cookie)-2] + "AA"

	tests := []struct {
		name         string
		hdr          map[string]string
		wantStatus   int
		wantLocation string
		wantJSON     bool
	}{
		{"valid cookie", map[string]string{"Cookie": CookieName + "=" + cookie}, 200, "", false},
		{"valid cookie on fetch", map[string]string{"Cookie": CookieName + "=" + cookie, "Sec-Fetch-Mode": "cors"}, 200, "", false},
		{"navigate", map[string]string{"Sec-Fetch-Mode": "navigate", "X-Forwarded-Uri": "/regicide/?code=ABCD&x=1"}, 302,
			"/gate/login?next=" + url.QueryEscape("/regicide/?code=ABCD&x=1"), false},
		{"navigate bad cookie", map[string]string{"Cookie": CookieName + "=" + bad, "Sec-Fetch-Mode": "navigate", "X-Forwarded-Uri": "/the-game/"}, 302,
			"/gate/login?next=%2Fthe-game%2F", false},
		{"navigate HEAD", map[string]string{"Sec-Fetch-Mode": "navigate", "X-Forwarded-Method": "HEAD", "X-Forwarded-Uri": "/a"}, 302, "/gate/login?next=%2Fa", false},
		{"accept html, no fetch mode", map[string]string{"Accept": "text/html,application/xhtml+xml", "X-Forwarded-Uri": "/fox-duet/"}, 302,
			"/gate/login?next=%2Ffox-duet%2F", false},
		{"navigate without uri", map[string]string{"Sec-Fetch-Mode": "navigate"}, 302, "/gate/login?next=", false},
		{"fetch", map[string]string{"Sec-Fetch-Mode": "cors", "Accept": "application/json"}, 401, "", true},
		{"fetch accepting html", map[string]string{"Sec-Fetch-Mode": "cors", "Accept": "text/html"}, 401, "", true},
		{"websocket", map[string]string{"Sec-Fetch-Mode": "websocket"}, 401, "", true},
		{"no headers", nil, 401, "", true},
		{"accept json only", map[string]string{"Accept": "application/json"}, 401, "", true},
		{"POST navigation", map[string]string{"Sec-Fetch-Mode": "navigate", "X-Forwarded-Method": "POST", "X-Forwarded-Uri": "/a"}, 401, "", true},
		{"POST accept html", map[string]string{"Accept": "text/html", "X-Forwarded-Method": "POST"}, 401, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := do(h, "GET", "/verify", tt.hdr, "")
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantStatus == 200 {
				if w.Body.Len() != 0 {
					t.Errorf("body %q, want empty", w.Body)
				}
				return
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Errorf("Cache-Control = %q", w.Header().Get("Cache-Control"))
			}
			if tt.wantStatus == 302 {
				if got := w.Header().Get("Location"); got != tt.wantLocation {
					t.Errorf("Location = %q, want %q", got, tt.wantLocation)
				}
			}
			if tt.wantJSON {
				if w.Header().Get("Content-Type") != "application/json" {
					t.Errorf("Content-Type = %q", w.Header().Get("Content-Type"))
				}
				var body struct {
					Error struct{ Code, Message string }
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Error.Code != "login_required" || body.Error.Message != "login required" {
					t.Errorf("body = %s", w.Body)
				}
			}
		})
	}
}

func TestVerifyExpiredCookie(t *testing.T) {
	h, c, _ := newTestHandler(t)
	cookie, _ := h.tokens.Issue("family")
	hdr := map[string]string{"Cookie": CookieName + "=" + cookie, "Sec-Fetch-Mode": "cors"}
	if w := do(h, "GET", "/verify", hdr, ""); w.Code != 200 {
		t.Fatalf("fresh: %d", w.Code)
	}
	c.t = t0.Add(MaxAge + time.Minute)
	if w := do(h, "GET", "/verify", hdr, ""); w.Code != 401 {
		t.Errorf("expired: %d", w.Code)
	}
}

func TestLoginPage(t *testing.T) {
	h, _, _ := newTestHandler(t)
	w := do(h, "GET", "/login?next=%2Fregicide%2F%3Fcode%3DABCD", nil, "")
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{
		`<form method="post" action="/gate/login">`,
		`name="password"`,
		`type="text" autocomplete="off"`, // no password-manager prompt
		`name="next" value="/regicide/?code=ABCD"`,
		`/board-games/impressum/`,
		`/board-games/datenschutz/`,
		`<title>Board Games</title>`,
		`name="robots" content="noindex"`,
		`mode-watcher-mode`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page lacks %q", want)
		}
	}
	if strings.Contains(body, `role="alert"`) {
		t.Error("error shown on a fresh page")
	}
	for k, want := range map[string]string{
		"Cache-Control":           "no-store",
		"X-Content-Type-Options":  "nosniff",
		"Referrer-Policy":         "same-origin",
		"X-Frame-Options":         "DENY",
		"Content-Security-Policy": csp,
	} {
		if got := w.Header().Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") {
		t.Errorf("Content-Type = %q", w.Header().Get("Content-Type"))
	}
}

func TestLoginPageEscapesNext(t *testing.T) {
	h, _, _ := newTestHandler(t)
	w := do(h, "GET", "/login?next="+url.QueryEscape(`/a"><script>alert(1)</script>`), nil, "")
	if strings.Contains(w.Body.String(), "<script>alert") {
		t.Error("next is not escaped")
	}
}

func TestLoginPageLoggedIn(t *testing.T) {
	h, _, _ := newTestHandler(t)
	cookie, _ := h.tokens.Issue("friends")
	hdr := map[string]string{"Cookie": CookieName + "=" + cookie}
	w := do(h, "GET", "/login?next=%2Fregicide%2F", hdr, "")
	if w.Code != 303 || w.Header().Get("Location") != "/regicide/" {
		t.Errorf("got %d %q", w.Code, w.Header().Get("Location"))
	}
	w = do(h, "GET", "/login?next=https%3A%2F%2Fevil.com", hdr, "")
	if w.Code != 303 || w.Header().Get("Location") != "/board-games/" {
		t.Errorf("evil next: got %d %q", w.Code, w.Header().Get("Location"))
	}
}

func TestLoginSuccess(t *testing.T) {
	for _, p := range testPws {
		t.Run(p.Label, func(t *testing.T) {
			h, _, logs := newTestHandler(t)
			w := do(h, "POST", "/login", nil, loginBody(p.Password, "/regicide/?code=ABCD"))
			if w.Code != 303 || w.Header().Get("Location") != "/regicide/?code=ABCD" {
				t.Fatalf("got %d %q", w.Code, w.Header().Get("Location"))
			}
			cs := w.Result().Cookies()
			if len(cs) != 1 {
				t.Fatalf("cookies = %v", cs)
			}
			c := cs[0]
			if c.Name != "bg_gate" || c.Path != "/" || c.MaxAge != 31536000 || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode {
				t.Errorf("cookie attributes: %+v", c)
			}
			if label, ok := h.tokens.Verify(c.Value); !ok || label != p.Label {
				t.Errorf("cookie verifies as %q, %v", label, ok)
			}
			if !strings.Contains(logs.String(), "login") || !strings.Contains(logs.String(), "label="+p.Label) {
				t.Errorf("log lacks login line: %s", logs)
			}
			if strings.Contains(logs.String(), p.Password) {
				t.Error("log leaks the password")
			}
			// the cookie then opens /verify
			w = do(h, "GET", "/verify", map[string]string{"Cookie": c.Name + "=" + c.Value}, "")
			if w.Code != 200 {
				t.Errorf("verify with new cookie: %d", w.Code)
			}
		})
	}
}

func TestLoginWrongPassword(t *testing.T) {
	h, _, logs := newTestHandler(t)
	for _, pw := range []string{"nope", "", "pw-famil", "pw-family ", "PW-FAMILY"} {
		w := do(h, "POST", "/login", nil, loginBody(pw, "/regicide/"))
		if w.Code != 401 {
			t.Errorf("%q: status = %d", pw, w.Code)
		}
		if len(w.Result().Cookies()) != 0 {
			t.Errorf("%q: cookie set", pw)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Wrong password.") || !strings.Contains(body, `role="alert"`) {
			t.Errorf("%q: no error message", pw)
		}
		if !strings.Contains(body, `value="/regicide/"`) {
			t.Errorf("%q: next lost", pw)
		}
	}
	if !strings.Contains(logs.String(), "login failed") {
		t.Error("no failure log")
	}
	if strings.Contains(logs.String(), "pw-famil") {
		t.Error("log leaks a password")
	}
}

func TestLoginBodyTooLarge(t *testing.T) {
	h, _, _ := newTestHandler(t)
	w := do(h, "POST", "/login", nil, "password="+strings.Repeat("a", 8<<10))
	if w.Code != 400 || len(w.Result().Cookies()) != 0 {
		t.Errorf("status = %d", w.Code)
	}
}

func TestLoginRateLimit(t *testing.T) {
	h, c, _ := newTestHandler(t)
	post := func(ip, pw string) *httptest.ResponseRecorder {
		return do(h, "POST", "/login", map[string]string{"X-Forwarded-For": ip}, loginBody(pw, ""))
	}
	for i := 0; i < LoginRate.Burst; i++ {
		if w := post("203.0.113.7", "bad"); w.Code != 401 {
			t.Fatalf("attempt %d: status %d", i, w.Code)
		}
	}
	w := post("203.0.113.7", "bad")
	if w.Code != 429 || !strings.Contains(w.Body.String(), "Too many attempts. Try again in a minute.") || w.Header().Get("Retry-After") == "" {
		t.Fatalf("over the burst: %d %s", w.Code, w.Body)
	}
	// Even the right password is refused while limited.
	if w := post("203.0.113.7", "pw-family"); w.Code != 429 || len(w.Result().Cookies()) != 0 {
		t.Errorf("right password while limited: %d", w.Code)
	}
	// Another client is unaffected.
	if w := post("203.0.113.8", "pw-family"); w.Code != 303 {
		t.Errorf("other client: %d", w.Code)
	}
	// Tokens come back over time.
	c.t = c.t.Add(LoginRate.Every)
	if w := post("203.0.113.7", "pw-family"); w.Code != 303 {
		t.Errorf("after refill: %d", w.Code)
	}
	// GET /login and /verify are not limited.
	if w := do(h, "GET", "/login", map[string]string{"X-Forwarded-For": "203.0.113.7"}, ""); w.Code != 200 {
		t.Errorf("GET /login: %d", w.Code)
	}
}

func TestSanitizeNext(t *testing.T) {
	h, _, _ := newTestHandler(t)
	const def = "/board-games/"
	tests := []struct{ in, want string }{
		{"/regicide/?code=ABCD", "/regicide/?code=ABCD"},
		{"/", "/"},
		{"/the-game/", "/the-game/"},
		{"/a%2Fb", "/a%2Fb"},
		{"", def},
		{"https://evil.com", def},
		{"http://evil.com/", def},
		{"//evil.com", def},
		{"///evil.com", def},
		{`/\evil.com`, def},
		{`/a\b`, def},
		{"evil.com", def},
		{"regicide/", def},
		{"javascript:alert(1)", def},
		{"/%0d%0a", "/%0d%0a"}, // still percent-encoded: harmless in a Location header
		{"/a\r\nSet-Cookie: x=1", def},
		{"/a\nb", def},
		{"/a\tb", def},
		{"/a\x00b", def},
		{"/a\x7fb", def},
	}
	for _, tt := range tests {
		if got := h.sanitizeNext(tt.in); got != tt.want {
			t.Errorf("sanitizeNext(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLoginOpenRedirect(t *testing.T) {
	h, _, _ := newTestHandler(t)
	for _, next := range []string{"https://evil.com", "//evil.com", `/\evil.com`, "/a\r\nX: y", ""} {
		w := do(h, "POST", "/login", nil, loginBody("pw-family", next))
		if w.Code != 303 || w.Header().Get("Location") != "/board-games/" {
			t.Errorf("next %q: %d %q", next, w.Code, w.Header().Get("Location"))
		}
	}
}

func TestOtherRoutes(t *testing.T) {
	h, _, _ := newTestHandler(t)
	tests := []struct {
		method, path string
		status       int
		location     string
	}{
		{"GET", "/", 302, "/gate/login"},
		{"GET", "/healthz", 200, ""},
		{"GET", "/nope", 404, ""},
		{"GET", "/login/extra", 404, ""},
		{"GET", "/verify/x", 404, ""},
		{"POST", "/verify", 405, ""},
		{"PUT", "/login", 405, ""},
		{"POST", "/", 405, ""},
	}
	for _, tt := range tests {
		w := do(h, tt.method, tt.path, nil, "")
		if w.Code != tt.status {
			t.Errorf("%s %s = %d, want %d", tt.method, tt.path, w.Code, tt.status)
		}
		if tt.location != "" && w.Header().Get("Location") != tt.location {
			t.Errorf("%s %s Location = %q", tt.method, tt.path, w.Header().Get("Location"))
		}
	}
	if w := do(h, "GET", "/healthz", nil, ""); w.Body.String() != "ok" {
		t.Errorf("healthz body %q", w.Body)
	}
}

func TestMatchChecksEveryPassword(t *testing.T) {
	h, _, _ := newTestHandler(t)
	for _, p := range testPws {
		if label, ok := h.match(p.Password); !ok || label != p.Label {
			t.Errorf("match(%q) = %q, %v", p.Password, label, ok)
		}
	}
	if _, ok := h.match("x"); ok {
		t.Error("matched a wrong password")
	}
}
