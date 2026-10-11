package ports

import (
	"context"

	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
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
	StartGame(ctx context.Context, id domain.GameID, player domain.PlayerID, level int) (domain.PlayerView, error)
	PlayCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) (domain.PlayerView, error)
	ChooseFoxesPlayer(ctx context.Context, id domain.GameID, player, chosen domain.PlayerID) (domain.PlayerView, error)
	SwapDecree(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) (domain.PlayerView, error)
	KeepDecree(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	GiveGift(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) (domain.PlayerView, error)
	// ChooseMove: delta is one of the pending move options (negative = toward the left end).
	ChooseMove(ctx context.Context, id domain.GameID, player domain.PlayerID, delta int) (domain.PlayerView, error)
	ChooseForestEnd(ctx context.Context, id domain.GameID, player domain.PlayerID, end domain.End) (domain.PlayerView, error)
	GiveUp(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	// EndGame lets the host close the lobby for everyone (phase "ended").
	EndGame(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	// Rematch returns the actor's view of the old game, which now carries NextGameID.
	Rematch(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
}
