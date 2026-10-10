package domain

import (
	"fmt"
	"slices"
	"strconv"
)

// Suit is a card suit. The Jester has no suit.
type Suit string

const (
	SuitNone     Suit = "" // the Jester
	SuitHearts   Suit = "hearts"
	SuitDiamonds Suit = "diamonds"
	SuitClubs    Suit = "clubs"
	SuitSpades   Suit = "spades"
)

// Suits is the canonical order (deck building, sorting, power listing).
var Suits = [4]Suit{SuitHearts, SuitDiamonds, SuitClubs, SuitSpades}

// suitIndex returns the position of s in Suits, or -1 (SuitNone and unknown suits).
func suitIndex(s Suit) int {
	for i, x := range Suits {
		if x == s {
			return i
		}
	}
	return -1
}

// suitSymbols are the glyphs used by Card.String, in Suits order.
var suitSymbols = map[Suit]string{
	SuitHearts:   "♥",
	SuitDiamonds: "♦",
	SuitClubs:    "♣",
	SuitSpades:   "♠",
}

// Rank is a card rank.
type Rank int

const (
	RankJester Rank = 0
	RankAce    Rank = 1 // "Animal Companion"; 2..10 are plain numbers
	RankJack   Rank = 11
	RankQueen  Rank = 12
	RankKing   Rank = 13
)

// Card is a playing card. Values are derived by methods, never stored.
type Card struct {
	Suit Suit
	Rank Rank
}

// NewCard returns the card of suit s and rank r. It does not validate; see Valid.
func NewCard(s Suit, r Rank) Card { return Card{Suit: s, Rank: r} }

// Jester returns the Jester card.
func Jester() Card { return Card{Suit: SuitNone, Rank: RankJester} }

// Valid reports whether c is the Jester or a suited card with rank Ace..King.
func (c Card) Valid() bool {
	if c.Rank == RankJester {
		return c.Suit == SuitNone
	}
	return suitIndex(c.Suit) >= 0 && c.Rank >= RankAce && c.Rank <= RankKing
}

// IsJester reports whether c is the Jester.
func (c Card) IsJester() bool { return c == Jester() }

// IsCompanion reports whether c is an Animal Companion (Ace).
func (c Card) IsCompanion() bool { return c.Rank == RankAce }

// IsNumber reports whether c is a number card 2..10.
func (c Card) IsNumber() bool { return c.Rank >= 2 && c.Rank <= 10 }

// IsRoyal reports whether c is a Jack, Queen or King.
func (c Card) IsRoyal() bool { return c.Rank >= RankJack && c.Rank <= RankKing }

// Value is the card's value in hand, used both for playing and for discarding:
// Jester 0, Ace 1, 2..10 face value, Jack 10, Queen 15, King 20.
func (c Card) Value() int {
	switch {
	case c.IsJester():
		return 0
	case c.IsCompanion(), c.IsNumber():
		return int(c.Rank)
	case c.Rank == RankJack:
		return 10
	case c.Rank == RankQueen:
		return 15
	case c.Rank == RankKing:
		return 20
	}
	return 0
}

// String renders the card for tests and error details: "5♥", "A♠", "10♦", "K♣", "Jester".
func (c Card) String() string {
	if c.IsJester() {
		return "Jester"
	}
	if !c.Valid() {
		return fmt.Sprintf("Card(%q,%d)", string(c.Suit), int(c.Rank))
	}
	var r string
	switch c.Rank {
	case RankAce:
		r = "A"
	case RankJack:
		r = "J"
	case RankQueen:
		r = "Q"
	case RankKing:
		r = "K"
	default:
		r = strconv.Itoa(int(c.Rank))
	}
	return r + suitSymbols[c.Suit]
}

// enemyStats is the enemy table (rules §2).
var enemyStats = map[Rank]struct{ Attack, Health int }{
	RankJack:  {Attack: 10, Health: 20},
	RankQueen: {Attack: 15, Health: 30},
	RankKing:  {Attack: 20, Health: 40},
}

// EnemyAttack is the printed attack of an enemy of rank r: J 10, Q 15, K 20; 0 otherwise.
func EnemyAttack(r Rank) int { return enemyStats[r].Attack }

// EnemyHealth is the health of an enemy of rank r: J 20, Q 30, K 40; 0 otherwise.
func EnemyHealth(r Rank) int { return enemyStats[r].Health }

// HandValue returns the sum of the card values.
func HandValue(cs []Card) int {
	sum := 0
	for _, c := range cs {
		sum += c.Value()
	}
	return sum
}

// SortCards sorts cards in place: by rank ascending (Jester first, then A, 2..10, J, Q, K),
// ties by Suits order.
func SortCards(cs []Card) {
	slices.SortFunc(cs, compareCards)
}

// compareCards orders cards by rank, then by suit (Suits order).
func compareCards(a, b Card) int {
	if a.Rank != b.Rank {
		return int(a.Rank) - int(b.Rank)
	}
	return suitIndex(a.Suit) - suitIndex(b.Suit)
}
