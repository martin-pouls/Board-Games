package domain

import (
	"strconv"
	"strings"
	"testing"
)

// bbPlaySpecs plays tricks through the engine. A spec is "winner:cards": the winner leads y8 (or the
// card written first with a * prefix, e.g. "2:*t3 t1") and wins; the listed cards go to the other
// seats in turn order, free places get low yellow cards. Before each trick every member's hand is
// reset to the one card they play, so follow suit never interferes. It stops when the
// attempt is over and returns the number of tricks played.
func bbPlaySpecs(t *testing.T, g *Game, specs ...string) int {
	t.Helper()
	a := g.Attempt
	n := len(a.Members)
	done := 0
	for _, s := range specs {
		if g.Phase != PhasePlaying {
			return done
		}
		ws, cards, _ := strings.Cut(s, ":")
		w, err := strconv.Atoi(ws)
		if err != nil {
			t.Fatalf("bad spec %q", s)
		}
		leader := w // the winner leads: its card wins whatever it is
		a.Trick = Trick{Leader: w}
		listed := strings.Fields(cards)
		play := make([]string, n) // by seat
		play[w] = "y8"
		if len(listed) > 0 && strings.HasPrefix(listed[0], "*") {
			play[w], listed = listed[0][1:], listed[1:]
		}
		free := 2
		for k := 0; k < n; k++ {
			seat := (leader + k) % n
			if seat == w {
				continue
			}
			if len(listed) > 0 {
				play[seat], listed = listed[0], listed[1:]
			} else {
				play[seat] = "y" + strconv.Itoa(free)
				free++
			}
		}
		if len(listed) > 0 {
			t.Fatalf("spec %q has too many cards", s)
		}
		for seat := range n {
			if a.Members[seat].Dummy {
				a.Members[seat].Stacks = []DummyStack{{Up: cp(play[seat])}}
				continue
			}
			a.Members[seat].Hand = cs(play[seat])
		}
		for k := 0; k < n; k++ {
			seat := (leader + k) % n
			if err := g.PlayCard(bbActor(g, seat), c(play[seat]), noShuffle{}); err != nil {
				t.Fatalf("trick %d, seat %d plays %s: %v", done+1, seat, play[seat], err)
			}
		}
		done++
	}
	return done
}

// bbStatusString renders task statuses: O open, F fulfilled, X failed.
func bbStatusString(ts []Status) string {
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

// bbCheckPlay starts mission 3 (n >= 3 players, no objectives), sets the tasks, plays the
// specs and checks the outcome. want is "LOST" (the attempt is lost, whatever the reason), a
// string of F only (the mission is completed after the last spec) or a string of F/O (still
// running with exactly these task statuses).
func bbCheckPlay(t *testing.T, n int, tasks, specs []string, want string) {
	t.Helper()
	g := startMission(t, n, 3)
	g.Attempt.Objectives = nil
	setTasks(g, tasks...)
	played := bbPlaySpecs(t, g, specs...)
	a := g.Attempt
	switch {
	case want == "LOST":
		if g.Phase != PhaseDebrief || a.Result == nil || a.Result.Success {
			ts, _ := a.Statuses()
			t.Fatalf("the attempt must be lost: phase %s result %+v statuses %s", g.Phase, a.Result, bbStatusString(ts))
		}
	case !strings.ContainsAny(want, "O"):
		if g.Phase != PhaseDebrief || a.Result == nil || !a.Result.Success || played != len(specs) {
			t.Fatalf("the mission must be completed after %d tricks: phase %s result %+v (played %d)", len(specs), g.Phase, a.Result, played)
		}
	default:
		if g.Phase != PhasePlaying || played != len(specs) {
			t.Fatalf("the attempt must still run: phase %s result %+v (played %d of %d)", g.Phase, a.Result, played, len(specs))
		}
		ts, _ := a.Statuses()
		if got := bbStatusString(ts); got != want {
			t.Fatalf("statuses %q, want %q", got, want)
		}
	}
}
