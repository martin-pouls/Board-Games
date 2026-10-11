package domain

// Helpers of the black-box rule tests (rules_*_test.go). They only use the public API, the
// exported state and helpers_test.go. Names start with bb to stay clear of other test files.

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// bbFailAny asserts that fn fails with one of wants and leaves g unchanged.
func bbFailAny(t *testing.T, g *Game, fn func(*Game) error, wants ...error) {
	t.Helper()
	before := g.Clone()
	err := fn(g)
	if err == nil {
		t.Fatalf("expected one of %v, got no error", wants)
	}
	ok := false
	for _, w := range wants {
		if errors.Is(err, w) {
			ok = true
		}
	}
	if !ok {
		t.Fatalf("err = %v, want one of %v", err, wants)
	}
	if !reflect.DeepEqual(before, g) {
		t.Fatalf("state changed on error %v", err)
	}
}

// bt builds a trick string "seat:card ..." for n seats: the cards are played by the seats
// leader, leader+1, ... (wrapping).
func bt(n, leader int, cards string) string {
	var out []string
	for i, f := range strings.Fields(cards) {
		out = append(out, fmt.Sprintf("%d:%s", (leader+i)%n, f))
	}
	return strings.Join(out, " ")
}

// bbPlays parses "seat:card seat:card ...".
func bbPlays(s string) []Play {
	var out []Play
	for _, f := range strings.Fields(s) {
		p := strings.Split(f, ":")
		seat, err := strconv.Atoi(p[0])
		if err != nil {
			panic("bad play " + f)
		}
		out = append(out, Play{Seat: seat, Card: c(p[1])})
	}
	return out
}

// bbNew starts a mission with n human players (no dummy: n >= 3) and replaces the deal by what
// the tricks need: every member's hand is exactly the cards they play in the tricks, the
// commander is the leader of the first trick and the tasks are set with setTasks. The attempt
// waits at the lead of trick 1.
func bbNew(t *testing.T, n, mission int, tasks []string, tricks ...string) *Game {
	t.Helper()
	g := startMission(t, n, mission)
	a := g.Attempt
	hands := make([]string, n)
	first := -1
	for _, tr := range tricks {
		for _, p := range bbPlays(tr) {
			if first < 0 {
				first = p.Seat
			}
			hands[p.Seat] += " " + p.Card.shortString()
		}
	}
	for i := range hands {
		a.Members[i].Hand = cs(hands[i])
	}
	setTasks(g, tasks...)
	if first >= 0 {
		a.Commander = first
		a.Trick = Trick{Leader: first}
	}
	g.PullEvents()
	return g
}

// shortString renders a card as the helper letter and rank, e.g. "p7".
func (c Card) shortString() string {
	for l, s := range suitLetters {
		if s == c.Suit {
			return string(l) + strconv.Itoa(int(c.Rank))
		}
	}
	return "?"
}

// bbPlayTricks plays the tricks in order, each card by the player of its seat. It stops without
// an error as soon as the attempt is over and returns the number of cards played.
func bbPlayTricks(t *testing.T, g *Game, tricks ...string) int {
	t.Helper()
	n := 0
	for _, tr := range tricks {
		for _, p := range bbPlays(tr) {
			if g.Phase != PhasePlaying {
				return n
			}
			by := g.Attempt.Members[p.Seat].PlayerID
			if err := g.PlayCard(by, p.Card, noShuffle{}); err != nil {
				t.Fatalf("seat %d plays %v (card %d): %v", p.Seat, p.Card, n+1, err)
			}
			n++
		}
	}
	return n
}

// bbExpectEnd asserts that the attempt is over with the given result after the given number of
// completed tricks.
func bbExpectEnd(t *testing.T, g *Game, success bool, reason Reason, tricks int) {
	t.Helper()
	if g.Phase != PhaseDebrief || g.Attempt.Result == nil {
		t.Fatalf("attempt not over: phase %s, tricks %d", g.Phase, len(g.Attempt.Tricks))
	}
	r := g.Attempt.Result
	if r.Success != success || r.Reason != reason {
		t.Fatalf("result %+v, want success=%v reason=%s", *r, success, reason)
	}
	if tricks >= 0 && len(g.Attempt.Tricks) != tricks {
		t.Fatalf("ended after %d tricks, want %d", len(g.Attempt.Tricks), tricks)
	}
}

