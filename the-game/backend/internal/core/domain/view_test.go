package domain

import (
	"errors"
	"math/rand/v2"
	"reflect"
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
	if len(v.Piles) != 4 || v.Piles[0] != (PileView{Index: 0, Direction: DirectionUp, StartValue: 1, TopValue: 1}) ||
		v.Piles[3] != (PileView{Index: 3, Direction: DirectionDown, StartValue: 100, TopValue: 100}) {
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
	if active.Piles[2] != (PileView{Index: 2, Direction: DirectionDown, StartValue: 100, TopValue: 45, CardCount: 2, TopCard: active.Piles[2].TopCard}) ||
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
			if !played {
				mustOK(t, g.EndTurn(active))
			}
		}
		check()
	}
}
