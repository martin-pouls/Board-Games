// Package httpapi is the driving adapter exposing the REST API (and, in
// production, the built frontend). It translates HTTP into calls on the core
// service ports and must not contain game rules.
package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Deps are the collaborators the router needs. Core service ports are added here.
type Deps struct {
	Logger    *slog.Logger
	Health    func(context.Context) error
	StaticDir string // optional: directory of the built SPA (frontend/build)
}

// NewRouter builds the HTTP handler. Routes use Go 1.22+ method/path patterns.
func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := d.Health(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	if d.StaticDir != "" {
		mux.Handle("/", spaHandler(d.StaticDir))
	}

	return logRequests(d.Logger, mux)
}

// spaHandler serves static files and falls back to index.html for client-side routes.
func spaHandler(dir string) http.Handler {
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Debug("request", "method", r.Method, "path", r.URL.Path, "dur", time.Since(start))
	})
}
