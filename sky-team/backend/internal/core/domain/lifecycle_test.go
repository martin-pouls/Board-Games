package domain

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestNewGameAndJoin(t *testing.T) {
	if _, err := NewGame("G", alice, "  "); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("empty name: %v", err)
	}
	if _, err := NewGame("G", alice, "abcdefghijklmnopqrstu"); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("long name: %v", err)
	}
	if _, err := NewGame("G", alice, "a\tb"); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("control char: %v", err)
	}
	g, err := NewGame("G", alice, " Alice ")
	must(t, err)
	if g.Phase != PhaseLobby || g.Revision != 1 || g.Players[0].Name != "Alice" || !g.IsHost(alice) {
		t.Fatalf("bad new game %+v", g)
	}
	mustFail(t, g, ErrNameTaken, func(g *Game) error { return g.AddPlayer(bob, "alice") })
	mustFail(t, g, ErrDuplicatePlayer, func(g *Game) error { return g.AddPlayer(alice, "Other") })
	mustFail(t, g, ErrInvalidName, func(g *Game) error { return g.AddPlayer(bob, " ") })
	must(t, g.AddPlayer(bob, "Bob"))
	if g.Revision != 2 || len(eventsOf[PlayerJoined](g)) != 1 {
		t.Fatalf("revision %d events %v", g.Revision, g.events)
	}
	mustFail(t, g, ErrGameFull, func(g *Game) error { return g.AddPlayer("carol", "Carol") })
	if s := g.Summary(); s.Joinable || s.PlayerCount != 2 || s.MaxPlayers != 2 || s.HostName != "Alice" {
		t.Fatalf("summary %+v", s)
	}
}

func TestAddPlayerOutsideLobby(t *testing.T) {
	g, err := NewGame("G", alice, "Alice")
	must(t, err)
	must(t, g.EndGame(alice))
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.AddPlayer(bob, "Bob") })
}

func TestLeave(t *testing.T) {
	g := newLobbyGame(t)
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.Leave("carol") })
	must(t, g.Leave(alice))
	if g.HostID != bob || len(g.Players) != 1 {
		t.Fatalf("host %s players %v", g.HostID, g.Players)
	}
	if ev := eventsOf[PlayerLeft](g); len(ev) != 1 || ev[0].HostID != bob {
		t.Fatalf("events %v", g.events)
	}
	mustFail(t, g, ErrCannotLeave, func(g *Game) error { return g.Leave(bob) })
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.Leave(alice) })

	g = newLobbyGame(t)
	must(t, g.EndGame(alice))
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.Leave(bob) })
}

func TestEndGame(t *testing.T) {
	g := newLobbyGame(t)
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.EndGame(bob) })
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.EndGame("carol") })
	must(t, g.EndGame(alice))
	if g.Phase != PhaseEnded || len(eventsOf[GameEnded](g)) != 1 {
		t.Fatalf("phase %s events %v", g.Phase, g.events)
	}
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.EndGame(alice) })
}

func TestNewGameDefaults(t *testing.T) {
	g := newLobbyGame(t)
	if g.Scenario != DefaultScenario || g.Abilities == nil || len(g.Abilities) != 0 || g.PilotSeat != 0 || g.Flight != nil {
		t.Fatalf("defaults %+v", g)
	}
	if g.RoleOf(0) != RolePilot || g.RoleOf(1) != RoleCopilot || g.PlayerOf(RolePilot) != alice || g.PlayerOf(RoleCopilot) != bob {
		t.Fatal("seat 0 must be the Pilot")
	}
}

func TestConfigure(t *testing.T) {
	withScenario(t, testRed, testRedTrack)
	g := newLobbyGame(t)
	must(t, g.Configure(alice, "tst-red", []AbilityID{AbilityControl}))
	if g.Scenario != "tst-red" || !slices.Equal(g.Abilities, []AbilityID{AbilityControl}) || g.Revision != 3 {
		t.Fatalf("configured %s %v rev %d", g.Scenario, g.Abilities, g.Revision)
	}
	if ev := eventsOf[FlightConfigured](g); len(ev) != 1 || ev[0].Scenario != "tst-red" || !slices.Equal(ev[0].Abilities, []AbilityID{AbilityControl}) {
		t.Fatalf("events %v", g.events)
	}
	must(t, g.Configure(alice, "tst-red", []AbilityID{AbilityMastery, AbilityAdaptation}))
	if !slices.Equal(g.Abilities, []AbilityID{AbilityMastery, AbilityAdaptation}) {
		t.Fatalf("abilities %v", g.Abilities)
	}
	must(t, g.Configure(alice, DefaultScenario, nil))
	if g.Scenario != DefaultScenario || g.Abilities == nil || len(g.Abilities) != 0 {
		t.Fatalf("reset %s %v", g.Scenario, g.Abilities)
	}
}

