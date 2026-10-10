package domain

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// noShuffle is a deterministic RNG that leaves every deck in order.
type noShuffle struct{}

func (noShuffle) Shuffle(int, func(i, j int)) {}

// reverseShuffle reverses every deck, to prove a shuffle really uses the RNG.
type reverseShuffle struct{}

func (reverseShuffle) Shuffle(n int, swap func(i, j int)) {
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

// countingRNG records the n of every Shuffle call and reverses the decks of the calls listed in
// reverse (0-based call index).
type countingRNG struct {
	calls   []int
	reverse map[int]bool
}

func (r *countingRNG) Shuffle(n int, swap func(i, j int)) {
	if r.reverse[len(r.calls)] {
		reverseShuffle{}.Shuffle(n, swap)
	}
	r.calls = append(r.calls, n)
}

var suitLetters = map[byte]Suit{'H': SuitHearts, 'D': SuitDiamonds, 'C': SuitClubs, 'S': SuitSpades}

// c parses a card: rank A, 2..10, J, Q, K followed by the suit H/D/C/S; "X" is the Jester.
func c(s string) Card {
	if s == "X" {
		return Jester()
	}
	suit, ok := suitLetters[s[len(s)-1]]
	if !ok {
		panic("bad suit in " + s)
	}
	var r Rank
	switch rs := s[:len(s)-1]; rs {
	case "A":
		r = RankAce
	case "J":
		r = RankJack
	case "Q":
		r = RankQueen
	case "K":
		r = RankKing
	default:
		n, err := strconv.Atoi(rs)
		if err != nil || n < 2 || n > 10 {
			panic("bad rank in " + s)
		}
		r = Rank(n)
	}
	return NewCard(suit, r)
}

// cs parses space-separated cards; "" is an empty, non-nil slice.
func cs(s string) []Card {
	out := []Card{}
	for _, f := range strings.Fields(s) {
		out = append(out, c(f))
	}
	return out
}

// sorted returns a sorted copy of cs("...").
func sorted(s string) []Card {
	out := cs(s)
	SortCards(out)
	return out
}

func pid(seat int) PlayerID { return PlayerID(fmt.Sprintf("p%d", seat)) }

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func wantErr(t *testing.T, err, sentinel error) {
	t.Helper()
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want %v", err, sentinel)
	}
}

// expectErr runs op, checks that it fails with want and that g is unchanged: same state,
// same Revision, no new events (atomicity).
func expectErr(t *testing.T, g *Game, want error, op func(g *Game) error) {
	t.Helper()
	before := g.Clone()
	err := op(g)
	wantErr(t, err, want)
	if !reflect.DeepEqual(before, g) {
		t.Fatalf("state changed on error %v:\nbefore %+v\nafter  %+v", err, before, g)
	}
}

// played is a set already played against the enemy, by seat.
type played struct {
	seat  int
	cards string
}

// table describes a game in progress for scenario tests. All card lists use cs notation;
// decks are listed bottom .. top.
type table struct {
	hands    []string // per seat (sorted on build); len = player count
	enemy    string   // "" = no enemy
	castle   string
	tavern   string
	discard  string
	played   []played
	seat     int  // active seat
	step     Step // default StepPlay
	yielded  []int
	soloUsed int
}

// build creates a game in PhasePlaying (Revision 1, Turn 1, no events).
func (tb table) build(t *testing.T) *Game {
	t.Helper()
	players := make([]Player, len(tb.hands))
	for i, h := range tb.hands {
		players[i] = Player{ID: pid(i), Name: fmt.Sprintf("Player %d", i), Hand: sorted(h)}
	}
	for _, s := range tb.yielded {
		players[s].YieldedLastTurn = true
	}
	var enemy *Card
	if tb.enemy != "" {
		e := c(tb.enemy)
		enemy = &e
	}
	var ps []PlayedSet
	for _, p := range tb.played {
		ps = append(ps, PlayedSet{By: pid(p.seat), Cards: sorted(p.cards)})
	}
	step := tb.step
	if step == StepNone {
		step = StepPlay
	}
	return &Game{
		ID:              "TEST2",
		Phase:           PhasePlaying,
		Step:            step,
		HostID:          pid(0),
		Players:         players,
		Castle:          cs(tb.castle),
		Enemy:           enemy,
		Tavern:          cs(tb.tavern),
		DiscardPile:     cs(tb.discard),
		Played:          ps,
		Turn:            Turn{Number: 1, ActiveSeat: tb.seat},
		SoloJestersUsed: tb.soloUsed,
		Revision:        1,
	}
}

// lobbyWith creates a lobby game with n players p0..p(n-1).
func lobbyWith(t *testing.T, n int) *Game {
	t.Helper()
	g, err := NewGame("LOBBY", pid(0), "Player 0")
	mustOK(t, err)
	for i := 1; i < n; i++ {
		mustOK(t, g.AddPlayer(pid(i), fmt.Sprintf("Player %d", i)))
	}
	g.PullEvents()
	return g
}

