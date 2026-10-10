package domain

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestNewGame(t *testing.T) {
	g, err := NewGame("K7MQ2", "h", "  Anna  ")
	mustOK(t, err)
	want := &Game{
		ID:          "K7MQ2",
		Phase:       PhaseLobby,
		HostID:      "h",
		Players:     []Player{{ID: "h", Name: "Anna", Hand: []Card{}}},
		Castle:      []Card{},
		Tavern:      []Card{},
		DiscardPile: []Card{},
		Revision:    1,
		CreatedAt:   g.CreatedAt,
	}
	if !reflect.DeepEqual(g, want) {
		t.Fatalf("NewGame =\n%+v\nwant\n%+v", g, want)
	}
	if evs := g.PullEvents(); len(evs) != 0 {
		t.Fatalf("NewGame emitted %v", evs)
	}
	if g.CreatedAt.IsZero() || time.Since(g.CreatedAt) > time.Minute || g.CreatedAt.Location() != time.UTC ||
		g.CreatedAt.Nanosecond() != 0 {
		t.Fatalf("CreatedAt = %v, want now (UTC, seconds)", g.CreatedAt)
	}
	checkShape(t, g)
}

func TestNewGameErrors(t *testing.T) {
	for name, host := range map[string]string{
		"empty":          "",
		"blank":          " \t ",
		"too long":       strings.Repeat("a", 21),
		"too long runes": strings.Repeat("ä", 21),
		"control char":   "a\tb",
		"nul":            "a\x00b",
		"invalid utf8":   "a\xffb",
	} {
		t.Run(name, func(t *testing.T) {
			g, err := NewGame("X", "h", host)
			if !errors.Is(err, ErrInvalidName) || g != nil {
				t.Fatalf("NewGame = %v, %v", g, err)
			}
		})
	}
}

func TestNameValidationAccepts(t *testing.T) {
	for _, name := range []string{"A", strings.Repeat("ä", 20), "Jörg-Ünal 😀", "x y"} {
		if _, err := NewGame("X", "h", name); err != nil {
			t.Errorf("NewGame(%q) = %v", name, err)
		}
	}
}

func TestAddPlayer(t *testing.T) {
	g := lobbyWith(t, 1)
	mustOK(t, g.AddPlayer("b", "  Ben "))
	if g.Revision != 2 || len(g.Players) != 2 || g.Players[1].ID != "b" || g.Players[1].Name != "Ben" || g.Players[1].Hand == nil {
		t.Fatalf("after AddPlayer: %+v", g)
	}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{PlayerJoined{PlayerID: "b", Name: "Ben"}}) {
		t.Fatalf("events = %#v", evs)
	}
	mustOK(t, g.AddPlayer("c", "Cara"))
	mustOK(t, g.AddPlayer("d", "Dan"))
	if len(g.Players) != MaxPlayers || g.Revision != 4 {
		t.Fatalf("players = %d, revision = %d", len(g.Players), g.Revision)
	}
}

func TestAddPlayerErrors(t *testing.T) {
	tests := []struct {
		name string
		g    *Game
		id   PlayerID
		pn   string
		want error
	}{
		{"after start", started(t, 2), "new", "New", ErrWrongPhase},
		{"full", lobbyWith(t, 4), "new", "New", ErrGameFull},
		{"invalid name", lobbyWith(t, 2), "new", "", ErrInvalidName},
		{"name taken exact", lobbyWith(t, 2), "new", "Player 1", ErrNameTaken},
		{"name taken case-insensitive", lobbyWith(t, 2), "new", "  PLAYER 0 ", ErrNameTaken},
		{"duplicate id", lobbyWith(t, 2), pid(1), "Other", ErrDuplicatePlayer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.AddPlayer(tt.id, tt.pn) })
		})
	}
}

func TestLeave(t *testing.T) {
	t.Run("guest", func(t *testing.T) {
		g := lobbyWith(t, 3)
		mustOK(t, g.Leave(pid(1)))
		if len(g.Players) != 2 || g.Players[0].ID != pid(0) || g.Players[1].ID != pid(2) || g.HostID != pid(0) || g.Revision != 4 {
			t.Fatalf("after leave: %+v", g)
		}
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{PlayerLeft{PlayerID: pid(1), HostID: pid(0)}}) {
			t.Fatalf("events = %#v", evs)
		}
	})
	t.Run("host hands over to seat 0", func(t *testing.T) {
		g := lobbyWith(t, 3)
		mustOK(t, g.Leave(pid(0)))
		if len(g.Players) != 2 || g.HostID != pid(1) || g.Players[0].ID != pid(1) {
			t.Fatalf("after host leave: %+v", g)
		}
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{PlayerLeft{PlayerID: pid(0), HostID: pid(1)}}) {
			t.Fatalf("events = %#v", evs)
		}
		// the remaining players can rejoin the vacated name and fill the game
		mustOK(t, g.AddPlayer("new", "Player 0"))
	})
	t.Run("does not alias", func(t *testing.T) {
		g := lobbyWith(t, 3)
		before := g.Clone()
		players := g.Players
		mustOK(t, g.Leave(pid(0)))
		if !reflect.DeepEqual(players, before.Players) {
			t.Fatal("Leave mutated the previous players slice")
		}
	})
	tests := []struct {
		name string
		g    *Game
		by   PlayerID
		want error
	}{
		{"unknown", lobbyWith(t, 2), "zz", ErrUnknownPlayer},
		{"sole player", lobbyWith(t, 1), pid(0), ErrCannotLeave},
		{"after start", started(t, 2), pid(1), ErrWrongPhase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.Leave(tt.by) })
		})
	}
}

