package domain

import (
	"strconv"
	"testing"
)

// bbWin wins the running attempt (3+ humans) in one trick: the commander holds the only task
// card b9 and everybody else plays b1.
func bbWin(t *testing.T, g *Game) {
	t.Helper()
	a := g.Attempt
	n := len(a.Members)
	cmd := a.Commander
	for i := range a.Members {
		a.Members[i].Hand = cs("b1")
	}
	a.Members[cmd].Hand = cs("b9")
	setTasks(g, "b9:"+strconv.Itoa(cmd))
	a.Objectives = nil
	a.Commander = cmd
	a.Trick = Trick{Leader: cmd}
	for i := range n {
		seat := (cmd + i) % n
		must(t, g.PlayCard(bbID(g, seat), a.Members[seat].Hand[0], noShuffle{}))
	}
	bbExpectEnd(t, g, true, ReasonCompleted, 1)
}

// rules §15: give up — any player, counts as a lost attempt.
func TestRulesGiveUp(t *testing.T) {
	for _, mission := range []int{1, 5, 11, 50} { // draft, poll, pick, pick
		g := startMission(t, 4, mission)
		must(t, g.GiveUp(carol))
		bbExpectEnd(t, g, false, ReasonGaveUp, 0)
		if g.Attempt.Result.Seat != 2 {
			t.Fatalf("mission %d: seat %d", mission, g.Attempt.Result.Seat)
		}
		if ev := eventsOf[AttemptEnded](g); len(ev) != 1 || ev[0].Success || ev[0].Reason != ReasonGaveUp {
			t.Fatalf("events %v", ev)
		}
		if e := g.Logbook[mission]; e.Attempts != 1 || e.Completed {
			t.Fatalf("logbook %+v", e)
		}
		mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.GiveUp(alice) })
	}
	g := bbReady(t, 3, 1)
	must(t, g.PlayCard(bbCmd(g), g.LegalCards(g.Attempt.Commander)[0], noShuffle{}))
	must(t, g.GiveUp(bob)) // in the middle of a trick
	bbExpectEnd(t, g, false, ReasonGaveUp, 0)
	g = newCrew(t, 3)
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.GiveUp(alice) })
	g = startMission(t, 3, 1)
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.GiveUp("nobody") })
}

// rules §15: the logbook counts attempts up to the completion and keeps the line of a completed
// mission when it is replayed.
func TestRulesLogbook(t *testing.T) {
	g := startMission(t, 4, 1)
	must(t, g.GiveUp(alice))
	if e := g.Logbook[1]; e.Attempts != 1 || e.Completed || e.Distress || e.Total() != 1 {
		t.Fatalf("after attempt 1: %+v", e)
	}
	must(t, g.NextAttempt(alice, noShuffle{}))
	if g.Attempt.Number != 2 {
		t.Fatalf("attempt %d", g.Attempt.Number)
	}
	must(t, g.GiveUp(alice))
	must(t, g.NextAttempt(alice, noShuffle{}))
	bbDraft(t, g)
	bbWin(t, g)
	if e := g.Logbook[1]; e.Attempts != 3 || !e.Completed || e.Total() != 3 {
		t.Fatalf("after the win: %+v", e)
	}
	// back to the lobby: the next unfinished mission is selected
	must(t, g.BackToLobby(alice))
	if g.Phase != PhaseLobby || g.Mission != 2 || g.Attempt != nil {
		t.Fatalf("phase %s mission %d", g.Phase, g.Mission)
	}
	if len(eventsOf[BackInLobby](g)) != 1 {
		t.Fatal("event")
	}
	// replaying the completed mission starts at attempt 1 and does not change its line
	must(t, g.SelectMission(alice, 1))
	must(t, g.StartMission(alice, noShuffle{}))
	if g.Attempt.Number != 1 {
		t.Fatalf("replay starts at attempt %d", g.Attempt.Number)
	}
	must(t, g.GiveUp(alice))
	if e := g.Logbook[1]; e.Attempts != 3 || !e.Completed {
		t.Fatalf("replay lost: %+v", e)
	}
	must(t, g.NextAttempt(alice, noShuffle{}))
	bbDraft(t, g)
	bbWin(t, g)
	if e := g.Logbook[1]; e.Attempts != 3 || !e.Completed || e.Distress {
		t.Fatalf("replay won: %+v", e)
	}
	// an unfinished mission keeps its attempts across the lobby
	must(t, g.BackToLobby(alice))
	must(t, g.SelectMission(alice, 3))
	must(t, g.StartMission(alice, noShuffle{}))
	must(t, g.GiveUp(alice))
	must(t, g.BackToLobby(alice))
	if g.Mission != 3 {
		t.Fatalf("after a loss the mission stays selected: %d", g.Mission)
	}
	if e := g.Logbook[3]; e.Attempts != 1 || e.Completed {
		t.Fatalf("logbook 3: %+v", e)
	}
	must(t, g.StartMission(alice, noShuffle{}))
	if g.Attempt.Number != 2 {
		t.Fatalf("an unfinished mission continues counting: attempt %d", g.Attempt.Number)
	}
	bbDraft(t, g)
	bbWin(t, g)
	if e := g.Logbook[3]; e.Attempts != 2 || !e.Completed {
		t.Fatalf("logbook 3: %+v", e)
	}
	// 1 and 3 are done: the lobby preselects 2
	must(t, g.BackToLobby(alice))
	if g.Mission != 2 {
		t.Fatalf("preselected %d, want 2", g.Mission)
	}
}

