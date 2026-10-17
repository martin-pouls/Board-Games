package domain

import (
	"errors"
	"slices"
	"testing"
)

func TestNewGameAndJoin(t *testing.T) {
	if _, err := NewGame("G", EditionPlanetNine, alice, "  ", Options{}); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("empty name: %v", err)
	}
	if _, err := NewGame("G", EditionPlanetNine, alice, "abcdefghijklmnopqrstu", Options{}); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("long name: %v", err)
	}
	if _, err := NewGame("G", EditionPlanetNine, alice, "a\tb", Options{}); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("control char: %v", err)
	}
	g, err := NewGame("G", EditionPlanetNine, alice, " Alice ", Options{})
	must(t, err)
	if g.Phase != PhaseLobby || g.Revision != 1 || g.Players[0].Name != "Alice" || !g.IsHost(alice) {
		t.Fatalf("bad new game %+v", g)
	}
	if g.Options.Distress {
		t.Fatalf("zero options should leave distress off")
	}
	if d, _ := NewGame("G", EditionPlanetNine, alice, "Alice", Options{Distress: true}); !d.Options.Distress {
		t.Fatalf("NewGame should keep the options")
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

func TestSetOptions(t *testing.T) {
	g := newLobbyGame(t)
	if g.Options != (Options{}) || g.Options.Distress {
		t.Fatalf("a new game starts with the defaults: %+v", g.Options)
	}

	for _, tc := range []struct {
		name string
		game func(t *testing.T) *Game
		by   PlayerID
		want error
	}{
		{"a guest may not", newLobbyGame, bob, ErrNotHost},
		{"a stranger may not", newLobbyGame, "nobody", ErrUnknownPlayer},
		{"not while playing", func(t *testing.T) *Game { return startMission(t, 2, 1) }, alice, ErrWrongPhase},
		{"not once the session ended", func(t *testing.T) *Game {
			g := newLobbyGame(t)
			must(t, g.EndGame(alice))
			return g
		}, alice, ErrWrongPhase},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mustFail(t, tc.game(t), tc.want, func(g *Game) error { return g.SetOptions(tc.by, Options{Distress: true}) })
		})
	}

	t.Run("the host turns it on and off", func(t *testing.T) {
		g := newLobbyGame(t)
		rev := g.Revision
		must(t, g.SetOptions(alice, Options{Distress: true}))
		if !g.Options.Distress || g.Revision != rev+1 {
			t.Fatalf("options %+v revision %d (was %d)", g.Options, g.Revision, rev)
		}
		if ev := eventsOf[OptionsChanged](g); len(ev) != 1 || !ev[0].Options.Distress {
			t.Fatalf("events %+v", ev)
		}
		if v, _ := g.ViewFor(bob); !v.Options.Distress {
			t.Fatal("every player sees the option")
		}
		g.PullEvents()
		must(t, g.SetOptions(alice, Options{}))
		if g.Options.Distress || g.Revision != rev+2 {
			t.Fatalf("options %+v revision %d", g.Options, g.Revision)
		}
		if v, _ := g.ViewFor(bob); v.Options.Distress {
			t.Fatal("the view follows the option")
		}
	})

	t.Run("the option survives the attempts and is kept by Clone", func(t *testing.T) {
		g := newLobbyGame(t)
		must(t, g.SetOptions(alice, Options{Distress: true}))
		must(t, g.StartMission(alice, noShuffle{}))
		if c := g.Clone(); !c.Options.Distress {
			t.Fatal("clone lost the option")
		}
		c := g.Clone()
		g.Options.Distress = false
		if !c.Options.Distress {
			t.Fatal("the clone is independent of later changes of g")
		}
	})
}

func TestDistressNeedsTheOption(t *testing.T) {
	g := fourAtPlay(t, 1)
	if g.Options.Distress {
		t.Fatal("off by default")
	}
	cmd := g.Attempt.Commander
	if g.CanSendDistress(cmd) == nil {
		t.Fatal("the signal is off: nothing may be offered")
	}
	if v, _ := g.ViewFor(alice); v.Attempt.Actions.SendDistress {
		t.Fatal("the commander is not asked while the option is off")
	}
	mustFail(t, g, ErrNotAllowed, func(g *Game) error { return g.SendDistress(alice, DirLeft) })
	mustFail(t, g, ErrNotAllowed, func(g *Game) error { return g.SkipDistress(alice) })
	if g.Logbook[1].Distress || g.Attempt.DistressDecided || g.Attempt.Step != StepPlay {
		t.Fatal("a rejected signal changes nothing")
	}
	// Play simply goes on without the question.
	must(t, g.PlayCard(alice, c("p1"), noShuffle{}))

	g = withDistress(fourAtPlay(t, 1))
	if g.CanSendDistress(g.Attempt.Commander) != nil {
		t.Fatal("the signal is on: the commander may send it")
	}
	if v, _ := g.ViewFor(alice); !v.Attempt.Actions.SendDistress {
		t.Fatal("the commander is asked while the option is on")
	}
	must(t, g.SendDistress(alice, DirLeft))
}
