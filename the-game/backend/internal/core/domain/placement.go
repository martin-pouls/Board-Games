package domain

import (
	"errors"
	"fmt"
	"slices"
)

// PlacementKind says which base rule allowed a placement.
type PlacementKind int

const (
	PlacementNormal         PlacementKind = iota + 1 // up: card > top; down: card < top
	PlacementBackwardsTrick                          // up: card == top-Step; down: card == top+Step
)

// Placement describes a candidate card placement that the base rule allowed.
type Placement struct {
	Card      Card
	PileIndex int
	Pile      Pile // current state of the target pile
	Kind      PlacementKind
}

// PlacementConstraint can veto a placement the base rule allowed. The constraints are
// derived from the visible Infinity commands (no backwards trick, single pile) and from
// the lightning commands of this turn and of the card itself (3!, STOP, Skull).
type PlacementConstraint interface {
	Check(g *Game, p Placement) error // nil = allowed; else a typed command sentinel (errors.go)
}

// basePlacement applies the base direction rules and the backwards trick.
// The backwards trick needs a played card on top: it is impossible on a start card.
func basePlacement(r Ruleset, pile Pile, c Card) (PlacementKind, bool) {
	top := pile.TopValue()
	_, hasTop := pile.TopCard()
	switch pile.Direction {
	case DirectionUp:
		if c.Value > top {
			return PlacementNormal, true
		}
		if hasTop && c.Value == top-r.BackwardsTrickStep {
			return PlacementBackwardsTrick, true
		}
	case DirectionDown:
		if c.Value < top {
			return PlacementNormal, true
		}
		if hasTop && c.Value == top+r.BackwardsTrickStep {
			return PlacementBackwardsTrick, true
		}
	}
	return 0, false
}

// noBackwardsTrickRule vetoes the backwards trick while a No backwards trick card is
// visible. This includes a trick that would cover that card itself (Q2). The state is
// the one before the placement, so an NBT card may itself be placed via the trick (Q3).
type noBackwardsTrickRule struct{ cs commandState }

func (r noBackwardsTrickRule) Check(_ *Game, p Placement) error {
	if r.cs.noBackwardsTrick && p.Kind == PlacementBackwardsTrick {
		return fmt.Errorf("%w: %d on pile %d would use the backwards trick", ErrBackwardsTrickBlocked, p.Card.Value, p.PileIndex)
	}
	return nil
}

// singlePileRule vetoes a card that does not go on the bound pile while Single pile is active.
type singlePileRule struct{ cs commandState }

func (r singlePileRule) Check(_ *Game, p Placement) error {
	if r.cs.singlePile && r.cs.boundPile >= 0 && p.PileIndex != r.cs.boundPile {
		return fmt.Errorf("%w: all cards this turn go on pile %d", ErrSinglePileViolation, r.cs.boundPile)
	}
	return nil
}

// exactCountRule enforces 3!: no card after the third once a 3! was played, no 3! as the
// 4th or later card, and no 3! when the turn started with fewer than 3 cards.
type exactCountRule struct{ cs commandState }

func (r exactCountRule) Check(g *Game, p Placement) error {
	played := len(g.Turn.Plays)
	if r.cs.exactCount > 0 && played >= r.cs.exactCount {
		return fmt.Errorf("%w: 3! allows exactly %d cards this turn", ErrExactCountExceeded, r.cs.exactCount)
	}
	if p.Card.Command == CommandExactlyThree {
		if played+1 > exactlyThreeCount {
			return fmt.Errorf("%w: 3! would be card %d of this turn", ErrExactCountExceeded, played+1)
		}
		if g.Turn.StartHandSize < exactlyThreeCount {
			return fmt.Errorf("%w: the turn started with %d cards", ErrExactCountUnreachable, g.Turn.StartHandSize)
		}
	}
	return nil
}

// afterStopRule vetoes every card once a STOP was played this turn: the player may only
// end the turn (or undo the STOP).
type afterStopRule struct{ cs commandState }

func (r afterStopRule) Check(_ *Game, _ Placement) error {
	if r.cs.stopPlayed {
		return fmt.Errorf("%w: no card may follow a STOP", ErrStopPlayed)
	}
	return nil
}

