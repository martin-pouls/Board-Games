package domain

import (
	"slices"
	"testing"
)

func TestSelectAndStartMission(t *testing.T) {
	g := newCrew(t, 1)
	mustFail(t, g, ErrNotEnoughPlayers, func(g *Game) error { return g.StartMission(alice, noShuffle{}) })
	g = newCrew(t, 2)
	if g.Mission != 1 || g.Edition != EditionPlanetNine {
		t.Fatalf("new game: mission %d edition %s", g.Mission, g.Edition)
	}
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.SelectMission(bob, 3) })
	mustFail(t, g, ErrInvalidMission, func(g *Game) error { return g.SelectMission(alice, 0) })
	mustFail(t, g, ErrInvalidMission, func(g *Game) error { return g.SelectMission(alice, 51) })
	must(t, g.SelectMission(alice, 50))
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.StartMission(bob, noShuffle{}) })
	must(t, g.StartMission(alice, noShuffle{}))
	if g.Phase != PhasePlaying || g.Attempt.Number != 1 || len(eventsOf[AttemptStarted](g)) != 1 {
		t.Fatalf("phase %s attempt %+v", g.Phase, g.Attempt)
	}
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.SelectMission(alice, 2) })
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.AddPlayer(carol, "Carol") })
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.Leave(bob) })
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.EndGame(alice) })
}

// lose gives up the current attempt.
func lose(t *testing.T, g *Game) {
	t.Helper()
	must(t, g.GiveUp(bob))
}

// win completes mission 1 for 4 players: alice leads her own task card p9 and wins.
func winMissionOne(t *testing.T, g *Game) {
	t.Helper()
	setHands(g, "t4 p9 b3", "p3 b2 g1", "g3 g4 y1", "y2 y3 t1")
	setTasks(g, "p9:0")
	playAll(t, g, "p9 p3 g3 y2")
	if g.Phase != PhaseDebrief || !g.Attempt.Result.Success {
		t.Fatalf("phase %s result %+v", g.Phase, g.Attempt.Result)
	}
}

func TestGiveUpAndNextAttempt(t *testing.T) {
	g := startMission(t, 4, 1)
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.GiveUp("nobody") })
	lose(t, g)
	r := g.Attempt.Result
	if g.Phase != PhaseDebrief || r.Success || r.Reason != ReasonGaveUp || r.Seat != 1 {
		t.Fatalf("phase %s result %+v", g.Phase, r)
	}
	if e := g.Logbook[1]; e.Attempts != 1 || e.Completed {
		t.Fatalf("logbook %+v", e)
	}
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.GiveUp(bob) })
	mustFail(t, g, ErrInvalidResultChoice, func(g *Game) error { return g.NextMission(alice, noShuffle{}) })
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.NextAttempt(bob, NextAttemptChoice{}, noShuffle{}) })
	must(t, g.NextAttempt(alice, NextAttemptChoice{}, noShuffle{}))
	if g.Phase != PhasePlaying || g.Attempt.Number != 2 || g.Attempt.Result != nil {
		t.Fatalf("phase %s attempt %d", g.Phase, g.Attempt.Number)
	}
	winMissionOne(t, g)
	if e := g.Logbook[1]; e.Attempts != 2 || !e.Completed || e.Total() != 2 {
		t.Fatalf("logbook %+v", e)
	}
	mustFail(t, g, ErrInvalidResultChoice, func(g *Game) error { return g.NextAttempt(alice, NextAttemptChoice{}, noShuffle{}) })
	must(t, g.NextMission(alice, noShuffle{}))
	if g.Mission != 2 || g.Attempt.Number != 1 || g.Phase != PhasePlaying {
		t.Fatalf("mission %d attempt %d", g.Mission, g.Attempt.Number)
	}
}

func TestDistressCountsInLogbook(t *testing.T) {
	g := withDistress(startMission(t, 4, 1))
	setHands(g, "t4 p9 b3 b4", "p3 b2 g1 g2", "g3 g4 y1 y4", "y2 y3 t1 t2")
	setTasks(g, "p9:0")
	must(t, g.SendDistress(alice, DirLeft))
	for i, card := range []string{"b4", "g2", "y4", "y3"} {
		must(t, g.ChoosePassCard(crewIDs[i], c(card)))
	}
	lose(t, g)
	if e := g.Logbook[1]; !e.Distress || e.Total() != 1 {
		t.Fatalf("the +1 counts only on completion: %+v", e)
	}
	if v, _ := g.ViewFor(bob); !v.Attempt.Distress {
		t.Fatal("the signal stays active until the mission is completed")
	}
	must(t, g.NextAttempt(alice, NextAttemptChoice{}, noShuffle{}))
	if g.Attempt.DistressDecided {
		t.Fatal("a new attempt may pass again")
	}
	winMissionOne(t, g)
	if e := g.Logbook[1]; e.Attempts != 2 || !e.Distress || e.Total() != 3 {
		t.Fatalf("logbook %+v", e)
	}
	// A replay of the completed mission starts without an active signal.
	must(t, g.BackToLobby(alice, false))
	must(t, g.SelectMission(alice, 1))
	must(t, g.StartMission(alice, noShuffle{}))
	if v, _ := g.ViewFor(bob); v.Attempt.Distress {
		t.Fatal("no active signal in a replay of a completed mission")
	}
}

