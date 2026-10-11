package domain

import (
	"fmt"
	"slices"
	"testing"
)

// bbObjCase plays tricks of a mission with special objectives. Specs are "winner:cards" (see
// bbPlaySpecs; "*card" first makes the winner's card). want: "run" (still playing after all specs),
// "win" (completed with the last spec) or "lose" (lost with the last spec, reason and, when >= 0,
// the failed objective's index).
type bbObjCase struct {
	name    string
	mission int
	n       int
	owners  map[int]int // objective index -> owner seat
	tasks   []string
	specs   []string
	want    string
	reason  Reason
	obj     int
}

func bbRunObjCases(t *testing.T, cases []bbObjCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := startMission(t, tc.n, tc.mission)
			a := g.Attempt
			for i, seat := range tc.owners {
				a.Objectives[i].Owner = seat
			}
			setTasks(g, tc.tasks...)
			played := bbPlaySpecs(t, g, tc.specs...)
			switch tc.want {
			case "run":
				if played != len(tc.specs) {
					t.Fatalf("the attempt ended after %d of %d tricks: %+v", played, len(tc.specs), a.Result)
				}
				bbExpectRunning(t, g, len(tc.specs))
			case "win":
				bbExpectEnd(t, g, true, ReasonCompleted, len(tc.specs))
			case "loseby":
				// lost at the latest after the last spec (an earlier loss is fine: the objective
				// can no longer be met)
				if g.Phase != PhaseDebrief || a.Result == nil || a.Result.Success || a.Result.Reason != tc.reason {
					t.Fatalf("phase %s result %+v", g.Phase, a.Result)
				}
			case "lose":
				if tc.reason == "" && a.Result != nil {
					bbExpectEnd(t, g, false, a.Result.Reason, len(tc.specs))

				} else {
					bbExpectEnd(t, g, false, tc.reason, len(tc.specs))
				}
				if tc.obj >= 0 && a.Result.Objective != tc.obj {
					t.Fatalf("failed objective %d, want %d", a.Result.Objective, tc.obj)
				}
			default:
				t.Fatal("bad want")
			}
		})
	}
}

// w builds count tricks won by the same seat; seq builds tricks won by the listed seats.
func bbW(count int, seat string) []string { return bbSpecs(0, count, seat) }

func bbSeq(seats ...int) []string {
	out := make([]string, len(seats))
	for i, s := range seats {
		out[i] = fmt.Sprintf("%d:", s)
	}
	return out
}

func cat(parts ...[]string) []string { return slices.Concat(parts...) }

func one(s ...string) []string { return s }

const lostObj = ReasonObjectiveFailed

// rules §13 mission 5: a member must win no trick.
func TestRulesObjectiveWinNoTrick(t *testing.T) {
	own2 := map[int]int{0: 2}
	bbRunObjCases(t, []bbObjCase{
		{"fulfilled only after the last trick", 5, 4, own2, nil, bbSeq(0, 1, 3, 0, 1, 3, 0, 1, 3, 0), "win", "", -1},
		{"still open before the last trick", 5, 4, own2, nil, bbSeq(0, 1, 3, 0, 1, 3, 0, 1, 3), "run", "", -1},
		{"fails when the member wins a trick", 5, 4, own2, nil, bbSeq(0, 1, 2), "lose", lostObj, 0},
		{"fails when the member wins the first trick", 5, 4, own2, nil, bbSeq(2), "lose", lostObj, 0},
		{"fails when the member wins the last trick", 5, 4, own2, nil, bbSeq(0, 1, 3, 0, 1, 3, 0, 1, 3, 2), "lose", lostObj, 0},
		{"others win as often as they like", 5, 4, own2, nil, bbSeq(0, 0, 0, 0), "run", "", -1},
		{"JARVIS as the member (2 players)", 5, 2, own2, nil, cat(bbSeq(0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0)), "win", "", -1},
		{"JARVIS wins a trick", 5, 2, own2, nil, bbSeq(0, 1, 2), "lose", lostObj, 0},
	})
}

