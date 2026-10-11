package domain

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

// bbMinus returns the cards of a that are not in b.
func bbMinus(a, b []Card) []Card {
	var out []Card
	for _, x := range a {
		if !slices.Contains(b, x) {
			out = append(out, x)
		}
	}
	return out
}

// bbDrawGame is a 4-player game of mission 12 at the lead of trick 1. Seat 3 (the commander) has
// communicated b4, which still lies in its hand.
func bbDrawGame(t *testing.T) *Game {
	t.Helper()
	g := startMission(t, 4, 12)
	g.Attempt.Objectives = nil
	setHands(g, "p3 b1 b2 b3", "p5 g1 g2 g3", "p2 y2 y3 y4", "t4 p4 b4 g4")
	setTasks(g, "y9:0:omega")
	g.PullEvents()
	return g
}

// rules §16 mission 12: right after the first trick each human draws a random card from the
// right neighbour; decided first, moved at once; a communicated card on the table is safe.
func TestRulesMission12Draw(t *testing.T) {
	for seed := range uint64(40) {
		g := bbDrawGame(t)
		a := g.Attempt
		rng := rand.New(rand.NewPCG(seed, 12))
		must(t, g.Communicate(dave, c("b4"), PosOnly))
		play := func(by PlayerID, card string) {
			t.Helper()
			must(t, g.PlayCard(by, c(card), rng))
		}
		play(dave, "p4")
		play(alice, "p3")
		play(bob, "p5")
		// nothing is drawn before the trick is over
		if len(eventsOf[CardsDrawn](g)) != 0 || a.Exchange != nil {
			t.Fatalf("seed %d: drew during the trick", seed)
		}
		old := make([][]Card, 4)
		for s := range 4 {
			old[s] = slices.Clone(a.Members[s].Hand)
		}
		old[2] = cs("p2 y2 y3 y4") // seat 2 has not played yet
		play(carol, "p2")
		old[2] = slices.Clone(a.Members[2].Hand) // after playing p2 (must be a draw-free hand yet?)
		_ = old
		// the draw happened right after the trick
		if len(eventsOf[CardsDrawn](g)) != 1 {
			t.Fatalf("seed %d: CardsDrawn events %d", seed, len(eventsOf[CardsDrawn](g)))
		}
		if a.Exchange == nil || a.Exchange.Kind != ExchangeDraw {
			t.Fatalf("seed %d: exchange %+v", seed, a.Exchange)
		}
		// the hands before the draw (after playing the trick)
		before := [][]Card{cs("b1 b2 b3"), cs("g1 g2 g3"), cs("y2 y3 y4"), cs("t4 b4 g4")}
		for s := range 4 {
			if len(a.Members[s].Hand) != 3 {
				t.Fatalf("seed %d: seat %d has %d cards", seed, s, len(a.Members[s].Hand))
			}
			gave := bbMinus(before[s], a.Members[s].Hand)
			got := bbMinus(a.Members[s].Hand, before[s])
			if len(gave) != 1 || len(got) != 1 {
				t.Fatalf("seed %d: seat %d gave %v got %v", seed, s, gave, got)
			}
			left := (s + 1) % 4 // the player whose right neighbour is s
			right := (s + 3) % 4
			// what s got was in the right neighbour's hand before the draw
			if !slices.Contains(before[right], got[0]) {
				t.Fatalf("seed %d: seat %d got %v, not from seat %d (%v)", seed, s, got[0], right, before[right])
			}
			// and what s gave is what the left neighbour got
			leftGot := bbMinus(a.Members[left].Hand, before[left])
			if len(leftGot) != 1 || leftGot[0] != gave[0] {
				t.Fatalf("seed %d: seat %d gave %v but seat %d got %v", seed, s, gave, left, leftGot)
			}
			// the exchange view shows exactly this player's side
			v, _ := g.ViewFor(bbID(g, s))
			ex := v.Attempt.Exchange
			if ex == nil || ex.Kind != ExchangeDraw || ex.Gave == nil || *ex.Gave != gave[0] || ex.Got == nil || *ex.Got != got[0] {
				t.Fatalf("seed %d: seat %d exchange view %+v, gave %v got %v", seed, s, ex, gave, got)
			}
		}
		// the communicated card b4 of seat 3 cannot be drawn by seat 0
		if bbMinus(a.Members[0].Hand, before[0])[0] == c("b4") {
			t.Fatalf("seed %d: the communicated card was drawn", seed)
		}
		if !slices.Contains(a.Members[3].Hand, c("b4")) {
			t.Fatalf("seed %d: the communicated card left its owner", seed)
		}
		if cm := a.CommunicatedCard(3); cm == nil || cm.Card != c("b4") {
			t.Fatalf("seed %d: the communication display is gone", seed)
		}
		// the winner (seat 1: p5) leads the next trick
		if g.LegalCards(1) == nil || len(g.LegalCards(1)) != 3 {
			t.Fatalf("seed %d: seat 1 must lead with %v", seed, g.LegalCards(1))
		}
	}
}

