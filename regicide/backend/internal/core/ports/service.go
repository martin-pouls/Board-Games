package ports

import (
	"context"

	"github.com/martin-pouls/regicide/backend/internal/core/domain"
)

// Session is returned when a player creates or joins a game.
type Session struct {
	GameID   domain.GameID
	PlayerID domain.PlayerID
	Token    string // raw token, returned only here
	View     domain.PlayerView
}

// GameService is the driving port used by the REST and WebSocket adapters. The lobby calls (REST)
// return what the caller needs; the gameplay commands (WebSocket) return only an error.
// Every mutation publishes the new state to every connected client.
type GameService interface {
	CreateGame(ctx context.Context, hostName string) (Session, error)
	JoinGame(ctx context.Context, id domain.GameID, name string) (Session, error)
	// Authenticate returns ErrUnauthorized for an unknown token AND for a player no longer seated (left).
	Authenticate(ctx context.Context, id domain.GameID, token string) (domain.PlayerID, error)
	Summary(ctx context.Context, id domain.GameID) (domain.GameSummary, error)
	View(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	LeaveGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// StartGame: first == "" picks the first player at random.
	StartGame(ctx context.Context, id domain.GameID, player, first domain.PlayerID) error
	Play(ctx context.Context, id domain.GameID, player domain.PlayerID, cards []domain.Card) error
	Yield(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	Discard(ctx context.Context, id domain.GameID, player domain.PlayerID, cards []domain.Card) error
	ChooseNext(ctx context.Context, id domain.GameID, player, next domain.PlayerID) error
	UseSoloJester(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	SetJesterSignal(ctx context.Context, id domain.GameID, player domain.PlayerID, s domain.JesterSignal) error
	GiveUp(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// EndGame lets the host close the lobby for everyone (phase "ended").
	EndGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// Rematch creates the next lobby; the old game then carries NextGameID. It is idempotent.
	Rematch(ctx context.Context, id domain.GameID, player domain.PlayerID) error
}
