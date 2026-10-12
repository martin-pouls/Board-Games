package domain

import (
	"reflect"
	"slices"
	"testing"
)

func TestPlayErrors(t *testing.T) {
	base := table{hands: []string{"5H 5C 7D X", "6H"}, enemy: "JS"}
	lost := base.build(t)
	mustOK(t, lost.GiveUp(pid(1)))
	lost.PullEvents()
	tests := []struct {
		name  string
		g     *Game
		by    PlayerID
		cards string
		want  error
	}{
		{"unknown", base.build(t), "zz", "5H", ErrUnknownPlayer},
		{"lobby", lobbyWith(t, 2), pid(0), "5H", ErrWrongPhase},
		{"after end", lost, pid(0), "5H", ErrWrongPhase},
		{"not your turn", base.build(t), pid(1), "6H", ErrNotYourTurn},
		{"wrong step", table{hands: base.hands, enemy: "JS", step: StepDiscard}.build(t), pid(0), "5H", ErrWrongStep},
		{"not in hand", base.build(t), pid(0), "6H", ErrCardNotInHand},
		{"too many copies", base.build(t), pid(0), "5H 5H", ErrCardNotInHand},
		{"mixed ranks", base.build(t), pid(0), "5H 7D", ErrIllegalPlay},
		{"jester not alone", base.build(t), pid(0), "X 5H", ErrIllegalPlay},
		{"empty", base.build(t), pid(0), "", ErrIllegalPlay},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.Play(tt.by, cs(tt.cards), noShuffle{}) })
		})
	}
}

// lastPlayed returns the CardsPlayed event among evs.
func lastPlayed(t *testing.T, evs []Event) CardsPlayed {
	t.Helper()
	for i := len(evs) - 1; i >= 0; i-- {
		if cp, ok := evs[i].(CardsPlayed); ok {
			return cp
		}
	}
	t.Fatalf("no CardsPlayed in %#v", evs)
	return CardsPlayed{}
}

func TestHeal(t *testing.T) {
	t.Run("heals N under the Tavern", func(t *testing.T) {
		g := table{hands: []string{"5H KC", ""}, enemy: "JS", tavern: "10C", discard: "2C 3C 4C 6D 7D 8D 9D"}.build(t)
		mustOK(t, g.Play(pid(0), cs("5H"), noShuffle{}))
		evs := wantEventTypes(t, g, "CardsPlayed")
		if cp := lastPlayed(t, evs); cp.Healed != 5 || cp.Drawn != 0 || cp.Damage != 5 || cp.Value != 5 || cp.BlockedSuits != nil {
			t.Fatalf("event %+v", cp)
		}
		if !slices.Equal(g.Tavern, cs("4C 6D 7D 8D 9D 10C")) || !slices.Equal(g.DiscardPile, cs("2C 3C")) {
			t.Fatalf("tavern %v discard %v", g.Tavern, g.DiscardPile)
		}
		if g.Step != StepDiscard || g.Revision != 2 {
			t.Fatalf("step %s rev %d", g.Step, g.Revision)
		}
		checkShape(t, g)
	})
	t.Run("fewer when the discard is small", func(t *testing.T) {
		g := table{hands: []string{"7H KC", ""}, enemy: "JS", tavern: "10C", discard: "2C 3C"}.build(t)
		mustOK(t, g.Play(pid(0), cs("7H"), noShuffle{}))
		if cp := lastPlayed(t, g.PullEvents()); cp.Healed != 2 || len(g.DiscardPile) != 0 || !slices.Equal(g.Tavern, cs("2C 3C 10C")) {
			t.Fatalf("healed %d tavern %v discard %v", cp.Healed, g.Tavern, g.DiscardPile)
		}
	})
	t.Run("shuffles the discard with the RNG", func(t *testing.T) {
		g := table{hands: []string{"2H KC", ""}, enemy: "JS", tavern: "10C", discard: "2C 3C 4C 6D 7D 8D 9D"}.build(t)
		mustOK(t, g.Play(pid(0), cs("2H"), reverseShuffle{}))
		if !slices.Equal(g.Tavern, cs("3C 2C 10C")) || !slices.Equal(g.DiscardPile, cs("9D 8D 7D 6D 4C")) {
			t.Fatalf("tavern %v discard %v", g.Tavern, g.DiscardPile)
		}
	})
	t.Run("hearts vs hearts enemy", func(t *testing.T) {
		g := table{hands: []string{"5H KC", ""}, enemy: "JH", tavern: "10C", discard: "2C 3C"}.build(t)
		mustOK(t, g.Play(pid(0), cs("5H"), noShuffle{}))
		cp := lastPlayed(t, g.PullEvents())
		if cp.Healed != 0 || cp.Damage != 5 || !slices.Equal(cp.BlockedSuits, []Suit{SuitHearts}) || len(g.DiscardPile) != 2 {
			t.Fatalf("event %+v discard %v", cp, g.DiscardPile)
		}
	})
}

