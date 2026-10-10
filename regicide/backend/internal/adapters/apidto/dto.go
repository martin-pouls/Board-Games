// Package apidto is the JSON wire contract shared by the REST (httpapi) and
// WebSocket (ws) adapters: tagged DTOs, mappers from domain types, and the
// error mapping. All JSON is camelCase and slices are never null.
// Keep in sync with frontend/src/lib/types.ts.
package apidto

// Card is a playing card. Suit "" (and rank 0) is the Jester.
type Card struct {
	Suit string `json:"suit"` // "hearts" | "diamonds" | "clubs" | "spades" | "" (Jester)
	Rank int    `json:"rank"` // 0 Jester, 1 Ace, 2..10, 11 Jack, 12 Queen, 13 King
}

// Rules is the public subset of the ruleset for the current player count.
type Rules struct {
	MinPlayers    int `json:"minPlayers"`
	MaxPlayers    int `json:"maxPlayers"`
	MaxHandSize   int `json:"maxHandSize"` // 0 if there are no players
	TavernJesters int `json:"tavernJesters"`
	SoloJesters   int `json:"soloJesters"`
}

// PlayerSummary is public information about a seated player.
type PlayerSummary struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Seat            int    `json:"seat"`
	HandCount       int    `json:"handCount"`
	IsHost          bool   `json:"isHost"`
	IsYou           bool   `json:"isYou"`
	IsActive        bool   `json:"isActive"`
	YieldedLastTurn bool   `json:"yieldedLastTurn"`
	Signal          string `json:"signal"`                 // "none" | "goodPlay" | "ratherNot"
	RevealedHand    []Card `json:"revealedHand,omitempty"` // only when won/lost
}

// PlayedSet is one play against the current enemy.
type PlayedSet struct {
	PlayerID     string   `json:"playerId"`
	Cards        []Card   `json:"cards"` // never null
	Value        int      `json:"value"`
	Damage       int      `json:"damage"`
	Doubled      bool     `json:"doubled"`
	BlockedSuits []string `json:"blockedSuits"` // never null
	Jester       bool     `json:"jester"`
}

// Enemy is the current enemy with all derived numbers.
type Enemy struct {
	Card              Card        `json:"card"`
	Attack            int         `json:"attack"`
	Health            int         `json:"health"`
	Damage            int         `json:"damage"`
	Shield            int         `json:"shield"`
	EffectiveAttack   int         `json:"effectiveAttack"`
	Immunity          string      `json:"immunity"`          // the enemy's suit
	ImmunityCancelled bool        `json:"immunityCancelled"` // a Jester was played against it
	Plays             []PlayedSet `json:"plays"`             // never null
}

// Turn is the public state of the current turn.
type Turn struct {
	Number         int    `json:"number"`
	ActivePlayerID string `json:"activePlayerId"`
	Step           string `json:"step"` // "play" | "discard" | "chooseNext" | "" (game over)
	CanYield       bool   `json:"canYield"`
	DamageToCover  int    `json:"damageToCover"`
}

// Result is the final result of a finished game.
type Result struct {
	Outcome         string `json:"outcome"`          // "won" | "lost"
	Reason          string `json:"reason,omitempty"` // "cannotCoverDamage" | "cannotAct" | "gaveUp"
	PlayerID        string `json:"playerId,omitempty"`
	Grade           string `json:"grade,omitempty"` // "gold" | "silver" | "bronze" (solo wins)
	EnemiesDefeated int    `json:"enemiesDefeated"`
	SoloJestersUsed int    `json:"soloJestersUsed"`
}

// PlayerView is the per-player projection.
type PlayerView struct {
	GameID          string          `json:"gameId"`
	Revision        int64           `json:"revision"`
	Phase           string          `json:"phase"` // "lobby" | "playing" | "won" | "lost"
	You             string          `json:"you"`
	HostID          string          `json:"hostId"`
	Rules           Rules           `json:"rules"`
	Players         []PlayerSummary `json:"players"` // never null
	Hand            []Card          `json:"hand"`    // never null
	Enemy           *Enemy          `json:"enemy"`
	CastleCount     int             `json:"castleCount"`
	TavernCount     int             `json:"tavernCount"`
	DiscardCount    int             `json:"discardCount"`
	Turn            *Turn           `json:"turn"`
	SoloJestersLeft int             `json:"soloJestersLeft"`
	Result          *Result         `json:"result"`
	NextGameID      string          `json:"nextGameId,omitempty"`
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
	EventPlayerJoined     = "playerJoined"
	EventPlayerLeft       = "playerLeft"
	EventGameStarted      = "gameStarted"
	EventEnemyRevealed    = "enemyRevealed"
	EventTurnStarted      = "turnStarted"
	EventCardsPlayed      = "cardsPlayed"
	EventJesterPlayed     = "jesterPlayed"
	EventYielded          = "yielded"
	EventEnemyDefeated    = "enemyDefeated"
	EventCardsDiscarded   = "cardsDiscarded"
	EventAttackBlocked    = "attackBlocked"
	EventNextPlayerChosen = "nextPlayerChosen"
	EventSignalSet        = "signalSet"
	EventSoloJesterUsed   = "soloJesterUsed"
	EventGameWon          = "gameWon"
	EventGameLost         = "gameLost"
	EventRematchCreated   = "rematchCreated"
)

// Event is a flattened union of domain events; Type is one of the Event* constants.
type Event struct {
	Type         string   `json:"type"`
	PlayerID     string   `json:"playerId,omitempty"`
	ByPlayerID   string   `json:"byPlayerId,omitempty"`
	HostID       string   `json:"hostId,omitempty"`
	Name         string   `json:"name,omitempty"`
	Number       *int     `json:"number,omitempty"`
	Card         *Card    `json:"card,omitempty"`
	Cards        []Card   `json:"cards,omitempty"`
	Value        *int     `json:"value,omitempty"`
	Healed       *int     `json:"healed,omitempty"`
	Drawn        *int     `json:"drawn,omitempty"`
	Damage       *int     `json:"damage,omitempty"`
	Attack       *int     `json:"attack,omitempty"`
	Left         *int     `json:"left,omitempty"`
	BlockedSuits []string `json:"blockedSuits,omitempty"`
	Exact        *bool    `json:"exact,omitempty"`
	Signal       string   `json:"signal,omitempty"`
	Grade        string   `json:"grade,omitempty"`
	Reason       string   `json:"reason,omitempty"`
	GameID       string   `json:"gameId,omitempty"`
}

// CreateGameRequest is the body of POST /api/games.
type CreateGameRequest struct {
	Name string `json:"name"`
}

// JoinGameRequest is the body of POST /api/games/{code}/players.
type JoinGameRequest struct {
	Name string `json:"name"`
}

// StartGameRequest is the body of POST /api/games/{code}/start. An empty first
// player (or an empty body) means random.
type StartGameRequest struct {
	FirstPlayerID string `json:"firstPlayerId"`
}

// CardsRequest is the body of play and discard. A nil Cards means bad_request.
type CardsRequest struct {
	Cards []Card `json:"cards"`
}

// ChooseNextRequest is the body of POST /api/games/{code}/next-player. An empty id means bad_request.
type ChooseNextRequest struct {
	PlayerID string `json:"playerId"`
}

// SignalRequest is the body of PUT /api/games/{code}/signal. A nil signal means bad_request.
type SignalRequest struct {
	Signal *string `json:"signal"`
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
