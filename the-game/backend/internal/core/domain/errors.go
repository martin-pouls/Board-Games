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
	ErrNothingToUndo     = errors.New("nothing to undo") // UndoPlay: no card played this turn
	ErrNotStuck          = errors.New("not stuck")       // GiveUp: a card fits or the turn can end

	// Extreme command vetoes (rules.md §2 Enforcement). None of them wraps
	// ErrIllegalPlacement or ErrMinimumNotReached.
	ErrStopPlayed            = errors.New("stop played")             // PlayCard: a STOP was played this turn
	ErrBackwardsTrickBlocked = errors.New("backwards trick blocked") // PlayCard: a No backwards trick card is visible
	ErrSinglePileViolation   = errors.New("single pile violation")   // PlayCard: card not on the bound pile
	ErrExactCountExceeded    = errors.New("exact count exceeded")    // PlayCard: a 4th card after 3!, or 3! as the 4th+ card
	ErrExactCountUnreachable = errors.New("exact count unreachable") // PlayCard: 3! when the turn started with < 3 cards
	ErrStopTooEarly          = errors.New("stop too early")          // PlayCard: STOP after 3! as the 1st or 2nd card
	ErrSkullUncovered        = errors.New("skull uncovered")         // EndTurn; PlayCard (STOP): a skull from this turn on top
	ErrExactCountRequired    = errors.New("exact count required")    // EndTurn: fewer than 3 cards after 3!
	ErrSkullUncoverable      = errors.New("skull uncoverable")       // PlayCard: skull could never be covered because of 3!
)
