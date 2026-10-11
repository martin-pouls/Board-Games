package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/martin-pouls/shared/backend/wskit"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
	"github.com/martin-pouls/the-crew/backend/internal/core/ports"
)

// Hub broadcasts committed updates to the live connections of a game.
// It implements ports.EventPublisher on top of the shared wskit.Hub.
type Hub struct {
	*wskit.Hub

	// viewFor builds a player's projection; injectable for tests.
	viewFor func(*domain.Game, domain.PlayerID) (domain.PlayerView, error)
}

var _ ports.EventPublisher = (*Hub)(nil)

// NewHub returns an empty hub. A nil logger discards logs.
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{Hub: wskit.NewHub(logger), viewFor: (*domain.Game).ViewFor}
}

// Publish sends each connected player their own fresh view plus the events.
// Views are built and marshalled once per distinct player; sockets of a player
// who is no longer seated (they left) are closed with 4401.
func (h *Hub) Publish(_ context.Context, u ports.GameUpdate) {
	if u.Game == nil {
		return
	}
	events := apidto.FromEvents(u.Events)
	h.Broadcast(string(u.Game.ID), func(pid string) ([]byte, bool) {
		return h.buildView(u.Game, domain.PlayerID(pid), events)
	})
}

// buildView returns the marshalled view message for pid. gone reports that pid
// is no longer seated in the game (they left).
func (h *Hub) buildView(g *domain.Game, pid domain.PlayerID, events []apidto.Event) (payload []byte, gone bool) {
	v, err := h.viewFor(g, pid)
	if errors.Is(err, domain.ErrUnknownPlayer) {
		return nil, true
	}
	if err != nil {
		h.Logger().Error("ws: build view", "game", g.ID, "player", pid, "err", err)
		return nil, false
	}
	b, err := json.Marshal(apidto.ViewMessage(apidto.FromView(v), events))
	if err != nil {
		h.Logger().Error("ws: marshal view", "game", g.ID, "player", pid, "err", err)
		return nil, false
	}
	return b, false
}

// Online returns the sorted IDs of players with at least one connection.
func (h *Hub) Online(id domain.GameID) []domain.PlayerID {
	var out []domain.PlayerID
	for _, p := range h.Hub.Online(string(id)) {
		out = append(out, domain.PlayerID(p))
	}
	return out
}
