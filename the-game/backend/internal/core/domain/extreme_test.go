package domain

import (
	"reflect"
	"slices"
	"testing"
)

// Card values: see the comment at the top of commands_test.go.

// playOK plays value on pile as the active player and fails the test on error.
func playOK(t *testing.T, g *Game, value, pile int) {
	t.Helper()
	mustOK(t, g.PlayCard(g.Players[g.Turn.ActiveSeat].ID, value, pile))
}

// expectPlayErr checks that the active player's PlayCard(value, pile) fails with want
// and leaves the game unchanged.
func expectPlayErr(t *testing.T, g *Game, want error, value, pile int) {
	t.Helper()
	by := g.Players[g.Turn.ActiveSeat].ID
	expectErr(t, g, want, func(g *Game) error { return g.PlayCard(by, value, pile) })
}

// expectEndErr checks that the active player's EndTurn fails with want and changes nothing.
func expectEndErr(t *testing.T, g *Game, want error) {
	t.Helper()
	by := g.Players[g.Turn.ActiveSeat].ID
	expectErr(t, g, want, func(g *Game) error { return g.EndTurn(by) })
}

// legalPilesOf returns the piles LegalMoves allows for card value v (active player).
func legalPilesOf(g *Game, v int) []int {
	var out []int
	for _, m := range g.LegalMoves(g.Players[g.Turn.ActiveSeat].ID) {
		if m.Card == v {
			out = append(out, m.Pile)
		}
	}
	return out
}

func TestNoBackwardsTrick(t *testing.T) {
	// NBT (70) visible on pile 2; pile 0 shows 40, pile 3 shows 54.
	nbt := func() *Game {
		return table{opts: extreme, hands: [][]int{{30, 50, 64, 80}, {99}}, draw: []int{90}, piles: [][]int{{40}, nil, {70}, {54}}}.build(t)
	}
	blocked := []struct {
		name       string
		card, pile int
	}{
		{"trick on an ascending pile", 30, 0},
		{"trick on a descending pile", 64, 3},
		{"trick onto the No backwards trick card itself (Q2)", 80, 2},
	}
	for _, tt := range blocked {
		t.Run(tt.name, func(t *testing.T) {
			expectPlayErr(t, nbt(), ErrBackwardsTrickBlocked, tt.card, tt.pile)
		})
	}

	t.Run("normal plays allowed, LegalMoves without tricks", func(t *testing.T) {
		g := nbt()
		want := []Move{{30, 1}, {30, 2}, {30, 3}, {50, 0}, {50, 1}, {50, 2}, {50, 3}, {64, 0}, {64, 1}, {64, 2}, {80, 0}, {80, 1}}
		if got := g.LegalMoves(pid(0)); !reflect.DeepEqual(got, want) {
			t.Fatalf("legal moves %v\nwant %v", got, want)
		}
		playOK(t, g, 50, 0)
		playOK(t, g, 30, 1)
	})

	t.Run("NBT card placed via the trick (Q3), then the next trick is blocked", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{21, 64, 99}, {98}}, draw: []int{90}, piles: [][]int{{31}, nil, nil, {54}}}.build(t)
		playOK(t, g, 21, 0) // 31 - 10
		expectPlayErr(t, g, ErrBackwardsTrickBlocked, 64, 3)
	})

	t.Run("NBT card via the trick while another NBT is visible (Q3)", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{21, 99}, {98}}, draw: []int{90}, piles: [][]int{{31}, nil, {70}}}.build(t)
		expectPlayErr(t, g, ErrBackwardsTrickBlocked, 21, 0)
	})

	t.Run("NBT played mid-turn: the earlier trick stays, later tricks are blocked", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{30, 59, 64, 99}, {98}}, draw: []int{90}, piles: [][]int{{40}, nil, nil, {54}}}.build(t)
		playOK(t, g, 30, 0) // trick, before the NBT card
		playOK(t, g, 59, 1)
		expectPlayErr(t, g, ErrBackwardsTrickBlocked, 64, 3)
		if got := values(g.Piles[0].Cards); !reflect.DeepEqual(got, []int{40, 30}) {
			t.Fatalf("pile 0 %v", got)
		}
	})

	t.Run("NBT covered mid-turn: the trick is legal again", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{64, 69, 99}, {98}}, draw: []int{90}, piles: [][]int{nil, nil, {70}, {54}}}.build(t)
		expectPlayErr(t, g, ErrBackwardsTrickBlocked, 64, 3)
		playOK(t, g, 69, 2) // covers the NBT card
		playOK(t, g, 64, 3)
	})
}

