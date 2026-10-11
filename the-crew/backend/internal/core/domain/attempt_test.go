package domain

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestDealPerPlayerCount(t *testing.T) {
	for _, tc := range []struct {
		players int
		sizes   []int // per seat: hand size, or for the dummy all its cards
		tricks  int
	}{
		{2, []int{13, 13, 14}, 13},
		{3, []int{14, 13, 13}, 13}, // IntN(3) = 0 gives seat 0 the extra card
		{4, []int{10, 10, 10, 10}, 10},
		{5, []int{8, 8, 8, 8, 8}, 8},
	} {
		g := startMission(t, tc.players, 1)
		a := g.Attempt
		if len(a.Members) != len(tc.sizes) {
			t.Fatalf("%d players: %d members", tc.players, len(a.Members))
		}
		all := []Card{}
		for i, want := range tc.sizes {
			m := a.Members[i]
			if got := m.CardCount(); got != want {
				t.Errorf("%d players: seat %d has %d cards, want %d", tc.players, i, got, want)
			}
			all = append(all, m.Cards()...)
			if !m.Dummy && !slices.IsSortedFunc(m.Hand, compareCards) {
				t.Errorf("%d players: seat %d hand not sorted", tc.players, i)
			}
		}
		SortCards(all)
		if want := NewDeck(); !slices.Equal(all, func() []Card { SortCards(want); return want }()) {
			t.Errorf("%d players: the deal is not the 40 cards", tc.players)
		}
		if !a.Members[a.Commander].holds(CommanderCard) || a.Members[a.Commander].Dummy {
			t.Errorf("%d players: commander seat %d does not hold t4", tc.players, a.Commander)
		}
		if a.history().Total != tc.tricks {
			t.Errorf("%d players: %d tricks, want %d", tc.players, a.history().Total, tc.tricks)
		}
	}
}

func TestDealTwoPlayersDummy(t *testing.T) {
	for seed := range uint64(30) {
		g := newCrew(t, 2)
		must(t, g.StartMission(alice, rand.New(rand.NewPCG(seed, 1))))
		d := g.Attempt.Members[2]
		if !d.Dummy || d.PlayerID != "" || len(d.Stacks) != DummyStacks {
			t.Fatalf("seat 2 is not the dummy: %+v", d)
		}
		for _, s := range d.Stacks {
			if s.Up == nil || s.Down == nil {
				t.Fatalf("stack not dealt: %+v", s)
			}
		}
		if slices.Contains(d.Cards(), CommanderCard) {
			t.Fatal("the dummy must never hold t4")
		}
	}
}

func TestDealThreePlayersExtraCardSeat(t *testing.T) {
	for want := range 3 {
		g := newCrew(t, 3)
		must(t, g.StartMission(alice, &fixedIntN{vals: []int{want}}))
		for i, m := range g.Attempt.Members {
			size := 13
			if i == want {
				size = 14
			}
			if len(m.Hand) != size {
				t.Fatalf("extra card to seat %d: seat %d has %d", want, i, len(m.Hand))
			}
		}
	}
}

func TestMissionOneTwoPlayersWin(t *testing.T) {
	g := startMission(t, 2, 1)
	setHands(g,
		"t4 p1 p2 p3 p4 p5 p6 p7 p8 p9 b1 b2 b3",
		"b4 b5 b6 b7 b8 b9 g1 g2 g3 g4 g5 g6 g7")
	setDummy(g, "g8 g9 y1 y2 y3 y4 y5", "y6 y7 y8 y9 t1 t2 t3")
	setTasks(g, "p9:0")
	if g.Attempt.Commander != 0 || actorID(g) != alice {
		t.Fatalf("alice must lead: commander %d", g.Attempt.Commander)
	}
	// alice leads p9, bob has no pink, the dummy (played by alice) has no pink either.
	playAll(t, g, "p9 b4 g8")
	if g.Phase != PhaseDebrief || !g.Attempt.Result.Success || g.Attempt.Result.Reason != ReasonCompleted {
		t.Fatalf("phase %s result %+v", g.Phase, g.Attempt.Result)
	}
	if e := g.Logbook[1]; e.Attempts != 1 || !e.Completed {
		t.Fatalf("logbook %+v", e)
	}
	if totalCards(g) != 40 {
		t.Fatalf("cards %d", totalCards(g))
	}
}