// rules §13 missions 9 and 26: a trick is won by a colour 1 (once / twice).
func TestRulesObjectiveColourOne(t *testing.T) {
	bbRunObjCases(t, []bbObjCase{
		// mission 9: once
		{"9: a winning colour 1 completes at once", 9, 4, nil, nil, one("0:*b1"), "win", "", -1},
		{"9: any colour", 9, 4, nil, nil, one("1:*g1"), "win", "", -1},
		{"9: the 1 in a later trick", 9, 4, nil, nil, cat(bbSeq(0, 1), one("2:*g1")), "win", "", -1},
		{"9: a rocket does not count", 9, 4, nil, nil, one("0:*t1"), "run", "", -1},
		{"9: a 9 does not count", 9, 4, nil, nil, one("0:*b9"), "run", "", -1},
		{"9: a 1 that loses does not count", 9, 4, nil, nil, one("0:*t1 b1"), "run", "", -1},
		{"9: three 1s wasted, one left", 9, 4, nil, nil, one("0:*p9 p1", "0:*b9 b1", "0:*g9 g1"), "run", "", -1},
		{"9: four 1s wasted", 9, 4, nil, nil, one("0:*p9 p1", "0:*b9 b1", "0:*g9 g1", "0:*y9 y1"), "lose", lostObj, 0},
		{"9: the last trick without a 1", 9, 4, nil, nil, bbSeq(0, 1, 2, 3, 0, 1, 2, 3, 0, 1), "lose", lostObj, 0},
		{"9: still open before the last trick", 9, 4, nil, nil, bbSeq(0, 1, 2, 3, 0, 1, 2, 3, 0), "run", "", -1},
		// mission 26: twice
		{"26: one winning 1 is not enough", 26, 4, nil, nil, one("0:*p1"), "run", "", -1},
		{"26: two winning 1s", 26, 4, nil, nil, one("0:*p1", "1:*b1"), "win", "", -1},
		{"26: two winning 1s of the same player", 26, 4, nil, nil, one("0:*p1", "0:*b1"), "win", "", -1},
		{"26: two wasted 1s leave enough", 26, 4, nil, nil, one("0:*p9 p1", "0:*b9 b1"), "run", "", -1},
		{"26: three wasted 1s leave too few", 26, 4, nil, nil, one("0:*p9 p1", "0:*b9 b1", "0:*g9 g1"), "lose", lostObj, 0},
		{"26: one win and one left is just enough", 26, 4, nil, nil, one("0:*p1", "0:*b9 b1", "0:*g9 g1"), "run", "", -1},
		{"26: one win and no 1 left", 26, 4, nil, nil, one("0:*p1", "0:*b9 b1", "0:*g9 g1", "0:*y9 y1"), "lose", lostObj, 0},
		{"26: rockets and 9s do not count", 26, 4, nil, nil, one("0:*t1", "1:*b9"), "run", "", -1},
	})
}

