package domain

import "fmt"

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

// PlacementConstraint can veto a placement the base rule allowed.
// Phase 2 adds constraints derived from visible Infinity commands (no backwards trick, single pile)
// and from lightning commands played this turn (3! count, STOP).
type PlacementConstraint interface {
	Check(g *Game, p Placement) error // nil = allowed; else wrap ErrIllegalPlacement (or a future specific error)
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

// placementConstraints returns the active constraints. Phase 1: none.
func (g *Game) placementConstraints() []PlacementConstraint {
	return nil
}

// checkPlacement runs the base rule, then every constraint. It does not mutate g.
func (g *Game) checkPlacement(c Card, pileIdx int) (Placement, error) {
	if pileIdx < 0 || pileIdx >= len(g.Piles) {
		return Placement{}, fmt.Errorf("%w: %d", ErrInvalidPile, pileIdx)
	}
	pile := g.Piles[pileIdx]
	kind, ok := basePlacement(g.Rules, pile, c)
	if !ok {
		return Placement{}, fmt.Errorf("%w: %d on %s pile %d showing %d", ErrIllegalPlacement, c.Value, pile.Direction, pileIdx, pile.TopValue())
	}
	p := Placement{Card: c, PileIndex: pileIdx, Pile: pile, Kind: kind}
	for _, con := range g.placementConstraints() {
		if err := con.Check(g, p); err != nil {
			return Placement{}, err
		}
	}
	return p, nil
}
