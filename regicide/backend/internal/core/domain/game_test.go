package domain

import (
	"reflect"
	"testing"
	"time"
)

func TestCloneDeepEqualAndIndependent(t *testing.T) {
	g := table{
		hands:   []string{"5H 6C", "7D"},
		enemy:   "JS",
		castle:  "KH QH",
		tavern:  "2H 3H",
		discard: "4H",
		played:  []played{{0, "8S"}, {1, "X"}},
		yielded: []int{1},
	}.build(t)
	g.Result = &Result{Reason: LossGaveUp, PlayerID: pid(0)}
	g.emit(Yielded{PlayerID: pid(0)})
	cl := g.Clone()
	if !reflect.DeepEqual(cl, g) {
		t.Fatalf("clone differs:\n%+v\n%+v", cl, g)
	}
	before := g.Clone()
	cl.Players[0].Hand[0] = c("KD")
	cl.Players[1].Name = "x"
	*cl.Enemy = c("QC")
	cl.Castle[0] = c("2C")
	cl.Tavern[0] = c("2C")
	cl.DiscardPile[0] = c("2C")
	cl.Played[0].Cards[0] = c("2C")
	cl.Played[1].By = "zz"
	cl.Result.Reason = LossCannotAct
	cl.events[0] = GameWon{}
	if !reflect.DeepEqual(before, g) {
		t.Fatal("mutating the clone changed the original")
	}
}

func TestClonePreservesNilAndEmpty(t *testing.T) {
	g := lobbyWith(t, 1)
	cl := g.Clone()
	if cl.Played != nil || cl.Enemy != nil || cl.Result != nil || cl.Castle == nil || cl.Tavern == nil ||
		cl.DiscardPile == nil || cl.Players[0].Hand == nil || cl.events != nil {
		t.Fatalf("clone %+v", cl)
	}
	if (*Game)(nil).Clone() != nil {
		t.Fatal("nil clone")
	}
}

func TestPullEvents(t *testing.T) {
	g := lobbyWith(t, 1)
	mustOK(t, g.AddPlayer("b", "Ben"))
	if evs := g.PullEvents(); len(evs) != 1 {
		t.Fatalf("events %v", evs)
	}
	if evs := g.PullEvents(); evs != nil {
		t.Fatalf("second pull %v", evs)
	}
}

func TestPlayerByIDAndIsHost(t *testing.T) {
	g := lobbyWith(t, 2)
	if p, seat, ok := g.PlayerByID(pid(1)); !ok || seat != 1 || p.ID != pid(1) {
		t.Fatalf("PlayerByID = %v %d %v", p, seat, ok)
	}
	if _, seat, ok := g.PlayerByID("zz"); ok || seat != -1 {
		t.Fatal("unknown player found")
	}
	if !g.IsHost(pid(0)) || g.IsHost(pid(1)) || g.IsHost("") {
		t.Fatal("IsHost")
	}
	g.HostID = ""
	if g.IsHost("") {
		t.Fatal(`IsHost("") with empty host`)
	}
}

func TestRulesAndSolo(t *testing.T) {
	g := &Game{}
	if g.Rules() != (Ruleset{}) || g.IsSolo() || g.SoloJestersLeft() != 0 {
		t.Fatal("zero game")
	}
	s := lobbyWith(t, 1)
	if !s.IsSolo() || s.SoloJestersLeft() != 2 {
		t.Fatal("solo")
	}
	s.SoloJestersUsed = 3
	if s.SoloJestersLeft() != 0 {
		t.Fatal("solo jesters left below 0")
	}
}

func TestSummary(t *testing.T) {
	g := lobbyWith(t, 2)
	g.CreatedAt = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	want := GameSummary{GameID: "LOBBY", Phase: PhaseLobby, HostName: "Player 0", PlayerCount: 2, MaxPlayers: 4, Joinable: true, CreatedAt: g.CreatedAt}
	if got := g.Summary(); got != want {
		t.Fatalf("Summary = %+v", got)
	}
	if lobbyWith(t, 4).Summary().Joinable || started(t, 2).Summary().Joinable {
		t.Fatal("full or started game joinable")
	}
	g.HostID = "zz"
	if g.Summary().HostName != "" {
		t.Fatal("host name for an unknown host")
	}
}

func TestRevisionPerSuccess(t *testing.T) {
	g := lobbyWith(t, 2) // NewGame 1, AddPlayer 2
	if g.Revision != 2 {
		t.Fatalf("revision %d", g.Revision)
	}
	mustOK(t, g.Start(pid(0), pid(0), noShuffle{}))
	mustOK(t, g.Play(pid(0), cs("8C"), noShuffle{}))
	mustOK(t, g.Discard(pid(0), cs("10S")))
	if g.Revision != 5 {
		t.Fatalf("revision %d, want one per command", g.Revision)
	}
}
