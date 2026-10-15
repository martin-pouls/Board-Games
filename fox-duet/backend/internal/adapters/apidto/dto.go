// Package apidto is the JSON wire contract shared by the REST (httpapi) and
// WebSocket (ws) adapters: tagged DTOs, mappers from domain types, and the
// error mapping. All JSON is camelCase and slices are never null.
// Keep in sync with frontend/src/lib/types.ts.
package apidto

import "github.com/martin-pouls/shared/backend/wire"

// Card is a playing card. Movement and ability are derived from the rank on the client too, but
// sent for convenience.
type Card struct {
	Suit     string `json:"suit"`     // "doves" | "roses" | "stars"
	Rank     int    `json:"rank"`     // 1..10
	Movement int    `json:"movement"` // 0..3; ignored on input
	Ability  string `json:"ability"`  // "" | "musician" | "foxes" | "gazelle" | "gift" | "royalHeir"; ignored on input
}

// PlayerSummary is public information about a seated player.
type PlayerSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Seat         int    `json:"seat"`
	End          string `json:"end"` // "left" | "right"
	HandCount    int    `json:"handCount"`
	IsHost       bool   `json:"isHost"`
	IsYou        bool   `json:"isYou"`
	IsDealer     bool   `json:"isDealer"`
	MustAct      bool   `json:"mustAct"`
	RevealedHand []Card `json:"revealedHand,omitempty"` // only when won/lost
}

// Space is one path space with its location.
type Space struct {
	Gems    int  `json:"gems"`
	Plus    bool `json:"plus"`
	Covered bool `json:"covered"`
}

// Board is the forest board. Space indices run from the left end to the right end.
type Board struct {
	Side              string  `json:"side"` // "A" | "B"
	Start             int     `json:"start"`
	Spaces            []Space `json:"spaces"` // never null
	Traveller         int     `json:"traveller"`
	ForestTokens      int     `json:"forestTokens"` // left to place
	ForestTokensTotal int     `json:"forestTokensTotal"`
	GemsLeft          int     `json:"gemsLeft"`
}

// Trick is the trick in progress.
type Trick struct {
	Number   int    `json:"number"`
	LeaderID string `json:"leaderId"`
	Lead     *Card  `json:"lead"`
	Follow   *Card  `json:"follow"`
	WinnerID string `json:"winnerId,omitempty"` // once both cards are on the table
}

