package domain

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// This table was transcribed independently from deep-sea/rules.md section 9 (and the encoding
// conventions agreed for it) without looking at tasks_deepsea.go. TestDeepSeaTaskCardTable
// compares every field of every card.

// dsDiff returns one line per leaf field that differs between want and got, named by its path.
// Nil and empty slices are equal.
func dsDiff(path string, want, got reflect.Value) []string {
	switch want.Kind() {
	case reflect.Struct:
		var out []string
		for i := 0; i < want.NumField(); i++ {
			out = append(out, dsDiff(path+"."+want.Type().Field(i).Name, want.Field(i), got.Field(i))...)
		}
		return out
	case reflect.Slice:
		if want.Len() != got.Len() {
			return []string{fmt.Sprintf("%s: expected %v, got %v", path, want.Interface(), got.Interface())}
		}
		var out []string
		for i := 0; i < want.Len(); i++ {
			out = append(out, dsDiff(fmt.Sprintf("%s[%d]", path, i), want.Index(i), got.Index(i))...)
		}
		return out
	case reflect.Array:
		var out []string
		for i := 0; i < want.Len(); i++ {
			out = append(out, dsDiff(fmt.Sprintf("%s[%d]", path, i), want.Index(i), got.Index(i))...)
		}
		return out
	}
	if !reflect.DeepEqual(want.Interface(), got.Interface()) {
		return []string{fmt.Sprintf("%s: expected %v, got %v", path, want.Interface(), got.Interface())}
	}
	return nil
}

func dsSub(n Rank) Card { return Card{SuitTrump, n} }

// dsCard builds a card with difficulty a/b/c.
func dsCard(id, a, b, c int, o Objective) TaskCard {
	return TaskCard{ID: id, Difficulty: [3]int{a, b, c}, Objective: o}
}

func dsWithCountBy(t TaskCard, by [3]int) TaskCard { t.CountBy = by; return t }

func dsWithRedeal(t TaskCard, r ...[]Card) TaskCard { t.Redeal = r; return t }

func dsWin(cards ...Card) Objective { return Objective{Kind: KindWinCards, Cards: cards} }

func dsTricks(kind ObjectiveKind, tricks ...int) Objective {
	return Objective{Kind: kind, Tricks: tricks}
}

func dsRankCount(cmp Comparison, n int, ranks ...Rank) Objective {
	return Objective{Kind: KindWinCount, Ranks: ranks, Cmp: cmp, Count: n}
}

func dsSuitCount(cmp Comparison, n int, suits ...Suit) Objective {
	return Objective{Kind: KindWinCount, Suits: suits, Cmp: cmp, Count: n}
}

func dsTrump(n int) Objective { return dsSuitCount(CmpExactly, n, SuitTrump) }

// dsTrumpIncl is "exactly one submarine, and it is submarine k".
func dsTrumpIncl(k Rank) Objective {
	return Objective{Kind: KindAll, Parts: []Objective{dsTrump(1), dsWin(dsSub(k))}}
}

func dsWinWith(with, target CardMatch) Objective {
	return Objective{Kind: KindWinWith, With: with, Target: target}
}

func dsRankMatch(r Rank) CardMatch { return CardMatch{Rank: r} }

func dsWhere(w TrickCondition, count int) Objective {
	return Objective{Kind: KindWinTrickWhere, Where: w, Count: count}
}

func dsAvoid(k int) Objective {
	t := make([]int, k)
	for i := range t {
		t[i] = i + 1
	}
	return Objective{Kind: KindAvoidTricks, Tricks: t}
}

var dsSubs1to4 = []Card{dsSub(1), dsSub(2), dsSub(3), dsSub(4)}

