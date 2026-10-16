package domain

import (
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"
)

func TestPlayCardErrors(t *testing.T) {
	playing := func() *Game {
		return table{
			hands: [][]int{{10, 20, 30, 40, 60, 70, 80}, {15, 25, 35, 45, 55, 65, 75}},
			draw:  []int{81, 82, 83},
			piles: [][]int{{50}, nil, {65}, nil},
		}.build(t)
	}
	choosing := lobbyWith(t, 2, DefaultOptions())
	mustOK(t, choosing.Start(pid(0), noShuffle{}))
	tests := []struct {
		name  string
		g     *Game
		by    PlayerID
		value int
		pile  int
		want  error
	}{
		{"unknown player", playing(), "zz", 60, 0, ErrUnknownPlayer},
		{"lobby", lobbyWith(t, 2, DefaultOptions()), pid(0), 10, 0, ErrWrongPhase},
		{"choosing first player", choosing, pid(0), 99, 0, ErrWrongPhase},
		{"finished", finishedGame(t), pid(0), 50, 3, ErrWrongPhase},
		{"not your turn", playing(), pid(1), 55, 0, ErrNotYourTurn},
		{"card not in hand", playing(), pid(0), 55, 0, ErrCardNotInHand},
		{"card nowhere", playing(), pid(0), 100, 0, ErrCardNotInHand},
		{"negative pile", playing(), pid(0), 60, -1, ErrInvalidPile},
		{"pile out of range", playing(), pid(0), 60, 4, ErrInvalidPile},
		{"up: smaller", playing(), pid(0), 10, 0, ErrIllegalPlacement},
		{"up: equal-ish not trick", playing(), pid(0), 30, 0, ErrIllegalPlacement},
		{"down: greater", playing(), pid(0), 70, 2, ErrIllegalPlacement},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.PlayCard(tt.by, tt.value, tt.pile) })
		})
	}
}

func TestPlayCardApplies(t *testing.T) {
	g := table{hands: [][]int{{10, 20, 30, 40, 60, 70, 80}, {15}}, draw: []int{81}, piles: [][]int{{50}, nil, {90}, nil}}.build(t)
	mustOK(t, g.PlayCard(pid(0), 40, 0)) // backwards trick
	mustOK(t, g.PlayCard(pid(0), 80, 2))
	if got := values(g.Players[0].Hand); !reflect.DeepEqual(got, []int{10, 20, 30, 60, 70}) {
		t.Fatalf("hand %v", got)
	}
	if got := values(g.Piles[0].Cards); !reflect.DeepEqual(got, []int{50, 40}) {
		t.Fatalf("pile 0 %v", got)
	}
	if got := values(g.Piles[2].Cards); !reflect.DeepEqual(got, []int{90, 80}) {
		t.Fatalf("pile 2 %v", got)
	}
	wantPlays := []Play{{NumberCard(40), 0}, {NumberCard(80), 2}}
	if !reflect.DeepEqual(g.Turn.Plays, wantPlays) || g.Revision != 3 || g.Phase != PhasePlaying {
		t.Fatalf("turn %+v rev %d", g.Turn, g.Revision)
	}
	wantEv := []Event{
		CardPlayed{PlayerID: pid(0), Card: NumberCard(40), Pile: 0},
		CardPlayed{PlayerID: pid(0), Card: NumberCard(80), Pile: 2},
	}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, wantEv) {
		t.Fatalf("events %#v", evs)
	}
}

