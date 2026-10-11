package domain

import (
	"errors"
	"slices"
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
	must(t, g.AddPlayer(bob, "Bob"))
	if g.Revision != 2 || len(eventsOf[PlayerJoined](g)) != 1 {
		t.Fatalf("revision %d events %v", g.Revision, g.events)
	}
	if s := g.Summary(); !s.Joinable || s.PlayerCount != 2 || s.MaxPlayers != 5 || s.HostName != "Alice" {
		t.Fatalf("summary %+v", s)
	}
}

func TestFiveSeats(t *testing.T) {
	g := newLobbyGame(t)
	must(t, g.AddPlayer(carol, "Carol"))
	must(t, g.AddPlayer("dave", "Dave"))
	if s := g.Summary(); !s.Joinable || s.PlayerCount != 4 {
		t.Fatalf("summary %+v", s)
	}
	must(t, g.AddPlayer("erin", "Erin"))
	if s := g.Summary(); s.Joinable || s.PlayerCount != 5 {
		t.Fatalf("summary %+v", s)
	}
	mustFail(t, g, ErrGameFull, func(g *Game) error { return g.AddPlayer("frank", "Frank") })
	for i, want := range []PlayerID{alice, bob, carol, "dave", "erin"} {
		if g.Players[i].ID != want {
			t.Fatalf("seat %d: %s, want %s", i, g.Players[i].ID, want)
		}
	}
}

func TestAddPlayerOutsideLobby(t *testing.T) {
	g := newLobbyGame(t)
	must(t, g.EndGame(alice))
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.AddPlayer(carol, "Carol") })
	if g.Summary().Joinable {
		t.Fatal("an ended game must not be joinable")
	}
}

func TestLeave(t *testing.T) {
	g := newLobbyGame(t)
	must(t, g.Leave(alice))
	if g.HostID != bob || len(g.Players) != 1 {
		t.Fatalf("host %s players %v", g.HostID, g.Players)
	}
	if evs := eventsOf[PlayerLeft](g); len(evs) != 1 || evs[0].PlayerID != alice || evs[0].HostID != bob {
		t.Fatalf("events %v", g.events)
	}
	mustFail(t, g, ErrCannotLeave, func(g *Game) error { return g.Leave(bob) })
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.Leave(alice) })
}

func TestLeaveKeepsSeatOrder(t *testing.T) {
	g := newLobbyGame(t)
	must(t, g.AddPlayer(carol, "Carol"))
	must(t, g.Leave(bob))
	if g.HostID != alice || !slices.Equal([]PlayerID{g.Players[0].ID, g.Players[1].ID}, []PlayerID{alice, carol}) {
		t.Fatalf("host %s players %v", g.HostID, g.Players)
	}
	must(t, g.Leave(alice))
	if g.HostID != carol {
		t.Fatalf("host %s, want carol", g.HostID)
	}
}

func TestLeaveOutsideLobby(t *testing.T) {
	g := newLobbyGame(t)
	g.Phase = PhasePlaying
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.Leave(bob) })
}

func TestEndGame(t *testing.T) {
	g := newLobbyGame(t)
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.EndGame(bob) })
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.EndGame(carol) })
	must(t, g.EndGame(alice))
	if g.Phase != PhaseEnded || len(eventsOf[GameEnded](g)) != 1 {
		t.Fatalf("phase %s events %v", g.Phase, g.events)
	}
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.EndGame(alice) })

	g = newLobbyGame(t)
	g.Phase = PhasePlaying
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.EndGame(alice) })
}

func TestClone(t *testing.T) {
	g := newLobbyGame(t)
	c := g.Clone()
	c.Players[0].Name = "Changed"
	if g.Players[0].Name != "Alice" {
		t.Fatal("clone shares the players slice")
	}
}