func TestDraw(t *testing.T) {
	t.Run("round robin from the active player, skipping full hands, until all are full", func(t *testing.T) {
		// 3 players, max hand 6. Seat 1 plays 8♦ (1 card left); seat 2 is full.
		g := table{
			hands:  []string{"2H 3H 4H 5H 6H", "8D 2C", "2S 3S 4S 5S 6S 7S"},
			enemy:  "JS",
			tavern: "AC 2D 3D 4D 5D 6D 7D 9C",
			seat:   1,
		}.build(t)
		mustOK(t, g.Play(pid(1), cs("8D"), noShuffle{}))
		cp := lastPlayed(t, g.PullEvents())
		if cp.Drawn != 6 {
			t.Fatalf("drawn %d, want 6 (everyone full)", cp.Drawn)
		}
		if !slices.Equal(g.Players[1].Hand, sorted("2C 9C 6D 5D 4D 3D")) || !slices.Equal(g.Players[0].Hand, sorted("2H 3H 4H 5H 6H 7D")) ||
			len(g.Players[2].Hand) != 6 || !slices.Equal(g.Tavern, cs("AC 2D")) {
			t.Fatalf("hands %v / %v / %v tavern %v", g.Players[0].Hand, g.Players[1].Hand, g.Players[2].Hand, g.Tavern)
		}
		checkShape(t, g)
	})
	t.Run("stops when the Tavern is empty", func(t *testing.T) {
		g := table{hands: []string{"5D KC", ""}, enemy: "JS", tavern: "9C 10C"}.build(t)
		mustOK(t, g.Play(pid(0), cs("5D"), noShuffle{}))
		if cp := lastPlayed(t, g.PullEvents()); cp.Drawn != 2 || len(g.Tavern) != 0 ||
			!slices.Equal(g.Players[0].Hand, sorted("KC 10C")) || !slices.Equal(g.Players[1].Hand, cs("9C")) {
			t.Fatalf("drawn %d hands %v / %v", cp.Drawn, g.Players[0].Hand, g.Players[1].Hand)
		}
	})
	t.Run("diamonds vs diamonds enemy", func(t *testing.T) {
		g := table{hands: []string{"5D KC", ""}, enemy: "JD", tavern: "9C 10C"}.build(t)
		mustOK(t, g.Play(pid(0), cs("5D"), noShuffle{}))
		if cp := lastPlayed(t, g.PullEvents()); cp.Drawn != 0 || len(g.Tavern) != 2 || !slices.Equal(cp.BlockedSuits, []Suit{SuitDiamonds}) {
			t.Fatalf("event %+v", cp)
		}
	})
}

func TestHeartsBeforeDiamonds(t *testing.T) {
	// The Tavern is empty: only the healed cards can be drawn.
	g := table{hands: []string{"5H 5D", ""}, enemy: "JS", discard: "2C 3C 4C 6D 7D"}.build(t)
	mustOK(t, g.Play(pid(0), cs("5H 5D"), noShuffle{}))
	cp := lastPlayed(t, g.PullEvents())
	if cp.Healed != 5 || cp.Drawn != 5 || cp.Damage != 10 || len(g.Tavern) != 0 {
		t.Fatalf("event %+v tavern %v", cp, g.Tavern)
	}
	if !slices.Equal(g.Players[0].Hand, sorted("7D 4C 2C")) || !slices.Equal(g.Players[1].Hand, sorted("6D 3C")) {
		t.Fatalf("hands %v / %v", g.Players[0].Hand, g.Players[1].Hand)
	}
}

func TestSameSuitPowerOnce(t *testing.T) {
	g := table{hands: []string{"AH 5H KC", ""}, enemy: "JS", discard: "2C 3C 4C 6D 7D 8D 9D 10D"}.build(t)
	mustOK(t, g.Play(pid(0), cs("AH 5H"), noShuffle{}))
	if cp := lastPlayed(t, g.PullEvents()); cp.Healed != 6 || cp.Value != 6 {
		t.Fatalf("event %+v", cp)
	}
	g = table{hands: []string{"AC 8C KD", ""}, enemy: "JS"}.build(t)
	mustOK(t, g.Play(pid(0), cs("AC 8C"), noShuffle{}))
	if cp := lastPlayed(t, g.PullEvents()); cp.Damage != 18 || g.Damage() != 18 {
		t.Fatalf("event %+v", cp)
	}
}

func TestRulesExamples(t *testing.T) {
	t.Run("3♦ + 3♠ + 3♣", func(t *testing.T) {
		g := table{hands: []string{"3D 3S 3C", ""}, enemy: "JH", tavern: "AH 2H 4H 5H 6H 7H 8H 9H 10H AC"}.build(t)
		mustOK(t, g.Play(pid(0), cs("3D 3S 3C"), noShuffle{}))
		cp := lastPlayed(t, g.PullEvents())
		if cp.Drawn != 9 || cp.Damage != 18 || g.Shield() != 9 || g.EffectiveAttack() != 1 || g.Step != StepDiscard {
			t.Fatalf("event %+v shield %d attack %d step %s", cp, g.Shield(), g.EffectiveAttack(), g.Step)
		}
		if len(g.Players[0].Hand) != 5 || len(g.Players[1].Hand) != 4 {
			t.Fatalf("hand sizes %d / %d", len(g.Players[0].Hand), len(g.Players[1].Hand))
		}
	})
	t.Run("8♦ + A♣", func(t *testing.T) {
		g := table{hands: []string{"8D AC", ""}, enemy: "JH", tavern: "AH 2H 4H 5H 6H 7H 8H 9H 10H 2C"}.build(t)
		mustOK(t, g.Play(pid(0), cs("8D AC"), noShuffle{}))
		cp := lastPlayed(t, g.PullEvents())
		if cp.Drawn != 9 || cp.Damage != 18 || g.Shield() != 0 {
			t.Fatalf("event %+v", cp)
		}
	})
}

