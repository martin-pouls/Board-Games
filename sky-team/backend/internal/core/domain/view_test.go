package domain

import (
	"errors"
	"testing"
)

func TestViewInLobby(t *testing.T) {
	g := newLobbyGame(t)
	v, err := g.ViewFor(bob)
	must(t, err)
	if v.Phase != PhaseLobby || v.You != bob || v.HostID != alice || len(v.Players) != 2 {
		t.Fatalf("lobby view %+v", v)
	}
	if !v.Players[1].IsYou || !v.Players[0].IsHost || v.Players[0].IsYou || v.Players[1].Seat != 1 {
		t.Fatalf("players %+v", v.Players)
	}
	if _, err := g.ViewFor("carol"); !errors.Is(err, ErrUnknownPlayer) {
		t.Fatalf("unknown viewer: %v", err)
	}
}
