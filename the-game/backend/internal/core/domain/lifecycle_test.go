package domain

import (
	"errors"
	"math/rand/v2"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestNewGame(t *testing.T) {
	g, err := NewGame("K7MQ2", "h", "  Anna  ", Options{Variant: VariantClassic, Pro: true})
	mustOK(t, err)
	if g.ID != "K7MQ2" || g.Phase != PhaseLobby || g.HostID != "h" || g.Revision != 1 {
		t.Fatalf("game = %+v", g)
	}
	if len(g.Players) != 1 || g.Players[0].ID != "h" || g.Players[0].Name != "Anna" || len(g.Players[0].Hand) != 0 {
		t.Fatalf("players = %+v", g.Players)
	}
	if !reflect.DeepEqual(g.Piles, NewPiles(g.Rules)) || g.Rules.MinCardsPerTurn != 3 || g.Result != nil || g.NextGameID != "" {
		t.Fatalf("table = %+v", g)
	}
	if evs := g.PullEvents(); len(evs) != 0 {
		t.Fatalf("NewGame emitted %v", evs)
	}
}

func TestNewGameErrors(t *testing.T) {
	tests := []struct {
		name string
		host string
		opts Options
		want error
	}{
		{"empty name", "", DefaultOptions(), ErrInvalidName},
		{"blank name", " \t ", DefaultOptions(), ErrInvalidName},
		{"too long", strings.Repeat("a", 21), DefaultOptions(), ErrInvalidName},
		{"too long runes", strings.Repeat("ä", 21), DefaultOptions(), ErrInvalidName},
		{"control char", "a\tb", DefaultOptions(), ErrInvalidName},
		{"nul", "a\x00b", DefaultOptions(), ErrInvalidName},
		{"invalid utf8", "a\xffb", DefaultOptions(), ErrInvalidName},
		{"extreme reduced without pro", "Anna", Options{Variant: VariantExtreme, ReducedHandSize: true}, ErrInvalidOptions},
		{"empty variant", "Anna", Options{}, ErrInvalidOptions},
		{"reduced without pro", "Anna", Options{Variant: VariantClassic, ReducedHandSize: true}, ErrInvalidOptions},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGame("X", "h", tt.host, tt.opts)
			if !errors.Is(err, tt.want) || g != nil {
				t.Fatalf("NewGame = %v, %v; want %v", g, err, tt.want)
			}
		})
	}
}

func TestNameValidationAccepts(t *testing.T) {
	for _, name := range []string{"A", strings.Repeat("ä", 20), "Jörg-Ünal 😀", "x y"} {
		if _, err := NewGame("X", "h", name, DefaultOptions()); err != nil {
			t.Errorf("NewGame(%q) = %v", name, err)
		}
	}
}

func TestAddPlayer(t *testing.T) {
	g := lobbyWith(t, 1, DefaultOptions())
	mustOK(t, g.AddPlayer("b", "  Ben "))
	if g.Revision != 2 || len(g.Players) != 2 || g.Players[1].ID != "b" || g.Players[1].Name != "Ben" || g.Players[1].Hand == nil {
		t.Fatalf("after AddPlayer: %+v", g)
	}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{PlayerJoined{PlayerID: "b", Name: "Ben"}}) {
		t.Fatalf("events = %#v", evs)
	}
	// up to MaxPlayers
	for i := 2; i < 5; i++ {
		mustOK(t, g.AddPlayer(pid(10+i), "N"+string(rune('a'+i))))
	}
	if len(g.Players) != 5 || g.Revision != 5 {
		t.Fatalf("players = %d, revision = %d", len(g.Players), g.Revision)
	}
}

func TestAddPlayerErrors(t *testing.T) {
	started := lobbyWith(t, 2, DefaultOptions())
	mustOK(t, started.Start(pid(0), noShuffle{}))
	tests := []struct {
		name string
		g    *Game
		id   PlayerID
		pn   string
		want error
	}{
		{"after start", started, "new", "New", ErrWrongPhase},
		{"full", lobbyWith(t, 5, DefaultOptions()), "new", "New", ErrGameFull},
		{"invalid name", lobbyWith(t, 2, DefaultOptions()), "new", "", ErrInvalidName},
		{"name taken exact", lobbyWith(t, 2, DefaultOptions()), "new", "Player 1", ErrNameTaken},
		{"name taken case-insensitive", lobbyWith(t, 2, DefaultOptions()), "new", "  PLAYER 0 ", ErrNameTaken},
		{"duplicate id", lobbyWith(t, 2, DefaultOptions()), pid(1), "Other", ErrDuplicatePlayer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.AddPlayer(tt.id, tt.pn) })
		})
	}
}

