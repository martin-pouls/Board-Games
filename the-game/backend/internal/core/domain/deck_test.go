package domain

import (
	"errors"
	"reflect"
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

func TestNewDeckExtreme(t *testing.T) {
	deck, err := NewDeck(mustRules(t, extreme))
	if err != nil {
		t.Fatal(err)
	}
	if len(deck) != 98 {
		t.Fatalf("len(deck) = %d, want 98", len(deck))
	}
	// rules.md §2 "Command → card values", copied independently of deck.go.
	want := map[Command][]int{
		CommandStop:             {6, 35, 57, 95},
		CommandSkull:            {17, 41, 61, 87},
		CommandExactlyThree:     {19, 28, 53, 81},
		CommandNoCommunication:  {26, 65, 78, 92},
		CommandNoBackwardsTrick: {21, 32, 59, 70},
		CommandSinglePile:       {9, 47, 68, 84},
		CommandDrawOnlyOne:      {13, 38, 43, 76},
	}
	got := map[Command][]int{}
	for i, c := range deck {
		if c.Value != i+2 {
			t.Fatalf("deck[%d].Value = %d, want %d (2..99 ascending, each once)", i, c.Value, i+2)
		}
		if c.Kind != CardKindNumber {
			t.Errorf("card %+v: kind %q, want number", c, c.Kind)
		}
		if c.Command != CommandNone {
			got[c.Command] = append(got[c.Command], c.Value)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command values\n got %v\nwant %v", got, want)
	}
	total := 0
	for _, vs := range got {
		total += len(vs)
	}
	if total != 28 || len(got) != 7 {
		t.Fatalf("%d command cards over %d commands, want 28 over 7", total, len(got))
	}
	// The extreme deck is a fresh slice: building it again yields the same deck.
	again, _ := NewDeck(mustRules(t, extremePro))
	if !reflect.DeepEqual(deck, again) {
		t.Fatal("extreme deck differs between builds/options")
	}
}

func TestNewDeckErrors(t *testing.T) {
	tests := []struct {
		name string
		r    Ruleset
	}{
		{"unknown variant", Ruleset{Variant: "bogus", LowestCard: 2, HighestCard: 99}},
		{"empty variant", Ruleset{LowestCard: 2, HighestCard: 99}},
		// Defensive: a command value outside [LowestCard, HighestCard].
		{"extreme, command value outside the deck", Ruleset{Variant: VariantExtreme, LowestCard: 2, HighestCard: 50}},
		{"extreme, empty range", Ruleset{Variant: VariantExtreme}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deck, err := NewDeck(tt.r)
			if !errors.Is(err, ErrInvalidOptions) || deck != nil {
				t.Fatalf("NewDeck = %v, %v; want ErrInvalidOptions", deck, err)
			}
		})
	}
}

func TestCommandClass(t *testing.T) {
	tests := []struct {
		cmd       Command
		class     CommandClass
		lightning bool
		infinity  bool
	}{
		{CommandNone, CommandClassNone, false, false},
		{"bogus", CommandClassNone, false, false},
		{CommandStop, CommandClassLightning, true, false},
		{CommandSkull, CommandClassLightning, true, false},
		{CommandExactlyThree, CommandClassLightning, true, false},
		{CommandNoCommunication, CommandClassInfinity, false, true},
		{CommandNoBackwardsTrick, CommandClassInfinity, false, true},
		{CommandSinglePile, CommandClassInfinity, false, true},
		{CommandDrawOnlyOne, CommandClassInfinity, false, true},
	}
	for _, tt := range tests {
		if got := tt.cmd.Class(); got != tt.class {
			t.Errorf("%q.Class() = %q, want %q", tt.cmd, got, tt.class)
		}
		if tt.cmd.IsLightning() != tt.lightning || tt.cmd.IsInfinity() != tt.infinity {
			t.Errorf("%q: IsLightning %v IsInfinity %v", tt.cmd, tt.cmd.IsLightning(), tt.cmd.IsInfinity())
		}
	}
	// Wire strings are camelCase and fixed (plan §1).
	wire := map[Command]string{
		CommandStop: "stop", CommandSkull: "skull", CommandExactlyThree: "exactlyThree",
		CommandNoCommunication: "noCommunication", CommandNoBackwardsTrick: "noBackwardsTrick",
		CommandSinglePile: "singlePile", CommandDrawOnlyOne: "drawOnlyOne",
	}
	for c, s := range wire {
		if string(c) != s {
			t.Errorf("command %q, want wire string %q", c, s)
		}
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
