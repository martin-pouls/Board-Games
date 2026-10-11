package ports

import (
	"context"

	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
)

// Session is returned when a player creates or joins a game.
type Session struct {
	GameID   domain.GameID
	PlayerID domain.PlayerID
	Token    string // raw token, returned only here
	View     domain.PlayerView
}

// GameService is the driving port used by the REST and WebSocket adapters.
// Mutating calls return the actor's fresh view and publish to every connected client.
type GameService interface {
	CreateGame(ctx context.Context, hostName string) (Session, error)
	JoinGame(ctx context.Context, id domain.GameID, name string) (Session, error)
	// Authenticate returns ErrUnauthorized for an unknown token AND for a player no longer seated (left).
	Authenticate(ctx context.Context, id domain.GameID, token string) (domain.PlayerID, error)
	Summary(ctx context.Context, id domain.GameID) (domain.GameSummary, error)
	View(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	LeaveGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// EndGame lets the host close the lobby for everyone (phase "ended").
	EndGame(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
}
