package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/martin-pouls/shared/backend/httpkit"

	"github.com/martin-pouls/regicide/backend/internal/adapters/apidto"
	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/ports"
)

type gamesAPI struct {
	svc    ports.GameService
	logger *slog.Logger
}

// registerGameRoutes registers the lobby routes. Gameplay (start, play, ...) goes over the WebSocket only.
func registerGameRoutes(mux *http.ServeMux, a *gamesAPI) {
	mux.HandleFunc("POST /api/games", a.createGame)
	mux.HandleFunc("GET /api/games/{code}", a.summary)
	mux.HandleFunc("POST /api/games/{code}/players", a.joinGame)
	mux.HandleFunc("GET /api/games/{code}/view", a.view)
	mux.HandleFunc("POST /api/games/{code}/leave", a.leave)
}

// gameCode returns the {code} path value, normalised to upper case.
func gameCode(r *http.Request) domain.GameID {
	return domain.GameID(httpkit.GameCode(r))
}

// decodeBody decodes a JSON body capped at httpkit.MaxBodyBytes. Every failure maps to bad_request.
func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	return httpkit.DecodeBody(w, r, v)
}

// writeError maps err to a status and error body; unmapped errors are logged.
func (a *gamesAPI) writeError(w http.ResponseWriter, r *http.Request, err error) {
	apidto.Errors.WriteError(w, r, a.logger, err)
}

func sessionResponse(s ports.Session) apidto.SessionResponse {
	return apidto.SessionResponse{
		GameID:   string(s.GameID),
		PlayerID: string(s.PlayerID),
		Token:    s.Token,
		View:     apidto.FromView(s.View),
	}
}

func (a *gamesAPI) createGame(w http.ResponseWriter, r *http.Request) {
	var req apidto.CreateGameRequest
	if err := decodeBody(w, r, &req); err != nil {
		a.writeError(w, r, err)
		return
	}
	s, err := a.svc.CreateGame(r.Context(), req.Name)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, sessionResponse(s))
}

func (a *gamesAPI) summary(w http.ResponseWriter, r *http.Request) {
	s, err := a.svc.Summary(r.Context(), gameCode(r))
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, apidto.FromSummary(s))
}

func (a *gamesAPI) joinGame(w http.ResponseWriter, r *http.Request) {
	var req apidto.JoinGameRequest
	if err := decodeBody(w, r, &req); err != nil {
		a.writeError(w, r, err)
		return
	}
	s, err := a.svc.JoinGame(r.Context(), gameCode(r), req.Name)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, sessionResponse(s))
}

// view answers with the caller's current view.
func (a *gamesAPI) view(w http.ResponseWriter, r *http.Request) {
	id := gameCode(r)
	player, err := a.authenticate(r, id)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	v, err := a.svc.View(r.Context(), id, player)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, apidto.FromView(v))
}

// leave removes the caller from a lobby; it answers 204 without a body.
func (a *gamesAPI) leave(w http.ResponseWriter, r *http.Request) {
	id := gameCode(r)
	player, err := a.authenticate(r, id)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	if err := a.svc.LeaveGame(r.Context(), id, player); err != nil {
		a.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
