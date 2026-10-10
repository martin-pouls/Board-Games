package domain

import (
	"fmt"
	"slices"
)

// PlayKind is the shape of a legal Step 1 play (rules §5).
type PlayKind string

const (
	PlaySingle    PlayKind = "single"
	PlayCombo     PlayKind = "combo"
	PlayCompanion PlayKind = "companion"
	PlayJester    PlayKind = "jester"
)

// maxComboValue is the highest total value of a combo (rules §5).
const maxComboValue = 10

// maxComboCards is the largest combo (four 2s).
const maxComboCards = 4

// Reason names in illegal play/discard errors. The TS mirror (rules.ts) uses the same names.
const (
	reasonEmpty          = "empty"
	reasonJesterNotAlone = "jesterNotAlone"
	reasonCompanionPair  = "companionPair"
	reasonRoyalCombo     = "royalCombo"
	reasonMixedRanks     = "mixedRanks"
	reasonTooManyCards   = "tooManyCards"
	reasonComboTooHigh   = "comboTooHigh"
	reasonNotEnough      = "notEnough"
	reasonSuperfluous    = "superfluous"
)

// ClassifyPlay checks the shape of a play (rules §5). Errors wrap ErrIllegalPlay; the detail is
// the reason name. The checks run in this order (the TS mirror uses the same order and names):
//  1. len 0                      → "empty"
//  2. any Jester: len 1 → PlayJester, else → "jesterNotAlone"
//  3. len 1                      → PlaySingle
//  4. any Ace: len 2 → PlayCompanion (other card may be number, Ace or royal), else → "companionPair"
//  5. any royal                  → "royalCombo"
//  6. ranks differ               → "mixedRanks"
//  7. len > 4                    → "tooManyCards"
//  8. Σ value > 10               → "comboTooHigh"
//  9. → PlayCombo
func ClassifyPlay(cards []Card) (PlayKind, error) {
	illegal := func(reason string) (PlayKind, error) {
		return "", fmt.Errorf("%w: %s", ErrIllegalPlay, reason)
	}
	if len(cards) == 0 {
		return illegal(reasonEmpty)
	}
	if slices.ContainsFunc(cards, Card.IsJester) {
		if len(cards) == 1 {
			return PlayJester, nil
		}
		return illegal(reasonJesterNotAlone)
	}
	if len(cards) == 1 {
		return PlaySingle, nil
	}
	if slices.ContainsFunc(cards, Card.IsCompanion) {
		if len(cards) == 2 {
			return PlayCompanion, nil
		}
		return illegal(reasonCompanionPair)
	}
	if slices.ContainsFunc(cards, Card.IsRoyal) {
		return illegal(reasonRoyalCombo)
	}
	for _, c := range cards[1:] {
		if c.Rank != cards[0].Rank {
			return illegal(reasonMixedRanks)
		}
	}
	if len(cards) > maxComboCards {
		return illegal(reasonTooManyCards)
	}
	if HandValue(cards) > maxComboValue {
		return illegal(reasonComboTooHigh)
	}
	return PlayCombo, nil
}

// ValidateDiscard checks a Step 4 discard set against attack (> 0) (rules §4 App decision).
// The set is legal iff Σ ≥ attack and Σ − min(values) < attack (no superfluous card).
// Errors wrap ErrIllegalDiscard: len 0 → "empty"; Σ < attack → "notEnough";
// Σ − min(values) ≥ attack → "superfluous".
func ValidateDiscard(cards []Card, attack int) error {
	if len(cards) == 0 {
		return fmt.Errorf("%w: %s", ErrIllegalDiscard, reasonEmpty)
	}
	sum := HandValue(cards)
	if sum < attack {
		return fmt.Errorf("%w: %s (%d < %d)", ErrIllegalDiscard, reasonNotEnough, sum, attack)
	}
	minValue := cards[0].Value()
	for _, c := range cards[1:] {
		minValue = min(minValue, c.Value())
	}
	if sum-minValue >= attack {
		return fmt.Errorf("%w: %s", ErrIllegalDiscard, reasonSuperfluous)
	}
	return nil
}

// removeCards returns hand without cards (a sub-multiset of hand). It returns a new slice and
// never aliases hand. Identical cards (the two Jesters) are interchangeable.
// Errors: ErrCardNotInHand when a card is missing or selected more often than the hand holds it.
func removeCards(hand, cards []Card) ([]Card, error) {
	rest := slices.Clone(hand)
	if rest == nil {
		rest = []Card{}
	}
	for _, c := range cards {
		i := slices.Index(rest, c)
		if i < 0 {
			return nil, fmt.Errorf("%w: %s", ErrCardNotInHand, c)
		}
		rest = slices.Delete(rest, i, i+1)
	}
	return rest, nil
}