func TestConfigureErrors(t *testing.T) {
	withScenario(t, testRed, testRedTrack)
	g := newLobbyGame(t)
	tests := []struct {
		name string
		by   PlayerID
		sc   ScenarioID
		ab   []AbilityID
		want error
	}{
		{"unknown player", "carol", DefaultScenario, nil, ErrUnknownPlayer},
		{"not host", bob, DefaultScenario, nil, ErrNotHost},
		{"unknown scenario", alice, "xyz-green", nil, ErrUnknownScenario},
		{"too many for yul", alice, DefaultScenario, []AbilityID{AbilityControl}, ErrInvalidAbilities},
		{"too many", alice, "tst-red", []AbilityID{AbilityControl, AbilityMastery, AbilityAdaptation}, ErrInvalidAbilities},
		{"unknown card", alice, "tst-red", []AbilityID{"luck"}, ErrInvalidAbilities},
		{"duplicate", alice, "tst-red", []AbilityID{AbilityControl, AbilityControl}, ErrInvalidAbilities},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mustFail(t, g, tt.want, func(g *Game) error { return g.Configure(tt.by, tt.sc, tt.ab) })
		})
	}
	started := newStarted(t, zeroRNG{}, DefaultScenario)
	mustFail(t, started, ErrWrongPhase, func(g *Game) error { return g.Configure(alice, DefaultScenario, nil) })
}

func TestSwapRoles(t *testing.T) {
	g := newLobbyGame(t)
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.SwapRoles(bob) })
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.SwapRoles("carol") })
	must(t, g.SwapRoles(alice))
	if g.PilotSeat != 1 || g.RoleOf(0) != RoleCopilot || g.SeatOf(RolePilot) != 1 || g.PlayerOf(RolePilot) != bob {
		t.Fatalf("pilot seat %d", g.PilotSeat)
	}
	if ev := eventsOf[RolesSwapped](g); len(ev) != 1 || ev[0].PilotSeat != 1 {
		t.Fatalf("events %v", g.events)
	}
	must(t, g.SwapRoles(alice))
	if g.PilotSeat != 0 {
		t.Fatalf("pilot seat %d after swapping back", g.PilotSeat)
	}
	started := newStarted(t, zeroRNG{}, DefaultScenario)
	mustFail(t, started, ErrWrongPhase, func(g *Game) error { return g.SwapRoles(alice) })
}

func TestLeaveKeepsRoles(t *testing.T) {
	// Alice (seat 0) is the Pilot and leaves: Bob moves to seat 0 and stays the Co-Pilot.
	g := newLobbyGame(t)
	must(t, g.Leave(alice))
	if g.RoleOf(0) != RoleCopilot || g.PlayerOf(RoleCopilot) != bob || g.PlayerOf(RolePilot) != "" {
		t.Fatalf("pilot seat %d", g.PilotSeat)
	}
	must(t, g.AddPlayer("carol", "Carol"))
	if g.PlayerOf(RolePilot) != "carol" {
		t.Fatal("the newcomer must take the free role")
	}
	// Bob (seat 1) leaves: nothing moves.
	g = newLobbyGame(t)
	must(t, g.SwapRoles(alice))
	must(t, g.Leave(bob))
	if g.PlayerOf(RoleCopilot) != alice || g.PilotSeat != 1 {
		t.Fatalf("pilot seat %d", g.PilotSeat)
	}
}

func TestStartYUL(t *testing.T) {
	g := newLobbyGame(t)
	must(t, g.Start(alice, ScriptedRNG()))
	f := g.Flight
	if g.Phase != PhasePlaying || g.Revision != 3 || f == nil {
		t.Fatalf("phase %s rev %d", g.Phase, g.Revision)
	}
	if f.Round != 1 || f.Position != 1 || !slices.Equal(f.Planes, []int{0, 0, 1, 2, 1, 3, 2}) {
		t.Fatalf("track %+v", f)
	}
	// §4.1: axis level, all switches red, no coffee; §5.1: the round-1 reroll is picked up.
	if f.Axis != 0 || f.Gear != [3]bool{} || f.Flaps != 0 || f.Brakes != 0 || f.Coffee != 0 || f.Rerolls != 1 {
		t.Fatalf("panel %+v", f)
	}
	if f.Fuel != 0 || f.Wind != 0 || f.Ice != 0 || f.Intern != nil || f.AdaptationUsed != [2]bool{} {
		t.Fatalf("modules %+v", f)
	}
	r := f.R
	if r.Step != StepDiscussion || r.Turn != RolePilot || r.Dice != [2][HandSize]int{} || len(r.Placements) != 0 {
		t.Fatalf("round %+v", r)
	}
	if !r.Start.RerollPicked || len(r.Start.Traffic) != 0 {
		t.Fatalf("round start %+v", r.Start)
	}
	if !slices.Equal(g.Actors(), []Role{RolePilot, RoleCopilot}) {
		t.Fatalf("actors %v", g.Actors())
	}
	if len(eventsOf[GameStarted](g)) != 1 {
		t.Fatalf("events %v", g.events)
	}
	if ev := eventsOf[RoundStarted](g); len(ev) != 1 || ev[0].Round != 1 || !ev[0].RerollPicked {
		t.Fatalf("events %v", g.events)
	}
	if !slices.Equal(g.RerollRounds(), []int{5}) || g.rerollsInBox() != 0 {
		t.Fatalf("reroll rounds %v", g.RerollRounds())
	}
}

