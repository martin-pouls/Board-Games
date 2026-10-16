package domain

// ActivePlayerStuck reports whether the active player has no legal card play
// and cannot end the turn: len(LegalMoves(activeID)) == 0 && CanEndTurn() != nil.
// Both sides include the Extreme commands: LegalMoves applies every placement constraint
// and CanEndTurn the Skull and 3! checks. There is no lookahead for reaching the minimum
// or fulfilling a command (plan 11.1). It is false unless playing.
// Mid-turn a stuck player may still undo cards or give up; see checkStuck.
func (g *Game) ActivePlayerStuck() bool {
	if g.Phase != PhasePlaying {
		return false
	}
	active := g.Players[g.Turn.ActiveSeat].ID
	return len(g.LegalMoves(active)) == 0 && g.CanEndTurn() != nil
}

// checkStuck ends the game as lost when the active player is stuck with nothing to undo
// (at the start of a turn).
func (g *Game) checkStuck() {
	if len(g.Turn.Plays) == 0 && g.ActivePlayerStuck() {
		g.lose(LossStuck, g.Players[g.Turn.ActiveSeat].ID)
	}
}