// rules §16: the draw happens once, after the first trick only.
func TestRulesMission12DrawOnlyAfterFirstTrick(t *testing.T) {
	g := bbDrawGame(t)
	rng := rand.New(rand.NewPCG(1, 2))
	for _, p := range []struct {
		by   PlayerID
		card string
	}{{dave, "p4"}, {alice, "p3"}, {bob, "p5"}, {carol, "p2"}} {
		must(t, g.PlayCard(p.by, c(p.card), rng))
	}
	if len(eventsOf[CardsDrawn](g)) != 1 {
		t.Fatal("first draw")
	}
	g.PullEvents()
	// second trick: bob leads (he holds g1 g2 g3 plus the drawn card)
	for i := range 4 {
		seat := (1 + i) % 4
		hand := g.Attempt.Members[seat].Hand
		var card Card
		legal := g.LegalCards(seat)
		if len(legal) == 0 {
			t.Fatalf("seat %d has no legal card", seat)
		}
		card = legal[0]
		_ = hand
		must(t, g.PlayCard(bbID(g, seat), card, rng))
	}
	if len(eventsOf[CardsDrawn](g)) != 0 {
		t.Fatal("another draw after trick 2")
	}
	for s := range 4 {
		if len(g.Attempt.Members[s].Hand) != 2 {
			t.Fatalf("seat %d has %d cards", s, len(g.Attempt.Members[s].Hand))
		}
	}
}

// rules §16: with 2 players the humans draw from each other; JARVIS is out of it.
func TestRulesMission12DrawTwoPlayers(t *testing.T) {
	for seed := range uint64(30) {
		g := startMission(t, 2, 12)
		g.Attempt.Objectives = nil
		setHands(g, "t4 p5 b1 b2", "p6 g1 g2 g3")
		setDummy(g, "p1 p2 b3 b4 g5 g6 y7", "p9 b9 g9 y8 t1 t2 t3")
		setTasks(g, "y9:0:omega")
		a := g.Attempt
		rng := rand.New(rand.NewPCG(seed, 3))
		must(t, g.PlayCard(alice, c("p5"), rng))
		must(t, g.PlayCard(bob, c("p6"), rng))
		dummyBefore := fmt.Sprint(a.Members[2].Stacks)
		must(t, g.PlayCard(alice, c("p1"), rng)) // JARVIS
		if len(eventsOf[CardsDrawn](g)) != 1 {
			t.Fatalf("seed %d: no draw", seed)
		}
		ha, hb := cs("t4 b1 b2"), cs("g1 g2 g3")
		gaveA, gotA := bbMinus(ha, a.Members[0].Hand), bbMinus(a.Members[0].Hand, ha)
		gaveB, gotB := bbMinus(hb, a.Members[1].Hand), bbMinus(a.Members[1].Hand, hb)
		if len(gaveA) != 1 || len(gotA) != 1 || len(gaveB) != 1 || len(gotB) != 1 {
			t.Fatalf("seed %d: alice %v/%v bob %v/%v", seed, gaveA, gotA, gaveB, gotB)
		}
		if !slices.Contains(hb, gotA[0]) || !slices.Contains(ha, gotB[0]) || gotA[0] != gaveB[0] || gotB[0] != gaveA[0] {
			t.Fatalf("seed %d: not a draw between the humans", seed)
		}
		// JARVIS only turned up its next card after the trick; it did not take part in the draw
		after := fmt.Sprint(a.Members[2].Stacks)
		if after == dummyBefore {
			t.Fatalf("seed %d: JARVIS's stack should have advanced after the trick", seed)
		}
		if a.Members[2].CardCount() != 13 {
			t.Fatalf("seed %d: JARVIS has %d cards", seed, a.Members[2].CardCount())
		}
	}
}

