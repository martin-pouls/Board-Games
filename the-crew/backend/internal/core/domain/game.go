package domain

import (
	"maps"
	"slices"
	"time"
)

// Phase is the lifecycle phase of a game (a session of one crew).
type Phase string

const (
	PhaseLobby   Phase = "lobby"   // players join; the host picks a mission
	PhasePlaying Phase = "playing" // an attempt is running
	PhaseDebrief Phase = "debrief" // an attempt is over; all hands are revealed
	PhaseEnded   Phase = "ended"   // closed by the host
)

// Crew sizes (rules §1, §16).
const (
	MinPlayers = 2
	MaxPlayers = 5
)

// Player is a human in the session. Join order is seat order.
type Player struct {
	ID   PlayerID
	Name string // trimmed, 1..20 runes, no control chars, unique case-insensitively
}

// LogEntry is the logbook line of one mission (rules §15).
type LogEntry struct {
	Attempts  int  // finished attempts until completion (attempts of a replay are not counted)
	Distress  bool // the distress signal was used
	Completed bool
}

// Total is the logbook count: the attempts, plus one with the distress signal.
func (e LogEntry) Total() int {
	if e.Distress && e.Completed {
		return e.Attempts + 1
	}
	return e.Attempts
}

// Game is the aggregate root. Fields are exported for the persistence mapper and tests; code
// outside this package must mutate a Game only through its methods. Statuses, actors, legal cards
// and the trick winner are derived, never stored.
type Game struct {
	ID        GameID
	Phase     Phase
	HostID    PlayerID
	Players   []Player // seat order = join order
	Edition   Edition
	Mission   int              // the selected mission (lobby) or the one being played
	Logbook   map[int]LogEntry // per mission number; never nil
	Attempt   *Attempt         // nil in the lobby; kept in the debrief for the reveal
	Revision  int64            // 1 after NewGame; +1 per successful mutator
	CreatedAt time.Time        // UTC, second precision
	events    []Event          // pending, not persisted
}

// Clone returns a deep copy, including pending events. Nil and empty slices are preserved.
func (g *Game) Clone() *Game {
	if g == nil {
		return nil
	}
	c := *g
	c.Players = slices.Clone(g.Players)
	c.Logbook = maps.Clone(g.Logbook)
	c.Attempt = g.Attempt.clone()
	// Events are immutable values whose slices are never shared with state, so a shallow slice
	// copy is enough.
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

// MissionDef returns the definition of the selected mission.
func (g *Game) MissionDef() MissionDef {
	m, _ := MissionFor(g.Edition, g.Mission)
	return m
}

// nextUnfinished returns the lowest mission number that is not completed (the last mission if
// all are).
// distressActive reports whether the distress signal is active for the selected mission: it was
// used and the mission is not completed yet (rules §9).
func (g *Game) distressActive() bool {
	e := g.Logbook[g.Mission]
	return e.Distress && !e.Completed
}

func (g *Game) nextUnfinished() int {
	n := Missions(g.Edition)
	for i := 1; i <= n; i++ {
		if !g.Logbook[i].Completed {
			return i
		}
	}
	return n
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
