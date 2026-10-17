package domain

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

// bbReady starts a mission with an open draft and completes the draft.
func bbReady(t *testing.T, n, mission int) *Game {
	t.Helper()
	g := startMission(t, n, mission)
	bbDraft(t, g)
	if g.Attempt.Step != StepPlay {
		t.Fatalf("mission %d: step %s after the draft, want play", mission, g.Attempt.Step)
	}
	return g
}

// bbID returns the player id of a human seat.
func bbID(g *Game, seat int) PlayerID { return g.Players[seat].ID }

// bbCmd is the commander's player id.
func bbCmd(g *Game) PlayerID { return bbID(g, g.Attempt.Commander) }

// bbAnswerAll lets every actor of the poll answer.
func bbAnswerAll(t *testing.T, g *Game, yes bool) {
	t.Helper()
	for _, s := range g.Actors() {
		must(t, g.AnswerPoll(bbID(g, s), yes))
	}
}

// bbCardOfSeat returns a card of the seat's hand that is not trump.
func bbNonTrump(g *Game, seat int) Card {
	for _, card := range g.Attempt.Members[seat].Hand {
		if !card.IsTrump() {
			return card
		}
	}
	panic("only trumps")
}

// ---------------------------------------------------------------- §9 distress

func TestRulesDistressWho(t *testing.T) {
	g := withDistress(bbReady(t, 4, 1))
	cmd := g.Attempt.Commander
	for seat := range 4 {
		if got := g.CanSendDistress(seat) == nil; got != (seat == cmd) {
			t.Errorf("CanSendDistress(%d) = %v (commander %d)", seat, got, cmd)
		}
		if seat != cmd {
			seat := seat
			mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.SendDistress(bbID(g, seat), DirLeft) })
			mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.SkipDistress(bbID(g, seat)) })
		}
	}
	v, _ := g.ViewFor(bbCmd(g))
	if !v.Attempt.Actions.SendDistress {
		t.Fatal("the commander is offered the distress signal")
	}
	v, _ = g.ViewFor(bbID(g, (cmd+1)%4))
	if v.Attempt.Actions.SendDistress {
		t.Fatal("others are not")
	}
	mustFail(t, g, ErrInvalidDirection, func(g *Game) error { return g.SendDistress(bbCmd(g), Direction("up")) })
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.SendDistress("nobody", DirLeft) })
}

func TestRulesDistressNotBeforeDistribution(t *testing.T) {
	g := withDistress(startMission(t, 4, 4)) // draft pending
	bbFailAny(t, g, func(g *Game) error { return g.SendDistress(bbCmd(g), DirLeft) }, ErrWrongStep, ErrNotAllowed)
	if g.CanSendDistress(g.Attempt.Commander) == nil {
		t.Fatal("no distress during the draft")
	}
}

