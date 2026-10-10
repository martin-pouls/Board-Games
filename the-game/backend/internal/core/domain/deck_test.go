package domain

import (
	"errors"
	"testing"
)

func TestNewDeckClassic(t *testing.T) {
	r, err := NewRuleset(DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	deck, err := NewDeck(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(deck) != 98 {
		t.Fatalf("len(deck) = %d, want 98", len(deck))
	}
	seen := map[int]bool{}
	for i, c := range deck {
		if c.Value < 2 || c.Value > 99 {
			t.Errorf("card %d has value %d out of range", i, c.Value)
		}
		if seen[c.Value] {
			t.Errorf("duplicate value %d", c.Value)
		}
		seen[c.Value] = true
		if c.Kind != CardKindNumber || c.Command != CommandNone {
			t.Errorf("card %+v: want plain number card", c)
		}
		if i > 0 && deck[i-1].Value >= c.Value {
			t.Errorf("deck not ascending at %d", i)
		}
	}
}

func TestNewDeckExtremeRejected(t *testing.T) {
	_, err := NewDeck(Ruleset{Variant: VariantExtreme})
	if !errors.Is(err, ErrInvalidOptions) {
		t.Fatalf("NewDeck(extreme) err = %v, want ErrInvalidOptions", err)
	}
}

func TestNumberCard(t *testing.T) {
	if got := NumberCard(42); got != (Card{Value: 42, Kind: CardKindNumber}) {
		t.Fatalf("NumberCard(42) = %+v", got)
	}
}

func TestNewPiles(t *testing.T) {
	r, _ := NewRuleset(DefaultOptions())
	piles := NewPiles(r)
	want := []struct {
		dir   Direction
		start int
	}{{DirectionUp, 1}, {DirectionUp, 1}, {DirectionDown, 100}, {DirectionDown, 100}}
	if len(piles) != len(want) {
		t.Fatalf("len(piles) = %d", len(piles))
	}
	for i, w := range want {
		p := piles[i]
		if p.Direction != w.dir || p.StartValue != w.start || p.Cards == nil || len(p.Cards) != 0 {
			t.Errorf("pile %d = %+v", i, p)
		}
		if p.TopValue() != w.start {
			t.Errorf("pile %d TopValue = %d, want %d", i, p.TopValue(), w.start)
		}
		if _, ok := p.TopCard(); ok {
			t.Errorf("pile %d TopCard ok on empty pile", i)
		}
	}
}

func TestPileTop(t *testing.T) {
	p := Pile{Direction: DirectionDown, StartValue: 100, Cards: []Card{NumberCard(90), NumberCard(55)}}
	if p.TopValue() != 55 {
		t.Fatalf("TopValue = %d", p.TopValue())
	}
	c, ok := p.TopCard()
	if !ok || c != NumberCard(55) {
		t.Fatalf("TopCard = %+v, %v", c, ok)
	}
}
