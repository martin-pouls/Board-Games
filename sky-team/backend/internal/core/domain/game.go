package domain

import (
	"fmt"
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

// Players is the number of seats (a game has exactly two, §14).
const Players = 2

// Player is a seated player. The role follows from the seat (Game.RoleOf).
type Player struct {
	ID   PlayerID
	Name string // trimmed, 1..20 runes, no control chars, unique case-insensitively
}

// Game is the aggregate root. Fields are exported for the persistence mapper and
// tests; code outside this package must mutate a Game only through its methods.
// Occupancy, thresholds, legal spaces and actors are derived, never stored (flight.go).
type Game struct {
	ID         GameID
	Phase      Phase
	HostID     PlayerID
	Players    []Player    // seat order = join order
	PilotSeat  int         // 0 or 1; the other seat is the Co-Pilot (SwapRoles)
	Scenario   ScenarioID  // DefaultScenario in a new lobby
	Abilities  []AbilityID // the chosen cards, in pick order; never nil once created
	Flight     *Flight     // nil in the lobby
	Result     *Result     // set iff Phase is won or lost
	NextGameID GameID      // set when a rematch was created; empty otherwise
	Revision   int64       // 1 after NewGame; +1 per successful mutator
	CreatedAt  time.Time   // UTC, second precision
	events     []Event     // pending, not persisted
}

// Clone returns a deep copy, including pending events. Nil and empty slices are preserved.
func (g *Game) Clone() *Game {
	if g == nil {
		return nil
	}
	c := *g
	c.Players = slices.Clone(g.Players)
	c.Abilities = slices.Clone(g.Abilities)
	c.Flight = g.Flight.clone()
	c.Result = g.Result.clone()
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

// RoleOf returns the role of seat.
func (g *Game) RoleOf(seat int) Role {
	if seat == g.PilotSeat {
		return RolePilot
	}
	return RoleCopilot
}

// SeatOf returns the seat of role.
func (g *Game) SeatOf(r Role) int {
	if r == RolePilot {
		return g.PilotSeat
	}
	return 1 - g.PilotSeat
}

// PlayerOf returns the ID of the player in role (empty when that seat is free).
func (g *Game) PlayerOf(r Role) PlayerID {
	if s := g.SeatOf(r); s < len(g.Players) {
		return g.Players[s].ID
	}
	return ""
}

// ScenarioDef returns the game's scenario (zero if the ID is unknown, which a valid game never has).
func (g *Game) ScenarioDef() ScenarioDef {
	s, _ := ScenarioFor(g.Scenario)
	return s
}

// track returns the game's approach track.
func (g *Game) track() TrackDef {
	t, _ := TrackFor(g.ScenarioDef().Track)
	return t
}

// hasAbility reports whether card a was chosen for this game.
func (g *Game) hasAbility(a AbilityID) bool {
	return slices.Contains(g.Abilities, a)
}

// Actors returns the roles that must act now (§5.3 of the plan): each role not yet ready in the
// discussion, the turn player while placing, each role that has not confirmed a reroll, the
// partner of the Working Together initiator, the Co-Pilot with the Synchronisation die. Nil when
// not playing.
func (g *Game) Actors() []Role {
	if g.Phase != PhasePlaying || g.Flight == nil {
		return nil
	}
	r := &g.Flight.R
	switch r.Step {
	case StepDiscussion:
		return rolesWithout(r.Ready)
	case StepPlace:
		return []Role{r.Turn}
	case StepReroll:
		return rolesWithout(r.RerollConfirmed)
	case StepSwap:
		if r.Swap != nil {
			return []Role{r.Swap.By.Other()}
		}
	case StepSync:
		return []Role{RoleCopilot}
	}
	return nil
}

// rolesWithout returns the roles whose flag is false, in index order.
func rolesWithout(done [2]bool) []Role {
	var out []Role
	for _, r := range Roles {
		if !done[r] {
			out = append(out, r)
		}
	}
	return out
}

// requireActor checks that by is seated, the game is playing in one of steps, and by must act.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn.
func (g *Game) requireActor(by PlayerID, steps ...Step) (Role, error) {
	role, err := g.requirePlaying(by, steps...)
	if err != nil {
		return 0, err
	}
	if !slices.Contains(g.Actors(), role) {
		return 0, fmt.Errorf("%w: %s", ErrNotYourTurn, g.Flight.R.Step)
	}
	return role, nil
}

// requirePlaying checks that by is seated and the game is playing in one of steps, without
// requiring by to be an actor (side intents, §13).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep.
func (g *Game) requirePlaying(by PlayerID, steps ...Step) (Role, error) {
	seat, err := g.requireSeated(by)
	if err != nil {
		return 0, err
	}
	if g.Phase != PhasePlaying {
		return 0, fmt.Errorf("%w: the game is not running", ErrWrongPhase)
	}
	if !slices.Contains(steps, g.Flight.R.Step) {
		return 0, fmt.Errorf("%w: the game waits for %s", ErrWrongStep, g.Flight.R.Step)
	}
	return g.RoleOf(seat), nil
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
		MaxPlayers:  Players,
		Joinable:    g.Phase == PhaseLobby && n < Players,
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
