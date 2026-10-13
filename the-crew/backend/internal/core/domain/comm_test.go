package domain

import (
	"slices"
	"testing"
)

func TestStatementFor(t *testing.T) {
	for _, tc := range []struct {
		hand, card string
		want       Position
		ok         bool
	}{
		{"p1 p5 p9 b3", "p9", PosHighest, true},
		{"p1 p5 p9 b3", "p1", PosLowest, true},
		{"p1 p5 p9 b3", "p5", "", false}, // a middle card
		{"p1 p5 p9 b3", "b3", PosOnly, true},
		{"p1 p5 b3 t2", "t2", "", false}, // trump never
		{"p1 p5 b3", "g1", "", false},    // not in hand
		{"p4 p5", "p5", PosHighest, true},
	} {
		pos, ok := StatementFor(cs(tc.hand), c(tc.card))
		if pos != tc.want || ok != tc.ok {
			t.Errorf("StatementFor(%s, %s) = %q %v, want %q %v", tc.hand, tc.card, pos, ok, tc.want, tc.ok)
		}
	}
}

// fourAtPlay returns 4 players playing mission (tasks y9 for bob) with small fixed hands; alice
// holds t4 and leads.
func fourAtPlay(t *testing.T, mission int) *Game {
	t.Helper()
	g := startMission(t, 4, mission)
	setHands(g, "t4 p1 p5 b3", "p3 b2 g1 g2", "g3 g4 y1 b9", "y2 y3 t1 t2")
	setTasks(g, "y9:1")
	return g
}

func TestCommunicate(t *testing.T) {
	g := fourAtPlay(t, 1)
	mustFail(t, g, ErrInvalidPosition, func(g *Game) error { return g.Communicate(bob, c("b2"), PosHighest) }) // a singleton is "only"
	mustFail(t, g, ErrInvalidPosition, func(g *Game) error { return g.Communicate(alice, c("p1"), PosOnly) })
	mustFail(t, g, ErrCannotCommunicate, func(g *Game) error { return g.Communicate(alice, c("t4"), PosHighest) })
	mustFail(t, g, ErrCardNotInHand, func(g *Game) error { return g.Communicate(alice, c("g3"), PosOnly) })
	must(t, g.Communicate(bob, c("b2"), PosOnly))
	if ev := eventsOf[Communicated](g); len(ev) != 1 || ev[0] != (Communicated{Seat: 1, Card: c("b2"), Position: PosOnly}) {
		t.Fatalf("events %v", ev)
	}
	mustFail(t, g, ErrAlreadyCommunicated, func(g *Game) error { return g.Communicate(bob, c("g1"), PosLowest) })
	must(t, g.Communicate(alice, c("p5"), PosHighest))
	must(t, g.PlayCard(alice, c("p1"), noShuffle{}))
	// Only before a trick: once the lead is on the table the window is closed.
	mustFail(t, g, ErrCannotCommunicate, func(g *Game) error { return g.Communicate(carol, c("y1"), PosOnly) })
	playAll(t, g, "p3 b9 y2") // bob wins and leads
	// The display disappears once the card is played; the communication stays used.
	if a := g.Attempt; a.CommunicatedCard(1) == nil {
		t.Fatal("b2 is still in bob's hand")
	}
	must(t, g.Communicate(carol, c("y1"), PosOnly)) // before trick 2
	playAll(t, g, "b2")
	if a := g.Attempt; a.CommunicatedCard(1) != nil || a.Comms[1] == nil {
		t.Fatal("b2 was played: no display, communication used")
	}
	v, _ := g.ViewFor(dave)
	if v.Attempt.Members[1].Communication != nil || !v.Attempt.Members[1].Communicated ||
		v.Attempt.Members[0].Communication == nil || v.Attempt.Members[0].Communication.Card != c("p5") {
		t.Fatalf("members %+v", v.Attempt.Members)
	}
}