// rules §9: direction and passing.
func TestRulesDistressPass(t *testing.T) {
	for _, tc := range []struct {
		players int
		dir     Direction
	}{
		{3, DirLeft}, {3, DirRight}, {4, DirLeft}, {4, DirRight}, {5, DirLeft}, {5, DirRight},
	} {
		t.Run(fmt.Sprintf("%dp %s", tc.players, tc.dir), func(t *testing.T) {
			n := tc.players
			g := withDistress(bbReady(t, n, 1))
			a := g.Attempt
			must(t, g.SendDistress(bbCmd(g), tc.dir))
			if a.Step != StepPass {
				t.Fatalf("step %s", a.Step)
			}
			if got := g.Actors(); len(got) != n {
				t.Fatalf("every human chooses a card: actors %v", got)
			}
			hands := make([][]Card, n)
			chosen := make([]Card, n)
			for s := range n {
				hands[s] = slices.Clone(a.Members[s].Hand)
				chosen[s] = bbNonTrump(g, s)
			}
			// trump may never be passed
			for s := range n {
				for _, card := range hands[s] {
					if card.IsTrump() {
						card := card
						s := s
						mustFail(t, g, ErrCannotPassTrump, func(g *Game) error { return g.ChoosePassCard(bbID(g, s), card) })
						break
					}
				}
			}
			// a card of somebody else
			other := hands[1][0]
			if slices.Contains(hands[0], other) {
				other = hands[1][len(hands[1])-1]
			}
			if !slices.Contains(hands[0], other) {
				mustFail(t, g, ErrCardNotInHand, func(g *Game) error { return g.ChoosePassCard(bbID(g, 0), other) })
			}
			for s := range n {
				if s == n-1 {
					// nothing moves before the last one has chosen; the others cannot see the choice
					for k := range n - 1 {
						if !slices.Equal(a.Members[k].Hand, hands[k]) {
							t.Fatal("cards moved early")
						}
					}
				}
				must(t, g.ChoosePassCard(bbID(g, s), chosen[s]))
				if s < n-1 {
					mustFail(t, g, ErrAlreadyAnswered, func(g *Game) error { return g.ChoosePassCard(bbID(g, s), chosen[s]) })
					// nobody else sees which card
					for o := range n {
						v, _ := g.ViewFor(bbID(g, o))
						if o > s && v.Attempt.Pass != nil && v.Attempt.Pass.Chosen != nil {
							t.Fatalf("player %d sees a pass choice", o)
						}
						if o <= s && (v.Attempt.Pass == nil || v.Attempt.Pass.Chosen == nil || *v.Attempt.Pass.Chosen != chosen[o]) {
							t.Fatalf("player %d does not see their own choice", o)
						}
						if v.Attempt.Pass != nil && v.Attempt.Pass.Direction != tc.dir {
							t.Fatalf("direction %s", v.Attempt.Pass.Direction)
						}
					}
				}
			}
			if a.Step != StepPlay {
				t.Fatalf("after the pass the step is %s", a.Step)
			}
			if len(eventsOf[PassCardChosen](g)) != n || len(eventsOf[CardsPassed](g)) != 1 || eventsOf[CardsPassed](g)[0].Direction != tc.dir {
				t.Fatalf("events %v", g.events)
			}
			for s := range n {
				to := (s + 1) % n // left = the next seat
				if tc.dir == DirRight {
					to = (s + n - 1) % n
				}
				if slices.Contains(a.Members[s].Hand, chosen[s]) && !slices.Contains(hands[s], chosen[s]) {
					t.Fatalf("seat %d still has the passed card", s)
				}
				if !slices.Contains(a.Members[to].Hand, chosen[s]) {
					t.Fatalf("seat %d did not receive %v from seat %d", to, chosen[s], s)
				}
				if len(a.Members[s].Hand) != len(hands[s]) {
					t.Fatalf("seat %d hand size changed", s)
				}
				if !slices.IsSortedFunc(a.Members[s].Hand, compareCards) {
					t.Fatal("hand not sorted")
				}
				// each player sees what they gave and got, others nothing
				v, _ := g.ViewFor(bbID(g, s))
				from := (s + n - 1) % n
				if tc.dir == DirRight {
					from = (s + 1) % n
				}
				ex := v.Attempt.Exchange
				if ex == nil || ex.Kind != ExchangePass || ex.Gave == nil || *ex.Gave != chosen[s] || ex.Got == nil || *ex.Got != chosen[from] {
					t.Fatalf("seat %d exchange view %+v", s, ex)
				}
			}
			// once per attempt
			if g.CanSendDistress(a.Commander) == nil {
				t.Fatal("a second signal in the same attempt")
			}
			bbFailAny(t, g, func(g *Game) error { return g.SendDistress(bbCmd(g), DirLeft) }, ErrNotAllowed, ErrWrongStep)
			bbFailAny(t, g, func(g *Game) error { return g.SkipDistress(bbCmd(g)) }, ErrNotAllowed, ErrWrongStep)
		})
	}
}

// rules §9 app decision: with two players the humans swap, JARVIS stays out; direction is irrelevant.
func TestRulesDistressTwoPlayers(t *testing.T) {
	for _, dir := range []Direction{DirLeft, DirRight} {
		g := withDistress(bbReady(t, 2, 1))
		a := g.Attempt
		dummyBefore := fmt.Sprint(a.Members[2].Stacks)
		must(t, g.SendDistress(bbCmd(g), dir))
		if got := g.Actors(); len(got) != 2 {
			t.Fatalf("actors %v", got)
		}
		ca, cb := bbNonTrump(g, 0), bbNonTrump(g, 1)
		ha, hb := slices.Clone(a.Members[0].Hand), slices.Clone(a.Members[1].Hand)
		must(t, g.ChoosePassCard(alice, ca))
		must(t, g.ChoosePassCard(bob, cb))
		if !slices.Contains(a.Members[1].Hand, ca) || !slices.Contains(a.Members[0].Hand, cb) {
			t.Fatalf("%s: the humans did not swap", dir)
		}
		if len(a.Members[0].Hand) != len(ha) || len(a.Members[1].Hand) != len(hb) {
			t.Fatal("hand sizes")
		}
		if fmt.Sprint(a.Members[2].Stacks) != dummyBefore {
			t.Fatal("JARVIS must not take part")
		}
	}
}

