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

// GameService is the driving port used by the REST and WebSocket adapters. REST covers the lobby
// (create, join, summary, view, leave); the gameplay commands run over the WebSocket only.
// Gameplay commands return only an error and publish the new state to every connected client.
type GameService interface {
	CreateGame(ctx context.Context, hostName string) (Session, error)
	JoinGame(ctx context.Context, id domain.GameID, name string) (Session, error)
	// Authenticate returns ErrUnauthorized for an unknown token AND for a player no longer seated (left).
	Authenticate(ctx context.Context, id domain.GameID, token string) (domain.PlayerID, error)
	Summary(ctx context.Context, id domain.GameID) (domain.GameSummary, error)
	View(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
	LeaveGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// SetOptions changes the lobby options (host, lobby).
	SetOptions(ctx context.Context, id domain.GameID, player domain.PlayerID, opts domain.Options) error
	// SelectMission picks the mission to play next (host, lobby).
	SelectMission(ctx context.Context, id domain.GameID, player domain.PlayerID, mission int) error
	// StartMission deals the first attempt of the selected mission (host, lobby).
	StartMission(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// AdjustTokens swaps (mission 23) or moves (mission 40) the token of task from to task to.
	AdjustTokens(ctx context.Context, id domain.GameID, player domain.PlayerID, from, to int) error
	SkipTokens(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	AnswerPoll(ctx context.Context, id domain.GameID, player domain.PlayerID, yes bool) error
	// PickSeats is the commander's choice in the pick step: one seat, or two for the roles of mission 50.
	PickSeats(ctx context.Context, id domain.GameID, player domain.PlayerID, seats []int) error
	// PickTask takes a task in the open distribution (the commander acts for the dummy).
	PickTask(ctx context.Context, id domain.GameID, player domain.PlayerID, taskID int) error
	SendDistress(ctx context.Context, id domain.GameID, player domain.PlayerID, dir domain.Direction) error
	SkipDistress(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	ChoosePassCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) error
	Handover(ctx context.Context, id domain.GameID, player domain.PlayerID, taskID, toSeat int) error
	// Communicate: pos is ignored in a dead zone.
	Communicate(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card, pos domain.Position) error
	// PlayCard plays for the seat whose turn it is (the commander plays for the dummy).
	PlayCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) error
	GiveUp(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// NextAttempt deals a new attempt of the same mission after a loss (host, debrief).
	NextAttempt(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// NextMission starts the next mission after a win (host, debrief).
	NextMission(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	BackToLobby(ctx context.Context, id domain.GameID, player domain.PlayerID) error
	// EndGame lets the host close the session for everyone (phase "ended").
	EndGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error
}
