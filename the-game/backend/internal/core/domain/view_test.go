package domain

import (
	"errors"
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"
)

func TestViewForUnknownPlayer(t *testing.T) {
	g := lobbyWith(t, 2, DefaultOptions())
	if _, err := g.ViewFor("zz"); !errors.Is(err, ErrUnknownPlayer) {
		t.Fatalf("err = %v", err)
	}
}

func TestViewForLobby(t *testing.T) {
	g := lobbyWith(t, 2, Options{Variant: VariantClassic, Pro: true, ReducedHandSize: true})
	v, err := g.ViewFor(pid(1))
	mustOK(t, err)
	if v.GameID != "LOBBY" || v.Revision != 2 || v.Phase != PhaseLobby || v.You != pid(1) || v.HostID != pid(0) {
		t.Fatalf("view header %+v", v)
	}
	if v.Rules != (RulesView{MinCardsPerTurn: 3, MinCardsDrawPileEmpty: 1, HandSize: 6, MinPlayers: 1, MaxPlayers: 5}) {
		t.Fatalf("rules %+v", v.Rules)
	}
	if v.Hand == nil || len(v.Hand) != 0 || v.Turn != nil || v.Result != nil || v.DrawPileCount != 0 || v.NextGameID != "" {
		t.Fatalf("lobby view %+v", v)
	}
	want := []PlayerSummary{
		{ID: pid(0), Name: "Player 0", Seat: 0, IsHost: true},
		{ID: pid(1), Name: "Player 1", Seat: 1, IsYou: true},
	}
	if !reflect.DeepEqual(v.Players, want) {
		t.Fatalf("players %+v", v.Players)
	}
	if len(v.Piles) != 4 || !reflect.DeepEqual(v.Piles[0], PileView{Index: 0, Direction: DirectionUp, StartValue: 1, TopValue: 1}) ||
		!reflect.DeepEqual(v.Piles[3], PileView{Index: 3, Direction: DirectionDown, StartValue: 100, TopValue: 100}) {
		t.Fatalf("piles %+v", v.Piles)
	}
}

func TestViewForChoosingFirstPlayer(t *testing.T) {
	g := lobbyWith(t, 2, DefaultOptions())
	mustOK(t, g.Start(pid(0), noShuffle{}))
	v, err := g.ViewFor(pid(1))
	mustOK(t, err)
	if v.Turn != nil || len(v.Hand) != 7 || v.DrawPileCount != 84 {
		t.Fatalf("view %+v", v)
	}
	for _, h := range v.Hand {
		if len(h.LegalPiles) != 0 {
			t.Fatalf("legal piles before the first turn: %+v", h)
		}
	}
	for _, p := range v.Players {
		if p.HandCount != 7 || p.IsActive || p.RevealedHand != nil {
			t.Fatalf("player %+v", p)
		}
	}
}