// rules §9: not after anyone communicated, not once a trick has started.
func TestRulesDistressTiming(t *testing.T) {
	t.Run("after a communication", func(t *testing.T) {
		g := withDistress(bbCommGame(t, 1))
		must(t, g.Communicate(alice, c("p9"), PosHighest))
		if g.CanSendDistress(3) == nil {
			t.Fatal("distress after a communication")
		}
		bbFailAny(t, g, func(g *Game) error { return g.SendDistress(dave, DirLeft) }, ErrNotAllowed, ErrCannotCommunicate)
	})
	t.Run("after the first card", func(t *testing.T) {
		g := withDistress(bbCommGame(t, 1))
		if g.CanSendDistress(3) != nil {
			t.Fatal("the commander may send before the lead")
		}
		must(t, g.PlayCard(dave, c("p4"), noShuffle{}))
		if g.CanSendDistress(3) == nil {
			t.Fatal("distress during trick 1")
		}
		bbFailAny(t, g, func(g *Game) error { return g.SendDistress(dave, DirLeft) }, ErrNotAllowed, ErrWrongStep)
	})
	t.Run("skipping closes the window", func(t *testing.T) {
		g := withDistress(bbReady(t, 4, 1))
		must(t, g.SkipDistress(bbCmd(g)))
		if g.CanSendDistress(g.Attempt.Commander) == nil {
			t.Fatal("distress after skipping")
		}
		bbFailAny(t, g, func(g *Game) error { return g.SendDistress(bbCmd(g), DirLeft) }, ErrNotAllowed, ErrWrongStep)
		bbFailAny(t, g, func(g *Game) error { return g.SkipDistress(bbCmd(g)) }, ErrNotAllowed, ErrWrongStep)
		if ev := eventsOf[DistressSignal](g); len(ev) != 1 || !ev[0].Skipped {
			t.Fatalf("events %v", ev)
		}
	})
	t.Run("the first card may be played without a decision", func(t *testing.T) {
		g := withDistress(bbReady(t, 4, 1))
		cmd := g.Attempt.Commander
		legal := g.LegalCards(cmd)
		must(t, g.PlayCard(bbCmd(g), legal[0], noShuffle{}))
	})
}

// rules §9 / §15: the logbook counts one more with the signal; it can be sent again later.
func TestRulesDistressLogbook(t *testing.T) {
	g := withDistress(bbReady(t, 3, 1))
	must(t, g.SendDistress(bbCmd(g), DirLeft))
	for s := range 3 {
		must(t, g.ChoosePassCard(bbID(g, s), bbNonTrump(g, s)))
	}
	must(t, g.GiveUp(alice))
	must(t, g.NextAttempt(alice, NextAttemptChoice{}, noShuffle{}))
	bbDraft(t, g)
	// allowed again on a later attempt, and not forced
	if g.CanSendDistress(g.Attempt.Commander) != nil {
		t.Fatal("distress is available again on the next attempt")
	}
	must(t, g.SkipDistress(bbCmd(g)))
	// win attempt 2 in one trick
	a := g.Attempt
	cmd := a.Commander
	for i := range a.Members {
		a.Members[i].Hand = cs("b1")
	}
	a.Members[cmd].Hand = cs("b9")
	setTasks(g, "b9:"+fmt.Sprint(cmd))
	a.Commander = cmd
	a.Trick = Trick{Leader: cmd}
	for i := range 3 {
		seat := (cmd + i) % 3
		must(t, g.PlayCard(bbID(g, seat), a.Members[seat].Hand[0], noShuffle{}))
	}
	bbExpectEnd(t, g, true, ReasonCompleted, 1)
	e := g.Logbook[1]
	if !e.Completed || !e.Distress || e.Attempts != 2 || e.Total() != 3 {
		t.Fatalf("logbook %+v total %d, want 2 attempts + 1 for the signal", e, e.Total())
	}
}

// ---------------------------------------------------------------- §10 polls and picks

