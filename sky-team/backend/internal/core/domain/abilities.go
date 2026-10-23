package domain

import "fmt"

// Adapt turns one of by's unplaced dice to its opposite face, v -> 7-v (Adaptation, §13): once
// per game per role, during placement, outside the turn order.
// RNG use: none.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrAbilityNotInPlay, ErrAbilityUsed,
// ErrNoSuchDie. Emits AdaptationUsed (without values).
func (g *Game) Adapt(by PlayerID, slot int, rng RNG) error {
	return fmt.Errorf("%w: Adapt", ErrNotImplemented)
}

// Anticipate rerolls one of by's dice (Anticipation, §13): the first player of the round, on
// their turn, before the first placement of the round, once per round.
// RNG use: one die.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrAbilityNotInPlay, ErrAbilityUsed,
// ErrNoSuchDie, ErrNotYourTurn. Emits AnticipationUsed (without values).
func (g *Game) Anticipate(by PlayerID, slot int, rng RNG) error {
	return fmt.Errorf("%w: Anticipate", ErrNotImplemented)
}

// OfferSwap puts one of by's unplaced dice on the Working Together card (§13): either role,
// once per round, during placement, while the partner holds dice. The partner answers next.
// RNG use: none.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrAbilityNotInPlay, ErrAbilityUsed,
// ErrNoSuchDie, ErrPartnerHasNoDice. Emits SwapOffered (the offered value is public).
func (g *Game) OfferSwap(by PlayerID, slot int, rng RNG) error {
	return fmt.Errorf("%w: OfferSwap", ErrNotImplemented)
}

// AnswerSwap is the partner's die for Working Together (§13): the two dice swap values and
// placement continues.
// RNG use: none.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn, ErrNoSuchDie.
// Emits DiceSwapped (without values).
func (g *Game) AnswerSwap(by PlayerID, slot int, rng RNG) error {
	return fmt.Errorf("%w: AnswerSwap", ErrNotImplemented)
}

// PlaceSync puts the Synchronisation traffic die on space (Co-Pilot, §13). It resolves like a
// normal die, without coffee and never on an intern space.
// RNG use: only if the placement ends the round (the next round's traffic rolls).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn, ErrIllegalSpace.
// Emits DiePlaced and whatever the space and the end of the round emit.
func (g *Game) PlaceSync(by PlayerID, space SpaceID, rng RNG) error {
	return fmt.Errorf("%w: PlaceSync", ErrNotImplemented)
}