func TestStartPerPlayerCount(t *testing.T) {
	tests := []struct {
		n, hand, jesters, soloLeft int
	}{
		{1, 8, 0, 2},
		{2, 7, 0, 0},
		{3, 6, 1, 0},
		{4, 5, 2, 0},
	}
	for _, tt := range tests {
		g := lobbyWith(t, tt.n)
		mustOK(t, g.Start(pid(0), "", reverseShuffle{}))
		checkInvariants(t, g)
		for _, p := range g.Players {
			if len(p.Hand) != tt.hand {
				t.Fatalf("%dp: %s holds %d cards", tt.n, p.ID, len(p.Hand))
			}
		}
		if g.Phase != PhasePlaying || g.Step != StepPlay || g.Enemy == nil || g.Enemy.Rank != RankJack ||
			len(g.Castle) != 11 || len(g.Tavern) != 40+tt.jesters-tt.n*tt.hand || len(g.DiscardPile) != 0 ||
			g.SoloJestersLeft() != tt.soloLeft || g.Rules().TavernJesters != tt.jesters || g.Turn.Number != 1 ||
			g.Revision != int64(tt.n+1) || g.Result != nil {
			t.Fatalf("%dp: %+v", tt.n, g)
		}
	}
}

func TestStartDealsWithNoShuffle(t *testing.T) {
	g := lobbyWith(t, 2)
	mustOK(t, g.Start(pid(0), pid(1), noShuffle{}))
	if !slices.Equal(g.Players[0].Hand, sorted("10S 8S 6S 4S 2S 10C 8C")) ||
		!slices.Equal(g.Players[1].Hand, sorted("9S 7S 5S 3S AS 9C 7C")) {
		t.Fatalf("hands %v / %v", g.Players[0].Hand, g.Players[1].Hand)
	}
	if *g.Enemy != c("JS") || g.Castle[len(g.Castle)-1] != c("JC") || g.Tavern[len(g.Tavern)-1] != c("6C") {
		t.Fatalf("enemy %v castle top %v tavern top %v", g.Enemy, g.Castle[len(g.Castle)-1], g.Tavern[len(g.Tavern)-1])
	}
	if g.Turn.ActiveSeat != 1 {
		t.Fatalf("active seat %d, want the chosen first player", g.Turn.ActiveSeat)
	}
	evs := g.PullEvents()
	want := []Event{GameStarted{FirstPlayerID: pid(1)}, EnemyRevealed{Enemy: c("JS")}, TurnStarted{PlayerID: pid(1), Number: 1}}
	if !reflect.DeepEqual(evs, want) {
		t.Fatalf("events = %#v", evs)
	}

	g3 := lobbyWith(t, 3)
	mustOK(t, g3.Start(pid(0), pid(0), noShuffle{}))
	if g3.Players[0].Hand[0] != Jester() || slices.Contains(g3.Players[1].Hand, Jester()) {
		t.Fatalf("3p: seat 0 should hold the Jester: %v", g3.Players[0].Hand)
	}
	g4 := lobbyWith(t, 4)
	mustOK(t, g4.Start(pid(0), pid(0), noShuffle{}))
	if g4.Players[0].Hand[0] != Jester() || g4.Players[1].Hand[0] != Jester() {
		t.Fatalf("4p: seats 0 and 1 should hold the Jesters")
	}
}

func TestStartRandomFirstPlayer(t *testing.T) {
	// RNG calls: 3 castle tiers, the Tavern, then the seats; reversing the seats picks the last one.
	rng := &countingRNG{reverse: map[int]bool{4: true}}
	g := lobbyWith(t, 3)
	mustOK(t, g.Start(pid(0), "", rng))
	if !slices.Equal(rng.calls, []int{4, 4, 4, 41, 3}) || g.Turn.ActiveSeat != 2 {
		t.Fatalf("calls %v active %d", rng.calls, g.Turn.ActiveSeat)
	}
	// With a chosen first player the seats are not shuffled.
	rng = &countingRNG{}
	g = lobbyWith(t, 3)
	mustOK(t, g.Start(pid(0), pid(1), rng))
	if !slices.Equal(rng.calls, []int{4, 4, 4, 41}) || g.Turn.ActiveSeat != 1 {
		t.Fatalf("calls %v active %d", rng.calls, g.Turn.ActiveSeat)
	}
	// The Tavern shuffle is used.
	a, b := lobbyWith(t, 2), lobbyWith(t, 2)
	mustOK(t, a.Start(pid(0), pid(0), noShuffle{}))
	mustOK(t, b.Start(pid(0), pid(0), reverseShuffle{}))
	if slices.Equal(a.Players[0].Hand, b.Players[0].Hand) || *b.Enemy != c("JH") {
		t.Fatal("Start ignores the RNG")
	}
}

