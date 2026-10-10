package domain

import (
	"reflect"
	"testing"
	"time"
)

func sampleGame(t *testing.T) *Game {
	t.Helper()
	r, err := NewRuleset(Options{Variant: VariantClassic, Pro: true})
	if err != nil {
		t.Fatal(err)
	}
	piles := NewPiles(r)
	piles[0].Cards = append(piles[0].Cards, NumberCard(10))
	return &Game{
		ID:       "ABCDE",
		Options:  Options{Variant: VariantClassic, Pro: true},
		Rules:    r,
		Phase:    PhasePlaying,
		HostID:   "p1",
		Players:  []Player{{ID: "p1", Name: "Anna", Hand: []Card{NumberCard(3), NumberCard(50)}}, {ID: "p2", Name: "Ben", Hand: []Card{}}},
		DrawPile: []Card{NumberCard(70), NumberCard(80), NumberCard(90)},
		Piles:    piles,
		Turn:     Turn{Number: 2, ActiveSeat: 0, Plays: []Play{{Card: NumberCard(10), Pile: 0}}, StartHandSize: 3},
		Result:   &Result{Outcome: OutcomeLost, Score: 5},
		Revision: 7,
		events:   []Event{CardPlayed{PlayerID: "p1", Card: NumberCard(10), Pile: 0}},
	}
}

func TestCloneDeepEqualAndIndependent(t *testing.T) {
	g := sampleGame(t)
	c := g.Clone()
	if !reflect.DeepEqual(g, c) {
		t.Fatalf("clone differs:\n%+v\n%+v", g, c)
	}
	before := sampleGame(t)

	c.Rules.HandSizes[1] = 99
	c.Players[0].Hand[0] = NumberCard(4)
	c.Players[0].Name = "X"
	c.Players = append(c.Players, Player{ID: "p3"})
	c.DrawPile[0] = NumberCard(71)
	c.Piles[0].Cards[0] = NumberCard(11)
	c.Piles[1].Direction = DirectionDown
	c.Turn.Plays[0].Pile = 3
	c.Result.Score = 42
	c.events[0] = GameStarted{}
	c.Revision++

	if !reflect.DeepEqual(g, before) {
		t.Fatalf("mutating clone changed original:\n%+v\n%+v", g, before)
	}
}

func TestClonePreservesNilAndEmpty(t *testing.T) {
	empty := &Game{
		Players:  []Player{{ID: "p1", Hand: []Card{}}},
		DrawPile: []Card{},
		Piles:    []Pile{{Cards: []Card{}}},
		Turn:     Turn{Plays: []Play{}},
		events:   []Event{},
	}
	c := empty.Clone()
	if c.Players[0].Hand == nil || c.DrawPile == nil || c.Piles[0].Cards == nil || c.Turn.Plays == nil || c.events == nil {
		t.Fatalf("empty slices became nil: %+v", c)
	}

	nilGame := &Game{Players: []Player{{ID: "p1"}}, Piles: []Pile{{}}}
	c = nilGame.Clone()
	if c.Players[0].Hand != nil || c.DrawPile != nil || c.Piles[0].Cards != nil || c.Turn.Plays != nil ||
		c.events != nil || c.Rules.HandSizes != nil || c.Result != nil {
		t.Fatalf("nil fields became non-nil: %+v", c)
	}
	if (&Game{}).Clone().Players != nil || (&Game{}).Clone().Piles != nil {
		t.Fatal("nil Players/Piles became non-nil")
	}
	if !reflect.DeepEqual(empty.Clone(), empty) || !reflect.DeepEqual(nilGame.Clone(), nilGame) {
		t.Fatal("clone not DeepEqual")
	}
}