func TestSetOptions(t *testing.T) {
	g := lobbyWith(t, 2, DefaultOptions())
	pro := Options{Variant: VariantClassic, Pro: true, ReducedHandSize: true}
	mustOK(t, g.SetOptions(pid(0), pro))
	if g.Options != pro || g.Rules.MinCardsPerTurn != 3 || g.Rules.HandSize(2) != 6 || g.Revision != 3 {
		t.Fatalf("after SetOptions: %+v", g)
	}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{OptionsChanged{Options: pro}}) {
		t.Fatalf("events = %#v", evs)
	}
	mustOK(t, g.SetOptions(pid(0), DefaultOptions()))
	if g.Rules.MinCardsPerTurn != 2 || g.Rules.HandSize(2) != 7 {
		t.Fatalf("rules not re-derived: %+v", g.Rules)
	}
	// Extreme is accepted, alone and combined with Pro and the reduced hand size (Q4).
	for _, o := range []Options{extreme, extremePro, extremeProReduced} {
		mustOK(t, g.SetOptions(pid(0), o))
		if g.Options != o || g.Rules.Variant != VariantExtreme {
			t.Fatalf("after SetOptions(%+v): %+v", o, g)
		}
	}
	if g.Rules.MinCardsPerTurn != 3 || g.Rules.HandSize(2) != 6 {
		t.Fatalf("extreme pro reduced rules: %+v", g.Rules)
	}
}

func TestNewGameExtreme(t *testing.T) {
	for _, o := range []Options{extreme, extremePro, extremeProReduced} {
		g, err := NewGame("K7MQ2", "h", "Anna", o)
		mustOK(t, err)
		if g.Options != o || g.Rules.Variant != VariantExtreme || g.Phase != PhaseLobby {
			t.Fatalf("NewGame(%+v) = %+v", o, g)
		}
	}
}

func TestStartExtremeDealsCommands(t *testing.T) {
	g := lobbyWith(t, 1, extreme)
	mustOK(t, g.Start(pid(0), noShuffle{}))
	if g.Phase != PhasePlaying {
		t.Fatalf("phase %s", g.Phase)
	}
	if got := values(g.Players[0].Hand); !reflect.DeepEqual(got, seq(92, 99)) {
		t.Fatalf("hand %v", got)
	}
	for _, c := range g.Players[0].Hand {
		want := CommandNone
		switch c.Value {
		case 92:
			want = CommandNoCommunication
		case 95:
			want = CommandStop
		}
		if c.Command != want || c.Kind != CardKindNumber {
			t.Errorf("card %+v, want command %q", c, want)
		}
	}
	commands := 0
	for _, c := range g.DrawPile {
		if c.Command != CommandNone {
			commands++
		}
	}
	if commands != 26 || len(g.DrawPile) != 90 {
		t.Fatalf("draw pile: %d cards, %d with commands", len(g.DrawPile), commands)
	}
	if top := g.DrawPile[len(g.DrawPile)-1]; top != NumberCard(91) {
		t.Fatalf("draw top %+v", top)
	}
}

func TestSetOptionsErrors(t *testing.T) {
	started := lobbyWith(t, 2, DefaultOptions())
	mustOK(t, started.Start(pid(0), noShuffle{}))
	tests := []struct {
		name string
		g    *Game
		by   PlayerID
		opts Options
		want error
	}{
		{"unknown", lobbyWith(t, 2, DefaultOptions()), "zz", DefaultOptions(), ErrUnknownPlayer},
		{"not host", lobbyWith(t, 2, DefaultOptions()), pid(1), DefaultOptions(), ErrNotHost},
		{"wrong phase", started, pid(0), DefaultOptions(), ErrWrongPhase},
		{"invalid", lobbyWith(t, 2, DefaultOptions()), pid(0), Options{Variant: VariantClassic, ReducedHandSize: true}, ErrInvalidOptions},
		{"extreme reduced without pro", lobbyWith(t, 2, DefaultOptions()), pid(0), Options{Variant: VariantExtreme, ReducedHandSize: true}, ErrInvalidOptions},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.SetOptions(tt.by, tt.opts) })
		})
	}
}

