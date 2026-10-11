package domain

import "slices"

// End is one end of the forest path. Seat 0 faces the left end, seat 1 the right end.
type End string

const (
	EndNone  End = ""
	EndLeft  End = "left"
	EndRight End = "right"
)

// endOf returns the end the player in seat faces.
func endOf(seat int) End {
	if seat == 0 {
		return EndLeft
	}
	return EndRight
}

// direction returns the index step toward end: -1 for the left end, +1 for the right end.
func direction(e End) int {
	if e == EndLeft {
		return -1
	}
	return 1
}

// otherEnd returns the opposite end.
func otherEnd(e End) End {
	if e == EndLeft {
		return EndRight
	}
	return EndLeft
}

// BoardSide is a side of the forest board.
type BoardSide string

const (
	SideA BoardSide = "A"
	SideB BoardSide = "B"
)

// Space is one path space with its location, as printed on the board (rules §3).
type Space struct {
	Squares  int  // hanging squares: starting gems on every level
	Diamonds int  // hanging diamonds: extra starting gems on level 3
	Plus     bool // gets a gem at the end of rounds 1 and 2
}

// boardDef is the static topology of one board side, spaces from the left end to the right end.
type boardDef struct {
	Spaces []Space
	Start  int // the circled start space
}

// boards holds the topology of both sides (rules §3).
var boards = map[BoardSide]boardDef{
	SideA: {
		Start: 5,
		Spaces: []Space{
			{Squares: 1},             // L5
			{Squares: 1, Plus: true}, // L4
			{Squares: 2},             // L3
			{Squares: 1, Plus: true}, // L2
			{Squares: 1},             // L1
			{Plus: true},             // C (start)
			{Squares: 1},             // R1
			{Squares: 1, Plus: true}, // R2
			{Squares: 2},             // R3
			{Squares: 1, Plus: true}, // R4
			{Squares: 1},             // R5
		},
	},
	SideB: {
		Start: 4,
		Spaces: []Space{
			{Squares: 1, Diamonds: 1}, // L4
			{Squares: 2, Plus: true},  // L3
			{Squares: 2, Plus: true},  // L2
			{Plus: true},              // L1
			{Squares: 2, Diamonds: 1}, // C (start)
			{Squares: 2, Plus: true},  // R1
			{Plus: true},              // R2
			{Squares: 2, Plus: true},  // R3
			{Squares: 1, Diamonds: 1}, // R4
			{Squares: 1},              // R5
		},
	},
}

// LevelDef is the setup of one difficulty level (rules §4).
type LevelDef struct {
	Level        int
	Side         BoardSide
	Diamonds     bool // diamonds count as starting gems
	ForestTokens int
	BaseScore    int
}

// Level limits.
const MinLevel, MaxLevel = 1, 3

// levels is the difficulty table keyed by level.
var levels = map[int]LevelDef{
	1: {Level: 1, Side: SideA, Diamonds: false, ForestTokens: 4, BaseScore: 10},
	2: {Level: 2, Side: SideB, Diamonds: false, ForestTokens: 3, BaseScore: 20},
	3: {Level: 3, Side: SideB, Diamonds: true, ForestTokens: 3, BaseScore: 30},
}

// LevelFor returns the setup of level; false outside MinLevel..MaxLevel.
func LevelFor(level int) (LevelDef, bool) {
	l, ok := levels[level]
	return l, ok
}

// Spaces returns a copy of the board topology for this level's side.
func (l LevelDef) Spaces() []Space { return slices.Clone(boards[l.Side].Spaces) }

// Start returns the start space of this level's side.
func (l LevelDef) Start() int { return boards[l.Side].Start }

// startingGems returns the gems per location at setup.
func (l LevelDef) startingGems() []int {
	spaces := boards[l.Side].Spaces
	out := make([]int, len(spaces))
	for i, s := range spaces {
		out[i] = s.Squares
		if l.Diamonds {
			out[i] += s.Diamonds
		}
	}
	return out
}

// Board is the mutable state of the forest. Positions are indices from the left end.
type Board struct {
	Gems         []int // gems on the location of each path space
	CoveredLeft  int   // forest tokens covering spaces from the left end inward
	CoveredRight int   // forest tokens covering spaces from the right end inward
	Traveller    int   // the space the traveller stands on
	ForestTokens int   // forest tokens left to place
}

// newBoard returns the board at setup for level l.
func newBoard(l LevelDef) Board {
	return Board{Gems: l.startingGems(), Traveller: l.Start(), ForestTokens: l.ForestTokens}
}

// clone returns a deep copy.
func (b Board) clone() Board {
	b.Gems = slices.Clone(b.Gems)
	return b
}

// lastUncovered returns the outermost uncovered space at end e.
func (b *Board) lastUncovered(e End) int {
	if e == EndLeft {
		return b.CoveredLeft
	}
	return len(b.Gems) - 1 - b.CoveredRight
}

// Covered reports whether space i is under a forest token.
func (b *Board) Covered(i int) bool {
	return i < b.lastUncovered(EndLeft) || i > b.lastUncovered(EndRight)
}

// GemsLeft returns the gems on the board.
func (b *Board) GemsLeft() int {
	n := 0
	for _, g := range b.Gems {
		n += g
	}
	return n
}

// beyond returns the end the traveller would leave the path at when moving by delta from
// its space, or EndNone if it stays on the path.
func (b *Board) beyond(delta int) End {
	to := b.Traveller + delta
	switch {
	case to < b.lastUncovered(EndLeft):
		return EndLeft
	case to > b.lastUncovered(EndRight):
		return EndRight
	}
	return EndNone
}

// cover places a forest token on the last uncovered space of end e and moves the gems of its
// location one location toward the centre. It returns the covered space and the gems moved.
// The caller checks that a token is left.
func (b *Board) cover(e End) (space, moved int) {
	space = b.lastUncovered(e)
	inward := space - direction(e)
	moved = b.Gems[space]
	b.Gems[inward] += moved
	b.Gems[space] = 0
	if e == EndLeft {
		b.CoveredLeft++
	} else {
		b.CoveredRight++
	}
	b.ForestTokens--
	return space, moved
}

// addRoundGems adds one gem per plus location (rules §7). A covered plus location gives its gem
// to the outermost uncovered location of its end (app decision). Returns the gems added.
func (b *Board) addRoundGems(spaces []Space) int {
	n := 0
	for i, s := range spaces {
		if !s.Plus {
			continue
		}
		at := i
		switch {
		case i < b.lastUncovered(EndLeft):
			at = b.lastUncovered(EndLeft)
		case i > b.lastUncovered(EndRight):
			at = b.lastUncovered(EndRight)
		}
		b.Gems[at]++
		n++
	}
	return n
}

// coverableEnds returns the ends a round-end forest token may go to: both, unless the traveller
// stands on the last uncovered space of one end (rules §7). Empty when no token is left.
func (b *Board) coverableEnds() []End {
	if b.ForestTokens == 0 {
		return nil
	}
	switch b.Traveller {
	case b.lastUncovered(EndLeft):
		return []End{EndRight}
	case b.lastUncovered(EndRight):
		return []End{EndLeft}
	}
	return []End{EndLeft, EndRight}
}
