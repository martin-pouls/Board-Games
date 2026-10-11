package domain

import (
	"slices"
	"testing"
)

func TestLevels(t *testing.T) {
	tests := []struct {
		level, spaces, start, gems, pluses, tokens, base int
		side                                             BoardSide
	}{
		{level: 1, side: SideA, spaces: 11, start: 5, gems: 12, pluses: 5, tokens: 4, base: 10},
		{level: 2, side: SideB, spaces: 10, start: 4, gems: 13, pluses: 6, tokens: 3, base: 20},
		{level: 3, side: SideB, spaces: 10, start: 4, gems: 16, pluses: 6, tokens: 3, base: 30},
	}
	for _, tt := range tests {
		l, ok := LevelFor(tt.level)
		if !ok {
			t.Fatalf("level %d missing", tt.level)
		}
		b := newBoard(l)
		pluses := 0
		for _, s := range l.Spaces() {
			if s.Plus {
				pluses++
			}
		}
		if l.Side != tt.side || len(b.Gems) != tt.spaces || l.Start() != tt.start || b.Traveller != tt.start ||
			b.GemsLeft() != tt.gems || pluses != tt.pluses || b.ForestTokens != tt.tokens || l.BaseScore != tt.base {
			t.Errorf("level %d: side %s, %d spaces, start %d, traveller %d, %d gems, %d pluses, %d tokens, base %d",
				tt.level, l.Side, len(b.Gems), l.Start(), b.Traveller, b.GemsLeft(), pluses, b.ForestTokens, l.BaseScore)
		}
	}
	for _, bad := range []int{0, 4, -1} {
		if _, ok := LevelFor(bad); ok {
			t.Errorf("LevelFor(%d) ok", bad)
		}
	}
}

func TestStartingGems(t *testing.T) {
	a, _ := LevelFor(1)
	b2, _ := LevelFor(2)
	b3, _ := LevelFor(3)
	tests := []struct {
		name string
		got  []int
		want []int
	}{
		{"side A", a.startingGems(), []int{1, 1, 2, 1, 1, 0, 1, 1, 2, 1, 1}},
		{"side B squares", b2.startingGems(), []int{1, 2, 2, 0, 2, 2, 0, 2, 1, 1}},
		{"side B squares and diamonds", b3.startingGems(), []int{2, 2, 2, 0, 3, 2, 0, 2, 2, 1}},
	}
	for _, tt := range tests {
		if !slices.Equal(tt.got, tt.want) {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

func TestStartIsNeverCovered(t *testing.T) {
	// Each end has more spaces than there are forest tokens (rules §7).
	for level := MinLevel; level <= MaxLevel; level++ {
		l, _ := LevelFor(level)
		n := len(l.Spaces())
		if l.Start() < l.ForestTokens+1 || n-1-l.Start() < l.ForestTokens+1 {
			t.Errorf("level %d: start %d of %d spaces can be covered with %d tokens", level, l.Start(), n, l.ForestTokens)
		}
	}
}

func TestCover(t *testing.T) {
	tests := []struct {
		name               string
		gems               []int
		coveredL, coveredR int
		end                End
		wantSpace, moved   int
		wantGems           []int
	}{
		{"left end", []int{2, 1, 0, 0, 0}, 0, 0, EndLeft, 0, 2, []int{0, 3, 0, 0, 0}},
		{"right end", []int{0, 0, 0, 1, 1}, 0, 0, EndRight, 4, 1, []int{0, 0, 0, 2, 0}},
		{"left end, already covered once", []int{0, 1, 0, 0, 0}, 1, 0, EndLeft, 1, 1, []int{0, 0, 1, 0, 0}},
		{"right end, no gems", []int{0, 0, 0, 1, 0}, 0, 1, EndRight, 3, 1, []int{0, 0, 1, 0, 0}},
		{"empty location", []int{0, 1, 0, 0, 0}, 0, 0, EndLeft, 0, 0, []int{0, 1, 0, 0, 0}},
	}
	for _, tt := range tests {
		b := Board{Gems: slices.Clone(tt.gems), CoveredLeft: tt.coveredL, CoveredRight: tt.coveredR, Traveller: 2, ForestTokens: 2}
		space, moved := b.cover(tt.end)
		if space != tt.wantSpace || moved != tt.moved || !slices.Equal(b.Gems, tt.wantGems) || b.ForestTokens != 1 {
			t.Errorf("%s: space %d moved %d gems %v tokens %d", tt.name, space, moved, b.Gems, b.ForestTokens)
		}
		if !b.Covered(space) {
			t.Errorf("%s: space %d not covered", tt.name, space)
		}
	}
}

func TestAddRoundGems(t *testing.T) {
	a, _ := LevelFor(1) // pluses at 1, 3, 5, 7, 9
	tests := []struct {
		name               string
		coveredL, coveredR int
		want               []int
	}{
		{"nothing covered", 0, 0, []int{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0}},
		// App decision: a covered plus location gives its gem to the outermost uncovered location.
		{"left plus covered", 2, 0, []int{0, 0, 1, 1, 0, 1, 0, 1, 0, 1, 0}},
		{"two left pluses covered", 4, 0, []int{0, 0, 0, 0, 2, 1, 0, 1, 0, 1, 0}},
		{"right plus covered", 0, 3, []int{0, 1, 0, 1, 0, 1, 0, 2, 0, 0, 0}},
	}
	for _, tt := range tests {
		b := Board{Gems: make([]int, 11), CoveredLeft: tt.coveredL, CoveredRight: tt.coveredR}
		if n := b.addRoundGems(a.Spaces()); n != 5 {
			t.Errorf("%s: added %d, want 5", tt.name, n)
		}
		if !slices.Equal(b.Gems, tt.want) {
			t.Errorf("%s: gems %v, want %v", tt.name, b.Gems, tt.want)
		}
	}
}

func TestCoverableEnds(t *testing.T) {
	tests := []struct {
		name                 string
		traveller, cl, cr, n int
		want                 []End
	}{
		{"middle", 5, 0, 0, 1, []End{EndLeft, EndRight}},
		{"on the left end", 0, 0, 0, 1, []End{EndRight}},
		{"on the right end", 10, 0, 0, 1, []End{EndLeft}},
		{"on the left end after covers", 2, 2, 0, 1, []End{EndRight}},
		{"on the right end after covers", 9, 0, 1, 1, []End{EndLeft}},
		{"next to the end", 1, 0, 0, 1, []End{EndLeft, EndRight}},
		{"no token left", 5, 0, 0, 0, nil},
	}
	for _, tt := range tests {
		b := Board{Gems: make([]int, 11), Traveller: tt.traveller, CoveredLeft: tt.cl, CoveredRight: tt.cr, ForestTokens: tt.n}
		if got := b.coverableEnds(); !slices.Equal(got, tt.want) {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestBeyond(t *testing.T) {
	b := Board{Gems: make([]int, 11), Traveller: 5, CoveredLeft: 1, CoveredRight: 2}
	tests := []struct {
		delta int
		want  End
	}{
		{0, EndNone}, {-4, EndNone}, {-5, EndLeft}, {3, EndNone}, {4, EndRight}, {-6, EndLeft},
	}
	for _, tt := range tests {
		if got := b.beyond(tt.delta); got != tt.want {
			t.Errorf("beyond(%d) = %q, want %q", tt.delta, got, tt.want)
		}
	}
}
