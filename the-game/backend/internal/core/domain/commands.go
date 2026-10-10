package domain

import "slices"

// Command parameters printed on the cards (not table config, so not in Ruleset).
const (
	exactlyThreeCount = 3 // 3!
	drawOnlyOneCount  = 1 // Draw only one
)

// ActiveCommand is a visible Infinity command and the pile it lies on (view type).
type ActiveCommand struct {
	Command Command
	Pile    int
}

// commandState is everything the Extreme commands imply right now.
// It is derived from Piles and Turn.Plays and never stored, so covering a card
// takes effect for the very next action. Classic games always derive the zero
// state (with boundPile -1): every command rule is a no-op there.
type commandState struct {
	active           []ActiveCommand // visible Infinity commands, pile order; nil if none
	noBackwardsTrick bool
	singlePile       bool
	boundPile        int // -1 = not bound (no Single pile, or the next card picks)
	drawOnlyOne      bool
	noCommunication  bool
	exactCount       int   // exactlyThreeCount once a 3! was played this turn, else 0
	stopPlayed       bool  // a STOP was played this turn
	skullPiles       []int // piles whose top card is a Skull played this turn; nil if none
}

// commands derives the command state of the current table and turn.
func (g *Game) commands() commandState { return deriveCommands(g.Piles, g.Turn.Plays) }

// deriveCommands derives the command state.
//   - Infinity: a command is in effect iff its card is the top card of a pile (any turn).
//   - Lightning: STOP and 3! count iff their card is in plays (this turn), covered or not.
//     A Skull counts iff it is on top and was played this turn. A lightning card on top
//     from an earlier turn means nothing.
func deriveCommands(piles []Pile, plays []Play) commandState {
	cs := commandState{boundPile: -1}
	for i, p := range piles {
		top, ok := p.TopCard()
		if !ok {
			continue
		}
		if top.Command.IsInfinity() {
			cs.active = append(cs.active, ActiveCommand{Command: top.Command, Pile: i})
		}
		switch top.Command {
		case CommandNoBackwardsTrick:
			cs.noBackwardsTrick = true
		case CommandSinglePile:
			cs.singlePile = true
		case CommandDrawOnlyOne:
			cs.drawOnlyOne = true
		case CommandNoCommunication:
			cs.noCommunication = true
		case CommandSkull:
			if playedThisTurn(plays, i) {
				cs.skullPiles = append(cs.skullPiles, i)
			}
		}
	}
	for _, p := range plays {
		switch p.Card.Command {
		case CommandStop:
			cs.stopPlayed = true
		case CommandExactlyThree:
			cs.exactCount = exactlyThreeCount
		}
	}
	if cs.singlePile {
		cs.boundPile = singlePileBinding(plays)
	}
	return cs
}

// singlePileBinding returns the pile every further card of this turn must go on while a
// Single pile card is visible, or -1 if the next card picks. Only call it when Single pile
// is active. Covering a Single pile card ends its binding at once (rules.md §2, Q1).
func singlePileBinding(plays []Play) int {
	// 1. The latest Single pile card played this turn that is still on top binds to its own pile.
	for i := len(plays) - 1; i >= 0; i-- {
		p := plays[i]
		if p.Card.Command != CommandSinglePile {
			continue
		}
		stillTop := !slices.ContainsFunc(plays[i+1:], func(q Play) bool { return q.Pile == p.Pile })
		if stillTop {
			return p.Pile
		}
	}
	// 2. Otherwise the visible Single pile card(s) were visible since the turn started:
	//    the first card picked the pile, and every play since went there.
	if len(plays) > 0 {
		return plays[0].Pile
	}
	// 3. Active but unbound: the next card picks.
	return -1
}

// playedThisTurn reports whether some play this turn went on pile, i.e. its top card
// was played this turn. It uses pile membership, not card values, so it stays correct
// for future duplicate-value card kinds.
func playedThisTurn(plays []Play, pile int) bool {
	return slices.ContainsFunc(plays, func(p Play) bool { return p.Pile == pile })
}
