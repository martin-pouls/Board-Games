package domain

import "slices"

// LossReason says why a game was lost (§7.2).
type LossReason string

const (
	LossSpin          LossReason = "spin"
	LossCollision     LossReason = "collision"
	LossOvershoot     LossReason = "overshoot"
	LossTurn          LossReason = "turn"
	LossMissedAirport LossReason = "missedAirport"
	LossMandatory     LossReason = "mandatory"
	LossKerosene      LossReason = "kerosene"
	LossLanding       LossReason = "landing"
	LossIntern        LossReason = "intern"
	LossIceBrakes     LossReason = "iceBrakes"
	LossGaveUp        LossReason = "gaveUp"
)

// Condition is a landing condition of the final round (§7.1) or a module's end condition.
type Condition string

const (
	ConditionA         Condition = "A"         // no plane tokens left
	ConditionB         Condition = "B"         // all gear and flaps deployed
	ConditionC         Condition = "C"         // axis at 0
	ConditionD         Condition = "D"         // final speed below the brake marker
	ConditionIntern    Condition = "intern"    // all intern tokens used (§12.2)
	ConditionIceBrakes Condition = "iceBrakes" // ice-brake marker past 5 (§12.6)
)

// Result is set iff Phase is won or lost.
type Result struct {
	Reason   LossReason  // empty on a win
	Failed   []Condition // the failed conditions of a landing/intern/iceBrakes loss, in Condition order
	PlayerID PlayerID    // who gave up; empty otherwise
}

// clone returns a deep copy; nil stays nil.
func (r *Result) clone() *Result {
	if r == nil {
		return nil
	}
	c := *r
	c.Failed = slices.Clone(r.Failed)
	return &c
}

// win ends the game as a victory. The flight stays as it is.
func (g *Game) win() {
	g.Phase = PhaseWon
	g.Result = &Result{}
	g.emit(GameWon{})
}

// lose ends the game as a loss. The flight, the panel and the hands stay as they are, so the
// view can reveal the unplaced dice (§7.2).
func (g *Game) lose(reason LossReason, failed []Condition, by PlayerID) {
	g.Phase = PhaseLost
	g.Result = &Result{Reason: reason, Failed: slices.Clone(failed), PlayerID: by}
	g.emit(GameLost{Reason: reason, Failed: slices.Clone(failed), PlayerID: by})
}
