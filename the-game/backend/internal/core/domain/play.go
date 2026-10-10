package domain

import (
	"fmt"
	"slices"
)

// EndTurnCheck is an end-of-turn hook (Phase 1: minimum check; Extreme adds Skull/3!).
type EndTurnCheck interface{ Check(g *Game) error }

// minimumCheck requires the active player to have played the required minimum this turn.
type minimumCheck struct{}

func (minimumCheck) Check(g *Game) error {
	played, need := len(g.Turn.Plays), g.RequiredMinimum()
	if played < need {
		return fmt.Errorf("%w: played %d of %d cards", ErrMinimumNotReached, played, need)
	}
	return nil
}

// endTurnChecks returns the end-of-turn hooks. Phase 1: the minimum check only.
func (g *Game) endTurnChecks() []EndTurnCheck {
	return []EndTurnCheck{minimumCheck{}}
}

// drawCount is the draw policy at turn end. Phase 1 "refill": up to the hand size,
// capped by the draw pile.
func (g *Game) drawCount() int {
	hand := len(g.Players[g.Turn.ActiveSeat].Hand)
	return min(max(0, g.Rules.HandSize(len(g.Players))-hand), len(g.DrawPile))
}

// requireActive checks that by is seated, the game is playing and it is by's turn.
func (g *Game) requireActive(by PlayerID) (int, error) {
	_, seat, ok := g.PlayerByID(by)
	if !ok {
		return -1, fmt.Errorf("%w: %s", ErrUnknownPlayer, by)
	}
	if g.Phase != PhasePlaying {
		return -1, fmt.Errorf("%w: game is not being played", ErrWrongPhase)
	}
	if seat != g.Turn.ActiveSeat {
		return -1, fmt.Errorf("%w: it is %s's turn", ErrNotYourTurn, g.Players[g.Turn.ActiveSeat].ID)
	}
	return seat, nil
}

// PlayCard plays a card from by's hand onto a pile.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn, ErrCardNotInHand, ErrInvalidPile, ErrIllegalPlacement.
func (g *Game) PlayCard(by PlayerID, value int, pile int) error {
	seat, err := g.requireActive(by)
	if err != nil {
		return err
	}
	hand := g.Players[seat].Hand
	idx := slices.IndexFunc(hand, func(c Card) bool { return c.Value == value })
	if idx < 0 {
		return fmt.Errorf("%w: %d", ErrCardNotInHand, value)
	}
	card := hand[idx]
	if _, err := g.checkPlacement(card, pile); err != nil {
		return err
	}

	// Apply.
	g.Players[seat].Hand = slices.Delete(hand, idx, idx+1)
	g.Piles[pile].Cards = append(g.Piles[pile].Cards, card)
	g.Turn.Plays = append(g.Turn.Plays, Play{Card: card, Pile: pile})
	g.emit(CardPlayed{PlayerID: by, Card: card, Pile: pile})

	// Follow-ups.
	switch {
	case g.noCardsLeft():
		if g.CanEndTurn() == nil {
			g.finish(OutcomeWon, "")
		} else {
			// Extreme hook: e.g. an uncovered Skull as the very last card.
			g.finish(OutcomeLost, by)
		}
	case len(g.Players[seat].Hand) == 0 && g.CanEndTurn() == nil:
		g.endTurn(true)
	default:
		g.checkStuck()
	}
	g.commit()
	return nil
}

// noCardsLeft reports whether all hands and the draw pile are empty.
func (g *Game) noCardsLeft() bool {
	hands, draw := g.Score()
	return hands == 0 && draw == 0
}

// EndTurn ends the active player's turn, draws and advances.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn, ErrMinimumNotReached.
func (g *Game) EndTurn(by PlayerID) error {
	if _, err := g.requireActive(by); err != nil {
		return err
	}
	if err := g.CanEndTurn(); err != nil {
		return err
	}
	g.endTurn(false)
	g.commit()
	return nil
}

// endTurn draws for the active player, emits TurnEnded and advances. Checks must have passed.
func (g *Game) endTurn(auto bool) {
	seat := g.Turn.ActiveSeat
	n := g.drawCount()
	if n > 0 {
		top := len(g.DrawPile) - n
		drawn := g.DrawPile[top:]
		hand := append(g.Players[seat].Hand, drawn...)
		sortCards(hand)
		g.Players[seat].Hand = hand
		g.DrawPile = slices.Clip(g.DrawPile[:top])
	}
	g.emit(TurnEnded{PlayerID: g.Players[seat].ID, CardsDrawn: n, Auto: auto})
	g.advanceTurn()
}

// advanceTurn passes the turn clockwise to the next player with cards, skipping
// empty hands (only possible once the draw pile is empty). The current seat comes last (solo).
func (g *Game) advanceTurn() {
	n := len(g.Players)
	for i := 1; i <= n; i++ {
		seat := (g.Turn.ActiveSeat + i) % n
		if len(g.Players[seat].Hand) > 0 {
			g.beginTurn(seat)
			return
		}
		g.emit(PlayerSkipped{PlayerID: g.Players[seat].ID})
	}
	// Defensive: PlayCard normally detects that every card was played.
	g.finish(OutcomeWon, "")
}

// beginTurn starts the next turn for seat and checks whether the player is stuck.
func (g *Game) beginTurn(seat int) {
	g.Turn = Turn{
		Number:        g.Turn.Number + 1,
		ActiveSeat:    seat,
		StartHandSize: len(g.Players[seat].Hand),
	}
	g.emit(TurnStarted{PlayerID: g.Players[seat].ID, Number: g.Turn.Number})
	g.checkStuck()
}

// finish ends the game with the given outcome and the score at this moment.
func (g *Game) finish(o Outcome, stuck PlayerID) {
	hands, draw := g.Score()
	g.Phase = PhaseFinished
	g.Result = &Result{
		Outcome:             o,
		Score:               hands + draw,
		CardsLeftInHands:    hands,
		CardsLeftInDrawPile: draw,
		StuckPlayerID:       stuck,
	}
	g.emit(GameFinished{Outcome: o, Score: hands + draw, StuckPlayerID: stuck})
}

// LegalMoves returns every legal (card, pile) for by; nil unless playing and by is active.
func (g *Game) LegalMoves(by PlayerID) []Move {
	if g.Phase != PhasePlaying {
		return nil
	}
	_, seat, ok := g.PlayerByID(by)
	if !ok || seat != g.Turn.ActiveSeat {
		return nil
	}
	var moves []Move
	for _, c := range g.Players[seat].Hand {
		for pi := range g.Piles {
			if _, err := g.checkPlacement(c, pi); err == nil {
				moves = append(moves, Move{Card: c.Value, Pile: pi})
			}
		}
	}
	return moves
}

// CanEndTurn returns nil if the active player may end the turn now (runs all end-turn checks).
func (g *Game) CanEndTurn() error {
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: game is not being played", ErrWrongPhase)
	}
	for _, c := range g.endTurnChecks() {
		if err := c.Check(g); err != nil {
			return err
		}
	}
	return nil
}

// RequiredMinimum returns Rules.MinCardsToPlay(len(DrawPile) == 0).
func (g *Game) RequiredMinimum() int {
	return g.Rules.MinCardsToPlay(len(g.DrawPile) == 0)
}
