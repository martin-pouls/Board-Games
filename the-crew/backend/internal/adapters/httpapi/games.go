package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/martin-pouls/shared/backend/httpkit"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
	"github.com/martin-pouls/the-crew/backend/internal/core/ports"
)

type gamesAPI struct {
	svc    ports.GameService
	logger *slog.Logger
}

func registerGameRoutes(mux *http.ServeMux, a *gamesAPI) {
	mux.HandleFunc("POST /api/games", a.createGame)
	mux.HandleFunc("GET /api/games/{code}", a.summary)
	mux.HandleFunc("POST /api/games/{code}/players", a.joinGame)
	mux.HandleFunc("GET /api/games/{code}/view", a.authed(a.view))
	mux.HandleFunc("POST /api/games/{code}/leave", a.leave)
	mux.HandleFunc("POST /api/games/{code}/end-game", a.authed(a.endGame))
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

func (a *gamesAPI) writeView(w http.ResponseWriter, r *http.Request, v domain.PlayerView, err error) {
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, apidto.FromView(v))
}

func sessionResponse(s ports.Session) apidto.SessionResponse {
	return apidto.SessionResponse{
		GameID:   string(s.GameID),
		PlayerID: string(s.PlayerID),
		Token:    s.Token,
		View:     apidto.FromView(s.View),
	}
}

// authedFunc handles an authenticated request and returns the actor's view.
type authedFunc func(w http.ResponseWriter, r *http.Request, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)

// authed wraps fn with bearer authentication and view/error writing.
func (a *gamesAPI) authed(fn authedFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := gameCode(r)
		player, err := a.authenticate(r, id)
		if err != nil {
			a.writeError(w, r, err)
			return
		}
		v, err := fn(w, r, id, player)
		a.writeView(w, r, v, err)
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

func (a *gamesAPI) view(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.View(r.Context(), id, p)
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

func (a *gamesAPI) endGame(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.EndGame(r.Context(), id, p)
}