// rules §13 missions 13 and 44: each rocket wins a trick (in the order 1, 2, 3, 4).
func TestRulesObjectiveRockets(t *testing.T) {
	bbRunObjCases(t, []bbObjCase{
		// mission 13
		{"13: one rocket wins", 13, 4, nil, nil, one("0:*t1"), "run", "", -1},
		{"13: four rockets win in any order", 13, 4, nil, nil, one("0:*t3", "1:*t1", "2:*t4", "3:*t2"), "win", "", -1},
		{"13: rockets win with colours between", 13, 4, nil, nil, one("0:*t3", "1:", "2:*t1", "0:", "1:*t4", "3:*t2"), "win", "", -1},
		{"13: a rocket is beaten by a higher rocket", 13, 4, nil, nil, one("0:*t3 t2"), "lose", lostObj, 0},
		{"13: the lowest rocket loses to a colour? no: it wins when led", 13, 4, nil, nil, one("0:*t1 p9"), "run", "", -1},
		{"13: a rocket played but not winning because trump 4 is higher", 13, 4, nil, nil, one("0:*t4 t1"), "lose", lostObj, 0},
		{"13: rockets that never get played (lost at the latest after the last trick)", 13, 4, nil, nil, bbSeq(0, 1, 2, 3, 0, 1, 2, 3, 0, 1), "loseby", lostObj, 0},
		{"13: three rockets won, the fourth still to come", 13, 4, nil, nil, one("0:*t1", "1:*t2", "2:*t3", "3:"), "run", "", -1},
		{"13: all four in the last tricks", 13, 4, nil, nil, cat(bbSeq(0, 1, 2, 3, 0, 1), one("0:*t1", "1:*t2", "2:*t3", "3:*t4")), "win", "", -1},
		// mission 44
		{"44: in order 1 2 3 4", 44, 4, nil, nil, one("0:*t1", "1:*t2", "2:*t3", "3:*t4"), "win", "", -1},
		{"44: in order with other tricks between", 44, 4, nil, nil, one("0:*t1", "1:", "1:*t2", "2:", "3:*t3", "0:*t4"), "win", "", -1},
		{"44: 1 then 2", 44, 4, nil, nil, one("0:*t1", "1:*t2"), "run", "", -1},
		{"44: 2 before 1", 44, 4, nil, nil, one("0:*t2"), "lose", lostObj, 0},
		{"44: 3 before 2", 44, 4, nil, nil, one("0:*t1", "1:*t3"), "lose", lostObj, 0},
		{"44: 4 first", 44, 4, nil, nil, one("0:*t4"), "lose", lostObj, 0},
		{"44: reversed order", 44, 4, nil, nil, one("0:*t4"), "lose", lostObj, 0},
		{"44: a rocket loses", 44, 4, nil, nil, one("0:*t1", "1:*t3 t2"), "lose", lostObj, 0},
		{"44: 1 then 4", 44, 4, nil, nil, one("0:*t1", "1:*t4"), "lose", lostObj, 0},
	})
}

// rules §13 missions 16 and 17: no trick is won by a colour 9; the mission is played to the end.
func TestRulesObjectiveNoNine(t *testing.T) {
	bbRunObjCases(t, []bbObjCase{
		{"16: no 9 wins: fulfilled after the last trick", 16, 4, nil, nil, bbSeq(0, 1, 2, 3, 0, 1, 2, 3, 0, 1), "win", "", -1},
		{"16: still open before the end", 16, 4, nil, nil, bbSeq(0, 1, 2, 3, 0, 1, 2, 3, 0), "run", "", -1},
		{"16: a 9 wins", 16, 4, nil, nil, one("0:*b9"), "lose", lostObj, 0},
		{"16: pink 9", 16, 4, nil, nil, one("0:", "1:*p9"), "lose", lostObj, 0},
		{"16: yellow 9 late", 16, 4, nil, nil, cat(bbSeq(0, 1, 2, 3, 0), one("3:*y9")), "lose", lostObj, 0},
		{"16: a 9 that loses to a rocket", 16, 4, nil, nil, one("0:*t1 b9"), "run", "", -1},
		{"16: a 9 that loses to a higher card of the colour? none: 9 is top; loses off colour", 16, 4, nil, nil, one("0:*p5 b9"), "run", "", -1},
		{"16: the winner is an 8", 16, 4, nil, nil, one("0:*b8 b7 b6 b5"), "run", "", -1},
		{"16: a 9 wins in the last trick", 16, 4, nil, nil, cat(bbSeq(0, 1, 2, 3, 0, 1, 2, 3, 0), one("1:*g9")), "lose", lostObj, 0},
		{"17: both tasks done, the mission goes on", 17, 4, nil, []string{"b2:0", "b3:1"}, one("0:*b2", "1:*b3"), "run", "", -1},
		{"17: played to the end", 17, 4, nil, []string{"b2:0", "b3:1"}, cat(one("0:*b2", "1:*b3"), bbSeq(2, 3, 0, 1, 2, 3, 0, 1)), "win", "", -1},
		{"17: a 9 wins after the tasks", 17, 4, nil, []string{"b2:0", "b3:1"}, one("0:*b2", "1:*b3", "2:*g9"), "lose", lostObj, 0},
		{"17: a 9 wins before the tasks", 17, 4, nil, []string{"b2:0", "b3:1"}, one("2:*g9"), "lose", lostObj, 0},
		{"17: a task fails", 17, 4, nil, []string{"b2:0", "b3:1"}, one("1:*b4 b2"), "lose", ReasonTaskFailed, -1},
		{"17: one task done, the other open", 17, 4, nil, []string{"b2:0", "b3:1"}, one("0:*b2"), "run", "", -1},
	})
}

