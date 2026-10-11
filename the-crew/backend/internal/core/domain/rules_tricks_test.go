package domain

import (
	"slices"
	"testing"
)

// bbTrickGame is a 4-player game of mission 1 whose only task (y9, owner seat 0) is never played,
// so the attempt keeps running. hands are in seat order; leader leads trick 1.
func bbTrickGame(t *testing.T, leader int, hands ...string) *Game {
	t.Helper()
	g := startMission(t, 4, 1)
	for i, h := range hands {
		g.Attempt.Members[i].Hand = cs(h)
	}
	setTasks(g, "y9:0")
	g.Attempt.Commander = leader
	g.Attempt.Trick = Trick{Leader: leader}
	g.PullEvents()
	return g
}

// rules §5: who wins a trick; the winner leads the next one.
func TestRulesTrickWinner(t *testing.T) {
	for _, tc := range []struct {
		name   string
		leader int
		hands  [4]string
		cards  string // in turn order from the leader
		winner int
	}{
		{"highest card of the led colour", 0, [4]string{"p3 b1", "p7 b1", "p5 b1", "p6 b1"}, "p3 p7 p5 p6", 1},
		{"the leader can win", 0, [4]string{"p9 b1", "p7 b1", "p5 b1", "p6 b1"}, "p9 p7 p5 p6", 0},
		{"the last player can win", 0, [4]string{"p1 b1", "p2 b1", "p3 b1", "p4 b1"}, "p1 p2 p3 p4", 3},
		{"off colour never wins", 0, [4]string{"p2 p3", "b9 b1", "g9 g1", "y7 y1"}, "p2 b9 g9 y7", 0},
		{"off colour never wins (high card is the last)", 0, [4]string{"p2 p3", "b9 b1", "g9 g1", "y8 y1"}, "p2 b9 g9 y8", 0},
		{"trump beats the led colour", 0, [4]string{"p9 b1", "t1 b1", "p8 b1", "p7 b1"}, "p9 t1 p8 p7", 1},
		{"highest trump wins", 0, [4]string{"p5 b1", "t2 b1", "t4 b1", "t3 b1"}, "p5 t2 t4 t3", 2},
		{"highest trump wins (earlier)", 0, [4]string{"p5 b1", "t3 b1", "t1 b1", "t2 b1"}, "p5 t3 t1 t2", 1},
		{"trump led: highest trump wins", 0, [4]string{"t2 b1", "t3 b1", "t1 b1", "t4 b1"}, "t2 t3 t1 t4", 3},
		{"trump led: nobody else has trump", 0, [4]string{"t1 b1", "b9 g1", "g9 g1", "y8 y1"}, "t1 b9 g9 y8", 0},
		{"lead from seat 2 wraps around", 2, [4]string{"p6 b1", "p2 b1", "p5 b1", "p9 b1"}, "p5 p9 p6 p2", 3},
		{"lead from seat 3", 3, [4]string{"p6 b1", "p8 b1", "p5 b1", "p3 b1"}, "p3 p6 p8 p5", 1},
		{"a low trump still beats a high colour", 1, [4]string{"p6 b1", "p1 b1", "p9 b1", "t1 b1"}, "p1 p9 t1 p6", 3},
		{"cannot follow: any colour may be played", 0, [4]string{"p4 p5", "b9 b1", "g7 g1", "y3 y1"}, "p4 b9 g7 y3", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := bbTrickGame(t, tc.leader, tc.hands[:]...)
			for i, f := range splitFields(tc.cards) {
				seat := (tc.leader + i) % 4
				must(t, g.PlayCard(g.Players[seat].ID, c(f), noShuffle{}))
			}
			a := g.Attempt
			if len(a.Tricks) != 1 || a.Tricks[0].Winner != tc.winner {
				t.Fatalf("tricks %+v, want winner %d", a.Tricks, tc.winner)
			}
			if a.Tricks[0].Leader != tc.leader || len(a.Tricks[0].Plays) != 4 {
				t.Fatalf("trick record %+v", a.Tricks[0])
			}
			if w := eventsOf[TrickWon](g); len(w) != 1 || w[0].Seat != tc.winner || w[0].Number != 1 {
				t.Fatalf("TrickWon events %v", w)
			}
			// the winner leads the next trick: only they may play
			for seat := range 4 {
				legal := g.LegalCards(seat)
				if (seat == tc.winner) != (len(legal) > 0) {
					t.Errorf("seat %d: LegalCards %v after the trick (winner %d)", seat, legal, tc.winner)
				}
			}
			if bbActor(g, tc.winner) != g.Players[tc.winner].ID {
				t.Errorf("wrong actor for the winner")
			}
			// the winner may open with any card of their hand
			if got := g.LegalCards(tc.winner); !bbCardsEqual(got, g.Attempt.Members[tc.winner].Hand) {
				t.Errorf("the leader may play any card: %v vs hand %v", got, g.Attempt.Members[tc.winner].Hand)
			}
		})
	}
}

