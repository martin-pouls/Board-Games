package domain

import (
	"errors"
	"reflect"
	"testing"
)

const (
	alice PlayerID = "alice" // seat 0, host
	bob   PlayerID = "bob"   // seat 1
	carol PlayerID = "carol" // seat 2
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
