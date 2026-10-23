package domain

import "fmt"

// PlaceIntent is one placement (§5.2): the die in hand Slot goes on Space, changed by Coffee
// (the net modifier; it costs |Coffee| tokens, §5.4). TokenSpace is the intern token's target
// and is required exactly when Space is an intern space (§12.2).
type PlaceIntent struct {
	Slot       int
	Space      SpaceID
	Coffee     int
	TokenSpace SpaceID
}

// Place puts one of by's dice on the panel on their turn and resolves the space (§6), then
// passes the turn (plan §5.3 resolution order). Intern training places the die and the token in
// the same intent.
// RNG use: only if the placement ends the round (the next round's traffic rolls).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn, ErrNoSuchDie,
// ErrNotEnoughCoffee, ErrValueOutOfRange, ErrIllegalSpace, ErrIllegalTokenSpace.
// Emits DiePlaced and whatever the space and the end of the round emit.
func (g *Game) Place(by PlayerID, in PlaceIntent, rng RNG) error {
	return fmt.Errorf("%w: Place", ErrNotImplemented)
}

// legalSpaces returns the spaces role may put piece of value on now, in table order: colour,
// value, free this round, flaps and brakes in order, open ice columns, the intern and
// concentration exceptions (§6, §12, §13).
// Not implemented yet: always empty.
func (g *Game) legalSpaces(role Role, piece PieceKind, value int) []SpaceID {
	return []SpaceID{}
}