func TestStartDeals(t *testing.T) {
	optsList := []Options{
		DefaultOptions(),
		{Variant: VariantClassic, Pro: true},
		{Variant: VariantClassic, Pro: true, ReducedHandSize: true},
		extreme,
		extremeProReduced,
	}
	for _, o := range optsList {
		for n := 1; n <= 5; n++ {
			g := lobbyWith(t, n, o)
			mustOK(t, g.Start(pid(0), rand.New(rand.NewPCG(uint64(n), 7))))
			hs := g.Rules.HandSize(n)
			if want := 98 - n*hs; len(g.DrawPile) != want {
				t.Errorf("%+v n=%d: draw pile %d, want %d", o, n, len(g.DrawPile), want)
			}
			seen := map[int]bool{}
			for _, c := range g.DrawPile {
				seen[c.Value] = true
			}
			for _, p := range g.Players {
				if len(p.Hand) != hs {
					t.Errorf("%+v n=%d: hand %d, want %d", o, n, len(p.Hand), hs)
				}
				if !slices.IsSortedFunc(p.Hand, func(a, b Card) int { return a.Value - b.Value }) {
					t.Errorf("hand not sorted: %v", values(p.Hand))
				}
				for _, c := range p.Hand {
					seen[c.Value] = true
				}
			}
			if len(seen) != 98 {
				t.Errorf("%+v n=%d: %d distinct cards, want 98", o, n, len(seen))
			}
			for v := 2; v <= 99; v++ {
				if !seen[v] {
					t.Errorf("card %d missing", v)
				}
			}
			for _, p := range g.Piles {
				if len(p.Cards) != 0 {
					t.Errorf("pile not empty: %+v", p)
				}
			}
		}
	}
}

func TestStartMultiplayer(t *testing.T) {
	g := lobbyWith(t, 2, DefaultOptions())
	mustOK(t, g.Start(pid(0), noShuffle{}))
	if g.Phase != PhaseChoosingFirstPlayer || g.Revision != 3 || g.Turn.Number != 0 {
		t.Fatalf("after Start: phase %s rev %d turn %+v", g.Phase, g.Revision, g.Turn)
	}
	// Deal round-robin from the top (highest card with noShuffle).
	if got, want := values(g.Players[0].Hand), []int{87, 89, 91, 93, 95, 97, 99}; !reflect.DeepEqual(got, want) {
		t.Fatalf("p0 hand %v, want %v", got, want)
	}
	if got, want := values(g.Players[1].Hand), []int{86, 88, 90, 92, 94, 96, 98}; !reflect.DeepEqual(got, want) {
		t.Fatalf("p1 hand %v, want %v", got, want)
	}
	if len(g.DrawPile) != 84 || g.DrawPile[len(g.DrawPile)-1].Value != 85 {
		t.Fatalf("draw pile top = %v", g.DrawPile[len(g.DrawPile)-1])
	}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{GameStarted{}}) {
		t.Fatalf("events = %#v", evs)
	}
	// Rules are frozen: SetOptions is no longer possible.
	expectErr(t, g, ErrWrongPhase, func(g *Game) error { return g.SetOptions(pid(0), DefaultOptions()) })
}

func TestStartUsesRNG(t *testing.T) {
	g := lobbyWith(t, 1, DefaultOptions())
	mustOK(t, g.Start(pid(0), reverseShuffle{}))
	if got := values(g.Players[0].Hand); !reflect.DeepEqual(got, seq(2, 9)) {
		t.Fatalf("hand %v", got)
	}
}

func TestStartDeterministic(t *testing.T) {
	a := lobbyWith(t, 3, DefaultOptions())
	b := lobbyWith(t, 3, DefaultOptions())
	mustOK(t, a.Start(pid(0), rand.New(rand.NewPCG(42, 1))))
	mustOK(t, b.Start(pid(0), rand.New(rand.NewPCG(42, 1))))
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed gave different games")
	}
	c := lobbyWith(t, 3, DefaultOptions())
	mustOK(t, c.Start(pid(0), rand.New(rand.NewPCG(43, 1))))
	if reflect.DeepEqual(a.DrawPile, c.DrawPile) {
		t.Fatal("different seeds gave identical draw piles")
	}
}

