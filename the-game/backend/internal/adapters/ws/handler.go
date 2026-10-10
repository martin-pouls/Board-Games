package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/coder/websocket"

	"github.com/martin-pouls/the-game/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-game/backend/internal/core/domain"
	"github.com/martin-pouls/the-game/backend/internal/core/ports"
)

// Options configures the WebSocket handler.
type Options struct {
	// OriginPatterns are extra allowed Origin host patterns (see
	// websocket.AcceptOptions). Empty means the default same-host check.
	OriginPatterns []string
}

// Handler serves GET /ws/games/{code}?token=<token>.
type Handler struct {
	hub  *Hub
	svc  ports.GameService
	opts Options
}

// NewHandler returns the WebSocket endpoint handler. It expects to be mounted
// on a pattern with a {code} path value.
func NewHandler(hub *Hub, svc ports.GameService, opts Options) http.Handler {
	return &Handler{hub: hub, svc: svc, opts: opts}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gameID := domain.GameID(strings.ToUpper(strings.TrimSpace(r.PathValue("code"))))
	token := r.URL.Query().Get("token")

	// Accept first so the browser can see the close code on auth failure.
	wsc, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: h.opts.OriginPatterns})
	if err != nil {
		h.hub.logger.Debug("ws accept failed", "path", r.URL.Path, "err", err)
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	playerID, err := h.svc.Authenticate(ctx, gameID, token)
	if token == "" && err == nil {
		err = ports.ErrUnauthorized
	}
	if err != nil {
		h.closeWithError(wsc, gameID, err)
		return
	}
	wsc.SetReadLimit(readLimit)

	c := newConn(wsc, gameID, playerID, h.hub.logger)
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		c.writeLoop(ctx)
	}()

	// Register before loading the view so no committed update is missed.
	h.hub.register(c)
	defer func() {
		if h.hub.unregister(c) {
			h.hub.broadcastPresence(gameID)
		}
		cancel()
		<-writerDone
		_ = wsc.CloseNow()
	}()

	v, err := h.svc.View(ctx, gameID, playerID)
	if err != nil {
		h.hub.unregister(c)
		cancel()
		<-writerDone
		h.closeWithError(wsc, gameID, err)
		return
	}
	if b, err := json.Marshal(apidto.ViewMessage(apidto.FromView(v), nil)); err == nil {
		c.enqueue(b)
	}
	h.hub.broadcastPresence(gameID)

	for {
		_, data, err := wsc.Read(ctx)
		if err != nil {
			return
		}
		h.handleMessage(ctx, c, data)
	}
}

// closeWithError closes an unregistered socket with a code derived from err.
func (h *Handler) closeWithError(wsc *websocket.Conn, gameID domain.GameID, err error) {
	switch {
	case errors.Is(err, ports.ErrUnauthorized):
		_ = wsc.Close(closeUnauthorized, apidto.CodeUnauthorized)
	case errors.Is(err, ports.ErrGameNotFound):
		_ = wsc.Close(closeGameNotFound, apidto.CodeGameNotFound)
	default:
		h.hub.logger.Error("ws connect failed", "game", gameID, "err", err)
		_ = wsc.Close(websocket.StatusInternalError, apidto.CodeInternal)
	}
}

// handleMessage decodes one client message, runs the matching service call
// and enqueues an ack or error. The service publishes before returning, so the
// broadcast view is queued before the ack.
func (h *Handler) handleMessage(ctx context.Context, c *conn, data []byte) {
	var msg apidto.ClientMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		h.reply(c, apidto.ErrorMessage("", h.errorBody(c, apidto.BadRequest("malformed message: %v", err))))
		return
	}
	if err := h.dispatch(ctx, c, msg); err != nil {
		h.reply(c, apidto.ErrorMessage(msg.RequestID, h.errorBody(c, err)))
		return
	}
	h.reply(c, apidto.AckMessage(msg.RequestID))
}

func (h *Handler) dispatch(ctx context.Context, c *conn, msg apidto.ClientMessage) error {
	id, p := c.gameID, c.playerID
	var err error
	switch msg.Type {
	case apidto.CmdSetOptions:
		if msg.Options == nil {
			return apidto.BadRequest("missing options")
		}
		_, err = h.svc.SetOptions(ctx, id, p, msg.Options.ToDomain())
	case apidto.CmdStartGame:
		_, err = h.svc.StartGame(ctx, id, p)
	case apidto.CmdChooseFirstPlayer:
		if msg.PlayerID == "" {
			return apidto.BadRequest("missing playerId")
		}
		_, err = h.svc.ChooseFirstPlayer(ctx, id, p, domain.PlayerID(msg.PlayerID))
	case apidto.CmdPlayCard:
		if msg.Card == nil || msg.Pile == nil {
			return apidto.BadRequest("card and pile are required")
		}
		_, err = h.svc.PlayCard(ctx, id, p, *msg.Card, *msg.Pile)
	case apidto.CmdEndTurn:
		_, err = h.svc.EndTurn(ctx, id, p)
	case apidto.CmdRematch:
		_, err = h.svc.Rematch(ctx, id, p)
	default:
		return apidto.BadRequest("unknown message type %q", msg.Type)
	}
	return err
}

func (h *Handler) errorBody(c *conn, err error) apidto.ErrorBody {
	status, body := apidto.ErrorFor(err)
	if status >= http.StatusInternalServerError {
		h.hub.logger.Error("ws command failed", "game", c.gameID, "player", c.playerID, "err", err)
	}
	return body
}

func (h *Handler) reply(c *conn, m apidto.ServerMessage) {
	b, err := json.Marshal(m)
	if err != nil {
		h.hub.logger.Error("ws: marshal reply", "err", err)
		return
	}
	c.enqueue(b)
}