func TestViewForPlaying(t *testing.T) {
	g := table{
		hands: [][]int{{10, 20, 95}, {30, 40}, {50}},
		draw:  []int{80, 81},
		piles: [][]int{{60}, nil, {70, 45}, nil},
		plays: []Play{{NumberCard(45), 2}},
	}.build(t)
	g.Revision = 7

	active, err := g.ViewFor(pid(0))
	mustOK(t, err)
	wantHand := []HandCard{
		{Card: NumberCard(10), LegalPiles: []int{1, 2, 3}},
		{Card: NumberCard(20), LegalPiles: []int{1, 2, 3}},
		{Card: NumberCard(95), LegalPiles: []int{0, 1, 3}},
	}
	if !reflect.DeepEqual(active.Hand, wantHand) {
		t.Fatalf("active hand %+v", active.Hand)
	}
	wantTurn := &TurnView{Number: 1, ActivePlayerID: pid(0), CardsPlayed: 1, RequiredMinimum: 2, CanEndTurn: false, Plays: []Play{{NumberCard(45), 2}}}
	if !reflect.DeepEqual(active.Turn, wantTurn) {
		t.Fatalf("turn %+v", active.Turn)
	}
	top := NumberCard(45)
	if !reflect.DeepEqual(active.Piles[2], PileView{Index: 2, Direction: DirectionDown, StartValue: 100, TopValue: 45, CardCount: 2, TopCard: active.Piles[2].TopCard, Cards: g.Piles[2].Cards}) ||
		active.Piles[2].TopCard == nil || *active.Piles[2].TopCard != top || active.Piles[1].TopCard != nil {
		t.Fatalf("piles %+v", active.Piles)
	}
	if active.DrawPileCount != 2 || active.Revision != 7 || active.Result != nil || active.CommunicationRestricted {
		t.Fatalf("view %+v", active)
	}

	other, err := g.ViewFor(pid(1))
	mustOK(t, err)
	// Only the viewer's own hand, never with legal piles when not active.
	if len(other.Hand) != 2 || other.Hand[0].Card.Value != 30 || other.Hand[1].Card.Value != 40 {
		t.Fatalf("other hand %+v", other.Hand)
	}
	for _, h := range other.Hand {
		if len(h.LegalPiles) != 0 {
			t.Fatalf("legal piles for a non-active viewer: %+v", h)
		}
	}
	wantPlayers := []PlayerSummary{
		{ID: pid(0), Name: "Player 0", Seat: 0, HandCount: 3, IsHost: true, IsActive: true},
		{ID: pid(1), Name: "Player 1", Seat: 1, HandCount: 2, IsYou: true},
		{ID: pid(2), Name: "Player 2", Seat: 2, HandCount: 1},
	}
	if !reflect.DeepEqual(other.Players, wantPlayers) {
		t.Fatalf("players %+v", other.Players)
	}
	if !reflect.DeepEqual(other.Turn, wantTurn) {
		t.Fatalf("turn %+v", other.Turn)
	}

	// The view does not alias game state.
	other.Turn.Plays[0].Pile = 3
	*active.Piles[2].TopCard = NumberCard(1)
	if g.Turn.Plays[0].Pile != 2 || g.Piles[2].Cards[1].Value != 45 {
		t.Fatal("view aliases game state")
	}

	// After the second card the turn may end.
	mustOK(t, g.PlayCard(pid(0), 95, 0))
	v, _ := g.ViewFor(pid(2))
	if !v.Turn.CanEndTurn || v.Turn.CardsPlayed != 2 || v.Revision != 8 {
		t.Fatalf("turn %+v rev %d", v.Turn, v.Revision)
	}
}

func TestViewForFinished(t *testing.T) {
	g := finishedGame(t) // p0 [50], p1 [51 52], p2 [], draw [53]
	for seat := range g.Players {
		v, err := g.ViewFor(pid(seat))
		mustOK(t, err)
		want := [][]Card{cards(50), cards(51, 52), {}}
		for i, p := range v.Players {
			if p.RevealedHand == nil || !reflect.DeepEqual(p.RevealedHand, want[i]) {
				t.Fatalf("revealed hand of seat %d = %v, want %v", i, p.RevealedHand, want[i])
			}
			if p.IsActive {
				t.Fatalf("active player in finished game: %+v", p)
			}
		}
		wantRes := &ResultView{Outcome: OutcomeLost, Score: 4, Excellent: true, CardsLeftInHands: 3, CardsLeftInDrawPile: 1, StuckPlayerID: pid(0)}
		if !reflect.DeepEqual(v.Result, wantRes) {
			t.Fatalf("result %+v", v.Result)
		}
		if v.Turn == nil || v.Turn.ActivePlayerID != pid(0) || v.Turn.CanEndTurn {
			t.Fatalf("turn %+v", v.Turn)
		}
		for _, h := range v.Hand {
			if len(h.LegalPiles) != 0 {
				t.Fatalf("legal piles after the game: %+v", h)
			}
		}
	}
	// Revealed hands are copies.
	v, _ := g.ViewFor(pid(1))
	v.Players[1].RevealedHand[0] = NumberCard(2)
	if g.Players[1].Hand[0].Value != 51 {
		t.Fatal("revealed hand aliases game state")
	}
	// NextGameID is shown once a rematch exists.
	_, err := g.Rematch(pid(2), "NEXT2")
	mustOK(t, err)
	if v, _ := g.ViewFor(pid(0)); v.NextGameID != "NEXT2" {
		t.Fatalf("NextGameID = %q", v.NextGameID)
	}
}

