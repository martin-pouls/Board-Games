package domain

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestViewInLobby(t *testing.T) {
	g := newLobbyGame(t)
	must(t, g.SwapRoles(alice))
	v, err := g.ViewFor(bob)
	must(t, err)
	if v.Phase != PhaseLobby || v.You != bob || v.HostID != alice || len(v.Players) != 2 || v.Flight != nil || v.Result != nil {
		t.Fatalf("lobby view %+v", v)
	}
	if !v.Players[1].IsYou || !v.Players[0].IsHost || v.Players[0].IsYou || v.Players[1].Seat != 1 {
		t.Fatalf("players %+v", v.Players)
	}
	if v.Players[0].Role != RoleCopilot || v.Players[1].Role != RolePilot || v.Players[0].MustAct || v.Players[0].RevealedDice != nil {
		t.Fatalf("roles %+v", v.Players)
	}
	if v.Scenario.ID != DefaultScenario || v.Scenario.Airport != "YUL" || v.Scenario.Spaces != 7 || v.Scenario.Modules == nil {
		t.Fatalf("scenario %+v", v.Scenario)
	}
	if v.Abilities == nil || len(v.Abilities) != 0 {
		t.Fatalf("abilities %v", v.Abilities)
	}
	if len(v.Catalogue) != len(scenarios) || !reflect.DeepEqual(v.Catalogue[0], v.Scenario) {
		t.Fatalf("catalogue %+v", v.Catalogue)
	}
	if _, err := g.ViewFor("carol"); !errors.Is(err, ErrUnknownPlayer) {
		t.Fatalf("unknown viewer: %v", err)
	}
	must(t, g.EndGame(alice))
	if v, _ := g.ViewFor(alice); v.Catalogue != nil || v.Flight != nil {
		t.Fatalf("ended view %+v", v)
	}
}

func TestViewInDiscussion(t *testing.T) {
	g := newStarted(t, zeroRNG{}, DefaultScenario)
	v, err := g.ViewFor(alice)
	must(t, err)
	if v.Catalogue != nil || v.Flight == nil || v.Result != nil {
		t.Fatalf("view %+v", v)
	}
	for _, p := range v.Players {
		if !p.MustAct || p.Ready || p.DiceLeft != 0 {
			t.Fatalf("player %+v", p)
		}
	}
	f := v.Flight
	if f.Round != 1 || f.FirstPlayer != RolePilot || f.Step != StepDiscussion || f.Turn != RolePilot {
		t.Fatalf("flight %+v", f)
	}
	if f.Approach.Position != 1 || len(f.Approach.Spaces) != 7 || f.Approach.Spaces[5] != (ApproachSpaceView{Planes: 3}) {
		t.Fatalf("approach %+v", f.Approach)
	}
	if !slices.Equal(f.RerollRounds, []int{5}) || f.Rerolls != 1 || f.Axis != 0 || f.AeroBlue != 4 || f.AeroOrange != 8 ||
		f.BrakeLimit != 2 || f.PlaneSupply != 3 || f.Coffee != 0 {
		t.Fatalf("panel %+v", f)
	}
	if f.Fuel != nil || f.Leak || f.Wind != nil || f.Ice != nil || f.Intern != nil || f.Speed != nil {
		t.Fatalf("modules %+v", f)
	}
	if !f.RoundStart.RerollPicked || f.RoundStart.Traffic == nil || len(f.Placements) != 0 || f.Placements == nil {
		t.Fatalf("round start %+v", f.RoundStart)
	}
	if f.YourDice != [HandSize]int{} || f.Legal != nil || f.Reroll != nil || f.Swap != nil || f.Sync != nil {
		t.Fatalf("step fields %+v", f)
	}

	g.Flight.R.Ready[RolePilot] = true
	v, _ = g.ViewFor(bob)
	if !v.Players[0].Ready || v.Players[0].MustAct || !v.Players[1].MustAct {
		t.Fatalf("ready %+v", v.Players)
	}
}

func TestViewModules(t *testing.T) {
	withScenario(t, testRed, testRedTrack)
	g := NewFlightForTest(FlightSetup{
		Scenario: "tst-red", Abilities: []AbilityID{AbilityControl, AbilityMastery},
		Fuel: new(7), Wind: 7, Ice: 2, Intern: []int{3, 1, 2, 6, 5, 4}, InternTaken: [2]int{1, 2},
		Dice: [2][HandSize]int{{1, 0, 0, 0}, {2, 0, 0, 0}},
	})
	v, _ := g.ViewFor(alice)
	f := v.Flight
	if f.Fuel == nil || *f.Fuel != 7 || f.Leak || f.Wind == nil || *f.Wind != (WindView{7, -2}) || f.Ice == nil || f.Ice.Marker != 2 || f.BrakeLimit != 4 {
		t.Fatalf("modules %+v", f)
	}
	if f.Approach.Spaces[0].Traffic != 3 || f.Approach.Spaces[1].Turn != "ooxxx" || f.Approach.Spaces[2].Turn != "" {
		t.Fatalf("approach %+v", f.Approach)
	}
	iv := f.Intern
	wantTokens := []InternTokenView{{3, true}, {1, false}, {2, false}, {6, false}, {5, true}, {4, true}}
	if iv == nil || !slices.Equal(iv.Tokens, wantTokens) || *iv.NextPilot != 1 || *iv.NextCopilot != 3 {
		t.Fatalf("intern %+v", iv)
	}
	g.Flight.InternLeft = 4
	if v, _ := g.ViewFor(alice); v.Flight.Intern.NextPilot != nil || v.Flight.Intern.NextCopilot != nil {
		t.Fatal("used-up row has no next token")
	}
}