func TestRulesPollChosenMember(t *testing.T) {
	for _, tc := range []struct {
		mission int
		kind    PollKind
		kindObj ObjectiveKind
	}{
		{5, PollWellIll, KindWinNoTricks},
		{33, PollYesNo, KindWinExactlyTricks},
		{41, PollYesNo, KindWinOnlyTricks},
	} {
		t.Run(fmt.Sprintf("mission %d", tc.mission), func(t *testing.T) {
			g := startMission(t, 4, tc.mission)
			a := g.Attempt
			cmd := a.Commander
			if a.Step != StepPoll {
				t.Fatalf("step %s", a.Step)
			}
			wantActors := []int{}
			for s := range 4 {
				if s != cmd {
					wantActors = append(wantActors, s)
				}
			}
			if got := g.Actors(); !slices.Equal(got, wantActors) {
				t.Fatalf("actors %v, want %v (everybody but the commander)", got, wantActors)
			}
			v, _ := g.ViewFor(bbID(g, wantActors[0]))
			if v.Attempt.Poll == nil || v.Attempt.Poll.Kind != tc.kind || v.Attempt.Poll.Objective != 0 {
				t.Fatalf("poll view %+v", v.Attempt.Poll)
			}
			// the commander does not answer
			bbFailAny(t, g, func(g *Game) error { return g.AnswerPoll(bbCmd(g), true) }, ErrNotAllowed, ErrNotYourTurn)
			bbFailAny(t, g, func(g *Game) error { return g.AnswerPoll("nobody", true) }, ErrUnknownPlayer)
			// answers come in any order and are visible at once
			order := []int{wantActors[2], wantActors[0], wantActors[1]}
			for k, s := range order {
				must(t, g.AnswerPoll(bbID(g, s), k%2 == 0))
				if k < 2 {
					mustFail(t, g, ErrAlreadyAnswered, func(g *Game) error { return g.AnswerPoll(bbID(g, s), true) })
					for obs := range 4 {
						v, _ := g.ViewFor(bbID(g, obs))
						ans := v.Attempt.Poll.Answers
						if ans[s] == nil || *ans[s] != (k%2 == 0) {
							t.Fatalf("observer %d does not see seat %d's answer: %v", obs, s, ans)
						}
						if ans[cmd] != nil {
							t.Fatal("the commander has no answer")
						}
					}
				}
			}
			if ev := eventsOf[PollAnswered](g); len(ev) != 3 || ev[0].Seat != order[0] || !ev[0].Yes || ev[1].Yes {
				t.Fatalf("events %v", ev)
			}
			// then the commander picks one other member
			if a.Step != StepPick || g.Pick() != PickObjective {
				t.Fatalf("step %s pick %s", a.Step, g.Pick())
			}
			if got := g.Actors(); !slices.Equal(got, []int{cmd}) {
				t.Fatalf("actors %v", got)
			}
			bbFailAny(t, g, func(g *Game) error { return g.AnswerPoll(bbID(g, wantActors[0]), true) }, ErrWrongStep, ErrNotAllowed)
			mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.PickSeats(bbID(g, wantActors[0]), []int{wantActors[1]}) })
			mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{cmd}) }) // never the commander
			mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{7}) })
			mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{-1}) })
			mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), nil) })
			mustFail(t, g, ErrInvalidSeat, func(g *Game) error {
				return g.PickSeats(bbCmd(g), []int{wantActors[0], wantActors[1]})
			})
			must(t, g.PickSeats(bbCmd(g), []int{wantActors[1]}))
			if a.Objectives[0].Owner != wantActors[1] || a.Objectives[0].Kind != tc.kindObj {
				t.Fatalf("objective %+v", a.Objectives[0])
			}
			if a.Step != StepPlay {
				t.Fatalf("step %s, want play (no tasks to draft)", a.Step)
			}
			if ev := eventsOf[Picked](g); len(ev) != 1 || !slices.Equal(ev[0].Seats, []int{wantActors[1]}) {
				t.Fatalf("events %v", ev)
			}
			bbFailAny(t, g, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{wantActors[0]}) }, ErrWrongStep)
			if err := g.Attempt.Poll; err != nil {
				t.Fatal("the poll must be over")
			}
		})
	}
}

// rules §10 app decisions: JARVIS does not answer but may be chosen.
func TestRulesPollTwoPlayersDummy(t *testing.T) {
	g := startMission(t, 2, 5)
	a := g.Attempt
	cmd := a.Commander
	other := 1 - cmd
	if got := g.Actors(); !slices.Equal(got, []int{other}) {
		t.Fatalf("actors %v, want only the other human", got)
	}
	must(t, g.AnswerPoll(bbID(g, other), false))
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{cmd}) })
	must(t, g.PickSeats(bbCmd(g), []int{2})) // JARVIS
	if a.Objectives[0].Owner != 2 {
		t.Fatalf("owner %d", a.Objectives[0].Owner)
	}
	// the dummy is a valid owner of the "win no trick" objective; the other human may be chosen too
	g = startMission(t, 2, 33)
	must(t, g.AnswerPoll(bbID(g, 1-g.Attempt.Commander), true))
	must(t, g.PickSeats(bbCmd(g), []int{1 - g.Attempt.Commander}))
	if g.Attempt.Objectives[0].Owner != 1-g.Attempt.Commander {
		t.Fatal("owner")
	}
}