func TestSinglePile(t *testing.T) {
	t.Run("visible at start: the first card picks the pile", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{30, 40, 96, 99}, {98}}, draw: []int{90}, piles: [][]int{nil, {9}}}.build(t)
		if got := legalPilesOf(g, 30); !reflect.DeepEqual(got, []int{0, 1, 2, 3}) {
			t.Fatalf("first card: legal piles %v", got)
		}
		playOK(t, g, 40, 2)
		expectPlayErr(t, g, ErrSinglePileViolation, 30, 0)
		expectPlayErr(t, g, ErrSinglePileViolation, 96, 0)
		playOK(t, g, 30, 2)
		// 96 and 99 fit only on other piles now: no legal move, but the turn can end.
		if g.LegalMoves(pid(0)) != nil || g.CanEndTurn() != nil || g.ActivePlayerStuck() {
			t.Fatalf("moves %v, can end %v", g.LegalMoves(pid(0)), g.CanEndTurn())
		}
	})

	t.Run("first card covers the Single pile card: free (Q1)", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{30, 40, 96, 99}, {98}}, draw: []int{90}, piles: [][]int{nil, {9}}}.build(t)
		playOK(t, g, 30, 1)
		playOK(t, g, 96, 2)
		playOK(t, g, 40, 0)
	})

	t.Run("played mid-turn: earlier cards stay, the next card goes on its pile, then free (Q1)", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{30, 47, 50, 60, 99}, {98}}, draw: []int{90}}.build(t)
		playOK(t, g, 30, 0)
		playOK(t, g, 47, 1)
		if got := values(g.Piles[0].Cards); !reflect.DeepEqual(got, []int{30}) {
			t.Fatalf("pile 0 %v", got)
		}
		expectPlayErr(t, g, ErrSinglePileViolation, 60, 0)
		if got := legalPilesOf(g, 60); !reflect.DeepEqual(got, []int{1}) {
			t.Fatalf("legal piles for 60: %v", got)
		}
		playOK(t, g, 50, 1) // covers the Single pile card
		playOK(t, g, 60, 0) // free again
	})

	t.Run("backwards trick on the bound pile is allowed", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{20, 30, 99}, {98}}, draw: []int{90}, piles: [][]int{{40}, nil, nil, {84}}}.build(t)
		playOK(t, g, 30, 0) // trick, binds pile 0
		playOK(t, g, 20, 0) // trick again
		expectPlayErr(t, g, ErrSinglePileViolation, 99, 1)
	})

	t.Run("with No backwards trick visible the trick is blocked on the bound pile too", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{40, 50, 99}, {98}}, draw: []int{90}, piles: [][]int{nil, nil, {70}, {84}}}.build(t)
		playOK(t, g, 50, 0) // binds pile 0
		expectPlayErr(t, g, ErrBackwardsTrickBlocked, 40, 0)
		expectPlayErr(t, g, ErrSinglePileViolation, 40, 1) // single pile is checked before the trick
	})
}

func TestSeveralInfinityActive(t *testing.T) {
	// NBT on pile 0, Single pile on 1, Draw only one on 2, No communication on 3.
	g := table{
		opts:  extreme,
		hands: [][]int{{11, 40, 50, 60, 62, 63, 64, 66}},
		draw:  []int{93, 94, 96},
		piles: [][]int{{21}, {9}, {76}, {92}},
	}.build(t)
	v, err := g.ViewFor(pid(0))
	mustOK(t, err)
	want := []ActiveCommand{{CommandNoBackwardsTrick, 0}, {CommandSinglePile, 1}, {CommandDrawOnlyOne, 2}, {CommandNoCommunication, 3}}
	if !reflect.DeepEqual(v.ActiveCommands, want) || !v.CommunicationRestricted || !g.CommunicationRestricted() {
		t.Fatalf("active %v, restricted %v", v.ActiveCommands, v.CommunicationRestricted)
	}
	expectPlayErr(t, g, ErrBackwardsTrickBlocked, 11, 0)
	playOK(t, g, 40, 0) // covers NBT, binds pile 0
	expectPlayErr(t, g, ErrSinglePileViolation, 50, 2)
	playOK(t, g, 50, 0)
	g.PullEvents()
	mustOK(t, g.EndTurn(pid(0)))
	evs := g.PullEvents()
	if len(evs) == 0 || evs[0] != (TurnEnded{PlayerID: pid(0), CardsDrawn: 1}) {
		t.Fatalf("events %#v", evs)
	}
	if len(g.Players[0].Hand) != 7 || len(g.DrawPile) != 2 {
		t.Fatalf("hand %v, draw %v", values(g.Players[0].Hand), values(g.DrawPile))
	}
}