func TestViewForExcellent(t *testing.T) {
	tests := []struct {
		score     int
		excellent bool
	}{{0, true}, {9, true}, {10, false}, {40, false}}
	for _, tt := range tests {
		g := finishedGame(t)
		g.Result.Score = tt.score
		v, _ := g.ViewFor(pid(0))
		if v.Result.Excellent != tt.excellent {
			t.Errorf("score %d: excellent = %v", tt.score, v.Result.Excellent)
		}
	}
}

// TestViewNeverRevealsHandsBeforeFinished walks random games and checks every view.
func TestViewNeverRevealsHandsBeforeFinished(t *testing.T) {
	for seed := range uint64(20) {
		rng := rand.New(rand.NewPCG(seed, 5))
		n := 2 + int(seed%4)
		g := lobbyWith(t, n, DefaultOptions())
		check := func() {
			t.Helper()
			for seat, p := range g.Players {
				v, err := g.ViewFor(p.ID)
				mustOK(t, err)
				if v.Revision != g.Revision {
					t.Fatalf("revision %d != %d", v.Revision, g.Revision)
				}
				if len(v.Hand) != len(p.Hand) {
					t.Fatalf("hand size %d != %d", len(v.Hand), len(p.Hand))
				}
				isActive := g.Phase == PhasePlaying && seat == g.Turn.ActiveSeat
				for _, h := range v.Hand {
					if !isActive && len(h.LegalPiles) != 0 {
						t.Fatalf("seed %d: legal piles for non-active viewer", seed)
					}
				}
				for _, s := range v.Players {
					if (g.Phase == PhaseFinished) != (s.RevealedHand != nil) {
						t.Fatalf("seed %d phase %s: revealed hand %v", seed, g.Phase, s.RevealedHand)
					}
					if s.HandCount != len(g.Players[s.Seat].Hand) {
						t.Fatalf("hand count mismatch")
					}
				}
			}
		}
		check()
		mustOK(t, g.Start(pid(0), rng))
		check()
		mustOK(t, g.ChooseFirstPlayer(pid(0), pid(1)))
		for g.Phase == PhasePlaying {
			check()
			active := g.Players[g.Turn.ActiveSeat].ID
			v, _ := g.ViewFor(active)
			if v.Turn.CanEndTurn && rng.IntN(2) == 0 {
				mustOK(t, g.EndTurn(active))
				continue
			}
			played := false
			for _, h := range v.Hand {
				if len(h.LegalPiles) > 0 {
					mustOK(t, g.PlayCard(active, h.Card.Value, h.LegalPiles[0]))
					played = true
					break
				}
			}
			switch {
			case played:
			case v.Turn.Stuck:
				mustOK(t, g.GiveUp(active))
			default:
				mustOK(t, g.EndTurn(active))
			}
		}
		check()
	}
}