// started creates a started game with n players (first = seat 0, noShuffle), events pulled.
func started(t *testing.T, n int) *Game {
	t.Helper()
	g := lobbyWith(t, n)
	mustOK(t, g.Start(pid(0), pid(0), noShuffle{}))
	g.PullEvents()
	return g
}

// allCards returns every card of the game in any zone.
func allCards(g *Game) []Card {
	var out []Card
	out = append(out, g.Tavern...)
	out = append(out, g.Castle...)
	if g.Enemy != nil {
		out = append(out, *g.Enemy)
	}
	out = append(out, g.DiscardPile...)
	for _, ps := range g.Played {
		out = append(out, ps.Cards...)
	}
	for _, p := range g.Players {
		out = append(out, p.Hand...)
	}
	return out
}

// checkShape checks the invariants that hold for every game, including hand-built tables:
// sorted hands within the max size, Step/Phase consistency, Result iff over.
func checkShape(t *testing.T, g *Game) {
	t.Helper()
	maxHand := g.Rules().MaxHandSize
	for _, p := range g.Players {
		if p.Hand == nil {
			t.Fatalf("%s: nil hand", p.ID)
		}
		if !slices.IsSortedFunc(p.Hand, compareCards) {
			t.Fatalf("%s: hand not sorted: %v", p.ID, p.Hand)
		}
		if g.Phase != PhaseLobby && len(p.Hand) > maxHand {
			t.Fatalf("%s: %d cards > max %d", p.ID, len(p.Hand), maxHand)
		}
	}
	switch g.Phase {
	case PhaseLobby:
		if g.Step != StepNone || g.Result != nil {
			t.Fatalf("lobby: step %q result %+v", g.Step, g.Result)
		}
	case PhasePlaying:
		if g.Step != StepPlay && g.Step != StepDiscard && g.Step != StepChooseNext {
			t.Fatalf("playing: step %q", g.Step)
		}
		if g.Enemy == nil || g.Result != nil {
			t.Fatalf("playing: enemy %v result %+v", g.Enemy, g.Result)
		}
		if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Players) {
			t.Fatalf("playing: active seat %d", g.Turn.ActiveSeat)
		}
		if g.Step == StepDiscard && g.EffectiveAttack() == 0 {
			t.Fatal("discard step with effective attack 0")
		}
	case PhaseWon:
		if g.Step != StepNone || g.Result == nil || *g.Result != (Result{}) || g.Enemy != nil || len(g.Castle) != 0 {
			t.Fatalf("won: step %q result %+v enemy %v castle %d", g.Step, g.Result, g.Enemy, len(g.Castle))
		}
	case PhaseLost:
		if g.Step != StepNone || g.Result == nil || g.Result.Reason == "" || g.Enemy == nil {
			t.Fatalf("lost: step %q result %+v enemy %v", g.Step, g.Result, g.Enemy)
		}
	default:
		t.Fatalf("phase %q", g.Phase)
	}
	if g.SoloJestersLeft() < 0 || (!g.IsSolo() && g.SoloJestersUsed != 0) {
		t.Fatalf("solo jesters used %d", g.SoloJestersUsed)
	}
}

// checkInvariants checks checkShape plus card conservation for games set up by Start:
// 52 + TavernJesters cards in total, all distinct except the Jesters.
func checkInvariants(t *testing.T, g *Game) {
	t.Helper()
	checkShape(t, g)
	all := allCards(g)
	if want := 52 + g.Rules().TavernJesters; len(all) != want {
		t.Fatalf("%d cards in play, want %d", len(all), want)
	}
	seen := map[Card]bool{}
	jesters := 0
	for _, x := range all {
		if !x.Valid() {
			t.Fatalf("invalid card %v", x)
		}
		if x.IsJester() {
			jesters++
			continue
		}
		if seen[x] {
			t.Fatalf("duplicate card %v", x)
		}
		seen[x] = true
	}
	if jesters != g.Rules().TavernJesters {
		t.Fatalf("%d jesters, want %d", jesters, g.Rules().TavernJesters)
	}
}

// eventTypes returns the type names of evs, for compact sequence assertions.
func eventTypes(evs []Event) []string {
	out := make([]string, len(evs))
	for i, e := range evs {
		out[i] = reflect.TypeOf(e).Name()
	}
	return out
}

func wantEventTypes(t *testing.T, g *Game, want ...string) []Event {
	t.Helper()
	evs := g.PullEvents()
	if got := eventTypes(evs); !slices.Equal(got, want) {
		t.Fatalf("events = %v, want %v\n%#v", got, want, evs)
	}
	return evs
}

// sameCards reports whether got holds the cards of want (cs notation) in any order.
func sameCards(got []Card, want string) bool {
	a, b := slices.Clone(got), cs(want)
	SortCards(a)
	SortCards(b)
	return slices.Equal(a, b)
}