func TestExactlyThreePlacement(t *testing.T) {
	tests := []struct {
		name  string
		hand  []int
		piles [][]int
		plays []Play
		card  int
		pile  int
		want  error
	}{
		{"4th card after 3! as 1st", []int{30, 99}, [][]int{{19, 20, 22}}, []Play{play(19, 0), play(20, 0), play(22, 0)}, 30, 0, ErrExactCountExceeded},
		{"4th card after 3! as 2nd", []int{30, 99}, [][]int{{18, 19, 20}}, []Play{play(18, 0), play(19, 0), play(20, 0)}, 30, 0, ErrExactCountExceeded},
		{"4th card after 3! as 3rd", []int{30, 99}, [][]int{{16, 18, 19}}, []Play{play(16, 0), play(18, 0), play(19, 0)}, 30, 0, ErrExactCountExceeded},
		{"4th card on another pile", []int{30, 99}, [][]int{{19, 20, 22}}, []Play{play(19, 0), play(20, 0), play(22, 0)}, 99, 3, ErrExactCountExceeded},
		{"4th card is a STOP", []int{95, 99}, [][]int{{19, 20, 22}}, []Play{play(19, 0), play(20, 0), play(22, 0)}, 95, 0, ErrExactCountExceeded},
		{"3! as the 4th card", []int{28, 99}, [][]int{{10, 11, 12}}, []Play{play(10, 0), play(11, 0), play(12, 0)}, 28, 0, ErrExactCountExceeded},
		{"3! as the 5th card", []int{28, 99}, [][]int{{10, 11, 12, 14}}, []Play{play(10, 0), play(11, 0), play(12, 0), play(14, 0)}, 28, 0, ErrExactCountExceeded},
		{"3! with start hand 1", []int{28}, nil, nil, 28, 0, ErrExactCountUnreachable},
		{"3! with start hand 2", []int{28, 99}, nil, nil, 28, 0, ErrExactCountUnreachable},
		{"3! as 2nd card with start hand 2", []int{28}, [][]int{{10}}, []Play{play(10, 0)}, 28, 0, ErrExactCountUnreachable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := table{opts: extreme, hands: [][]int{tt.hand, {98}}, draw: []int{90}, piles: tt.piles, plays: tt.plays}.build(t)
			expectPlayErr(t, g, tt.want, tt.card, tt.pile)
		})
	}

	t.Run("start hand 3 is enough", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{28, 30, 40}, {98}}, draw: []int{90}}.build(t)
		playOK(t, g, 28, 0)
		if g.commands().exactCount != 3 {
			t.Fatal("3! not registered")
		}
	})

	t.Run("draw pile empty (minimum 1): still exactly 3", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{28, 30, 40, 50}, {98}}}.build(t)
		playOK(t, g, 28, 0)
		expectEndErr(t, g, ErrExactCountRequired)
		playOK(t, g, 30, 0)
		expectEndErr(t, g, ErrExactCountRequired)
		playOK(t, g, 40, 0)
		if err := g.CanEndTurn(); err != nil {
			t.Fatalf("CanEndTurn = %v", err)
		}
		expectPlayErr(t, g, ErrExactCountExceeded, 50, 0)
	})

	t.Run("Pro (minimum 3) is consistent", func(t *testing.T) {
		g := table{opts: extremePro, hands: [][]int{{28, 30, 40, 50}, {98}}, draw: []int{90}}.build(t)
		playOK(t, g, 28, 0)
		playOK(t, g, 30, 0)
		expectEndErr(t, g, ErrExactCountRequired)
		playOK(t, g, 40, 0)
		if err := g.CanEndTurn(); err != nil {
			t.Fatalf("CanEndTurn = %v", err)
		}
		expectPlayErr(t, g, ErrExactCountExceeded, 50, 0)
	})

	t.Run("two 3! in one turn", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{19, 28, 30, 40}, {98}}, draw: []int{90}}.build(t)
		playOK(t, g, 19, 0)
		playOK(t, g, 28, 0)
		playOK(t, g, 30, 0)
		if err := g.CanEndTurn(); err != nil {
			t.Fatalf("CanEndTurn = %v", err)
		}
		expectPlayErr(t, g, ErrExactCountExceeded, 40, 0)
	})
}