// LastTrick is the last completed trick, shown until the next card is led.
type LastTrick struct {
	Number    int    `json:"number"`
	LeaderID  string `json:"leaderId"`
	Lead      Card   `json:"lead"`
	Follow    Card   `json:"follow"`
	WinnerID  string `json:"winnerId"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Delta     int    `json:"delta"`
	Beyond    string `json:"beyond"` // "" | "left" | "right"
	Collected bool   `json:"collected"`
}

// MoveOption is one distinct result the trick winner can choose.
type MoveOption struct {
	Delta  int    `json:"delta"` // negative = toward the left end
	To     int    `json:"to"`
	Beyond string `json:"beyond"` // "" | "left" | "right"
	Gem    bool   `json:"gem"`
	Lost   bool   `json:"lost"`
}

// Pending is the decision the game waits for.
type Pending struct {
	Step        string       `json:"step"`      // "lead" | "follow" | "foxesChoose" | "foxesSwap" | "gift" | "move" | "forestEnd"
	PlayerIDs   []string     `json:"playerIds"` // who must act; never null
	ChosenID    string       `json:"chosenId,omitempty"`
	MoveOptions []MoveOption `json:"moveOptions"` // never null
	Ends        []string     `json:"ends"`        // never null
	YourGift    *Card        `json:"yourGift"`
}

// Gift is the viewer's side of the last gift exchange.
type Gift struct {
	Gave     Card `json:"gave"`
	Received Card `json:"received"`
}

// Score is the victory score breakdown.
type Score struct {
	Level        int `json:"level"`
	Cards        int `json:"cards"`
	EarlyWin     int `json:"earlyWin"`
	ForestTokens int `json:"forestTokens"`
	Total        int `json:"total"`
}

// Result is the final result of a finished game.
type Result struct {
	Outcome  string `json:"outcome"`          // "won" | "lost"
	Reason   string `json:"reason,omitempty"` // "lostInForest" | "outOfTime" | "gaveUp"
	PlayerID string `json:"playerId,omitempty"`
	Round    int    `json:"round"`
	GemsLeft int    `json:"gemsLeft"`
	Score    *Score `json:"score"` // only on a win
}

// PlayerView is the per-player projection.
type PlayerView struct {
	GameID     string          `json:"gameId"`
	Revision   int64           `json:"revision"`
	Phase      string          `json:"phase"` // "lobby" | "playing" | "won" | "lost" | "ended"
	You        string          `json:"you"`
	HostID     string          `json:"hostId"`
	Level      int             `json:"level"`    // 0 in the lobby
	Players    []PlayerSummary `json:"players"`  // never null
	Hand       []Card          `json:"hand"`     // never null
	Playable   []Card          `json:"playable"` // never null
	Board      *Board          `json:"board"`
	Round      int             `json:"round"`
	Decree     *Card           `json:"decree"`
	Trick      *Trick          `json:"trick"`
	LastTrick  *LastTrick      `json:"lastTrick"`
	Pending    *Pending        `json:"pending"`
	LastGift   *Gift           `json:"lastGift"`
	Result     *Result         `json:"result"`
	NextGameID string          `json:"nextGameId,omitempty"`
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
}

// Event type strings.
const (
	EventPlayerJoined      = "playerJoined"
	EventPlayerLeft        = "playerLeft"
	EventGameStarted       = "gameStarted"
	EventRoundStarted      = "roundStarted"
	EventCardPlayed        = "cardPlayed"
	EventFoxesChosen       = "foxesChosen"
	EventDecreeSwapped     = "decreeSwapped"
	EventDecreeKept        = "decreeKept"
	EventGiftChosen        = "giftChosen"
	EventGiftExchanged     = "giftExchanged"
	EventTrickWon          = "trickWon"
	EventTravellerMoved    = "travellerMoved"
	EventForestTokenPlaced = "forestTokenPlaced"
	EventGemCollected      = "gemCollected"
	EventGemsAdded         = "gemsAdded"
	EventGameWon           = "gameWon"
	EventGameLost          = "gameLost"
	EventGameEnded         = "gameEnded"
	EventRematchCreated    = "rematchCreated"
)

// Event is a flattened union of domain events; Type is one of the Event* constants.
// Pointer fields are set exactly when the event carries them, so 0 and false survive omitempty.
type Event struct {
	Type       string `json:"type"`
	PlayerID   string `json:"playerId,omitempty"`
	ByPlayerID string `json:"byPlayerId,omitempty"`
	HostID     string `json:"hostId,omitempty"`
	DealerID   string `json:"dealerId,omitempty"`
	Name       string `json:"name,omitempty"`
	GameID     string `json:"gameId,omitempty"`
	Level      *int   `json:"level,omitempty"`     // gameStarted
	Round      *int   `json:"round,omitempty"`     // roundStarted
	Number     *int   `json:"number,omitempty"`    // trickWon
	Card       *Card  `json:"card,omitempty"`      // cardPlayed; roundStarted: the decree
	Old        *Card  `json:"old,omitempty"`       // decreeSwapped
	New        *Card  `json:"new,omitempty"`       // decreeSwapped
	Lead       *bool  `json:"lead,omitempty"`      // cardPlayed
	Skipped    *bool  `json:"skipped,omitempty"`   // foxesChosen
	Auto       *bool  `json:"auto,omitempty"`      // giftChosen
	From       *int   `json:"from,omitempty"`      // travellerMoved
	To         *int   `json:"to,omitempty"`        // travellerMoved
	Delta      *int   `json:"delta,omitempty"`     // travellerMoved
	End        string `json:"end,omitempty"`       // travellerMoved (beyond), forestTokenPlaced
	Space      *int   `json:"space,omitempty"`     // forestTokenPlaced, gemCollected
	GemsMoved  *int   `json:"gemsMoved,omitempty"` // forestTokenPlaced
	RoundEnd   *bool  `json:"roundEnd,omitempty"`  // forestTokenPlaced
	Left       *int   `json:"left,omitempty"`      // forestTokenPlaced: tokens left; gemCollected: gems left
	Count      *int   `json:"count,omitempty"`     // gemsAdded
	Score      *int   `json:"score,omitempty"`     // gameWon
	Reason     string `json:"reason,omitempty"`    // gameLost
}

// CreateGameRequest is the body of POST /api/games.
type CreateGameRequest struct {
	Name string `json:"name"`
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