func TestDefeatExact(t *testing.T) {
	// 10 + 2×5 = 20 = the Jack's health.
	g := table{
		hands:   []string{"5C 2D KH", "6H"},
		enemy:   "JS",
		castle:  "KS JC",
		tavern:  "9H",
		played:  []played{{1, "10H"}},
		yielded: []int{1},
	}.build(t)
	mustOK(t, g.Play(pid(0), cs("5C"), noShuffle{}))
	evs := g.PullEvents()
	want := []Event{
		CardsPlayed{PlayerID: pid(0), Cards: cs("5C"), Value: 5, Damage: 10},
		EnemyDefeated{Enemy: c("JS"), Exact: true},
		EnemyRevealed{Enemy: c("JC")},
		TurnStarted{PlayerID: pid(0), Number: 2},
	}
	if !reflect.DeepEqual(evs, want) {
		t.Fatalf("events =\n%#v\nwant\n%#v", evs, want)
	}
	if g.Tavern[len(g.Tavern)-1] != c("JS") || !sameCards(g.DiscardPile, "10H 5C") || g.Played != nil ||
		*g.Enemy != c("JC") || !slices.Equal(g.Castle, cs("KS")) {
		t.Fatalf("tavern %v discard %v played %v enemy %v castle %v", g.Tavern, g.DiscardPile, g.Played, g.Enemy, g.Castle)
	}
	if g.Turn.ActiveSeat != 0 || g.Step != StepPlay || g.Damage() != 0 || g.Shield() != 0 || g.JesterPlayed() ||
		!g.Players[1].YieldedLastTurn || g.Players[0].YieldedLastTurn {
		t.Fatalf("after defeat: %+v", g)
	}
	// The next ♦ draws the defeated Jack.
	mustOK(t, g.Play(pid(0), cs("2D"), noShuffle{}))
	if !slices.Contains(g.Players[0].Hand, c("JS")) || !slices.Contains(g.Players[1].Hand, c("9H")) {
		t.Fatalf("hands %v / %v", g.Players[0].Hand, g.Players[1].Hand)
	}
	checkShape(t, g)
}

func TestDefeatOverkill(t *testing.T) {
	g := table{hands: []string{"6C KH", "6H"}, enemy: "JS", castle: "JC", played: []played{{1, "10H"}, {1, "X"}}}.build(t)
	mustOK(t, g.Play(pid(0), cs("6C"), noShuffle{}))
	evs := wantEventTypes(t, g, "CardsPlayed", "EnemyDefeated", "EnemyRevealed", "TurnStarted")
	if evs[1] != (EnemyDefeated{Enemy: c("JS"), Exact: false}) {
		t.Fatalf("defeat %#v", evs[1])
	}
	if !sameCards(g.DiscardPile, "10H X 6C JS") || len(g.Tavern) != 0 {
		t.Fatalf("discard %v tavern %v", g.DiscardPile, g.Tavern)
	}
	// Immunity is back for the new enemy.
	if g.ImmuneSuit() != SuitClubs {
		t.Fatalf("immune %q", g.ImmuneSuit())
	}
}

func TestDamageAccumulatesAcrossTurns(t *testing.T) {
	g := table{hands: []string{"7H 9S KC", "8D QH"}, enemy: "JC", castle: "QC"}.build(t)
	mustOK(t, g.Play(pid(0), cs("7H"), noShuffle{}))
	mustOK(t, g.Discard(pid(0), cs("KC")))
	mustOK(t, g.Play(pid(1), cs("8D"), noShuffle{}))
	if g.Damage() != 15 || g.Step != StepDiscard {
		t.Fatalf("damage %d step %s", g.Damage(), g.Step)
	}
	mustOK(t, g.Discard(pid(1), cs("QH")))
	mustOK(t, g.Play(pid(0), cs("9S"), noShuffle{}))
	if *g.Enemy != c("QC") || g.Turn.ActiveSeat != 0 || g.Step != StepPlay {
		t.Fatalf("enemy %v seat %d step %s", g.Enemy, g.Turn.ActiveSeat, g.Step)
	}
}

func TestWinOnLastKing(t *testing.T) {
	g := table{hands: []string{"2H 5S", "6H"}, enemy: "KS", played: []played{{1, "10H"}, {1, "10D"}, {1, "9C"}}}.build(t)
	mustOK(t, g.Play(pid(0), cs("2H"), noShuffle{}))
	evs := g.PullEvents()
	want := []Event{
		CardsPlayed{PlayerID: pid(0), Cards: cs("2H"), Value: 2, Damage: 2},
		EnemyDefeated{Enemy: c("KS"), Exact: true},
		GameWon{Grade: GradeNone},
	}
	if !reflect.DeepEqual(evs, want) {
		t.Fatalf("events %#v", evs)
	}
	if g.Phase != PhaseWon || g.Enemy != nil || g.Tavern[len(g.Tavern)-1] != c("KS") || *g.Result != (Result{}) || g.Step != StepNone {
		t.Fatalf("after win: %+v", g)
	}
	checkShape(t, g)
}

