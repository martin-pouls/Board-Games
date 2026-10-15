// Package apidto is the JSON wire contract shared by the REST (httpapi) and
// WebSocket (ws) adapters: tagged DTOs, mappers from domain types, and the
// error mapping. All JSON is camelCase and slices are never null.
// Keep in sync with frontend/src/lib/types.ts.
package apidto

import "github.com/martin-pouls/shared/backend/wire"

// Card is a playable card.
type Card struct {
	Value   int    `json:"value"`
	Kind    string `json:"kind"`              // "number"
	Command string `json:"command,omitempty"` // stop|skull|exactlyThree|noCommunication|noBackwardsTrick|singlePile|drawOnlyOne
}

// HandCard is a card in the viewer's hand with the piles it may go on now.
type HandCard struct {
	Card
	LegalPiles []int  `json:"legalPiles"`
	Blocked    string `json:"blocked,omitempty"` // error code of the command that blocks this card (only when legalPiles is empty)
	// Vetoes: piles the base rules allow but a command forbids, with the error code.
	Vetoes []PileVeto `json:"vetoes,omitempty"`
}

// PileVeto is a command veto of a placement on Pile.
type PileVeto struct {
	Pile int    `json:"pile"`
	Code string `json:"code"`
}

// GameOptions are the host-selectable settings.
type GameOptions struct {
	Variant         string `json:"variant"` // "classic" | "extreme"; empty on input means classic
	Pro             bool   `json:"pro"`
	ReducedHandSize bool   `json:"reducedHandSize"`
}

// Rules is the public subset of the ruleset.
type Rules struct {
	MinCardsPerTurn       int `json:"minCardsPerTurn"`
	MinCardsDrawPileEmpty int `json:"minCardsDrawPileEmpty"`
	HandSize              int `json:"handSize"`
	MinPlayers            int `json:"minPlayers"`
	MaxPlayers            int `json:"maxPlayers"`
}

// PlayerSummary is public information about a seated player.
type PlayerSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Seat         int    `json:"seat"`
	HandCount    int    `json:"handCount"`
	IsHost       bool   `json:"isHost"`
	IsYou        bool   `json:"isYou"`
	IsActive     bool   `json:"isActive"`
	RevealedHand []Card `json:"revealedHand,omitempty"`
}

// Pile is the public state of a pile.
type Pile struct {
	Index         int    `json:"index"`
	Direction     string `json:"direction"` // "up" | "down"
	StartValue    int    `json:"startValue"`
	TopValue      int    `json:"topValue"`
	TopCard       *Card  `json:"topCard"` // null when only the start card shows
	CardCount     int    `json:"cardCount"`
	CommandActive bool   `json:"commandActive"` // topCard.command is in effect now
}

// ActiveCommand is a visible Infinity command and the pile it lies on.
type ActiveCommand struct {
	Command string `json:"command"`
	Pile    int    `json:"pile"`
}

// TurnConstraints are the command constraints on the current turn.
type TurnConstraints struct {
	ExactCount *int  `json:"exactCount"` // 3 after a 3!, else null
	SinglePile bool  `json:"singlePile"`
	BoundPile  *int  `json:"boundPile"`  // null = not bound
	SkullPiles []int `json:"skullPiles"` // never null
	StopPlayed bool  `json:"stopPlayed"`
}

// Play is a card played this turn.
type Play struct {
	Card Card `json:"card"`
	Pile int  `json:"pile"`
}

// Turn is the public state of the current turn.
type Turn struct {
	Number          int             `json:"number"`
	ActivePlayerID  string          `json:"activePlayerId"`
	CardsPlayed     int             `json:"cardsPlayed"`
	RequiredMinimum int             `json:"requiredMinimum"`
	CanEndTurn      bool            `json:"canEndTurn"`
	Stuck           bool            `json:"stuck"` // no card fits and the turn cannot end: undo or give up
	Plays           []Play          `json:"plays"`
	Constraints     TurnConstraints `json:"constraints"`
}