// rules §13 missions 29 and 34: balance — nobody ever has two tricks more than another.
func TestRulesObjectiveBalance(t *testing.T) {
	bbRunObjCases(t, []bbObjCase{
		{"29: round robin is balanced", 29, 4, nil, nil, bbSeq(0, 1, 2, 3, 0, 1, 2, 3, 0, 1), "win", "", -1},
		{"29: the order within the round does not matter", 29, 4, nil, nil, bbSeq(3, 2, 1, 0, 1, 0, 3, 2, 0, 3), "win", "", -1},
		{"29: one ahead is fine", 29, 4, nil, nil, bbSeq(0), "run", "", -1},
		{"29: two tricks more than somebody", 29, 4, nil, nil, bbSeq(0, 0), "lose", lostObj, 0},
		{"29: two more than a member with none", 29, 4, nil, nil, bbSeq(0, 1, 0), "lose", lostObj, 0},
		{"29: three members won twice, one none", 29, 4, nil, nil, bbSeq(0, 1, 2, 0), "lose", lostObj, 0},
		{"29: open before the last trick", 29, 4, nil, nil, bbSeq(0, 1, 2, 3, 0, 1, 2, 3, 0), "run", "", -1},
		{"29: unbalanced in the last trick", 29, 4, nil, nil, bbSeq(0, 1, 2, 3, 0, 1, 2, 3, 0, 0), "lose", lostObj, 0},
		{"29: 2 players, JARVIS counts as a member", 29, 2, nil, nil, bbSeq(0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2, 0), "win", "", -1},
		{"29: 2 players, JARVIS never wins", 29, 2, nil, nil, bbSeq(0, 1, 0), "lose", lostObj, 0},
		{"29: 3 players", 29, 3, nil, nil, bbSeq(0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2, 0), "win", "", -1},
		{"29: 5 players", 29, 5, nil, nil, bbSeq(0, 1, 2, 3, 4, 0, 1, 2), "win", "", -1},
		{"29: 5 players unbalanced", 29, 5, nil, nil, bbSeq(0, 1, 2, 3, 4, 0, 0), "lose", lostObj, 0},
		// mission 34: the commander (seat 3) wins the first and the last trick, and balance
		{"34: commander wins first and last, balanced", 34, 4, map[int]int{1: 3}, nil, bbSeq(3, 0, 1, 2, 0, 1, 2, 3, 0, 3), "win", "", -1},
		{"34: another wins the first trick", 34, 4, map[int]int{1: 3}, nil, bbSeq(0), "lose", lostObj, 1},
		{"34: the commander wins the first, still open", 34, 4, map[int]int{1: 3}, nil, bbSeq(3, 0), "run", "", -1},
		{"34: another wins the last trick", 34, 4, map[int]int{1: 3}, nil, bbSeq(3, 0, 1, 2, 0, 1, 2, 3, 0, 1), "lose", lostObj, 1},
		{"34: balance broken", 34, 4, map[int]int{1: 3}, nil, bbSeq(3, 0, 3), "lose", lostObj, 0},
		{"34: the commander may win other tricks too", 34, 4, map[int]int{1: 3}, nil, bbSeq(3, 0, 1, 2, 3), "run", "", -1},
	})
}

