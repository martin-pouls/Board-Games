package domain

import "errors"

// Sentinel domain errors. Methods wrap them with detail via fmt.Errorf("%w: ...");
// callers match them with errors.Is.
var (
	ErrInvalidName         = errors.New("invalid name")
	ErrNameTaken           = errors.New("name taken")
	ErrGameFull            = errors.New("game full")
	ErrDuplicatePlayer     = errors.New("duplicate player")
	ErrWrongPhase          = errors.New("wrong phase")
	ErrWrongStep           = errors.New("wrong step")
	ErrNotHost             = errors.New("not host")
	ErrUnknownPlayer       = errors.New("unknown player")
	ErrNotYourTurn         = errors.New("not your turn")
	ErrCannotLeave         = errors.New("cannot leave") // sole player in the lobby
	ErrNotEnoughPlayers    = errors.New("not enough players")
	ErrUnknownScenario     = errors.New("unknown scenario")
	ErrInvalidAbilities    = errors.New("invalid abilities")    // unknown, duplicate or too many
	ErrAbilitiesIncomplete = errors.New("abilities incomplete") // fewer than the scenario's count at Start
	ErrAlreadyReady        = errors.New("already ready")
	ErrNoSuchDie           = errors.New("no such die") // slot out of range or not in the hand
	ErrNotEnoughCoffee     = errors.New("not enough coffee")
	ErrValueOutOfRange     = errors.New("value out of range") // die value with coffee outside 1..6
	ErrIllegalSpace        = errors.New("illegal space")
	ErrIllegalTokenSpace   = errors.New("illegal token space") // intern token target
	ErrNoReroll            = errors.New("no reroll")
	ErrAlreadyConfirmed    = errors.New("already confirmed")
	ErrAbilityNotInPlay    = errors.New("ability not in play")
	ErrAbilityUsed         = errors.New("ability used")
	ErrPartnerHasNoDice    = errors.New("partner has no dice")
	ErrRematchExists       = errors.New("rematch exists")
	ErrNotImplemented      = errors.New("not implemented")
)
