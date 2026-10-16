// Package apidto is the JSON wire contract shared by the REST (httpapi) and
// WebSocket (ws) adapters: tagged DTOs, mappers from domain types, and the
// error mapping. All JSON is camelCase and slices are never null, except where
// a field says "null" (hidden or not yet known). Seats are ints; the dummy
// (JARVIS) has playerId "" and dummy true.
// Keep in sync with frontend/src/lib/types.ts.
package apidto

import "github.com/martin-pouls/shared/backend/wire"

// Card is a playing card. Suit is "pink" | "blue" | "green" | "yellow" | "trump"; rank is 1..9 for
// colours and 1..4 for trump.
type Card struct {
	Suit string `json:"suit"`
	Rank int    `json:"rank"`
}

// PlayerSummary is public information about a seated player.
type PlayerSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Seat   int    `json:"seat"`
	IsHost bool   `json:"isHost"`
	IsYou  bool   `json:"isYou"`
}

// LogEntry is the logbook line of one mission.
type LogEntry struct {
	Mission   int  `json:"mission"`
	Attempts  int  `json:"attempts"`
	Distress  bool `json:"distress"`
	Completed bool `json:"completed"`
	Total     int  `json:"total"` // attempts, plus one with the distress signal
}

// Objective is a task or special objective. Only the fields its kind uses are meaningful; the
// others are zero. Owner is a seat, -1 (the whole crew) or -2 (not assigned yet).
type Objective struct {
	Kind         string `json:"kind"`
	Owner        int    `json:"owner"`
	Cards        []Card `json:"cards"`  // never null
	Count        int    `json:"count"`  //
	Tricks       []int  `json:"tricks"` // never null; 1-based trick numbers, -1 = the last trick
	Suit         string `json:"suit"`
	Rank         int    `json:"rank"`
	NotWithTrump bool   `json:"notWithTrump"`
	Ordered      bool   `json:"ordered"`
}

// ObjectiveSpec is a special objective of a mission: a template plus the rule for its owner.
type ObjectiveSpec struct {
	Objective
	OwnerRule string `json:"ownerRule"`           // "crew" | "commander" | "chosen" | "roleA" | "roleB" | "leftOfHolder"
	Holder    *Card  `json:"holder"`              // leftOfHolder: the card whose holder's left neighbour owns it, else null
	TaskToken string `json:"taskToken,omitempty"` // the objective's card is the card of the task with this token
}

// MissionDef is one row of the mission table.
type MissionDef struct {
	Number       int             `json:"number"`
	Tasks        int             `json:"tasks"`
	Tokens       []string        `json:"tokens"`       // never null; placement order
	Distribution string          `json:"distribution"` // "" | "open" | "commanderDecision" | "commanderDistribution"
	DeadZone     bool            `json:"deadZone"`
	FromTrick    int             `json:"fromTrick"` // disruption: nobody communicates before this trick; 0 = no restriction
	Golden       bool            `json:"golden"`
	Setup        []string        `json:"setup"` // never null; "swapTokens" | "moveToken" | "silenceMember" | "drawAfterFirstTrick"
	Poll         string          `json:"poll"`  // "" | "yesNo" | "wellIll"
	Objectives   []ObjectiveSpec `json:"objectives"`
}

// MissionCatalogue is the body of GET /api/missions.
type MissionCatalogue struct {
	Edition  string       `json:"edition"`
	Missions []MissionDef `json:"missions"` // in mission order; never null
}

// Communication is a communicated card: Position is "highest" | "only" | "lowest", or "" in a dead zone.
type Communication struct {
	Card     Card   `json:"card"`
	Position string `json:"position"`
	Trick    int    `json:"trick"` // the trick it was made before
}

// Stack is a stack of the dummy: the face-up card (null once played) and whether a hidden card
// lies under it. The hidden card itself is sent only in the debrief, in Member.Cards.
type Stack struct {
	Up     *Card `json:"up"`
	Hidden bool  `json:"hidden"`
}

// Member is a seat in the attempt.
type Member struct {
	Seat          int            `json:"seat"`
	PlayerID      string         `json:"playerId"` // "" for the dummy
	Name          string         `json:"name"`     // "" for the dummy
	Dummy         bool           `json:"dummy"`
	CardCount     int            `json:"cardCount"`
	TricksWon     int            `json:"tricksWon"`
	Communication *Communication `json:"communication"` // while the card is still held, else null
	Communicated  bool           `json:"communicated"`  // once per attempt
	Stacks        []Stack        `json:"stacks"`        // the dummy's stacks; [] for a human
	MustAct       bool           `json:"mustAct"`
	Cards         []Card         `json:"cards"` // null until the debrief; then every card still held
}

