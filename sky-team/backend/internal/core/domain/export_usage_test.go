package domain_test

// Checks that the helpers in export_test.go serve the black-box tests: they compile from
// package domain_test and build the states they promise.

import (
	"slices"
	"testing"

	"github.com/martin-pouls/sky-team/backend/internal/core/domain"
)

func TestScriptedRNG(t *testing.T) {
	rng := domain.ScriptedRNG(domain.DieRoll(6), domain.TrafficDie(2), domain.TrafficDie(3), domain.TrafficDie(4), domain.TrafficDie(5))
	got := []int{1 + rng.IntN(6)}
	for range 4 {
		got = append(got, domain.TrafficFaces()[rng.IntN(6)])
	}
	if !slices.Equal(got, []int{6, 2, 3, 4, 5}) || rng.Left() != 0 {
		t.Fatalf("rolls %v, %d left", got, rng.Left())
	}
	if !slices.Equal(domain.DieRolls(1, 6), []int{0, 5}) {
		t.Fatal("DieRolls")
	}
	row := []int{1, 2, 3}
	rng.Shuffle(len(row), func(i, j int) { row[i], row[j] = row[j], row[i] })
	if !slices.Equal(row, []int{1, 2, 3}) {
		t.Fatal("Shuffle must be the identity")
	}
	for name, fn := range map[string]func(){
		"exhausted":    func() { rng.IntN(6) },
		"out of range": func() { domain.ScriptedRNG(6).IntN(6) },
		"bad die":      func() { domain.DieRoll(7) },
		"bad face":     func() { domain.TrafficDie(6) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			fn()
		}()
	}
}

func TestNewFlightForTest(t *testing.T) {
	placed := []domain.Placement{{Space: domain.SpaceAxisPilot, Piece: domain.PiecePilot, Value: 4, Role: domain.RolePilot, Slot: 3}}
	g := domain.NewFlightForTest(domain.FlightSetup{
		Round: 3, Position: 4, Planes: []int{0, 0, 0, 1, 1, 0, 2}, Axis: -1, Gear: [3]bool{true, false, false},
		Flaps: 2, Brakes: 1, Coffee: 2, Rerolls: 1,
		Dice:       [2][domain.HandSize]int{{1, 2, 3, 0}, {6, 6, 0, 5}},
		Discarded:  [2]int{0, 1},
		Turn:       domain.RoleCopilot,
		Placements: placed,
	})
	if len(g.PullEvents()) != 0 || g.Phase != domain.PhasePlaying || g.PlayerOf(domain.RolePilot) != domain.Alice || g.PlayerOf(domain.RoleCopilot) != domain.Bob {
		t.Fatalf("game %+v", g)
	}
	f := g.Flight
	if f.Round != 3 || f.Position != 4 || !slices.Equal(f.Planes, []int{0, 0, 0, 1, 1, 0, 2}) || f.Axis != -1 ||
		f.Gear != [3]bool{true, false, false} || f.Flaps != 2 || f.Brakes != 1 || f.Coffee != 2 || f.Rerolls != 1 || f.Fuel != 0 || f.Intern != nil {
		t.Fatalf("flight %+v", f)
	}
	r := f.R
	if r.Step != domain.StepPlace || r.Turn != domain.RoleCopilot || r.Dice[1] != [domain.HandSize]int{6, 6, 0, 5} ||
		r.Discarded != [2]int{0, 1} || !slices.Equal(r.Placements, placed) {
		t.Fatalf("round %+v", r)
	}
	if !slices.Equal(g.Actors(), []domain.Role{domain.RoleCopilot}) {
		t.Fatalf("actors %v", g.Actors())
	}
	placed[0].Value = 1
	if g.Flight.R.Placements[0].Value != 4 {
		t.Fatal("placements shared with the setup")
	}

	// Defaults: the scenario's track, round 1, space 1.
	g = domain.NewFlightForTest(domain.FlightSetup{})
	if g.Scenario != domain.DefaultScenario || g.Flight.Round != 1 || g.Flight.Position != 1 || !slices.Equal(g.Flight.Planes, []int{0, 0, 1, 2, 1, 3, 2}) {
		t.Fatalf("defaults %+v", g.Flight)
	}
}

func TestFailUnchanged(t *testing.T) {
	g := domain.NewFlightForTest(domain.FlightSetup{Dice: [2][domain.HandSize]int{{1, 0, 0, 0}, {2, 0, 0, 0}}})
	domain.FailUnchanged(t, g, func(g *domain.Game) error { return g.GiveUp("carol") }, domain.ErrUnknownPlayer)
	if spaces := domain.LegalSpaces(g, domain.RolePilot, 1); spaces == nil {
		t.Fatal("LegalSpaces returns a slice")
	}
}