// bbExpectRunning asserts that the attempt is still being played after the given number of tricks.
func bbExpectRunning(t *testing.T, g *Game, tricks int) {
	t.Helper()
	if g.Phase != PhasePlaying || g.Attempt.Result != nil {
		t.Fatalf("attempt over: phase %s result %+v", g.Phase, g.Attempt.Result)
	}
	if len(g.Attempt.Tricks) != tricks {
		t.Fatalf("%d tricks played, want %d", len(g.Attempt.Tricks), tricks)
	}
}

// bbFiller is a neutral completed trick won by winner: the winner plays y8, the others low yellow
// cards. (Keep yellow out of the test's tasks, and rank 1 and 9 out of the objectives.)
func bbFiller(seats, winner int) TrickRecord {
	r := TrickRecord{Leader: winner, Winner: winner}
	r.Plays = append(r.Plays, Play{Seat: winner, Card: c("y8")})
	for i := 1; i < seats; i++ {
		r.Plays = append(r.Plays, Play{Seat: (winner + i) % seats, Card: c("y" + strconv.Itoa(1+i))})
	}
	return r
}

// bbFillers returns neutral tricks won by the given seats in turn.
func bbFillers(seats int, winners ...int) []TrickRecord {
	var out []TrickRecord
	for _, w := range winners {
		out = append(out, bbFiller(seats, w))
	}
	return out
}

// bbHistory replaces the completed tricks and makes the winner of the last one lead the next.
func bbHistory(g *Game, recs ...TrickRecord) {
	a := g.Attempt
	a.Tricks = slices.Clone(recs)
	a.Trick = Trick{Leader: recs[len(recs)-1].Winner}
}

// bbFinalTrick sets up the last trick of an attempt: history is the completed tricks (it must be
// TotalTricks-1 long) and hands are the last cards ("" leaves a hand as it is).
func bbFinalTrick(t *testing.T, g *Game, hands []string, history ...TrickRecord) {
	t.Helper()
	a := g.Attempt
	humans := len(g.Players)
	seats := len(a.Members)
	if len(history) != TotalTricks(humans)-1 {
		t.Fatalf("history of %d tricks, want %d", len(history), TotalTricks(humans)-1)
	}
	for i, h := range hands {
		if h != "" {
			a.Members[i].Hand = cs(h)
		}
	}
	_ = seats
	bbHistory(g, history...)
}

// bbSeatOf returns the player acting for seat in g's attempt.
func bbActor(g *Game, seat int) PlayerID {
	a := g.Attempt
	return a.Members[a.actorFor(seat)].PlayerID
}

// bbHas reports whether cards contains the card.
func bbHas(cards []Card, s string) bool { return slices.Contains(cards, c(s)) }

// bbCardsEqual compares card lists ignoring order.
func bbCardsEqual(a, b []Card) bool {
	x, y := slices.Clone(a), slices.Clone(b)
	SortCards(x)
	SortCards(y)
	return slices.Equal(x, y)
}

// bbDraft completes the open draft of a started game: every pick is made by the actor in turn,
// taking the lowest free task. It returns the owner by task ID.
func bbDraft(t *testing.T, g *Game) {
	t.Helper()
	a := g.Attempt
	for a.Step == StepDraft {
		seat := a.DraftTurn()
		id := -1
		for _, tk := range a.Tasks {
			if tk.Owner() == OwnerUnassigned {
				id = tk.ID
				break
			}
		}
		if id < 0 {
			t.Fatal("draft step without a free task")
		}
		by := a.Members[a.actorFor(seat)].PlayerID
		if err := g.PickTask(by, id); err != nil {
			t.Fatalf("pick task %d for seat %d: %v", id, seat, err)
		}
	}
}