// rules §9 / §15: a signal used on a replay of a completed mission does not change its line.
func TestRulesLogbookReplayWithDistress(t *testing.T) {
	g := withDistress(bbReady(t, 3, 1))
	bbWin(t, g)
	before := g.Logbook[1]
	must(t, g.BackToLobby(alice))
	must(t, g.SelectMission(alice, 1))
	must(t, g.StartMission(alice, noShuffle{}))
	bbDraft(t, g)
	must(t, g.SendDistress(bbCmd(g), DirLeft))
	for s := range 3 {
		must(t, g.ChoosePassCard(bbID(g, s), bbNonTrump(g, s)))
	}
	bbWin(t, g)
	if g.Logbook[1] != before {
		t.Fatalf("logbook %+v, want %+v", g.Logbook[1], before)
	}
}

// rules §16: after an attempt the host chooses; the transitions and their errors.
func TestRulesAfterAttempt(t *testing.T) {
	lost := func(t *testing.T) *Game {
		g := startMission(t, 4, 1)
		must(t, g.GiveUp(alice))
		return g
	}
	won := func(t *testing.T) *Game {
		g := bbReady(t, 4, 1)
		bbWin(t, g)
		return g
	}
	t.Run("next attempt after a loss", func(t *testing.T) {
		g := lost(t)
		before := g.Attempt.Members[0].Hand
		g.PullEvents()
		must(t, g.NextAttempt(alice, &fixedIntN{}))
		a := g.Attempt
		if g.Phase != PhasePlaying || a.Number != 2 || len(a.Tricks) != 0 || a.Result != nil || a.Step != StepDraft {
			t.Fatalf("phase %s attempt %+v", g.Phase, a)
		}
		_ = before
		if ev := eventsOf[AttemptStarted](g); len(ev) != 1 || ev[0].Number != 2 || ev[0].Mission != 1 {
			t.Fatalf("events %v", ev)
		}
		if g.Mission != 1 {
			t.Fatal("same mission")
		}
	})
	t.Run("next attempt is refused after a win", func(t *testing.T) {
		g := won(t)
		mustFail(t, g, ErrInvalidResultChoice, func(g *Game) error { return g.NextAttempt(alice, noShuffle{}) })
	})
	t.Run("next mission after a win", func(t *testing.T) {
		g := won(t)
		g.PullEvents()
		must(t, g.NextMission(alice, noShuffle{}))
		if g.Mission != 2 || g.Phase != PhasePlaying || g.Attempt.Number != 1 || len(g.Attempt.Tasks) != 2 {
			t.Fatalf("mission %d attempt %+v", g.Mission, g.Attempt)
		}
		if ev := eventsOf[AttemptStarted](g); len(ev) != 1 || ev[0].Mission != 2 || ev[0].Number != 1 {
			t.Fatalf("events %v", ev)
		}
	})
	t.Run("next mission is refused after a loss", func(t *testing.T) {
		g := lost(t)
		mustFail(t, g, ErrInvalidResultChoice, func(g *Game) error { return g.NextMission(alice, noShuffle{}) })
	})
	t.Run("next mission after the last one", func(t *testing.T) {
		g := newCrew(t, 3)
		must(t, g.SelectMission(alice, 50))
		must(t, g.StartMission(alice, noShuffle{}))
		bbWin(t, g)
		mustFail(t, g, ErrInvalidMission, func(g *Game) error { return g.NextMission(alice, noShuffle{}) })
		must(t, g.BackToLobby(alice))
		if g.Mission != 1 {
			t.Fatalf("preselected %d, want the first unfinished mission 1", g.Mission)
		}
	})
	t.Run("only the host", func(t *testing.T) {
		g := won(t)
		mustFail(t, g, ErrNotHost, func(g *Game) error { return g.NextMission(bob, noShuffle{}) })
		mustFail(t, g, ErrNotHost, func(g *Game) error { return g.BackToLobby(bob) })
		mustFail(t, g, ErrNotHost, func(g *Game) error { return g.EndGame(bob) })
		g = lost(t)
		mustFail(t, g, ErrNotHost, func(g *Game) error { return g.NextAttempt(bob, noShuffle{}) })
		mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.NextAttempt("nobody", noShuffle{}) })
	})
	t.Run("wrong phase", func(t *testing.T) {
		g := startMission(t, 4, 1)
		mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.NextAttempt(alice, noShuffle{}) })
		mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.NextMission(alice, noShuffle{}) })
		mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.BackToLobby(alice) })
		mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.EndGame(alice) })
		mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.StartMission(alice, noShuffle{}) })
		mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.SelectMission(alice, 2) })
	})
	t.Run("back to the lobby after a loss keeps the mission", func(t *testing.T) {
		g := lost(t)
		must(t, g.BackToLobby(alice))
		if g.Mission != 1 || g.Phase != PhaseLobby || g.Attempt != nil {
			t.Fatalf("mission %d phase %s", g.Mission, g.Phase)
		}
	})
	t.Run("only the lobby lets players come and go", func(t *testing.T) {
		g := won(t)
		mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.AddPlayer(erin, "Erin") })
		mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.Leave(bob) })
		must(t, g.BackToLobby(alice))
		must(t, g.AddPlayer(erin, "Erin"))
		must(t, g.Leave(bob))
	})
	t.Run("the host may end the session after an attempt", func(t *testing.T) {
		g := lost(t)
		must(t, g.EndGame(alice))
		if g.Phase != PhaseEnded {
			t.Fatal(g.Phase)
		}
		v, err := g.ViewFor(bob)
		must(t, err)
		if v.Phase != PhaseEnded {
			t.Fatal(v.Phase)
		}
	})
}