func TestCommunicateNotBeforeDistribution(t *testing.T) {
	g := startMission(t, 4, 10) // draft
	mustFail(t, g, ErrWrongStep, func(g *Game) error { return g.Communicate(alice, c("p9"), PosHighest) })
}

func TestCommunicateDeadZone(t *testing.T) {
	g := fourAtPlay(t, 6)
	g.Attempt.Members[3].Hand = cs("y2 y3 y4 t1")
	mustFail(t, g, ErrInvalidPosition, func(g *Game) error { return g.Communicate(dave, c("y3"), PosHidden) }) // a middle card never qualifies
	must(t, g.Communicate(dave, c("y2"), PosHighest))                                                          // the stated position is ignored
	if cm := g.Attempt.Comms[3]; cm.Position != PosHidden {
		t.Fatalf("dead zone position %q", cm.Position)
	}
	v, _ := g.ViewFor(alice)
	if opts := v.Attempt.Actions.Communicate; len(opts) == 0 || opts[0].Position != PosHidden {
		t.Fatalf("options %+v", opts)
	}
}

func TestCommunicateDisruption(t *testing.T) {
	g := fourAtPlay(t, 18) // not before trick 2
	mustFail(t, g, ErrCannotCommunicate, func(g *Game) error { return g.Communicate(bob, c("b2"), PosOnly) })
	v, _ := g.ViewFor(bob)
	if len(v.Attempt.Actions.Communicate) != 0 {
		t.Fatal("no options before trick 2")
	}
	playAll(t, g, "p1 p3 b9 y2")
	must(t, g.Communicate(bob, c("b2"), PosOnly)) // directly before trick 2
}

func TestCommunicateSilenced(t *testing.T) {
	g := fourAtPlay(t, 11)
	g.Attempt.Silenced = 2
	mustFail(t, g, ErrCannotCommunicate, func(g *Game) error { return g.Communicate(carol, c("y1"), PosOnly) })
	must(t, g.Communicate(bob, c("b2"), PosOnly))
}

func TestDummyNeverCommunicates(t *testing.T) {
	g := startMission(t, 2, 1)
	setTasks(g, "p1:2")
	if err := g.CanCommunicate(2); err == nil {
		t.Fatal("the dummy must not communicate")
	}
}

func TestDistressPass(t *testing.T) {
	g := withDistress(fourAtPlay(t, 1))
	mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.SendDistress(bob, DirLeft) })
	mustFail(t, g, ErrInvalidDirection, func(g *Game) error { return g.SendDistress(alice, "up") })
	must(t, g.SendDistress(alice, DirLeft))
	a := g.Attempt
	if a.Step != StepPass || !slices.Equal(g.Actors(), []int{0, 1, 2, 3}) || !g.Logbook[1].Distress {
		t.Fatalf("step %s actors %v logbook %+v", a.Step, g.Actors(), g.Logbook[1])
	}
	mustFail(t, g, ErrCannotPassTrump, func(g *Game) error { return g.ChoosePassCard(alice, c("t4")) })
	mustFail(t, g, ErrCardNotInHand, func(g *Game) error { return g.ChoosePassCard(alice, c("y1")) })
	mustFail(t, g, ErrWrongStep, func(g *Game) error { return g.PlayCard(alice, c("p1"), noShuffle{}) })
	mustFail(t, g, ErrWrongStep, func(g *Game) error { return g.Communicate(bob, c("b2"), PosOnly) })
	must(t, g.ChoosePassCard(alice, c("b3")))
	mustFail(t, g, ErrAlreadyAnswered, func(g *Game) error { return g.ChoosePassCard(alice, c("p1")) })
	// Other players' choices stay secret.
	v, _ := g.ViewFor(bob)
	if v.Attempt.Pass == nil || v.Attempt.Pass.Chosen != nil || !slices.Equal(v.Attempt.Pass.Waiting, []int{1, 2, 3}) {
		t.Fatalf("bob's pass view %+v", v.Attempt.Pass)
	}
	must(t, g.ChoosePassCard(bob, c("g2")))
	must(t, g.ChoosePassCard(carol, c("g4")))
	must(t, g.ChoosePassCard(dave, c("y3")))
	// Left = the next seat: alice → bob → carol → dave → alice.
	for i, want := range []string{"t4 p1 p5 y3", "p3 b2 g1 b3", "g3 y1 b9 g2", "y2 t1 t2 g4"} {
		if !slices.Equal(a.Members[i].Hand, cs(want)) {
			t.Errorf("seat %d hand %v, want %s", i, a.Members[i].Hand, want)
		}
	}
	if a.Step != StepPlay || a.Pass != nil {
		t.Fatalf("step %s", a.Step)
	}
	v, _ = g.ViewFor(bob)
	if ex := v.Attempt.Exchange; ex == nil || *ex.Gave != c("g2") || *ex.Got != c("b3") || ex.Kind != ExchangePass {
		t.Fatalf("bob's exchange %+v", v.Attempt.Exchange)
	}
	if ev := eventsOf[CardsPassed](g); len(ev) != 1 || ev[0].Direction != DirLeft {
		t.Fatalf("events %v", ev)
	}
	// Once per attempt.
	mustFail(t, g, ErrNotAllowed, func(g *Game) error { return g.SendDistress(alice, DirRight) })
	mustFail(t, g, ErrNotAllowed, func(g *Game) error { return g.SkipDistress(alice) })
}

