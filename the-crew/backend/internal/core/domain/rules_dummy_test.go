package domain

import (
	"math/rand/v2"
	"slices"
	"testing"
)

const (
	bbUp   = "p1 p2 b3 b4 g5 g6 y7"
	bbDown = "p9 b9 g9 y8 t1 t2 t3"
)

// bbDummyGame is a 2-player game of mission 1 (task y9 of seat 0, never played) with the given
// hands and dummy stacks. Whoever holds t4 is the commander.
func bbDummyGame(t *testing.T, aliceHand, bobHand, up, down string) *Game {
	t.Helper()
	g := startMission(t, 2, 1)
	setHands(g, aliceHand, bobHand)
	setDummy(g, up, down)
	setTasks(g, "y9:0")
	g.PullEvents()
	return g
}

// rules §12: the commander plays for JARVIS; the other human may not.
func TestRulesDummyPlayedByCommander(t *testing.T) {
	t.Run("alice commander: order alice, bob, dummy", func(t *testing.T) {
		g := bbDummyGame(t, "t4 p5 b5", "p6 b6", bbUp, bbDown)
		if g.Attempt.Commander != 0 {
			t.Fatalf("commander %d", g.Attempt.Commander)
		}
		must(t, g.PlayCard(alice, c("p5"), noShuffle{}))
		must(t, g.PlayCard(bob, c("p6"), noShuffle{}))
		if a := g.Actors(); len(a) != 1 || a[0] != 0 {
			t.Fatalf("on the dummy's turn the game waits for the commander: %v", a)
		}
		bbFailAny(t, g, func(g *Game) error { return g.PlayCard(bob, c("p1"), noShuffle{}) }, ErrNotYourTurn, ErrNotCommander)
		must(t, g.PlayCard(alice, c("p1"), noShuffle{}))
		if len(g.Attempt.Tricks) != 1 {
			t.Fatal("trick not completed")
		}
		if g.Attempt.Tricks[0].Plays[2].Seat != 2 || g.Attempt.Tricks[0].Plays[2].Card != c("p1") {
			t.Fatalf("plays %+v", g.Attempt.Tricks[0].Plays)
		}
	})
	t.Run("bob commander: order bob, dummy, alice", func(t *testing.T) {
		g := bbDummyGame(t, "p5 b5", "t4 p6 b6", bbUp, bbDown)
		if g.Attempt.Commander != 1 {
			t.Fatalf("commander %d", g.Attempt.Commander)
		}
		must(t, g.PlayCard(bob, c("p6"), noShuffle{}))
		if a := g.Actors(); len(a) != 1 || a[0] != 1 {
			t.Fatalf("on the dummy's turn the game waits for the commander: %v", a)
		}
		bbFailAny(t, g, func(g *Game) error { return g.PlayCard(alice, c("p1"), noShuffle{}) }, ErrNotYourTurn, ErrNotCommander)
		must(t, g.PlayCard(bob, c("p2"), noShuffle{})) // the commander plays JARVIS's p2
		if a := g.Actors(); len(a) != 1 || a[0] != 0 {
			t.Fatalf("now alice: %v", a)
		}
		bbFailAny(t, g, func(g *Game) error { return g.PlayCard(bob, c("p5"), noShuffle{}) }, ErrNotYourTurn)
		must(t, g.PlayCard(alice, c("p5"), noShuffle{}))
		if g.Attempt.Tricks[0].Winner != 1 {
			t.Fatalf("winner %d", g.Attempt.Tricks[0].Winner)
		}
	})
	t.Run("the commander cannot play JARVIS's card on a human's turn", func(t *testing.T) {
		g := bbDummyGame(t, "t4 p5 b5", "p6 b6", bbUp, bbDown)
		must(t, g.PlayCard(alice, c("p5"), noShuffle{}))
		// bob's turn: the commander alice cannot play bob's or JARVIS's cards
		bbFailAny(t, g, func(g *Game) error { return g.PlayCard(alice, c("p1"), noShuffle{}) }, ErrNotYourTurn)
		bbFailAny(t, g, func(g *Game) error { return g.PlayCard(alice, c("p6"), noShuffle{}) }, ErrNotYourTurn, ErrCardNotInHand)
	})
}

