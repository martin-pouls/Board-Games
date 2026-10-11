package domain

import (
	"fmt"
	"slices"
)

// Suit is a card suit: one of the four colours or trump. Planet Nine calls trump "rockets",
// Deep Sea "submarines"; the domain says trump.
type Suit string

const (
	SuitPink   Suit = "pink"
	SuitBlue   Suit = "blue"
	SuitGreen  Suit = "green"
	SuitYellow Suit = "yellow"
	SuitTrump  Suit = "trump"
)

// Colours are the four colour suits in canonical order.
var Colours = [4]Suit{SuitPink, SuitBlue, SuitGreen, SuitYellow}

// Suits is the canonical order of all suits (deck building, sorting).
var Suits = [5]Suit{SuitPink, SuitBlue, SuitGreen, SuitYellow, SuitTrump}

// suitIndex returns the position of s in Suits, or -1.
func suitIndex(s Suit) int {
	for i, x := range Suits {
		if x == s {
			return i
		}
	}
	return -1
}

// Rank is a card value: 1..9 for colours, 1..4 for trump.
type Rank int

const (
	MaxColourRank Rank = 9
	MaxTrumpRank  Rank = 4
)

// Card is a playing card.
type Card struct {
	Suit Suit
	Rank Rank
}

// NewCard returns the card of suit s and rank r. It does not validate; see Valid.
func NewCard(s Suit, r Rank) Card { return Card{Suit: s, Rank: r} }

// IsTrump reports whether c is a trump card.
func (c Card) IsTrump() bool { return c.Suit == SuitTrump }

// Valid reports whether c is one of the 40 cards.
func (c Card) Valid() bool {
	switch {
	case c.Suit == SuitTrump:
		return c.Rank >= 1 && c.Rank <= MaxTrumpRank
	case suitIndex(c.Suit) >= 0:
		return c.Rank >= 1 && c.Rank <= MaxColourRank
	}
	return false
}

// String renders the card for tests and error details, e.g. "pink 7".
func (c Card) String() string {
	return fmt.Sprintf("%s %d", string(c.Suit), int(c.Rank))
}

// CommanderCard is the card whose holder is the commander (rules §4).
var CommanderCard = Card{Suit: SuitTrump, Rank: 4}

// ColourCards returns the 36 colour cards, unshuffled; the task deck has one task per colour card.
func ColourCards() []Card {
	out := make([]Card, 0, len(Colours)*int(MaxColourRank))
	for _, s := range Colours {
		for r := Rank(1); r <= MaxColourRank; r++ {
			out = append(out, NewCard(s, r))
		}
	}
	return out
}

// NewDeck returns the 40 playing cards, unshuffled: the colour cards, then trump 1..4.
func NewDeck() []Card {
	out := ColourCards()
	for r := Rank(1); r <= MaxTrumpRank; r++ {
		out = append(out, NewCard(SuitTrump, r))
	}
	return out
}

// SortCards sorts cards in place: by suit (Suits order), then by rank ascending.
func SortCards(cs []Card) {
	slices.SortFunc(cs, compareCards)
}

func compareCards(a, b Card) int {
	if a.Suit != b.Suit {
		return suitIndex(a.Suit) - suitIndex(b.Suit)
	}
	return int(a.Rank) - int(b.Rank)
}

// cloneCards copies cs; the result is never nil.
func cloneCards(cs []Card) []Card {
	out := make([]Card, len(cs))
	copy(out, cs)
	return out
}

// cloneCard copies a card pointer.
func cloneCard(c *Card) *Card {
	if c == nil {
		return nil
	}
	v := *c
	return &v
}

// removeCard returns cs without the first occurrence of c, and whether it was found. It does not
// modify cs.
func removeCard(cs []Card, c Card) ([]Card, bool) {
	i := slices.Index(cs, c)
	if i < 0 {
		return cs, false
	}
	out := make([]Card, 0, len(cs)-1)
	out = append(out, cs[:i]...)
	return append(out, cs[i+1:]...), true
}
