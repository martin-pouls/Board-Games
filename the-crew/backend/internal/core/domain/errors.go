package domain

import "errors"

// Sentinel domain errors. Methods wrap them with detail via fmt.Errorf("%w: ...");
// callers match them with errors.Is.
var (
	ErrInvalidName      = errors.New("invalid name")
	ErrNameTaken        = errors.New("name taken")
	ErrGameFull         = errors.New("game full")
	ErrDuplicatePlayer  = errors.New("duplicate player")
	ErrWrongPhase       = errors.New("wrong phase")
	ErrNotHost          = errors.New("not host")
	ErrUnknownPlayer    = errors.New("unknown player")
	ErrCannotLeave      = errors.New("cannot leave") // sole player in the lobby
	ErrNotEnoughPlayers = errors.New("not enough players")
)
