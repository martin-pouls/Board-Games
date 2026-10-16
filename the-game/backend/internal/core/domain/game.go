package domain

import (
	"maps"
	"slices"
	"time"
)

// Phase is the lifecycle phase of a game.
type Phase string

const (
	PhaseLobby               Phase = "lobby"
	PhaseChoosingFirstPlayer Phase = "choosingFirstPlayer"
	PhasePlaying             Phase = "playing"
	PhaseWon                 Phase = "won"   // every card played; score 0
	PhaseLost                Phase = "lost"  // stuck or given up; score ≥ 0 (0 only when the last card was an uncovered Skull)
	PhaseEnded               Phase = "ended" // closed by the host in the lobby, before it started
)

// LossReason says why a game was lost.
type LossReason string

const (
	LossStuck  LossReason = "stuck"  // the active player could neither play nor end the turn at its start
	LossGaveUp LossReason = "gaveUp" // a player gave up
)

// Player is a seated player.
type Player struct {
	ID   PlayerID
	Name string // trimmed, 1..20 runes, no control chars, unique case-insensitively within the game
	Hand []Card // kept sorted ascending by Value
}

// Play is one card played in the current turn.
type Play struct {
	Card Card
	Pile int
}

// Turn is the explicit state of the current turn.
type Turn struct {
	Number        int    // 1-based; 0 before the first turn
	ActiveSeat    int    // index into Game.Players
	Plays         []Play // cards played this turn, in order
	StartHandSize int    // hand size at turn start (Extreme 3! hook)
}

// PilesTouched returns the distinct pile indices played on this turn, in first-touched order.
// It returns nil when no card has been played.
func (t Turn) PilesTouched() []int {
	var out []int
	for _, p := range t.Plays {
		if !slices.Contains(out, p.Pile) {
			out = append(out, p.Pile)
		}
	}
	return out
}

// Result is the final result of a game that was won or lost (the phase says which).
type Result struct {
	Score               int // CardsLeftInHands + CardsLeftInDrawPile
	CardsLeftInHands    int
	CardsLeftInDrawPile int
	Reason              LossReason // "" when won
	PlayerID            PlayerID   // the stuck player or the one who gave up; "" when won
}

// Move is a legal (card value, pile index) pair.
type Move struct{ Card, Pile int }

// Game is the aggregate root. Fields are exported for the persistence mapper and
// tests; code outside this package must mutate a Game only through its methods.
type Game struct {
	ID         GameID
	Options    Options
	Rules      Ruleset // re-derived on SetOptions in lobby; frozen from Start on
	Phase      Phase
	HostID     PlayerID
	Players    []Player // seat order = join order = clockwise
	DrawPile   []Card   // last element = top
	Piles      []Pile
	Turn       Turn
	Result     *Result   // non-nil iff Phase is won or lost
	NextGameID GameID    // set when a rematch was created; empty otherwise
	Revision   int64     // starts at 1 in NewGame; +1 exactly once per successful mutating method
	CreatedAt  time.Time // UTC, second precision; zero for games persisted before it was tracked
	events     []Event   // pending, not persisted
}

// Clone returns a deep copy, including pending events. Nil and empty slices are preserved.
func (g *Game) Clone() *Game {
	if g == nil {
		return nil
	}
	c := *g
	c.Rules.HandSizes = maps.Clone(g.Rules.HandSizes)
	if g.Players != nil {
		c.Players = make([]Player, len(g.Players))
		for i, p := range g.Players {
			p.Hand = slices.Clone(p.Hand)
			c.Players[i] = p
		}
	}
	c.DrawPile = slices.Clone(g.DrawPile)
	if g.Piles != nil {
		c.Piles = make([]Pile, len(g.Piles))
		for i, p := range g.Piles {
			p.Cards = slices.Clone(p.Cards)
			c.Piles[i] = p
		}
	}
	c.Turn.Plays = slices.Clone(g.Turn.Plays)
	if g.Result != nil {
		r := *g.Result
		c.Result = &r
	}
	// Event values contain no references, so a shallow slice copy is a deep copy.
	c.events = slices.Clone(g.events)
	return &c
}

// PullEvents returns the pending events and clears them.
func (g *Game) PullEvents() []Event {
	evs := g.events
	g.events = nil
	return evs
}

// Score returns the cards left in all hands and in the draw pile.
func (g *Game) Score() (hands, drawPile int) {
	for _, p := range g.Players {
		hands += len(p.Hand)
	}
	return hands, len(g.DrawPile)
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

// IsOver reports whether the game was won or lost.
func (g *Game) IsOver() bool { return g.Phase == PhaseWon || g.Phase == PhaseLost }

// IsHost reports whether id is the host.
func (g *Game) IsHost(id PlayerID) bool {
	return id != "" && id == g.HostID
}

// CommunicationRestricted reports whether chat and hints must be disabled: while playing
// and a "No communication" card is visible. It is false in every other phase (Q6).
func (g *Game) CommunicationRestricted() bool {
	return g.Phase == PhasePlaying && g.commands().noCommunication
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
		MaxPlayers:  g.Rules.MaxPlayers,
		Joinable:    g.Phase == PhaseLobby && n < g.Rules.MaxPlayers,
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
