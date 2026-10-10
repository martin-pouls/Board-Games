package domain

// EndTurnCheck is an end-of-turn hook (Phase 1: minimum check; Extreme adds Skull/3!).
type EndTurnCheck interface{ Check(g *Game) error }

// PlayCard plays a card from by's hand onto a pile.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn, ErrCardNotInHand, ErrInvalidPile, ErrIllegalPlacement.
func (g *Game) PlayCard(by PlayerID, value int, pile int) error {
	panic("not implemented: WP3")
}

// EndTurn ends the active player's turn, draws and advances.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn, ErrMinimumNotReached.
func (g *Game) EndTurn(by PlayerID) error {
	panic("not implemented: WP3")
}

// LegalMoves returns every legal (card, pile) for by; nil unless playing and by is active.
func (g *Game) LegalMoves(by PlayerID) []Move {
	panic("not implemented: WP3")
}

// CanEndTurn returns nil if the active player may end the turn now (runs all end-turn checks).
func (g *Game) CanEndTurn() error {
	panic("not implemented: WP3")
}

// RequiredMinimum returns Rules.MinCardsToPlay(len(DrawPile) == 0).
func (g *Game) RequiredMinimum() int {
	panic("not implemented: WP3")
}

// ActivePlayerStuck reports whether the active player has no legal card play
// and cannot end the turn: len(LegalMoves(activeID)) == 0 && CanEndTurn() != nil.
// There is no lookahead for reaching the minimum (plan 11.1).
func (g *Game) ActivePlayerStuck() bool {
	panic("not implemented: WP3")
}
