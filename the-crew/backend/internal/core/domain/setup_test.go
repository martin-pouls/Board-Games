package domain

import (
	"slices"
	"testing"
)

// With noShuffle the deal is fixed: 4 players → dave (seat 3) holds y4..y9 t1..t4 and is the
// commander; 2 players → bob (seat 1) holds y1..y9 t1..t4; 5 players → erin (seat 4). The tasks
// are the first colour cards: p1, p2, p3, …

func TestCommanderWithNoShuffle(t *testing.T) {
	for n, want := range map[int]int{2: 1, 3: 2, 4: 3, 5: 4} {
		g := startMission(t, n, 1)
		if g.Attempt.Commander != want {
			t.Errorf("%d players: commander %d, want %d", n, g.Attempt.Commander, want)
		}
	}
}

func TestOpenDraftOrder(t *testing.T) {
	g := startMission(t, 2, 10) // 4 tasks, open; commander bob (seat 1), then the dummy, then alice
	a := g.Attempt
	if a.Step != StepDraft || a.DraftTurn() != 1 {
		t.Fatalf("step %s turn %d", a.Step, a.DraftTurn())
	}
	for _, task := range a.Tasks {
		if task.FaceDown || task.Owner() != OwnerUnassigned {
			t.Fatalf("open tasks must be face up and unassigned: %+v", task)
		}
	}
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.PickTask(alice, 0) })
	mustFail(t, g, ErrInvalidTask, func(g *Game) error { return g.PickTask(bob, 9) })
	must(t, g.PickTask(bob, 2))
	mustFail(t, g, ErrInvalidTask, func(g *Game) error { return g.PickTask(bob, 2) }) // taken
	// The dummy's turn: the commander picks for it.
	if a.DraftTurn() != 2 || !slices.Equal(g.Actors(), []int{1}) {
		t.Fatalf("dummy turn %d actors %v", a.DraftTurn(), g.Actors())
	}
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.PickTask(alice, 0) })
	must(t, g.PickTask(bob, 0))
	must(t, g.PickTask(alice, 1))
	must(t, g.PickTask(bob, 3))
	owners := []int{a.Tasks[0].Owner(), a.Tasks[1].Owner(), a.Tasks[2].Owner(), a.Tasks[3].Owner()}
	if !slices.Equal(owners, []int{2, 0, 1, 1}) {
		t.Fatalf("owners %v", owners)
	}
	if a.Step != StepPlay || a.Trick.Leader != 1 || !slices.Equal(g.Actors(), []int{1}) {
		t.Fatalf("after the draft: step %s leader %d", a.Step, a.Trick.Leader)
	}
	if n := len(eventsOf[TaskPicked](g)); n != 4 {
		t.Fatalf("%d TaskPicked events", n)
	}
}

func TestSwapTokensMission23(t *testing.T) {
	g := startMission(t, 4, 23)
	a := g.Attempt
	if a.Step != StepAdjustTokens || !slices.Equal(g.Actors(), []int{3}) {
		t.Fatalf("step %s actors %v", a.Step, g.Actors())
	}
	mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.AdjustTokens(alice, 0, 1) })
	mustFail(t, g, ErrInvalidTask, func(g *Game) error { return g.AdjustTokens(dave, 0, 0) })
	mustFail(t, g, ErrInvalidTask, func(g *Game) error { return g.AdjustTokens(dave, 0, 5) })
	mustFail(t, g, ErrWrongStep, func(g *Game) error { return g.PickTask(dave, 0) })
	must(t, g.AdjustTokens(dave, 0, 4))
	if a.Tasks[0].Token != Token5 || a.Tasks[4].Token != Token1 || a.Step != StepDraft {
		t.Fatalf("tokens %s %s step %s", a.Tasks[0].Token, a.Tasks[4].Token, a.Step)
	}
	mustFail(t, g, ErrWrongStep, func(g *Game) error { return g.SkipTokens(dave) }) // once per attempt

	g = startMission(t, 4, 23)
	must(t, g.SkipTokens(dave))
	if g.Attempt.Tasks[0].Token != Token1 || g.Attempt.Step != StepDraft {
		t.Fatal("skip must keep the tokens")
	}
}

