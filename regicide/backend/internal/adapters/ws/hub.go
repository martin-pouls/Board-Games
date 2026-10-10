package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"slices"
	"sync"

	"github.com/martin-pouls/regicide/backend/internal/adapters/apidto"
	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/ports"
)

// Hub tracks live connections per game and broadcasts committed updates.
// It implements ports.EventPublisher. The mutex is never held during network
// I/O: enqueueing onto a conn's buffered channel is non-blocking.
type Hub struct {
	logger *slog.Logger

	mu    sync.Mutex
	rooms map[domain.GameID]map[*conn]struct{}

	// viewFor builds a player's projection; injectable for tests.
	viewFor func(*domain.Game, domain.PlayerID) (domain.PlayerView, error)
}

var _ ports.EventPublisher = (*Hub)(nil)

// NewHub returns an empty hub. A nil logger discards logs.
func NewHub(logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Hub{
		logger:  logger,
		rooms:   make(map[domain.GameID]map[*conn]struct{}),
		viewFor: (*domain.Game).ViewFor,
	}
}

// Publish sends each connected player their own fresh view plus the events.
// Views are built and marshalled once per distinct player, then enqueued
// without blocking; a conn with a full buffer is closed as a slow consumer.
func (h *Hub) Publish(_ context.Context, u ports.GameUpdate) {
	if u.Game == nil {
		return
	}
	conns := h.snapshot(u.Game.ID)
	if len(conns) == 0 {
		return
	}
	events := apidto.FromEvents(u.Events)
	payloads := make(map[domain.PlayerID][]byte)
	left := make(map[domain.PlayerID]bool)
	for _, c := range conns {
		if _, done := payloads[c.playerID]; done || left[c.playerID] {
			continue
		}
		payload, gone := h.buildView(u.Game, c.playerID, events)
		if gone {
			left[c.playerID] = true
			continue
		}
		payloads[c.playerID] = payload
	}
	for _, c := range conns {
		if left[c.playerID] {
			c.kick(closeUnauthorized, apidto.CodeUnauthorized)
			continue
		}
		if p := payloads[c.playerID]; p != nil {
			c.enqueue(p)
		}
	}
}

// buildView returns the marshalled view message for pid. gone reports that pid
// is no longer seated in the game (they left); the caller kicks their sockets.
func (h *Hub) buildView(g *domain.Game, pid domain.PlayerID, events []apidto.Event) (payload []byte, gone bool) {
	v, err := h.viewFor(g, pid)
	if errors.Is(err, domain.ErrUnknownPlayer) {
		return nil, true
	}
	if err != nil {
		h.logger.Error("ws: build view", "game", g.ID, "player", pid, "err", err)
		return nil, false
	}
	b, err := json.Marshal(apidto.ViewMessage(apidto.FromView(v), events))
	if err != nil {
		h.logger.Error("ws: marshal view", "game", g.ID, "player", pid, "err", err)
		return nil, false
	}
	return b, false
}

// snapshot copies the room's conns so callers can work without the lock.
func (h *Hub) snapshot(id domain.GameID) []*conn {
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
func (h *Hub) broadcastPresence(id domain.GameID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.rooms[id]
	if len(room) == 0 {
		return
	}
	seen := make(map[domain.PlayerID]struct{}, len(room))
	online := make([]string, 0, len(room))
	for c := range room {
		if _, ok := seen[c.playerID]; ok {
			continue
		}
		seen[c.playerID] = struct{}{}
		online = append(online, string(c.playerID))
	}
	slices.Sort(online)
	b, err := json.Marshal(apidto.PresenceMessage(online))
	if err != nil {
		h.logger.Error("ws: marshal presence", "game", id, "err", err)
		return
	}
	for c := range room {
		c.enqueue(b)
	}
}

// Online returns the sorted IDs of players with at least one connection.
func (h *Hub) Online(id domain.GameID) []domain.PlayerID {
	h.mu.Lock()
	defer h.mu.Unlock()
	seen := make(map[domain.PlayerID]struct{})
	var out []domain.PlayerID
	for c := range h.rooms[id] {
		if _, ok := seen[c.playerID]; !ok {
			seen[c.playerID] = struct{}{}
			out = append(out, c.playerID)
		}
	}
	slices.Sort(out)
	return out
}
