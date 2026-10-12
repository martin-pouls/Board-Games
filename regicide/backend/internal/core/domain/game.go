package domain

import (
	"slices"
	"time"
)

// Phase is the lifecycle phase of a game.
type Phase string

const (
	PhaseLobby   Phase = "lobby"
	PhasePlaying Phase = "playing"
	PhaseWon     Phase = "won"
	PhaseLost    Phase = "lost"
	PhaseEnded   Phase = "ended" // closed by the host in the lobby, before it started
)

// Step is the step of the current turn while playing (rules §4).
type Step string

const (
	StepNone       Step = ""           // not playing
	StepPlay       Step = "play"       // Step 1: play cards, yield (or use a solo Jester)
	StepDiscard    Step = "discard"    // Step 4: discard to cover the enemy's attack
	StepChooseNext Step = "chooseNext" // after a Jester: the Jester player picks the next player
)

// JesterSignal is the public "who wants to go next" hint while the Jester player chooses (rules §7).
type JesterSignal string

const (
	SignalNone      JesterSignal = ""
	SignalGoodPlay  JesterSignal = "goodPlay"
	SignalRatherNot JesterSignal = "ratherNot"
)

// LossReason says why a game was lost.
type LossReason string

const (
	LossCannotCover LossReason = "cannotCoverDamage" // Step 4: hand total < effective attack
	LossCannotAct   LossReason = "cannotAct"         // Step 1: empty hand and no allowed action
	LossGaveUp      LossReason = "gaveUp"
)

// Grade is the solo victory grade (rules §11).
type Grade string

const (
	GradeNone   Grade = ""
	GradeGold   Grade = "gold"
	GradeSilver Grade = "silver"
	GradeBronze Grade = "bronze"
)

// Player is a seated player.
type Player struct {
	ID              PlayerID
	Name            string       // trimmed, 1..20 runes, no control chars, unique case-insensitively
	Hand            []Card       // always sorted (SortCards); never nil once seated ([]Card{})
	YieldedLastTurn bool         // rules §8 flag
	Signal          JesterSignal // only meaningful while Step == chooseNext
}

// PlayedSet is one play against the current enemy (a Jester play is a 1-card set).
type PlayedSet struct {
	By    PlayerID
	Cards []Card
}

// Turn is the explicit state of the current turn.
type Turn struct {
	Number     int // 1-based; +1 on every beginTurn (incl. same player after a defeat)
	ActiveSeat int // index into Players; in chooseNext: the Jester player (the chooser)
}

// LastTurn is what the most recent turn did, for the table's Played and Discarded piles. The next
// turn replaces it as soon as its player plays or yields, so a defeated enemy's killing play stays
// visible until then.
type LastTurn struct {
	By        PlayerID
	Played    []Card // sorted; nil when the player yielded
	Defeated  *Card  // the enemy this play defeated, if any
	Exact     bool   // the defeat was exact: the enemy went on top of the Tavern
	Discarded []Card // sorted; nil until the player discards (or when no damage was taken)
}

// clone returns a deep copy; nil stays nil.
func (lt *LastTurn) clone() *LastTurn {
	if lt == nil {
		return nil
	}
	c := *lt
	c.Played = slices.Clone(lt.Played)
	c.Discarded = slices.Clone(lt.Discarded)
	if lt.Defeated != nil {
		e := *lt.Defeated
		c.Defeated = &e
	}
	return &c
}

// Result is set iff Phase is won or lost. On a win both fields are empty.
type Result struct {
	Reason   LossReason
	PlayerID PlayerID // who could not cover / could not act / gave up
}

