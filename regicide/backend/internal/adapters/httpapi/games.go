package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/martin-pouls/regicide/backend/internal/adapters/apidto"
	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/ports"
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
	mux.HandleFunc("POST /api/games/{code}/leave", a.leave)
	mux.HandleFunc("POST /api/games/{code}/start", a.authed(a.start))
	mux.HandleFunc("POST /api/games/{code}/play", a.authed(a.play))
	mux.HandleFunc("POST /api/games/{code}/yield", a.authed(a.yield))
	mux.HandleFunc("POST /api/games/{code}/discard", a.authed(a.discard))
	mux.HandleFunc("POST /api/games/{code}/next-player", a.authed(a.chooseNext))
	mux.HandleFunc("POST /api/games/{code}/solo-jester", a.authed(a.soloJester))
	mux.HandleFunc("PUT /api/games/{code}/signal", a.authed(a.setSignal))
	mux.HandleFunc("POST /api/games/{code}/give-up", a.authed(a.giveUp))
	mux.HandleFunc("POST /api/games/{code}/rematch", a.authed(a.rematch))
}

// gameCode returns the {code} path value, normalised to upper case.
func gameCode(r *http.Request) domain.GameID {
	return domain.GameID(strings.ToUpper(strings.TrimSpace(r.PathValue("code"))))
}

// errEmptyBody is returned by decodeBody for a body with no content.
var errEmptyBody = apidto.BadRequest("empty request body")

// decodeBody decodes a JSON body capped at maxBodyBytes. Every failure maps to bad_request.
func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return apidto.BadRequest("request body exceeds %d bytes", maxBodyBytes)
		}
		if errors.Is(err, io.EOF) {
			return errEmptyBody
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

// start accepts an optional body; an empty body means a random first player.
func (a *gamesAPI) start(w http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	var req apidto.StartGameRequest
	if err := decodeBody(w, r, &req); err != nil && !errors.Is(err, errEmptyBody) {
		return domain.PlayerView{}, err
	}
	return a.svc.StartGame(r.Context(), id, p, domain.PlayerID(req.FirstPlayerID))
}

// decodeCards decodes a {cards} body; a missing or null list is bad_request.
func decodeCards(w http.ResponseWriter, r *http.Request) ([]domain.Card, error) {
	var req apidto.CardsRequest
	if err := decodeBody(w, r, &req); err != nil {
		return nil, err
	}
	if req.Cards == nil {
		return nil, apidto.BadRequest("cards are required")
	}
	return apidto.CardsToDomain(req.Cards)
}

func (a *gamesAPI) play(w http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	cards, err := decodeCards(w, r)
	if err != nil {
		return domain.PlayerView{}, err
	}
	return a.svc.Play(r.Context(), id, p, cards)
}

func (a *gamesAPI) yield(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.Yield(r.Context(), id, p)
}

func (a *gamesAPI) discard(w http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	cards, err := decodeCards(w, r)
	if err != nil {
		return domain.PlayerView{}, err
	}
	return a.svc.Discard(r.Context(), id, p, cards)
}

func (a *gamesAPI) chooseNext(w http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	var req apidto.ChooseNextRequest
	if err := decodeBody(w, r, &req); err != nil {
		return domain.PlayerView{}, err
	}
	if req.PlayerID == "" {
		return domain.PlayerView{}, apidto.BadRequest("missing playerId")
	}
	return a.svc.ChooseNext(r.Context(), id, p, domain.PlayerID(req.PlayerID))
}

func (a *gamesAPI) soloJester(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.UseSoloJester(r.Context(), id, p)
}

func (a *gamesAPI) setSignal(w http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	var req apidto.SignalRequest
	if err := decodeBody(w, r, &req); err != nil {
		return domain.PlayerView{}, err
	}
	if req.Signal == nil {
		return domain.PlayerView{}, apidto.BadRequest("missing signal")
	}
	return a.svc.SetJesterSignal(r.Context(), id, p, apidto.SignalFromWire(*req.Signal))
}

func (a *gamesAPI) giveUp(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.GiveUp(r.Context(), id, p)
}

func (a *gamesAPI) rematch(_ http.ResponseWriter, r *http.Request, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return a.svc.Rematch(r.Context(), id, p)
}
