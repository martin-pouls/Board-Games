package domain

import (
	"errors"
	"slices"
	"testing"
)

func TestViewInLobby(t *testing.T) {
	g := newLobbyGame(t)
	v, err := g.ViewFor(bob)
	must(t, err)
	if v.Phase != PhaseLobby || v.Board != nil || v.Trick != nil || v.Pending != nil || v.Decree != nil || len(v.Hand) != 0 || len(v.Playable) != 0 {
		t.Fatalf("lobby view %+v", v)
	}
	if v.Players[0].End != EndLeft || v.Players[1].End != EndRight || !v.Players[1].IsYou || !v.Players[0].IsHost || v.Players[0].IsDealer {
		t.Fatalf("players %+v", v.Players)
	}
	if _, err := g.ViewFor("carol"); !errors.Is(err, ErrUnknownPlayer) {
		t.Fatalf("unknown viewer: %v", err)
	}
}

func TestViewWhilePlaying(t *testing.T) {
	g := newPlaying(t, 3)
	va, _ := g.ViewFor(alice)
	vb, _ := g.ViewFor(bob)
	if !slices.Equal(va.Hand, g.Players[0].Hand) || !slices.Equal(vb.Hand, g.Players[1].Hand) {
		t.Fatal("hands not the viewer's own")
	}
	if va.Players[1].HandCount != 11 || va.Players[1].RevealedHand != nil {
		t.Fatalf("teammate %+v", va.Players[1])
	}
	// Bob leads: only he has playable cards.
	if len(va.Playable) != 0 || !slices.Equal(vb.Playable, vb.Hand) {
		t.Fatalf("playable %v / %v", va.Playable, vb.Playable)
	}
	if va.Pending == nil || va.Pending.Step != StepLead || !slices.Equal(va.Pending.PlayerIDs, []PlayerID{bob}) || !va.Players[1].MustAct || va.Players[0].MustAct {
		t.Fatalf("pending %+v", va.Pending)
	}
	if !va.Players[0].IsDealer || va.Players[1].IsDealer || va.Round != 1 || va.Level != 3 || *va.Decree != *g.Decree {
		t.Fatalf("view %+v", va)
	}
	b := va.Board
	if b.Side != SideB || b.Start != 4 || len(b.Spaces) != 10 || b.Traveller != 4 || b.ForestTokens != 3 || b.ForestTokensTotal != 3 || b.GemsLeft != 16 {
		t.Fatalf("board %+v", b)
	}
	if b.Spaces[0] != (SpaceView{Gems: 2}) || b.Spaces[1] != (SpaceView{Gems: 2, Plus: true}) {
		t.Fatalf("spaces %+v", b.Spaces)
	}
	if va.Trick == nil || va.Trick.Number != 1 || va.Trick.LeaderID != bob || va.Trick.Lead != nil || va.Trick.WinnerID != "" {
		t.Fatalf("trick %+v", va.Trick)
	}

	g.Board.CoveredLeft = 1
	va, _ = g.ViewFor(alice)
	if !va.Board.Spaces[0].Covered || va.Board.Spaces[1].Covered {
		t.Fatalf("covered %+v", va.Board.Spaces)
	}
}

func TestViewPendingDecisions(t *testing.T) {
	g := newPlaying(t, 1)
	setup(g, "1D 2R", "6D 8S", "2S", 0, 1)
	play(t, g, alice, "1D")
	play(t, g, bob, "6D")
	v, _ := g.ViewFor(alice)
	if v.Pending.Step != StepMove || len(v.Pending.MoveOptions) != 2 || v.Trick.WinnerID != bob || v.Trick.Follow == nil {
		t.Fatalf("pending %+v trick %+v", v.Pending, v.Trick)
	}

	g = newPlaying(t, 1)
	setup(g, "3D 5S", "2D 8S", "2R", 0, 1)
	play(t, g, alice, "3D")
	must(t, g.ChooseFoxesPlayer(alice, bob, noShuffle{}))
	v, _ = g.ViewFor(alice)
	if v.Pending.Step != StepFoxesSwap || v.Pending.ChosenID != bob || !slices.Equal(v.Pending.PlayerIDs, []PlayerID{bob}) {
		t.Fatalf("pending %+v", v.Pending)
	}

	g = newPlaying(t, 1)
	setup(g, "4D", "6D", "2S", 0, 11)
	play(t, g, alice, "4D")
	play(t, g, bob, "6D")
	v, _ = g.ViewFor(bob)
	if v.Pending.Step != StepForestEnd || !slices.Equal(v.Pending.Ends, []End{EndLeft, EndRight}) || !slices.Equal(v.Pending.PlayerIDs, []PlayerID{alice}) || v.Trick != nil {
		t.Fatalf("pending %+v trick %+v", v.Pending, v.Trick)
	}
}

func TestViewWhenOver(t *testing.T) {
	g := newPlaying(t, 1)
	must(t, g.GiveUp(alice))
	v, _ := g.ViewFor(bob)
	if v.Pending != nil || v.Trick != nil || v.Result == nil || v.Result.Outcome != PhaseLost || v.Result.Reason != LossGaveUp ||
		v.Result.PlayerID != alice || v.Result.Score != nil || v.Result.GemsLeft != 12 || v.Result.Round != 1 {
		t.Fatalf("view %+v result %+v", v, v.Result)
	}
	if !slices.Equal(v.Players[0].RevealedHand, g.Players[0].Hand) || len(v.Playable) != 0 {
		t.Fatalf("revealed %v", v.Players[0].RevealedHand)
	}

	g = newPlaying(t, 1)
	setup(g, "6D 2R", "4D 8S", "2S", 0, 5)
	g.Board.Gems = []int{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}
	play(t, g, alice, "6D")
	play(t, g, bob, "4D")
	v, _ = g.ViewFor(alice)
	if v.Result == nil || v.Result.Outcome != PhaseWon || v.Result.Score == nil || v.Result.Score.Total != 34 || v.LastTrick == nil {
		t.Fatalf("result %+v", v.Result)
	}
}
