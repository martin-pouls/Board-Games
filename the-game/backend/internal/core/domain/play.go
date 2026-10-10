package domain

import (
	"fmt"
	"slices"
)

// EndTurnCheck is an end-of-turn hook: the minimum, plus the Skull and 3! commands.
type EndTurnCheck interface{ Check(g *Game) error }

// minimumCheck requires the active player to have played the required minimum this turn.
// A STOP played this turn overrides the minimum (rules.md §2).
type minimumCheck struct{}

func (minimumCheck) Check(g *Game) error {
	if g.commands().stopPlayed {
		return nil
	}
	played, need := len(g.Turn.Plays), g.RequiredMinimum()
	if played < need {
		return fmt.Errorf("%w: played %d of %d cards", ErrMinimumNotReached, played, need)
	}
	return nil
}

// skullCheck forbids ending the turn while a skull played this turn is on top of a pile.
type skullCheck struct{}

func (skullCheck) Check(g *Game) error {
	if piles := g.commands().skullPiles; len(piles) > 0 {
		return fmt.Errorf("%w: skull on pile %d", ErrSkullUncovered, piles[0])
	}
	return nil
}

// exactCountCheck requires exactly 3 cards once a 3! was played this turn.
type exactCountCheck struct{}

func (exactCountCheck) Check(g *Game) error {
	cs := g.commands()
	if played := len(g.Turn.Plays); cs.exactCount > 0 && played != cs.exactCount {
		return fmt.Errorf("%w: played %d, need exactly %d", ErrExactCountRequired, played, cs.exactCount)
	}
	return nil
}

// endTurnChecks returns the end-of-turn hooks, most specific message first.
func (g *Game) endTurnChecks() []EndTurnCheck {
	return []EndTurnCheck{skullCheck{}, exactCountCheck{}, minimumCheck{}}
}

// runEndTurnChecks runs every end-turn check without the phase check
// (used by CanEndTurn and on withPlay projections).
func (g *Game) runEndTurnChecks() error {
	for _, c := range g.endTurnChecks() {
		if err := c.Check(g); err != nil {
			return err
		}
	}
	return nil
}

// drawCount is the draw policy at turn end: refill up to the hand size; only 1 card
// while a Draw only one card is visible (evaluated at turn end, rules.md §2); capped
// by the draw pile.
func (g *Game) drawCount() int {
	hand := len(g.Players[g.Turn.ActiveSeat].Hand)
	n := max(0, g.Rules.HandSize(len(g.Players))-hand) // refill
	if g.commands().drawOnlyOne {
		n = min(n, drawOnlyOneCount)
	}
	return min(n, len(g.DrawPile))
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

// PlayCard plays a card from by's hand onto a pile. A STOP ends the turn automatically,
// as does an empty hand once the turn may end.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn, ErrCardNotInHand, ErrInvalidPile,
// ErrIllegalPlacement, ErrBackwardsTrickBlocked, ErrSinglePileViolation, ErrExactCountExceeded,
// ErrExactCountUnreachable, ErrStopTooEarly, ErrSkullUncovered, ErrSkullUncoverable.
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
			// An uncovered Skull as the very last card: lost with score 0 (rules.md §2).
			g.finish(OutcomeLost, by)
		}
	case (card.Command == CommandStop || len(g.Players[seat].Hand) == 0) && g.CanEndTurn() == nil:
		// STOP always lands here: stopRule pre-validated the after-state.
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
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn, ErrSkullUncovered,
// ErrExactCountRequired, ErrMinimumNotReached.
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
	cons := g.placementConstraints()
	var moves []Move
	for _, c := range g.Players[seat].Hand {
		piles, _ := g.cardLegality(cons, c)
		for _, pi := range piles {
			moves = append(moves, Move{Card: c.Value, Pile: pi})
		}
	}
	return moves
}

// cardLegality checks c against every pile with the given constraints: the legal piles
// (nil if none), and, when there are none, the first command veto (in pile order) of a
// placement the base rules allowed (nil otherwise). It does not check phase or turn.
func (g *Game) cardLegality(cons []PlacementConstraint, c Card) (piles []int, blocked error) {
	for pi := range g.Piles {
		p, err := g.basePlacementAt(c, pi)
		if err != nil {
			continue
		}
		if err := g.checkConstraints(cons, p); err != nil {
			if blocked == nil {
				blocked = err
			}
			continue
		}
		piles = append(piles, pi)
	}
	if len(piles) > 0 {
		return piles, nil
	}
	return nil, blocked
}

// CanEndTurn returns nil if the active player may end the turn now (phase check, then all
// end-turn checks: skull, 3!, minimum).
func (g *Game) CanEndTurn() error {
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: game is not being played", ErrWrongPhase)
	}
	return g.runEndTurnChecks()
}

// RequiredMinimum returns Rules.MinCardsToPlay(len(DrawPile) == 0).
func (g *Game) RequiredMinimum() int {
	return g.Rules.MinCardsToPlay(len(g.DrawPile) == 0)
}