func TestStartModulesAndTraffic(t *testing.T) {
	withScenario(t, testRed, testRedTrack)
	g := newLobbyGame(t)
	must(t, g.Configure(alice, "tst-red", []AbilityID{AbilityControl, AbilityMastery}))
	// Three traffic icons on space 1: rolls 5 (space 5), 2 (space 2), 3 (space 3). The supply
	// holds one token (11 on the track), so only the first roll adds one.
	rng := reverseRNG{ScriptedRNG(TrafficDie(5), TrafficDie(2), TrafficDie(3))}
	must(t, g.Start(alice, rng))
	f := g.Flight
	if rng.Left() != 0 {
		t.Fatalf("%d rolls unused", rng.Left())
	}
	if !slices.Equal(f.Planes, []int{0, 2, 3, 3, 4}) || f.PlaneSupply() != 0 {
		t.Fatalf("planes %v", f.Planes)
	}
	want := []TrafficRoll{{5, 5, true}, {2, 2, false}, {3, 3, false}}
	if !slices.Equal(f.R.Start.Traffic, want) {
		t.Fatalf("traffic %+v", f.R.Start.Traffic)
	}
	if ev := eventsOf[RoundStarted](g); len(ev) != 1 || !slices.Equal(ev[0].Traffic, want) {
		t.Fatalf("events %v", g.events)
	}
	// The intern row is shuffled with the RNG (reversed here).
	if !slices.Equal(f.Intern, []int{6, 5, 4, 3, 2, 1}) || f.InternLeft != 0 || f.InternRight != 5 {
		t.Fatalf("intern %v %d %d", f.Intern, f.InternLeft, f.InternRight)
	}
	if f.Fuel != StartFuel || f.Wind != 0 || f.Ice != 0 {
		t.Fatalf("modules %+v", f)
	}
	// Red side: the round-1 token is picked up, none is left on the track, one stays in the box.
	if f.Rerolls != 1 || len(g.RerollRounds()) != 0 || g.rerollsInBox() != 1 {
		t.Fatalf("rerolls %d %v", f.Rerolls, g.RerollRounds())
	}
}

func TestRoundStartTraffic(t *testing.T) {
	withScenario(t, testRed, testRedTrack)
	g := newStarted(t, ScriptedRNG(TrafficDie(5), TrafficDie(5), TrafficDie(5)), "tst-red", AbilityControl, AbilityMastery)
	f := g.Flight
	// Round 2 on space 4: no traffic icon, nothing is rolled; the Co-Pilot is first (§4.2).
	f.Planes = []int{0, 0, 0, 0, 0}
	f.Round, f.Position = 2, 4
	g.startRound(ScriptedRNG())
	if len(f.R.Start.Traffic) != 0 || f.R.Start.RerollPicked || f.R.Turn != RoleCopilot || f.R.Step != StepDiscussion {
		t.Fatalf("round start %+v", f.R)
	}
	// On space 1 with room in the supply: 5 lands on the airport, 2 on space 2, 4 on space 4.
	f.Position = 1
	g.startRound(ScriptedRNG(TrafficDie(5), TrafficDie(2), TrafficDie(4)))
	if !slices.Equal(f.Planes, []int{0, 1, 0, 1, 1}) {
		t.Fatalf("planes %v", f.Planes)
	}
	// On space 4, a roll beyond the airport puts the token on the airport (§11.1).
	withScenario(t, ScenarioDef{ID: "tst-green", Colour: ColourGreen, Track: "TST-G"},
		TrackDef{ID: "TST-G", Spaces: []TrackSpace{{}, {}, {}, {0, 1, 0}, {}}})
	g = newStarted(t, ScriptedRNG(), "tst-green")
	f = g.Flight
	f.Round, f.Position = 5, 4
	g.startRound(ScriptedRNG(TrafficDie(4)))
	if !slices.Equal(f.Planes, []int{0, 0, 0, 0, 1}) || f.R.Start.Traffic[0] != (TrafficRoll{4, 5, true}) {
		t.Fatalf("planes %v traffic %+v", f.Planes, f.R.Start.Traffic)
	}
	// Round 5 on the green side picks up the second reroll token.
	if !f.R.Start.RerollPicked || f.Rerolls != 2 {
		t.Fatalf("pickup %+v rerolls %d", f.R.Start, f.Rerolls)
	}
}

