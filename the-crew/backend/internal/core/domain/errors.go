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
	ErrInvalidMission      = errors.New("invalid mission")
	ErrNotCommander        = errors.New("not commander")
	ErrCardNotInHand       = errors.New("card not in hand")
	ErrMustFollowSuit      = errors.New("must follow suit")
	ErrCannotCommunicate   = errors.New("cannot communicate")
	ErrAlreadyCommunicated = errors.New("already communicated")
	ErrInvalidPosition     = errors.New("invalid position") // the statement about the card is false
	ErrInvalidSeat         = errors.New("invalid seat")
	ErrInvalidTask         = errors.New("invalid task")
	ErrUnevenDistribution  = errors.New("uneven distribution")
	ErrCannotPassTrump     = errors.New("cannot pass trump")
	ErrAlreadyAnswered     = errors.New("already answered") // poll answer or pass card already given
	ErrNotAllowed          = errors.New("not allowed")      // the action is not available now
	ErrInvalidDirection    = errors.New("invalid direction")
	ErrInvalidTokenMove    = errors.New("invalid token move")
	ErrInvalidResultChoice = errors.New("invalid result choice") // next attempt after a win, next mission after a loss
)
