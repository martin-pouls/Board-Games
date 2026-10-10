package domain

import (
	"errors"
	"testing"
)

func TestBasePlacement(t *testing.T) {
	r := mustRules(t, DefaultOptions())
	up := func(played ...int) Pile { return Pile{Direction: DirectionUp, StartValue: 1, Cards: cards(played...)} }
	down := func(played ...int) Pile {
		return Pile{Direction: DirectionDown, StartValue: 100, Cards: cards(played...)}
	}
	tests := []struct {
		name string
		pile Pile
		card int
		want PlacementKind // 0 = rejected
	}{
		{"up start: lowest card", up(), 2, PlacementNormal},
		{"up start: highest card", up(), 99, PlacementNormal},
		{"up greater", up(30), 31, PlacementNormal},
		{"up much greater", up(30), 99, PlacementNormal},
		{"up equal", up(30), 30, 0},
		{"up less", up(30), 29, 0},
		{"up trick -10", up(30), 20, PlacementBackwardsTrick},
		{"up -9", up(30), 21, 0},
		{"up -11", up(30), 19, 0},
		{"up +10 is normal", up(30), 40, PlacementNormal},
		{"up trick uses top only", up(20, 30), 10, 0},
		{"down start: highest card", down(), 99, PlacementNormal},
		{"down start: lowest card", down(), 2, PlacementNormal},
		{"down smaller", down(70), 69, PlacementNormal},
		{"down equal", down(70), 70, 0},
		{"down greater", down(70), 71, 0},
		{"down trick +10", down(70), 80, PlacementBackwardsTrick},
		{"down +9", down(70), 79, 0},
		{"down +11", down(70), 81, 0},
		{"down -10 is normal", down(70), 60, PlacementNormal},
		// The trick needs a played card: it is impossible on the start cards 1 and 100.
		{"up start no trick", up(), 1 - 10, 0},
		{"down start no trick", down(), 100 + 10, 0},
		{"unknown direction", Pile{Direction: "sideways", StartValue: 50}, 60, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, ok := basePlacement(r, tt.pile, NumberCard(tt.card))
			if ok != (tt.want != 0) || kind != tt.want {
				t.Fatalf("basePlacement(%d) = %v, %v; want %v", tt.card, kind, ok, tt.want)
			}
		})
	}
}

func TestBackwardsTrickOnStartCardIsImpossible(t *testing.T) {
	// With an untouched pile, every card in the deck is a normal move; none uses the trick.
	r := mustRules(t, DefaultOptions())
	for _, p := range NewPiles(r) {
		for v := r.LowestCard; v <= r.HighestCard; v++ {
			kind, ok := basePlacement(r, p, NumberCard(v))
			if !ok || kind != PlacementNormal {
				t.Fatalf("%d on untouched %s pile: %v %v", v, p.Direction, kind, ok)
			}
		}
	}
}

func TestCheckPlacement(t *testing.T) {
	g := table{hands: [][]int{{10, 20}}, draw: []int{5}, piles: [][]int{{30}, nil, {70}, nil}}.build(t)
	tests := []struct {
		name string
		card int
		pile int
		kind PlacementKind
		want error
	}{
		{"normal", 31, 0, PlacementNormal, nil},
		{"trick", 20, 0, PlacementBackwardsTrick, nil},
		{"illegal", 29, 0, 0, ErrIllegalPlacement},
		{"down trick", 80, 2, PlacementBackwardsTrick, nil},
		{"negative pile", 31, -1, 0, ErrInvalidPile},
		{"pile out of range", 31, 4, 0, ErrInvalidPile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := g.Clone()
			p, err := g.checkPlacement(NumberCard(tt.card), tt.pile)
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if err == nil && (p.Kind != tt.kind || p.PileIndex != tt.pile || p.Card.Value != tt.card || p.Pile.TopValue() != g.Piles[tt.pile].TopValue()) {
				t.Fatalf("placement = %+v", p)
			}
			if g.Revision != before.Revision || len(g.Piles[0].Cards) != 1 {
				t.Fatal("checkPlacement mutated the game")
			}
		})
	}
	// Classic cards carry no commands, so no constraint ever vetoes a base-legal placement.
	for _, con := range g.placementConstraints() {
		for _, p := range []Placement{
			{Card: NumberCard(31), PileIndex: 0, Pile: g.Piles[0], Kind: PlacementNormal},
			{Card: NumberCard(20), PileIndex: 0, Pile: g.Piles[0], Kind: PlacementBackwardsTrick},
		} {
			if err := con.Check(g, p); err != nil {
				t.Fatalf("%T vetoed a classic placement: %v", con, err)
			}
		}
	}
}
