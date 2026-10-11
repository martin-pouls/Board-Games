package domain

import (
	"strconv"
	"strings"
	"testing"
)

// bbRec builds a completed trick won by seat w (n seats): w leads y8, the task cards listed in
// cards are played by the following seats, free places get low yellow cards. Off-colour cards
// never win, so w really wins. Keep yellow cards and rank 1/9 yellow out of the tests that use it.
func bbRec(n, w int, cards string) TrickRecord {
	r := TrickRecord{Leader: w, Winner: w}
	r.Plays = append(r.Plays, Play{Seat: w, Card: c("y8")})
	fs := strings.Fields(cards)
	for i := 1; i < n; i++ {
		card := c("y" + strconv.Itoa(i+1))
		if i-1 < len(fs) {
			card = c(fs[i-1])
		}
		r.Plays = append(r.Plays, Play{Seat: (w + i) % n, Card: card})
	}
	return r
}

// bbRecs builds tricks from "winner:cards" strings, e.g. "0:p1 p2", "2:".
func bbRecs(n int, specs ...string) []TrickRecord {
	var out []TrickRecord
	for _, s := range specs {
		w, cards, _ := strings.Cut(s, ":")
		seat, err := strconv.Atoi(w)
		if err != nil {
			panic("bad spec " + s)
		}
		out = append(out, bbRec(n, seat, cards))
	}
	return out
}

// bbStatuses sets the tasks and the completed tricks of an n-player attempt and renders the task
// statuses: O open, F fulfilled, X failed.
func bbStatuses(t *testing.T, n int, tasks []string, recs []TrickRecord) string {
	t.Helper()
	g := startMission(t, n, 3) // mission 3: no objectives
	g.Attempt.Objectives = nil
	setTasks(g, tasks...)
	g.Attempt.Tricks = recs
	ts, _ := g.Attempt.Statuses()
	var sb strings.Builder
	for _, s := range ts {
		switch s {
		case StatusOpen:
			sb.WriteByte('O')
		case StatusFulfilled:
			sb.WriteByte('F')
		case StatusFailed:
			sb.WriteByte('X')
		default:
			sb.WriteByte('?')
		}
	}
	return sb.String()
}

// lost marks a case where the attempt is lost (some task failed); the exact statuses of the
// other tasks are not specified by the rules.
const lost = "LOST"

// rules §6: fulfilling and failing tasks, derived from the completed tricks.
func TestRulesTaskStatuses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		tasks  []string
		tricks []string // "winner:cards"
		want   string
	}{
		{"nothing played", []string{"p1:0", "p2:1"}, nil, "OO"},
		{"owner wins a trick with the card", []string{"p1:0"}, []string{"0:p1"}, "F"},
		{"someone else wins the card", []string{"p1:0"}, []string{"1:p1"}, "X"},
		{"unrelated trick leaves the task open", []string{"p1:0"}, []string{"1:p2 p3"}, "O"},
		{"fulfilled in a later trick", []string{"p1:0"}, []string{"1:p2", "3:", "0:p1"}, "F"},
		{"failed in a later trick", []string{"p1:0"}, []string{"1:p2", "3:", "2:p1"}, "X"},
		{"two tasks, same owner, same trick", []string{"p1:0", "p2:0"}, []string{"0:p1 p2"}, "FF"},
		{"two tasks, same owner, same trick, three cards", []string{"p1:0", "p2:0", "p3:0"}, []string{"0:p1 p2 p3"}, "FFF"},
		{"two tasks, same owner, different tricks", []string{"p1:0", "p2:0"}, []string{"0:p1", "0:p2"}, "FF"},
		{"two tasks of different owners, each wins their own", []string{"p1:0", "p2:1"}, []string{"0:p1", "1:p2"}, "FF"},
		{"one task done, the other still open", []string{"p1:0", "p2:1"}, []string{"0:p1"}, "FO"},
		{"two tasks in one trick, the winner owns only one", []string{"p1:0", "p2:1"}, []string{"0:p1 p2"}, lost},
		{"the other task fails with the trick of the first", []string{"p1:0", "p2:1"}, []string{"1:p1 p2"}, lost},
		{"a task of another member is not mine", []string{"p1:2"}, []string{"0:p1"}, "X"},
		{"owner is the dummy-free seat 3", []string{"p9:3"}, []string{"3:p9"}, "F"},
		{"all tricks played, task card never played", []string{"p1:0"}, bbSpecs(4, 10, "1"), "X"},
		{"all tricks played, one task done, one never played", []string{"p1:0", "p2:0"}, append([]string{"0:p1"}, bbSpecs(4, 9, "1")...), "FX"},
		{"all tricks played, both done", []string{"p1:0", "p2:0"}, append([]string{"0:p1 p2"}, bbSpecs(4, 9, "1")...), "FF"},
		{"the last trick fulfils the task", []string{"p1:0"}, append(bbSpecs(4, 9, "1"), "0:p1"), "F"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := bbStatuses(t, 4, tc.tasks, bbRecs(4, tc.tricks...))
			assertStatuses(t, got, tc.want)
		})
	}
}