func TestPullEvents(t *testing.T) {
	g := &Game{}
	g.emit(GameStarted{})
	g.emit(TurnStarted{PlayerID: "p1", Number: 1})
	evs := g.PullEvents()
	if len(evs) != 2 || evs[0] != (GameStarted{}) || evs[1] != (TurnStarted{PlayerID: "p1", Number: 1}) {
		t.Fatalf("PullEvents = %#v", evs)
	}
	if again := g.PullEvents(); len(again) != 0 {
		t.Fatalf("second PullEvents = %#v", again)
	}
}

func TestCommit(t *testing.T) {
	g := &Game{Revision: 1}
	g.commit()
	if g.Revision != 2 {
		t.Fatalf("Revision = %d", g.Revision)
	}
}

func TestScore(t *testing.T) {
	g := sampleGame(t)
	h, d := g.Score()
	if h != 2 || d != 3 {
		t.Fatalf("Score() = %d, %d", h, d)
	}
}

func TestPlayerByIDAndIsHost(t *testing.T) {
	g := sampleGame(t)
	p, seat, ok := g.PlayerByID("p2")
	if !ok || seat != 1 || p.Name != "Ben" {
		t.Fatalf("PlayerByID(p2) = %+v, %d, %v", p, seat, ok)
	}
	if _, seat, ok := g.PlayerByID("nope"); ok || seat != -1 {
		t.Fatalf("PlayerByID(nope) found: %d", seat)
	}
	if !g.IsHost("p1") || g.IsHost("p2") || g.IsHost("") {
		t.Fatal("IsHost wrong")
	}
	if (&Game{}).IsHost("") {
		t.Fatal("empty id must not be host")
	}
	if g.CommunicationRestricted() {
		t.Fatal("CommunicationRestricted must be false in Phase 1")
	}
}

func TestPilesTouched(t *testing.T) {
	tests := []struct {
		plays []Play
		want  []int
	}{
		{nil, nil},
		{[]Play{{Pile: 2}}, []int{2}},
		{[]Play{{Pile: 2}, {Pile: 0}, {Pile: 2}, {Pile: 3}, {Pile: 0}}, []int{2, 0, 3}},
	}
	for _, tt := range tests {
		if got := (Turn{Plays: tt.plays}).PilesTouched(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("PilesTouched(%v) = %v, want %v", tt.plays, got, tt.want)
		}
	}
}

func TestSummary(t *testing.T) {
	r, _ := NewRuleset(DefaultOptions())
	players := func(n int) []Player {
		ps := make([]Player, n)
		for i := range ps {
			ps[i] = Player{ID: PlayerID(rune('a' + i)), Name: string(rune('A' + i))}
		}
		return ps
	}
	tests := []struct {
		name     string
		phase    Phase
		n        int
		joinable bool
	}{
		{"lobby open", PhaseLobby, 2, true},
		{"lobby full", PhaseLobby, 5, false},
		{"started", PhasePlaying, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)
			g := &Game{ID: "K7MQ2", Phase: tt.phase, HostID: "a", Rules: r, Players: players(tt.n), CreatedAt: created}
			want := GameSummary{GameID: "K7MQ2", Phase: tt.phase, HostName: "A", PlayerCount: tt.n, MaxPlayers: 5, Joinable: tt.joinable, CreatedAt: created}
			if got := g.Summary(); got != want {
				t.Fatalf("Summary() = %+v, want %+v", got, want)
			}
		})
	}
}

func TestSummaryOutcome(t *testing.T) {
	r, _ := NewRuleset(DefaultOptions())
	for _, o := range []Outcome{OutcomeWon, OutcomeLost, OutcomeAborted} {
		g := &Game{ID: "K7MQ2", Phase: PhaseFinished, Rules: r, Result: &Result{Outcome: o}}
		if got := g.Summary().Outcome; got != o {
			t.Errorf("Summary().Outcome = %q, want %q", got, o)
		}
	}
	if got := (&Game{Phase: PhasePlaying, Rules: r}).Summary().Outcome; got != "" {
		t.Errorf("running game Outcome = %q, want empty", got)
	}
}