func TestDistressRightAndTwoPlayers(t *testing.T) {
	g := withDistress(fourAtPlay(t, 1))
	must(t, g.SendDistress(alice, DirRight))
	for i, card := range []string{"b3", "g2", "g4", "y3"} {
		must(t, g.ChoosePassCard(crewIDs[i], c(card)))
	}
	// Right = the previous seat: alice → dave.
	if !slices.Contains(g.Attempt.Members[3].Hand, c("b3")) || !slices.Contains(g.Attempt.Members[0].Hand, c("g2")) {
		t.Fatalf("hands %+v", g.Attempt.Members)
	}

	g = withDistress(startMission(t, 2, 1))
	setTasks(g, "p1:0")
	must(t, g.SendDistress(bob, DirRight))
	if !slices.Equal(g.Actors(), []int{0, 1}) {
		t.Fatalf("the dummy does not pass: actors %v", g.Actors())
	}
	must(t, g.ChoosePassCard(alice, c("b6")))
	must(t, g.ChoosePassCard(bob, c("y1")))
	a := g.Attempt
	if !slices.Contains(a.Members[0].Hand, c("y1")) || !slices.Contains(a.Members[1].Hand, c("b6")) || a.Members[2].CardCount() != 14 {
		t.Fatal("the two humans swap")
	}
}

func TestDistressWindow(t *testing.T) {
	g := withDistress(fourAtPlay(t, 1))
	must(t, g.Communicate(bob, c("b2"), PosOnly))
	mustFail(t, g, ErrNotAllowed, func(g *Game) error { return g.SendDistress(alice, DirLeft) })

	g = withDistress(fourAtPlay(t, 1))
	must(t, g.PlayCard(alice, c("p1"), noShuffle{}))
	mustFail(t, g, ErrNotAllowed, func(g *Game) error { return g.SendDistress(alice, DirLeft) })

	g = withDistress(fourAtPlay(t, 1))
	must(t, g.SkipDistress(alice))
	if g.Logbook[1].Distress || g.Attempt.Step != StepPlay {
		t.Fatal("skipping does not use the signal")
	}
	mustFail(t, g, ErrNotAllowed, func(g *Game) error { return g.SendDistress(alice, DirLeft) })
	v, _ := g.ViewFor(alice)
	if v.Attempt.Actions.SendDistress {
		t.Fatal("no distress action after skipping")
	}
}

