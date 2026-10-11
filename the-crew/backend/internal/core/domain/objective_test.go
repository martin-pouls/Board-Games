package domain

import "testing"

// hist builds a history of 4 seats with the given total and tricks.
func hist(total int, tricks ...TrickRecord) History {
	return History{Tricks: tricks, Total: total, Seats: 4}
}

// Tricks used below (4 seats). The winner is computed by trick().
var (
	t0p7 = trick("1:p1", "2:p2", "3:p3", "0:p7") // seat 0 wins with p7
	t1p7 = trick("0:p7", "1:p9", "2:p2", "3:p3") // seat 1 wins, p7 inside
	t0b9 = trick("0:b9", "1:b1", "2:b2", "3:b3") // seat 0 wins with a colour
	t1b9 = trick("1:b9", "2:b4", "3:b5", "0:b6") // seat 1 wins with a colour
	t2g9 = trick("2:g9", "3:g1", "0:g2", "1:g3") // seat 2 wins with g9
	t0t1 = trick("1:y1", "2:y2", "3:y3", "0:t1") // seat 0 wins with trump 1
	t1t1 = trick("0:y4", "1:t1", "2:y5", "3:y6") // seat 1 wins with trump 1
	t3y9 = trick("0:y4", "1:y5", "2:y6", "3:y9") // seat 3 wins with y9
)

