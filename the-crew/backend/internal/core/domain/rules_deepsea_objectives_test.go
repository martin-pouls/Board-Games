package domain

import "testing"

// Rules deep-sea/rules.md §8: the mission-level crew objectives (missions 8, 12, 21, 23, 27) and
// the side quests Q1 and Q2. The history DSL is explained in rules_deepsea_tasks_test.go.
func TestRulesDeepSeaMissionObjectives(t *testing.T) {
	balance := func(rank Rank) Objective {
		return Objective{Kind: KindCrewBalance, Owner: OwnerCrew, Count: 1, Rank: rank}
	}
	neverLead := Objective{Kind: KindNeverLead, Owner: OwnerCrew, Suits: []Suit{SuitPink, SuitTrump}}
	ahead := Objective{Kind: KindLeaderStaysAhead, Owner: OwnerCrew}
	yellow5 := Objective{Kind: KindLastCardPlayed, Owner: OwnerCrew, Cards: []Card{{SuitYellow, 5}}}
	q1 := func(n int) Objective {
		return Objective{Kind: KindWinExactlyTricks, Owner: 0, Count: n, Without: []Rank{7, 8, 9}}
	}
	quota := func(targets ...int) Objective {
		return Objective{Kind: KindTrickQuota, Owner: OwnerCrew, Targets: targets}
	}

	cases := []tkCase{
		// 8: at no point may a member have won two more 9s than any other member (dummy included)
		tkm("mission08", "a lead of one is allowed", balance(9), tkF, "0|p9", "1|b9", "2|g9", "0|y9", "~1"),
		tkm("mission08", "a lead of two fails", balance(9), tkX, "0|p9 b9"),
		tkm("mission08", "a lead over the dummy counts", balance(9), tkX, "0|p9", "1|b9", "0|g9"),
		tkm("mission08", "catching up later does not undo it", balance(9), tkX, "0|p9 b9", "1|g9 y9"),
		tkm("mission08", "open", balance(9), tkO, "0|p9"),

		// 21: the same for 1s (submarine 1 is not a 1)
		tkm("mission21", "a lead of one is allowed", balance(1), tkF, "0:t1 1:p1 2:b2=0", "1|b1", "2|g1", "0|y1", "~1"),
		tkm("mission21", "a lead of two fails", balance(1), tkX, "0|p1 b1"),
		tkm("mission21", "open", balance(1), tkO, "0|p1"),

		// 12: no trick may be led with a pink card or a submarine, by any member
		tkm("mission12", "nobody leads pink or a submarine (4 members)", neverLead, tkF, "~0:5", "~1:5").mem(4).nolead("pt"),
		tkm("mission12", "nobody leads pink or a submarine (3 members)", neverLead, tkF, "~0:7", "~1:6").nolead("pt"),
		tkm("mission12", "pink lead", neverLead, tkX, "0:p9 1:b1 2:g1=0"),
		tkm("mission12", "submarine lead", neverLead, tkX, "0:t1 1:b1 2:g1=0"),
		tkm("mission12", "the dummy leads pink", neverLead, tkX, "2:p5 0:b1 1:g1=2"),
		tkm("mission12", "following pink or a submarine is fine", neverLead, tkO, "0:y9 1:p1 2:t1=2"),
		tkm("mission12", "open", neverLead, tkO, "0:y9 1:y1 2:y2=0"),

		// 23: the winner of the first trick always has strictly more tricks than every other member
		tkm("mission23", "stays ahead to the end", ahead, tkF, tkWins("0 0 1 0 2 0 0 1 0 2 0 0 1")...),
		tkm("mission23", "stays ahead, first winner is seat 1", ahead, tkF, tkWins("1 1 0 1 1 2 1 1 0 1 1 2 1")...),
		tkm("mission23", "a tie after trick two fails", ahead, tkX, tkWins("0 1")...),
		tkm("mission23", "a tie with the dummy fails", ahead, tkX, tkWins("0 2")...),
		tkm("mission23", "another member overtakes", ahead, tkX, tkWins("0 1 1")...),
		tkm("mission23", "the first winner is overtaken later", ahead, tkX, tkWins("0 0 1 1 1")...),
		tkm("mission23", "open", ahead, tkO, tkWins("0 0")...),

		// 27: the yellow 5 is the last card played in the last trick
		tkm("mission27", "last card of the last trick", yellow5, tkF, "~1:12", "0:g9 1:b1 2:y5=0"),
		tkm("mission27", "4 members", yellow5, tkF, "~1:9", "0:g9 1:b1 2:p1 3:y5=0").mem(4),
		tkm("mission27", "5 members", yellow5, tkF, "~1:7", "0:g9 1:b1 2:p1 3:y1 4:y5=0").mem(5),
		tkm("mission27", "played first in the last trick", yellow5, tkX, "~1:12", "0:y5 1:g1 2:b1=0"),
		tkm("mission27", "played in an earlier trick", yellow5, tkX, "0:y9 1:y5 2:y1=0", "~1"),
		tkm("mission27", "another card is played last", yellow5, tkX, "~1:12", "0:g9 1:b1 2:y1=0"),
		tkm("mission27", "it is the left-over card (3 players)", yellow5, tkX, "~1").lft("y5"),
		tkm("mission27", "open", yellow5, tkO, "~1:5"),

		// Q1: the captain wins exactly 2 (stage B: 3) tricks, none containing a 7, 8 or 9
		tkm("Q1", "two clean tricks", q1(2), tkF, "0:y5 1:y1 2:y2=0", "0:b5 1:b1 2:b2=0", "~1"),
		tkm("Q1", "a submarine is fine", q1(2), tkF, "0:t4 1:y1 2:y2=0", "0:y5 1:b1 2:b2=0", "~1"),
		tkm("Q1", "a 7 in a counted trick", q1(2), tkX, "0:y7 1:y1 2:y2=0"),
		tkm("Q1", "an off-colour 9 in a counted trick", q1(2), tkX, "0:y5 1:g9 2:y1=0"),
		tkm("Q1", "an 8 played by another member", q1(2), tkX, "0:y5 1:b8 2:y1=0"),
		tkm("Q1", "three tricks", q1(2), tkX, "0:y5 1:y1 2:y2=0", "0:b5 1:b1 2:b2=0", "0:g5 1:g1 2:g2=0"),
		tkm("Q1", "one trick at the end", q1(2), tkX, "0:y5 1:y1 2:y2=0", "~1"),
		tkm("Q1", "one clean trick is open", q1(2), tkO, "0:y5 1:y1 2:y2=0"),
		tkm("Q1", "7-9 cards in tricks of other members are fine", q1(2), tkF, "0:y5 1:y1 2:y2=0", "0:b5 1:b1 2:b2=0", "1|p9 p8 p7", "~1"),
		tkm("Q1", "stage B: three clean tricks", q1(3), tkF, "0:y5 1:y1 2:y2=0", "0:b5 1:b1 2:b2=0", "0:g5 1:g1 2:g2=0", "~1"),
		tkm("Q1", "stage B: an 8 in the third trick", q1(3), tkX, "0:y5 1:y1 2:y2=0", "0:b5 1:b1 2:b2=0", "0:g8 1:g1 2:g2=0"),
		tkm("Q1", "stage B: two tricks only", q1(3), tkX, "0:y5 1:y1 2:y2=0", "0:b5 1:b1 2:b2=0", "~1"),

		// Q2: every member wins exactly as many tricks as the sum of the difficulties of the cards
		// selected (the dummy is a member); open while the targets are unknown
		tkm("Q2", "everyone hits the target", quota(5, 5, 3), tkF, "~0:5", "~1:5", "~2:3"),
		tkm("Q2", "4 members", quota(3, 3, 2, 2), tkF, "~0:3", "~1:3", "~2:2", "~3:2").mem(4),
		tkm("Q2", "a member with no cards wins no trick", quota(0, 6, 7), tkF, "~1:6", "~2:7"),
		tkm("Q2", "a member with no cards wins a trick", quota(0, 6, 7), tkX, "~0:1"),
		tkm("Q2", "one trick too many", quota(5, 5, 3), tkX, "~0:6"),
		tkm("Q2", "the dummy misses its target", quota(5, 5, 3), tkX, "~0:5", "~1:4", "~2:4"),
		tkm("Q2", "on the way", quota(5, 5, 3), tkO, "~0:2"),
		tkm("Q2", "targets not set yet", Objective{Kind: KindTrickQuota, Owner: OwnerCrew}, tkO, "~0:2"),
	}
	tkRun(t, cases)
	t.Logf("%d mission objective cases", len(cases))
}
