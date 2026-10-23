package domain

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

const (
	alice PlayerID = "alice" // seat 0, host
	bob   PlayerID = "bob"   // seat 1
)

// newLobbyGame returns a lobby with alice (host) and bob seated.
func newLobbyGame(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame("GAME1", alice, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer(bob, "Bob"); err != nil {
		t.Fatal(err)
	}
	g.PullEvents()
	return g
}

// must fails the test on err.
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// mustFail asserts that err matches want and that fn left g unchanged (validate before mutate).
func mustFail(t *testing.T, g *Game, want error, fn func(*Game) error) {
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

// eventsOf returns the pending events of type T.
func eventsOf[T Event](g *Game) []T {
	var out []T
	for _, e := range g.events {
		if v, ok := e.(T); ok {
			out = append(out, v)
		}
	}
	return out
}

// withScenario adds a scenario and its track to the data tables for the duration of the test,
// so lifecycle and view tests can exercise modules, traffic and the red side without depending
// on the transcribed data.
func withScenario(t *testing.T, s ScenarioDef, tr TrackDef) {
	t.Helper()
	oldS, oldT := scenarios, tracks
	scenarios = append(slices.Clone(scenarios), s)
	tracks = append(slices.Clone(tracks), tr)
	t.Cleanup(func() { scenarios, tracks = oldS, oldT })
}

// testRed is a red test scenario with every module but the leak and two abilities. Its track
// (TST-R, N = 5) has 3 traffic icons at the start and 11 plane tokens, so only one traffic
// token fits into the supply.
var (
	testRed = ScenarioDef{
		ID: "tst-red", Airport: "TST", City: "Test", Colour: ColourRed, Track: "TST-R",
		Modules:   []Module{ModuleKerosene, ModuleIntern, ModuleWind, ModuleIceBrakes},
		Abilities: 2,
	}
	testRedTrack = TrackDef{ID: "TST-R", Spaces: []TrackSpace{
		{0, 3, turn("")},
		{2, 0, turn("ooxxx")},
		{3, 0, turn("")},
		{3, 0, turn("xxooo")},
		{3, 0, turn("")},
	}}
)

// newStarted returns a started game (step discussion) of scenario with the given abilities.
func newStarted(t *testing.T, rng RNG, scenario ScenarioID, abilities ...AbilityID) *Game {
	t.Helper()
	g := newLobbyGame(t)
	must(t, g.Configure(alice, scenario, abilities))
	must(t, g.Start(alice, rng))
	g.PullEvents()
	return g
}

// reverseRNG is a ScriptedRNG whose Shuffle reverses, to prove a shuffle uses the RNG.
type reverseRNG struct{ *TestRNG }

func (reverseRNG) Shuffle(n int, swap func(i, j int)) {
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