func TestObjectiveStatus(t *testing.T) {
	win := func(owner int, card string) Objective {
		return Objective{Kind: KindWinCards, Owner: owner, Cards: []Card{c(card)}}
	}
	ones := []Card{c("p1"), c("b1"), c("g1"), c("y1")}
	trumps := []Card{c("t1"), c("t2"), c("t3"), c("t4")}
	for _, tc := range []struct {
		name string
		o    Objective
		h    History
		want Status
	}{
		// Tasks (§6).
		{"task open", win(0, "p7"), hist(3), StatusOpen},
		{"task won by owner", win(0, "p7"), hist(3, t0p7), StatusFulfilled},
		{"task won by another", win(0, "p7"), hist(3, t1p7), StatusFailed},
		{"task unplayed at the end", win(0, "p7"), hist(2, t0b9, t1b9), StatusFailed},
		{"task not yet played", win(0, "p7"), hist(3, t0b9, t1b9), StatusOpen},

		// Win no trick (5).
		{"no trick: open", Objective{Kind: KindWinNoTricks, Owner: 1}, hist(3, t0b9), StatusOpen},
		{"no trick: won one", Objective{Kind: KindWinNoTricks, Owner: 1}, hist(3, t0b9, t1b9), StatusFailed},
		{"no trick: done", Objective{Kind: KindWinNoTricks, Owner: 1}, hist(2, t0b9, t2g9), StatusFulfilled},

		// Win exactly one trick, not with trump (33).
		{"exactly 1: one won, not over", exactly(1), hist(3, t1b9), StatusOpen},
		{"exactly 1: one won, over", exactly(1), hist(2, t1b9, t0b9), StatusFulfilled},
		{"exactly 1: won with trump", exactly(1), hist(3, t1t1), StatusFailed},
		{"exactly 1: won two", exactly(1), hist(3, t1b9, t1p7), StatusFailed},
		{"exactly 1: none, one left", exactly(1), hist(3, t0b9, t2g9), StatusOpen},
		{"exactly 1: none, none left", exactly(1), hist(2, t0b9, t2g9), StatusFailed},

		// Win only the first and the last trick, not with trump (41).
		{"only 1+last: first won", only(0, true, 1, LastTrick), hist(3, t0b9), StatusOpen},
		{"only 1+last: first lost", only(0, true, 1, LastTrick), hist(3, t1b9), StatusFailed},
		{"only 1+last: won the middle", only(0, true, 1, LastTrick), hist(3, t0b9, t0p7), StatusFailed},
		{"only 1+last: done", only(0, true, 1, LastTrick), hist(3, t0b9, t1b9, t0p7), StatusFulfilled},
		{"only 1+last: last lost", only(0, true, 1, LastTrick), hist(3, t0b9, t1b9, t2g9), StatusFailed},
		{"only 1+last: first by trump", only(0, true, 1, LastTrick), hist(3, t0t1), StatusFailed},
		{"only 1..4 (50 A): trump allowed", only(0, false, 1, 2, 3, 4), hist(5, t0t1, t0b9, t0p7, t0b9), StatusOpen},
		{"only 1..4 (50 A): wins the 5th", only(0, false, 1, 2, 3, 4), hist(6, t0t1, t0b9, t0p7, t0b9, t0b9), StatusFailed},
		{"only last (50 B): wins early", only(1, false, LastTrick), hist(5, t1b9), StatusFailed},

		// Win the first and the last trick (34, commander).
		{"first+last: first won", firstLast(0), hist(3, t0b9), StatusOpen},
		{"first+last: first lost", firstLast(0), hist(3, t1b9), StatusFailed},
		{"first+last: middle irrelevant", firstLast(0), hist(3, t0b9, t0p7, t0b9), StatusFulfilled},
		{"first+last: last lost", firstLast(0), hist(3, t0b9, t1b9, t2g9), StatusFailed},

		// Win all pink cards (46).
		{"all pink: someone else wins pink", allPink(0), hist(10, t1p7), StatusFailed},
		{"all pink: some won", allPink(0), hist(10, t0p7), StatusOpen},
		{"all pink: all won", allPink(0), hist(10,
			trick("0:p9", "1:p1", "2:p2", "3:p3"),
			trick("0:p8", "1:p4", "2:p5", "3:p6"),
			trick("0:p7", "1:b1", "2:b2", "3:b3")), StatusFulfilled},

		// A trick won by a colour 1, once / twice (9, 26).
		{"any 1: won by p1", anyOf(ones, 1), hist(10, trick("0:p1", "1:b2", "2:g3", "3:y4")), StatusFulfilled},
		{"any 1: lost by p1", anyOf(ones, 1), hist(10, trick("0:p1", "1:p2", "2:b1", "3:g1")), StatusOpen},
		{"any 1: all four lost", anyOf(ones, 1), hist(10,
			trick("0:p1", "1:p2", "2:b1", "3:g1"),
			trick("0:y1", "1:y2", "2:b3", "3:b4")), StatusFailed},
		{"two 1s: too few tricks left", anyOf(ones, 2), hist(2, t0b9), StatusFailed},
		{"two 1s: one won, one left", anyOf(ones, 2), hist(3, trick("0:p1", "1:b2", "2:g3", "3:y4")), StatusOpen},
		{"two 1s: both", anyOf(ones, 2), hist(10,
			trick("0:p1", "1:b2", "2:g3", "3:y4"),
			trick("0:g1", "1:b3", "2:p3", "3:y5")), StatusFulfilled},

		// Each trump wins a trick, unordered / ordered (13, 44).
		{"each trump: t2 won", each(trumps, false), hist(10, trick("0:t2", "1:p1", "2:p2", "3:p3")), StatusOpen},
		{"each trump: t1 lost to t3", each(trumps, false), hist(10, trick("0:p1", "1:t1", "2:t3", "3:p3")), StatusFailed},
		{"each trump: too few tricks", each(trumps, false), hist(4, t0b9), StatusFailed},
		{"each trump: all four", each(trumps, false), hist(10,
			trick("0:t2", "1:p1", "2:p2", "3:p3"),
			trick("0:t4", "1:p4", "2:p5", "3:p6"),
			trick("0:t1", "1:b4", "2:b5", "3:b6"),
			trick("0:t3", "1:g4", "2:g5", "3:g6")), StatusFulfilled},
		{"ordered: t2 before t1", each(trumps, true), hist(10, trick("0:t2", "1:p1", "2:p2", "3:p3")), StatusFailed},
		{"ordered: t1 then t2", each(trumps, true), hist(10,
			trick("0:t1", "1:p1", "2:p2", "3:p3"),
			trick("0:t2", "1:p4", "2:p5", "3:p6")), StatusOpen},

		// No trick won by a 9 (16, 17).
		{"no 9: won by g9", no9(), hist(3, t2g9), StatusFailed},
		{"no 9: trump beats p9", no9(), hist(3, trick("0:p9", "1:t1", "2:p2", "3:p3")), StatusOpen},
		{"no 9: done", no9(), hist(2, t0p7, t0t1), StatusFulfilled},

		// Balance (29, 34): never two tricks more than another member.
		{"balance: one ahead", balance(), hist(10, t0b9), StatusOpen},
		{"balance: two ahead", balance(), hist(10, t0b9, t0p7), StatusFailed},
		{"balance: all even", balance(), hist(4, t0b9, t1b9, t2g9, t3y9), StatusFulfilled},
		{"balance: was two ahead earlier", balance(), hist(10, t0b9, t0p7, t1b9, t2g9, t3y9), StatusFailed},

		// The Ω task's card in the last trick (48).
		{"last trick card: too early", lastCard("p7"), hist(3, t0p7), StatusFailed},
		{"last trick card: in the last trick", lastCard("p7"), hist(2, t0b9, t0p7), StatusFulfilled},
		{"last trick card: never played", lastCard("p7"), hist(2, t0b9, t1b9), StatusFailed},
		{"last trick card: not yet", lastCard("p7"), hist(3, t0b9), StatusOpen},
	} {
		if got := tc.o.Status(tc.h); got != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, got, tc.want)
		}
	}
}