// Result is the final result of a finished game.
type Result struct {
	Outcome             string `json:"outcome"` // "won" | "lost" | "aborted"
	Score               int    `json:"score"`
	Excellent           bool   `json:"excellent"`
	CardsLeftInHands    int    `json:"cardsLeftInHands"`
	CardsLeftInDrawPile int    `json:"cardsLeftInDrawPile"`
	StuckPlayerID       string `json:"stuckPlayerId,omitempty"`
}

// PlayerView is the per-player projection.
type PlayerView struct {
	GameID                  string          `json:"gameId"`
	Revision                int64           `json:"revision"`
	Phase                   string          `json:"phase"`
	You                     string          `json:"you"`
	HostID                  string          `json:"hostId"`
	Options                 GameOptions     `json:"options"`
	Rules                   Rules           `json:"rules"`
	Players                 []PlayerSummary `json:"players"`
	Hand                    []HandCard      `json:"hand"`
	Piles                   []Pile          `json:"piles"`
	DrawPileCount           int             `json:"drawPileCount"`
	Turn                    *Turn           `json:"turn"`
	CommunicationRestricted bool            `json:"communicationRestricted"`
	Result                  *Result         `json:"result"`
	NextGameID              string          `json:"nextGameId,omitempty"`
	ActiveCommands          []ActiveCommand `json:"activeCommands"` // never null
}

// GameSummary is public, pre-join information about a game.
type GameSummary struct {
	GameID      string `json:"gameId"`
	Phase       string `json:"phase"`
	HostName    string `json:"hostName"`
	PlayerCount int    `json:"playerCount"`
	MaxPlayers  int    `json:"maxPlayers"`
	Joinable    bool   `json:"joinable"`
	CreatedAt   string `json:"createdAt,omitempty"` // RFC3339 UTC; omitted when unknown
	Outcome     string `json:"outcome,omitempty"`   // won/lost/aborted once finished
}

// Event type strings.
const (
	EventPlayerJoined      = "playerJoined"
	EventOptionsChanged    = "optionsChanged"
	EventGameStarted       = "gameStarted"
	EventFirstPlayerChosen = "firstPlayerChosen"
	EventTurnStarted       = "turnStarted"
	EventCardPlayed        = "cardPlayed"
	EventPlayUndone        = "playUndone"
	EventTurnEnded         = "turnEnded"
	EventPlayerSkipped     = "playerSkipped"
	EventGameFinished      = "gameFinished"
	EventRematchCreated    = "rematchCreated"
)

// Event is a flattened union of domain events; Type is one of the Event* constants.
type Event struct {
	Type          string `json:"type"`
	PlayerID      string `json:"playerId,omitempty"`
	Name          string `json:"name,omitempty"`
	Card          *Card  `json:"card,omitempty"`
	Pile          *int   `json:"pile,omitempty"`
	Number        *int   `json:"number,omitempty"`
	CardsDrawn    *int   `json:"cardsDrawn,omitempty"`
	Outcome       string `json:"outcome,omitempty"`
	Score         *int   `json:"score,omitempty"`
	StuckPlayerID string `json:"stuckPlayerId,omitempty"`
	GameID        string `json:"gameId,omitempty"`
}

// CreateGameRequest is the body of POST /api/games.
type CreateGameRequest struct {
	Name    string       `json:"name"`
	Options *GameOptions `json:"options,omitempty"`
}

// JoinGameRequest is the body of POST /api/games/{code}/players.
type JoinGameRequest struct {
	Name string `json:"name"`
}

// SessionResponse is returned when creating or joining a game.
type SessionResponse struct {
	GameID   string     `json:"gameId"`
	PlayerID string     `json:"playerId"`
	Token    string     `json:"token"`
	View     PlayerView `json:"view"`
}

// ErrorBody is the error payload used on REST and WS (shared envelope).
type ErrorBody = wire.ErrorBody

// ErrorResponse is the body of every non-2xx REST response (shared envelope).
type ErrorResponse = wire.ErrorResponse