// Task is a task card. Card is null while the task is face down.
type Task struct {
	ID       int    `json:"id"`
	Card     *Card  `json:"card"`
	Token    string `json:"token"` // "" | "1".."5" | "omega" | "arrow1".."arrow4"
	Owner    int    `json:"owner"` // seat, or -2 while not assigned
	Status   string `json:"status"`
	FaceDown bool   `json:"faceDown"`
}

// ObjectiveView is a special objective with its status ("open" | "fulfilled" | "failed").
type ObjectiveView struct {
	Objective
	Status string `json:"status"`
}

// Poll is the open poll. Answers has one entry per seat: null = not answered or not asked.
type Poll struct {
	Kind      string  `json:"kind"` // "yesNo" | "wellIll"
	Answers   []*bool `json:"answers"`
	Objective int     `json:"objective"` // the objective whose owner is chosen, else -1
	Task      int     `json:"task"`      // commander's distribution: the revealed task, else -1
}

// Pass is the distress pass in progress.
type Pass struct {
	Direction string `json:"direction"` // "left" | "right"
	Chosen    *Card  `json:"chosen"`    // the viewer's own choice, else null
	Waiting   []int  `json:"waiting"`   // seats that have not chosen yet; never null
}

// Exchange is the viewer's side of the last pass or draw.
type Exchange struct {
	Kind string `json:"kind"` // "pass" | "draw"
	Gave *Card  `json:"gave"`
	Got  *Card  `json:"got"`
}

// CommOption is a card the viewer may communicate, with the true statement ("" in a dead zone).
type CommOption struct {
	Card     Card   `json:"card"`
	Position string `json:"position"`
}

// Actions are what the viewer may do now, besides the step's action.
type Actions struct {
	Communicate  []CommOption `json:"communicate"` // never null; empty unless the viewer may communicate now
	SendDistress bool         `json:"sendDistress"`
	Handover     bool         `json:"handover"`
}

// Play is one card played to a trick.
type Play struct {
	Seat int  `json:"seat"`
	Card Card `json:"card"`
}

// CompletedTrick is the last completed trick.
type CompletedTrick struct {
	Leader int    `json:"leader"`
	Plays  []Play `json:"plays"`
	Winner int    `json:"winner"`
}

// AttemptResult is how an attempt ended. Task, Objective and Seat are -1 when not applicable.
type AttemptResult struct {
	Success   bool   `json:"success"`
	Reason    string `json:"reason"` // "completed" | "taskFailed" | "tokenOrder" | "objectiveFailed" | "gaveUp"
	Task      int    `json:"task"`
	Objective int    `json:"objective"`
	Trick     int    `json:"trick"`
	Seat      int    `json:"seat"`
}

// Attempt is the running attempt (playing) or the one just finished (debrief).
type Attempt struct {
	Number      int             `json:"number"`
	Seat        int             `json:"seat"` // the viewer's seat
	Commander   int             `json:"commander"`
	Step        string          `json:"step"` // "adjustTokens" | "poll" | "pick" | "draft" | "pass" | "play" | "" (debrief)
	Pick        string          `json:"pick"` // step pick: "silence" | "roles" | "objective" | "decision" | "distribute"; else ""
	Actors      []int           `json:"actors"`
	Turn        int             `json:"turn"` // step play: the seat to play; step draft: the seat to take a task; else -1
	TrickNumber int             `json:"trickNumber"`
	TotalTricks int             `json:"totalTricks"`
	Members     []Member        `json:"members"`
	Hand        []Card          `json:"hand"`
	Playable    []Card          `json:"playable"` // the viewer's own, or the dummy's when the viewer acts for it
	Tasks       []Task          `json:"tasks"`
	Objectives  []ObjectiveView `json:"objectives"`
	Silenced    int             `json:"silenced"` // mission 11: the silenced seat, else -1
	Pickable    []int           `json:"pickable"` // step pick, for the commander: the seats they may pick; never null
	Poll        *Poll           `json:"poll"`
	Pass        *Pass           `json:"pass"`
	Exchange    *Exchange       `json:"exchange"`
	Trick       []Play          `json:"trick"` // in play order; never null
	TrickLeader int             `json:"trickLeader"`
	LastTrick   *CompletedTrick `json:"lastTrick"`
	Distress    bool            `json:"distress"`
	Actions     Actions         `json:"actions"`
	Result      *AttemptResult  `json:"result"`
}