func TestEndTurnErrors(t *testing.T) {
	hands := [][]int{{20, 30, 40, 50, 60}, {15, 25, 35, 45, 55, 65, 75}}
	tests := []struct {
		name string
		g    *Game
		by   PlayerID
		want error
	}{
		{"unknown player", table{hands: hands, draw: []int{81}}.build(t), "zz", ErrUnknownPlayer},
		{"lobby", lobbyWith(t, 2, DefaultOptions()), pid(0), ErrWrongPhase},
		{"finished", finishedGame(t), pid(0), ErrWrongPhase},
		{"not your turn", table{hands: hands, draw: []int{81}, piles: [][]int{{5, 10}}, plays: []Play{{NumberCard(5), 0}, {NumberCard(10), 0}}}.build(t), pid(1), ErrNotYourTurn},
		{"no card played", table{hands: hands, draw: []int{81}}.build(t), pid(0), ErrMinimumNotReached},
		{"one of two", table{hands: hands, draw: []int{81}, piles: [][]int{{10}}, plays: []Play{{NumberCard(10), 0}}}.build(t), pid(0), ErrMinimumNotReached},
		{"pro: two of three", table{opts: Options{Variant: VariantClassic, Pro: true}, hands: hands, draw: []int{81}, piles: [][]int{{5, 10}}, plays: []Play{{NumberCard(5), 0}, {NumberCard(10), 0}}}.build(t), pid(0), ErrMinimumNotReached},
		{"draw empty: none played", table{hands: hands}.build(t), pid(0), ErrMinimumNotReached},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.EndTurn(tt.by) })
		})
	}
}

