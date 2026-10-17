package domain

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

// rules §4: hand sizes, 40 distinct cards, the t4 holder is the commander, number of tricks.
func TestRulesDealPerPlayerCount(t *testing.T) {
	for _, tc := range []struct {
		players int
		sizes   []int // humans (plus the dummy's card count last)
		tricks  int
	}{
		{2, []int{13, 13, 14}, 13},
		{3, []int{14, 13, 13}, 13},
		{4, []int{10, 10, 10, 10}, 10},
		{5, []int{8, 8, 8, 8, 8}, 8},
	} {
		t.Run(fmt.Sprintf("%d players", tc.players), func(t *testing.T) {
			if TotalTricks(tc.players) != tc.tricks {
				t.Fatalf("TotalTricks = %d, want %d", TotalTricks(tc.players), tc.tricks)
			}
			for seed := range uint64(25) {
				g := newCrew(t, tc.players)
				must(t, g.StartMission(alice, rand.New(rand.NewPCG(seed, 99))))
				a := g.Attempt
				var all []Card
				extra := 0
				for i, m := range a.Members {
					all = slices.Concat(all, m.Cards())
					want := tc.sizes[0]
					if tc.players == 3 {
						want = 13
					}
					if tc.players == 2 {
						want = 13
						if i == 2 {
							want = 14
						}
					}
					if tc.players == 3 && m.CardCount() == 14 {
						extra++
						continue
					}
					if m.CardCount() != want {
						t.Fatalf("seed %d seat %d has %d cards, want %d", seed, i, m.CardCount(), want)
					}
				}
				if tc.players == 3 && extra != 1 {
					t.Fatalf("seed %d: %d seats with 14 cards, want exactly 1", seed, extra)
				}
				SortCards(all)
				deck := NewDeck()
				SortCards(deck)
				if !slices.Equal(all, deck) {
					t.Fatalf("seed %d: the deal is not the 40 cards", seed)
				}
				holder := -1
				for i, m := range a.Members {
					if slices.Contains(m.Cards(), c("t4")) {
						holder = i
					}
				}
				if holder != a.Commander || a.Members[holder].Dummy {
					t.Fatalf("seed %d: t4 with seat %d, commander %d", seed, holder, a.Commander)
				}
				for _, m := range a.Members {
					if !m.Dummy && !slices.IsSortedFunc(m.Hand, compareCards) {
						t.Fatalf("hand not sorted")
					}
				}
				if got := eventsOf[AttemptStarted](g); len(got) != 1 || got[0].Commander != holder || got[0].Mission != 1 || got[0].Number != 1 {
					t.Fatalf("AttemptStarted %v", got)
				}
				if g.Phase != PhasePlaying || a.Number != 1 {
					t.Fatalf("phase %s number %d", g.Phase, a.Number)
				}
			}
		})
	}
}

// rules §4: with 3 players one random seat gets the 14th card.
func TestRulesDealThreePlayersExtraCard(t *testing.T) {
	for want := range 3 {
		g := newCrew(t, 3)
		must(t, g.StartMission(alice, &fixedIntN{vals: []int{want}}))
		for i, m := range g.Attempt.Members {
			size := 13
			if i == want {
				size = 14
			}
			if m.CardCount() != size {
				t.Errorf("extra card to %d: seat %d has %d", want, i, m.CardCount())
			}
		}
	}
}

// rules §4: every attempt starts with a fresh deal.
func TestRulesNewAttemptNewDeal(t *testing.T) {
	g := newCrew(t, 4)
	must(t, g.StartMission(alice, rand.New(rand.NewPCG(1, 1))))
	first := g.Attempt.Members[0].Hand
	must(t, g.GiveUp(alice))
	must(t, g.NextAttempt(alice, NextAttemptChoice{}, rand.New(rand.NewPCG(2, 2))))
	if g.Attempt.Number != 2 || slices.Equal(first, g.Attempt.Members[0].Hand) {
		t.Fatalf("attempt %d hands equal %v", g.Attempt.Number, slices.Equal(first, g.Attempt.Members[0].Hand))
	}
	if len(g.Attempt.Tricks) != 0 || g.Attempt.Result != nil {
		t.Fatal("the new attempt must be fresh")
	}
}