// rules §12: JARVIS plays only face-up cards.
func TestRulesDummyOnlyFaceUp(t *testing.T) {
	g := bbDummyGame(t, "t4 p5 b5", "p6 b6", bbUp, bbDown)
	must(t, g.PlayCard(alice, c("p5"), noShuffle{}))
	must(t, g.PlayCard(bob, c("p6"), noShuffle{}))
	// p9 lies face down under p1; b9 under b3; t1 under g6
	for _, hidden := range []string{"p9", "b9", "g9", "y8", "t1", "t2", "t3"} {
		hidden := hidden
		bbFailAny(t, g, func(g *Game) error { return g.PlayCard(alice, c(hidden), noShuffle{}) }, ErrCardNotInHand, ErrNotAllowed, ErrMustFollowSuit)
	}
	// a card nobody holds
	bbFailAny(t, g, func(g *Game) error { return g.PlayCard(alice, c("g1"), noShuffle{}) }, ErrCardNotInHand)
	if got := g.LegalCards(2); !bbCardsEqual(got, cs("p1 p2")) {
		t.Fatalf("legal %v", got)
	}
	// the face-up p2 is fine too
	must(t, g.PlayCard(alice, c("p2"), noShuffle{}))
}

// rules §12 app decision: follow suit applies to the face-up cards only.
func TestRulesDummyFollowSuitFaceUpOnly(t *testing.T) {
	const greenHidden = "p1 p2 b3 b4 y5 y6 y7" // green only face down
	for _, tc := range []struct {
		name      string
		up, down  string
		lead, bob string // alice leads, bob answers
		legal     string // the dummy's legal cards
	}{
		{"colour led, face-up matches (hidden p9 irrelevant)", bbUp, bbDown, "p5", "p6", "p1 p2"},
		{"colour led, two face-up matches", bbUp, bbDown, "b5", "b6", "b3 b4"},
		{"a single face-up match", bbUp, bbDown, "y3", "p6", "y7"},
		{"no face-up match although a hidden card matches: free", greenHidden, "p9 b9 g9 g8 t1 t2 t3", "g2", "g1", greenHidden},
		{"trump led, only hidden trumps: free", bbUp, bbDown, "t4", "p6", bbUp},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := bbDummyGame(t, "t4 p5 b5 y3 g2", "p6 b6 g1", tc.up, tc.down)
			must(t, g.PlayCard(alice, c(tc.lead), noShuffle{}))
			must(t, g.PlayCard(bob, c(tc.bob), noShuffle{}))
			if got := g.LegalCards(2); !bbCardsEqual(got, cs(tc.legal)) {
				t.Fatalf("LegalCards(dummy) = %v, want %v", got, cs(tc.legal))
			}
			// every non-legal face-up card is refused, every legal card accepted
			for _, f := range splitFields(tc.up) {
				if slices.Contains(cs(tc.legal), c(f)) {
					continue
				}
				f := f
				mustFail(t, g, ErrMustFollowSuit, func(g *Game) error { return g.PlayCard(alice, c(f), noShuffle{}) })
			}
			must(t, g.PlayCard(alice, cs(tc.legal)[0], noShuffle{}))
		})
	}
}

func TestRulesDummyFollowSuitGreenLead(t *testing.T) {
	// green is only hidden under the dummy's cards; the lead is green: JARVIS may play anything face-up
	up, down := "p1 p2 b3 b4 y5 y6 y7", "p9 b9 g9 g8 t1 t2 t3"
	g := bbDummyGame(t, "t4 g2 b5", "g1 b6", up, down)
	must(t, g.PlayCard(alice, c("g2"), noShuffle{}))
	must(t, g.PlayCard(bob, c("g1"), noShuffle{}))
	if got := g.LegalCards(2); !bbCardsEqual(got, cs(up)) {
		t.Fatalf("LegalCards(dummy) = %v", got)
	}
	must(t, g.PlayCard(alice, c("y7"), noShuffle{}))
	if g.Attempt.Tricks[0].Winner != 0 {
		t.Fatalf("winner %d", g.Attempt.Tricks[0].Winner)
	}
}

