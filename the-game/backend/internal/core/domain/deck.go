package domain

import "fmt"

// extremeCommandValues maps each Extreme command to the values of its 4 cards
// (rules.md §2 "Command → card values"). It is deck data, not a ruleset parameter.
var extremeCommandValues = map[Command][4]int{
	CommandStop:             {6, 35, 57, 95},
	CommandSkull:            {17, 41, 61, 87},
	CommandExactlyThree:     {19, 28, 53, 81},
	CommandNoCommunication:  {26, 65, 78, 92},
	CommandNoBackwardsTrick: {21, 32, 59, 70},
	CommandSinglePile:       {9, 47, 68, 84},
	CommandDrawOnlyOne:      {13, 38, 43, 76},
}

// NewDeck builds the unshuffled deck for the ruleset's variant (variant-specific factory).
// classic: NumberCard(LowestCard..HighestCard) ascending.
// extreme: the same cards, 28 of which carry a command (extremeCommandValues).
// Errors: ErrInvalidOptions for an unknown variant or a command value outside the deck.
func NewDeck(r Ruleset) ([]Card, error) {
	switch r.Variant {
	case VariantClassic:
		return numberDeck(r), nil
	case VariantExtreme:
		deck := numberDeck(r)
		for cmd, vals := range extremeCommandValues {
			for _, v := range vals {
				if v < r.LowestCard || v > r.HighestCard {
					return nil, fmt.Errorf("%w: %s card %d outside the deck %d..%d", ErrInvalidOptions, cmd, v, r.LowestCard, r.HighestCard)
				}
				deck[v-r.LowestCard].Command = cmd
			}
		}
		return deck, nil
	default:
		return nil, fmt.Errorf("%w: no deck for variant %q", ErrInvalidOptions, r.Variant)
	}
}

// numberDeck returns NumberCard(LowestCard..HighestCard) ascending.
func numberDeck(r Ruleset) []Card {
	deck := make([]Card, 0, max(0, r.HighestCard-r.LowestCard+1))
	for v := r.LowestCard; v <= r.HighestCard; v++ {
		deck = append(deck, NumberCard(v))
	}
	return deck
}