func TestRequiredMinimum(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		draw []int
		want int
	}{
		{"classic", classic, []int{90}, 2},
		{"classic, draw pile empty", classic, nil, 1},
		{"pro", Options{Variant: VariantClassic, Pro: true}, []int{90}, 3},
		{"pro, draw pile empty", Options{Variant: VariantClassic, Pro: true}, nil, 1},
		{"pro reduced, draw pile empty", Options{Variant: VariantClassic, Pro: true, ReducedHandSize: true}, []int{}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := table{opts: tt.opts, hands: [][]int{{10}}, draw: tt.draw}.build(t)
			if got := g.RequiredMinimum(); got != tt.want {
				t.Fatalf("RequiredMinimum = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEndTurnRefillsToHandSize(t *testing.T) {
	g := table{
		hands: [][]int{{10, 20, 30, 40, 50, 60, 70}, {11, 21, 31, 41, 51, 61, 71}},
		draw:  seq(81, 90), // top = 90
	}.build(t)
	for _, v := range []int{10, 20, 30} {
		mustOK(t, g.PlayCard(pid(0), v, 0))
	}
	g.PullEvents()
	mustOK(t, g.EndTurn(pid(0)))
	if got := values(g.Players[0].Hand); !reflect.DeepEqual(got, []int{40, 50, 60, 70, 88, 89, 90}) {
		t.Fatalf("hand %v", got)
	}
	if got := values(g.DrawPile); !reflect.DeepEqual(got, seq(81, 87)) {
		t.Fatalf("draw pile %v", got)
	}
	want := []Event{TurnEnded{PlayerID: pid(0), CardsDrawn: 3}, TurnStarted{PlayerID: pid(1), Number: 2}}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, want) {
		t.Fatalf("events %#v", evs)
	}
	if g.Turn.ActiveSeat != 1 || g.Turn.Number != 2 || g.Turn.Plays != nil || g.Turn.StartHandSize != 7 || g.Revision != 5 {
		t.Fatalf("turn %+v rev %d", g.Turn, g.Revision)
	}
}

func TestEndTurnDrawCappedByDrawPile(t *testing.T) {
	g := table{
		hands: [][]int{{10, 20, 30, 40, 50, 60, 70}, {11, 21, 31, 41, 51, 61, 71}},
		draw:  []int{81},
	}.build(t)
	mustOK(t, g.PlayCard(pid(0), 10, 0))
	mustOK(t, g.PlayCard(pid(0), 20, 0))
	mustOK(t, g.EndTurn(pid(0)))
	if got := values(g.Players[0].Hand); !reflect.DeepEqual(got, []int{30, 40, 50, 60, 70, 81}) {
		t.Fatalf("hand %v", got)
	}
	if len(g.DrawPile) != 0 || g.RequiredMinimum() != 1 {
		t.Fatalf("draw pile %v, minimum %d", g.DrawPile, g.RequiredMinimum())
	}
	// Next player: minimum is now 1.
	mustOK(t, g.PlayCard(pid(1), 71, 2))
	mustOK(t, g.EndTurn(pid(1)))
	if g.Turn.ActiveSeat != 0 || g.Phase != PhasePlaying {
		t.Fatalf("turn %+v", g.Turn)
	}
}

func TestDrawCount(t *testing.T) {
	tests := []struct {
		name  string
		opts  Options
		hands [][]int
		draw  []int
		want  int
	}{
		{"refill solo", classic, [][]int{{10, 20}}, seq(30, 50), 6},
		{"refill pro reduced 2p", Options{Variant: VariantClassic, Pro: true, ReducedHandSize: true}, [][]int{{10, 20}, {11}}, seq(30, 50), 4},
		{"capped", classic, [][]int{{10, 20}}, []int{30, 31}, 2},
		{"empty draw", classic, [][]int{{10}}, nil, 0},
		{"hand already full", classic, [][]int{seq(10, 17)}, []int{30}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := table{opts: tt.opts, hands: tt.hands, draw: tt.draw}.build(t)
			if got := g.drawCount(); got != tt.want {
				t.Fatalf("drawCount = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTurnOrder(t *testing.T) {
	tests := []struct {
		name     string
		hands    [][]int
		seat     int
		draw     []int
		wantSeat int
		skipped  []PlayerID
	}{
		{"next clockwise", [][]int{{20, 30}, {21}, {22}}, 0, []int{80, 81}, 1, nil},
		{"wraps around", [][]int{{20}, {21}, {22, 30}}, 2, []int{80, 81}, 0, nil},
		{"solo stays", [][]int{{22, 30}}, 0, []int{80, 81}, 0, nil},
		{"skips empty hand", [][]int{{20, 30}, {}, {22}}, 0, nil, 2, []PlayerID{pid(1)}},
		{"skips several, wrapping", [][]int{{}, {21}, {}, {}, {24, 30}}, 4, nil, 1, []PlayerID{pid(0)}},
		{"skips all others: back to self", [][]int{{}, {21, 30}, {}}, 1, nil, 1, []PlayerID{pid(2), pid(0)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := table{hands: tt.hands, draw: tt.draw, seat: tt.seat, piles: [][]int{{10}}, plays: []Play{{NumberCard(10), 0}}}.build(t)
			if len(tt.draw) > 0 { // need two cards played while the draw pile has cards
				g.Piles[0].Cards = cards(5, 10)
				g.Turn.Plays = []Play{{NumberCard(5), 0}, {NumberCard(10), 0}}
			}
			mustOK(t, g.EndTurn(pid(tt.seat)))
			if g.Turn.ActiveSeat != tt.wantSeat || g.Turn.Number != 2 {
				t.Fatalf("turn %+v, want seat %d", g.Turn, tt.wantSeat)
			}
			var skipped []PlayerID
			for _, e := range g.PullEvents() {
				if s, ok := e.(PlayerSkipped); ok {
					skipped = append(skipped, s.PlayerID)
				}
			}
			if !reflect.DeepEqual(skipped, tt.skipped) {
				t.Fatalf("skipped %v, want %v", skipped, tt.skipped)
			}
		})
	}
}

func TestDrawPileEmptyMinimumOne(t *testing.T) {
	g := table{opts: Options{Variant: VariantClassic, Pro: true}, hands: [][]int{{20, 30}, {21}}}.build(t)
	mustOK(t, g.PlayCard(pid(0), 20, 0))
	if err := g.CanEndTurn(); err != nil {
		t.Fatalf("CanEndTurn = %v", err)
	}
	g.PullEvents()
	mustOK(t, g.EndTurn(pid(0)))
	want := []Event{TurnEnded{PlayerID: pid(0), CardsDrawn: 0}, TurnStarted{PlayerID: pid(1), Number: 2}}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, want) {
		t.Fatalf("events %#v", evs)
	}
	if got := values(g.Players[0].Hand); !reflect.DeepEqual(got, []int{30}) {
		t.Fatalf("hand %v", got)
	}
}

// TestEmptyHandWaitsForEndTurn: an empty hand never ends the turn by itself.
func TestEmptyHandWaitsForEndTurn(t *testing.T) {
	tests := []struct {
		name     string
		hands    [][]int
		draw     []int
		play     []int
		wantHand []int
		drawn    int
		nextSeat int
		skipped  []PlayerID
	}{
		{"draw pile empty", [][]int{{10}, {30}}, nil, []int{10}, []int{}, 0, 1, nil},
		{"draw pile empty, skip", [][]int{{10}, {}, {30}}, nil, []int{10}, []int{}, 0, 2, []PlayerID{pid(1)}},
		{"refills from draw pile", [][]int{{10, 20}, {30}}, []int{81, 82, 83}, []int{10, 20}, []int{81, 82, 83}, 3, 1, nil},
		{"solo refills and continues", [][]int{{10, 20}}, seq(81, 90), []int{10, 20}, seq(83, 90), 8, 0, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := table{hands: tt.hands, draw: tt.draw}.build(t)
			for _, v := range tt.play {
				mustOK(t, g.PlayCard(pid(0), v, 0))
			}
			if g.Phase != PhasePlaying || g.Turn.ActiveSeat != 0 || g.Turn.Number != 1 || len(g.Players[0].Hand) != 0 {
				t.Fatalf("phase %s turn %+v", g.Phase, g.Turn)
			}
			g.PullEvents()
			mustOK(t, g.EndTurn(pid(0)))
			if g.Phase != PhasePlaying || g.Turn.ActiveSeat != tt.nextSeat || g.Turn.Number != 2 {
				t.Fatalf("phase %s turn %+v", g.Phase, g.Turn)
			}
			if got := values(g.Players[0].Hand); !reflect.DeepEqual(got, tt.wantHand) {
				t.Fatalf("hand %v, want %v", got, tt.wantHand)
			}
			if g.Revision != int64(2+len(tt.play)) {
				t.Fatalf("revision %d", g.Revision)
			}
			evs := g.PullEvents()
			want := []Event{TurnEnded{PlayerID: pid(0), CardsDrawn: tt.drawn}}
			for _, s := range tt.skipped {
				want = append(want, PlayerSkipped{PlayerID: s})
			}
			want = append(want, TurnStarted{PlayerID: pid(tt.nextSeat), Number: 2})
			if got := evs; !reflect.DeepEqual(got, want) {
				t.Fatalf("events %#v\nwant %#v", got, want)
			}
		})
	}
}

func TestWin(t *testing.T) {
	g := table{hands: [][]int{{10, 20}, {}}, piles: [][]int{{5}}}.build(t)
	mustOK(t, g.PlayCard(pid(0), 10, 0))
	if g.Phase != PhasePlaying {
		t.Fatal("finished too early")
	}
	g.PullEvents()
	mustOK(t, g.PlayCard(pid(0), 20, 1))
	if g.Phase != PhaseWon || g.Revision != 3 {
		t.Fatalf("phase %s rev %d", g.Phase, g.Revision)
	}
	if want := (Result{}); *g.Result != want {
		t.Fatalf("result %+v", *g.Result)
	}
	want := []Event{CardPlayed{PlayerID: pid(0), Card: NumberCard(20), Pile: 1}, GameWon{}}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, want) {
		t.Fatalf("events %#v", evs)
	}
	// Nothing is possible after the game ended.
	expectErr(t, g, ErrWrongPhase, func(g *Game) error { return g.EndTurn(pid(0)) })
	if g.ActivePlayerStuck() || g.LegalMoves(pid(0)) != nil || g.CanEndTurn() == nil {
		t.Fatal("finished game must have no moves")
	}
}

func TestWinSolo(t *testing.T) {
	g := table{hands: [][]int{{99}}}.build(t)
	mustOK(t, g.PlayCard(pid(0), 99, 3))
	if g.Phase != PhaseWon || g.Result.Score != 0 {
		t.Fatalf("phase %s result %+v", g.Phase, g.Result)
	}
}

func TestAdvanceTurnNoCardsLeftWins(t *testing.T) {
	// Defensive path: nobody has cards when the turn advances.
	g := table{hands: [][]int{{}, {}}}.build(t)
	g.advanceTurn()
	if g.Phase != PhaseWon {
		t.Fatalf("phase %s result %+v", g.Phase, g.Result)
	}
}

// stuckTable: p0 has played 94, 95 and ends the turn (drawing 1 of 2 draw cards, so the
// minimum stays 2); piles show 95, 96 (up) and 5, 6 (down).
func stuckTable(p1 []int) table {
	return table{
		hands: [][]int{{97, 40, 41, 42, 43, 44}, p1},
		draw:  []int{52, 53},
		piles: [][]int{{94, 95}, {96}, {5}, {6}},
		plays: []Play{{NumberCard(94), 0}, {NumberCard(95), 0}},
	}
}

func TestStuckAtTurnStart(t *testing.T) {
	g := stuckTable([]int{50, 51}).build(t)
	mustOK(t, g.EndTurn(pid(0)))
	if g.Phase != PhaseLost || g.Revision != 2 {
		t.Fatalf("phase %s rev %d", g.Phase, g.Revision)
	}
	want := Result{Score: 10, CardsLeftInHands: 9, CardsLeftInDrawPile: 1, Reason: LossStuck, PlayerID: pid(1)}
	if *g.Result != want {
		t.Fatalf("result %+v, want %+v", *g.Result, want)
	}
	wantEv := []Event{
		TurnEnded{PlayerID: pid(0), CardsDrawn: 1},
		TurnStarted{PlayerID: pid(1), Number: 2},
		GameLost{Reason: LossStuck, PlayerID: pid(1), Score: 10},
	}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, wantEv) {
		t.Fatalf("events %#v", evs)
	}
	if g.Turn.ActiveSeat != 1 || g.Turn.Number != 2 {
		t.Fatalf("turn must show the stuck player: %+v", g.Turn)
	}
}

func TestNotStuckWhileACardFits(t *testing.T) {
	// p1 can play 98 but never reach the minimum of 2: no lookahead, keep playing.
	g := stuckTable([]int{50, 98}).build(t)
	mustOK(t, g.EndTurn(pid(0)))
	if g.Phase != PhasePlaying || g.ActivePlayerStuck() {
		t.Fatalf("lost too early: phase %s", g.Phase)
	}
	if got, want := g.LegalMoves(pid(1)), []Move{{98, 0}, {98, 1}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("legal moves %v, want %v", got, want)
	}
	g.PullEvents()
	mustOK(t, g.PlayCard(pid(1), 98, 1))
	// Stuck mid-turn: not lost yet, the player may undo or give up.
	if g.Phase != PhasePlaying || !g.ActivePlayerStuck() {
		t.Fatalf("phase %s stuck %v", g.Phase, g.ActivePlayerStuck())
	}
	mustOK(t, g.GiveUp(pid(1)))
	want := Result{Score: 9, CardsLeftInHands: 8, CardsLeftInDrawPile: 1, Reason: LossGaveUp, PlayerID: pid(1)}
	if g.Phase != PhaseLost || *g.Result != want {
		t.Fatalf("phase %s result %+v", g.Phase, g.Result)
	}
	wantEv := []Event{
		CardPlayed{PlayerID: pid(1), Card: NumberCard(98), Pile: 1},
		GameLost{Reason: LossGaveUp, PlayerID: pid(1), Score: 9},
	}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, wantEv) {
		t.Fatalf("events %#v", evs)
	}
}

func TestStuckWhenHandEmptiesBelowMinimum(t *testing.T) {
	// Artificial: a 1-card hand while the draw pile still has cards (minimum 2).
	g := table{hands: [][]int{{10}, {30}}, draw: []int{81}}.build(t)
	mustOK(t, g.PlayCard(pid(0), 10, 0))
	if g.Phase != PhasePlaying || !g.ActivePlayerStuck() {
		t.Fatalf("phase %s stuck %v", g.Phase, g.ActivePlayerStuck())
	}
	mustOK(t, g.GiveUp(pid(0)))
	want := Result{Score: 2, CardsLeftInHands: 1, CardsLeftInDrawPile: 1, Reason: LossGaveUp, PlayerID: pid(0)}
	if g.Phase != PhaseLost || *g.Result != want {
		t.Fatalf("phase %s result %+v", g.Phase, g.Result)
	}
}

func TestUndoPlay(t *testing.T) {
	t.Run("takes back cards one by one, last first", func(t *testing.T) {
		g := table{hands: [][]int{{10, 20, 95}, {30}}, draw: []int{81, 82}, piles: [][]int{nil, nil, {50}}}.build(t)
		start := g.Clone()
		mustOK(t, g.PlayCard(pid(0), 10, 0))
		mustOK(t, g.PlayCard(pid(0), 20, 0))
		mustOK(t, g.PlayCard(pid(0), 95, 1))
		g.PullEvents()

		mustOK(t, g.UndoPlay(pid(0)))
		if got := values(g.Players[0].Hand); !reflect.DeepEqual(got, []int{95}) || len(g.Piles[1].Cards) != 0 || len(g.Turn.Plays) != 2 {
			t.Fatalf("after 1st undo: hand %v plays %v", got, g.Turn.Plays)
		}
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{PlayUndone{PlayerID: pid(0), Card: NumberCard(95), Pile: 1}}) {
			t.Fatalf("events %#v", evs)
		}
		mustOK(t, g.UndoPlay(pid(0)))
		mustOK(t, g.UndoPlay(pid(0)))
		if !reflect.DeepEqual(g.Players, start.Players) || !reflect.DeepEqual(g.Piles, start.Piles) || !reflect.DeepEqual(g.Turn, start.Turn) {
			t.Fatalf("not back at turn start:\n%+v\n%+v", g.Turn, start.Turn)
		}
		if g.Revision != start.Revision+6 {
			t.Fatalf("revision %d", g.Revision)
		}
		expectErr(t, g, ErrNothingToUndo, func(g *Game) error { return g.UndoPlay(pid(0)) })
	})

	t.Run("plays of an ended turn are final", func(t *testing.T) {
		g := table{hands: [][]int{{10, 20, 30}, {40}}, draw: []int{81, 82}}.build(t)
		mustOK(t, g.PlayCard(pid(0), 10, 0))
		mustOK(t, g.PlayCard(pid(0), 20, 0))
		mustOK(t, g.EndTurn(pid(0)))
		expectErr(t, g, ErrNotYourTurn, func(g *Game) error { return g.UndoPlay(pid(0)) })
		expectErr(t, g, ErrNothingToUndo, func(g *Game) error { return g.UndoPlay(pid(1)) })
	})

	t.Run("not your turn, unknown player, wrong phase", func(t *testing.T) {
		g := table{hands: [][]int{{10, 20}, {40}}, draw: []int{81}, plays: []Play{play(5, 0)}, piles: [][]int{{5}}}.build(t)
		expectErr(t, g, ErrNotYourTurn, func(g *Game) error { return g.UndoPlay(pid(1)) })
		expectErr(t, g, ErrUnknownPlayer, func(g *Game) error { return g.UndoPlay("nobody") })
		lobby := lobbyWith(t, 2, classic)
		expectErr(t, lobby, ErrWrongPhase, func(g *Game) error { return g.UndoPlay(pid(0)) })
	})

	t.Run("the winning card cannot be undone", func(t *testing.T) {
		g := table{hands: [][]int{{10}, {}}}.build(t)
		mustOK(t, g.PlayCard(pid(0), 10, 0))
		expectErr(t, g, ErrWrongPhase, func(g *Game) error { return g.UndoPlay(pid(0)) })
	})
}

