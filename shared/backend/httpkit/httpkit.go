// Package httpkit holds the HTTP plumbing every game's REST adapter shares: JSON in and out,
// error mapping, bearer tokens, the SPA handler, request logging, the router skeleton and the
// per-client abuse limits. Game routes are registered by the game.
package httpkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/martin-pouls/shared/backend/wire"
)

// MaxBodyBytes caps every request body read by DecodeBody.
const MaxBodyBytes = 4 << 10

// ErrBadRequest marks malformed input (bad JSON, missing field, unknown WS type).
// Wrap it with detail via BadRequest.
var ErrBadRequest = errors.New("bad request")

// BadRequest returns an error wrapping ErrBadRequest with a human-readable reason.
func BadRequest(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrBadRequest, fmt.Sprintf(format, args...))
}

// ErrRateLimited marks a request refused by the per-client abuse limits (429).
var ErrRateLimited = errors.New("too many requests, please wait a moment")

// ErrEmptyBody is returned by DecodeBody for a body with no content. It wraps ErrBadRequest.
var ErrEmptyBody = BadRequest("empty request body")

// DecodeBody decodes a JSON body capped at MaxBodyBytes. Every failure wraps ErrBadRequest.
func DecodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return BadRequest("request body exceeds %d bytes", MaxBodyBytes)
		}
		if errors.Is(err, io.EOF) {
			return ErrEmptyBody
		}
		return BadRequest("malformed JSON: %v", err)
	}
	return nil
}

// WriteJSON writes v as JSON with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ErrorEntry maps a sentinel error to its wire code and HTTP status.
type ErrorEntry struct {
	Err    error
	Code   string
	Status int
}

// CommonErrors are the entries every game maps the same way. Games append their port and
// domain errors (unauthorized, game_not_found, conflict, ...).
var CommonErrors = []ErrorEntry{
	{ErrBadRequest, wire.CodeBadRequest, http.StatusBadRequest},
	{ErrRateLimited, wire.CodeRateLimited, http.StatusTooManyRequests},
}

// ErrorMapper maps errors to an HTTP status and wire error body, for REST and WS alike.
type ErrorMapper []ErrorEntry

// Map matches err against the entries in order (errors.Is). Malformed JSON
// (*json.SyntaxError, *json.UnmarshalTypeError) and oversized bodies (*http.MaxBytesError) map
// to bad_request. Anything else maps to 500 "internal" with a generic message; the caller
// should log the real error.
func (m ErrorMapper) Map(err error) (int, wire.ErrorBody) {
	if err == nil {
		return http.StatusInternalServerError, wire.ErrorBody{Code: wire.CodeInternal, Message: wire.InternalMessage}
	}
	for _, e := range m {
		if errors.Is(err, e.Err) {
			return e.Status, wire.ErrorBody{Code: e.Code, Message: err.Error()}
		}
	}
	var (
		syntaxErr *json.SyntaxError
		typeErr   *json.UnmarshalTypeError
		maxErr    *http.MaxBytesError
	)
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) || errors.As(err, &maxErr) {
		return http.StatusBadRequest, wire.ErrorBody{Code: wire.CodeBadRequest, Message: err.Error()}
	}
	return http.StatusInternalServerError, wire.ErrorBody{Code: wire.CodeInternal, Message: wire.InternalMessage}
}

// WriteError writes err as a wire error response. Errors that map to 5xx are logged.
func (m ErrorMapper) WriteError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	status, body := m.Map(err)
	if status >= http.StatusInternalServerError && logger != nil {
		logger.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	}
	WriteJSON(w, status, wire.ErrorResponse{Error: body})
}

// BearerToken extracts the token from "Authorization: Bearer <token>".
// The scheme is matched case-insensitively. It returns "" when absent or malformed.
func BearerToken(r *http.Request) string {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "bearer "
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}

// SPA serves static files from dir and falls back to index.html for client-side routes.
func SPA(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(dir, filepath.FromSlash(filepath.Clean("/"+r.URL.Path)))
		if info, err := os.Stat(p); err != nil || info.IsDir() && r.URL.Path != "/" {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		files.ServeHTTP(w, r)
	})
}

// LogRequests logs method and URL.Path only. Never log the query: the WS token is there.
func LogRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Debug("request", "method", r.Method, "path", r.URL.Path, "dur", time.Since(start))
	})
}

// RouterDeps are the collaborators of NewRouter.
type RouterDeps struct {
	Logger    *slog.Logger
	Health    func(context.Context) error
	StaticDir string                   // optional: directory of the built SPA (frontend/build)
	API       func(mux *http.ServeMux) // optional: registers the game's /api routes
	WS        http.Handler             // optional: mounted at GET /ws/games/{code} when non-nil
	Limits    *RateLimits              // optional: per-client abuse limits; nil disables them
}

// NewRouter builds a game server's HTTP handler: GET /api/health, the game's API routes, a 404
// catch-all for unknown /api paths, the WebSocket endpoint and the SPA, behind the abuse limits
// and request logging. Routes use Go 1.22+ method/path patterns.
func NewRouter(d RouterDeps) http.Handler {
	if d.Logger == nil {
		d.Logger = slog.New(slog.DiscardHandler)
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if d.Health != nil {
			if err := d.Health(ctx); err != nil {
				WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
				return
			}
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	if d.API != nil {
		d.API(mux)
	}

	// Catch-all so the SPA fallback never swallows API typos.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusNotFound, wire.ErrorResponse{Error: wire.ErrorBody{
			Code:    wire.CodeNotFound,
			Message: "no such endpoint",
		}})
	})

	if d.WS != nil {
		mux.Handle("GET /ws/games/{code}", d.WS)
	}

	if d.StaticDir != "" {
		mux.Handle("/", SPA(d.StaticDir))
	}

	var h http.Handler = mux
	if d.Limits != nil {
		h = newRateLimiter(*d.Limits, d.Logger, time.Now).wrap(mux)
	}
	return LogRequests(d.Logger, h)
}

// GameCode returns the {code} path value, normalised like a typed join code.
func GameCode(r *http.Request) string {
	return strings.ToUpper(strings.TrimSpace(r.PathValue("code")))
}