// stopRule allows a STOP only if the turn could legally end right after it, because no
// card may follow a STOP (single-card look via withPlay, not a search).
type stopRule struct{}

func (stopRule) Check(g *Game, p Placement) error {
	if p.Card.Command != CommandStop {
		return nil
	}
	err := g.withPlay(p).runEndTurnChecks()
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrExactCountRequired):
		return fmt.Errorf("%w: STOP must be the third card after 3!", ErrStopTooEarly)
	default:
		return fmt.Errorf("no card may follow STOP: %w", err)
	}
}

// skullConflictRule (Q5) vetoes any card after which no further card is allowed
// (3! reached) while a skull from this turn stays uncovered: it could never be covered.
// This covers a Skull or 3! as the 3rd card, and a plain 3rd card that leaves a skull showing.
type skullConflictRule struct{}

func (skullConflictRule) Check(g *Game, p Placement) error {
	if len(g.Turn.Plays)+1 < exactlyThreeCount {
		return nil // 3! cannot be reached by this card: skip building the after-state
	}
	after := g.withPlay(p)
	cs := after.commands()
	if cs.exactCount > 0 && len(after.Turn.Plays) >= cs.exactCount && len(cs.skullPiles) > 0 {
		return fmt.Errorf("%w: after 3! no further card may cover the skull on pile %d", ErrSkullUncoverable, cs.skullPiles[0])
	}
	return nil
}

// placementConstraints returns the constraints derived from the current command state.
// In classic games none of them ever vetoes (no card carries a command).
func (g *Game) placementConstraints() []PlacementConstraint {
	cs := g.commands()
	return []PlacementConstraint{
		afterStopRule{cs},  // pile-independent
		exactCountRule{cs}, // count first: pile-independent
		singlePileRule{cs},
		noBackwardsTrickRule{cs},
		stopRule{},          // needs the after-state
		skullConflictRule{}, // needs the after-state
	}
}

// withPlay returns a read-only projection of g after placing p (card out of the active hand,
// onto the pile, appended to Turn.Plays). It shares unchanged slices and maps with g:
// never mutate it, emit on it or commit it.
func (g *Game) withPlay(p Placement) *Game {
	after := *g
	after.events = nil
	after.Piles = slices.Clone(g.Piles)
	after.Piles[p.PileIndex].Cards = append(slices.Clip(g.Piles[p.PileIndex].Cards), p.Card)
	after.Turn.Plays = append(slices.Clip(g.Turn.Plays), Play{Card: p.Card, Pile: p.PileIndex})
	seat := g.Turn.ActiveSeat
	after.Players = slices.Clone(g.Players)
	after.Players[seat].Hand = slices.DeleteFunc(slices.Clone(g.Players[seat].Hand),
		func(c Card) bool { return c.Value == p.Card.Value })
	return &after
}

// basePlacementAt validates the pile index and applies the base rule.
// Errors: ErrInvalidPile, ErrIllegalPlacement.
func (g *Game) basePlacementAt(c Card, pileIdx int) (Placement, error) {
	if pileIdx < 0 || pileIdx >= len(g.Piles) {
		return Placement{}, fmt.Errorf("%w: %d", ErrInvalidPile, pileIdx)
	}
	pile := g.Piles[pileIdx]
	kind, ok := basePlacement(g.Rules, pile, c)
	if !ok {
		return Placement{}, fmt.Errorf("%w: %d on %s pile %d showing %d", ErrIllegalPlacement, c.Value, pile.Direction, pileIdx, pile.TopValue())
	}
	return Placement{Card: c, PileIndex: pileIdx, Pile: pile, Kind: kind}, nil
}

// checkConstraints returns the first veto of cons for p, or nil.
func (g *Game) checkConstraints(cons []PlacementConstraint, p Placement) error {
	for _, con := range cons {
		if err := con.Check(g, p); err != nil {
			return err
		}
	}
	return nil
}

// checkPlacement runs the base rule, then every constraint. It does not mutate g.
func (g *Game) checkPlacement(c Card, pileIdx int) (Placement, error) {
	p, err := g.basePlacementAt(c, pileIdx)
	if err != nil {
		return Placement{}, err
	}
	if err := g.checkConstraints(g.placementConstraints(), p); err != nil {
		return Placement{}, err
	}
	return p, nil
}