func TestExactlyThreeEndTurn(t *testing.T) {
	ends := []struct {
		name  string
		piles [][]int
		plays []Play
		want  error
	}{
		{"1 play (3!)", [][]int{{19}}, []Play{play(19, 0)}, ErrExactCountRequired},
		{"2 plays", [][]int{{19, 20}}, []Play{play(19, 0), play(20, 0)}, ErrExactCountRequired},
		{"2 plays, 3! second", [][]int{{18, 19}}, []Play{play(18, 0), play(19, 0)}, ErrExactCountRequired},
		{"3 plays", [][]int{{18, 19, 20}}, []Play{play(18, 0), play(19, 0), play(20, 0)}, nil},
	}
	for _, tt := range ends {
		t.Run(tt.name, func(t *testing.T) {
			g := table{opts: extreme, hands: [][]int{{30, 40, 50}, {98}}, draw: []int{90, 91, 93}, piles: tt.piles, plays: tt.plays}.build(t)
			if tt.want != nil {
				expectEndErr(t, g, tt.want)
				return
			}
			mustOK(t, g.EndTurn(pid(0)))
			if g.Turn.ActiveSeat != 1 || len(g.Players[0].Hand) != 6 {
				t.Fatalf("turn %+v hand %v", g.Turn, values(g.Players[0].Hand))
			}
		})
	}

	t.Run("hand empties at exactly 3: no auto-end, End turn draws 3", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{19, 20, 22}, {98}}, draw: []int{90, 91, 93}}.build(t)
		playOK(t, g, 19, 0)
		playOK(t, g, 20, 0)
		g.PullEvents()
		playOK(t, g, 22, 0)
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{CardPlayed{PlayerID: pid(0), Card: NumberCard(22), Pile: 0}}) {
			t.Fatalf("events %#v", evs)
		}
		mustOK(t, g.EndTurn(pid(0)))
		want := []Event{
			TurnEnded{PlayerID: pid(0), CardsDrawn: 3},
			TurnStarted{PlayerID: pid(1), Number: 2},
		}
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, want) {
			t.Fatalf("events %#v", evs)
		}
	})

	t.Run("3 reached with cards left: not stuck, the player ends the turn", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{19, 20, 22, 30}, {98}}, draw: []int{90}}.build(t)
		playOK(t, g, 19, 0)
		playOK(t, g, 20, 0)
		playOK(t, g, 22, 0)
		if g.LegalMoves(pid(0)) != nil || g.CanEndTurn() != nil || g.ActivePlayerStuck() {
			t.Fatalf("moves %v can end %v", g.LegalMoves(pid(0)), g.CanEndTurn())
		}
		if g.Phase != PhasePlaying || g.Turn.Number != 1 || g.Turn.ActiveSeat != 0 {
			t.Fatalf("phase %s turn %+v", g.Phase, g.Turn)
		}
		mustOK(t, g.EndTurn(pid(0)))
	})

	t.Run("stranded 3!: the only card at turn start is lost at once", func(t *testing.T) {
		// p0 ends the turn; p1 starts with only a 3! (start hand 1) and the draw pile is empty.
		g := table{opts: extreme, hands: [][]int{{30}, {81}}, piles: [][]int{{20}}, plays: []Play{play(20, 0)}}.build(t)
		mustOK(t, g.EndTurn(pid(0)))
		want := Result{Score: 2, CardsLeftInHands: 2, Reason: LossStuck, PlayerID: pid(1)}
		if g.Phase != PhaseLost || *g.Result != want {
			t.Fatalf("phase %s result %+v", g.Phase, g.Result)
		}
		wantEv := []Event{
			TurnEnded{PlayerID: pid(0)},
			TurnStarted{PlayerID: pid(1), Number: 2},
			GameLost{Reason: LossStuck, PlayerID: pid(1), Score: 2},
		}
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, wantEv) {
			t.Fatalf("events %#v", evs)
		}
	})
}