// rules §10: commander's decision (20, 27, 37): all tasks, face down, to one other member.
func TestRulesCommanderDecision(t *testing.T) {
	for _, tc := range []struct{ players, mission, tasks int }{{3, 20, 2}, {4, 20, 2}, {4, 27, 3}, {5, 37, 4}, {2, 20, 2}, {2, 27, 3}} {
		t.Run(fmt.Sprintf("%dp mission %d", tc.players, tc.mission), func(t *testing.T) {
			g := startMission(t, tc.players, tc.mission)
			a := g.Attempt
			cmd := a.Commander
			if len(a.Tasks) != tc.tasks || a.Step != StepPoll {
				t.Fatalf("tasks %d step %s", len(a.Tasks), a.Step)
			}
			// face down: nobody sees the cards, not even the commander
			for s := range tc.players {
				v, _ := g.ViewFor(bbID(g, s))
				for _, tv := range v.Attempt.Tasks {
					if tv.Card != nil || !tv.FaceDown {
						t.Fatalf("seat %d sees a face-down task: %+v", s, tv)
					}
				}
			}
			bbAnswerAll(t, g, true)
			if a.Step != StepPick || g.Pick() != PickDecision {
				t.Fatalf("step %s pick %s", a.Step, g.Pick())
			}
			mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{cmd}) })
			mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{99}) })
			target := (cmd + 1) % len(a.Members) // in a 2-player game this is JARVIS
			if tc.players == 2 {
				target = 2
			}
			notCmd := (cmd + 1) % tc.players
			mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.PickSeats(bbID(g, notCmd), []int{target}) })
			must(t, g.PickSeats(bbCmd(g), []int{target}))
			for _, tk := range a.Tasks {
				if tk.Owner() != target || tk.FaceDown {
					t.Fatalf("task %+v: owner %d, want %d and revealed", tk, tk.Owner(), target)
				}
			}
			if a.Step != StepPlay {
				t.Fatalf("step %s", a.Step)
			}
			for s := range tc.players {
				v, _ := g.ViewFor(bbID(g, s))
				for _, tv := range v.Attempt.Tasks {
					if tv.Card == nil {
						t.Fatalf("seat %d: revealed task without card", s)
					}
				}
			}
		})
	}
}

func TestRulesCommanderDecisionOtherHuman(t *testing.T) {
	g := startMission(t, 4, 27)
	bbAnswerAll(t, g, false)
	cmd := g.Attempt.Commander
	target := (cmd + 2) % 4
	must(t, g.PickSeats(bbCmd(g), []int{target}))
	for _, tk := range g.Attempt.Tasks {
		if tk.Owner() != target {
			t.Fatalf("owner %d", tk.Owner())
		}
	}
}

// rules §10: commander's distribution (24, 32, 36, 43): one task at a time, poll, give.
func TestRulesCommanderDistribution(t *testing.T) {
	g := startMission(t, 4, 24)
	a := g.Attempt
	cmd := a.Commander
	if len(a.Tasks) != 6 {
		t.Fatalf("tasks %d", len(a.Tasks))
	}
	// only the first task is revealed at the start
	for i, tk := range a.Tasks {
		if tk.FaceDown != (i != 0) {
			t.Fatalf("task %d face down = %v", i, tk.FaceDown)
		}
	}
	for s := range 4 {
		v, _ := g.ViewFor(bbID(g, s))
		for i, tv := range v.Attempt.Tasks {
			if (tv.Card != nil) != (i == 0) {
				t.Fatalf("seat %d: task %d card visible = %v", s, i, tv.Card != nil)
			}
		}
	}
	// 6 tasks, 4 members: nobody may end with 3 (the others could not reach 2)
	give := func(task, seat int) error {
		if a.Step != StepPoll {
			t.Fatalf("task %d: step %s, want poll", task, a.Step)
		}
		v, _ := g.ViewFor(bbCmd(g))
		if v.Attempt.Poll == nil || v.Attempt.Poll.Task != task {
			t.Fatalf("poll view %+v, want task %d", v.Attempt.Poll, task)
		}
		bbAnswerAll(t, g, task%2 == 0)
		if g.Pick() != PickDistribute {
			t.Fatalf("pick %s", g.Pick())
		}
		return g.PickSeats(bbCmd(g), []int{seat})
	}
	must(t, give(0, 0))
	if a.Tasks[0].Owner() != 0 {
		t.Fatal("owner of task 0")
	}
	if a.Tasks[1].FaceDown {
		// the next task is revealed when its turn comes
		v, _ := g.ViewFor(bbCmd(g))
		if v.Attempt.Tasks[1].Card == nil {
			t.Fatal("task 1 must be revealed now")
		}
	}
	must(t, give(1, 0))
	// seat 0 would have 3 tasks: an even split is unreachable
	bbAnswerAll(t, g, true)
	mustFail(t, g, ErrUnevenDistribution, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{0}) })
	must(t, g.PickSeats(bbCmd(g), []int{1}))
	must(t, give(3, 1))
	// 0:2 1:2 so far; two tasks left for seats 2 and 3 (one each)... a third for seat 1 is rejected
	bbAnswerAll(t, g, true)
	mustFail(t, g, ErrUnevenDistribution, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{1}) })
	mustFail(t, g, ErrUnevenDistribution, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{0}) })
	must(t, g.PickSeats(bbCmd(g), []int{cmd}))
	_ = cmd
}