func TestUndoRevertsCommands(t *testing.T) {
	tests := []struct {
		name  string
		hand  []int
		piles [][]int
		plays []int // played on pile 0 in order; the last one is undone
		check func(t *testing.T, g *Game)
	}{
		{"3! no longer demands 3 cards", []int{10, 20, 81, 90}, nil, []int{10, 20, 81}, func(t *testing.T, g *Game) {
			if g.commands().exactCount != 0 || g.CanEndTurn() != nil {
				t.Fatalf("exact count %d, can end %v", g.commands().exactCount, g.CanEndTurn())
			}
		}},
		{"skull no longer needs covering", []int{10, 20, 41, 90}, nil, []int{10, 20, 41}, func(t *testing.T, g *Game) {
			if len(g.commands().skullPiles) != 0 || g.CanEndTurn() != nil {
				t.Fatalf("skull piles %v, can end %v", g.commands().skullPiles, g.CanEndTurn())
			}
		}},
		{"Single pile binding lifted", []int{10, 68, 90}, nil, []int{68}, func(t *testing.T, g *Game) {
			if g.commands().singlePile || !slices.Equal(legalPilesOf(g, 90), []int{0, 1, 2, 3}) {
				t.Fatalf("legal piles of 90: %v", legalPilesOf(g, 90))
			}
		}},
		{"covered Single pile binds again", []int{72, 99}, [][]int{{68}}, []int{72}, func(t *testing.T, g *Game) {
			if cs := g.commands(); cs.boundPile != -1 || !cs.singlePile {
				t.Fatalf("command state %+v", cs)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := table{opts: extreme, hands: [][]int{tt.hand, {98}}, draw: []int{90, 91}, piles: tt.piles}.build(t)
			for _, v := range tt.plays {
				playOK(t, g, v, 0)
			}
			mustOK(t, g.UndoPlay(pid(0)))
			tt.check(t, g)
		})
	}
}

func TestBackwardsTrickEnablesSequence(t *testing.T) {
	// 55 fits nowhere until 50 is played on 60 via the backwards trick.
	g := table{hands: [][]int{{50, 55, 20}}, draw: []int{80}, piles: [][]int{{60}, {98}, {4}, {3}}}.build(t)
	if g.ActivePlayerStuck() {
		t.Fatal("stuck at start")
	}
	if got := g.LegalMoves(pid(0)); !reflect.DeepEqual(got, []Move{{50, 0}}) {
		t.Fatalf("legal moves %v", got)
	}
	mustOK(t, g.PlayCard(pid(0), 50, 0))
	if got := g.LegalMoves(pid(0)); !reflect.DeepEqual(got, []Move{{55, 0}}) {
		t.Fatalf("legal moves %v", got)
	}
	mustOK(t, g.PlayCard(pid(0), 55, 0))
	// Minimum met, no card fits (20), but the turn can end: not stuck.
	if g.Phase != PhasePlaying || g.ActivePlayerStuck() || len(g.LegalMoves(pid(0))) != 0 || g.CanEndTurn() != nil {
		t.Fatalf("phase %s stuck %v", g.Phase, g.ActivePlayerStuck())
	}
	mustOK(t, g.EndTurn(pid(0)))
	if g.Phase != PhasePlaying || g.Turn.Number != 2 || !reflect.DeepEqual(values(g.Players[0].Hand), []int{20, 80}) {
		t.Fatalf("phase %s turn %+v hand %v", g.Phase, g.Turn, values(g.Players[0].Hand))
	}
}

func TestLegalMovesAndCanEndTurnOutsideTurn(t *testing.T) {
	g := table{hands: [][]int{{10, 95}, {20}}, draw: []int{80}, piles: [][]int{nil, {96}, {50}, nil}}.build(t)
	if got, want := g.LegalMoves(pid(0)), []Move{{10, 0}, {10, 2}, {10, 3}, {95, 0}, {95, 3}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("legal moves %v, want %v", got, want)
	}
	if g.LegalMoves(pid(1)) != nil || g.LegalMoves("zz") != nil {
		t.Fatal("LegalMoves must be nil for non-active players")
	}
	lobby := lobbyWith(t, 2, DefaultOptions())
	if lobby.LegalMoves(pid(0)) != nil || lobby.CanEndTurn() == nil || lobby.ActivePlayerStuck() {
		t.Fatal("lobby must have no moves")
	}
}

// TestRandomGames plays many random games end to end and checks invariants after every command.
func TestRandomGames(t *testing.T) {
	optsList := []Options{
		classic,
		{Variant: VariantClassic, Pro: true},
		{Variant: VariantClassic, Pro: true, ReducedHandSize: true},
		extreme,
		extremePro,
	}
	const games = 250
	won := 0
	for seed := range uint64(games) {
		rng := rand.New(rand.NewPCG(seed, 99))
		n := 1 + int(seed%5)
		g := lobbyWith(t, n, optsList[(seed/5)%uint64(len(optsList))]) // every option set with every player count
		mustOK(t, g.Start(pid(0), rng))
		if n > 1 {
			mustOK(t, g.ChooseFirstPlayer(pid(0), pid(rng.IntN(n))))
		}
		for steps := 0; g.Phase == PhasePlaying; steps++ {
			if steps > 400 {
				t.Fatalf("seed %d: game does not end", seed)
			}
			rev := g.Revision
			active := g.Players[g.Turn.ActiveSeat].ID
			moves := g.LegalMoves(active)
			stuck := len(moves) == 0 && g.CanEndTurn() != nil
			if stuck != g.ActivePlayerStuck() {
				t.Fatalf("seed %d: ActivePlayerStuck disagrees", seed)
			}
			if stuck && len(g.Turn.Plays) == 0 {
				t.Fatalf("seed %d: stuck at turn start left in playing", seed)
			}
			playedStop := false
			switch {
			case stuck:
				// Mid-turn: the bot gives up instead of undoing.
				mustOK(t, g.GiveUp(active))
			case g.CanEndTurn() == nil && (len(moves) == 0 || rng.IntN(3) == 0):
				mustOK(t, g.EndTurn(active))
			default:
				// Prefer the smallest jump (the trick counts as a negative jump).
				best := moves[0]
				jump := func(m Move) int {
					top := g.Piles[m.Pile].TopValue()
					if g.Piles[m.Pile].Direction == DirectionUp {
						return m.Card - top
					}
					return top - m.Card
				}
				for _, m := range moves {
					if jump(m) < jump(best) {
						best = m
					}
				}
				_, seat, _ := g.PlayerByID(active)
				idx := slices.IndexFunc(g.Players[seat].Hand, func(c Card) bool { return c.Value == best.Card })
				playedStop = g.Players[seat].Hand[idx].Command == CommandStop
				before := g.Clone()
				mustOK(t, g.PlayCard(active, best.Card, best.Pile))
				// Undo restores the table exactly as it was before the play.
				if g.Phase == PhasePlaying {
					u := g.Clone()
					mustOK(t, u.UndoPlay(active))
					if !reflect.DeepEqual(u.Piles, before.Piles) || !reflect.DeepEqual(u.Players, before.Players) ||
						!reflect.DeepEqual(u.Turn, before.Turn) || u.Revision != g.Revision+1 {
						t.Fatalf("seed %d: undo of %v did not restore the state", seed, best)
					}
				}
			}
			if g.Revision != rev+1 {
				t.Fatalf("seed %d: revision %d -> %d", seed, rev, g.Revision)
			}
			if playedStop && g.Phase == PhasePlaying && (g.LegalMoves(active) != nil || g.CanEndTurn() != nil) {
				t.Fatalf("seed %d: after STOP a card fits or the turn cannot end", seed)
			}
			if totalCards(g) != 98 {
				t.Fatalf("seed %d: %d cards", seed, totalCards(g))
			}
			for _, p := range g.Players {
				if !slices.IsSortedFunc(p.Hand, func(a, b Card) int { return a.Value - b.Value }) {
					t.Fatalf("seed %d: unsorted hand", seed)
				}
			}
			evs := g.PullEvents()
			if len(evs) == 0 {
				t.Fatalf("seed %d: no events", seed)
			}
			// The piles do not change while a turn ends, so this is the state at turn end.
			for _, e := range evs {
				if te, ok := e.(TurnEnded); ok && g.commands().drawOnlyOne && te.CardsDrawn > 1 {
					t.Fatalf("seed %d: drew %d cards while Draw only one was visible", seed, te.CardsDrawn)
				}
			}
		}
		h, d := g.Score()
		r := g.Result
		if r == nil || r.Score != h+d || r.CardsLeftInHands != h || r.CardsLeftInDrawPile != d {
			t.Fatalf("seed %d: result %+v vs %d+%d", seed, r, h, d)
		}
		switch g.Phase {
		case PhaseWon:
			won++
			if r.Score != 0 || r.Reason != "" || r.PlayerID != "" {
				t.Fatalf("seed %d: won with %+v", seed, r)
			}
		case PhaseLost:
			// The bot gives up only when stuck mid-turn; otherwise it lost at a turn start.
			if r.PlayerID != g.Players[g.Turn.ActiveSeat].ID || (r.Reason != LossStuck && r.Reason != LossGaveUp) {
				t.Fatalf("seed %d: lost with %+v", seed, r)
			}
			// Score 0 only when the last card was an uncovered Skull.
			if r.Score == 0 && len(g.commands().skullPiles) == 0 {
				t.Fatalf("seed %d: lost with score 0 and no skull showing", seed)
			}
		}
	}
	t.Logf("won %d of %d random games", won, games)
}
