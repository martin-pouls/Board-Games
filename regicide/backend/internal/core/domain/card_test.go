package domain

import (
	"slices"
	"testing"
)

func TestCardValue(t *testing.T) {
	tests := []struct {
		card Card
		want int
	}{
		{Jester(), 0},
		{c("AH"), 1},
		{c("2D"), 2},
		{c("7C"), 7},
		{c("10S"), 10},
		{c("JH"), 10},
		{c("QD"), 15},
		{c("KC"), 20},
	}
	for _, tt := range tests {
		if got := tt.card.Value(); got != tt.want {
			t.Errorf("%v.Value() = %d, want %d", tt.card, got, tt.want)
		}
	}
}

func TestCardKinds(t *testing.T) {
	tests := []struct {
		card                                 Card
		jester, companion, number, royal, ok bool
	}{
		{Jester(), true, false, false, false, true},
		{c("AS"), false, true, false, false, true},
		{c("2S"), false, false, true, false, true},
		{c("10H"), false, false, true, false, true},
		{c("JC"), false, false, false, true, true},
		{c("KD"), false, false, false, true, true},
		{NewCard(SuitNone, 5), false, false, true, false, false},
		{NewCard("stars", 5), false, false, true, false, false},
		{NewCard(SuitHearts, 0), false, false, false, false, false},
		{NewCard(SuitHearts, 14), false, false, false, false, false},
	}
	for _, tt := range tests {
		k := tt.card
		if k.IsJester() != tt.jester || k.IsCompanion() != tt.companion || k.IsNumber() != tt.number ||
			k.IsRoyal() != tt.royal || k.Valid() != tt.ok {
			t.Errorf("%v: jester %v companion %v number %v royal %v valid %v", k,
				k.IsJester(), k.IsCompanion(), k.IsNumber(), k.IsRoyal(), k.Valid())
		}
	}
}

func TestCardString(t *testing.T) {
	for in, want := range map[string]string{"5H": "5♥", "AS": "A♠", "10D": "10♦", "KC": "K♣", "QH": "Q♥", "JS": "J♠", "X": "Jester"} {
		if got := c(in).String(); got != want {
			t.Errorf("%s.String() = %q, want %q", in, got, want)
		}
	}
	if got := NewCard("stars", 3).String(); got != `Card("stars",3)` {
		t.Errorf("invalid card String() = %q", got)
	}
}

func TestEnemyTable(t *testing.T) {
	tests := []struct {
		r              Rank
		attack, health int
	}{
		{RankJack, 10, 20},
		{RankQueen, 15, 30},
		{RankKing, 20, 40},
		{RankAce, 0, 0},
		{5, 0, 0},
		{RankJester, 0, 0},
	}
	for _, tt := range tests {
		if EnemyAttack(tt.r) != tt.attack || EnemyHealth(tt.r) != tt.health {
			t.Errorf("rank %d: attack %d health %d", tt.r, EnemyAttack(tt.r), EnemyHealth(tt.r))
		}
	}
}

func TestHandValue(t *testing.T) {
	if got := HandValue(cs("AH 5C KS QD JH X")); got != 1+5+20+15+10 {
		t.Fatalf("HandValue = %d", got)
	}
	if HandValue(nil) != 0 {
		t.Fatal("HandValue(nil) != 0")
	}
}

func TestSortCards(t *testing.T) {
	got := cs("KS 5S X 5H AD 10C JH QC 5C AH")
	SortCards(got)
	if want := cs("X AH AD 5H 5C 5S 10C JH QC KS"); !slices.Equal(got, want) {
		t.Fatalf("SortCards = %v, want %v", got, want)
	}
}

func TestRulesFor(t *testing.T) {
	tests := []struct {
		n                    int
		hand, tavernJ, soloJ int
		ok                   bool
	}{
		{0, 0, 0, 0, false},
		{1, 8, 0, 2, true},
		{2, 7, 0, 0, true},
		{3, 6, 1, 0, true},
		{4, 5, 2, 0, true},
		{5, 0, 0, 0, false},
	}
	for _, tt := range tests {
		r, ok := RulesFor(tt.n)
		if ok != tt.ok || r.MaxHandSize != tt.hand || r.TavernJesters != tt.tavernJ || r.SoloJesters != tt.soloJ {
			t.Errorf("RulesFor(%d) = %+v, %v", tt.n, r, ok)
		}
	}
}

func TestNewTavern(t *testing.T) {
	for jesters := range 3 {
		tv := NewTavern(jesters)
		if len(tv) != 40+jesters {
			t.Fatalf("NewTavern(%d): %d cards", jesters, len(tv))
		}
		if tv[0] != c("AH") || tv[9] != c("10H") || tv[10] != c("AD") || tv[39] != c("10S") {
			t.Fatalf("NewTavern(%d) order: %v", jesters, tv[:40])
		}
		seen := map[Card]bool{}
		for i, x := range tv {
			if i >= 40 {
				if !x.IsJester() {
					t.Fatalf("card %d = %v, want Jester", i, x)
				}
				continue
			}
			if !x.Valid() || x.IsRoyal() || x.IsJester() || seen[x] {
				t.Fatalf("card %d = %v", i, x)
			}
			seen[x] = true
		}
	}
}

func TestNewCastle(t *testing.T) {
	got := NewCastle(noShuffle{})
	want := cs("KH KD KC KS QH QD QC QS JH JD JC JS")
	if !slices.Equal(got, want) {
		t.Fatalf("NewCastle(noShuffle) = %v", got)
	}
	// Each tier is shuffled separately and stays grouped.
	if got := NewCastle(reverseShuffle{}); !slices.Equal(got, cs("KS KC KD KH QS QC QD QH JS JC JD JH")) {
		t.Fatalf("NewCastle(reverse) = %v", got)
	}
	// Shuffle order: Jacks, Queens, Kings (reverse only the first call).
	rng := &countingRNG{reverse: map[int]bool{0: true}}
	got = NewCastle(rng)
	if !slices.Equal(got, cs("KH KD KC KS QH QD QC QS JS JC JD JH")) || !slices.Equal(rng.calls, []int{4, 4, 4}) {
		t.Fatalf("NewCastle(first call reversed) = %v, calls %v", got, rng.calls)
	}
}