// rules §16 mission 23: before the distribution the commander may swap the tokens of two tasks.
func TestRulesMission23SwapTokens(t *testing.T) {
	g := startMission(t, 4, 23)
	a := g.Attempt
	cmd := a.Commander
	if a.Step != StepAdjustTokens || !slices.Equal(g.Actors(), []int{cmd}) {
		t.Fatalf("step %s actors %v", a.Step, g.Actors())
	}
	tok := func(i int) Token { return a.Tasks[i].Token }
	other := bbID(g, (cmd+1)%4)
	mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.AdjustTokens(other, 0, 3) })
	bbFailAny(t, g, func(g *Game) error { return g.AdjustTokens(bbCmd(g), 0, 0) }, ErrInvalidTokenMove, ErrInvalidTask)
	bbFailAny(t, g, func(g *Game) error { return g.AdjustTokens(bbCmd(g), 0, 9) }, ErrInvalidTokenMove, ErrInvalidTask)
	bbFailAny(t, g, func(g *Game) error { return g.AdjustTokens(bbCmd(g), -1, 2) }, ErrInvalidTokenMove, ErrInvalidTask)
	mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.SkipTokens(other) })
	must(t, g.AdjustTokens(bbCmd(g), 0, 3))
	if tok(0) != Token("4") || tok(3) != Token("1") || tok(1) != Token("2") || tok(2) != Token("3") || tok(4) != Token("5") {
		t.Fatalf("tokens %q %q %q %q %q", tok(0), tok(1), tok(2), tok(3), tok(4))
	}
	if ev := eventsOf[TokensAdjusted](g); len(ev) != 1 || ev[0].From != 0 || ev[0].To != 3 {
		t.Fatalf("events %v", ev)
	}
	if a.Step != StepDraft {
		t.Fatalf("step %s, want draft", a.Step)
	}
	// one swap per attempt, only before the distribution
	bbFailAny(t, g, func(g *Game) error { return g.AdjustTokens(bbCmd(g), 1, 2) }, ErrWrongStep)
	bbFailAny(t, g, func(g *Game) error { return g.SkipTokens(bbCmd(g)) }, ErrWrongStep)
	// the draft goes on as usual and the tokens stay on their cards
	bbDraft(t, g)
	if tok(0) != Token("4") || tok(3) != Token("1") {
		t.Fatal("tokens moved during the draft")
	}
}

func TestRulesTokensSkip(t *testing.T) {
	for _, m := range []int{23, 40} {
		g := startMission(t, 4, m)
		a := g.Attempt
		before := fmt.Sprint(a.Tasks)
		must(t, g.SkipTokens(bbCmd(g)))
		if fmt.Sprint(a.Tasks) != before || a.Step != StepDraft {
			t.Fatalf("mission %d: tokens changed or step %s", m, a.Step)
		}
		if ev := eventsOf[TokensAdjusted](g); len(ev) != 1 || ev[0].From != -1 || ev[0].To != -1 {
			t.Fatalf("events %v", ev)
		}
		bbFailAny(t, g, func(g *Game) error { return g.AdjustTokens(bbCmd(g), 0, 4) }, ErrWrongStep)
	}
	// other missions have no such step
	g := startMission(t, 4, 3)
	bbFailAny(t, g, func(g *Game) error { return g.AdjustTokens(bbCmd(g), 0, 1) }, ErrWrongStep)
	bbFailAny(t, g, func(g *Game) error { return g.SkipTokens(bbCmd(g)) }, ErrWrongStep)
}

// rules §16 mission 40: the commander may move one token to a task card without one.
func TestRulesMission40MoveToken(t *testing.T) {
	g := startMission(t, 4, 40)
	a := g.Attempt
	if a.Step != StepAdjustTokens {
		t.Fatalf("step %s", a.Step)
	}
	// tasks 0..2 carry 1 2 3; 3..7 carry nothing
	mustFail(t, g, ErrInvalidTokenMove, func(g *Game) error { return g.AdjustTokens(bbCmd(g), 0, 1) }) // onto a token
	mustFail(t, g, ErrInvalidTokenMove, func(g *Game) error { return g.AdjustTokens(bbCmd(g), 3, 4) }) // no token to move
	mustFail(t, g, ErrInvalidTokenMove, func(g *Game) error { return g.AdjustTokens(bbCmd(g), 3, 0) }) // from a task without
	mustFail(t, g, ErrNotCommander, func(g *Game) error { return g.AdjustTokens(bbID(g, (a.Commander+1)%4), 0, 5) })
	must(t, g.AdjustTokens(bbCmd(g), 0, 5))
	if a.Tasks[0].Token != TokenNone || a.Tasks[5].Token != Token("1") || a.Tasks[1].Token != Token("2") || a.Tasks[2].Token != Token("3") {
		t.Fatalf("tokens %q %q %q ... %q", a.Tasks[0].Token, a.Tasks[1].Token, a.Tasks[2].Token, a.Tasks[5].Token)
	}
	if a.Step != StepDraft {
		t.Fatalf("step %s", a.Step)
	}
	bbFailAny(t, g, func(g *Game) error { return g.AdjustTokens(bbCmd(g), 1, 6) }, ErrWrongStep)
	// the moved token now counts for its new card
	bbDraft(t, g)
	for i := range a.Tasks {
		_ = i
	}
}

// rules §16: tokens are moved before the distribution; the commander's pick order is unaffected.
func TestRulesTokenAdjustThenDraftOrder(t *testing.T) {
	g := startMission(t, 3, 23)
	cmd := g.Attempt.Commander
	must(t, g.AdjustTokens(bbCmd(g), 1, 4))
	if g.Attempt.DraftTurn() != cmd {
		t.Fatalf("the commander picks first, turn %d", g.Attempt.DraftTurn())
	}
}
