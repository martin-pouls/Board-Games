package domain

import "fmt"

// SpendReroll spends a reroll token from the supply during placement (either role, §5.3) and
// opens the reroll step. A role without unplaced dice is confirmed at once.
// RNG use: none unless the step ends at once.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNoReroll. Emits RerollSpent.
func (g *Game) SpendReroll(by PlayerID, rng RNG) error {
	return fmt.Errorf("%w: SpendReroll", ErrNotImplemented)
}

// ConfirmReroll rerolls by's dice in slots (possibly none) and confirms (§5.3). When both have
// confirmed, placement continues.
// RNG use order: the given slots, ascending.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrAlreadyConfirmed, ErrNoSuchDie
// (a slot out of range, empty or repeated). Emits RerollConfirmed (without values).
func (g *Game) ConfirmReroll(by PlayerID, slots []int, rng RNG) error {
	return fmt.Errorf("%w: ConfirmReroll", ErrNotImplemented)
}