func TestStop(t *testing.T) {
	t.Run("STOP as the first card: no further card, minimum overridden, End turn", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{30, 40, 50, 60, 62, 63, 95}, {99}}, draw: []int{10, 11, 12}}.build(t)
		playOK(t, g, 95, 0)
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{CardPlayed{PlayerID: pid(0), Card: xcard(95), Pile: 0}}) {
			t.Fatalf("events %#v", evs)
		}
		if g.Turn.Number != 1 || !g.commands().stopPlayed || g.CanEndTurn() != nil || g.LegalMoves(pid(0)) != nil || g.ActivePlayerStuck() {
			t.Fatalf("turn %+v can end %v moves %v", g.Turn, g.CanEndTurn(), g.LegalMoves(pid(0)))
		}
		expectPlayErr(t, g, ErrStopPlayed, 30, 1)
		expectPlayErr(t, g, ErrStopPlayed, 62, 2)
		mustOK(t, g.EndTurn(pid(0)))
		want := []Event{
			TurnEnded{PlayerID: pid(0), CardsDrawn: 1},
			TurnStarted{PlayerID: pid(1), Number: 2},
		}
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, want) {
			t.Fatalf("events %#v", evs)
		}
		if g.Revision != 3 || g.Turn.ActiveSeat != 1 || g.Turn.Plays != nil {
			t.Fatalf("rev %d turn %+v", g.Revision, g.Turn)
		}
		if got := values(g.Players[0].Hand); !reflect.DeepEqual(got, []int{12, 30, 40, 50, 60, 62, 63}) {
			t.Fatalf("hand %v", got)
		}
	})

	t.Run("STOP undone: cards may be played again", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{30, 40, 95}, {99}}, draw: []int{10}}.build(t)
		playOK(t, g, 95, 0)
		mustOK(t, g.UndoPlay(pid(0)))
		if g.commands().stopPlayed || g.CanEndTurn() == nil {
			t.Fatalf("STOP still counts after undo: can end %v", g.CanEndTurn())
		}
		playOK(t, g, 30, 0)
		playOK(t, g, 40, 0)
		mustOK(t, g.EndTurn(pid(0)))
	})

	t.Run("STOP as the second card with Pro (minimum 3)", func(t *testing.T) {
		g := table{opts: extremePro, hands: [][]int{{40, 50, 95}, {99}}, draw: []int{90}, piles: [][]int{{30}}, plays: []Play{play(30, 0)}}.build(t)
		playOK(t, g, 95, 1)
		if g.Turn.Number != 1 || g.CanEndTurn() != nil || g.LegalMoves(pid(0)) != nil {
			t.Fatalf("turn %+v can end %v", g.Turn, g.CanEndTurn())
		}
		mustOK(t, g.EndTurn(pid(0)))
		if g.Turn.ActiveSeat != 1 || g.Turn.Number != 2 || g.Phase != PhasePlaying {
			t.Fatalf("turn %+v", g.Turn)
		}
	})

	t.Run("STOP in solo: same player, new turn", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{40, 95}}, draw: []int{90, 91}}.build(t)
		g.PullEvents()
		playOK(t, g, 95, 0)
		mustOK(t, g.EndTurn(pid(0)))
		evs := g.PullEvents()
		if g.Turn.ActiveSeat != 0 || g.Turn.Number != 2 || g.Turn.StartHandSize != 3 ||
			!reflect.DeepEqual(evs[len(evs)-1], TurnStarted{PlayerID: pid(0), Number: 2}) {
			t.Fatalf("turn %+v events %#v", g.Turn, evs)
		}
	})

	t.Run("STOP as the last card overall: won, score 0", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{95}, {}}}.build(t)
		playOK(t, g, 95, 0)
		if g.Phase != PhaseWon || *g.Result != (Result{}) {
			t.Fatalf("phase %s result %+v", g.Phase, g.Result)
		}
		want := []Event{CardPlayed{PlayerID: pid(0), Card: xcard(95), Pile: 0}, GameWon{}}
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, want) {
			t.Fatalf("events %#v", evs)
		}
	})

	t.Run("STOP as the last hand card (draw pile empty): the turn waits for End turn", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{95}, {30}}}.build(t)
		playOK(t, g, 95, 0)
		if g.Turn.Number != 1 || g.Phase != PhasePlaying || g.ActivePlayerStuck() {
			t.Fatalf("phase %s turn %+v", g.Phase, g.Turn)
		}
		mustOK(t, g.EndTurn(pid(0)))
		want := []Event{
			CardPlayed{PlayerID: pid(0), Card: xcard(95), Pile: 0},
			TurnEnded{PlayerID: pid(0)},
			TurnStarted{PlayerID: pid(1), Number: 2},
		}
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, want) {
			t.Fatalf("events %#v", evs)
		}
	})

	vetoes := []struct {
		name  string
		piles [][]int
		plays []Play
		card  int
		pile  int
		want  error
	}{
		{"after 3! as the 2nd card", [][]int{{19}}, []Play{play(19, 0)}, 95, 0, ErrStopTooEarly},
		{"after 3! as the 2nd card, other pile", [][]int{{19}}, []Play{play(19, 0)}, 95, 3, ErrStopTooEarly},
		{"low STOP after 3! as the 2nd card", [][]int{nil, nil, {19}}, []Play{play(19, 2)}, 6, 2, ErrStopTooEarly},
		{"skull from this turn on another pile", [][]int{{41}}, []Play{play(41, 0)}, 95, 1, ErrSkullUncovered},
		{"skull from this turn on another pile after 3! as the 3rd card", [][]int{{41}, {19}}, []Play{play(19, 1), play(41, 0)}, 95, 1, ErrSkullUncovered},
	}
	for _, tt := range vetoes {
		t.Run(tt.name, func(t *testing.T) {
			g := table{opts: extreme, hands: [][]int{{6, 95, 99}, {98}}, draw: []int{90}, piles: tt.piles, plays: tt.plays}.build(t)
			expectPlayErr(t, g, tt.want, tt.card, tt.pile)
		})
	}

	allowed := []struct {
		name  string
		piles [][]int
		plays []Play
		card  int
		pile  int
	}{
		{"3! then STOP as the 3rd card", [][]int{{19, 20}}, []Play{play(19, 0), play(20, 0)}, 95, 0},
		{"STOP on the skull's own pile", [][]int{{41}}, []Play{play(41, 0)}, 95, 0},
		{"skull from an earlier turn showing", [][]int{{41}}, nil, 95, 1},
		{"3!, skull, STOP on the skull as the 3rd card", [][]int{{41}, {19}}, []Play{play(19, 1), play(41, 0)}, 95, 0},
	}
	for _, tt := range allowed {
		t.Run(tt.name, func(t *testing.T) {
			g := table{opts: extreme, hands: [][]int{{95, 99}, {98}}, draw: []int{90}, piles: tt.piles, plays: tt.plays}.build(t)
			playOK(t, g, tt.card, tt.pile)
			if g.CanEndTurn() != nil || g.LegalMoves(pid(0)) != nil {
				t.Fatalf("after STOP: can end %v moves %v", g.CanEndTurn(), g.LegalMoves(pid(0)))
			}
			mustOK(t, g.EndTurn(pid(0)))
		})
	}
}

