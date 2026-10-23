package ports

import (
	"context"

	"github.com/martin-pouls/sky-team/backend/internal/core/domain"
)

// GameUpdate is a committed change to a game.
type GameUpdate struct {
	Game   *domain.Game // read-only; must not be retained after Publish returns
	Events []domain.Event
}

// EventPublisher pushes committed updates to connected clients (driven port).
type EventPublisher interface {
	// Publish must not block on network I/O: compute per-player payloads synchronously, enqueue, return.
	Publish(ctx context.Context, u GameUpdate)
}

// NopPublisher discards every update (tests, CLI).
type NopPublisher struct{}

// Publish does nothing.
func (NopPublisher) Publish(context.Context, GameUpdate) {}

var _ EventPublisher = NopPublisher{}