func splitFields(s string) []string {
	var out []string
	cur := ""
	for _, r := range s + " " {
		if r == ' ' {
			if cur != "" {
				out = append(out, cur)
			}
			cur = ""
			continue
		}
		cur += string(r)
	}
	return out
}

// rules §5: follow suit, trump is a suit, no obligation to win.
func TestRulesFollowSuitLegalCards(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lead  string // seat 0's card
		hand  string // seat 1's hand
		legal string
	}{
		{"must follow the led colour", "p5", "p2 p9 b1 t1", "p2 p9"},
		{"no card of the colour: anything", "p5", "b1 t1 g4", "b1 t1 g4"},
		{"trump is not a way out when holding the colour", "p5", "p2 t1", "p2"},
		{"no obligation to play higher", "p5", "p2 p9", "p2 p9"},
		{"no obligation to play lower", "p5", "p6 p9", "p6 p9"},
		{"trump led: must play trump", "t3", "t1 t2 p2 b4", "t1 t2"},
		{"trump led: even the commander card", "t1", "t4 p2", "t4"},
		{"trump led: no trump in hand", "t3", "b1 p2", "b1 p2"},
		{"colour led, only trump in hand", "b4", "t1 t2", "t1 t2"},
		{"yellow led with a trump", "y3", "y1 t4", "y1"},
		{"single card", "g3", "g1", "g1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := bbTrickGame(t, 0, tc.lead+" b9", tc.hand, "b2 b3", "b4 b5")
			// before the lead: only the leader may play, with any card
			if len(g.LegalCards(1)) != 0 || len(g.LegalCards(2)) != 0 || len(g.LegalCards(3)) != 0 {
				t.Fatal("only the leader may play")
			}
			if got := g.LegalCards(0); !bbCardsEqual(got, cs(tc.lead+" b9")) {
				t.Fatalf("the leader may play any card, got %v", got)
			}
			must(t, g.PlayCard(alice, c(tc.lead), noShuffle{}))
			if got := g.LegalCards(1); !bbCardsEqual(got, cs(tc.legal)) {
				t.Fatalf("LegalCards = %v, want %v", got, cs(tc.legal))
			}
			if len(g.LegalCards(0)) != 0 || len(g.LegalCards(2)) != 0 {
				t.Fatal("not their turn")
			}
			// every card of the hand: legal ones are accepted, the others rejected without a change
			for _, card := range cs(tc.hand) {
				if slices.Contains(cs(tc.legal), card) {
					continue
				}
				card := card
				mustFail(t, g, ErrMustFollowSuit, func(g *Game) error { return g.PlayCard(bob, card, noShuffle{}) })
			}
			must(t, g.PlayCard(bob, cs(tc.legal)[0], noShuffle{}))
		})
	}
}