// expectedDeepSeaTasks is in the order of rules.md section 9, not by number; the test indexes it.
var expectedDeepSeaTasks = []TaskCard{
	// Trick counts and comparisons.
	dsCard(1, 2, 3, 3, Objective{Kind: KindCmpTricks, Cmp: CmpMore}),
	dsCard(2, 3, 4, 5, Objective{Kind: KindCmpTricks, Cmp: CmpMoreThanAll}),
	dsCard(3, 2, 2, 3, Objective{Kind: KindCmpTricks, Cmp: CmpFewer}),
	dsCard(4, 2, 2, 3, Objective{Kind: KindCmpCaptain, Cmp: CmpMore}),
	dsCard(5, 2, 2, 2, Objective{Kind: KindCmpCaptain, Cmp: CmpFewer}),
	dsCard(6, 4, 3, 3, Objective{Kind: KindCmpCaptain, Cmp: CmpEqual}),
	dsCard(74, 4, 3, 3, Objective{Kind: KindWinNoTricks}),
	dsCard(83, 3, 2, 2, Objective{Kind: KindWinExactlyTricks, Count: 1}),
	dsCard(84, 2, 2, 2, Objective{Kind: KindWinExactlyTricks, Count: 2}),
	dsCard(87, 2, 3, 5, Objective{Kind: KindWinExactlyTricks, Count: 4}),
	dsCard(90, 3, 2, 2, Objective{Kind: KindPredict}),
	dsCard(91, 4, 3, 3, Objective{Kind: KindPredict, Hidden: true}),

	// Specific tricks and sequences.
	dsCard(71, 1, 2, 3, dsAvoid(4)),
	dsCard(72, 1, 2, 2, dsAvoid(3)),
	dsCard(73, 2, 3, 3, dsAvoid(5)),
	dsCard(75, 3, 2, 2, Objective{Kind: KindNoConsecutive}),
	dsCard(76, 2, 3, 3, dsTricks(KindWinTricks, LastTrick)),
	dsCard(77, 2, 3, 4, dsTricks(KindWinTricks, 1, 2, 3)),
	dsCard(78, 1, 1, 2, dsTricks(KindWinTricks, 1, 2)),
	dsCard(79, 1, 1, 1, dsTricks(KindWinTricks, 1)),
	dsCard(80, 3, 4, 4, dsTricks(KindWinTricks, 1, LastTrick)),
	dsCard(81, 4, 4, 4, dsTricks(KindWinOnlyTricks, LastTrick)),
	dsCard(82, 4, 3, 3, dsTricks(KindWinOnlyTricks, 1)),
	dsCard(85, 1, 1, 1, Objective{Kind: KindConsecutive, Count: 2}),
	dsCard(86, 2, 3, 4, Objective{Kind: KindConsecutive, Count: 3}),
	dsCard(88, 3, 3, 4, Objective{Kind: KindConsecutive, Count: 3, Exact: true}),
	dsCard(89, 3, 3, 3, Objective{Kind: KindConsecutive, Count: 2, Exact: true}),

	// Winning a trick in a certain way.
	dsCard(9, 2, 3, 3, dsWinWith(dsRankMatch(6), CardMatch{})),
	dsCard(10, 2, 3, 4, dsWinWith(dsRankMatch(5), CardMatch{})),
	dsCard(11, 3, 4, 5, dsWinWith(dsRankMatch(3), CardMatch{})),
	dsCard(15, 3, 4, 5, dsWinWith(dsRankMatch(2), CardMatch{})),
	dsCard(12, 1, 2, 2, dsWinWith(dsRankMatch(7), dsRankMatch(5))),
	dsCard(13, 3, 4, 5, dsWinWith(dsRankMatch(4), dsRankMatch(8))),
	dsCard(14, 2, 3, 4, dsWinWith(dsRankMatch(6), dsRankMatch(6))),
	dsCard(58, 3, 3, 3, dsWinWith(CardMatch{Suit: SuitTrump}, CardMatch{Suit: SuitPink, Rank: 7})),
	dsCard(59, 3, 3, 3, dsWinWith(CardMatch{Suit: SuitTrump}, CardMatch{Suit: SuitGreen, Rank: 9})),

	// Trick values.
	dsCard(7, 2, 3, 3, dsWhere(WhereBelow, 7)),
	dsCard(8, 2, 3, 4, dsWhere(WhereAbove, 5)),
	dsCard(46, 2, 5, 6, dsWhere(WhereEven, 0)),
	dsCard(47, 2, 4, 5, dsWhere(WhereOdd, 0)),
	dsWithCountBy(dsCard(48, 3, 3, 4, dsWhere(WhereSumAtLeast, 0)), [3]int{23, 28, 31}),
	dsWithCountBy(dsCard(49, 3, 3, 4, dsWhere(WhereSumBelow, 0)), [3]int{8, 12, 16}),
	dsCard(50, 3, 3, 4, Objective{Kind: KindWinTrickWhere, Where: WhereSumBetween, Count: 22, Max: 23}),

	// Specific cards.
	dsCard(16, 1, 1, 1, dsWin(Card{SuitPink, 3})),
	dsCard(17, 1, 1, 1, dsWin(Card{SuitYellow, 1})),
	dsCard(18, 1, 1, 1, dsWin(Card{SuitBlue, 4})),
	dsCard(19, 1, 1, 1, dsWin(Card{SuitGreen, 6})),
	dsCard(27, 2, 3, 3, dsWin(Card{SuitBlue, 1}, Card{SuitBlue, 2}, Card{SuitBlue, 3})),
	dsCard(28, 2, 2, 3, dsWin(Card{SuitBlue, 6}, Card{SuitYellow, 7})),
	dsCard(29, 2, 2, 3, dsWin(Card{SuitPink, 5}, Card{SuitYellow, 6})),
	dsCard(30, 2, 2, 3, dsWin(Card{SuitGreen, 5}, Card{SuitBlue, 8})),
	dsCard(31, 2, 2, 3, dsWin(Card{SuitBlue, 5}, Card{SuitPink, 8})),
	dsCard(32, 2, 2, 3, dsWin(Card{SuitPink, 9}, Card{SuitYellow, 8})),
	dsCard(33, 2, 2, 2, dsWin(Card{SuitPink, 1}, Card{SuitGreen, 7})),
	dsCard(34, 2, 3, 3, dsWin(Card{SuitYellow, 9}, Card{SuitBlue, 7})),
	dsCard(35, 3, 4, 4, dsWin(Card{SuitGreen, 3}, Card{SuitYellow, 4}, Card{SuitYellow, 5})),
	dsCard(36, 3, 4, 5, Objective{Kind: KindAll, Parts: []Objective{
		dsWin(Card{SuitGreen, 2}),
		{Kind: KindCardInLastTrick, Cards: []Card{{SuitGreen, 2}}},
	}}),
	dsCard(54, 1, 1, 1, dsWin(dsSub(3))),

	// Counts of a kind.
	dsCard(20, 3, 4, 5, dsRankCount(CmpExactly, 4, 3)),
	dsCard(21, 3, 4, 5, dsRankCount(CmpAtLeast, 3, 5)),
	dsCard(22, 3, 4, 5, dsRankCount(CmpAtLeast, 3, 9)),
	dsCard(23, 2, 2, 2, dsRankCount(CmpAtLeast, 2, 7)),
	dsCard(24, 4, 5, 6, dsRankCount(CmpExactly, 4, 9)),
	dsCard(25, 3, 4, 4, dsRankCount(CmpExactly, 3, 6)),
	dsCard(26, 2, 3, 3, dsRankCount(CmpExactly, 2, 9)),
	dsCard(37, 4, 4, 4, Objective{Kind: KindAll, Parts: []Objective{
		dsSuitCount(CmpExactly, 1, SuitPink),
		dsSuitCount(CmpExactly, 1, SuitGreen),
	}}),
	dsCard(38, 3, 3, 3, dsSuitCount(CmpAtLeast, 7, SuitYellow)),
	dsCard(39, 2, 3, 3, dsSuitCount(CmpAtLeast, 5, SuitPink)),
	dsCard(40, 3, 4, 4, dsSuitCount(CmpExactly, 2, SuitGreen)),
	dsCard(41, 3, 4, 4, dsSuitCount(CmpExactly, 2, SuitBlue)),
	dsCard(42, 3, 3, 4, dsSuitCount(CmpExactly, 1, SuitPink)),
	dsCard(43, 2, 2, 2, dsSuitCount(CmpExactly, 0, SuitPink)),
	dsCard(62, 2, 2, 2, dsSuitCount(CmpExactly, 0, SuitGreen)),
	dsCard(63, 2, 2, 2, dsSuitCount(CmpExactly, 0, SuitYellow)),
	dsCard(64, 3, 3, 3, dsSuitCount(CmpExactly, 0, SuitPink, SuitBlue)),
	dsCard(65, 3, 3, 3, dsSuitCount(CmpExactly, 0, SuitYellow, SuitGreen)),
	dsCard(66, 3, 3, 2, dsRankCount(CmpExactly, 0, 8, 9)),
	dsCard(67, 1, 1, 1, dsRankCount(CmpExactly, 0, 9)),
	dsCard(68, 1, 2, 2, dsRankCount(CmpExactly, 0, 5)),
	dsCard(69, 2, 2, 2, dsRankCount(CmpExactly, 0, 1)),
	dsCard(70, 3, 3, 3, dsRankCount(CmpExactly, 0, 1, 2, 3)),
	dsCard(44, 2, 3, 4, Objective{Kind: KindWinEachColour}),
	dsCard(45, 3, 4, 5, Objective{Kind: KindWinAllOfAnyColour}),

	// Submarines.
	dsWithRedeal(dsCard(51, 3, 3, 3, dsTrump(1)), dsSubs1to4),
	dsWithRedeal(dsCard(52, 3, 3, 3, dsTrumpIncl(1)), []Card{dsSub(1), dsSub(4)}, []Card{dsSub(1), dsSub(2), dsSub(3)}),
	dsWithRedeal(dsCard(53, 3, 3, 3, dsTrumpIncl(2)), []Card{dsSub(2), dsSub(4)}, []Card{dsSub(1), dsSub(2), dsSub(3)}),
	dsWithRedeal(dsCard(55, 3, 3, 4, dsTrump(2)), []Card{dsSub(2), dsSub(3), dsSub(4)}),
	dsWithRedeal(dsCard(56, 3, 4, 4, dsTrump(3)), dsSubs1to4),
	dsCard(57, 1, 1, 1, dsTrump(0)),

	// Leading.
	dsCard(60, 4, 3, 3, Objective{Kind: KindNeverLead, Suits: []Suit{SuitPink, SuitYellow, SuitBlue}}),
	dsCard(61, 2, 1, 1, Objective{Kind: KindNeverLead, Suits: []Suit{SuitPink, SuitGreen}}),

	// Colour comparisons.
	dsCard(92, 4, 4, 4, Objective{Kind: KindEqualSuits, Suits: []Suit{SuitPink, SuitYellow}}),
	dsCard(93, 2, 3, 3, Objective{Kind: KindEqualSuits, Suits: []Suit{SuitGreen, SuitYellow}, InOneTrick: true}),
	dsCard(94, 2, 3, 3, Objective{Kind: KindEqualSuits, Suits: []Suit{SuitPink, SuitBlue}, InOneTrick: true}),
	dsCard(95, 1, 1, 1, Objective{Kind: KindMoreSuit, Suits: []Suit{SuitYellow, SuitBlue}}),
	dsCard(96, 1, 1, 1, Objective{Kind: KindMoreSuit, Suits: []Suit{SuitPink, SuitGreen}}),
}

