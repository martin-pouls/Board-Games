package wskit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coder/websocket"

	"github.com/martin-pouls/shared/backend/idgen"
	"github.com/martin-pouls/shared/backend/wire"
)

// Backend is what a game plugs into the endpoint: authentication, the first view, its
// commands and its error mapping. Mutating commands publish through the game's service (and so
// through Hub.Broadcast) before Handle returns, so the broadcast view reaches the socket before
// the ack.
type Backend interface {
	// Authenticate resolves a bearer token to a player seated in game gameID.
	Authenticate(ctx context.Context, gameID, token string) (playerID string, err error)
	// Hello returns the first message of a new connection: the player's current view message.
	Hello(ctx context.Context, gameID, playerID string) ([]byte, error)
	// Handle runs one client command. cmd is the decoded envelope, data the whole message.
	// nil means success (the endpoint acks); an error is mapped with MapError.
	Handle(ctx context.Context, gameID, playerID string, cmd wire.Command, data []byte) error
	// MapError maps an error to its HTTP status and wire body. Connect errors that map to 401
	// close with CloseUnauthorized, 404 with CloseGameNotFound, anything else with 1011.
	MapError(err error) (status int, body wire.ErrorBody)
}

// Options configures the endpoint.
type Options struct {
	// OriginPatterns are extra allowed Origin host patterns (see
	// websocket.AcceptOptions). Empty means the default same-host check.
	OriginPatterns []string
}

// handler serves GET /ws/games/{code}?token=<token>.
type handler struct {
	hub     *Hub
	backend Backend
	opts    Options
}

// NewHandler returns the WebSocket endpoint. Mount it on a pattern with a {code} path value
// (httpkit.NewRouter does: GET /ws/games/{code}).
func NewHandler(hub *Hub, backend Backend, opts Options) http.Handler {
	return &handler{hub: hub, backend: backend, opts: opts}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gameID := idgen.NormalizeCode(r.PathValue("code"))
	token := r.URL.Query().Get("token")

	// Accept first so the browser can see the close code on auth failure. Views are repetitive
	// JSON: compress them (the library leaves messages under 512 bytes, e.g. acks, as they are),
	// without context takeover so a connection keeps no compression state between messages.
	wsc, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns:  h.opts.OriginPatterns,
		CompressionMode: websocket.CompressionNoContextTakeover,
	})
	if err != nil {
		h.hub.logger.Debug("ws accept failed", "path", r.URL.Path, "err", err)
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// An unknown game answers 4404 even without a token.
	playerID, err := h.backend.Authenticate(ctx, gameID, token)
	if err == nil && (token == "" || playerID == "") {
		_ = wsc.Close(CloseUnauthorized, wire.CodeUnauthorized)
		return
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

	hello, err := h.backend.Hello(ctx, gameID, playerID)
	if err != nil {
		h.hub.unregister(c)
		cancel()
		<-writerDone
		h.closeWithError(wsc, gameID, err)
		return
	}
	c.enqueue(hello)
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
func (h *handler) closeWithError(wsc *websocket.Conn, gameID string, err error) {
	status, body := h.backend.MapError(err)
	switch status {
	case http.StatusUnauthorized:
		_ = wsc.Close(CloseUnauthorized, body.Code)
	case http.StatusNotFound:
		_ = wsc.Close(CloseGameNotFound, body.Code)
	default:
		h.hub.logger.Error("ws connect failed", "game", gameID, "err", err)
		_ = wsc.Close(websocket.StatusInternalError, wire.CodeInternal)
	}
}

// handleMessage decodes one client message, lets the backend run it and enqueues an ack or
// error.
func (h *handler) handleMessage(ctx context.Context, c *conn, data []byte) {
	var cmd wire.Command
	if err := json.Unmarshal(data, &cmd); err != nil {
		h.reply(c)(wire.ErrorMessage("", wire.ErrorBody{
			Code:    wire.CodeBadRequest,
			Message: fmt.Sprintf("bad request: malformed message: %v", err),
		}))
		return
	}
	if err := h.backend.Handle(ctx, c.gameID, c.playerID, cmd, data); err != nil {
		status, body := h.backend.MapError(err)
		if status >= http.StatusInternalServerError {
			h.hub.logger.Error("ws command failed", "game", c.gameID, "player", c.playerID, "err", err)
		}
		h.reply(c)(wire.ErrorMessage(cmd.RequestID, body))
		return
	}
	h.reply(c)(wire.AckMessage(cmd.RequestID))
}

// reply returns a sink for a marshalled message to c: h.reply(c)(wire.AckMessage(id)).
func (h *handler) reply(c *conn) func(b []byte, err error) {
	return func(b []byte, err error) {
		if err != nil {
			h.hub.logger.Error("ws: marshal reply", "err", err)
			return
		}
		c.enqueue(b)
	}
}