func TestSufferDamage(t *testing.T) {
	t.Run("attack fully shielded skips Step 4", func(t *testing.T) {
		g := table{hands: []string{"2C 3H", "6H", "7H"}, enemy: "JH", played: []played{{1, "10S"}}}.build(t)
		mustOK(t, g.Play(pid(0), cs("2C"), noShuffle{}))
		evs := wantEventTypes(t, g, "CardsPlayed", "AttackBlocked", "TurnStarted")
		if evs[1] != (AttackBlocked{PlayerID: pid(0)}) || g.Turn.ActiveSeat != 1 || g.Step != StepPlay || g.Turn.Number != 2 {
			t.Fatalf("seat %d step %s", g.Turn.ActiveSeat, g.Step)
		}
	})
	t.Run("discard passes the turn clockwise", func(t *testing.T) {
		for _, seat := range []int{1, 2} {
			hands := []string{"6H", "7H", "8H"}
			hands[seat] = "5H 7C KS"
			g := table{hands: hands, enemy: "JH", seat: seat, step: StepDiscard, played: []played{{0, "3S"}}}.build(t)
			mustOK(t, g.Discard(pid(seat), cs("7C")))
			evs := g.PullEvents()
			next := (seat + 1) % 3
			want := []Event{
				CardsDiscarded{PlayerID: pid(seat), Cards: cs("7C"), Attack: 7},
				TurnStarted{PlayerID: pid(next), Number: 2},
			}
			if !reflect.DeepEqual(evs, want) || g.Turn.ActiveSeat != next || !slices.Equal(g.Players[seat].Hand, cs("5H KS")) ||
				!slices.Equal(g.DiscardPile, cs("7C")) {
				t.Fatalf("seat %d: events %#v hand %v", seat, evs, g.Players[seat].Hand)
			}
		}
	})
	t.Run("discard down to an empty hand", func(t *testing.T) {
		g := table{hands: []string{"KS", "6H"}, enemy: "JH", step: StepDiscard}.build(t)
		mustOK(t, g.Discard(pid(0), cs("KS")))
		if len(g.Players[0].Hand) != 0 || g.Players[0].Hand == nil || g.Phase != PhasePlaying {
			t.Fatalf("hand %v phase %s", g.Players[0].Hand, g.Phase)
		}
	})
	t.Run("hand too weak loses at Step 4 start", func(t *testing.T) {
		g := table{hands: []string{"2H 3C 4D", "KH"}, enemy: "JS"}.build(t)
		mustOK(t, g.Play(pid(0), cs("2H"), noShuffle{}))
		evs := wantEventTypes(t, g, "CardsPlayed", "GameLost")
		if evs[1] != (GameLost{Reason: LossCannotCover, PlayerID: pid(0)}) || g.Phase != PhaseLost || g.Enemy == nil {
			t.Fatalf("after: %+v", g)
		}
		checkShape(t, g)
	})
	t.Run("hand exactly covering is not a loss", func(t *testing.T) {
		g := table{hands: []string{"2H 3C 7D", "KH"}, enemy: "JS"}.build(t)
		mustOK(t, g.Play(pid(0), cs("2H"), noShuffle{}))
		if g.Phase != PhasePlaying || g.Step != StepDiscard {
			t.Fatalf("phase %s step %s", g.Phase, g.Step)
		}
	})
}

func TestDiscardErrors(t *testing.T) {
	disc := table{hands: []string{"AH 5H 7C X KS", "6H"}, enemy: "JH", step: StepDiscard}
	tests := []struct {
		name  string
		g     *Game
		by    PlayerID
		cards string
		want  error
	}{
		{"unknown", disc.build(t), "zz", "KS", ErrUnknownPlayer},
		{"lobby", lobbyWith(t, 2), pid(0), "KS", ErrWrongPhase},
		{"not your turn", disc.build(t), pid(1), "6H", ErrNotYourTurn},
		{"wrong step", table{hands: disc.hands, enemy: "JH"}.build(t), pid(0), "KS", ErrWrongStep},
		{"not in hand", disc.build(t), pid(0), "QS", ErrCardNotInHand},
		{"not enough", disc.build(t), pid(0), "5H", ErrIllegalDiscard},
		{"superfluous", disc.build(t), pid(0), "AH KS", ErrIllegalDiscard},
		{"jester superfluous", disc.build(t), pid(0), "X KS", ErrIllegalDiscard},
		{"empty", disc.build(t), pid(0), "", ErrIllegalDiscard},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectErr(t, tt.g, tt.want, func(g *Game) error { return g.Discard(tt.by, cs(tt.cards)) })
		})
	}
}

