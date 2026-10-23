package domain

import (
	"fmt"
	"slices"
)

// Step is the pending decision while playing (plan §5.3). Rounds live inside PhasePlaying.
type Step string

const (
	StepNone       Step = ""           // not playing
	StepDiscussion Step = "discussion" // both players press Ready, then the dice are rolled (§5.1)
	StepPlace      Step = "place"      // the turn player places one die (§5.2)
	StepReroll     Step = "reroll"     // both players pick dice to reroll and confirm (§5.3)
	StepSwap       Step = "swap"       // the partner answers Working Together (§13)
	StepSync       Step = "sync"       // the Co-Pilot places the Synchronisation die (§13)
)

// HandSize is the number of dice per player and round.
const HandSize = 4

// Round is the state of the current round; it is replaced at every round start.
type Round struct {
	Step  Step
	Turn  Role    // the role placing next (StepPlace); the first player at round start
	Ready [2]bool // StepDiscussion, by role
	// Dice holds each role's unplaced dice by hand slot; 0 = not in the hand (not rolled yet,
	// placed or discarded). Slots are stable: a placed die leaves its slot empty.
	Dice            [2][HandSize]int
	Discarded       [2]int      // dice discarded this round, by role (§5.2)
	Placements      []Placement // in placement order
	Speed           *Speed      // set when the second engine piece is placed
	RerollConfirmed [2]bool     // StepReroll, by role
	Swap            *SwapOffer  // StepSwap: the Working Together die on the card
	SyncValue       int         // StepSync: the traffic die waiting for the Co-Pilot
	// Once-per-round abilities (§13).
	AnticipationUsed, TogetherUsed, SyncUsed bool
	Start                                    RoundStart
}

// clone returns a deep copy.
func (r Round) clone() Round {
	c := r
	c.Placements = slices.Clone(r.Placements)
	if r.Speed != nil {
		s := *r.Speed
		c.Speed = &s
	}
	if r.Swap != nil {
		s := *r.Swap
		c.Swap = &s
	}
	c.Start.Traffic = slices.Clone(r.Start.Traffic)
	return c
}

// Placement is a piece placed this round. Role is who placed it (the trainer for an intern
// token, the Co-Pilot for the Synchronisation die). Slot is the hand slot of a player's die, -1
// for an intern token or the traffic die. Value is the final value, after Coffee (the net
// modifier spent).
type Placement struct {
	Space  SpaceID
	Piece  PieceKind
	Value  int
	Coffee int
	Role   Role
	Slot   int
}

// Speed is the result of the second engine piece (§6.2): the dice Sum plus Wind. Move is the
// number of approach spaces (0..2) in rounds 1-6; unused in the final round.
type Speed struct {
	Sum, Wind, Move int
}

// Total returns the speed including wind.
func (s Speed) Total() int { return s.Sum + s.Wind }

// SwapOffer is the die the Working Together initiator put on the card (§13).
type SwapOffer struct {
	By   Role
	Slot int
}

// RoundStart records what happened before the discussion (§5.1): the reroll pickup and the
// traffic rolls (§11.1).
type RoundStart struct {
	RerollPicked bool
	Traffic      []TrafficRoll
}

// TrafficRoll is one traffic-die roll at round start: the face, the approach space it points
// at (capped at the airport), and whether a token was added (false when the supply was empty).
type TrafficRoll struct {
	Roll, Space int
	Added       bool
}

// startRound sets up the round g.Flight.Round (§5.1 steps 1-2): the reroll pickup, the traffic
// rolls of the current approach space, then the discussion. RNG use: one IntN(6) per traffic
// icon. Emits RoundStarted.
func (g *Game) startRound(rng RNG) {
	f := g.Flight
	row := AltitudeRow(f.Round)
	f.R = Round{Step: StepDiscussion, Turn: row.First}
	if row.Reroll(g.ScenarioDef().Colour) {
		f.Rerolls++
		f.R.Start.RerollPicked = true
	}
	for range g.track().Spaces[f.Position-1].Traffic {
		roll := rollTraffic(rng)
		space := min(f.Position+roll-1, len(f.Planes))
		added := f.PlaneSupply() > 0
		if added {
			f.Planes[space-1]++
		}
		f.R.Start.Traffic = append(f.R.Start.Traffic, TrafficRoll{Roll: roll, Space: space, Added: added})
	}
	g.emit(RoundStarted{
		Round:        f.Round,
		RerollPicked: f.R.Start.RerollPicked,
		Traffic:      slices.Clone(f.R.Start.Traffic),
	})
}

// occupant returns the piece on space this round.
func (r *Round) occupant(space SpaceID) (Placement, bool) {
	for _, p := range r.Placements {
		if p.Space == space {
			return p, true
		}
	}
	return Placement{}, false
}

// diceLeft returns the number of unplaced dice of role.
func (r *Round) diceLeft(role Role) int {
	n := 0
	for _, v := range r.Dice[role] {
		if v != 0 {
			n++
		}
	}
	return n
}

// handValues returns the values of role's unplaced dice in slot order.
func (r *Round) handValues(role Role) []int {
	out := []int{}
	for _, v := range r.Dice[role] {
		if v != 0 {
			out = append(out, v)
		}
	}
	return out
}

// Ready marks by as ready in the discussion (§5.1). When both are ready the server rolls all 8
// dice and placement starts with the first player.
// RNG use order: Pilot slots 0-3, then Co-Pilot slots 0-3.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrAlreadyReady.
// Emits PlayerReady, and DiceRolled with the second.
func (g *Game) Ready(by PlayerID, rng RNG) error {
	return fmt.Errorf("%w: Ready", ErrNotImplemented)
}