func TestViewHidesPartnerDice(t *testing.T) {
	g := NewFlightForTest(FlightSetup{
		Turn: RoleCopilot,
		Dice: [2][HandSize]int{{1, 0, 3, 4}, {6, 5, 0, 0}},
		Placements: []Placement{
			{Space: SpaceAxisPilot, Piece: PiecePilot, Value: 2, Coffee: -1, Role: RolePilot, Slot: 1},
			{Space: SpaceAxisCopilot, Piece: PieceCopilot, Value: 4, Role: RoleCopilot, Slot: 2},
		},
	})
	va, _ := g.ViewFor(alice)
	vb, _ := g.ViewFor(bob)
	if va.Flight.YourDice != [HandSize]int{1, 0, 3, 4} || vb.Flight.YourDice != [HandSize]int{6, 5, 0, 0} {
		t.Fatalf("own dice %v / %v", va.Flight.YourDice, vb.Flight.YourDice)
	}
	if va.Players[1].DiceLeft != 2 || vb.Players[0].DiceLeft != 3 || va.Players[1].RevealedDice != nil {
		t.Fatalf("partner %+v / %+v", va.Players[1], vb.Players[0])
	}
	if !va.Players[1].MustAct || va.Players[0].MustAct {
		t.Fatalf("must act %+v", va.Players)
	}
	want := []PlacementView{
		{Space: SpaceAxisPilot, Piece: PiecePilot, Value: 2, Coffee: -1, Role: RolePilot},
		{Space: SpaceAxisCopilot, Piece: PieceCopilot, Value: 4, Role: RoleCopilot},
	}
	if !slices.Equal(va.Flight.Placements, want) || !slices.Equal(vb.Flight.Placements, want) {
		t.Fatalf("placements %+v", va.Flight.Placements)
	}

	// Game over: both hands are revealed (§7.2).
	must(t, g.GiveUp(alice))
	v, _ := g.ViewFor(bob)
	if !slices.Equal(v.Players[0].RevealedDice, []int{1, 3, 4}) || !slices.Equal(v.Players[1].RevealedDice, []int{6, 5}) {
		t.Fatalf("revealed %+v", v.Players)
	}
	if v.Result == nil || v.Result.Won || v.Result.Reason != LossGaveUp || v.Result.PlayerID != alice || v.Result.Failed == nil {
		t.Fatalf("result %+v", v.Result)
	}
	if v.Players[0].MustAct || v.Players[1].MustAct {
		t.Fatal("nobody acts after the game")
	}
}

func TestViewSteps(t *testing.T) {
	g := NewFlightForTest(FlightSetup{Dice: [2][HandSize]int{{1, 2, 3, 0}, {6, 5, 0, 0}}})
	r := &g.Flight.R

	r.Step, r.RerollConfirmed = StepReroll, [2]bool{true, false}
	va, _ := g.ViewFor(alice)
	if va.Flight.Reroll == nil || va.Flight.Reroll.Confirmed != [2]bool{true, false} || !va.Players[1].MustAct || va.Players[0].MustAct {
		t.Fatalf("reroll %+v", va.Flight.Reroll)
	}

	// Working Together: the offered value is public, the slot only for the initiator.
	r.Step, r.Swap = StepSwap, &SwapOffer{By: RolePilot, Slot: 2}
	va, _ = g.ViewFor(alice)
	vb, _ := g.ViewFor(bob)
	if sv := va.Flight.Swap; sv == nil || sv.By != RolePilot || sv.Value != 3 || sv.YourSlot == nil || *sv.YourSlot != 2 {
		t.Fatalf("initiator %+v", va.Flight.Swap)
	}
	if sv := vb.Flight.Swap; sv == nil || sv.Value != 3 || sv.YourSlot != nil || !vb.Players[1].MustAct {
		t.Fatalf("partner %+v", vb.Flight.Swap)
	}

	r.Step, r.Swap, r.SyncValue = StepSync, nil, 4
	va, _ = g.ViewFor(alice)
	vb, _ = g.ViewFor(bob)
	if va.Flight.Sync == nil || va.Flight.Sync.Value != 4 || va.Flight.Sync.Targets != nil || va.Flight.Swap != nil {
		t.Fatalf("pilot sync %+v", va.Flight.Sync)
	}
	if vb.Flight.Sync == nil || vb.Flight.Sync.Targets == nil {
		t.Fatalf("co-pilot sync %+v", vb.Flight.Sync)
	}
}

func TestViewSpeed(t *testing.T) {
	g := NewFlightForTest(FlightSetup{Round: 6, Dice: [2][HandSize]int{{1, 0, 0, 0}, {2, 0, 0, 0}}})
	g.Flight.R.Speed = &Speed{Sum: 9, Wind: -2, Move: 1}
	v, _ := g.ViewFor(alice)
	if s := v.Flight.Speed; s == nil || s.Sum != 9 || s.Wind != -2 || s.Total != 7 || s.Move == nil || *s.Move != 1 {
		t.Fatalf("speed %+v", v.Flight.Speed)
	}
	// The final round does not move (§6.2).
	g.Flight.Round = Rounds
	if v, _ := g.ViewFor(alice); v.Flight.Speed.Move != nil {
		t.Fatal("move in the final round")
	}
}