func TestStartErrors(t *testing.T) {
	tests := []struct {
		name      string
		g         *Game
		by, first PlayerID
		want      error
	}{
		{"unknown", lobbyWith(t, 2), "zz", "", ErrUnknownPlayer},
		{"not host", lobbyWith(t, 2), pid(1), "", ErrNotHost},
		{"started", started(t, 2), pid(0), "", ErrWrongPhase},
		{"unknown first", lobbyWith(t, 2), pid(0), "zz", ErrUnknownPlayer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.Start(tt.by, tt.first, noShuffle{}) })
		})
	}
}

func TestGiveUp(t *testing.T) {
	for _, step := range []Step{StepPlay, StepDiscard, StepChooseNext} {
		g := table{hands: []string{"5H", "6H"}, enemy: "JS", step: step, played: []played{{0, "2C"}}}.build(t)
		mustOK(t, g.GiveUp(pid(1))) // any player, not only the active one
		if g.Phase != PhaseLost || g.Step != StepNone || *g.Result != (Result{Reason: LossGaveUp, PlayerID: pid(1)}) || g.Revision != 2 {
			t.Fatalf("%s: %+v", step, g)
		}
		if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{GameLost{Reason: LossGaveUp, PlayerID: pid(1)}}) {
			t.Fatalf("events %#v", evs)
		}
		checkShape(t, g)
		expectErr(t, g, ErrWrongPhase, func(g *Game) error { return g.GiveUp(pid(0)) })
	}
	expectErr(t, lobbyWith(t, 2), ErrWrongPhase, func(g *Game) error { return g.GiveUp(pid(0)) })
	expectErr(t, started(t, 2), ErrUnknownPlayer, func(g *Game) error { return g.GiveUp("zz") })
}

// finishedGame is a lost 3-player game at Revision 2.
func finishedGame(t *testing.T) *Game {
	t.Helper()
	g := table{hands: []string{"5H", "6H", "7H"}, enemy: "JS", yielded: []int{1}}.build(t)
	mustOK(t, g.GiveUp(pid(0)))
	g.PullEvents()
	return g
}

func TestRematch(t *testing.T) {
	for _, by := range []PlayerID{pid(0), pid(1), pid(2)} { // any player, not only the host
		t.Run(string(by), func(t *testing.T) {
			g := finishedGame(t)
			next, err := g.Rematch(by, "NEXT2")
			mustOK(t, err)
			if g.NextGameID != "NEXT2" || g.Revision != 3 || g.Phase != PhaseLost {
				t.Fatalf("old game: next %q rev %d phase %s", g.NextGameID, g.Revision, g.Phase)
			}
			if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{RematchCreated{GameID: "NEXT2"}}) {
				t.Fatalf("events = %#v", evs)
			}
			want := &Game{
				ID:     "NEXT2",
				Phase:  PhaseLobby,
				HostID: by,
				Players: []Player{
					{ID: pid(0), Name: "Player 0", Hand: []Card{}},
					{ID: pid(1), Name: "Player 1", Hand: []Card{}},
					{ID: pid(2), Name: "Player 2", Hand: []Card{}},
				},
				Castle:      []Card{},
				Tavern:      []Card{},
				DiscardPile: []Card{},
				Revision:    1,
				CreatedAt:   next.CreatedAt,
			}
			if next.CreatedAt.IsZero() || time.Since(next.CreatedAt) > time.Minute {
				t.Fatalf("CreatedAt = %v, want about now", next.CreatedAt)
			}
			if !reflect.DeepEqual(next, want) {
				t.Fatalf("new game:\n got %+v\nwant %+v", next, want)
			}
			next.Players[0].Name = "changed"
			if g.Players[0].Name != "Player 0" {
				t.Fatal("new game aliases old players")
			}
			mustOK(t, next.Start(by, "", noShuffle{}))
		})
	}
}

func TestRematchErrors(t *testing.T) {
	exists := finishedGame(t)
	exists.NextGameID = "OLD22"
	tests := []struct {
		name string
		g    *Game
		by   PlayerID
		want error
	}{
		{"unknown", finishedGame(t), "zz", ErrUnknownPlayer},
		{"playing", started(t, 2), pid(0), ErrWrongPhase},
		{"lobby", lobbyWith(t, 2), pid(0), ErrWrongPhase},
		{"exists", exists, pid(1), ErrRematchExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error {
				next, err := g.Rematch(tt.by, "NEW22")
				if next != nil {
					t.Errorf("Rematch returned a game on error")
				}
				return err
			})
		})
	}
}