func TestStartSolo(t *testing.T) {
	g := lobbyWith(t, 1, DefaultOptions())
	mustOK(t, g.Start(pid(0), noShuffle{}))
	if g.Phase != PhasePlaying || g.Revision != 2 {
		t.Fatalf("phase %s rev %d", g.Phase, g.Revision)
	}
	if g.Turn.Number != 1 || g.Turn.ActiveSeat != 0 || g.Turn.StartHandSize != 8 || g.Turn.Plays != nil {
		t.Fatalf("turn %+v", g.Turn)
	}
	want := []Event{GameStarted{}, FirstPlayerChosen{PlayerID: pid(0)}, TurnStarted{PlayerID: pid(0), Number: 1}}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, want) {
		t.Fatalf("events = %#v", evs)
	}
	if len(g.DrawPile) != 90 {
		t.Fatalf("draw pile %d", len(g.DrawPile))
	}
}

func TestStartErrors(t *testing.T) {
	started := lobbyWith(t, 2, DefaultOptions())
	mustOK(t, started.Start(pid(0), noShuffle{}))
	tests := []struct {
		name string
		g    *Game
		by   PlayerID
		want error
	}{
		{"unknown", lobbyWith(t, 2, DefaultOptions()), "zz", ErrUnknownPlayer},
		{"not host", lobbyWith(t, 2, DefaultOptions()), pid(1), ErrNotHost},
		{"already started", started, pid(0), ErrWrongPhase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.Start(tt.by, noShuffle{}) })
		})
	}
}

func TestChooseFirstPlayer(t *testing.T) {
	g := lobbyWith(t, 3, DefaultOptions())
	mustOK(t, g.Start(pid(0), noShuffle{}))
	g.PullEvents()
	mustOK(t, g.ChooseFirstPlayer(pid(0), pid(2)))
	if g.Phase != PhasePlaying || g.Turn.Number != 1 || g.Turn.ActiveSeat != 2 || g.Turn.StartHandSize != 6 || g.Revision != 5 {
		t.Fatalf("after choose: phase %s turn %+v rev %d", g.Phase, g.Turn, g.Revision)
	}
	want := []Event{FirstPlayerChosen{PlayerID: pid(2)}, TurnStarted{PlayerID: pid(2), Number: 1}}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, want) {
		t.Fatalf("events = %#v", evs)
	}
}

func TestChooseFirstPlayerErrors(t *testing.T) {
	choosing := func() *Game {
		g := lobbyWith(t, 2, DefaultOptions())
		mustOK(t, g.Start(pid(0), noShuffle{}))
		return g
	}
	playing := choosing()
	mustOK(t, playing.ChooseFirstPlayer(pid(0), pid(1)))
	tests := []struct {
		name      string
		g         *Game
		by, first PlayerID
		want      error
	}{
		{"unknown by", choosing(), "zz", pid(1), ErrUnknownPlayer},
		{"unknown first", choosing(), pid(0), "zz", ErrUnknownPlayer},
		{"not host", choosing(), pid(1), pid(1), ErrNotHost},
		{"lobby", lobbyWith(t, 2, DefaultOptions()), pid(0), pid(1), ErrWrongPhase},
		{"already playing", playing, pid(0), pid(0), ErrWrongPhase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.ChooseFirstPlayer(tt.by, tt.first) })
		})
	}
}

// finishedGame returns a lost 3-player Pro game (p1 stuck).
func finishedGame(t *testing.T) *Game {
	t.Helper()
	o := Options{Variant: VariantClassic, Pro: true, ReducedHandSize: true}
	g := table{
		opts:  o,
		hands: [][]int{{50}, {51, 52}, {}},
		draw:  []int{53},
		piles: [][]int{{95}, {96}, {5}, {6}},
	}.build(t)
	g.HostID = pid(2)
	g.checkStuck()
	if g.Phase != PhaseFinished {
		t.Fatal("setup: game not finished")
	}
	g.Revision = 9
	g.PullEvents()
	return g
}

