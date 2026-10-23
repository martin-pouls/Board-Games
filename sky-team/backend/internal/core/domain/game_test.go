package domain

import (
	"reflect"
	"slices"
	"testing"
)

func TestRole(t *testing.T) {
	if RolePilot.Other() != RoleCopilot || RoleCopilot.Other() != RolePilot {
		t.Fatal("Other")
	}
	for _, r := range Roles {
		if back, ok := ParseRole(r.String()); !ok || back != r || !r.Valid() {
			t.Fatalf("round trip %s", r)
		}
	}
	if _, ok := ParseRole("navigator"); ok || Role(2).Valid() {
		t.Fatal("invalid role accepted")
	}
	if RolePilot.Colour() != ColourBlue || RoleCopilot.Colour() != ColourOrange ||
		RolePilot.Piece() != PiecePilot || RoleCopilot.Piece() != PieceCopilot {
		t.Fatal("colours and pieces (§1)")
	}
}

func TestActors(t *testing.T) {
	g := NewFlightForTest(FlightSetup{Turn: RoleCopilot, Dice: [2][HandSize]int{{1, 0, 0, 0}, {2, 0, 0, 0}}})
	r := &g.Flight.R
	tests := []struct {
		name  string
		setup func()
		want  []Role
	}{
		{"place: the turn player", func() {}, []Role{RoleCopilot}},
		{"discussion: both", func() { r.Step, r.Ready = StepDiscussion, [2]bool{} }, []Role{RolePilot, RoleCopilot}},
		{"discussion: the one not ready", func() { r.Step, r.Ready = StepDiscussion, [2]bool{true, false} }, []Role{RoleCopilot}},
		{"reroll: both", func() { r.Step = StepReroll }, []Role{RolePilot, RoleCopilot}},
		{"reroll: the one not confirmed", func() { r.Step, r.RerollConfirmed = StepReroll, [2]bool{false, true} }, []Role{RolePilot}},
		{"swap: the partner", func() { r.Step, r.Swap = StepSwap, &SwapOffer{By: RoleCopilot, Slot: 0} }, []Role{RolePilot}},
		{"sync: the Co-Pilot", func() { r.Step = StepSync }, []Role{RoleCopilot}},
	}
	base := g.Clone()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g.Flight.R = base.Flight.R.clone()
			r = &g.Flight.R
			tt.setup()
			if got := g.Actors(); !slices.Equal(got, tt.want) {
				t.Fatalf("actors %v, want %v", got, tt.want)
			}
		})
	}
	if newLobbyGame(t).Actors() != nil {
		t.Fatal("actors in the lobby")
	}
}

func TestRequireActor(t *testing.T) {
	g := NewFlightForTest(FlightSetup{Turn: RoleCopilot, Dice: [2][HandSize]int{{1, 0, 0, 0}, {2, 0, 0, 0}}})
	if role, err := g.requireActor(bob, StepPlace); err != nil || role != RoleCopilot {
		t.Fatalf("bob: %v %v", role, err)
	}
	for _, tt := range []struct {
		by    PlayerID
		steps []Step
		want  error
	}{
		{"carol", []Step{StepPlace}, ErrUnknownPlayer},
		{alice, []Step{StepPlace}, ErrNotYourTurn},
		{bob, []Step{StepDiscussion}, ErrWrongStep},
	} {
		mustFail(t, g, tt.want, func(g *Game) error { _, err := g.requireActor(tt.by, tt.steps...); return err })
	}
	// Side intents only need the step (§13).
	if role, err := g.requirePlaying(alice, StepPlace); err != nil || role != RolePilot {
		t.Fatalf("alice: %v %v", role, err)
	}
	mustFail(t, newLobbyGame(t), ErrWrongPhase, func(g *Game) error { _, err := g.requirePlaying(alice, StepPlace); return err })
}

func TestFlightDerived(t *testing.T) {
	f := &Flight{Planes: []int{1, 0, 2, 3}}
	if f.AeroBlue() != 4 || f.AeroOrange() != 8 || f.PlaneSupply() != 6 || f.AtAirport() {
		t.Fatalf("start %d %d %d", f.AeroBlue(), f.AeroOrange(), f.PlaneSupply())
	}
	f.Gear = [3]bool{true, false, true}
	f.Flaps = 4
	f.Position = 4
	if f.GearDown() != 2 || f.AeroBlue() != 6 || f.AeroOrange() != 12 || !f.AtAirport() {
		t.Fatalf("deployed %d %d", f.AeroBlue(), f.AeroOrange())
	}
	f.Gear = [3]bool{true, true, true}
	if f.AeroBlue() != 7 {
		t.Fatalf("all gear %d", f.AeroBlue())
	}
	// §3.4: highest landing speed 1, 2, 4, 6; §12.6: the ice marker, 5 or less once past 5.
	for brakes, limit := range []int{2, 3, 5, 7} {
		f.Brakes = brakes
		if f.BrakeLimit(false) != limit {
			t.Errorf("%d brakes: limit %d, want %d", brakes, f.BrakeLimit(false), limit)
		}
	}
	for ice, limit := range []int{2, 3, 4, 5, 6} {
		f.Ice = ice
		if f.BrakeLimit(true) != limit {
			t.Errorf("ice %d: limit %d, want %d", ice, f.BrakeLimit(true), limit)
		}
	}
}

