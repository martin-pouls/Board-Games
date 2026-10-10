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
// All mutating calls return the actor's fresh view and publish to every connected client.
type GameService interface {
	CreateGame(ctx context.Context, hostName string, opts domain.Options) (Session, error)
	JoinGame(ctx context.Context, id domain.GameID, name string) (Session, error)
	Authenticate(ctx context.Context, id domain.GameID, token string) (domain.PlayerID, error) // ErrUnauthorized / ErrGameNotFound
	Summary(ctx context.Context, id domain.GameID) (domain.GameSummary, error)
	View(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	SetOptions(ctx context.Context, id domain.GameID, player domain.PlayerID, opts domain.Options) (domain.PlayerView, error)
	StartGame(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	ChooseFirstPlayer(ctx context.Context, id domain.GameID, player, first domain.PlayerID) (domain.PlayerView, error)
	PlayCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card, pile int) (domain.PlayerView, error)
	// UndoPlay takes back the active player's last card of the current turn.
	UndoPlay(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	EndTurn(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	// GiveUp loses the game for a player stuck mid-turn (instead of undoing).
	GiveUp(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	// EndGame lets the host end a game early, from the lobby on (outcome "aborted").
	EndGame(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	// Rematch returns the actor's view of the old game, which now carries NextGameID.
	Rematch(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
}
