package domain

import "errors"

// Sentinel domain errors. Methods wrap them with detail via fmt.Errorf("%w: ...");
// callers match them with errors.Is.
var (
	ErrInvalidName       = errors.New("invalid name")
	ErrNameTaken         = errors.New("name taken")
	ErrGameFull          = errors.New("game full")
	ErrDuplicatePlayer   = errors.New("duplicate player")
	ErrInvalidOptions    = errors.New("invalid options")
	ErrWrongPhase        = errors.New("wrong phase")
	ErrNotHost           = errors.New("not host")
	ErrUnknownPlayer     = errors.New("unknown player")
	ErrNotYourTurn       = errors.New("not your turn")
	ErrCardNotInHand     = errors.New("card not in hand")
	ErrInvalidPile       = errors.New("invalid pile")
	ErrIllegalPlacement  = errors.New("illegal placement")
	ErrMinimumNotReached = errors.New("minimum not reached")
	ErrRematchExists     = errors.New("rematch exists")
)