// expectedDeepSeaByID indexes the expectation table; it fails the test if it is not exactly 1..96.
func expectedDeepSeaByID(t *testing.T) [TaskCards + 1]TaskCard {
	t.Helper()
	var by [TaskCards + 1]TaskCard
	for _, c := range expectedDeepSeaTasks {
		if c.ID < 1 || c.ID > TaskCards {
			t.Fatalf("expectation table: card %d out of range", c.ID)
		}
		if by[c.ID].ID != 0 {
			t.Fatalf("expectation table: card %d listed twice", c.ID)
		}
		by[c.ID] = c
	}
	for id := 1; id <= TaskCards; id++ {
		if by[id].ID == 0 {
			t.Fatalf("expectation table: card %d missing", id)
		}
	}
	return by
}

func TestDeepSeaExpectationTableComplete(t *testing.T) {
	// Count check of rules.md section 9: 12 + 15 + 9 + 7 + 15 + 25 + 6 + 2 + 5 = 96.
	if got := len(expectedDeepSeaTasks); got != 96 {
		t.Fatalf("expectation table has %d cards, want 96", got)
	}
	if TaskCards != 96 {
		t.Errorf("TaskCards = %d, want 96", TaskCards)
	}
	expectedDeepSeaByID(t)
}

func TestDeepSeaTaskCardTable(t *testing.T) {
	want := expectedDeepSeaByID(t)
	var all []string
	for id := 1; id <= TaskCards; id++ {
		got, ok := TaskCardFor(id)
		if !ok {
			all = append(all, fmt.Sprintf("card %d: TaskCardFor returned !ok", id))
			continue
		}
		for _, d := range dsDiff("", reflect.ValueOf(want[id]), reflect.ValueOf(got)) {
			all = append(all, fmt.Sprintf("card %d, field %s", id, strings.TrimPrefix(d, ".")))
		}
	}
	if len(all) > 0 {
		t.Errorf("%d mismatches:\n%s", len(all), strings.Join(all, "\n"))
	}
}

func TestDeepSeaTaskCardBounds(t *testing.T) {
	for _, id := range []int{-1, 0, 97, 1000} {
		if _, ok := TaskCardFor(id); ok {
			t.Errorf("TaskCardFor(%d) succeeded", id)
		}
	}
}

func TestDeepSeaNotForCaptain(t *testing.T) {
	for id := 1; id <= TaskCards; id++ {
		c, ok := TaskCardFor(id)
		if !ok {
			t.Errorf("card %d: TaskCardFor returned !ok", id)
			continue
		}
		want := id == 4 || id == 5 || id == 6
		if got := c.NotForCaptain(); got != want {
			t.Errorf("card %d: NotForCaptain() = %v, want %v", id, got, want)
		}
	}
}