func TestHandCardBlocked(t *testing.T) {
	tests := []struct {
		name    string
		tb      table
		viewer  PlayerID
		card    int
		legal   []int
		blocked error // nil = no reason
	}{
		{"3! with start hand 2", table{opts: extreme, hands: [][]int{{28, 99}, {50}}, draw: []int{90}}, pid(0), 28, nil, ErrExactCountUnreachable},
		{"base-illegal card", table{opts: extreme, hands: [][]int{{50, 99}, {98}}, draw: []int{90}, piles: [][]int{{96}, {97}, {5}, {4}}}, pid(0), 50, nil, nil},
		{"legal somewhere, vetoed elsewhere", table{opts: extreme, hands: [][]int{{80, 99}, {98}}, draw: []int{90}, piles: [][]int{nil, nil, {70}}}, pid(0), 80, []int{0, 1, 3}, nil},
		{"only the backwards trick fits", table{opts: extreme, hands: [][]int{{30, 99}, {98}}, draw: []int{90}, piles: [][]int{{40}, {32}, {5}, {4}}}, pid(0), 30, nil, ErrBackwardsTrickBlocked},
		{"single pile", table{opts: extreme, hands: [][]int{{15, 99}, {98}}, draw: []int{90}, piles: [][]int{{20}, {9}, {5}, {4}}, plays: []Play{play(20, 0)}}, pid(0), 15, nil, ErrSinglePileViolation},
		{"4th card after 3!", table{opts: extreme, hands: [][]int{{30}, {98}}, draw: []int{90}, piles: [][]int{{19, 20, 22}}, plays: []Play{play(19, 0), play(20, 0), play(22, 0)}}, pid(0), 30, nil, ErrExactCountExceeded},
		{"STOP too early", table{opts: extreme, hands: [][]int{{95, 99}, {98}}, draw: []int{90}, piles: [][]int{{19}}, plays: []Play{play(19, 0)}}, pid(0), 95, nil, ErrStopTooEarly},
		{"non-active viewer", table{opts: extreme, hands: [][]int{{50}, {28}}, draw: []int{90}}, pid(1), 28, nil, nil},
		{"classic", table{hands: [][]int{{28, 99}, {50}}, draw: []int{90}}, pid(0), 28, []int{0, 1, 2, 3}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.tb.build(t)
			v, err := g.ViewFor(tt.viewer)
			mustOK(t, err)
			i := slices.IndexFunc(v.Hand, func(h HandCard) bool { return h.Card.Value == tt.card })
			if i < 0 {
				t.Fatalf("card %d not in hand %+v", tt.card, v.Hand)
			}
			h := v.Hand[i]
			if !reflect.DeepEqual(h.LegalPiles, tt.legal) {
				t.Fatalf("legal piles %v, want %v", h.LegalPiles, tt.legal)
			}
			// Every veto is exactly what PlayCard returns on that pile.
			for _, veto := range h.Vetoes {
				err := g.Clone().PlayCard(tt.viewer, tt.card, veto.Pile)
				if err == nil || err.Error() != veto.Err.Error() {
					t.Fatalf("PlayCard on pile %d = %v, veto = %v", veto.Pile, err, veto.Err)
				}
			}
			if tt.blocked == nil {
				if h.Blocked != nil {
					t.Fatalf("blocked = %v, want nil", h.Blocked)
				}
				return
			}
			if !errors.Is(h.Blocked, tt.blocked) {
				t.Fatalf("blocked = %v, want %v", h.Blocked, tt.blocked)
			}
			// The reason is exactly what PlayCard returns for the first base-legal pile.
			for pile := range g.Piles {
				if _, ok := basePlacement(g.Rules, g.Piles[pile], h.Card); ok {
					err := g.Clone().PlayCard(tt.viewer, tt.card, pile)
					if err == nil || err.Error() != h.Blocked.Error() {
						t.Fatalf("PlayCard on pile %d = %v, blocked = %v", pile, err, h.Blocked)
					}
					break
				}
			}
		})
	}
}

