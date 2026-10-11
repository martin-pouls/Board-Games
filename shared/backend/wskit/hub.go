package wskit

import (
	"log/slog"
	"slices"
	"sync"

	"github.com/martin-pouls/shared/backend/wire"
)

// Hub tracks live connections per game and broadcasts to them. A game's publisher (its
// EventPublisher port) calls Broadcast with a per-player payload builder. The mutex is never
// held during network I/O: enqueueing onto a conn's buffered channel is non-blocking.
type Hub struct {
	logger *slog.Logger

	mu    sync.Mutex
	rooms map[string]map[*conn]struct{}
}

// NewHub returns an empty hub. A nil logger discards logs.
func NewHub(logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Hub{logger: logger, rooms: make(map[string]map[*conn]struct{})}
}

// Logger returns the hub's logger (never nil).
func (h *Hub) Logger() *slog.Logger { return h.logger }

// Broadcast sends every connection of gameID the payload built for its player. build runs once
// per distinct connected player (not under the hub lock). It returns gone=true when the player
// is no longer seated (they left): their sockets are closed with CloseUnauthorized. A nil
// payload with gone=false sends nothing to that player (build failed; build logs it).
// Payloads are enqueued without blocking; a conn with a full buffer is closed as a slow consumer.
func (h *Hub) Broadcast(gameID string, build func(playerID string) (payload []byte, gone bool)) {
	conns := h.snapshot(gameID)
	if len(conns) == 0 {
		return
	}
	payloads := make(map[string][]byte)
	left := make(map[string]bool)
	for _, c := range conns {
		if _, done := payloads[c.playerID]; done || left[c.playerID] {
			continue
		}
		payload, gone := build(c.playerID)
		if gone {
			left[c.playerID] = true
			continue
		}
		payloads[c.playerID] = payload
	}
	for _, c := range conns {
		if left[c.playerID] {
			c.kick(CloseUnauthorized, wire.CodeUnauthorized)
			continue
		}
		if p := payloads[c.playerID]; p != nil {
			c.enqueue(p)
		}
	}
}

// snapshot copies the room's conns so callers can work without the lock.
func (h *Hub) snapshot(id string) []*conn {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.rooms[id]
	out := make([]*conn, 0, len(room))
	for c := range room {
		out = append(out, c)
	}
	return out
}

// register adds c to its game's room.
func (h *Hub) register(c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.rooms[c.gameID]
	if room == nil {
		room = make(map[*conn]struct{})
		h.rooms[c.gameID] = room
	}
	room[c] = struct{}{}
}

// unregister removes c from its room (idempotent) and reports whether it was present.
func (h *Hub) unregister(c *conn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.rooms[c.gameID]
	if _, ok := room[c]; !ok {
		return false
	}
	delete(room, c)
	if len(room) == 0 {
		delete(h.rooms, c.gameID)
	}
	return true
}

// broadcastPresence sends the sorted list of online players to every conn in
// the room. It runs under the lock (enqueue never blocks) so that presence
// messages reach every conn in the same order as the membership changes.
func (h *Hub) broadcastPresence(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.rooms[id]
	if len(room) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(room))
	online := make([]string, 0, len(room))
	for c := range room {
		if _, ok := seen[c.playerID]; ok {
			continue
		}
		seen[c.playerID] = struct{}{}
		online = append(online, c.playerID)
	}
	slices.Sort(online)
	b, err := wire.PresenceMessage(online)
	if err != nil {
		h.logger.Error("ws: marshal presence", "game", id, "err", err)
		return
	}
	for c := range room {
		c.enqueue(b)
	}
}

// Online returns the sorted IDs of players with at least one connection to gameID.
func (h *Hub) Online(gameID string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	seen := make(map[string]struct{})
	var out []string
	for c := range h.rooms[gameID] {
		if _, ok := seen[c.playerID]; !ok {
			seen[c.playerID] = struct{}{}
			out = append(out, c.playerID)
		}
	}
	slices.Sort(out)
	return out
}
