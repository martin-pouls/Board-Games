package domain

import (
	"errors"
	"testing"
)

func TestNewGameAndJoin(t *testing.T) {
	if _, err := NewGame("G", alice, "  "); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("empty name: %v", err)
	}
	if _, err := NewGame("G", alice, "abcdefghijklmnopqrstu"); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("long name: %v", err)
	}
	if _, err := NewGame("G", alice, "a\tb"); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("control char: %v", err)
	}
	g, err := NewGame("G", alice, " Alice ")
	must(t, err)
	if g.Phase != PhaseLobby || g.Revision != 1 || g.Players[0].Name != "Alice" || !g.IsHost(alice) {
		t.Fatalf("bad new game %+v", g)
	}
	mustFail(t, g, ErrNameTaken, func(g *Game) error { return g.AddPlayer(bob, "alice") })
	mustFail(t, g, ErrDuplicatePlayer, func(g *Game) error { return g.AddPlayer(alice, "Other") })
	mustFail(t, g, ErrInvalidName, func(g *Game) error { return g.AddPlayer(bob, " ") })
	must(t, g.AddPlayer(bob, "Bob"))
	if g.Revision != 2 || len(eventsOf[PlayerJoined](g)) != 1 {
		t.Fatalf("revision %d events %v", g.Revision, g.events)
	}
	mustFail(t, g, ErrGameFull, func(g *Game) error { return g.AddPlayer("carol", "Carol") })
	if s := g.Summary(); s.Joinable || s.PlayerCount != 2 || s.MaxPlayers != 2 || s.HostName != "Alice" {
		t.Fatalf("summary %+v", s)
	}
}

func TestAddPlayerOutsideLobby(t *testing.T) {
	g, err := NewGame("G", alice, "Alice")
	must(t, err)
	must(t, g.EndGame(alice))
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.AddPlayer(bob, "Bob") })
}

func TestLeave(t *testing.T) {
	g := newLobbyGame(t)
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.Leave("carol") })
	must(t, g.Leave(alice))
	if g.HostID != bob || len(g.Players) != 1 {
		t.Fatalf("host %s players %v", g.HostID, g.Players)
	}
	if ev := eventsOf[PlayerLeft](g); len(ev) != 1 || ev[0].HostID != bob {
		t.Fatalf("events %v", g.events)
	}
	mustFail(t, g, ErrCannotLeave, func(g *Game) error { return g.Leave(bob) })
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.Leave(alice) })

	g = newLobbyGame(t)
	must(t, g.EndGame(alice))
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.Leave(bob) })
}

func TestEndGame(t *testing.T) {
	g := newLobbyGame(t)
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.EndGame(bob) })
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.EndGame("carol") })
	must(t, g.EndGame(alice))
	if g.Phase != PhaseEnded || len(eventsOf[GameEnded](g)) != 1 {
		t.Fatalf("phase %s events %v", g.Phase, g.events)
	}
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.EndGame(alice) })
}
