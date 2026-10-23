package domain

// Test helpers exported for the black-box tests in package domain_test (rules_*_test.go,
// data_expect_test.go). They build states through the public lifecycle and then put the flight
// in place, so the black-box tests need no access to unexported code.

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"
)

// The players of NewFlightForTest: Alice is the host in seat 0 and the Pilot, Bob the Co-Pilot.
const (
	Alice = alice
	Bob   = bob
)

// TestRNG is a scripted RNG: IntN returns the scripted ints in order (each must lie in [0, n)),
// Shuffle leaves everything in order. It panics when the script runs out, so a test notices an
// unexpected roll.
type TestRNG struct {
	ints []int
	next int
}

// ScriptedRNG returns a TestRNG that answers IntN with ints in order. Build them with DieRoll,
// DieRolls and TrafficDie.
func ScriptedRNG(ints ...int) *TestRNG {
	return &TestRNG{ints: slices.Clone(ints)}
}

// IntN returns the next scripted int.
func (r *TestRNG) IntN(n int) int {
	if r.next >= len(r.ints) {
		panic(fmt.Sprintf("ScriptedRNG: no value left for IntN(%d) after %d rolls", n, r.next))
	}
	v := r.ints[r.next]
	if v < 0 || v >= n {
		panic(fmt.Sprintf("ScriptedRNG: value %d out of range for IntN(%d)", v, n))
	}
	r.next++
	return v
}

// Shuffle is the identity.
func (r *TestRNG) Shuffle(int, func(i, j int)) {}

// Left returns the number of scripted ints not used yet.
func (r *TestRNG) Left() int { return len(r.ints) - r.next }

// DieRoll returns the IntN(6) result that rolls a player die of value v (1..6).
func DieRoll(v int) int {
	if v < 1 || v > 6 {
		panic(fmt.Sprintf("DieRoll(%d): want 1..6", v))
	}
	return v - 1
}

// DieRolls maps DieRoll over vs.
func DieRolls(vs ...int) []int {
	out := make([]int, len(vs))
	for i, v := range vs {
		out[i] = DieRoll(v)
	}
	return out
}

// TrafficDie returns the IntN(6) result that rolls face r (2..5) on the traffic die.
func TrafficDie(r int) int {
	if i := slices.Index(trafficFaces[:], r); i >= 0 {
		return i
	}
	panic(fmt.Sprintf("TrafficDie(%d): the traffic die has faces %v", r, trafficFaces))
}

// FlightSetup describes a game in the middle of a round, at step place. Fields left zero take
// the value noted; everything else is put in place literally, without resolving anything:
// placements do not move the axis or deploy switches (set those fields too), and Speed, the
// ready flags of later steps and the ability flags are set on g.Flight directly.
type FlightSetup struct {
	Scenario    ScenarioID  // "" = DefaultScenario
	Abilities   []AbilityID // the scenario's full number
	Round       int         // 0 = 1
	Position    int         // 0 = 1
	Planes      []int       // nil = the track's setup; else one count per approach space
	Axis        int
	Gear        [3]bool
	Flaps       int
	Brakes      int
	Ice         int
	Coffee      int
	Rerolls     int  // the supply, literally (no pickup is applied)
	Fuel        *int // nil = StartFuel with a kerosene track, 0 without
	Wind        int
	Intern      []int            // the token row; nil = 1..6 in order with the intern module, none without
	InternTaken [2]int           // tokens already taken by the Pilot (from the left) and the Co-Pilot (from the right)
	Dice        [2][HandSize]int // unplaced dice by role and slot; 0 = empty slot
	Discarded   [2]int
	Turn        Role
	Placements  []Placement
}

// zeroRNG rolls 0 for every IntN and shuffles nothing.
type zeroRNG struct{}

func (zeroRNG) IntN(int) int                { return 0 }
func (zeroRNG) Shuffle(int, func(i, j int)) {}

// NewFlightForTest returns a started game at step place as described by s, with no pending
// events. It panics on an invalid setup.
func NewFlightForTest(s FlightSetup) *Game {
	g, err := NewGame("TEST1", alice, "Alice")
	if err == nil {
		err = g.AddPlayer(bob, "Bob")
	}
	scenario := s.Scenario
	if scenario == "" {
		scenario = DefaultScenario
	}
	if err == nil {
		err = g.Configure(alice, scenario, s.Abilities)
	}
	if err == nil {
		err = g.Start(alice, zeroRNG{})
	}
	if err != nil {
		panic(fmt.Sprintf("NewFlightForTest: %v", err))
	}
	g.PullEvents()

	sc := g.ScenarioDef()
	t := g.track()
	f := g.Flight
	f.Round = max(s.Round, 1)
	f.Position = max(s.Position, 1)
	if s.Planes != nil {
		if len(s.Planes) != t.Len() {
			panic(fmt.Sprintf("NewFlightForTest: %d plane counts for %d approach spaces", len(s.Planes), t.Len()))
		}
		f.Planes = slices.Clone(s.Planes)
	} else {
		for i, sp := range t.Spaces {
			f.Planes[i] = sp.Planes
		}
	}
	if f.Round > Rounds || f.Position > t.Len() {
		panic(fmt.Sprintf("NewFlightForTest: round %d, position %d of %d", f.Round, f.Position, t.Len()))
	}
	f.Axis = s.Axis
	f.Gear = s.Gear
	f.Flaps = s.Flaps
	f.Brakes = s.Brakes
	f.Ice = s.Ice
	f.Coffee = s.Coffee
	f.Rerolls = s.Rerolls
	switch {
	case s.Fuel != nil:
		f.Fuel = *s.Fuel
	case sc.hasKerosene():
		f.Fuel = StartFuel
	default:
		f.Fuel = 0
	}
	f.Wind = s.Wind
	switch {
	case s.Intern != nil:
		f.Intern = slices.Clone(s.Intern)
	case sc.Has(ModuleIntern):
		f.Intern = []int{1, 2, 3, 4, 5, 6}
	default:
		f.Intern = nil
	}
	f.InternLeft, f.InternRight = 0, 0
	if f.Intern != nil {
		f.InternLeft = s.InternTaken[0]
		f.InternRight = len(f.Intern) - 1 - s.InternTaken[1]
	}
	f.R = Round{
		Step:       StepPlace,
		Turn:       s.Turn,
		Ready:      [2]bool{true, true},
		Dice:       s.Dice,
		Discarded:  s.Discarded,
		Placements: slices.Clone(s.Placements),
	}
	return g
}

// MustPlace places for role and fails the test on error.
func MustPlace(t *testing.T, g *Game, role Role, in PlaceIntent, rng RNG) {
	t.Helper()
	if err := g.Place(g.PlayerOf(role), in, rng); err != nil {
		t.Fatalf("place %+v by %s: %v", in, role, err)
	}
}

// LegalSpaces returns the spaces role may put one of their own dice of value on now (after
// coffee), in table order.
func LegalSpaces(g *Game, role Role, value int) []SpaceID {
	return g.legalSpaces(role, role.Piece(), value)
}

// FailUnchanged asserts that fn fails with want and leaves g unchanged, pending events included
// (validate before mutate).
func FailUnchanged(t *testing.T, g *Game, fn func(*Game) error, want error) {
	t.Helper()
	before := g.Clone()
	err := fn(g)
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
	if !reflect.DeepEqual(before, g) {
		t.Fatalf("state changed on error %v", err)
	}
}
