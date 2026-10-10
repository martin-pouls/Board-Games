package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/martin-pouls/the-game/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-game/backend/internal/core/domain"
	"github.com/martin-pouls/the-game/backend/internal/core/ports"
)

// maxBodyBytes caps every request body.
const maxBodyBytes = 4 << 10

type gamesAPI struct {
	svc    ports.GameService
	logger *slog.Logger
}

func registerGameRoutes(mux *http.ServeMux, a *gamesAPI) {
	mux.HandleFunc("POST /api/games", a.createGame)
	mux.HandleFunc("GET /api/games/{code}", a.summary)
	mux.HandleFunc("POST /api/games/{code}/players", a.joinGame)
	mux.HandleFunc("GET /api/games/{code}/view", a.authed(a.view))
	mux.HandleFunc("PUT /api/games/{code}/options", a.authed(a.setOptions))
	mux.HandleFunc("POST /api/games/{code}/start", a.authed(a.start))
	mux.HandleFunc("POST /api/games/{code}/first-player", a.authed(a.chooseFirstPlayer))
	mux.HandleFunc("POST /api/games/{code}/plays", a.authed(a.playCard))
	mux.HandleFunc("POST /api/games/{code}/undo", a.authed(a.undoPlay))
	mux.HandleFunc("POST /api/games/{code}/end-turn", a.authed(a.endTurn))
	mux.HandleFunc("POST /api/games/{code}/give-up", a.authed(a.giveUp))
	mux.HandleFunc("POST /api/games/{code}/end-game", a.authed(a.endGame))
	mux.HandleFunc("POST /api/games/{code}/rematch", a.authed(a.rematch))
}

// gameCode returns the {code} path value, normalised to upper case.
func gameCode(r *http.Request) domain.GameID {
	return domain.GameID(strings.ToUpper(strings.TrimSpace(r.PathValue("code"))))
}

// decodeBody decodes a JSON body capped at maxBodyBytes. Every failure maps to bad_request.
func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return apidto.BadRequest("request body exceeds %d bytes", maxBodyBytes)
		}
		if errors.Is(err, io.EOF) {
			return apidto.BadRequest("empty request body")
		}
		return apidto.BadRequest("malformed JSON: %v", err)
	}
	return nil
}

// writeError maps err to a status and error body; unmapped errors are logged.
func (a *gamesAPI) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, body := apidto.ErrorFor(err)
	if status >= http.StatusInternalServerError {
		a.logger.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	}
	writeJSON(w, status, apidto.ErrorResponse{Error: body})
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
	opts := domain.DefaultOptions()
	if req.Options != nil {
		opts = req.Options.ToDomain()
	}
	s, err := a.svc.CreateGame(r.Context(), req.Name, opts)
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

func (a *gamesAPI) setOptions(w http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	var req apidto.GameOptions
	if err := decodeBody(w, r, &req); err != nil {
		return domain.PlayerView{}, err
	}
	return a.svc.SetOptions(r.Context(), id, p, req.ToDomain())
}

func (a *gamesAPI) start(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.StartGame(r.Context(), id, p)
}

func (a *gamesAPI) chooseFirstPlayer(w http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	var req apidto.ChooseFirstPlayerRequest
	if err := decodeBody(w, r, &req); err != nil {
		return domain.PlayerView{}, err
	}
	if req.PlayerID == "" {
		return domain.PlayerView{}, apidto.BadRequest("missing playerId")
	}
	return a.svc.ChooseFirstPlayer(r.Context(), id, p, domain.PlayerID(req.PlayerID))
}

func (a *gamesAPI) playCard(w http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	var req apidto.PlayCardRequest
	if err := decodeBody(w, r, &req); err != nil {
		return domain.PlayerView{}, err
	}
	if req.Card == nil || req.Pile == nil {
		return domain.PlayerView{}, apidto.BadRequest("card and pile are required")
	}
	return a.svc.PlayCard(r.Context(), id, p, *req.Card, *req.Pile)
}

func (a *gamesAPI) undoPlay(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.UndoPlay(r.Context(), id, p)
}

func (a *gamesAPI) endTurn(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.EndTurn(r.Context(), id, p)
}

func (a *gamesAPI) giveUp(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.GiveUp(r.Context(), id, p)
}

func (a *gamesAPI) rematch(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.Rematch(r.Context(), id, p)
}

func (a *gamesAPI) endGame(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.EndGame(r.Context(), id, p)
}