func TestBackToLobbyAndReplay(t *testing.T) {
	g := startMission(t, 4, 1)
	winMissionOne(t, g)
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.BackToLobby(bob, false) })
	must(t, g.BackToLobby(alice, false))
	if g.Phase != PhaseLobby || g.Attempt != nil || g.Mission != 2 {
		t.Fatalf("phase %s mission %d", g.Phase, g.Mission)
	}
	// Players may change in the lobby.
	must(t, g.AddPlayer(erin, "Erin"))
	must(t, g.Leave(erin))
	// A replay of a completed mission does not touch its logbook line and counts from 1.
	must(t, g.SelectMission(alice, 1))
	must(t, g.StartMission(alice, noShuffle{}))
	if g.Attempt.Number != 1 {
		t.Fatalf("replay attempt %d", g.Attempt.Number)
	}
	lose(t, g)
	must(t, g.NextAttempt(alice, NextAttemptChoice{}, noShuffle{}))
	if g.Attempt.Number != 2 {
		t.Fatalf("replay attempt %d", g.Attempt.Number)
	}
	winMissionOne(t, g)
	if e := g.Logbook[1]; e.Attempts != 1 || !e.Completed {
		t.Fatalf("logbook changed by a replay: %+v", e)
	}
	// A lost replay of a completed mission keeps it selected (rules §16).
	must(t, g.BackToLobby(alice, false))
	must(t, g.SelectMission(alice, 1))
	must(t, g.StartMission(alice, noShuffle{}))
	lose(t, g)
	must(t, g.BackToLobby(alice, false))
	if g.Mission != 1 {
		t.Fatalf("after a lost replay the mission stays selected, got %d", g.Mission)
	}
	must(t, g.StartMission(alice, noShuffle{}))
	winMissionOne(t, g)
	// An unfinished mission continues its count from the lobby.
	must(t, g.BackToLobby(alice, false))
	if g.Mission != 2 {
		t.Fatalf("next unfinished %d", g.Mission)
	}
	must(t, g.StartMission(alice, noShuffle{}))
	lose(t, g)
	must(t, g.BackToLobby(alice, false))
	if g.Mission != 2 {
		t.Fatal("after a loss the mission stays selected")
	}
	must(t, g.StartMission(alice, noShuffle{}))
	if g.Attempt.Number != 2 {
		t.Fatalf("continued attempt %d", g.Attempt.Number)
	}
}

func TestNextMissionAfterTheLast(t *testing.T) {
	g := startMission(t, 4, 50)
	g.Mission = 50
	g.Attempt.Result = &AttemptResult{Success: true}
	g.Phase = PhaseDebrief
	mustFail(t, g, ErrInvalidMission, func(g *Game) error { return g.NextMission(alice, noShuffle{}) })
}

func TestEndGameFromDebrief(t *testing.T) {
	g := startMission(t, 2, 1)
	lose(t, g)
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.EndGame(bob) })
	must(t, g.EndGame(alice))
	if g.Phase != PhaseEnded {
		t.Fatal("ended")
	}
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.NextAttempt(alice, NextAttemptChoice{}, noShuffle{}) })
}

func TestCloneIsDeep(t *testing.T) {
	g := startMission(t, 2, 24)
	must(t, g.AnswerPoll(alice, true))
	c := g.Clone()
	c.Attempt.Members[0].Hand[0] = Card{SuitPink, 9}
	c.Attempt.Members[2].Stacks[0].Up.Rank = 9
	c.Attempt.Tasks[0].Objective.Cards[0].Rank = 9
	*c.Attempt.Poll.Answers[0] = false
	c.Logbook[7] = LogEntry{Attempts: 3}
	if g.Attempt.Members[0].Hand[0] == c.Attempt.Members[0].Hand[0] ||
		g.Attempt.Members[2].Stacks[0].Up.Rank == 9 ||
		g.Attempt.Tasks[0].Objective.Cards[0].Rank == 9 ||
		!*g.Attempt.Poll.Answers[0] || g.Logbook[7].Attempts != 0 {
		t.Fatal("clone shares state")
	}
	if !slices.Equal(g.Actors(), c.Actors()) {
		t.Fatal("clone differs")
	}
}
