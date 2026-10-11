package domain

import (
	"slices"
	"strings"
	"testing"
)

func TestLegalCards(t *testing.T) {
	for _, tc := range []struct {
		hand, lead, want string
	}{
		{"p1 p5 b3 t2", "", "p1 p5 b3 t2"},  // leading: anything
		{"p1 p5 b3 t2", "p9", "p1 p5"},      // follow the colour
		{"p1 p5 b3 t2", "g4", "p1 p5 b3 t2"}, // cannot follow: anything, trump too
		{"p1 p5 b3 t2", "t4", "t2"},          // trump is a suit too
		{"p1 p5 b3", "t1", "p1 p5 b3"},       // no trump to follow
		{"", "p1", ""},
	} {
		var lead *Card
		if tc.lead != "" {
			lead = cp(tc.lead)
		}
		got := legalCards(cs(tc.hand), lead)
		if !slices.Equal(got, cs(tc.want)) {
			t.Errorf("legalCards(%s, %s) = %v, want %s", tc.hand, tc.lead, got, tc.want)
		}
		if got == nil {
			t.Errorf("legalCards(%s, %s) is nil", tc.hand, tc.lead)
		}
	}
}

func TestTrickWinner(t *testing.T) {
	for _, tc := range []struct {
		cards string // in play order
		want  int
	}{
		{"p5 p9 p1", 1},       // highest of the led colour
		{"p5 b9 g9", 0},       // other colours never win
		{"p5 t1 p9", 1},       // any trump beats colours
		{"p5 t1 t3 t2", 2},    // the highest trump
		{"t2 t1 p9", 0},       // trump led
		{"y1 y2", 1},          //
		{"b9 p9 g9 y9 b8", 0}, // five players
	} {
		var cards []Card
		for _, f := range strings.Fields(tc.cards) {
			cards = append(cards, c(f))
		}
		if got := trickWinner(cards); got != tc.want {
			t.Errorf("trickWinner(%s) = %d, want %d", tc.cards, got, tc.want)
		}
	}
}

func TestPlayCardRules(t *testing.T) {
	g := startMission(t, 4, 1)
	setHands(g, "t4 p1 p2 b1", "p3 b2 b3 g1", "g2 g3 g4 y1", "y2 y3 t1 t2")
	setTasks(g, "y9:1")
	// alice (t4) leads.
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.PlayCard(bob, c("p3"), noShuffle{}) })
	mustFail(t, g, ErrCardNotInHand, func(g *Game) error { return g.PlayCard(alice, c("p3"), noShuffle{}) })
	must(t, g.PlayCard(alice, c("p1"), noShuffle{}))
	mustFail(t, g, ErrMustFollowSuit, func(g *Game) error { return g.PlayCard(bob, c("b2"), noShuffle{}) })
	if got := g.LegalCards(1); !slices.Equal(got, cs("p3")) {
		t.Fatalf("bob's legal cards %v", got)
	}
	if got := g.LegalCards(0); len(got) != 0 {
		t.Fatalf("not alice's turn, legal %v", got)
	}
	playAll(t, g, "p3 g2 t1")
	// dave won with a trump and leads trick 2.
	a := g.Attempt
	if len(a.Tricks) != 1 || a.Tricks[0].Winner != 3 || a.Trick.Leader != 3 || actorID(g) != dave {
		t.Fatalf("tricks %+v, leader %d", a.Tricks, a.Trick.Leader)
	}
	if ev := eventsOf[TrickWon](g); len(ev) != 1 || ev[0] != (TrickWon{Number: 1, Seat: 3}) {
		t.Fatalf("events %v", ev)
	}
	if len(eventsOf[CardPlayed](g)) != 4 {
		t.Fatal("want 4 CardPlayed events")
	}
	if g.Revision == 0 || totalCards(g) != 16 {
		t.Fatalf("cards %d", totalCards(g))
	}
}

func TestDummyPlay(t *testing.T) {
	g := startMission(t, 2, 1)
	setHands(g,
		"t4 p1 p2 p3 p4 p5 p6 p7 p8 b1 b2 b3 b4",
		"b5 b6 b7 b8 b9 g1 g2 g3 g4 g5 g6 g7 g8")
	// The dummy: face up p9 g9 y1 y2 y3 y4 y5, hidden y6 y7 y8 y9 t1 t2 t3.
	setDummy(g, "p9 g9 y1 y2 y3 y4 y5", "y6 y7 y8 y9 t1 t2 t3")
	setTasks(g, "y9:1")

	must(t, g.PlayCard(alice, c("b1"), noShuffle{}))
	must(t, g.PlayCard(bob, c("b5"), noShuffle{}))
	// The dummy's turn: the commander (alice) acts; it has no blue face up, so any face-up card.
	if got := g.Actors(); !slices.Equal(got, []int{0}) {
		t.Fatalf("actors %v", got)
	}
	if got := g.LegalCards(2); !slices.Equal(got, cs("p9 g9 y1 y2 y3 y4 y5")) {
		t.Fatalf("dummy legal %v", got)
	}
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.PlayCard(bob, c("y1"), noShuffle{}) })
	mustFail(t, g, ErrCardNotInHand, func(g *Game) error { return g.PlayCard(alice, c("y6"), noShuffle{}) }) // hidden
	must(t, g.PlayCard(alice, c("y1"), noShuffle{}))
	d := g.Attempt.Members[2]
	if *d.Stacks[2].Up != c("y8") || d.Stacks[2].Down != nil {
		t.Fatalf("stack 2 must turn up y8 after the trick: %+v", d.Stacks[2])
	}
	// bob won (b5) and leads; lead yellow: the dummy follows among face-up yellows only.
	must(t, g.PlayCard(bob, c("g1"), noShuffle{}))
	if got := g.LegalCards(2); !slices.Equal(got, cs("g9")) {
		t.Fatalf("dummy must follow green with g9: %v", got)
	}
	must(t, g.PlayCard(alice, c("g9"), noShuffle{}))
	// mid-trick: the stack of g9 must not turn up yet.
	if d := g.Attempt.Members[2]; d.Stacks[1].Up != nil || *d.Stacks[1].Down != c("y7") {
		t.Fatalf("stack 1 turned up during the trick: %+v", d.Stacks[1])
	}
	must(t, g.PlayCard(alice, c("p1"), noShuffle{}))
	if d := g.Attempt.Members[2]; *d.Stacks[1].Up != c("y7") {
		t.Fatalf("stack 1 after the trick: %+v", d.Stacks[1])
	}
}

func TestDummyFollowsAmongFaceUpOnly(t *testing.T) {
	m := Member{Dummy: true, Stacks: []DummyStack{
		{Up: cp("p3"), Down: cp("b9")},
		{Up: cp("g3"), Down: nil},
		{Up: nil, Down: cp("b1")},
	}}
	if got := m.Playable(cp("b5")); !slices.Equal(got, cs("p3 g3")) {
		t.Fatalf("a hidden blue must not force the dummy: %v", got)
	}
	if got := m.Playable(cp("p5")); !slices.Equal(got, cs("p3")) {
		t.Fatalf("follow pink: %v", got)
	}
	if got := m.Cards(); !slices.Equal(got, cs("p3 b9 g3 b1")) {
		t.Fatalf("cards %v", got)
	}
}