func TestRulesCommanderDistributionFull(t *testing.T) {
	// all of: the commander and JARVIS may receive tasks; 2 players = 3 members, 6 tasks: 2 each
	g := startMission(t, 2, 24)
	a := g.Attempt
	counts := map[int]int{}
	for i := range a.Tasks {
		if a.Step != StepPoll {
			t.Fatalf("task %d: step %s", i, a.Step)
		}
		if got := g.Actors(); len(got) != 1 || got[0] == a.Commander {
			t.Fatalf("actors %v: only the other human answers", got)
		}
		bbAnswerAll(t, g, true)
		seat := []int{2, 2, 0, 0, 1, 1}[i]
		if i == 2 {
			// a third for JARVIS cannot happen, but giving task 2 to JARVIS would be 3: reject
			mustFail(t, g, ErrUnevenDistribution, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{2}) })
		}
		must(t, g.PickSeats(bbCmd(g), []int{seat}))
		counts[seat]++
	}
	if a.Step != StepPlay {
		t.Fatalf("step %s", a.Step)
	}
	for _, tk := range a.Tasks {
		if tk.Owner() < 0 || tk.FaceDown {
			t.Fatalf("task %+v", tk)
		}
	}
	if counts[0] != 2 || counts[1] != 2 || counts[2] != 2 {
		t.Fatalf("counts %v", counts)
	}
}

// rules §10: 7 tasks among 3 members must end 3-2-2; a second member with 3 is unreachable.
func TestRulesCommanderDistributionSevenAmongThree(t *testing.T) {
	g := startMission(t, 3, 32)
	a := g.Attempt
	seq := []int{0, 0, 0, 1, 1}
	for i, seat := range seq {
		bbAnswerAll(t, g, true)
		must(t, g.PickSeats(bbCmd(g), []int{seat}))
		_ = i
	}
	bbAnswerAll(t, g, true)
	// seat 0 has 3, seat 1 has 2: seat 1 getting a 3rd leaves (3,3,1)
	mustFail(t, g, ErrUnevenDistribution, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{1}) })
	mustFail(t, g, ErrUnevenDistribution, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{0}) })
	must(t, g.PickSeats(bbCmd(g), []int{2}))
	bbAnswerAll(t, g, true)
	mustFail(t, g, ErrUnevenDistribution, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{0}) })
	must(t, g.PickSeats(bbCmd(g), []int{2}))
	if a.Step != StepPlay {
		t.Fatalf("step %s", a.Step)
	}
	cnt := [3]int{}
	for _, tk := range a.Tasks {
		cnt[tk.Owner()]++
	}
	if cnt != [3]int{3, 2, 2} {
		t.Fatalf("counts %v", cnt)
	}
}

// rules §10: tokens of mission 36 go with the first two revealed cards.
func TestRulesCommanderDistributionTokens(t *testing.T) {
	g := startMission(t, 4, 36)
	a := g.Attempt
	if a.Tasks[0].Token != Token("1") || a.Tasks[1].Token != Token("2") {
		t.Fatalf("tokens %q %q", a.Tasks[0].Token, a.Tasks[1].Token)
	}
	for i := 2; i < len(a.Tasks); i++ {
		if a.Tasks[i].Token != TokenNone {
			t.Fatalf("task %d has a token", i)
		}
	}
	// the token travels with the card when it is given
	bbAnswerAll(t, g, true)
	must(t, g.PickSeats(bbCmd(g), []int{2}))
	if a.Tasks[0].Owner() != 2 || a.Tasks[0].Token != Token("1") {
		t.Fatalf("task 0 %+v", a.Tasks[0])
	}
}

// rules §10: wrong step.
func TestRulesPickWrongStep(t *testing.T) {
	g := startMission(t, 4, 1)
	bbFailAny(t, g, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{0}) }, ErrWrongStep)
	bbFailAny(t, g, func(g *Game) error { return g.AnswerPoll(alice, true) }, ErrWrongStep, ErrNotAllowed)
	if g.Pick() != PickNone {
		t.Fatalf("pick %s", g.Pick())
	}
}

// rules §8 mission 11: the commander silences a human (the commander included, never JARVIS).
func TestRulesSilencedMember(t *testing.T) {
	g := startMission(t, 4, 11)
	a := g.Attempt
	cmd := a.Commander
	if a.Step != StepPick || g.Pick() != PickSilence || !slices.Equal(g.Actors(), []int{cmd}) {
		t.Fatalf("step %s pick %s actors %v", a.Step, g.Pick(), g.Actors())
	}
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{9}) })
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{0, 1}) })
	mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.PickSeats(bbID(g, (cmd+1)%4), []int{0}) })
	must(t, g.PickSeats(bbCmd(g), []int{1}))
	if a.Silenced != 1 || a.Step != StepDraft {
		t.Fatalf("silenced %d step %s", a.Silenced, a.Step)
	}
	bbDraft(t, g)
	for seat := range 4 {
		if got := g.CanCommunicate(seat) == nil; got != (seat != 1) {
			t.Errorf("seat %d may communicate = %v", seat, got)
		}
	}
	card := a.Members[1].Hand[0]
	if pos, ok := StatementFor(a.Members[1].Hand, card); ok {
		bbFailAny(t, g, func(g *Game) error { return g.Communicate(bob, card, pos) }, errCommNotAllowed...)
	}
	v, _ := g.ViewFor(bob)
	if v.Attempt.Silenced != 1 || len(v.Attempt.Actions.Communicate) != 0 {
		t.Fatalf("view %+v", v.Attempt.Actions)
	}
	// the commander can silence themselves
	g = startMission(t, 4, 11)
	must(t, g.PickSeats(bbCmd(g), []int{g.Attempt.Commander}))
	bbDraft(t, g)
	if g.CanCommunicate(g.Attempt.Commander) == nil {
		t.Fatal("the silenced commander must not communicate")
	}
	// JARVIS cannot be silenced
	g = startMission(t, 2, 11)
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{2}) })
}