func TestJesterFlow(t *testing.T) {
	g := table{hands: []string{"X 5C", "6H", "7H 3C"}, enemy: "JC", yielded: []int{0}, played: []played{{1, "4C"}}}.build(t)
	g.Players[2].Signal = SignalGoodPlay // stale signal from an earlier choice
	mustOK(t, g.Play(pid(0), cs("X"), noShuffle{}))
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, []Event{JesterPlayed{PlayerID: pid(0)}}) {
		t.Fatalf("events %#v", evs)
	}
	if g.Step != StepChooseNext || g.Turn.ActiveSeat != 0 || g.Players[0].YieldedLastTurn || g.Players[2].Signal != SignalNone ||
		!g.JesterPlayed() || g.ImmuneSuit() != SuitNone || g.Damage() != 4 {
		t.Fatalf("after jester: %+v", g)
	}
	checkShape(t, g)

	// Signals by others.
	expectErr(t, g, ErrIllegalSignal, func(g *Game) error { return g.SetJesterSignal(pid(0), SignalGoodPlay) })
	expectErr(t, g, ErrIllegalSignal, func(g *Game) error { return g.SetJesterSignal(pid(1), "maybe") })
	expectErr(t, g, ErrUnknownPlayer, func(g *Game) error { return g.SetJesterSignal("zz", SignalGoodPlay) })
	mustOK(t, g.SetJesterSignal(pid(1), SignalGoodPlay))
	mustOK(t, g.SetJesterSignal(pid(2), SignalRatherNot))
	mustOK(t, g.SetJesterSignal(pid(2), SignalRatherNot)) // unchanged still emits
	if evs := g.PullEvents(); len(evs) != 3 || evs[2] != (JesterSignalSet{PlayerID: pid(2), Signal: SignalRatherNot}) {
		t.Fatalf("events %#v", evs)
	}
	if g.Players[1].Signal != SignalGoodPlay || g.Players[2].Signal != SignalRatherNot {
		t.Fatal("signals not set")
	}

	// Only the chooser chooses; next must be seated.
	expectErr(t, g, ErrNotYourTurn, func(g *Game) error { return g.ChooseNext(pid(1), pid(1)) })
	expectErr(t, g, ErrUnknownPlayer, func(g *Game) error { return g.ChooseNext(pid(0), "zz") })
	expectErr(t, g, ErrWrongStep, func(g *Game) error { return g.Play(pid(0), cs("5C"), noShuffle{}) })
	expectErr(t, g, ErrWrongStep, func(g *Game) error { return g.Yield(pid(0)) })

	mustOK(t, g.ChooseNext(pid(0), pid(2)))
	want := []Event{NextPlayerChosen{ByPlayerID: pid(0), PlayerID: pid(2)}, TurnStarted{PlayerID: pid(2), Number: 2}}
	if evs := g.PullEvents(); !reflect.DeepEqual(evs, want) {
		t.Fatalf("events %#v", evs)
	}
	if g.Turn.ActiveSeat != 2 || g.Step != StepPlay || g.Players[1].Signal != SignalNone || g.Players[2].Signal != SignalNone {
		t.Fatalf("after choose: %+v", g)
	}
	expectErr(t, g, ErrWrongStep, func(g *Game) error { return g.SetJesterSignal(pid(1), SignalNone) })
	expectErr(t, g, ErrWrongStep, func(g *Game) error { return g.ChooseNext(pid(2), pid(0)) })

	// Clubs now double against the ♣ enemy; the earlier 4♣ is not re-doubled.
	mustOK(t, g.Play(pid(2), cs("3C"), noShuffle{}))
	if g.Damage() != 4+6 {
		t.Fatalf("damage %d", g.Damage())
	}
}

func TestJesterChooseSelf(t *testing.T) {
	g := table{hands: []string{"X 5C", "6H"}, enemy: "JS"}.build(t)
	mustOK(t, g.Play(pid(0), cs("X"), noShuffle{}))
	mustOK(t, g.ChooseNext(pid(0), pid(0)))
	if g.Turn.ActiveSeat != 0 || g.Step != StepPlay || g.Turn.Number != 2 {
		t.Fatalf("after choosing self: %+v", g)
	}
}

func TestJesterChosenPlayerCannotAct(t *testing.T) {
	// The chooser played their last card; the others yielded, so choosing themselves loses.
	g := table{hands: []string{"X", "6H", "7H"}, enemy: "JS", yielded: []int{1, 2}}.build(t)
	mustOK(t, g.Play(pid(0), cs("X"), noShuffle{}))
	mustOK(t, g.ChooseNext(pid(0), pid(0)))
	evs := wantEventTypes(t, g, "JesterPlayed", "NextPlayerChosen", "TurnStarted", "GameLost")
	if evs[3] != (GameLost{Reason: LossCannotAct, PlayerID: pid(0)}) {
		t.Fatalf("loss %#v", evs[3])
	}
}

func TestJesterGoesToDiscardOnDefeat(t *testing.T) {
	g := table{hands: []string{"10H 2C", "6H"}, enemy: "JS", castle: "JC", played: []played{{1, "X"}, {1, "10D"}}}.build(t)
	mustOK(t, g.Play(pid(0), cs("10H"), noShuffle{}))
	if !slices.Contains(g.DiscardPile, Jester()) || g.Tavern[len(g.Tavern)-1] != c("JS") {
		t.Fatalf("discard %v tavern %v", g.DiscardPile, g.Tavern)
	}
}

func TestSignalErrors(t *testing.T) {
	expectErr(t, lobbyWith(t, 2), ErrWrongPhase, func(g *Game) error { return g.SetJesterSignal(pid(1), SignalGoodPlay) })
	g := table{hands: []string{"5C", "6H"}, enemy: "JS"}.build(t)
	expectErr(t, g, ErrWrongStep, func(g *Game) error { return g.SetJesterSignal(pid(1), SignalGoodPlay) })
}