// bbSpecs returns count tricks without task cards, all won by the given seat.
func bbSpecs(n, count int, winner string) []string {
	out := make([]string, count)
	for i := range out {
		out[i] = winner + ":"
	}
	return out
}

func assertStatuses(t *testing.T, got, want string) {
	t.Helper()
	if want == lost {
		if !strings.Contains(got, "X") {
			t.Fatalf("statuses %q, want the attempt lost (some X)", got)
		}
		return
	}
	if got != want {
		t.Fatalf("statuses %q, want %q", got, want)
	}
}

// rules §7: numbered tokens, the generalised same-trick rule (b < k <= b+s), open-task failure.
func TestRulesTokenNumbered(t *testing.T) {
	for _, tc := range []struct {
		name   string
		tasks  []string
		tricks []string
		want   string
	}{
		{"1, 2, 3 in order", []string{"p1:0:1", "p2:1:2", "p3:2:3"}, []string{"0:p1", "1:p2", "2:p3"}, "FFF"},
		{"1 and 2 in the same trick", []string{"p1:0:1", "p2:0:2"}, []string{"0:p1 p2"}, "FF"},
		{"1 and 2 in the same trick, tokens on swapped cards", []string{"p1:0:2", "p2:0:1"}, []string{"0:p1 p2"}, "FF"},
		{"2 before 1", []string{"p1:0:1", "p2:1:2"}, []string{"1:p2"}, lost},
		{"3 before 1 and 2", []string{"p1:0:1", "p2:1:2", "p3:2:3"}, []string{"2:p3"}, lost},
		{"1 is not 'in the first trick'", []string{"p1:0:1", "p2:1:2"}, []string{"3:", "0:p1", "1:p2"}, "FF"},
		{"1 after a tokenless task fails", []string{"p1:0:1", "p2:1"}, []string{"1:p2"}, lost},
		{"1 fails already when a tokenless task is done", []string{"p1:0:1", "p2:1", "p3:2"}, []string{"2:p3"}, lost},
		{"2 after one tokenless task counts", []string{"p1:0:2", "p2:1"}, []string{"1:p2", "0:p1"}, "FF"},
		{"2 as the first of all fails", []string{"p1:0:2", "p2:1"}, []string{"0:p1"}, lost},
		{"tokenless task together with 1", []string{"p1:0:1", "p2:0"}, []string{"0:p1 p2"}, "FF"},
		{"1 and 3 together while 2 is open", []string{"p1:0:1", "p2:1:2", "p3:0:3"}, []string{"0:p1 p3"}, lost},
		{"1 and 3 together, 2 done before: 2 first is wrong", []string{"p1:0:1", "p2:1:2", "p3:0:3"}, []string{"1:p2", "0:p1 p3"}, lost},
		{"1 and 2 together, then 3", []string{"p1:0:1", "p2:0:2", "p3:1:3"}, []string{"0:p1 p2", "1:p3"}, "FFF"},
		{"1, then 2 and 3 together", []string{"p1:0:1", "p2:1:2", "p3:1:3"}, []string{"0:p1", "1:p2 p3"}, "FFF"},
		{"1, 2 and 3 all in one trick", []string{"p1:0:1", "p2:0:2", "p3:0:3"}, []string{"0:p1 p2 p3"}, "FFF"},
		{"token 3 stays open while two others are done", []string{"p1:0", "p2:1", "p3:2:3"}, []string{"0:p1", "1:p2"}, "FFO"},
		{"token 3 open, fulfilled as the third", []string{"p1:0", "p2:1", "p3:2:3"}, []string{"0:p1", "1:p2", "2:p3"}, "FFF"},
		{"token 3 fails when three others are done", []string{"p1:0", "p2:1", "p3:2:3", "p4:0"}, []string{"0:p1", "1:p2", "0:p4"}, lost},
		{"token 3 together with the two others", []string{"p1:2", "p2:2", "p3:2:3"}, []string{"2:p1 p2 p3"}, "FFF"},
		{"five tokens in order", []string{"p1:0:1", "p2:1:2", "p3:2:3", "p4:3:4", "p5:0:5"}, []string{"0:p1", "1:p2", "2:p3", "3:p4", "0:p5"}, "FFFFF"},
		{"five tokens, 3 before 2", []string{"p1:0:1", "p2:1:2", "p3:2:3", "p4:3:4", "p5:0:5"}, []string{"0:p1", "2:p3"}, lost},
		{"five tokens, 5 first", []string{"p1:0:1", "p2:1:2", "p3:2:3", "p4:3:4", "p5:0:5"}, []string{"0:p5"}, lost},
		{"tokens 1-3 together, then 4", []string{"p1:0:1", "p2:0:2", "p3:0:3", "p4:1:4"}, []string{"0:p1 p2 p3", "1:p4"}, "FFFF"},
		{"4 together with 1,2 while 3 is open", []string{"p1:0:1", "p2:0:2", "p3:1:3", "p4:0:4"}, []string{"0:p1 p2 p4"}, lost},
		{"someone else wins a token task", []string{"p1:0:1", "p2:1:2"}, []string{"1:p1"}, lost},
		{"open task with token 2 is still open after one", []string{"p1:0:1", "p2:1:2"}, []string{"0:p1"}, "FO"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bbCheckPlay(t, 4, tc.tasks, tc.tricks, tc.want)
		})
	}
}