// rules §13 mission 50: the commander assigns two different members to roles A and B.
func TestRulesRolesMission50(t *testing.T) {
	g := startMission(t, 4, 50)
	a := g.Attempt
	cmd := a.Commander
	if a.Step != StepPick || g.Pick() != PickRoles {
		t.Fatalf("step %s pick %s", a.Step, g.Pick())
	}
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{1}) })
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{1, 1}) })
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{1, 2, 3}) })
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.PickSeats(bbCmd(g), []int{1, 8}) })
	mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.PickSeats(bbID(g, (cmd+1)%4), []int{1, 2}) })
	must(t, g.PickSeats(bbCmd(g), []int{cmd, 1})) // the commander may take a role
	if a.Objectives[0].Owner != cmd || a.Objectives[1].Owner != 1 {
		t.Fatalf("owners %d %d", a.Objectives[0].Owner, a.Objectives[1].Owner)
	}
	if a.Step != StepPlay {
		t.Fatalf("step %s", a.Step)
	}
	// JARVIS may take a role
	g = startMission(t, 2, 50)
	must(t, g.PickSeats(bbCmd(g), []int{2, 0}))
	if g.Attempt.Objectives[0].Owner != 2 || g.Attempt.Objectives[1].Owner != 0 {
		t.Fatalf("owners %d %d", g.Attempt.Objectives[0].Owner, g.Attempt.Objectives[1].Owner)
	}
}

// rules §13: mission 46 — the left neighbour of the pink 9's holder, fixed at the deal.
func TestRulesMission46Owner(t *testing.T) {
	for _, n := range []int{2, 3, 4, 5} {
		for seed := range uint64(20) {
			g := newCrew(t, n)
			must(t, g.SelectMission(alice, 46))
			must(t, g.StartMission(alice, rand.New(rand.NewPCG(seed, 5))))
			a := g.Attempt
			holder := -1
			for i, m := range a.Members {
				if slices.Contains(m.Cards(), c("p9")) {
					holder = i
				}
			}
			want := (holder + 1) % len(a.Members)
			if a.Objectives[0].Kind != KindWinAllOfSuit || a.Objectives[0].Owner != want || a.Objectives[0].Suit != SuitPink {
				t.Fatalf("%dp seed %d: holder %d objective %+v, want owner %d", n, seed, holder, a.Objectives[0], want)
			}
			// the objective is shown to everyone
			v, _ := g.ViewFor(alice)
			if v.Attempt.Objectives[0].Objective.Owner != want {
				t.Fatal("view")
			}
		}
	}
	// fixed before a distress pass: passing the pink 9 does not change the owner
	g := withDistress(startMission(t, 4, 46))
	a := g.Attempt
	holder := -1
	for i, m := range a.Members {
		if slices.Contains(m.Hand, c("p9")) {
			holder = i
		}
	}
	want := (holder + 1) % 4
	if a.Objectives[0].Owner != want {
		t.Fatalf("owner %d, want %d", a.Objectives[0].Owner, want)
	}
	must(t, g.SendDistress(bbCmd(g), DirLeft))
	for s := range 4 {
		card := bbNonTrump(g, s)
		if s == holder {
			card = c("p9")
		}
		must(t, g.ChoosePassCard(bbID(g, s), card))
	}
	if a.Objectives[0].Owner != want {
		t.Fatalf("the owner changed with the pass: %d, want %d", a.Objectives[0].Owner, want)
	}
}

// rules §14 mission 34: the commander owns the first-and-last objective.
func TestRulesMission34Owner(t *testing.T) {
	for _, n := range []int{2, 3, 4, 5} {
		g := startMission(t, n, 34)
		a := g.Attempt
		if len(a.Objectives) != 2 {
			t.Fatalf("objectives %d", len(a.Objectives))
		}
		var bal, fl *Objective
		for i := range a.Objectives {
			switch a.Objectives[i].Kind {
			case KindCrewBalance:
				bal = &a.Objectives[i]
			case KindWinTricks:
				fl = &a.Objectives[i]
			}
		}
		if bal == nil || fl == nil || bal.Owner != OwnerCrew || fl.Owner != a.Commander {
			t.Fatalf("%dp: objectives %+v", n, a.Objectives)
		}
		if !slices.Equal(fl.Tricks, []int{1, LastTrick}) {
			t.Fatalf("tricks %v", fl.Tricks)
		}
	}
}

