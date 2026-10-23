// Package httpapi is the driving adapter exposing the REST API (and, in
// production, the built frontend). It translates HTTP into calls on the core
// service ports and must not contain game rules. The router skeleton, SPA
// handler and abuse limits are shared (shared/backend httpkit).
package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/martin-pouls/shared/backend/httpkit"

	"github.com/martin-pouls/sky-team/backend/internal/core/ports"
)

// RateLimits are the per-client abuse limits (see httpkit.RateLimits).
type RateLimits = httpkit.RateLimits

// DefaultRateLimits are the production limits.
func DefaultRateLimits() RateLimits { return httpkit.DefaultRateLimits() }

// Deps are the collaborators the router needs.
type Deps struct {
	Logger    *slog.Logger
	Health    func(context.Context) error
	StaticDir string            // optional: directory of the built SPA (frontend/build)
	Games     ports.GameService // optional: when nil, the /api/games routes are not registered
	WS        http.Handler      // optional: mounted at GET /ws/games/{code} when non-nil
	Limits    *RateLimits       // optional: per-client abuse limits; nil disables them
}

// NewRouter builds the HTTP handler. Routes use Go 1.22+ method/path patterns.
func NewRouter(d Deps) http.Handler {
	if d.Logger == nil {
		d.Logger = slog.New(slog.DiscardHandler)
	}
	var api func(*http.ServeMux)
	if d.Games != nil {
		a := &gamesAPI{svc: d.Games, logger: d.Logger}
		api = func(mux *http.ServeMux) { registerGameRoutes(mux, a) }
	}
	return httpkit.NewRouter(httpkit.RouterDeps{
		Logger:    d.Logger,
		Health:    d.Health,
		StaticDir: d.StaticDir,
		API:       api,
		WS:        d.WS,
		Limits:    d.Limits,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) { httpkit.WriteJSON(w, status, v) }