func TestHandover(t *testing.T) {
	g := startMission(t, 5, 25) // golden
	setTasks(g, "p1:0", "p2:1")
	mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.Handover(alice, 0, 2) })
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.Handover(erin, 0, 0) }) // already the owner
	mustFail(t, g, ErrInvalidSeat, func(g *Game) error { return g.Handover(erin, 0, 5) })
	mustFail(t, g, ErrInvalidTask, func(g *Game) error { return g.Handover(erin, 2, 1) })
	must(t, g.Handover(erin, 0, 3))
	if g.Attempt.Tasks[0].Owner() != 3 {
		t.Fatal("task 0 must belong to dave")
	}
	mustFail(t, g, ErrNotAllowed, func(g *Game) error { return g.Handover(erin, 1, 2) }) // once

	g = startMission(t, 5, 24) // not golden
	setTasks(g, "p1:0")
	mustFail(t, g, ErrNotAllowed, func(g *Game) error { return g.Handover(erin, 0, 2) })
	g = startMission(t, 4, 25) // golden, 4 players
	setTasks(g, "p1:0")
	mustFail(t, g, ErrNotAllowed, func(g *Game) error { return g.Handover(dave, 0, 2) })
	g = startMission(t, 5, 25)
	setTasks(g, "p1:0")
	must(t, g.PlayCard(erin, c("y6"), noShuffle{}))
	mustFail(t, g, ErrNotAllowed, func(g *Game) error { return g.Handover(erin, 0, 2) })
}

func TestMission12Draw(t *testing.T) {
	g := fourAtPlay(t, 12)
	must(t, g.Communicate(bob, c("b2"), PosOnly)) // a communicated card cannot be drawn
	playAll(t, g, "p1 p3 b9 y2")                  // bob wins; noShuffle draws the first card of each pool
	a := g.Attempt
	// Each human draws from the right neighbour (previous seat): alice ← dave (y3), bob ← alice
	// (p5), carol ← bob (g1, not the communicated b2), dave ← carol (g3).
	for i, want := range []string{"t4 b3 y3", "b2 g2 p5", "g4 y1 g1", "t1 t2 g3"} {
		if !slices.Equal(a.Members[i].Hand, cs(want)) {
			t.Errorf("seat %d hand %v, want %s", i, a.Members[i].Hand, want)
		}
	}
	if len(eventsOf[CardsDrawn](g)) != 1 || a.Exchange.Kind != ExchangeDraw {
		t.Fatal("want a draw")
	}
	v, _ := g.ViewFor(carol)
	if ex := v.Attempt.Exchange; *ex.Gave != c("g3") || *ex.Got != c("g1") {
		t.Fatalf("carol's exchange %+v", ex)
	}
	// Only after the first trick.
	playAll(t, g, "b2 g4 t1 b3")
	if len(eventsOf[CardsDrawn](g)) != 1 {
		t.Fatal("only one draw")
	}
}

func TestMission12DrawTwoPlayers(t *testing.T) {
	g := startMission(t, 2, 12)
	setHands(g, "p1 p2 b1", "t4 p3 g1")
	setDummy(g, "p4 g2 y1 y2 y3 y4 y5", "- - - - - - -")
	setTasks(g, "y9:0")
	playAll(t, g, "p3 p4 p1") // bob (commander) leads, plays p4 for the dummy, which wins
	a := g.Attempt
	if a.Members[2].CardCount() != 6 || len(a.Members[0].Hand)+len(a.Members[1].Hand) != 4 {
		t.Fatalf("the dummy takes no part: %+v", a.Members)
	}
	if !slices.Equal(a.Members[0].Hand, cs("b1 g1")) || !slices.Equal(a.Members[1].Hand, cs("p2 t4")) {
		t.Fatalf("the humans draw from each other: %v %v", a.Members[0].Hand, a.Members[1].Hand)
	}
}