// rules §16: lobby — host picks any mission 1-50 and starts it with 2-5 players.
func TestRulesLobbyMissions(t *testing.T) {
	g := newCrew(t, 3)
	for _, n := range []int{0, -1, 51, 100} {
		n := n
		mustFail(t, g, ErrInvalidMission, func(g *Game) error { return g.SelectMission(alice, n) })
	}
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.SelectMission(bob, 5) })
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.StartMission(bob, noShuffle{}) })
	for _, n := range []int{1, 25, 50} {
		must(t, g.SelectMission(alice, n))
		if g.Mission != n {
			t.Fatalf("mission %d", g.Mission)
		}
	}
	solo, err := NewGame("G", alice, "Alice")
	must(t, err)
	mustFail(t, solo, ErrNotEnoughPlayers, func(g *Game) error { return g.StartMission(alice, noShuffle{}) })
	if solo.Mission != 1 {
		t.Fatalf("mission 1 is preselected in a new game, got %d", solo.Mission)
	}
	for n := 2; n <= 5; n++ {
		for _, m := range []int{1, 20, 50} {
			g := newCrew(t, n)
			must(t, g.SelectMission(alice, m))
			must(t, g.StartMission(alice, noShuffle{}))
			if g.Phase != PhasePlaying {
				t.Fatalf("%d players mission %d", n, m)
			}
			wantMembers := n
			if n == 2 {
				wantMembers = 3
			}
			if len(g.Attempt.Members) != wantMembers {
				t.Fatalf("members %d", len(g.Attempt.Members))
			}
		}
	}
}
