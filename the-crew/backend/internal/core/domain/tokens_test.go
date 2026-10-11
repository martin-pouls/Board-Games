package domain

import (
	"strings"
	"testing"
)

// tokenCase is a set of tasks given as "token@trick" ("-" = no token, trick 0 = open) and the
// index of the task whose order condition breaks (-1 = none).
type tokenCase struct {
	name  string
	tasks string
	want  int
}

func parseTokenCase(s string) ([]Token, []int) {
	var tokens []Token
	var at []int
	for _, f := range strings.Fields(s) {
		parts := strings.Split(f, "@")
		tok := Token(parts[0])
		if parts[0] == "-" {
			tok = TokenNone
		}
		n := 0
		for _, r := range parts[1] {
			n = n*10 + int(r-'0')
		}
		tokens = append(tokens, tok)
		at = append(at, n)
	}
	return tokens, at
}

func TestOrderViolation(t *testing.T) {
	for _, tc := range []tokenCase{
		// Numbered tokens (rules §7): k-th of all tasks fulfilled.
		{"1 2 in order", "1@1 2@3", -1},
		{"2 before 1: the open 1 breaks", "1@3 2@1", 0},
		{"1 first of all tasks, others count", "1@2 -@1", 0},
		{"1 then untokened", "1@1 -@2", -1},
		{"open 1 fails once another is done", "1@0 -@1", 0},
		{"open 2 fine after one other", "2@0 -@1", -1},
		{"open 2 fails after two others", "2@0 -@1 -@2", 0},
		{"1 and 2 in the same trick", "1@2 2@2", -1},
		{"1 2 3 same trick", "1@1 2@1 3@1", -1},
		{"2 with an untokened in the same trick, nothing before", "2@1 -@1", -1},
		{"1 with two others in the same trick", "1@1 -@1 -@1", -1},
		{"3 in a trick with 1 other after 1 before", "3@2 -@1 -@2", -1},
		{"3 third after two earlier", "3@3 -@1 -@2", -1},
		{"3 too late", "3@4 -@1 -@2 -@3", 0},
		{"3 too early", "3@1 -@2 -@3", 0},
		{"1 and 2 in the same trick after another task", "1@2 2@2 -@1 -@3", 0}, // 1 needs b < 1
		{"1 2 3 4 5 all fine", "1@1 2@2 3@3 4@4 5@5", -1},
		{"4 and 5 swapped", "1@1 2@2 3@3 4@5 5@4", 3},
		// Ω: the last task fulfilled.
		{"omega last", "omega@3 -@1 -@2", -1},
		{"omega with the last others", "omega@2 -@1 -@2", -1},
		{"omega while another is open", "omega@1 -@0", 0},
		{"omega before another", "omega@1 -@2", 0},
		{"omega open is fine", "omega@0 -@1", -1},
		{"1 and omega", "1@1 omega@2 -@2", -1},
		{"1 and omega same trick, only two tasks", "1@1 omega@1", -1},
		// Arrows: relative order, ties allowed.
		{"arrows in order", "arrow1@1 arrow2@2 arrow3@4", -1},
		{"arrows same trick", "arrow1@2 arrow2@2", -1},
		{"arrow2 before arrow1", "arrow1@3 arrow2@2", 0},
		{"arrow2 done while arrow1 open", "arrow1@0 arrow2@2", 0},
		{"arrow3 done while arrow1 open", "arrow1@0 arrow2@0 arrow3@1", 0},
		{"arrow1 open, nothing later done", "arrow1@0 arrow2@0 -@1", -1},
		{"untokened tasks do not matter for arrows", "-@1 arrow1@2 -@3 arrow2@4", -1},
		{"four arrows, last two swapped", "arrow1@1 arrow2@2 arrow3@4 arrow4@3", 2},
		{"no tokens", "-@0 -@1", -1},
	} {
		tokens, at := parseTokenCase(tc.tasks)
		if got := orderViolation(tokens, at); got != tc.want {
			t.Errorf("%s (%s): got %d, want %d", tc.name, tc.tasks, got, tc.want)
		}
	}
}

func TestTokenHelpers(t *testing.T) {
	if Token3.Number() != 3 || TokenArrow2.Arrow() != 2 || TokenOmega.Number() != 0 || TokenNone.Arrow() != 0 {
		t.Fatal("token numbers")
	}
	for _, tok := range []Token{TokenNone, Token1, Token5, TokenOmega, TokenArrow4} {
		if !tok.Valid() {
			t.Fatalf("%q must be valid", tok)
		}
	}
	if Token("6").Valid() || Token("arrow5").Valid() {
		t.Fatal("invalid tokens accepted")
	}
}

func TestTokenOrderEndsAttempt(t *testing.T) {
	g := startMission(t, 4, 3)
	setHands(g, "t4 p1 p2 b1", "p3 b2 b3 g1", "g2 g3 g4 y1", "y2 y3 t1 t2")
	setTasks(g, "p3:0:1", "g4:0:2")
	playAll(t, g, "b1 b2 g4 y2") // bob wins with b2: g4 was won by bob, not alice: task fails
	r := g.Attempt.Result
	if g.Phase != PhaseDebrief || r.Reason != ReasonTaskFailed || r.Task != 1 || r.Seat != 1 {
		t.Fatalf("phase %s result %+v", g.Phase, r)
	}

	g = startMission(t, 4, 3)
	setHands(g, "t4 p1 p2 g9", "p3 b2 b3 g1", "g2 g3 g4 y1", "y2 y3 t1 t2")
	setTasks(g, "p3:0:1", "g4:0:2")
	playAll(t, g, "g9 g1 g4 y2") // alice wins g4 (token 2) before p3 (token 1)
	r = g.Attempt.Result
	if r == nil || r.Reason != ReasonTokenOrder || r.Task != 0 {
		t.Fatalf("result %+v", r)
	}
}
