package domain

import (
	"slices"
	"time"
)

// Phase is the lifecycle phase of a game.
type Phase string

const (
	PhaseLobby   Phase = "lobby"
	PhasePlaying Phase = "playing" // an attempt at a mission is running
	PhaseDebrief Phase = "debrief" // an attempt is over; the session continues from here
	PhaseEnded   Phase = "ended"   // closed by the host
)

// Table sizes. With two humans, the dummy (JARVIS) takes the third seat.
const (
	MinPlayers = 2
	MaxPlayers = 5
)

// Player is a seated human player.
type Player struct {
	ID   PlayerID
	Name string // trimmed, 1..20 runes, no control chars, unique case-insensitively
}

// Game is the aggregate root. Fields are exported for the persistence mapper and
// tests; code outside this package must mutate a Game only through its methods.
// Derived values are computed, never stored.
type Game struct {
	ID        GameID
	Phase     Phase
	HostID    PlayerID
	Players   []Player  // humans; seat order = join order
	Revision  int64     // 1 after NewGame; +1 per successful mutator
	CreatedAt time.Time // UTC, second precision
	events    []Event   // pending, not persisted
}

// Clone returns a deep copy, including pending events. Nil and empty slices are preserved.
func (g *Game) Clone() *Game {
	if g == nil {
		return nil
	}
	c := *g
	c.Players = slices.Clone(g.Players)
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