func TestYield(t *testing.T) {
	t.Run("goes to Step 4 without powers", func(t *testing.T) {
		g := table{hands: []string{"KC", "6H"}, enemy: "JS", tavern: "2H", discard: "3H"}.build(t)
		mustOK(t, g.Yield(pid(0)))
		wantEventTypes(t, g, "Yielded")
		if g.Step != StepDiscard || !g.Players[0].YieldedLastTurn || g.Damage() != 0 || len(g.Tavern) != 1 || g.Revision != 2 {
			t.Fatalf("after yield: %+v", g)
		}
		mustOK(t, g.Discard(pid(0), cs("KC")))
		// 2 players: the other player just yielded, so seat 1 cannot.
		expectErr(t, g, ErrYieldNotAllowed, func(g *Game) error { return g.Yield(pid(1)) })
		mustOK(t, g.Play(pid(1), cs("6H"), noShuffle{}))
		if g.Players[1].YieldedLastTurn {
			t.Fatal("play did not clear the yield flag of seat 1")
		}
	})
	t.Run("3p: two yields in a row forbid the third", func(t *testing.T) {
		g := table{hands: []string{"KC", "6H", "7H"}, enemy: "JS", yielded: []int{1, 2}}.build(t)
		expectErr(t, g, ErrYieldNotAllowed, func(g *Game) error { return g.Yield(pid(0)) })
		g = table{hands: []string{"KC", "6H", "7H"}, enemy: "JS", yielded: []int{1}}.build(t)
		mustOK(t, g.Yield(pid(0)))
	})
	t.Run("a play clears the flag", func(t *testing.T) {
		g := table{hands: []string{"KC 2D", "6H", "7H"}, enemy: "JS", yielded: []int{0, 1}}.build(t)
		mustOK(t, g.Play(pid(0), cs("2D"), noShuffle{}))
		if g.Players[0].YieldedLastTurn || !g.Players[1].YieldedLastTurn {
			t.Fatal("flags")
		}
	})
	t.Run("fully shielded yield passes the turn", func(t *testing.T) {
		g := table{hands: []string{"KC", "6H"}, enemy: "JH", played: []played{{1, "10S"}}}.build(t)
		mustOK(t, g.Yield(pid(0)))
		wantEventTypes(t, g, "Yielded", "AttackBlocked", "TurnStarted")
	})
	t.Run("yield with a weak hand loses", func(t *testing.T) {
		g := table{hands: []string{"2C", "6H"}, enemy: "JS"}.build(t)
		mustOK(t, g.Yield(pid(0)))
		wantEventTypes(t, g, "Yielded", "GameLost")
	})
	t.Run("solo may always yield", func(t *testing.T) {
		g := table{hands: []string{"KC KD"}, enemy: "JS", yielded: []int{0}}.build(t)
		mustOK(t, g.Yield(pid(0)))
		mustOK(t, g.Discard(pid(0), cs("KC")))
		mustOK(t, g.Yield(pid(0)))
	})
	t.Run("errors", func(t *testing.T) {
		expectErr(t, lobbyWith(t, 2), ErrWrongPhase, func(g *Game) error { return g.Yield(pid(0)) })
		g := table{hands: []string{"KC", "6H"}, enemy: "JS"}.build(t)
		expectErr(t, g, ErrNotYourTurn, func(g *Game) error { return g.Yield(pid(1)) })
		expectErr(t, g, ErrUnknownPlayer, func(g *Game) error { return g.Yield("zz") })
		g.Step = StepDiscard
		expectErr(t, g, ErrWrongStep, func(g *Game) error { return g.Yield(pid(0)) })
	})
}

func TestCannotActLosses(t *testing.T) {
	t.Run("next player empty and may not yield", func(t *testing.T) {
		g := table{hands: []string{"KC", ""}, enemy: "JS", step: StepDiscard, yielded: []int{0}}.build(t)
		mustOK(t, g.Discard(pid(0), cs("KC")))
		evs := wantEventTypes(t, g, "CardsDiscarded", "TurnStarted", "GameLost")
		if evs[2] != (GameLost{Reason: LossCannotAct, PlayerID: pid(1)}) {
			t.Fatalf("loss %#v", evs[2])
		}
	})
	t.Run("next player empty but may yield", func(t *testing.T) {
		g := table{hands: []string{"KC", ""}, enemy: "JS", step: StepDiscard}.build(t)
		mustOK(t, g.Discard(pid(0), cs("KC")))
		if g.Phase != PhasePlaying || g.Turn.ActiveSeat != 1 {
			t.Fatalf("phase %s", g.Phase)
		}
	})
	t.Run("same player after a defeat with an empty hand", func(t *testing.T) {
		g := table{hands: []string{"10H", "6H"}, enemy: "JS", castle: "JC", played: []played{{1, "10D"}}, yielded: []int{1}}.build(t)
		mustOK(t, g.Play(pid(0), cs("10H"), noShuffle{}))
		evs := wantEventTypes(t, g, "CardsPlayed", "EnemyDefeated", "EnemyRevealed", "TurnStarted", "GameLost")
		if evs[4] != (GameLost{Reason: LossCannotAct, PlayerID: pid(0)}) {
			t.Fatalf("loss %#v", evs[4])
		}
	})
}

func TestCommandsAfterEnd(t *testing.T) {
	for _, phase := range []Phase{PhaseWon, PhaseLost} {
		g := table{hands: []string{"5C 5H"}, enemy: "JS"}.build(t)
		g.Phase, g.Step, g.Result = phase, StepNone, &Result{Reason: LossGaveUp, PlayerID: pid(0)}
		ops := map[string]func(g *Game) error{
			"play":    func(g *Game) error { return g.Play(pid(0), cs("5C"), noShuffle{}) },
			"yield":   func(g *Game) error { return g.Yield(pid(0)) },
			"discard": func(g *Game) error { return g.Discard(pid(0), cs("5C")) },
			"choose":  func(g *Game) error { return g.ChooseNext(pid(0), pid(0)) },
			"signal":  func(g *Game) error { return g.SetJesterSignal(pid(0), SignalNone) },
			"jester":  func(g *Game) error { return g.UseSoloJester(pid(0)) },
			"giveUp":  func(g *Game) error { return g.GiveUp(pid(0)) },
			"start":   func(g *Game) error { return g.Start(pid(0), "", noShuffle{}) },
			"leave":   func(g *Game) error { return g.Leave(pid(0)) },
		}
		for name, op := range ops {
			t.Run(string(phase)+"/"+name, func(t *testing.T) {
				expectErr(t, g, ErrWrongPhase, op)
			})
		}
	}
}