// Game is the aggregate root. Fields are exported for the persistence mapper and
// tests; code outside this package must mutate a Game only through its methods.
// Derived values (shield, damage, effective attack, immunity) are computed from Played
// and Enemy, never stored.
type Game struct {
	ID              GameID
	Phase           Phase
	Step            Step // StepNone unless Phase == playing
	HostID          PlayerID
	Players         []Player    // seat order = join order = clockwise
	Castle          []Card      // facedown enemies, last = top; excludes Enemy
	Enemy           *Card       // current enemy; nil in lobby and after a win
	Tavern          []Card      // last = top; "under the Tavern" = index 0
	DiscardPile     []Card      // order irrelevant (face up)
	Played          []PlayedSet // plays against the current enemy, in order
	LastTurn        *LastTurn   // nil before the first play or yield
	Turn            Turn
	SoloJestersUsed int
	Result          *Result
	NextGameID      GameID    // set when a rematch was created; empty otherwise
	Revision        int64     // 1 after NewGame; +1 per successful mutator
	CreatedAt       time.Time // UTC, second precision
	events          []Event   // pending, not persisted
}

// Clone returns a deep copy, including pending events. Nil and empty slices are preserved.
func (g *Game) Clone() *Game {
	if g == nil {
		return nil
	}
	c := *g
	if g.Players != nil {
		c.Players = make([]Player, len(g.Players))
		for i, p := range g.Players {
			p.Hand = slices.Clone(p.Hand)
			c.Players[i] = p
		}
	}
	c.Castle = slices.Clone(g.Castle)
	if g.Enemy != nil {
		e := *g.Enemy
		c.Enemy = &e
	}
	c.Tavern = slices.Clone(g.Tavern)
	c.DiscardPile = slices.Clone(g.DiscardPile)
	if g.Played != nil {
		c.Played = make([]PlayedSet, len(g.Played))
		for i, ps := range g.Played {
			ps.Cards = slices.Clone(ps.Cards)
			c.Played[i] = ps
		}
	}
	c.LastTurn = g.LastTurn.clone()
	if g.Result != nil {
		r := *g.Result
		c.Result = &r
	}
	// Events are immutable values whose slices are never shared with state,
	// so a shallow slice copy is enough.
	c.events = slices.Clone(g.events)
	return &c
}

// PullEvents returns the pending events and clears them.
func (g *Game) PullEvents() []Event {
	evs := g.events
	g.events = nil
	return evs
}

// PlayerByID returns the player with the given ID, their seat, and whether they were found.
func (g *Game) PlayerByID(id PlayerID) (Player, int, bool) {
	for i, p := range g.Players {
		if p.ID == id {
			return p, i, true
		}
	}
	return Player{}, -1, false
}

// IsHost reports whether id is the host.
func (g *Game) IsHost(id PlayerID) bool {
	return id != "" && id == g.HostID
}

// Rules returns the ruleset for the current player count (zero Ruleset with no players).
// The player count is frozen once the game starts.
func (g *Game) Rules() Ruleset {
	r, _ := RulesFor(len(g.Players))
	return r
}

// IsSolo reports whether exactly one player is seated.
func (g *Game) IsSolo() bool { return len(g.Players) == 1 }

// SoloJestersLeft returns the unused solo Jester powers (0 in multiplayer).
func (g *Game) SoloJestersLeft() int {
	return max(0, g.Rules().SoloJesters-g.SoloJestersUsed)
}

// Summary returns the public, pre-join summary of the game.
func (g *Game) Summary() GameSummary {
	var hostName string
	if host, _, ok := g.PlayerByID(g.HostID); ok {
		hostName = host.Name
	}
	n := len(g.Players)
	return GameSummary{
		GameID:      g.ID,
		Phase:       g.Phase,
		HostName:    hostName,
		PlayerCount: n,
		MaxPlayers:  MaxPlayers,
		Joinable:    g.Phase == PhaseLobby && n < MaxPlayers,
		CreatedAt:   g.CreatedAt,
	}
}

// emit queues a pending domain event.
func (g *Game) emit(e Event) {
	g.events = append(g.events, e)
}

// commit marks a successful mutation: Revision++ exactly once per exported mutator.
func (g *Game) commit() {
	g.Revision++
}
