package ports

import (
	"context"

	"github.com/martin-pouls/the-game/backend/internal/core/domain"
)

// Session is returned when a player creates or joins a game.
type Session struct {
	GameID   domain.GameID
	PlayerID domain.PlayerID
	Token    string // raw token, returned only here
	View     domain.PlayerView
}

// GameService is the driving port used by the REST and WebSocket adapters.
// Every mutating call publishes the new state to every connected client; create and join also
// return a Session carrying the caller's first view.
type GameService interface {
	CreateGame(ctx context.Context, hostName string, opts domain.Options) (Session, error)
	JoinGame(ctx context.Context, id domain.GameID, name string) (Session, error)
	Authenticate(ctx context.Context, id domain.GameID, token string) (domain.PlayerID, error) // ErrUnauthorized / ErrGameNotFound
	Summary(ctx context.Context, id domain.GameID) (domain.GameSummary, error)
	View(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	// LeaveGame removes the player from the lobby (REST); their token stops working for this game.
	LeaveGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// The gameplay calls below are used by the WebSocket adapter only. They return no view:
	// the service publishes the new state to every connected client, the actor included.
	SetOptions(ctx context.Context, id domain.GameID, player domain.PlayerID, opts domain.Options) error
	StartGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	ChooseFirstPlayer(ctx context.Context, id domain.GameID, player, first domain.PlayerID) error
	PlayCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card, pile int) error
	// UndoPlay takes back the active player's last card of the current turn.
	UndoPlay(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	EndTurn(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// GiveUp lets any seated player end a running game as lost.
	GiveUp(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// EndGame lets the host close the lobby (phase "ended"); a running game is given up instead.
	EndGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// Rematch creates the next lobby game; the old game then carries NextGameID. It is idempotent.
	Rematch(ctx context.Context, id domain.GameID, player domain.PlayerID) error
}