func TestHandCardVetoes(t *testing.T) {
	// A skull from this turn on pile 0: STOP may only cover it; elsewhere it is vetoed.
	g := table{opts: extreme, hands: [][]int{{3, 95}, {50}}, draw: []int{90}, piles: [][]int{{41}}, plays: []Play{play(41, 0)}}.build(t)
	v, err := g.ViewFor(pid(0))
	mustOK(t, err)
	stop := v.Hand[1]
	if stop.Card.Value != 95 || !reflect.DeepEqual(stop.LegalPiles, []int{0}) || stop.Blocked != nil {
		t.Fatalf("STOP %+v", stop)
	}
	var piles []int
	for _, veto := range stop.Vetoes {
		if !errors.Is(veto.Err, ErrSkullUncovered) {
			t.Fatalf("veto %+v", veto)
		}
		piles = append(piles, veto.Pile)
	}
	if !reflect.DeepEqual(piles, []int{1, 2, 3}) {
		t.Fatalf("vetoed piles %v", piles)
	}
	// 3 is base-illegal on pile 0 (not a veto) and legal on the descending piles.
	if low := v.Hand[0]; low.Vetoes != nil || !reflect.DeepEqual(low.LegalPiles, []int{1, 2, 3}) {
		t.Fatalf("3: %+v", low)
	}
	// Other players get no vetoes.
	if other, _ := g.ViewFor(pid(1)); other.Hand[0].Vetoes != nil {
		t.Fatalf("other player's vetoes %+v", other.Hand[0].Vetoes)
	}
}

func TestViewForExtreme(t *testing.T) {
	// Pile 0: a skull from an earlier turn. Pile 1: 3! and a skull played this turn.
	// Pile 2: No backwards trick. Pile 3: Single pile (visible since the turn started).
	g := table{
		opts:  extreme,
		hands: [][]int{{18, 45, 50}, {99}},
		draw:  []int{90},
		piles: [][]int{{17}, {10, 28, 41}, {70}, {84}},
		plays: []Play{play(28, 1), play(41, 1)},
	}.build(t)

	v, err := g.ViewFor(pid(0))
	mustOK(t, err)
	wantActive := []ActiveCommand{{CommandNoBackwardsTrick, 2}, {CommandSinglePile, 3}}
	if !reflect.DeepEqual(v.ActiveCommands, wantActive) || v.CommunicationRestricted {
		t.Fatalf("active %+v restricted %v", v.ActiveCommands, v.CommunicationRestricted)
	}
	for i, want := range []bool{false, true, true, true} {
		if v.Piles[i].CommandActive != want {
			t.Errorf("pile %d CommandActive = %v, want %v", i, v.Piles[i].CommandActive, want)
		}
	}
	if top := v.Piles[1].TopCard; top == nil || *top != xcard(41) || top.Command != CommandSkull {
		t.Fatalf("pile 1 top %+v", top)
	}
	one := 1
	wantCons := TurnConstraints{ExactCount: 3, SinglePile: true, BoundPile: &one, SkullPiles: []int{1}}
	if !reflect.DeepEqual(v.Turn.Constraints, wantCons) || v.Turn.CanEndTurn || v.Turn.CardsPlayed != 2 {
		t.Fatalf("turn %+v", v.Turn)
	}
	h18, h45 := v.Hand[0], v.Hand[1]
	if h18.LegalPiles != nil || !errors.Is(h18.Blocked, ErrSinglePileViolation) {
		t.Fatalf("hand card 18: %+v", h18)
	}
	if !reflect.DeepEqual(h45.LegalPiles, []int{1}) || h45.Blocked != nil {
		t.Fatalf("hand card 45: %+v", h45)
	}

	// Other players see the same public command state, but no legal piles or reasons.
	other, err := g.ViewFor(pid(1))
	mustOK(t, err)
	if !reflect.DeepEqual(other.ActiveCommands, wantActive) || !reflect.DeepEqual(other.Turn.Constraints, wantCons) ||
		other.Hand[0].LegalPiles != nil || other.Hand[0].Blocked != nil {
		t.Fatalf("other view %+v", other)
	}

	// The view does not alias game state.
	v.Turn.Constraints.SkullPiles[0] = 3
	*v.Turn.Constraints.BoundPile = 3
	v.ActiveCommands[0].Pile = 0
	again, _ := g.ViewFor(pid(0))
	if !reflect.DeepEqual(again.ActiveCommands, wantActive) || !reflect.DeepEqual(again.Turn.Constraints, wantCons) {
		t.Fatalf("view aliases game state: %+v %+v", again.ActiveCommands, again.Turn.Constraints)
	}

	// Covering the skull with the 3rd card: the turn may end, no further card fits.
	mustOK(t, g.PlayCard(pid(0), 45, 1))
	v, _ = g.ViewFor(pid(0))
	wantCons = TurnConstraints{ExactCount: 3, SinglePile: true, BoundPile: &one}
	if !reflect.DeepEqual(v.Turn.Constraints, wantCons) || !v.Turn.CanEndTurn || v.Piles[1].CommandActive {
		t.Fatalf("after covering: turn %+v pile %+v", v.Turn, v.Piles[1])
	}
	if h := v.Hand[1]; h.Card.Value != 50 || h.LegalPiles != nil || !errors.Is(h.Blocked, ErrExactCountExceeded) {
		t.Fatalf("hand card 50: %+v", h)
	}

	// After the turn the next player sees the earlier lightning tops as inactive.
	mustOK(t, g.EndTurn(pid(0)))
	v, _ = g.ViewFor(pid(1))
	if !reflect.DeepEqual(v.Turn.Constraints, TurnConstraints{SinglePile: true}) {
		t.Fatalf("next turn constraints %+v", v.Turn.Constraints)
	}
	for i, want := range []bool{false, false, true, true} {
		if v.Piles[i].CommandActive != want {
			t.Errorf("next turn: pile %d CommandActive = %v, want %v", i, v.Piles[i].CommandActive, want)
		}
	}
}

