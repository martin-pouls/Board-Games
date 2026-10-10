package domain

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// noShuffle is a deterministic RNG that leaves the deck in order (ascending; top = highest).
type noShuffle struct{}

func (noShuffle) Shuffle(int, func(i, j int)) {}

// reverseShuffle reverses the deck (top = lowest), to prove Start really uses the RNG.
type reverseShuffle struct{}

func (reverseShuffle) Shuffle(n int, swap func(i, j int)) {
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func cards(vals ...int) []Card {
	out := make([]Card, len(vals))
	for i, v := range vals {
		out[i] = NumberCard(v)
	}
	return out
}

func values(cs []Card) []int {
	out := make([]int, len(cs))
	for i, c := range cs {
		out[i] = c.Value
	}
	return out
}

func pid(seat int) PlayerID { return PlayerID(fmt.Sprintf("p%d", seat)) }

func mustRules(t *testing.T, o Options) Ruleset {
	t.Helper()
	r, err := NewRuleset(o)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// table describes a game in progress for scenario tests.
type table struct {
	opts  Options
	hands [][]int // per seat, any order (sorted on build)
	draw  []int   // bottom .. top
	piles [][]int // played cards per pile (len 4); nil = all piles untouched
	seat  int     // active seat
	plays []Play  // already played this turn (cards must already be on the piles)
}

// Extreme option sets for tests.
var (
	classic           = Options{Variant: VariantClassic}
	extreme           = Options{Variant: VariantExtreme}
	extremePro        = Options{Variant: VariantExtreme, Pro: true}
	extremeProReduced = Options{Variant: VariantExtreme, Pro: true, ReducedHandSize: true}
)

// play is a Play of card value v on pile; table.build maps the value to the deck's card,
// so with extreme options the command applies automatically.
func play(v, pile int) Play { return Play{Card: NumberCard(v), Pile: pile} }

// xcard returns the extreme deck's card with value v (with its command, if any).
func xcard(v int) Card {
	r, err := NewRuleset(extreme)
	if err != nil {
		panic(err)
	}
	deck, err := NewDeck(r)
	if err != nil {
		panic(err)
	}
	for _, c := range deck {
		if c.Value == v {
			return c
		}
	}
	return NumberCard(v)
}

// build creates a game in PhasePlaying at the start of a turn (Revision 1, no events).
// Every value (hands, draw pile, piles, plays) is mapped to the card of NewDeck(rules)
// with that value, falling back to NumberCard(v): with extreme options the real command
// values apply (95 = STOP, 41 = Skull, ...); classic is unchanged.
func (tb table) build(t *testing.T) *Game {
	t.Helper()
	if tb.opts == (Options{}) {
		tb.opts = classic
	}
	r := mustRules(t, tb.opts)
	deck, err := NewDeck(r)
	if err != nil {
		t.Fatal(err)
	}
	byValue := make(map[int]Card, len(deck))
	for _, c := range deck {
		byValue[c.Value] = c
	}
	card := func(v int) Card {
		if c, ok := byValue[v]; ok {
			return c
		}
		return NumberCard(v)
	}
	deckCards := func(vals []int) []Card {
		out := make([]Card, len(vals))
		for i, v := range vals {
			out[i] = card(v)
		}
		return out
	}
	players := make([]Player, len(tb.hands))
	for i, h := range tb.hands {
		hand := deckCards(h)
		sortCards(hand)
		players[i] = Player{ID: pid(i), Name: fmt.Sprintf("Player %d", i), Hand: hand}
	}
	piles := NewPiles(r)
	for i, pc := range tb.piles {
		piles[i].Cards = append(piles[i].Cards, deckCards(pc)...)
	}
	var plays []Play
	if tb.plays != nil {
		plays = make([]Play, len(tb.plays))
		for i, p := range tb.plays {
			plays[i] = Play{Card: card(p.Card.Value), Pile: p.Pile}
		}
	}
	return &Game{
		ID:       "TEST2",
		Options:  tb.opts,
		Rules:    r,
		Phase:    PhasePlaying,
		HostID:   pid(0),
		Players:  players,
		DrawPile: deckCards(tb.draw),
		Piles:    piles,
		Turn: Turn{
			Number:        1,
			ActiveSeat:    tb.seat,
			Plays:         plays,
			StartHandSize: len(players[tb.seat].Hand) + len(tb.plays),
		},
		Revision: 1,
	}
}

// seq returns the ints from..to inclusive (descending if from > to).
func seq(from, to int) []int {
	var out []int
	step := 1
	if from > to {
		step = -1
	}
	for v := from; ; v += step {
		out = append(out, v)
		if v == to {
			return out
		}
	}
}

// expectErr runs op, checks that it fails with want, and that g is unchanged (atomicity).
func expectErr(t *testing.T, g *Game, want error, op func(g *Game) error) {
	t.Helper()
	before := g.Clone()
	err := op(g)
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
	if !reflect.DeepEqual(before, g) {
		t.Fatalf("state changed on error %v:\nbefore %+v\nafter  %+v", err, before, g)
	}
}

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// lobbyWith creates a lobby game with n players p0..p(n-1) and the given options.
func lobbyWith(t *testing.T, n int, o Options) *Game {
	t.Helper()
	g, err := NewGame("LOBBY", pid(0), "Player 0", o)
	mustOK(t, err)
	for i := 1; i < n; i++ {
		mustOK(t, g.AddPlayer(pid(i), fmt.Sprintf("Player %d", i)))
	}
	g.PullEvents()
	return g
}

// totalCards counts every card in hands, draw pile and piles.
func totalCards(g *Game) int {
	n := len(g.DrawPile)
	for _, p := range g.Players {
		n += len(p.Hand)
	}
	for _, p := range g.Piles {
		n += len(p.Cards)
	}
	return n
}