func TestRerollAccounting(t *testing.T) {
	withScenario(t, testRed, testRedTrack)
	green := NewFlightForTest(FlightSetup{})
	red := NewFlightForTest(FlightSetup{Scenario: "tst-red", Abilities: []AbilityID{AbilityControl, AbilityMastery}})
	tests := []struct {
		g              *Game
		round, supply  int
		onTrack, inBox int
	}{
		{green, 1, 1, 1, 0},
		{green, 4, 0, 1, 1},
		{green, 5, 2, 0, 0},
		{green, 7, 0, 0, 2},
		{red, 1, 1, 0, 1},
		{red, 3, 0, 0, 2},
	}
	for _, tt := range tests {
		tt.g.Flight.Round, tt.g.Flight.Rerolls = tt.round, tt.supply
		if len(tt.g.RerollRounds()) != tt.onTrack || tt.g.rerollsInBox() != tt.inBox {
			t.Errorf("%s round %d supply %d: track %v box %d", tt.g.Scenario, tt.round, tt.supply, tt.g.RerollRounds(), tt.g.rerollsInBox())
		}
	}
}

func TestCloneIsDeep(t *testing.T) {
	withScenario(t, testRed, testRedTrack)
	g := NewFlightForTest(FlightSetup{
		Scenario: "tst-red", Abilities: []AbilityID{AbilityControl, AbilityMastery},
		Dice:       [2][HandSize]int{{1, 2, 3, 4}, {5, 6, 1, 2}},
		Placements: []Placement{{Space: SpaceAxisPilot, Piece: PiecePilot, Value: 3, Slot: 0}},
	})
	g.Flight.R.Speed = &Speed{Sum: 7}
	g.Flight.R.Swap = &SwapOffer{By: RolePilot, Slot: 1}
	g.Flight.R.Start.Traffic = []TrafficRoll{{2, 2, true}}
	g.Result = &Result{Failed: []Condition{ConditionA}}
	g.emit(DiceRolled{})
	c := g.Clone()
	if !reflect.DeepEqual(c, g) {
		t.Fatal("clone differs")
	}
	c.Players[0].Name = "X"
	c.Abilities[0] = AbilityAdaptation
	c.Flight.Planes[1] = 9
	c.Flight.Intern[0] = 9
	c.Flight.R.Placements[0].Value = 6
	c.Flight.R.Speed.Sum = 1
	c.Flight.R.Swap.Slot = 3
	c.Flight.R.Start.Traffic[0].Roll = 5
	c.Result.Failed[0] = ConditionB
	c.events[0] = GameWon{}
	f := g.Flight
	if g.Players[0].Name != "Alice" || g.Abilities[0] != AbilityControl || f.Planes[1] != 2 || f.Intern[0] != 1 ||
		f.R.Placements[0].Value != 3 || f.R.Speed.Sum != 7 || f.R.Swap.Slot != 1 || f.R.Start.Traffic[0].Roll != 2 ||
		g.Result.Failed[0] != ConditionA || g.events[0] != (DiceRolled{}) {
		t.Fatal("clone shares state")
	}
	var nilGame *Game
	if nilGame.Clone() != nil {
		t.Fatal("nil clone")
	}
}

func TestInternRow(t *testing.T) {
	f := &Flight{}
	if _, ok := f.internNext(RolePilot); ok || f.internLeft() != 0 {
		t.Fatal("no module: no tokens")
	}
	f.Intern = []int{4, 1, 6, 2, 5, 3}
	f.InternRight = 5
	if i, _ := f.internNext(RolePilot); i != 0 {
		t.Fatalf("pilot takes from the left: %d", i)
	}
	if i, _ := f.internNext(RoleCopilot); i != 5 {
		t.Fatalf("co-pilot takes from the right: %d", i)
	}
	f.InternLeft, f.InternRight = 3, 3 // three taken by the Pilot, two by the Co-Pilot
	if f.internLeft() != 1 || f.internTaken(3) || !f.internTaken(2) || !f.internTaken(4) {
		t.Fatalf("left %d", f.internLeft())
	}
	if i, _ := f.internNext(RoleCopilot); i != 3 {
		t.Fatalf("last token %d", i)
	}
	f.InternLeft = 4
	if _, ok := f.internNext(RolePilot); ok || f.internLeft() != 0 {
		t.Fatal("row used up")
	}
}