func exactly(n int) Objective {
	return Objective{Kind: KindWinExactlyTricks, Owner: 1, Count: n, NotWithTrump: true}
}

func only(owner int, notTrump bool, tricks ...int) Objective {
	return Objective{Kind: KindWinOnlyTricks, Owner: owner, Tricks: tricks, NotWithTrump: notTrump}
}

func firstLast(owner int) Objective {
	return Objective{Kind: KindWinTricks, Owner: owner, Tricks: []int{1, LastTrick}}
}

func allPink(owner int) Objective {
	return Objective{Kind: KindWinAllOfSuit, Owner: owner, Suit: SuitPink}
}

func anyOf(cards []Card, n int) Objective {
	return Objective{Kind: KindCrewWinWithAny, Owner: OwnerCrew, Cards: cards, Count: n}
}

func each(cards []Card, ordered bool) Objective {
	return Objective{Kind: KindCrewWinWithEach, Owner: OwnerCrew, Cards: cards, Ordered: ordered}
}

func no9() Objective { return Objective{Kind: KindCrewNeverWinWith, Owner: OwnerCrew, Rank: 9} }

func balance() Objective { return Objective{Kind: KindCrewBalance, Owner: OwnerCrew, Count: 1} }

func lastCard(card string) Objective {
	return Objective{Kind: KindCardInLastTrick, Owner: OwnerCrew, Cards: []Card{c(card)}}
}

func TestFulfilledAt(t *testing.T) {
	o := Objective{Kind: KindWinCards, Owner: 0, Cards: []Card{c("p7")}}
	if got := o.FulfilledAt(hist(5, t0b9, t0p7, t1b9)); got != 2 {
		t.Fatalf("FulfilledAt = %d, want 2", got)
	}
	if got := o.FulfilledAt(hist(5, t0b9)); got != 0 {
		t.Fatalf("open: FulfilledAt = %d, want 0", got)
	}
	if got := o.FulfilledAt(hist(5, t1p7)); got != 0 {
		t.Fatalf("failed: FulfilledAt = %d, want 0", got)
	}
}