func TestRematch(t *testing.T) {
	for _, by := range []PlayerID{pid(0), pid(1), pid(2)} { // any player, not only the host
		t.Run(string(by), func(t *testing.T) {
			g := finishedGame(t)
			next, err := g.Rematch(by, "NEXT2")
			mustOK(t, err)
			// old game
			if g.NextGameID != "NEXT2" || g.Revision != 10 || g.Phase != PhaseFinished {
				t.Fatalf("old game: next %q rev %d phase %s", g.NextGameID, g.Revision, g.Phase)
			}
			if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{RematchCreated{GameID: "NEXT2"}}) {
				t.Fatalf("events = %#v", evs)
			}
			// new game
			r := mustRules(t, g.Options)
			want := &Game{
				ID:      "NEXT2",
				Options: g.Options,
				Rules:   r,
				Phase:   PhaseLobby,
				HostID:  by,
				Players: []Player{
					{ID: pid(0), Name: "Player 0", Hand: []Card{}},
					{ID: pid(1), Name: "Player 1", Hand: []Card{}},
					{ID: pid(2), Name: "Player 2", Hand: []Card{}},
				},
				DrawPile: []Card{},
				Piles:    NewPiles(r),
				Revision: 1,
			}
			if !reflect.DeepEqual(next, want) {
				t.Fatalf("new game:\n got %+v\nwant %+v", next, want)
			}
			// independent of the old game
			next.Players[0].Name = "changed"
			if g.Players[0].Name != "Player 0" {
				t.Fatal("new game aliases old players")
			}
			// the new game is a working lobby
			mustOK(t, next.Start(by, noShuffle{}))
		})
	}
}

func TestRematchErrors(t *testing.T) {
	playing := table{hands: [][]int{{10, 20}, {30, 40}}, draw: []int{50}}.build(t)
	exists := finishedGame(t)
	exists.NextGameID = "OLD22"
	tests := []struct {
		name string
		g    *Game
		by   PlayerID
		want error
	}{
		{"unknown", finishedGame(t), "zz", ErrUnknownPlayer},
		{"playing", playing, pid(0), ErrWrongPhase},
		{"lobby", lobbyWith(t, 2, DefaultOptions()), pid(0), ErrWrongPhase},
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

func TestEndGame(t *testing.T) {
	playing := func() *Game {
		return table{hands: [][]int{{10, 20}, {30}}, draw: []int{40, 50, 60}, seat: 1}.build(t)
	}
	choosing := func() *Game {
		g := lobbyWith(t, 2, DefaultOptions())
		mustOK(t, g.Start(pid(0), noShuffle{}))
		g.PullEvents()
		return g
	}
	tests := []struct {
		name        string
		g           *Game
		hands, draw int
	}{
		{"playing", playing(), 3, 3},
		{"choosing first player", choosing(), 14, 84},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.g
			rev := g.Revision
			mustOK(t, g.EndGame(pid(0)))
			score := tt.hands + tt.draw
			want := &Result{Outcome: OutcomeAborted, Score: score, CardsLeftInHands: tt.hands, CardsLeftInDrawPile: tt.draw}
			if g.Phase != PhaseFinished || !reflect.DeepEqual(g.Result, want) || g.Revision != rev+1 {
				t.Fatalf("phase %s result %+v rev %d", g.Phase, g.Result, g.Revision)
			}
			if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{GameFinished{Outcome: OutcomeAborted, Score: score}}) {
				t.Fatalf("events = %#v", evs)
			}
			v, err := g.ViewFor(pid(1))
			mustOK(t, err)
			if v.Result == nil || v.Result.Outcome != OutcomeAborted || v.Players[0].RevealedHand == nil {
				t.Fatalf("view after end: %+v", v)
			}
			if _, err := g.Rematch(pid(1), "NEXT2"); err != nil {
				t.Fatalf("rematch after end: %v", err)
			}
		})
	}
}

func TestEndGameErrors(t *testing.T) {
	playing := func() *Game { return table{hands: [][]int{{10, 20}, {30}}, draw: []int{40}}.build(t) }
	tests := []struct {
		name string
		g    *Game
		by   PlayerID
		want error
	}{
		{"unknown", playing(), "zz", ErrUnknownPlayer},
		{"not host", playing(), pid(1), ErrNotHost},
		{"lobby", lobbyWith(t, 2, DefaultOptions()), pid(0), ErrWrongPhase},
		{"finished", finishedGame(t), pid(2), ErrWrongPhase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.EndGame(tt.by) })
		})
	}
}
