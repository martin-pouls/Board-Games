package domain

import "errors"

// Sentinel domain errors. Methods wrap them with detail via fmt.Errorf("%w: ...");
// callers match them with errors.Is.
var (
	ErrInvalidName     = errors.New("invalid name")
	ErrNameTaken       = errors.New("name taken")
	ErrGameFull        = errors.New("game full")
	ErrDuplicatePlayer = errors.New("duplicate player")
	ErrWrongPhase      = errors.New("wrong phase")
	ErrWrongStep       = errors.New("wrong step")
	ErrNotHost         = errors.New("not host")
	ErrUnknownPlayer   = errors.New("unknown player")
	ErrNotYourTurn     = errors.New("not your turn")
	ErrCannotLeave     = errors.New("cannot leave") // sole player in the lobby
	ErrCardNotInHand   = errors.New("card not in hand")
	ErrIllegalPlay     = errors.New("illegal play")
	ErrIllegalDiscard  = errors.New("illegal discard")
	ErrYieldNotAllowed = errors.New("yield not allowed")
	ErrNotSolo         = errors.New("not solo")
	ErrNoSoloJester    = errors.New("no solo jester left")
	ErrIllegalSignal   = errors.New("illegal signal")
	ErrRematchExists   = errors.New("rematch exists")
)