// rules §12: the card below turns up after the trick ends, not during it.
func TestRulesDummyHiddenCardTurnsUpAfterTrick(t *testing.T) {
	g := bbDummyGame(t, "t4 p5 b5", "p6 b6", bbUp, bbDown)
	must(t, g.PlayCard(alice, c("p5"), noShuffle{}))
	must(t, g.PlayCard(bob, c("p6"), noShuffle{}))
	must(t, g.PlayCard(alice, c("p1"), noShuffle{})) // JARVIS plays p1; p9 lies under it
	// the trick is complete: the hidden p9 is face up now
	s := g.Attempt.Members[2].Stacks[0]
	if s.Up == nil || *s.Up != c("p9") || s.Down != nil {
		t.Fatalf("stack 0 after the trick: %+v", s)
	}
	v, err := g.ViewFor(alice)
	must(t, err)
	st := v.Attempt.Members[2].Stacks[0]
	if st.Up == nil || *st.Up != c("p9") || st.Hidden {
		t.Fatalf("view of stack 0: %+v", st)
	}
	// bob won (p6); the new face-up card is playable the next time
	if got := g.LegalCards(1); len(got) == 0 {
		t.Fatal("bob leads")
	}
}

func TestRulesDummyHiddenCardStaysHiddenDuringTrick(t *testing.T) {
	// the dummy plays second: after its card the trick is still open
	g := bbDummyGame(t, "p5 b5", "t4 p6 b6", bbUp, bbDown)
	must(t, g.PlayCard(bob, c("p6"), noShuffle{}))
	must(t, g.PlayCard(bob, c("p1"), noShuffle{})) // JARVIS
	s := g.Attempt.Members[2].Stacks[0]
	if s.Up != nil || s.Down == nil || *s.Down != c("p9") {
		t.Fatalf("during the trick the card below must stay face down: %+v", s)
	}
	for _, id := range []PlayerID{alice, bob} {
		v, err := g.ViewFor(id)
		must(t, err)
		st := v.Attempt.Members[2].Stacks[0]
		if st.Up != nil || !st.Hidden {
			t.Fatalf("%s sees %+v during the trick", id, st)
		}
	}
	if got := g.LegalCards(2); len(got) != 0 {
		t.Fatalf("dummy not to play: %v", got)
	}
	must(t, g.PlayCard(alice, c("p5"), noShuffle{}))
	if s := g.Attempt.Members[2].Stacks[0]; s.Up == nil || *s.Up != c("p9") || s.Down != nil {
		t.Fatalf("after the trick: %+v", s)
	}
}

func TestRulesDummyWinsAndLeads(t *testing.T) {
	// dummy up "p9 ...": it wins trick 1 and leads trick 2, played by the commander
	g := bbDummyGame(t, "t4 p1 b1", "p2 b6", "p9 b3 b4 g5 g6 y6 y7", "p8 b9 g9 y8 t1 t2 t3")
	must(t, g.PlayCard(alice, c("p1"), noShuffle{}))
	must(t, g.PlayCard(bob, c("p2"), noShuffle{}))
	must(t, g.PlayCard(alice, c("p9"), noShuffle{}))
	if g.Attempt.Tricks[0].Winner != 2 {
		t.Fatalf("winner %d", g.Attempt.Tricks[0].Winner)
	}
	// JARVIS leads: its face-up cards (the revealed p8 included) are all legal
	if got := g.LegalCards(2); !bbCardsEqual(got, cs("p8 b3 b4 g5 g6 y6 y7")) {
		t.Fatalf("the dummy leads with %v", got)
	}
	if len(g.LegalCards(0)) != 0 || len(g.LegalCards(1)) != 0 {
		t.Fatal("humans must wait")
	}
	if a := g.Actors(); len(a) != 1 || a[0] != 0 {
		t.Fatalf("actors %v", a)
	}
	bbFailAny(t, g, func(g *Game) error { return g.PlayCard(bob, c("b3"), noShuffle{}) }, ErrNotYourTurn, ErrNotCommander)
	must(t, g.PlayCard(alice, c("b3"), noShuffle{}))
	// blue led by JARVIS: bob (b6) must follow
	if got := g.LegalCards(0); len(got) != 1 && !bbHas(got, "b1") {
		t.Fatalf("alice must follow with blue: %v", got)
	}
}