// GameOptions are the host-selectable lobby settings.
type GameOptions struct {
	Distress bool `json:"distress"` // the distress signal (rules §9) is on; off by default
}

// PlayerView is the per-player projection.
type PlayerView struct {
	GameID     string          `json:"gameId"`
	Revision   int64           `json:"revision"`
	Phase      string          `json:"phase"` // "lobby" | "playing" | "debrief" | "ended"
	You        string          `json:"you"`
	HostID     string          `json:"hostId"`
	Players    []PlayerSummary `json:"players"` // never null
	Edition    string          `json:"edition"`
	Options    GameOptions     `json:"options"`
	Mission    int             `json:"mission"`    // selected (lobby) or being played
	MissionDef *MissionDef     `json:"missionDef"` // null if the mission is unknown
	Logbook    []LogEntry      `json:"logbook"`    // never null; sorted by mission
	Attempt    *Attempt        `json:"attempt"`    // null in the lobby
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
	EventPlayerJoined    = "playerJoined"
	EventPlayerLeft      = "playerLeft"
	EventMissionSelected = "missionSelected"
	EventOptionsChanged  = "optionsChanged"
	EventAttemptStarted  = "attemptStarted"
	EventTokensAdjusted  = "tokensAdjusted"
	EventPollAnswered    = "pollAnswered"
	EventPicked          = "picked"
	EventTaskPicked      = "taskPicked"
	EventDistressSignal  = "distressSignal"
	EventPassCardChosen  = "passCardChosen"
	EventCardsPassed     = "cardsPassed"
	EventTaskHandedOver  = "taskHandedOver"
	EventCommunicated    = "communicated"
	EventCardPlayed      = "cardPlayed"
	EventTrickWon        = "trickWon"
	EventCardsDrawn      = "cardsDrawn"
	EventAttemptEnded    = "attemptEnded"
	EventBackInLobby     = "backInLobby"
	EventGameEnded       = "gameEnded"
)

// Event is a flattened union of domain events; Type is one of the Event* constants. Events are
// public: they never carry a hidden card. Pointer fields are set exactly when the event carries
// them, so 0 and false survive omitempty.
type Event struct {
	Type       string `json:"type"`
	PlayerID   string `json:"playerId,omitempty"`
	ByPlayerID string `json:"byPlayerId,omitempty"`
	HostID     string `json:"hostId,omitempty"`
	Name       string `json:"name,omitempty"`
	Mission    *int   `json:"mission,omitempty"`   // missionSelected, attemptStarted
	Number     *int   `json:"number,omitempty"`    // attemptStarted: attempt number; trickWon: trick number
	Commander  *int   `json:"commander,omitempty"` // attemptStarted
	Seat       *int   `json:"seat,omitempty"`      // pollAnswered, taskPicked, passCardChosen, communicated, cardPlayed, trickWon
	Seats      []int  `json:"seats,omitempty"`     // picked
	TaskID     *int   `json:"taskId,omitempty"`    // taskPicked, taskHandedOver
	From       *int   `json:"from,omitempty"`      // tokensAdjusted (-1 = skipped), taskHandedOver
	To         *int   `json:"to,omitempty"`        // tokensAdjusted (-1 = skipped), taskHandedOver
	Yes        *bool  `json:"yes,omitempty"`       // pollAnswered
	Direction  string `json:"direction,omitempty"` // distressSignal, cardsPassed
	Skipped    *bool  `json:"skipped,omitempty"`   // distressSignal
	Card       *Card  `json:"card,omitempty"`      // communicated, cardPlayed
	Position   string `json:"position,omitempty"`  // communicated; omitted in a dead zone
	Success    *bool  `json:"success,omitempty"`   // attemptEnded
	Reason     string `json:"reason,omitempty"`    // attemptEnded
}

// CreateGameRequest is the body of POST /api/games.
type CreateGameRequest struct {
	Name    string       `json:"name"`
	Options *GameOptions `json:"options,omitempty"` // nil: the defaults
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
