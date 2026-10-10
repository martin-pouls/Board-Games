package domain

import "fmt"

// Variant selects the game edition.
type Variant string

const (
	VariantClassic Variant = "classic"
	VariantExtreme Variant = "extreme" // command cards (rules.md §2); same hand sizes and minimums
)

// Options are the host-selectable game settings.
type Options struct {
	Variant         Variant
	Pro             bool // min 3 cards per turn
	ReducedHandSize bool // hand size -1; only allowed together with Pro
}

// DefaultOptions returns the classic, non-Pro options.
func DefaultOptions() Options {
	return Options{Variant: VariantClassic}
}

// Validate reports whether the options are allowed. Errors wrap ErrInvalidOptions.
// An empty Variant is invalid; adapters map "empty on input" to classic.
// Pro and the reduced hand size combine with either variant.
func (o Options) Validate() error {
	switch o.Variant {
	case VariantClassic, VariantExtreme:
	default:
		return fmt.Errorf("%w: unknown variant %q", ErrInvalidOptions, o.Variant)
	}
	if o.ReducedHandSize && !o.Pro {
		return fmt.Errorf("%w: reduced hand size requires the Pro variant", ErrInvalidOptions)
	}
	return nil
}

// Ruleset holds every numeric rule parameter. Rule code reads these values
// instead of hard-coding them.
type Ruleset struct {
	Variant                         Variant
	MinPlayers, MaxPlayers          int
	MinCardsPerTurn                 int
	MinCardsDrawPileEmpty           int
	HandSizes                       map[int]int // player count -> base hand size
	HandSizeReduction               int
	BackwardsTrickStep              int
	AscendingStart                  int
	DescendingStart                 int
	LowestCard, HighestCard         int
	AscendingPiles, DescendingPiles int
}

// NewRuleset validates o and derives the ruleset from it.
func NewRuleset(o Options) (Ruleset, error) {
	if err := o.Validate(); err != nil {
		return Ruleset{}, err
	}
	r := Ruleset{
		Variant:               o.Variant,
		MinPlayers:            1,
		MaxPlayers:            5,
		MinCardsPerTurn:       2,
		MinCardsDrawPileEmpty: 1,
		HandSizes:             map[int]int{1: 8, 2: 7, 3: 6, 4: 6, 5: 6},
		BackwardsTrickStep:    10,
		AscendingStart:        1,
		DescendingStart:       100,
		LowestCard:            2,
		HighestCard:           99,
		AscendingPiles:        2,
		DescendingPiles:       2,
	}
	if o.Pro {
		r.MinCardsPerTurn = 3
		if o.ReducedHandSize {
			r.HandSizeReduction = 1
		}
	}
	return r, nil
}

// HandSize returns the hand size for the given player count, or 0 if the
// count is not in the table.
func (r Ruleset) HandSize(players int) int {
	base, ok := r.HandSizes[players]
	if !ok {
		return 0
	}
	return max(0, base-r.HandSizeReduction)
}

// MinCardsToPlay returns the minimum number of cards the active player must play this turn.
func (r Ruleset) MinCardsToPlay(drawPileEmpty bool) int {
	if drawPileEmpty {
		return r.MinCardsDrawPileEmpty
	}
	return r.MinCardsPerTurn
}