func TestSkull(t *testing.T) {
	// p0 has played 30 and the skull 41 on pile 1 (minimum met).
	skulled := func() *Game {
		return table{
			opts:  extreme,
			hands: [][]int{{45, 61, 99}, {98}},
			draw:  []int{90},
			piles: [][]int{{30}, {41}},
			plays: []Play{play(30, 0), play(41, 1)},
		}.build(t)
	}

	t.Run("cannot end the turn with the skull on top; covering it allows it", func(t *testing.T) {
		g := skulled()
		expectEndErr(t, g, ErrSkullUncovered)
		playOK(t, g, 45, 1)
		mustOK(t, g.EndTurn(pid(0)))
	})

	t.Run("skull on skull: the new skull must be covered", func(t *testing.T) {
		g := skulled()
		playOK(t, g, 61, 1)
		expectEndErr(t, g, ErrSkullUncovered)
		if got := g.commands().skullPiles; !reflect.DeepEqual(got, []int{1}) {
			t.Fatalf("skull piles %v", got)
		}
	})

	t.Run("no covering card: stuck; undo or give up", func(t *testing.T) {
		// 29 fits nowhere once the skull lies on pile 1.
		g := table{
			opts:  extreme,
			hands: [][]int{{29, 41}, {98}},
			draw:  []int{50},
			piles: [][]int{{96}, {39, 40}, {5}, {4}},
			plays: []Play{play(39, 1), play(40, 1)},
		}.build(t)
		playOK(t, g, 41, 1)
		if g.Phase != PhasePlaying || !g.ActivePlayerStuck() {
			t.Fatalf("phase %s stuck %v", g.Phase, g.ActivePlayerStuck())
		}
		if v, _ := g.ViewFor(pid(0)); !v.Turn.Stuck {
			t.Fatal("view does not show the player as stuck")
		}
		// Undo is the way out ...
		undone := g.Clone()
		mustOK(t, undone.UndoPlay(pid(0)))
		if undone.ActivePlayerStuck() || undone.CanEndTurn() != nil {
			t.Fatalf("still stuck after undo: %v", undone.CanEndTurn())
		}
		// ... or giving up loses the game.
		g.PullEvents()
		mustOK(t, g.GiveUp(pid(0)))
		want := Result{Score: 3, CardsLeftInHands: 2, CardsLeftInDrawPile: 1, Reason: LossGaveUp, PlayerID: pid(0)}
		if g.Phase != PhaseLost || *g.Result != want {
			t.Fatalf("phase %s result %+v", g.Phase, g.Result)
		}
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{GameLost{Reason: LossGaveUp, PlayerID: pid(0), Score: 3}}) {
			t.Fatalf("events %#v", evs)
		}
	})

	t.Run("skull as the last card overall: stuck, not won; give up scores 0", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{41}, {}}, piles: [][]int{{30}}}.build(t)
		playOK(t, g, 41, 0)
		if g.Phase != PhasePlaying || !g.ActivePlayerStuck() {
			t.Fatalf("phase %s stuck %v", g.Phase, g.ActivePlayerStuck())
		}
		mustOK(t, g.GiveUp(pid(0)))
		if want := (Result{Reason: LossGaveUp, PlayerID: pid(0)}); g.Phase != PhaseLost || *g.Result != want {
			t.Fatalf("phase %s result %+v", g.Phase, g.Result)
		}
	})

	t.Run("skull as the last hand card while others hold cards: stuck, no auto-end", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{41}, {50}}, piles: [][]int{{30}}}.build(t)
		playOK(t, g, 41, 0)
		if g.Phase != PhasePlaying || g.Turn.Number != 1 || !g.ActivePlayerStuck() {
			t.Fatalf("phase %s turn %+v", g.Phase, g.Turn)
		}
		mustOK(t, g.GiveUp(pid(0)))
		if want := (Result{Score: 1, CardsLeftInHands: 1, Reason: LossGaveUp, PlayerID: pid(0)}); *g.Result != want {
			t.Fatalf("result %+v", g.Result)
		}
	})

	t.Run("skull covered but minimum unmet", func(t *testing.T) {
		g := table{opts: extremePro, hands: [][]int{{41, 45, 50}, {98}}, draw: []int{90}}.build(t)
		playOK(t, g, 41, 0)
		expectEndErr(t, g, ErrSkullUncovered)
		playOK(t, g, 45, 0)
		expectEndErr(t, g, ErrMinimumNotReached)
	})

	// Q5: a skull that could never be covered because 3! allows no further card is blocked.
	q5 := []struct {
		name  string
		piles [][]int
		plays []Play
		card  int
		pile  int
		want  error // nil = allowed
	}{
		{"skull as the 3rd card after 3!", [][]int{{19, 20}}, []Play{play(19, 0), play(20, 0)}, 41, 0, ErrSkullUncoverable},
		{"skull as the 3rd card after 3!, other pile", [][]int{{19, 20}}, []Play{play(19, 0), play(20, 0)}, 61, 1, ErrSkullUncoverable},
		{"3! as the 3rd card with a skull uncovered", [][]int{{30}, {41}}, []Play{play(30, 0), play(41, 1)}, 81, 0, ErrSkullUncoverable},
		{"3! as the 3rd card covering the skull", [][]int{{30}, {41}}, []Play{play(30, 0), play(41, 1)}, 53, 1, nil},
		{"3! as the 2nd card with a skull uncovered", [][]int{nil, {41}}, []Play{play(41, 1)}, 81, 0, nil},
		{"skull as the 2nd card after 3!", [][]int{{19}}, []Play{play(19, 0)}, 41, 0, nil},
	}
	for _, tt := range q5 {
		t.Run(tt.name, func(t *testing.T) {
			g := table{opts: extreme, hands: [][]int{{41, 53, 61, 81, 99}, {98}}, draw: []int{90}, piles: tt.piles, plays: tt.plays}.build(t)
			if tt.want != nil {
				expectPlayErr(t, g, tt.want, tt.card, tt.pile)
				return
			}
			playOK(t, g, tt.card, tt.pile)
		})
	}

	t.Run("3!, skull, then a plain 3rd card elsewhere: blocked (owner decision)", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{41, 50, 99}, {98}}, draw: []int{90}, piles: [][]int{{19}}, plays: []Play{play(19, 0)}}.build(t)
		playOK(t, g, 41, 1)
		expectPlayErr(t, g, ErrSkullUncoverable, 50, 0)
		playOK(t, g, 50, 1) // covering the skull as the 3rd card is fine
		if err := g.EndTurn(pid(0)); err != nil {
			t.Fatalf("EndTurn: %v", err)
		}
	})
}