func TestRulesPlayCardErrors(t *testing.T) {
	g := bbTrickGame(t, 0, "p5 b9", "p2 b1", "b2 p3", "b4 g5")
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.PlayCard(bob, c("p2"), noShuffle{}) })
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.PlayCard(dave, c("b4"), noShuffle{}) })
	mustFail(t, g, ErrCardNotInHand, func(g *Game) error { return g.PlayCard(alice, c("p6"), noShuffle{}) })
	mustFail(t, g, ErrCardNotInHand, func(g *Game) error { return g.PlayCard(alice, c("p2"), noShuffle{}) }) // bob's card
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.PlayCard("nobody", c("p5"), noShuffle{}) })
	must(t, g.PlayCard(alice, c("p5"), noShuffle{}))
	if bbHas(g.Attempt.Members[0].Hand, "p5") || len(g.Attempt.Trick.Plays) != 1 {
		t.Fatalf("the card must leave the hand and lie in the trick: %+v", g.Attempt.Trick)
	}
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.PlayCard(alice, c("b9"), noShuffle{}) })
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.PlayCard(carol, c("p3"), noShuffle{}) })
	mustFail(t, g, ErrMustFollowSuit, func(g *Game) error { return g.PlayCard(bob, c("b1"), noShuffle{}) })
	// a played card cannot be played again
	must(t, g.PlayCard(bob, c("p2"), noShuffle{}))
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.PlayCard(bob, c("p2"), noShuffle{}) })
	// a card the viewer never had, even though the colour matches
	mustFail(t, g, ErrCardNotInHand, func(g *Game) error { return g.PlayCard(carol, c("p4"), noShuffle{}) })
	// carol holds p3 and must follow suit
	mustFail(t, g, ErrMustFollowSuit, func(g *Game) error { return g.PlayCard(carol, c("b2"), noShuffle{}) })
}

func TestRulesPlayOutsideAttempt(t *testing.T) {
	g := newCrew(t, 3)
	bbFailAny(t, g, func(g *Game) error { return g.PlayCard(alice, c("p1"), noShuffle{}) }, ErrWrongPhase)
}

// rules §5: events of a card and a completed trick; the trick record keeps play order.
func TestRulesTrickRecordAndEvents(t *testing.T) {
	g := bbTrickGame(t, 2, "p1 b1", "p2 b1", "p9 b1", "p3 b1")
	for i, f := range []string{"p9", "p3", "p1", "p2"} {
		before := len(g.Attempt.Tricks)
		must(t, g.PlayCard(g.Players[(2+i)%4].ID, c(f), noShuffle{}))
		if i < 3 && len(g.Attempt.Tricks) != before {
			t.Fatalf("trick completed early after %d cards", i+1)
		}
	}
	if got := eventsOf[CardPlayed](g); len(got) != 4 || got[0].Seat != 2 || got[0].Card != c("p9") || got[3].Seat != 1 {
		t.Fatalf("CardPlayed %v", got)
	}
	rec := g.Attempt.Tricks[0]
	wantSeats := []int{2, 3, 0, 1}
	for i, p := range rec.Plays {
		if p.Seat != wantSeats[i] {
			t.Fatalf("plays %v, want seats %v", rec.Plays, wantSeats)
		}
	}
	if rec.Winner != 2 {
		t.Fatalf("winner %d, want 2", rec.Winner)
	}
	if len(g.Attempt.Trick.Plays) != 0 {
		t.Fatalf("the trick in progress must be empty: %+v", g.Attempt.Trick)
	}
	if totalCards(g) != 8 {
		t.Fatalf("cards %d", totalCards(g))
	}
}

// rules §4 / §5: nobody is forced to win, the game goes on until the tasks decide.
func TestRulesNoObligationToWin(t *testing.T) {
	g := bbTrickGame(t, 0, "p5 b9", "p9 p2", "b2 p3", "b4 g5")
	must(t, g.PlayCard(alice, c("p5"), noShuffle{}))
	must(t, g.PlayCard(bob, c("p2"), noShuffle{})) // bob holds p9 but plays low
	must(t, g.PlayCard(carol, c("p3"), noShuffle{}))
	must(t, g.PlayCard(dave, c("b4"), noShuffle{}))
	if g.Attempt.Tricks[0].Winner != 0 {
		t.Fatalf("winner %d", g.Attempt.Tricks[0].Winner)
	}
}
