package domain

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// noShuffle is a deterministic RNG that leaves every deck in order.
type noShuffle struct{}

func (noShuffle) Shuffle(int, func(i, j int)) {}

// reverseShuffle reverses every slice, to prove a shuffle really uses the RNG.
type reverseShuffle struct{}

func (reverseShuffle) Shuffle(n int, swap func(i, j int)) {
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

var suitLetters = map[byte]Suit{'D': SuitDoves, 'R': SuitRoses, 'S': SuitStars}

// c parses a card: rank 1..10 followed by the suit D (doves), R (roses) or S (stars).
func c(s string) Card {
	suit, ok := suitLetters[s[len(s)-1]]
	if !ok {
		panic("bad suit in " + s)
	}
	n, err := strconv.Atoi(s[:len(s)-1])
	if err != nil || n < 1 || n > 10 {
		panic("bad rank in " + s)
	}
	return NewCard(suit, Rank(n))
}

// cp returns a pointer to the parsed card.
func cp(s string) *Card {
	v := c(s)
	return &v
}

// cs parses space-separated cards into a sorted, non-nil slice.
func cs(s string) []Card {
	out := []Card{}
	for _, f := range strings.Fields(s) {
		out = append(out, c(f))
	}
	SortCards(out)
	return out
}

const (
	alice PlayerID = "alice" // seat 0, host, left end
	bob   PlayerID = "bob"   // seat 1, right end
)

// newLobbyGame returns a lobby with alice (host) and bob seated.
func newLobbyGame(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame("GAME1", alice, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer(bob, "Bob"); err != nil {
		t.Fatal(err)
	}
	g.PullEvents()
	return g
}

// newPlaying returns a started game at level with noShuffle: alice (seat 0) deals round 1, so
// bob leads. Tests then overwrite hands, decree and board as they need.
func newPlaying(t *testing.T, level int) *Game {
	t.Helper()
	g := newLobbyGame(t)
	if err := g.Start(alice, level, noShuffle{}); err != nil {
		t.Fatal(err)
	}
	g.PullEvents()
	return g
}

// setup puts the hands and the decree in place and starts trick number with leader leading.
func setup(g *Game, hand0, hand1, decree string, leader, number int) {
	g.Players[0].Hand = cs(hand0)
	g.Players[1].Hand = cs(hand1)
	g.Decree = cp(decree)
	g.Trick = Trick{Number: number, Leader: leader}
	g.Step = StepLead
	g.LastTrick = nil
	g.LastGift = nil
}

// must fails the test on err.
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// mustFail asserts that err matches want and that fn left g unchanged (validate before mutate).
func mustFail(t *testing.T, g *Game, want error, fn func(*Game) error) {
	t.Helper()
	before := g.Clone()
	err := fn(g)
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
	if !reflect.DeepEqual(before, g) {
		t.Fatalf("state changed on error %v", err)
	}
}

// play plays a card and fails the test on error.
func play(t *testing.T, g *Game, by PlayerID, card string) {
	t.Helper()
	must(t, g.PlayCard(by, c(card), noShuffle{}))
}

// eventsOf returns the pending events of type T.
func eventsOf[T Event](g *Game) []T {
	var out []T
	for _, e := range g.events {
		if v, ok := e.(T); ok {
			out = append(out, v)
		}
	}
	return out
}

// totalCards counts every card the game still tracks: hands, deck, decree and the trick.
func totalCards(g *Game) int {
	n := len(g.Deck)
	for _, p := range g.Players {
		n += len(p.Hand)
	}
	if g.Decree != nil {
		n++
	}
	if g.Trick.Lead != nil {
		n++
	}
	if g.Trick.Follow != nil {
		n++
	}
	return n
}
