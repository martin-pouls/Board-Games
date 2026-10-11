package domain

import (
	"fmt"
	"slices"
)

// Suit is a card suit.
type Suit string

const (
	SuitDoves Suit = "doves"
	SuitRoses Suit = "roses"
	SuitStars Suit = "stars"
)

// Suits is the canonical order (deck building, sorting).
var Suits = [3]Suit{SuitDoves, SuitRoses, SuitStars}

// suitIndex returns the position of s in Suits, or -1.
func suitIndex(s Suit) int {
	for i, x := range Suits {
		if x == s {
			return i
		}
	}
	return -1
}

// Rank is a card rank, 1..10.
type Rank int

const (
	MinRank Rank = 1
	MaxRank Rank = 10
)

// Ranks with an ability (rules §2, §6).
const (
	RankMusician  Rank = 1
	RankFoxes     Rank = 3
	RankGazelle   Rank = 5
	RankGift      Rank = 7
	RankRoyalHeir Rank = 9
)

// Ability is the ability of an odd rank; AbilityNone for even ranks.
type Ability string

const (
	AbilityNone      Ability = ""
	AbilityMusician  Ability = "musician"
	AbilityFoxes     Ability = "foxes"
	AbilityGazelle   Ability = "gazelle"
	AbilityGift      Ability = "gift"
	AbilityRoyalHeir Ability = "royalHeir"
)

// movement is the movement value per rank (rules §2); index 0 is unused.
var movement = [MaxRank + 1]int{0, 0, 3, 2, 1, 1, 2, 0, 2, 0, 3}

// abilities maps the odd ranks to their ability.
var abilities = map[Rank]Ability{
	RankMusician:  AbilityMusician,
	RankFoxes:     AbilityFoxes,
	RankGazelle:   AbilityGazelle,
	RankGift:      AbilityGift,
	RankRoyalHeir: AbilityRoyalHeir,
}

// Card is a playing card. Movement and ability are derived from the rank.
type Card struct {
	Suit Suit
	Rank Rank
}

// NewCard returns the card of suit s and rank r. It does not validate; see Valid.
func NewCard(s Suit, r Rank) Card { return Card{Suit: s, Rank: r} }

// Valid reports whether c has a known suit and a rank in 1..10.
func (c Card) Valid() bool {
	return suitIndex(c.Suit) >= 0 && c.Rank >= MinRank && c.Rank <= MaxRank
}

// Movement is the number of spaces the card moves the traveller (0..3).
func (c Card) Movement() int {
	if c.Rank < MinRank || c.Rank > MaxRank {
		return 0
	}
	return movement[c.Rank]
}

// Ability is the card's ability (AbilityNone for even ranks).
func (c Card) Ability() Ability { return abilities[c.Rank] }

// String renders the card for tests and error details, e.g. "7 roses".
func (c Card) String() string {
	return fmt.Sprintf("%d %s", int(c.Rank), string(c.Suit))
}

// NewDeck returns the 30 cards, unshuffled: for s in Suits, ranks 1..10.
func NewDeck() []Card {
	out := make([]Card, 0, len(Suits)*int(MaxRank))
	for _, s := range Suits {
		for r := MinRank; r <= MaxRank; r++ {
			out = append(out, NewCard(s, r))
		}
	}
	return out
}

// SortCards sorts cards in place: by suit (Suits order), then by rank ascending.
func SortCards(cs []Card) {
	slices.SortFunc(cs, func(a, b Card) int {
		if a.Suit != b.Suit {
			return suitIndex(a.Suit) - suitIndex(b.Suit)
		}
		return int(a.Rank) - int(b.Rank)
	})
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