// rules §13 missions 33 and 41: a chosen member's trick conditions, not with rockets.
func TestRulesObjectiveChosenMember(t *testing.T) {
	own1 := map[int]int{0: 1}
	bbRunObjCases(t, []bbObjCase{
		// 33: exactly one trick, not with a rocket
		{"33: exactly one trick (the last)", 33, 4, own1, nil, cat(bbW(9, "0"), bbSeq(1)), "win", "", -1},
		{"33: exactly one trick (the first)", 33, 4, own1, nil, cat(bbSeq(1), bbW(9, "0")), "win", "", -1},
		{"33: one trick, still open before the end", 33, 4, own1, nil, bbSeq(1, 0), "run", "", -1},
		{"33: a second trick", 33, 4, own1, nil, bbSeq(1, 0, 1), "lose", lostObj, 0},
		{"33: one trick won with a rocket", 33, 4, own1, nil, one("1:*t1"), "lose", lostObj, 0},
		{"33: the winner uses a rocket against a higher card", 33, 4, own1, nil, one("1:*t3 t1"), "lose", lostObj, 0},
		{"33: not a single trick", 33, 4, own1, nil, bbW(10, "0"), "lose", lostObj, 0},
		{"33: no trick yet, still open", 33, 4, own1, nil, bbW(9, "0"), "run", "", -1},
		{"33: the trick of another member is a rocket", 33, 4, own1, nil, one("0:*t1", "1:"), "run", "", -1},
		// 41: only the first and the last trick, not with rockets
		{"41: first and last", 41, 4, own1, nil, cat(one("1:"), bbW(8, "0"), one("1:")), "win", "", -1},
		{"41: first trick missed", 41, 4, own1, nil, bbSeq(0), "lose", lostObj, 0},
		{"41: first won, a trick between", 41, 4, own1, nil, bbSeq(1, 0, 1), "lose", lostObj, 0},
		{"41: first won, second won", 41, 4, own1, nil, bbSeq(1, 1), "lose", lostObj, 0},
		{"41: first with a rocket", 41, 4, own1, nil, one("1:*t1"), "lose", lostObj, 0},
		{"41: last missed", 41, 4, own1, nil, cat(one("1:"), bbW(9, "0")), "lose", lostObj, 0},
		{"41: last with a rocket", 41, 4, own1, nil, cat(one("1:"), bbW(8, "0"), one("1:*t2")), "lose", lostObj, 0},
		{"41: first won, waiting", 41, 4, own1, nil, bbSeq(1, 0, 2), "run", "", -1},
		{"41: first and last in a 5-player game", 41, 5, own1, nil, cat(one("1:"), bbW(6, "0"), one("1:")), "win", "", -1},
		{"41: first and last in a 3-player game", 41, 3, own1, nil, cat(one("1:"), bbW(11, "0"), one("1:")), "win", "", -1},
	})
}

// rules §13 mission 50: A wins only the first four tricks, B only the last.
func TestRulesObjectiveRoles(t *testing.T) {
	roles := map[int]int{0: 0, 1: 1}
	bbRunObjCases(t, []bbObjCase{
		{"A first four, B last, others between", 50, 4, roles, nil, bbSeq(0, 0, 0, 0, 2, 3, 2, 3, 2, 1), "win", "", -1},
		{"A first four, waiting", 50, 4, roles, nil, bbSeq(0, 0, 0, 0), "run", "", -1},
		{"A misses trick 3", 50, 4, roles, nil, bbSeq(0, 0, 2), "lose", lostObj, 0},
		{"A misses trick 1", 50, 4, roles, nil, bbSeq(2), "lose", lostObj, 0},
		{"A wins trick 5", 50, 4, roles, nil, bbSeq(0, 0, 0, 0, 0), "lose", lostObj, 0},
		{"A wins a later trick", 50, 4, roles, nil, bbSeq(0, 0, 0, 0, 2, 3, 0), "lose", lostObj, 0},
		{"B wins before the last trick", 50, 4, roles, nil, bbSeq(0, 0, 0, 0, 1), "lose", lostObj, 1},
		{"B wins within the first four", 50, 4, roles, nil, bbSeq(0, 1), "lose", lostObj, -9},
		{"B misses the last trick", 50, 4, roles, nil, bbSeq(0, 0, 0, 0, 2, 3, 2, 3, 2, 3), "lose", lostObj, 1},
		{"A wins the last trick", 50, 4, roles, nil, bbSeq(0, 0, 0, 0, 2, 3, 2, 3, 2, 0), "lose", lostObj, -9},
		{"others win freely between", 50, 4, roles, nil, bbSeq(0, 0, 0, 0, 2, 2, 2, 2, 3, 1), "win", "", -1},
		{"5 players", 50, 5, roles, nil, bbSeq(0, 0, 0, 0, 2, 3, 4, 1), "win", "", -1},
		{"3 players: all tricks", 50, 3, roles, nil, cat(bbSeq(0, 0, 0, 0), bbW(8, "2"), bbSeq(1)), "win", "", -1},
	})
}