// rules §7: the omega token (valid iff no task is still open after its trick).
func TestRulesTokenOmega(t *testing.T) {
	for _, tc := range []struct {
		name   string
		tasks  []string
		tricks []string
		want   string
	}{
		{"omega last", []string{"p1:0", "p2:1:omega"}, []string{"0:p1", "1:p2"}, "FF"},
		{"omega waits", []string{"p1:0", "p2:1:omega"}, []string{"0:p1"}, "FO"},
		{"omega before another task", []string{"p1:0", "p2:1:omega"}, []string{"1:p2"}, lost},
		{"omega together with the last open task", []string{"p1:0", "p2:0:omega"}, []string{"0:p1 p2"}, "FF"},
		{"single omega task", []string{"p1:0:omega"}, []string{"0:p1"}, "F"},
		{"omega with the others in one trick, one still open", []string{"p1:0", "p2:1", "p3:1:omega", "p4:2"}, []string{"0:p1", "1:p2 p3"}, lost},
		{"omega with the last two tasks", []string{"p1:0", "p2:1", "p3:2", "p4:2:omega"}, []string{"0:p1", "1:p2", "2:p3 p4"}, "FFFF"},
		{"omega fulfilled in a later trick with a third task open", []string{"p1:0", "p2:1:omega", "p3:2"}, []string{"0:p1", "1:p2"}, lost},
		{"omega fulfilled in the middle but not last", []string{"p1:0", "p2:1:omega", "p3:2"}, []string{"1:p2", "0:p1", "2:p3"}, lost},
		{"omega after fillers", []string{"p1:0", "p2:1:omega"}, []string{"3:", "0:p1", "2:", "1:p2"}, "FF"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bbCheckPlay(t, 4, tc.tasks, tc.tricks, tc.want)
		})
	}
}

