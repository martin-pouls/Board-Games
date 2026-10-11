package domain

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const (
	alice PlayerID = "alice" // seat 0, host
	bob   PlayerID = "bob"   // seat 1
	carol PlayerID = "carol" // seat 2
	dave  PlayerID = "dave"  // seat 3
	erin  PlayerID = "erin"  // seat 4
)

// crewIDs are the players of newCrew in seat order.
var crewIDs = []PlayerID{alice, bob, carol, dave, erin}

// noShuffle is a deterministic RNG: it leaves every deck in order and IntN returns 0.
type noShuffle struct{}

func (noShuffle) Shuffle(int, func(i, j int)) {}
func (noShuffle) IntN(int) int                { return 0 }

// fixedIntN leaves decks in order and returns IntN values from the list (then 0).
type fixedIntN struct{ vals []int }

func (f *fixedIntN) Shuffle(int, func(i, j int)) {}
func (f *fixedIntN) IntN(n int) int {
	if len(f.vals) == 0 {
		return 0
	}
	v := f.vals[0]
	f.vals = f.vals[1:]
	return v % n
}

var suitLetters = map[byte]Suit{'p': SuitPink, 'b': SuitBlue, 'g': SuitGreen, 'y': SuitYellow, 't': SuitTrump}

// c parses a card: the suit letter p (pink), b (blue), g (green), y (yellow) or t (trump),
// followed by the rank, e.g. "p7", "t4".
func c(s string) Card {
	suit, ok := suitLetters[s[0]]
	if !ok {
		panic("bad suit in " + s)
	}
	n, err := strconv.Atoi(s[1:])
	card := NewCard(suit, Rank(n))
	if err != nil || !card.Valid() {
		panic("bad card " + s)
	}
	return card
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

// newLobbyGame returns a lobby with alice (host) and bob seated.
func newLobbyGame(t *testing.T) *Game {
	t.Helper()
	return newCrew(t, 2)
}

// newCrew returns a lobby with the first n of alice, bob, carol, dave and erin (host alice).
func newCrew(t *testing.T, n int) *Game {
	t.Helper()
	g, err := NewGame("GAME1", alice, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < n; i++ {
		id := crewIDs[i]
		if err := g.AddPlayer(id, strings.ToUpper(string(id[:1]))+string(id[1:])); err != nil {
			t.Fatal(err)
		}
	}
	g.PullEvents()
	return g
}

// startMission returns a game of n players playing mission with noShuffle (pending events
// cleared). With noShuffle the deal is fixed; tests then overwrite what they need.
func startMission(t *testing.T, n, mission int) *Game {
	t.Helper()
	g := newCrew(t, n)
	must(t, g.SelectMission(alice, mission))
	must(t, g.StartMission(alice, noShuffle{}))
	g.PullEvents()
	return g
}

// setHands overwrites the humans' hands (seat order) and makes the holder of t4 the commander.
// Empty strings leave a hand as it is.
func setHands(g *Game, hands ...string) {
	a := g.Attempt
	for i, h := range hands {
		if h != "" {
			a.Members[i].Hand = cs(h)
		}
	}
	for i := range a.Members {
		if a.Members[i].holds(CommanderCard) {
			a.Commander = i
		}
	}
}

// setDummy overwrites the dummy's stacks: up and down are space-separated cards per stack, "-"
// for an empty position. Both lists have DummyStacks entries.
func setDummy(g *Game, up, down string) {
	a := g.Attempt
	ups, downs := strings.Fields(up), strings.Fields(down)
	for i := range a.Members {
		if !a.Members[i].Dummy {
			continue
		}
		a.Members[i].Stacks = make([]DummyStack, DummyStacks)
		for k := range DummyStacks {
			if ups[k] != "-" {
				a.Members[i].Stacks[k].Up = cp(ups[k])
			}
			if downs[k] != "-" {
				a.Members[i].Stacks[k].Down = cp(downs[k])
			}
		}
	}
}

// setTasks replaces the tasks: each entry is "card:owner" or "card:owner:token", owner a seat
// (or "-" for unassigned). The attempt moves to StepPlay with the commander leading trick 1.
func setTasks(g *Game, tasks ...string) {
	a := g.Attempt
	a.Tasks = nil
	for i, s := range tasks {
		f := strings.Split(s, ":")
		owner := OwnerUnassigned
		if f[1] != "-" {
			owner, _ = strconv.Atoi(f[1])
		}
		t := Task{ID: i, Objective: Objective{Kind: KindWinCards, Owner: owner, Cards: []Card{c(f[0])}}}
		if len(f) > 2 {
			t.Token = Token(f[2])
		}
		a.Tasks = append(a.Tasks, t)
	}
	toPlay(g)
}

// toPlay puts the attempt at the start of trick 1, led by the commander.
func toPlay(g *Game) {
	a := g.Attempt
	a.Step = StepPlay
	a.Poll = nil
	a.Trick = Trick{Leader: a.Commander}
	a.Tricks = nil
}

// actorID returns the player acting for the seat whose turn it is.
func actorID(g *Game) PlayerID {
	a := g.Attempt
	return a.Members[a.actorFor(a.turn())].PlayerID
}

// playAll plays the cards in turn order, each by whoever acts for the seat to play.
func playAll(t *testing.T, g *Game, cards string) {
	t.Helper()
	for _, f := range strings.Fields(cards) {
		if err := g.PlayCard(actorID(g), c(f), noShuffle{}); err != nil {
			t.Fatalf("play %s: %v", f, err)
		}
	}
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

// totalCards counts every card of the attempt: hands, dummy, the trick in progress and the
// completed tricks.
func totalCards(g *Game) int {
	a := g.Attempt
	n := len(a.Trick.Plays)
	for i := range a.Members {
		n += a.Members[i].CardCount()
	}
	for _, t := range a.Tricks {
		n += len(t.Plays)
	}
	return n
}

// trick builds a completed trick record: leader, then "seat:card" plays; the winner is computed.
func trick(plays ...string) TrickRecord {
	r := TrickRecord{}
	var cards []Card
	for i, p := range plays {
		f := strings.Split(p, ":")
		seat, err := strconv.Atoi(f[0])
		if err != nil {
			panic(fmt.Sprint("bad play ", p))
		}
		if i == 0 {
			r.Leader = seat
		}
		r.Plays = append(r.Plays, Play{Seat: seat, Card: c(f[1])})
		cards = append(cards, c(f[1]))
	}
	r.Winner = r.Plays[trickWinner(cards)].Seat
	return r
}