// rules §13 mission 46: the owner wins all pink cards.
func TestRulesObjectiveAllPink(t *testing.T) {
	own0 := map[int]int{0: 0}
	bbRunObjCases(t, []bbObjCase{
		{"all pink cards won by the owner", 46, 4, own0, nil, one("0:*p9 p1 p2", "0:*p8 p3 p4", "0:*p7 p5 p6"), "win", "", -1},
		{"some pink won, the rest still to come", 46, 4, own0, nil, one("0:*p9 p1 p2"), "run", "", -1},
		{"another member wins a pink card", 46, 4, own0, nil, one("1:*p9"), "lose", lostObj, 0},
		{"another member wins a trick with a pink card in it", 46, 4, own0, nil, one("1:*b9 p1"), "lose", lostObj, 0},
		{"the owner wins a pink card with a rocket", 46, 4, own0, nil, one("0:*t1 p1 p2"), "run", "", -1},
		{"another member trumps a pink card", 46, 4, own0, nil, one("1:*t1 p5"), "lose", lostObj, 0},
		{"tricks without pink do not matter", 46, 4, own0, nil, bbSeq(1, 2, 3), "run", "", -1},
		{"a pink card never played", 46, 4, own0, nil, cat(one("0:*p9 p1 p2"), bbSeq(1, 2, 3, 1, 2, 3, 1, 2, 3)), "lose", lostObj, 0},
		{"no pink card at all", 46, 4, own0, nil, bbSeq(0, 1, 2, 3, 0, 1, 2, 3, 0, 1), "lose", lostObj, 0},
		{"the owner is JARVIS", 46, 2, map[int]int{0: 2}, nil, one("2:*p9 p1 p2", "2:*p8 p3 p4", "2:*p7 p5 p6"), "win", "", -1},
	})
}

// rules §13 mission 48: the omega task is won in the last trick.
func TestRulesObjectiveOmegaInLastTrick(t *testing.T) {
	tasks := []string{"p1:0:omega", "p2:0", "p3:0"}
	bbRunObjCases(t, []bbObjCase{
		{"omega task in the last trick", 48, 4, nil, tasks, cat(one("0:*p3 p2"), bbSeq(1, 2, 3, 1, 2, 3, 1, 2), one("0:*p1")), "win", "", -1},
		{"the omega task earlier", 48, 4, nil, tasks, one("0:*p3 p2", "0:*p1"), "lose", lostObj, 0},
		{"all three tasks in one trick", 48, 4, nil, tasks, one("0:*p3 p1 p2"), "lose", lostObj, 0},
		{"waiting", 48, 4, nil, tasks, one("0:*p3 p2", "1:", "2:"), "run", "", -1},
		{"omega won by someone else in the last trick", 48, 4, nil, tasks, cat(one("0:*p3 p2"), bbSeq(1, 2, 3, 1, 2, 3, 1, 2), one("1:*p1")), "lose", "", -1},
		{"omega never played", 48, 4, nil, tasks, cat(one("0:*p3 p2"), bbSeq(1, 2, 3, 1, 2, 3, 1, 2, 3)), "lose", "", -1},
	})
}