func TestSoloJester(t *testing.T) {
	tavern := "AH 2H 3H 4H 5H 6H 7H 8H 9H 10H"
	t.Run("in Step 1", func(t *testing.T) {
		g := table{hands: []string{"2C 3C"}, enemy: "JD", tavern: tavern, played: []played{{0, "4D"}}}.build(t)
		mustOK(t, g.UseSoloJester(pid(0)))
		evs := g.PullEvents()
		if !reflect.DeepEqual(evs, []Event{SoloJesterUsed{PlayerID: pid(0), Drawn: 8, Left: 1}}) {
			t.Fatalf("events %#v", evs)
		}
		if !slices.Equal(g.Players[0].Hand, sorted("3H 4H 5H 6H 7H 8H 9H 10H")) || !slices.Equal(g.DiscardPile, cs("2C 3C")) ||
			!slices.Equal(g.Tavern, cs("AH 2H")) || g.SoloJestersUsed != 1 || g.Step != StepPlay || g.Revision != 2 {
			t.Fatalf("after: %+v", g)
		}
		// Does not cancel immunity, does not touch the played cards.
		if g.ImmuneSuit() != SuitDiamonds || len(g.Played) != 1 || g.JesterPlayed() {
			t.Fatal("solo jester changed the fight")
		}
		mustOK(t, g.UseSoloJester(pid(0)))
		if g.SoloJestersLeft() != 0 || len(g.Players[0].Hand) != 2 || len(g.Tavern) != 0 {
			t.Fatalf("second use with a short Tavern: hand %v", g.Players[0].Hand)
		}
		expectErr(t, g, ErrNoSoloJester, func(g *Game) error { return g.UseSoloJester(pid(0)) })
	})
	t.Run("Step 4 loss is deferred while a Jester is left", func(t *testing.T) {
		g := table{hands: []string{"2C 3C 4C"}, enemy: "JS", tavern: "2D 3D", soloUsed: 1}.build(t)
		mustOK(t, g.Play(pid(0), cs("2C"), noShuffle{}))
		wantEventTypes(t, g, "CardsPlayed")
		if g.Step != StepDiscard || g.Phase != PhasePlaying {
			t.Fatalf("step %s phase %s", g.Step, g.Phase)
		}
		expectErr(t, g, ErrIllegalDiscard, func(g *Game) error { return g.Discard(pid(0), cs("3C 4C")) })
		// The last Jester does not help enough: the loss triggers now.
		mustOK(t, g.UseSoloJester(pid(0)))
		evs := wantEventTypes(t, g, "SoloJesterUsed", "GameLost")
		if evs[1] != (GameLost{Reason: LossCannotCover, PlayerID: pid(0)}) {
			t.Fatalf("loss %#v", evs[1])
		}
	})
	t.Run("Step 4 rescue", func(t *testing.T) {
		g := table{hands: []string{"2C 3C 4C"}, enemy: "JS", tavern: "10D KD"}.build(t)
		mustOK(t, g.Play(pid(0), cs("2C"), noShuffle{}))
		mustOK(t, g.UseSoloJester(pid(0)))
		if g.Phase != PhasePlaying || g.Step != StepDiscard || !slices.Equal(g.Players[0].Hand, sorted("10D KD")) {
			t.Fatalf("after: %+v", g)
		}
		mustOK(t, g.Discard(pid(0), cs("10D")))
		if g.Step != StepPlay || g.Turn.Number != 2 {
			t.Fatalf("step %s", g.Step)
		}
	})
	t.Run("usable right after a defeat", func(t *testing.T) {
		g := table{hands: []string{"10H"}, enemy: "JS", castle: "JC", tavern: tavern, played: []played{{0, "10D"}}}.build(t)
		mustOK(t, g.Play(pid(0), cs("10H"), noShuffle{}))
		if g.Step != StepPlay || g.Phase != PhasePlaying || len(g.Players[0].Hand) != 0 {
			t.Fatalf("after defeat: %+v", g)
		}
		mustOK(t, g.UseSoloJester(pid(0)))
	})
	t.Run("empty hand without Jesters loses at Step 1", func(t *testing.T) {
		g := table{hands: []string{"10H"}, enemy: "JS", castle: "JC", played: []played{{0, "10D"}}, soloUsed: 2}.build(t)
		mustOK(t, g.Play(pid(0), cs("10H"), noShuffle{}))
		evs := wantEventTypes(t, g, "CardsPlayed", "EnemyDefeated", "EnemyRevealed", "TurnStarted", "GameLost")
		if evs[4] != (GameLost{Reason: LossCannotAct, PlayerID: pid(0)}) {
			t.Fatalf("loss %#v", evs[4])
		}
	})
	t.Run("Jester into an empty Tavern with an empty hand loses at Step 1", func(t *testing.T) {
		g := table{hands: []string{""}, enemy: "JS", soloUsed: 1}.build(t)
		mustOK(t, g.UseSoloJester(pid(0)))
		wantEventTypes(t, g, "SoloJesterUsed", "GameLost")
	})
	t.Run("errors", func(t *testing.T) {
		expectErr(t, table{hands: []string{"5C", "6C"}, enemy: "JS"}.build(t), ErrNotSolo, func(g *Game) error { return g.UseSoloJester(pid(0)) })
		expectErr(t, table{hands: []string{"5C", "6C"}, enemy: "JS"}.build(t), ErrNotSolo, func(g *Game) error { return g.UseSoloJester(pid(1)) })
		expectErr(t, lobbyWith(t, 1), ErrWrongPhase, func(g *Game) error { return g.UseSoloJester(pid(0)) })
		expectErr(t, table{hands: []string{"5C"}, enemy: "JS"}.build(t), ErrUnknownPlayer, func(g *Game) error { return g.UseSoloJester("zz") })
		expectErr(t, table{hands: []string{"5C"}, enemy: "JS", step: StepChooseNext}.build(t), ErrWrongStep, func(g *Game) error { return g.UseSoloJester(pid(0)) })
	})
	t.Run("grade on a solo win", func(t *testing.T) {
		g := table{hands: []string{"2H"}, enemy: "KS", played: []played{{0, "10H"}, {0, "10D"}, {0, "9C"}}, soloUsed: 1}.build(t)
		mustOK(t, g.Play(pid(0), cs("2H"), noShuffle{}))
		evs := g.PullEvents()
		if evs[len(evs)-1] != (GameWon{Grade: GradeSilver}) || g.Grade() != GradeSilver {
			t.Fatalf("events %#v", evs)
		}
	})
}

