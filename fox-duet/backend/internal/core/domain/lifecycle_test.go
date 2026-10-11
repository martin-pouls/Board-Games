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
	mustFail(t, g, ErrGameFull, func(g *Game) error { return g.AddPlayer("carol", "Carol") })
	if s := g.Summary(); s.Joinable || s.PlayerCount != 2 || s.MaxPlayers != 2 || s.HostName != "Alice" {
		t.Fatalf("summary %+v", s)
	}
}

func TestLeave(t *testing.T) {
	g := newLobbyGame(t)
	must(t, g.Leave(alice))
	if g.HostID != bob || len(g.Players) != 1 {
		t.Fatalf("host %s players %v", g.HostID, g.Players)
	}
	mustFail(t, g, ErrCannotLeave, func(g *Game) error { return g.Leave(bob) })
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.Leave(alice) })

	g = newPlaying(t, 1)
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.Leave(bob) })
}

func TestStartErrors(t *testing.T) {
	g := newLobbyGame(t)
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.Start(bob, 1, noShuffle{}) })
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.Start("carol", 1, noShuffle{}) })
	for _, level := range []int{0, 4} {
		mustFail(t, g, ErrInvalidLevel, func(g *Game) error { return g.Start(alice, level, noShuffle{}) })
	}

	solo, err := NewGame("G", alice, "Alice")
	must(t, err)
	mustFail(t, solo, ErrNotEnoughPlayers, func(g *Game) error { return g.Start(alice, 1, noShuffle{}) })

	g = newPlaying(t, 1)
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.Start(alice, 1, noShuffle{}) })
}

func TestStart(t *testing.T) {
	for level := MinLevel; level <= MaxLevel; level++ {
		g := newLobbyGame(t)
		must(t, g.Start(alice, level, noShuffle{}))
		l, _ := LevelFor(level)
		if g.Phase != PhasePlaying || g.Step != StepLead || g.Level != level || g.Round != 1 {
			t.Fatalf("level %d: phase %s step %s level %d round %d", level, g.Phase, g.Step, g.Level, g.Round)
		}
		if !slices.Equal(g.Board.Gems, l.startingGems()) || g.Board.Traveller != l.Start() || g.Board.ForestTokens != l.ForestTokens {
			t.Fatalf("level %d: board %+v", level, g.Board)
		}
		for _, p := range g.Players {
			if len(p.Hand) != HandSize {
				t.Fatalf("level %d: %s has %d cards", level, p.ID, len(p.Hand))
			}
		}
		if len(g.Deck) != 7 || g.Decree == nil || totalCards(g) != 30 {
			t.Fatalf("level %d: deck %d decree %v total %d", level, len(g.Deck), g.Decree, totalCards(g))
		}
		// noShuffle keeps the seats in order: seat 0 deals, so seat 1 leads.
		if g.Dealer != 0 || g.Trick.Leader != 1 || g.Trick.Number != 1 {
			t.Fatalf("level %d: dealer %d leader %d trick %d", level, g.Dealer, g.Trick.Leader, g.Trick.Number)
		}
		started := eventsOf[GameStarted](g)
		rounds := eventsOf[RoundStarted](g)
		if len(started) != 1 || started[0].Level != level || len(rounds) != 1 || rounds[0].DealerID != alice || rounds[0].Decree != *g.Decree {
			t.Fatalf("level %d: events %v", level, g.events)
		}
	}
}

func TestStartDealsFromTheShuffledDeck(t *testing.T) {
	g := newLobbyGame(t)
	must(t, g.Start(alice, 1, noShuffle{}))
	// Unshuffled deck, top = last: the non-dealer (seat 1) gets the first card dealt.
	if !slices.Contains(g.Players[1].Hand, c("10S")) || !slices.Contains(g.Players[0].Hand, c("9S")) {
		t.Fatalf("hands %v / %v", g.Players[0].Hand, g.Players[1].Hand)
	}
	if *g.Decree != c("8D") || len(g.Deck) != 7 || g.Deck[6] != c("7D") {
		t.Fatalf("decree %v", *g.Decree)
	}

	// reverseShuffle swaps the seats: seat 1 deals and seat 0 leads.
	g = newLobbyGame(t)
	must(t, g.Start(alice, 1, reverseShuffle{}))
	if g.Dealer != 1 || g.Trick.Leader != 0 {
		t.Fatalf("dealer %d leader %d", g.Dealer, g.Trick.Leader)
	}
}

func TestGiveUp(t *testing.T) {
	g := newLobbyGame(t)
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.GiveUp(alice) })
	g = newPlaying(t, 2)
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.GiveUp("carol") })
	must(t, g.GiveUp(bob))
	if g.Phase != PhaseLost || g.Step != StepNone || g.Result.Reason != LossGaveUp || g.Result.PlayerID != bob {
		t.Fatalf("phase %s result %+v", g.Phase, g.Result)
	}
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.GiveUp(alice) })
}

func TestEndGame(t *testing.T) {
	g := newLobbyGame(t)
	mustFail(t, g, ErrNotHost, func(g *Game) error { return g.EndGame(bob) })
	must(t, g.EndGame(alice))
	if g.Phase != PhaseEnded {
		t.Fatalf("phase %s", g.Phase)
	}
	g = newPlaying(t, 1)
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.EndGame(alice) })
}

func TestRematch(t *testing.T) {
	g := newPlaying(t, 3)
	if _, err := g.Rematch(alice, "NEXT1"); !errors.Is(err, ErrWrongPhase) {
		t.Fatalf("running game: %v", err)
	}
	must(t, g.GiveUp(alice))
	next, err := g.Rematch(bob, "NEXT1")
	must(t, err)
	if g.NextGameID != "NEXT1" || next.ID != "NEXT1" || next.HostID != bob || next.Phase != PhaseLobby || len(next.Players) != 2 {
		t.Fatalf("rematch %+v / %+v", g, next)
	}
	if next.Players[0].ID != alice || len(next.Players[0].Hand) != 0 || next.Level != 0 {
		t.Fatalf("rematch players %+v level %d", next.Players, next.Level)
	}
	if _, err := g.Rematch(alice, "NEXT2"); !errors.Is(err, ErrRematchExists) {
		t.Fatalf("second rematch: %v", err)
	}
}
