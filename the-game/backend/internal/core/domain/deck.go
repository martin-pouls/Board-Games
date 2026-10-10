package domain

import "fmt"

// NewDeck builds the unshuffled deck for the ruleset's variant (variant-specific factory).
// classic: NumberCard(LowestCard..HighestCard) ascending. extreme: ErrInvalidOptions (Phase 2).
func NewDeck(r Ruleset) ([]Card, error) {
	switch r.Variant {
	case VariantClassic:
		deck := make([]Card, 0, max(0, r.HighestCard-r.LowestCard+1))
		for v := r.LowestCard; v <= r.HighestCard; v++ {
			deck = append(deck, NumberCard(v))
		}
		return deck, nil
	default:
		return nil, fmt.Errorf("%w: no deck for variant %q", ErrInvalidOptions, r.Variant)
	}
}