// rules §12: the deal for two players.
func TestRulesDealTwoPlayersShape(t *testing.T) {
	for seed := range uint64(40) {
		g := newCrew(t, 2)
		must(t, g.StartMission(alice, rand.New(rand.NewPCG(seed, 7))))
		a := g.Attempt
		if len(a.Members) != 3 || !a.Members[2].Dummy || a.Members[0].Dummy || a.Members[1].Dummy {
			t.Fatalf("members %+v", a.Members)
		}
		if len(a.Members[0].Hand) != 13 || len(a.Members[1].Hand) != 13 {
			t.Fatalf("hands %d %d", len(a.Members[0].Hand), len(a.Members[1].Hand))
		}
		d := a.Members[2]
		if len(d.Stacks) != 7 || d.CardCount() != 14 || len(d.Hand) != 0 {
			t.Fatalf("dummy %+v", d)
		}
		for _, s := range d.Stacks {
			if s.Up == nil || s.Down == nil {
				t.Fatalf("stack %+v", s)
			}
		}
		if a.Commander == 2 || !a.Members[a.Commander].holds(CommanderCard) {
			t.Fatalf("commander %d must be the human holding t4", a.Commander)
		}
		if bbHas(d.Cards(), "t4") {
			t.Fatal("JARVIS holds t4")
		}
		all := slices.Concat(a.Members[0].Hand, a.Members[1].Hand, d.Cards())
		deck := NewDeck()
		SortCards(all)
		SortCards(deck)
		if !slices.Equal(all, deck) {
			t.Fatalf("not the 40 cards: %v", all)
		}
		if TotalTricks(2) != 13 {
			t.Fatalf("tricks %d", TotalTricks(2))
		}
	}
}

// rules §4 / §12: both humans have 13 cards and JARVIS 14 so one JARVIS card stays unplayed after
// 13 tricks; the unplayed card matters for tasks.
func TestRulesDummyUnplayedLastCard(t *testing.T) {
	t.Run("a task card left under JARVIS's last card fails the attempt", func(t *testing.T) {
		g := startMission(t, 2, 1)
		setHands(g, "t4", "b6")
		setDummy(g, "p1 - - - - - -", "p9 - - - - - -")
		setTasks(g, "p9:0")
		bbHistory(g, bbFillers(3, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0)...)
		g.Attempt.Commander = 0
		must(t, g.PlayCard(alice, c("t4"), noShuffle{}))
		must(t, g.PlayCard(bob, c("b6"), noShuffle{}))
		must(t, g.PlayCard(alice, c("p1"), noShuffle{}))
		// the 13th trick is complete; p9 turned up but cannot be played any more
		bbExpectEnd(t, g, false, ReasonTaskFailed, 13)
		if g.Attempt.Members[2].CardCount() != 1 {
			t.Fatalf("JARVIS has %d cards left", g.Attempt.Members[2].CardCount())
		}
		if g.Attempt.Result.Task != 0 {
			t.Fatalf("task %d", g.Attempt.Result.Task)
		}
	})
	t.Run("the unplayed card does not matter without a task on it", func(t *testing.T) {
		g := startMission(t, 2, 1)
		setHands(g, "t4", "b6")
		setDummy(g, "p1 - - - - - -", "p8 - - - - - -")
		setTasks(g, "t4:0")
		bbHistory(g, bbFillers(3, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0)...)
		g.Attempt.Commander = 0
		must(t, g.PlayCard(alice, c("t4"), noShuffle{}))
		must(t, g.PlayCard(bob, c("b6"), noShuffle{}))
		must(t, g.PlayCard(alice, c("p1"), noShuffle{}))
		bbExpectEnd(t, g, true, ReasonCompleted, 13)
	})
}
