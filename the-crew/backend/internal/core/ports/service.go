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
	// SetOptions changes the lobby options (host, lobby).
	SetOptions(ctx context.Context, id domain.GameID, player domain.PlayerID, opts domain.Options) (domain.PlayerView, error)
	// SelectMission picks the mission to play next (host, lobby).
	SelectMission(ctx context.Context, id domain.GameID, player domain.PlayerID, mission int) (domain.PlayerView, error)
	// StartMission deals the first attempt of the selected mission (host, lobby).
	StartMission(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	// AdjustTokens swaps (mission 23) or moves (mission 40) the token of task from to task to.
	AdjustTokens(ctx context.Context, id domain.GameID, player domain.PlayerID, from, to int) (domain.PlayerView, error)
	SkipTokens(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	AnswerPoll(ctx context.Context, id domain.GameID, player domain.PlayerID, yes bool) (domain.PlayerView, error)
	// PickSeats is the commander's choice in the pick step: one seat, or two for the roles of mission 50.
	PickSeats(ctx context.Context, id domain.GameID, player domain.PlayerID, seats []int) (domain.PlayerView, error)
	// PickTask takes a task in the open distribution (the commander acts for the dummy).
	PickTask(ctx context.Context, id domain.GameID, player domain.PlayerID, taskID int) (domain.PlayerView, error)
	SendDistress(ctx context.Context, id domain.GameID, player domain.PlayerID, dir domain.Direction) (domain.PlayerView, error)
	SkipDistress(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	ChoosePassCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) (domain.PlayerView, error)
	Handover(ctx context.Context, id domain.GameID, player domain.PlayerID, taskID, toSeat int) (domain.PlayerView, error)
	// Communicate: pos is ignored in a dead zone.
	Communicate(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card, pos domain.Position) (domain.PlayerView, error)
	// PlayCard plays for the seat whose turn it is (the commander plays for the dummy).
	PlayCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) (domain.PlayerView, error)
	GiveUp(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	// NextAttempt deals a new attempt of the same mission after a loss (host, debrief).
	NextAttempt(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	// NextMission starts the next mission after a win (host, debrief).
	NextMission(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	BackToLobby(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	// EndGame lets the host close the session for everyone (phase "ended").
	EndGame(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
}