// rules §7: arrow tokens (trick of › <= trick of ›› <= ...), tasks without an arrow are free.
func TestRulesTokenArrows(t *testing.T) {
	for _, tc := range []struct {
		name   string
		tasks  []string
		tricks []string
		want   string
	}{
		{"› before ››", []string{"p1:0:arrow1", "p2:1:arrow2"}, []string{"0:p1", "1:p2"}, "FF"},
		{"›› before ›", []string{"p1:0:arrow1", "p2:1:arrow2"}, []string{"1:p2"}, lost},
		{"› and ›› in the same trick", []string{"p1:0:arrow1", "p2:0:arrow2"}, []string{"0:p1 p2"}, "FF"},
		{"››› and › in the same trick while ›› is open", []string{"p1:0:arrow1", "p2:1:arrow2", "p3:0:arrow3"}, []string{"0:p1 p3"}, lost},
		{"› ›› ››› in order", []string{"p1:0:arrow1", "p2:1:arrow2", "p3:2:arrow3"}, []string{"0:p1", "1:p2", "2:p3"}, "FFF"},
		{"› then ›› and ››› together", []string{"p1:0:arrow1", "p2:1:arrow2", "p3:1:arrow3"}, []string{"0:p1", "1:p2 p3"}, "FFF"},
		{"all three in one trick", []string{"p1:0:arrow1", "p2:0:arrow2", "p3:0:arrow3"}, []string{"0:p1 p2 p3"}, "FFF"},
		{"tasks without an arrow are free", []string{"p1:0:arrow1", "p2:1:arrow2", "p3:2"}, []string{"2:p3", "0:p1", "1:p2"}, "FFF"},
		{"task without arrow in the middle", []string{"p1:0:arrow1", "p2:1:arrow2", "p3:2"}, []string{"0:p1", "2:p3", "1:p2"}, "FFF"},
		{"four arrows in order", []string{"p1:0:arrow1", "p2:1:arrow2", "p3:2:arrow3", "p4:3:arrow4"}, []string{"0:p1", "1:p2", "2:p3", "3:p4"}, "FFFF"},
		{"›››› before ›››", []string{"p1:0:arrow1", "p2:1:arrow2", "p3:2:arrow3", "p4:3:arrow4"}, []string{"0:p1", "1:p2", "3:p4"}, lost},
		{"›› waits for ›", []string{"p1:0:arrow1", "p2:1:arrow2"}, []string{"0:p1"}, "FO"},
		{"› waits, nothing fulfilled", []string{"p1:0:arrow1", "p2:1:arrow2"}, []string{"3:"}, "OO"},
		{"arrows 1 and 3 only, 3 first", []string{"p1:0:arrow1", "p3:1:arrow3"}, []string{"1:p3"}, lost},
		{"arrows 1 and 3 only, in order", []string{"p1:0:arrow1", "p3:1:arrow3"}, []string{"0:p1", "1:p3"}, "FF"},
		{"arrow 1 may be late", []string{"p1:0:arrow1", "p2:1:arrow2"}, []string{"3:", "3:", "0:p1", "1:p2"}, "FF"},
		{"arrow task failed by another winner", []string{"p1:0:arrow1", "p2:1:arrow2"}, []string{"1:p1"}, lost},
		{"arrows 2 and 3 in different order than 1", []string{"p1:0:arrow1", "p2:1:arrow2", "p3:2:arrow3"}, []string{"0:p1", "2:p3"}, lost},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bbCheckPlay(t, 4, tc.tasks, tc.tricks, tc.want)
		})
	}
}