func TestDrawOnlyOne(t *testing.T) {
	five := []int{50, 60, 62, 63, 64}
	tests := []struct {
		name  string
		hands [][]int
		draw  []int
		piles [][]int
		plays []Play
		want  int
	}{
		{"visible at turn end", [][]int{five}, seq(80, 90), [][]int{{30, 40}, nil, {76}}, []Play{play(30, 0), play(40, 0)}, 1},
		{"played by the active player this turn", [][]int{five}, seq(80, 90), [][]int{{30}, nil, {76}}, []Play{play(30, 0), play(76, 2)}, 1},
		{"two visible", [][]int{five}, seq(80, 90), [][]int{{13}, {30}, {76}}, []Play{play(30, 1)}, 1},
		{"visible at start but covered", [][]int{five}, seq(80, 90), [][]int{{30}, nil, {76, 72}}, []Play{play(30, 0), play(72, 2)}, 3},
		{"earlier shrunk hand, Draw only one gone: refill", [][]int{{50, 60, 62}}, seq(80, 90), [][]int{{30, 40}}, []Play{play(30, 0), play(40, 0)}, 5},
		{"2 players: refill to 7 without it", [][]int{{50, 60}, {98}}, seq(80, 90), [][]int{{30, 40}}, []Play{play(30, 0), play(40, 0)}, 5},
		{"draw pile empty", [][]int{five}, nil, [][]int{{30, 40}, nil, {76}}, []Play{play(30, 0), play(40, 0)}, 0},
		{"one card left", [][]int{five}, []int{90}, [][]int{{30, 40}, nil, {76}}, []Play{play(30, 0), play(40, 0)}, 1},
		{"hand already full", [][]int{{50, 51, 52, 54, 55, 56, 58, 60}}, seq(80, 90), [][]int{nil, nil, {76}}, nil, 0},
		{"lightning or other infinity tops do not limit the draw", [][]int{five}, seq(80, 90), [][]int{{30, 41}, {21}, {92}}, []Play{play(30, 0), play(41, 0)}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := table{opts: extreme, hands: tt.hands, draw: tt.draw, piles: tt.piles, plays: tt.plays}.build(t)
			if got := g.drawCount(); got != tt.want {
				t.Fatalf("drawCount = %d, want %d", got, tt.want)
			}
		})
	}

	t.Run("end of turn draws one; the hand shrinks below the hand size", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{30, 40, 50, 60, 62, 63, 64}, {98}}, draw: seq(80, 90), piles: [][]int{nil, nil, {76}}}.build(t)
		playOK(t, g, 30, 0)
		playOK(t, g, 40, 0)
		g.PullEvents()
		mustOK(t, g.EndTurn(pid(0)))
		if evs := g.PullEvents(); evs[0] != (TurnEnded{PlayerID: pid(0), CardsDrawn: 1}) {
			t.Fatalf("events %#v", evs)
		}
		if got := values(g.Players[0].Hand); !reflect.DeepEqual(got, []int{50, 60, 62, 63, 64, 90}) {
			t.Fatalf("hand %v", got)
		}
	})

	t.Run("Q7: a 1-card hand while the draw pile has cards: plays it, then lost", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{10, 20}}, draw: []int{30, 40}, piles: [][]int{nil, nil, {76}}}.build(t)
		playOK(t, g, 10, 0)
		playOK(t, g, 20, 0)
		mustOK(t, g.EndTurn(pid(0))) // draws only 40
		if g.Turn.Number != 2 || g.Phase != PhasePlaying || !reflect.DeepEqual(values(g.Players[0].Hand), []int{40}) || g.RequiredMinimum() != 2 {
			t.Fatalf("turn %+v hand %v", g.Turn, values(g.Players[0].Hand))
		}
		if g.ActivePlayerStuck() {
			t.Fatal("stuck while a card fits")
		}
		playOK(t, g, 40, 0)
		mustOK(t, g.GiveUp(pid(0)))
		want := Result{Score: 1, CardsLeftInDrawPile: 1, Reason: LossGaveUp, PlayerID: pid(0)}
		if g.Phase != PhaseLost || *g.Result != want {
			t.Fatalf("phase %s result %+v", g.Phase, g.Result)
		}
	})
}

