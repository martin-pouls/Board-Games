package domain

import (
	"errors"
	"testing"
)

func TestViewInLobby(t *testing.T) {
	g := newLobbyGame(t)
	must(t, g.AddPlayer(carol, "Carol"))
	v, err := g.ViewFor(bob)
	must(t, err)
	if v.Phase != PhaseLobby || v.You != bob || v.HostID != alice || v.GameID != "GAME1" || len(v.Players) != 3 {
		t.Fatalf("lobby view %+v", v)
	}
	if !v.Players[1].IsYou || v.Players[0].IsYou || !v.Players[0].IsHost || v.Players[1].IsHost || v.Players[2].Seat != 2 {
		t.Fatalf("players %+v", v.Players)
	}
	if _, err := g.ViewFor("nobody"); !errors.Is(err, ErrUnknownPlayer) {
		t.Fatalf("unknown viewer: %v", err)
	}
}
