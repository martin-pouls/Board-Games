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
// Lobby calls (create, join, leave) are REST; the gameplay commands (start and below)
// come in over the WebSocket. Every mutating call publishes to all connected clients,
// who receive their fresh views that way.
type GameService interface {
	CreateGame(ctx context.Context, hostName string) (Session, error)
	JoinGame(ctx context.Context, id domain.GameID, name string) (Session, error)
	// Authenticate returns ErrUnauthorized for an unknown token AND for a player no longer seated (left).
	Authenticate(ctx context.Context, id domain.GameID, token string) (domain.PlayerID, error)
	Summary(ctx context.Context, id domain.GameID) (domain.GameSummary, error)
	View(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	LeaveGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	StartGame(ctx context.Context, id domain.GameID, player domain.PlayerID, level int) error
	PlayCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) error
	ChooseFoxesPlayer(ctx context.Context, id domain.GameID, player, chosen domain.PlayerID) error
	SwapDecree(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) error
	KeepDecree(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	GiveGift(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) error
	// ChooseMove: delta is one of the pending move options (negative = toward the left end).
	ChooseMove(ctx context.Context, id domain.GameID, player domain.PlayerID, delta int) error
	ChooseForestEnd(ctx context.Context, id domain.GameID, player domain.PlayerID, end domain.End) error
	GiveUp(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// EndGame lets the host close the lobby for everyone (phase "ended").
	EndGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// Rematch creates the follow-up lobby; the old game then carries NextGameID. It is idempotent.
	Rematch(ctx context.Context, id domain.GameID, player domain.PlayerID) error
}