// rules §4 / §6: the draft starts with the commander, goes clockwise round after round and the
// commander then leads trick 1. The dummy's picks are made by the commander.
func TestRulesOpenDraft(t *testing.T) {
	for _, tc := range []struct {
		players, mission, tasks int
	}{
		{2, 1, 1}, {2, 2, 2}, {2, 4, 3}, {2, 10, 4}, {2, 18, 5},
		{3, 4, 3}, {3, 10, 4}, {3, 24 - 6, 5}, {3, 47, 10},
		{4, 1, 1}, {4, 2, 2}, {4, 4, 3}, {4, 10, 4}, {4, 18, 5}, {4, 47, 10},
		{5, 1, 1}, {5, 10, 4}, {5, 18, 5}, {5, 22, 5}, {5, 47, 10}, {5, 42, 9},
	} {
		t.Run(fmt.Sprintf("%dp mission %d", tc.players, tc.mission), func(t *testing.T) {
			g := startMission(t, tc.players, tc.mission)
			a := g.Attempt
			seats := len(a.Members)
			cmd := a.Commander
			if len(a.Tasks) != tc.tasks {
				t.Fatalf("%d tasks, want %d", len(a.Tasks), tc.tasks)
			}
			if a.Step != StepDraft {
				t.Fatalf("step %s", a.Step)
			}
			var taken []int
			for i := range a.Tasks {
				seat := (cmd + i) % seats
				if a.DraftTurn() != seat {
					t.Fatalf("pick %d: turn %d, want %d", i, a.DraftTurn(), seat)
				}
				acts := g.Actors()
				wantActor := seat
				if a.Members[seat].Dummy {
					wantActor = cmd // the commander picks for JARVIS
				}
				if len(acts) != 1 || acts[0] != wantActor {
					t.Fatalf("pick %d: actors %v, want [%d]", i, acts, wantActor)
				}
				// somebody else may not pick
				other := (wantActor + 1) % len(g.Players)
				if other == wantActor {
					other = (wantActor + 2) % len(g.Players)
				}
				if !a.Members[other].Dummy {
					bbFailAny(t, g, func(g *Game) error { return g.PickTask(g.Players[other].ID, i) }, ErrNotYourTurn, ErrNotCommander)
				}
				// an invalid and an already taken task are refused
				mustFail(t, g, ErrInvalidTask, func(g *Game) error { return g.PickTask(g.Players[wantActor].ID, len(a.Tasks)+3) })
				if len(taken) > 0 {
					mustFail(t, g, ErrInvalidTask, func(g *Game) error { return g.PickTask(g.Players[wantActor].ID, taken[len(taken)-1]) })
				}
				// any free task may be taken, not only the lowest: take the last free one
				pick := len(a.Tasks) - 1
				for ; a.Tasks[pick].Owner() != OwnerUnassigned; pick-- {
				}
				must(t, g.PickTask(g.Players[wantActor].ID, pick))
				taken = append(taken, pick)
				if a.Tasks[pick].Owner() != seat {
					t.Fatalf("task %d owned by %d, want %d", pick, a.Tasks[pick].Owner(), seat)
				}
			}
			ev := eventsOf[TaskPicked](g)
			if len(ev) != tc.tasks {
				t.Fatalf("TaskPicked %v", ev)
			}
			for i, e := range ev {
				if e.Seat != (cmd+i)%seats {
					t.Fatalf("TaskPicked %d: seat %d", i, e.Seat)
				}
			}
			if a.Step != StepPlay {
				t.Fatalf("after the draft the step is %s, want play", a.Step)
			}
			// the commander leads trick 1
			for s := range seats {
				got := g.LegalCards(s)
				if (s == cmd) != (len(got) > 0) {
					t.Fatalf("seat %d legal %v (commander %d)", s, got, cmd)
				}
			}
			if a := g.Actors(); len(a) != 1 || a[0] != cmd {
				t.Fatalf("actors %v", a)
			}
			// the draft is over
			bbFailAny(t, g, func(g *Game) error { return g.PickTask(g.Players[cmd].ID, 0) }, ErrWrongStep)
		})
	}
}

// rules §7: tokens stay with their task cards when the task is taken.
func TestRulesDraftKeepsTokens(t *testing.T) {
	g := startMission(t, 3, 3)
	a := g.Attempt
	toks := []Token{a.Tasks[0].Token, a.Tasks[1].Token}
	bbDraft(t, g)
	if a.Tasks[0].Token != toks[0] || a.Tasks[1].Token != toks[1] || toks[0] != Token("1") || toks[1] != Token("2") {
		t.Fatalf("tokens %v -> %v %v", toks, a.Tasks[0].Token, a.Tasks[1].Token)
	}
}

// rules §6: only a started attempt has a draft.
func TestRulesPickTaskErrors(t *testing.T) {
	g := newCrew(t, 3)
	bbFailAny(t, g, func(g *Game) error { return g.PickTask(alice, 0) }, ErrWrongPhase)
	must(t, g.StartMission(alice, noShuffle{}))
	bbFailAny(t, g, func(g *Game) error { return g.PickTask("nobody", 0) }, ErrUnknownPlayer)
}
