package ws

import (
	"context"
	"log/slog"

	"github.com/martin-pouls/shared/backend/wskit"

	"github.com/martin-pouls/fox-duet/backend/internal/adapters/apidto"
	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
	"github.com/martin-pouls/fox-duet/backend/internal/core/ports"
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
	h.PublishViews(string(u.Game.ID), domain.ErrUnknownPlayer, func(pid string) (any, error) {
		v, err := h.viewFor(u.Game, domain.PlayerID(pid))
		if err != nil {
			return nil, err
		}
		return apidto.ViewMessage(apidto.FromView(v), events), nil
	})
}

// Online returns the sorted IDs of players with at least one connection.
func (h *Hub) Online(id domain.GameID) []domain.PlayerID {
	return wskit.OnlineAs[domain.PlayerID](h.Hub, string(id))
}