func TestCommunicationRestricted(t *testing.T) {
	g := table{opts: extreme, hands: [][]int{{91, 99}, {98}}, draw: []int{90}, piles: [][]int{nil, nil, nil, {92}}}.build(t)
	if !g.CommunicationRestricted() {
		t.Fatal("not restricted while No communication is visible")
	}
	playOK(t, g, 91, 3) // covers it
	if g.CommunicationRestricted() {
		t.Fatal("still restricted after covering")
	}
	if v, _ := g.ViewFor(pid(1)); v.CommunicationRestricted {
		t.Fatal("view still restricted after covering")
	}

	t.Run("covered while another is visible", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{91, 99}, {98}}, draw: []int{90}, piles: [][]int{nil, nil, {26}, {92}}}.build(t)
		playOK(t, g, 91, 3)
		if !g.CommunicationRestricted() {
			t.Fatal("not restricted while another No communication card is visible")
		}
	})

	for _, ph := range []Phase{PhaseLobby, PhaseChoosingFirstPlayer, PhaseLost, PhaseEnded} {
		t.Run(string(ph), func(t *testing.T) {
			g := table{opts: extreme, hands: [][]int{{99}}, piles: [][]int{nil, nil, nil, {92}}}.build(t)
			g.Phase = ph
			if ph == PhaseLost {
				g.Result = &Result{Score: 1, CardsLeftInHands: 1, Reason: LossGaveUp, PlayerID: pid(0)}
			}
			if g.CommunicationRestricted() {
				t.Fatalf("restricted in %s", ph)
			}
			if v, _ := g.ViewFor(pid(0)); v.CommunicationRestricted {
				t.Fatalf("view restricted in %s", ph)
			}
		})
	}
}

// TestCommandsAreNoOpsInClassic: the same card values carry no commands in classic,
// so none of the Extreme rules apply.
func TestCommandsAreNoOpsInClassic(t *testing.T) {
	g := table{hands: [][]int{{19, 20, 22, 30, 41, 95}, {98}}, draw: []int{90}, piles: [][]int{{40}, nil, {70}, {92}}}.build(t)
	playOK(t, g, 30, 0) // trick next to a classic 70
	playOK(t, g, 41, 0)
	if err := g.CanEndTurn(); err != nil { // 41 is no skull
		t.Fatalf("CanEndTurn = %v", err)
	}
	playOK(t, g, 95, 0) // 95 is no STOP
	if g.Turn.Number != 1 || g.CommunicationRestricted() {
		t.Fatalf("turn %+v", g.Turn)
	}
	v, _ := g.ViewFor(pid(0))
	if v.ActiveCommands != nil || !reflect.DeepEqual(v.Turn.Constraints, TurnConstraints{}) ||
		slices.ContainsFunc(v.Piles, func(p PileView) bool { return p.CommandActive }) {
		t.Fatalf("classic view has command state: %+v", v)
	}
}