// ---------------------------------------------------------------- §11 handover

func TestRulesHandover(t *testing.T) {
	g := bbReady(t, 5, 31) // golden mission, 6 tasks, tokens 1 2 3
	a := g.Attempt
	cmd := a.Commander
	if g.CanHandover(cmd) != nil {
		t.Fatalf("handover must be possible: %v", g.CanHandover(cmd))
	}
	for s := range 5 {
		if s != cmd && g.CanHandover(s) == nil {
			t.Fatalf("seat %d: only the commander carries it out", s)
		}
	}
	v, _ := g.ViewFor(bbCmd(g))
	if !v.Attempt.Actions.Handover {
		t.Fatal("the commander is offered the handover")
	}
	owner := a.Tasks[0].Owner() // task 0 carries token 1
	to := (owner + 2) % 5
	other := (cmd + 1) % 5
	mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.Handover(bbID(g, other), 0, to) })
	mustFail(t, g, ErrInvalidTask, func(g *Game) error { return g.Handover(bbCmd(g), 40, to) })
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.Handover(bbCmd(g), 0, 9) })
	bbFailAny(t, g, func(g *Game) error { return g.Handover(bbCmd(g), 0, owner) }, ErrInvalidSeat, ErrNotAllowed)
	must(t, g.Handover(bbCmd(g), 0, to))
	if a.Tasks[0].Owner() != to || a.Tasks[0].Token != Token("1") {
		t.Fatalf("task 0 %+v: owner %d token %q", a.Tasks[0], a.Tasks[0].Owner(), a.Tasks[0].Token)
	}
	if ev := eventsOf[TaskHandedOver](g); len(ev) != 1 || ev[0].TaskID != 0 || ev[0].From != owner || ev[0].To != to {
		t.Fatalf("events %v", ev)
	}
	// at most once
	if g.CanHandover(cmd) == nil {
		t.Fatal("a second handover")
	}
	bbFailAny(t, g, func(g *Game) error { return g.Handover(bbCmd(g), 1, to) }, ErrNotAllowed, ErrWrongStep)
	v, _ = g.ViewFor(bbCmd(g))
	if v.Attempt.Actions.Handover {
		t.Fatal("no more offers")
	}
}

func TestRulesHandoverRestrictions(t *testing.T) {
	t.Run("not a golden mission", func(t *testing.T) {
		g := bbReady(t, 5, 10)
		if g.CanHandover(g.Attempt.Commander) == nil {
			t.Fatal("mission 10 has no golden frame")
		}
		bbFailAny(t, g, func(g *Game) error { return g.Handover(bbCmd(g), 0, 1) }, ErrNotAllowed)
	})
	t.Run("not with fewer than five players", func(t *testing.T) {
		g := bbReady(t, 4, 31)
		if g.CanHandover(g.Attempt.Commander) == nil {
			t.Fatal("four players have no handover")
		}
		bbFailAny(t, g, func(g *Game) error { return g.Handover(bbCmd(g), 0, 1) }, ErrNotAllowed)
	})
	t.Run("not before the distribution", func(t *testing.T) {
		g := startMission(t, 5, 31)
		if g.CanHandover(g.Attempt.Commander) == nil {
			t.Fatal("handover during the draft")
		}
		bbFailAny(t, g, func(g *Game) error { return g.Handover(bbCmd(g), 0, 1) }, ErrNotAllowed, ErrWrongStep)
	})
	t.Run("not after the first card", func(t *testing.T) {
		g := bbReady(t, 5, 31)
		cmd := g.Attempt.Commander
		must(t, g.PlayCard(bbCmd(g), g.LegalCards(cmd)[0], noShuffle{}))
		if g.CanHandover(cmd) == nil {
			t.Fatal("handover in trick 1")
		}
		owner := g.Attempt.Tasks[0].Owner()
		bbFailAny(t, g, func(g *Game) error { return g.Handover(bbCmd(g), 0, (owner+1)%5) }, ErrNotAllowed, ErrWrongStep)
	})
	t.Run("golden missions with 5 players", func(t *testing.T) {
		for _, m := range []int{25, 27, 28, 30, 31, 32, 35, 36, 37, 38, 39, 40, 42, 43, 45, 47, 48, 49} {
			if !MissionDefOf(m).Golden {
				t.Errorf("mission %d must be golden", m)
			}
		}
		for _, m := range []int{1, 10, 18, 22, 24, 26, 29, 33, 41, 44, 46, 50} {
			if MissionDefOf(m).Golden {
				t.Errorf("mission %d must not be golden", m)
			}
		}
	})
}