// rules §6 / §7: the attempt ends as soon as a task or an order condition fails; the mission is
// completed as soon as every task is fulfilled.
func TestRulesTaskPlayOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mission int
		tasks   []string
		tricks  []string
		end     bool
		ok      bool
		reason  Reason
		after   int
	}{
		{"owner wins: completed", 1, []string{"p5:0"}, []string{bt(4, 0, "p5 p2 p3 p4")}, true, true, ReasonCompleted, 1},
		{"task card won by another: failed", 1, []string{"p5:0"}, []string{bt(4, 0, "p5 p9 p3 p4")}, true, false, ReasonTaskFailed, 1},
		{"task card played by another, owner wins it", 1, []string{"p5:0"}, []string{bt(4, 1, "p5 p3 p4 p9")}, true, true, ReasonCompleted, 1},
		{"the owner trumps the task card", 1, []string{"p5:3"}, []string{bt(4, 0, "p5 p3 p4 t1")}, true, true, ReasonCompleted, 1},
		{"someone else trumps the task card", 1, []string{"p5:0"}, []string{bt(4, 0, "p5 p3 t1 p4")}, true, false, ReasonTaskFailed, 1},
		{"a task card in a trick won by the owner of another task", 2, []string{"p5:0", "b5:1"}, []string{bt(4, 0, "p5 b1 b5 p3")}, true, false, ReasonTaskFailed, 1},
		{"two tasks one trick same owner", 2, []string{"p5:0", "p3:0"}, []string{bt(4, 0, "p5 p2 p3 p4")}, true, true, ReasonCompleted, 1},
		{"first task done, game goes on", 2, []string{"p5:0", "b5:1"}, []string{bt(4, 0, "p5 p2 p3 p4")}, false, false, "", 1},
		{"second task finishes the mission", 2, []string{"p5:0", "b5:1"}, []string{bt(4, 0, "p5 p2 p3 p4"), bt(4, 0, "b1 b5 b2 b3")}, true, true, ReasonCompleted, 2},
		{"token order: 2 before 1", 3, []string{"p5:0:2", "b5:1:1"}, []string{bt(4, 0, "p5 p2 p3 p4")}, true, false, ReasonTokenOrder, 1},
		{"token order: same trick 1+2 is fine", 3, []string{"p5:0:1", "p4:0:2"}, []string{bt(4, 0, "p5 p4 p3 p2")}, true, true, ReasonCompleted, 1},
		{"token order: omega first", 7, []string{"p5:0:omega", "b5:1"}, []string{bt(4, 0, "p5 p2 p3 p4")}, true, false, ReasonTokenOrder, 1},
		{"token order: arrows reversed", 6, []string{"p5:0:arrow2", "b5:1:arrow1"}, []string{bt(4, 0, "p5 p2 p3 p4")}, true, false, ReasonTokenOrder, 1},
		{"token order: tokenless task first makes 1 fail", 3, []string{"b5:0:1", "p4:1"}, []string{bt(4, 1, "p4 p2 p3 b1")}, true, false, ReasonTokenOrder, 1},
		{"token order kept: waits for the next trick", 3, []string{"p5:0:1", "b5:1:2"}, []string{bt(4, 0, "p5 p2 p3 p4")}, false, false, "", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := bbNew(t, 4, tc.mission, tc.tasks, tc.tricks...)
			g.Attempt.Objectives = nil
			bbPlayTricks(t, g, tc.tricks...)
			if tc.end {
				bbExpectEnd(t, g, tc.ok, tc.reason, tc.after)
			} else {
				bbExpectRunning(t, g, tc.after)
			}
		})
	}
}

// rules §6: a task that is still open when the last trick is over fails (also an unplayed card).
func TestRulesTaskOpenAfterLastTrick(t *testing.T) {
	for _, n := range []int{3, 4, 5} {
		for _, tc := range []struct {
			name   string
			last   string // the cards of the last trick, leader first; the leader is seat 0
			tasks  []string
			ok     bool
			reason Reason
		}{
			{"task card never played", "b1 b2 b3 b4 b5", []string{"p9:0"}, false, ReasonTaskFailed},
			{"task card played in the last trick, owner wins", "b6 b1 b2 b3 b4", []string{"b6:0"}, true, ReasonCompleted},
			{"task card played in the last trick, other wins", "b1 b6 b2 b3 b4", []string{"b6:0"}, false, ReasonTaskFailed},
			{"one task done earlier, one never played", "b1 b2 b3 b4 b5", []string{"y8:0", "p9:0"}, false, ReasonTaskFailed},
		} {
			t.Run(tc.name+" "+strconv.Itoa(n)+"p", func(t *testing.T) {
				g := startMission(t, n, 2)
				setTasks(g, tc.tasks...)
				total := TotalTricks(n)
				bbHistory(g, bbRecs(n, bbSpecs(n, total-1, "0")...)...)
				// the card "y8:0" was won by seat 0 in the filler tricks (it is the leader's card)
				a := g.Attempt
				fs := splitFields(tc.last)
				for i := 0; i < n; i++ {
					a.Members[i].Hand = cs(fs[i])
				}
				a.Commander = 0
				a.Trick = Trick{Leader: 0}
				for i := range n {
					must(t, g.PlayCard(g.Players[i].ID, c(fs[i]), noShuffle{}))
				}
				bbExpectEnd(t, g, tc.ok, tc.reason, total)
			})
		}
	}
}