func TestStartErrors(t *testing.T) {
	withScenario(t, testRed, testRedTrack)
	g := newLobbyGame(t)
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.Start("carol", zeroRNG{}) })
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.Start(bob, zeroRNG{}) })
	must(t, g.Configure(alice, "tst-red", []AbilityID{AbilityControl}))
	mustFail(t, g, ErrAbilitiesIncomplete, func(g *Game) error { return g.Start(alice, zeroRNG{}) })

	alone, err := NewGame("G", alice, "Alice")
	must(t, err)
	mustFail(t, alone, ErrNotEnoughPlayers, func(g *Game) error { return g.Start(alice, zeroRNG{}) })

	started := newStarted(t, zeroRNG{}, DefaultScenario)
	mustFail(t, started, ErrWrongPhase, func(g *Game) error { return g.Start(alice, zeroRNG{}) })
}

func TestGiveUp(t *testing.T) {
	g := newLobbyGame(t)
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.GiveUp(alice) })
	g = NewFlightForTest(FlightSetup{Dice: [2][HandSize]int{{1, 2, 0, 0}, {3, 0, 0, 0}}})
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.GiveUp("carol") })
	for _, step := range []Step{StepDiscussion, StepPlace, StepReroll, StepSwap, StepSync} {
		g := g.Clone()
		g.Flight.R.Step = step
		must(t, g.GiveUp(bob))
		if g.Phase != PhaseLost || g.Result == nil || g.Result.Reason != LossGaveUp || g.Result.PlayerID != bob {
			t.Fatalf("%s: phase %s result %+v", step, g.Phase, g.Result)
		}
		if ev := eventsOf[GameLost](g); len(ev) != 1 || ev[0].Reason != LossGaveUp || ev[0].PlayerID != bob {
			t.Fatalf("%s: events %v", step, g.events)
		}
		if g.Actors() != nil {
			t.Fatalf("%s: actors %v after the game", step, g.Actors())
		}
		mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.GiveUp(alice) })
	}
}

func TestRematch(t *testing.T) {
	withScenario(t, testRed, testRedTrack)
	g := newLobbyGame(t)
	must(t, g.SwapRoles(alice))
	must(t, g.Configure(alice, "tst-red", []AbilityID{AbilityControl, AbilityMastery}))
	rematch := func(by PlayerID) func(g *Game) error {
		return func(g *Game) error { _, err := g.Rematch(by, "NEXT1"); return err }
	}
	mustFail(t, g, ErrWrongPhase, rematch(alice))
	must(t, g.Start(alice, zeroRNG{}))
	mustFail(t, g, ErrWrongPhase, rematch(alice))
	must(t, g.GiveUp(alice))
	g.PullEvents()
	mustFail(t, g, ErrUnknownPlayer, rematch("carol"))

	next, err := g.Rematch(bob, "NEXT1")
	must(t, err)
	if g.NextGameID != "NEXT1" || len(eventsOf[RematchCreated](g)) != 1 {
		t.Fatalf("old game %+v", g)
	}
	if next.ID != "NEXT1" || next.Phase != PhaseLobby || next.HostID != bob || next.Revision != 1 || next.Flight != nil || next.Result != nil {
		t.Fatalf("new game %+v", next)
	}
	// Same seats, roles, scenario and abilities (§7.2), not shared with the old game.
	if !reflect.DeepEqual(next.Players, g.Players) || next.PilotSeat != 1 || next.Scenario != "tst-red" ||
		!slices.Equal(next.Abilities, g.Abilities) {
		t.Fatalf("new game %+v", next)
	}
	next.Abilities[0] = AbilityAdaptation
	next.Players[0].Name = "X"
	if g.Abilities[0] != AbilityControl || g.Players[0].Name != "Alice" {
		t.Fatal("rematch shares slices with the old game")
	}
	mustFail(t, g, ErrRematchExists, rematch(alice))
}
