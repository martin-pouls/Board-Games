package domain

import (
	"slices"
	"testing"
)

func TestMovementAndAbility(t *testing.T) {
	tests := []struct {
		rank     Rank
		movement int
		ability  Ability
	}{
		{1, 0, AbilityMusician},
		{2, 3, AbilityNone},
		{3, 2, AbilityFoxes},
		{4, 1, AbilityNone},
		{5, 1, AbilityGazelle},
		{6, 2, AbilityNone},
		{7, 0, AbilityGift},
		{8, 2, AbilityNone},
		{9, 0, AbilityRoyalHeir},
		{10, 3, AbilityNone},
	}
	for _, tt := range tests {
		for _, s := range Suits {
			card := NewCard(s, tt.rank)
			if got := card.Movement(); got != tt.movement {
				t.Errorf("%s movement = %d, want %d", card, got, tt.movement)
			}
			if got := card.Ability(); got != tt.ability {
				t.Errorf("%s ability = %q, want %q", card, got, tt.ability)
			}
		}
	}
}

func TestCardValid(t *testing.T) {
	tests := []struct {
		card Card
		want bool
	}{
		{NewCard(SuitDoves, 1), true},
		{NewCard(SuitStars, 10), true},
		{NewCard(SuitRoses, 0), false},
		{NewCard(SuitRoses, 11), false},
		{NewCard("hearts", 5), false},
		{Card{}, false},
	}
	for _, tt := range tests {
		if got := tt.card.Valid(); got != tt.want {
			t.Errorf("%v.Valid() = %v, want %v", tt.card, got, tt.want)
		}
	}
	if got := NewCard(SuitRoses, 11).Movement(); got != 0 {
		t.Errorf("invalid rank movement = %d, want 0", got)
	}
}

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if len(deck) != 30 {
		t.Fatalf("deck has %d cards, want 30", len(deck))
	}
	seen := map[Card]bool{}
	for _, card := range deck {
		if !card.Valid() || seen[card] {
			t.Fatalf("bad or duplicate card %v", card)
		}
		seen[card] = true
	}
}

func TestSortCards(t *testing.T) {
	got := []Card{c("3S"), c("10D"), c("1R"), c("2D"), c("9S")}
	SortCards(got)
	want := []Card{c("2D"), c("10D"), c("1R"), c("3S"), c("9S")}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
