package domain

// ActivePlayerStuck reports whether the active player has no legal card play
// and cannot end the turn: len(LegalMoves(activeID)) == 0 && CanEndTurn() != nil.
// There is no lookahead for reaching the minimum (plan 11.1). It is false unless playing.
func (g *Game) ActivePlayerStuck() bool {
	if g.Phase != PhasePlaying {
		return false
	}
	active := g.Players[g.Turn.ActiveSeat].ID
	return len(g.LegalMoves(active)) == 0 && g.CanEndTurn() != nil
}

// checkStuck ends the game as lost when the active player is stuck.
func (g *Game) checkStuck() {
	if g.ActivePlayerStuck() {
		g.finish(OutcomeLost, g.Players[g.Turn.ActiveSeat].ID)
	}
}
