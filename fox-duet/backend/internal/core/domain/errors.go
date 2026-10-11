package domain

import "errors"

// Sentinel domain errors. Methods wrap them with detail via fmt.Errorf("%w: ...");
// callers match them with errors.Is.
var (
	ErrInvalidName       = errors.New("invalid name")
	ErrNameTaken         = errors.New("name taken")
	ErrGameFull          = errors.New("game full")
	ErrDuplicatePlayer   = errors.New("duplicate player")
	ErrWrongPhase        = errors.New("wrong phase")
	ErrWrongStep         = errors.New("wrong step")
	ErrNotHost           = errors.New("not host")
	ErrUnknownPlayer     = errors.New("unknown player")
	ErrNotYourTurn       = errors.New("not your turn")
	ErrCannotLeave       = errors.New("cannot leave") // sole player in the lobby
	ErrNotEnoughPlayers  = errors.New("not enough players")
	ErrInvalidLevel      = errors.New("invalid level")
	ErrCardNotInHand     = errors.New("card not in hand")
	ErrMustFollowSuit    = errors.New("must follow suit")
	ErrGiftAlreadyChosen = errors.New("gift already chosen")
	ErrIllegalMove       = errors.New("illegal move")
	ErrIllegalEnd        = errors.New("illegal end")
	ErrRematchExists     = errors.New("rematch exists")
)
