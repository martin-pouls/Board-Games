package ws

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/martin-pouls/the-game/backend/internal/core/domain"
)

const (
	sendBuffer   = 32
	writeTimeout = 5 * time.Second
	pingInterval = 25 * time.Second
	pingTimeout  = 10 * time.Second
	readLimit    = 4096
)

// Close codes sent to the browser.
const (
	closeUnauthorized websocket.StatusCode = 4401
	closeGameNotFound websocket.StatusCode = 4404
)

// conn is one WebSocket connection of one player in one game.
// All writes happen on the writer goroutine; others only enqueue.
type conn struct {
	ws       *websocket.Conn
	gameID   domain.GameID
	playerID domain.PlayerID
	logger   *slog.Logger

	send chan []byte

	closeOnce   sync.Once
	closing     chan struct{} // closed by kick: the writer closes the socket
	closeCode   websocket.StatusCode
	closeReason string

	pingInterval time.Duration
}

func newConn(c *websocket.Conn, gameID domain.GameID, playerID domain.PlayerID, logger *slog.Logger) *conn {
	return &conn{
		ws:           c,
		gameID:       gameID,
		playerID:     playerID,
		logger:       logger,
		send:         make(chan []byte, sendBuffer),
		closing:      make(chan struct{}),
		pingInterval: pingInterval,
	}
}

// enqueue queues msg without blocking. If the buffer is full, the connection
// is closed with 1008 "slow consumer" (asynchronously, by the writer). It
// reports whether msg was queued.
func (c *conn) enqueue(msg []byte) bool {
	select {
	case <-c.closing:
		return false
	default:
	}
	select {
	case c.send <- msg:
		return true
	default:
		c.kick(websocket.StatusPolicyViolation, "slow consumer")
		return false
	}
}

// kick asks the writer goroutine to close the connection with code/reason.
// It never does network I/O itself. Only the first call wins.
func (c *conn) kick(code websocket.StatusCode, reason string) {
	c.closeOnce.Do(func() {
		c.closeCode = code
		c.closeReason = reason
		close(c.closing)
	})
}

// writeLoop owns all writes to the socket. It exits when ctx is done, after
// a kick, or on a write/ping failure; in the failure cases it closes the
// socket so the read loop unblocks.
func (c *conn) writeLoop(ctx context.Context) {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.closing:
			_ = c.ws.Close(c.closeCode, c.closeReason)
			return
		case msg := <-c.send:
			wctx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.ws.Write(wctx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				c.logger.Debug("ws write failed", "game", c.gameID, "player", c.playerID, "err", err)
				_ = c.ws.CloseNow()
				return
			}
		case <-ticker.C:
			pctx, cancel := context.WithTimeout(ctx, pingTimeout)
			err := c.ws.Ping(pctx)
			cancel()
			if err != nil {
				c.logger.Debug("ws ping failed", "game", c.gameID, "player", c.playerID, "err", err)
				_ = c.ws.CloseNow()
				return
			}
		}
	}
}
