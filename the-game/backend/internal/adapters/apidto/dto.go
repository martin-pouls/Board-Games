// Package apidto is the JSON wire contract shared by the REST (httpapi) and
// WebSocket (ws) adapters: tagged DTOs, mappers from domain types, and the
// error mapping. All JSON is camelCase and slices are never null.
// Keep in sync with frontend/src/lib/types.ts.
package apidto

// Card is a playable card.
type Card struct {
	Value   int    `json:"value"`
	Kind    string `json:"kind"`              // "number"
	Command string `json:"command,omitempty"` // Phase 2
}

// HandCard is a card in the viewer's hand with the piles it may go on now.
type HandCard struct {
	Card
	LegalPiles []int `json:"legalPiles"`
}

// GameOptions are the host-selectable settings.
type GameOptions struct {
	Variant         string `json:"variant"` // "classic"; empty on input means classic
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
	Index      int    `json:"index"`
	Direction  string `json:"direction"` // "up" | "down"
	StartValue int    `json:"startValue"`
	TopValue   int    `json:"topValue"`
	TopCard    *Card  `json:"topCard"` // null when only the start card shows
	CardCount  int    `json:"cardCount"`
}

// Play is a card played this turn.
type Play struct {
	Card Card `json:"card"`
	Pile int  `json:"pile"`
}

// Turn is the public state of the current turn.
type Turn struct {
	Number          int    `json:"number"`
	ActivePlayerID  string `json:"activePlayerId"`
	CardsPlayed     int    `json:"cardsPlayed"`
	RequiredMinimum int    `json:"requiredMinimum"`
	CanEndTurn      bool   `json:"canEndTurn"`
	Plays           []Play `json:"plays"`
}

// Result is the final result of a finished game.
type Result struct {
	Outcome             string `json:"outcome"` // "won" | "lost"
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
}

// GameSummary is public, pre-join information about a game.
type GameSummary struct {
	GameID      string `json:"gameId"`
	Phase       string `json:"phase"`
	HostName    string `json:"hostName"`
	PlayerCount int    `json:"playerCount"`
	MaxPlayers  int    `json:"maxPlayers"`
	Joinable    bool   `json:"joinable"`
}

// Event type strings.
const (
	EventPlayerJoined      = "playerJoined"
	EventOptionsChanged    = "optionsChanged"
	EventGameStarted       = "gameStarted"
	EventFirstPlayerChosen = "firstPlayerChosen"
	EventTurnStarted       = "turnStarted"
	EventCardPlayed        = "cardPlayed"
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
	Auto          bool   `json:"auto,omitempty"`
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

// ChooseFirstPlayerRequest is the body of POST /api/games/{code}/first-player.
type ChooseFirstPlayerRequest struct {
	PlayerID string `json:"playerId"`
}

// PlayCardRequest is the body of POST /api/games/{code}/plays. A nil field means bad_request.
type PlayCardRequest struct {
	Card *int `json:"card"`
	Pile *int `json:"pile"`
}

// SessionResponse is returned when creating or joining a game.
type SessionResponse struct {
	GameID   string     `json:"gameId"`
	PlayerID string     `json:"playerId"`
	Token    string     `json:"token"`
	View     PlayerView `json:"view"`
}

// ErrorBody is the error payload used on REST and WS.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the body of every non-2xx REST response.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}