func TestMoveTokenMission40(t *testing.T) {
	g := startMission(t, 4, 40)                                                                    // 8 tasks, tokens 1 2 3 on tasks 0..2
	mustFail(t, g, ErrInvalidTokenMove, func(g *Game) error { return g.AdjustTokens(dave, 0, 1) }) // target has a token
	mustFail(t, g, ErrInvalidTokenMove, func(g *Game) error { return g.AdjustTokens(dave, 5, 6) }) // source has none
	must(t, g.AdjustTokens(dave, 1, 7))
	a := g.Attempt
	if a.Tasks[1].Token != TokenNone || a.Tasks[7].Token != Token2 {
		t.Fatalf("tokens %s %s", a.Tasks[1].Token, a.Tasks[7].Token)
	}
}

func TestSilenceMission11(t *testing.T) {
	g := startMission(t, 2, 11)
	a := g.Attempt
	if a.Step != StepPick || g.Pick() != PickSilence {
		t.Fatalf("step %s pick %s", a.Step, g.Pick())
	}
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bob, []int{2}) }) // the dummy
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bob, []int{0, 1}) })
	mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.PickSeats(alice, []int{0}) })
	must(t, g.PickSeats(bob, []int{1})) // the commander may silence themselves
	if a.Silenced != 1 || a.Step != StepDraft {
		t.Fatalf("silenced %d step %s", a.Silenced, a.Step)
	}
}

func TestRolesMission50(t *testing.T) {
	g := startMission(t, 2, 50)
	if g.Pick() != PickRoles {
		t.Fatalf("pick %s", g.Pick())
	}
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bob, []int{1}) })
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bob, []int{2, 2}) })
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bob, []int{0, 3}) })
	must(t, g.PickSeats(bob, []int{2, 1})) // the dummy and the commander are allowed
	a := g.Attempt
	if a.Objectives[0].Owner != 2 || a.Objectives[1].Owner != 1 || a.Step != StepPlay {
		t.Fatalf("owners %d %d step %s", a.Objectives[0].Owner, a.Objectives[1].Owner, a.Step)
	}
}

func TestChosenMemberPollMission5(t *testing.T) {
	g := startMission(t, 4, 5)
	a := g.Attempt
	if a.Step != StepPoll || !slices.Equal(g.Actors(), []int{0, 1, 2}) {
		t.Fatalf("step %s actors %v", a.Step, g.Actors())
	}
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.AnswerPoll(dave, true) })
	mustFail(t, g, ErrWrongStep, func(g *Game) error { return g.PickSeats(dave, []int{0}) })
	must(t, g.AnswerPoll(bob, false))
	mustFail(t, g, ErrAlreadyAnswered, func(g *Game) error { return g.AnswerPoll(bob, true) })
	must(t, g.AnswerPoll(alice, true))
	if !slices.Equal(g.Actors(), []int{2}) || a.Step != StepPoll {
		t.Fatalf("actors %v", g.Actors())
	}
	// Answers are public as they come in.
	v, _ := g.ViewFor(carol)
	if p := v.Attempt.Poll; p == nil || p.Kind != PollWellIll || p.Objective != 0 || *p.Answers[1] || !*p.Answers[0] || p.Answers[2] != nil {
		t.Fatalf("poll view %+v", v.Attempt.Poll)
	}
	must(t, g.AnswerPoll(carol, true))
	if a.Step != StepPick || g.Pick() != PickObjective || !slices.Equal(g.Actors(), []int{3}) {
		t.Fatalf("step %s pick %s", a.Step, g.Pick())
	}
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(dave, []int{3}) }) // not themselves
	must(t, g.PickSeats(dave, []int{1}))
	if a.Objectives[0].Owner != 1 || a.Step != StepPlay || a.Poll != nil {
		t.Fatalf("owner %d step %s", a.Objectives[0].Owner, a.Step)
	}
}

func TestChosenMemberMayBeTheDummy(t *testing.T) {
	g := startMission(t, 2, 33)
	must(t, g.AnswerPoll(alice, true)) // the only poller: bob commands, the dummy does not answer
	must(t, g.PickSeats(bob, []int{2}))
	if g.Attempt.Objectives[0].Owner != 2 {
		t.Fatal("the dummy may be chosen")
	}
}