func TestViewForExtremeFinished(t *testing.T) {
	t.Run("won with STOP", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{95}, {}}}.build(t)
		mustOK(t, g.PlayCard(pid(0), 95, 0))
		v, _ := g.ViewFor(pid(1))
		if v.Result.Outcome != OutcomeWon || !reflect.DeepEqual(v.Turn.Constraints, TurnConstraints{StopPlayed: true}) ||
			!v.Piles[0].CommandActive || v.CommunicationRestricted {
			t.Fatalf("view %+v turn %+v", v, v.Turn)
		}
	})
	t.Run("lost with an uncovered skull", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{41}}, piles: [][]int{nil, nil, nil, {92}}}.build(t)
		mustOK(t, g.PlayCard(pid(0), 41, 3))
		if v, _ := g.ViewFor(pid(0)); v.Phase != PhasePlaying || !v.Turn.Stuck || v.Turn.CanEndTurn {
			t.Fatalf("before giving up: phase %s turn %+v", v.Phase, v.Turn)
		}
		mustOK(t, g.GiveUp(pid(0)))
		v, _ := g.ViewFor(pid(0))
		if v.Result.Outcome != OutcomeLost || v.Result.Score != 0 || v.Result.StuckPlayerID != pid(0) ||
			!reflect.DeepEqual(v.Turn.Constraints, TurnConstraints{SkullPiles: []int{3}}) || v.CommunicationRestricted {
			t.Fatalf("view %+v turn %+v", v.Result, v.Turn)
		}
	})
	t.Run("single pile active but unbound", func(t *testing.T) {
		g := table{opts: extreme, hands: [][]int{{50}}, draw: []int{90}, piles: [][]int{nil, {9}}}.build(t)
		v, _ := g.ViewFor(pid(0))
		if !reflect.DeepEqual(v.Turn.Constraints, TurnConstraints{SinglePile: true}) {
			t.Fatalf("constraints %+v", v.Turn.Constraints)
		}
	})
}