func TestLastTurn(t *testing.T) {
	t.Run("play then discard", func(t *testing.T) {
		g := table{hands: []string{"7H 5C KC", "6H"}, enemy: "JC"}.build(t)
		mustOK(t, g.Play(pid(0), cs("7H"), noShuffle{}))
		want := LastTurn{By: pid(0), Played: cs("7H")}
		if !reflect.DeepEqual(*g.LastTurn, want) {
			t.Fatalf("after play %+v", *g.LastTurn)
		}
		mustOK(t, g.Discard(pid(0), cs("KC")))
		want.Discarded = cs("KC")
		if !reflect.DeepEqual(*g.LastTurn, want) {
			t.Fatalf("after discard %+v", *g.LastTurn)
		}
		// The next play replaces it, dropping the old discard.
		mustOK(t, g.Play(pid(1), cs("6H"), noShuffle{}))
		if !reflect.DeepEqual(*g.LastTurn, LastTurn{By: pid(1), Played: cs("6H")}) {
			t.Fatalf("after next play %+v", *g.LastTurn)
		}
	})
	t.Run("an exact defeat records the enemy", func(t *testing.T) {
		g := table{hands: []string{"5C KH", "6H"}, enemy: "JS", castle: "JC", played: []played{{1, "10H"}}}.build(t)
		mustOK(t, g.Play(pid(0), cs("5C"), noShuffle{}))
		want := LastTurn{By: pid(0), Played: cs("5C"), Defeated: new(c("JS")), Exact: true}
		if !reflect.DeepEqual(*g.LastTurn, want) {
			t.Fatalf("after defeat %+v", *g.LastTurn)
		}
	})
	t.Run("an overkill is not exact", func(t *testing.T) {
		g := table{hands: []string{"6C KH", "6H"}, enemy: "JS", castle: "JC", played: []played{{1, "10H"}}}.build(t)
		mustOK(t, g.Play(pid(0), cs("6C"), noShuffle{}))
		want := LastTurn{By: pid(0), Played: cs("6C"), Defeated: new(c("JS"))}
		if !reflect.DeepEqual(*g.LastTurn, want) {
			t.Fatalf("after defeat %+v", *g.LastTurn)
		}
	})
	t.Run("yield has no played cards", func(t *testing.T) {
		g := table{hands: []string{"KC", "6H"}, enemy: "JS"}.build(t)
		mustOK(t, g.Yield(pid(0)))
		mustOK(t, g.Discard(pid(0), cs("KC")))
		if !reflect.DeepEqual(*g.LastTurn, LastTurn{By: pid(0), Discarded: cs("KC")}) {
			t.Fatalf("after yield %+v", *g.LastTurn)
		}
	})
	t.Run("a blocked attack has no discard", func(t *testing.T) {
		g := table{hands: []string{"2C 3H", "6H"}, enemy: "JH", played: []played{{1, "10S"}}}.build(t)
		mustOK(t, g.Play(pid(0), cs("2C"), noShuffle{}))
		if !reflect.DeepEqual(*g.LastTurn, LastTurn{By: pid(0), Played: cs("2C")}) {
			t.Fatalf("after blocked %+v", *g.LastTurn)
		}
	})
	t.Run("a Jester play", func(t *testing.T) {
		g := table{hands: []string{"X 2C", "6H"}, enemy: "JH"}.build(t)
		mustOK(t, g.Play(pid(0), cs("X"), noShuffle{}))
		if !reflect.DeepEqual(*g.LastTurn, LastTurn{By: pid(0), Played: cs("X")}) {
			t.Fatalf("after Jester %+v", *g.LastTurn)
		}
	})
	t.Run("an illegal play keeps it", func(t *testing.T) {
		g := table{hands: []string{"7H 5C KC", "6H"}, enemy: "JC"}.build(t)
		mustOK(t, g.Play(pid(0), cs("7H"), noShuffle{}))
		mustOK(t, g.Discard(pid(0), cs("KC")))
		before := *g.LastTurn
		if err := g.Play(pid(1), cs("6H 6H"), noShuffle{}); err == nil {
			t.Fatal("illegal play accepted")
		}
		if !reflect.DeepEqual(*g.LastTurn, before) {
			t.Fatalf("changed to %+v", *g.LastTurn)
		}
	})
	t.Run("clone and view copy it", func(t *testing.T) {
		g := table{hands: []string{"5C KH", "6H"}, enemy: "JS", castle: "JC", played: []played{{1, "10H"}}}.build(t)
		mustOK(t, g.Play(pid(0), cs("5C"), noShuffle{}))
		cl := g.Clone()
		v, err := g.ViewFor(pid(1))
		mustOK(t, err)
		g.LastTurn.Played[0] = c("9D")
		*g.LastTurn.Defeated = c("KD")
		for _, lt := range []*LastTurn{cl.LastTurn, v.LastTurn} {
			if lt.Played[0] != c("5C") || *lt.Defeated != c("JS") {
				t.Fatalf("aliased: %+v", *lt)
			}
		}
	})
}