func TestCommanderDecisionMission20(t *testing.T) {
	g := startMission(t, 4, 20)
	a := g.Attempt
	for _, task := range a.Tasks {
		if !task.FaceDown {
			t.Fatal("tasks must be face down")
		}
	}
	for _, id := range []PlayerID{alice, bob, carol} {
		must(t, g.AnswerPoll(id, true))
	}
	if g.Pick() != PickDecision {
		t.Fatalf("pick %s", g.Pick())
	}
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(dave, []int{3}) })
	must(t, g.PickSeats(dave, []int{2}))
	for _, task := range a.Tasks {
		if task.FaceDown || task.Owner() != 2 {
			t.Fatalf("task %+v", task)
		}
	}
	if a.Step != StepPlay {
		t.Fatalf("step %s", a.Step)
	}
}

func TestCommanderDistributionMission24(t *testing.T) {
	g := startMission(t, 4, 24) // 6 tasks
	a := g.Attempt
	give := func(seat int) {
		t.Helper()
		for _, id := range []PlayerID{alice, bob, carol} {
			must(t, g.AnswerPoll(id, true))
		}
		must(t, g.PickSeats(dave, []int{seat}))
	}
	if a.Tasks[0].FaceDown || !a.Tasks[1].FaceDown {
		t.Fatalf("task 0 is revealed, the others are not: %+v", a.Tasks)
	}
	give(3) // the commander may take a task
	if a.Tasks[1].FaceDown || !a.Tasks[2].FaceDown {
		t.Fatal("the next task is revealed one at a time")
	}
	give(3)
	// 3 has 2, the others 0, 4 tasks left: giving 3 a third needs 3+... others to 2 = 6 > 3 left.
	for _, id := range []PlayerID{alice, bob, carol} {
		must(t, g.AnswerPoll(id, true))
	}
	mustFail(t, g, ErrUnevenDistribution, func(g *Game) error { return g.PickSeats(dave, []int{3}) })
	must(t, g.PickSeats(dave, []int{0}))
	give(1)
	// counts 1,1,0,2 with 2 left: giving 0 a second → 2,1,0,2: carol needs 1, 1 left: fine.
	give(0)
	// counts 2,1,0,2 with 1 left: only carol can take it.
	for _, id := range []PlayerID{alice, bob, carol} {
		must(t, g.AnswerPoll(id, true))
	}
	mustFail(t, g, ErrUnevenDistribution, func(g *Game) error { return g.PickSeats(dave, []int{1}) })
	must(t, g.PickSeats(dave, []int{2}))
	if a.Step != StepPlay {
		t.Fatalf("step %s", a.Step)
	}
}

func TestEvenReachable(t *testing.T) {
	a := &Attempt{Members: make([]Member, 3)}
	for _, owner := range []int{0, 0, OwnerUnassigned, OwnerUnassigned} {
		a.Tasks = append(a.Tasks, Task{Objective: Objective{Owner: owner}})
	}
	// counts 2,0,0 and this task plus one more: 3 for seat 0 is never even.
	if a.evenReachable(0) {
		t.Fatal("seat 0 must not get a third")
	}
	// seat 1 → 2,1,0 and one left for seat 2.
	if !a.evenReachable(1) {
		t.Fatal("seat 1 must be allowed")
	}
}

func TestMission46OwnerAndMission48Card(t *testing.T) {
	g := startMission(t, 4, 46) // noShuffle: alice holds p9 → bob owns the objective
	if o := g.Attempt.Objectives[0]; o.Owner != 1 || g.Attempt.Step != StepPlay {
		t.Fatalf("owner %d step %s", o.Owner, g.Attempt.Step)
	}
	g = startMission(t, 2, 46) // the dummy (seat 2) holds p9 face up → alice, the next seat, owns it
	if o := g.Attempt.Objectives[0]; o.Owner != 0 {
		t.Fatalf("owner %d", o.Owner)
	}
	g = startMission(t, 4, 48) // tasks p1 (Ω), p2, p3
	if o := g.Attempt.Objectives[0]; o.Owner != OwnerCrew || !slices.Equal(o.Cards, cs("p1")) {
		t.Fatalf("objective %+v", o)
	}
	if def := g.MissionDef(); len(def.Objectives[0].Objective.Cards) != 0 {
		t.Fatal("the mission table must not be modified")
	}
}

func TestMission34CommanderObjective(t *testing.T) {
	g := startMission(t, 5, 34)
	a := g.Attempt
	if a.Objectives[0].Owner != OwnerCrew || a.Objectives[1].Owner != 4 || a.Step != StepPlay {
		t.Fatalf("objectives %+v step %s", a.Objectives, a.Step)
	}
}
