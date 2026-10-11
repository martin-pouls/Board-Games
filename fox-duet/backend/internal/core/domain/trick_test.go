package domain

import (
	"slices"
	"testing"
)

// Level 1 (side A) unless stated: 11 spaces, start 5, gems [1 1 2 1 1 0 1 1 2 1 1].
// Alice (seat 0) faces the left end, so moves toward her are negative.

func TestFollowSuit(t *testing.T) {
	tests := []struct {
		name     string
		lead     string
		hand     string
		card     string
		legal    string
		wantFail bool
	}{
		{"must follow the lead suit", "4D", "2D 6R", "6R", "2D", true},
		{"following is fine", "4D", "2D 6R", "2D", "2D", false},
		{"void: any card", "4D", "6R 8S", "8S", "6R 8S", false},
		{"Royal Heir led: any card", "9D", "2D 6R", "6R", "2D 6R", false},
		{"Royal Heir led: following still allowed", "9D", "2D 6R", "2D", "2D 6R", false},
		{"Royal Heir played second has no effect", "4R", "9D 6R", "9D", "6R", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newPlaying(t, 1)
			setup(g, tt.lead+" 10S", tt.hand, "2S", 0, 1)
			play(t, g, alice, tt.lead)
			if got := g.LegalCards(1); !slices.Equal(got, cs(tt.legal)) {
				t.Fatalf("legal %v, want %v", got, cs(tt.legal))
			}
			if tt.wantFail {
				mustFail(t, g, ErrMustFollowSuit, func(g *Game) error { return g.PlayCard(bob, c(tt.card), noShuffle{}) })
			} else {
				must(t, g.PlayCard(bob, c(tt.card), noShuffle{}))
			}
		})
	}
}

func TestPlayCardErrors(t *testing.T) {
	g := newLobbyGame(t)
	mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.PlayCard(alice, c("2D"), noShuffle{}) })

	g = newPlaying(t, 1)
	setup(g, "4D 10S", "2D 6R", "2S", 0, 1)
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.PlayCard(bob, c("2D"), noShuffle{}) })
	mustFail(t, g, ErrCardNotInHand, func(g *Game) error { return g.PlayCard(alice, c("2D"), noShuffle{}) })
	mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.PlayCard("carol", c("4D"), noShuffle{}) })
	mustFail(t, g, ErrWrongStep, func(g *Game) error { return g.ChooseMove(alice, 0, noShuffle{}) })
	mustFail(t, g, ErrWrongStep, func(g *Game) error { return g.GiveGift(alice, c("4D"), noShuffle{}) })
	mustFail(t, g, ErrWrongStep, func(g *Game) error { return g.ChooseForestEnd(alice, EndLeft, noShuffle{}) })
	play(t, g, alice, "4D")
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.PlayCard(alice, c("10S"), noShuffle{}) })
}

func TestLeadWins(t *testing.T) {
	tests := []struct {
		lead, follow string
		trump        Suit
		want         bool
	}{
		{"4D", "6D", SuitStars, false},  // same suit: higher wins
		{"6D", "4D", SuitStars, true},   // same suit: higher wins
		{"2D", "10R", SuitStars, true},  // off-suit, not trump: the lead wins
		{"2D", "1S", SuitStars, false},  // follow trumps
		{"2S", "10D", SuitStars, true},  // lead is trump
		{"2S", "10S", SuitStars, false}, // both trump: higher wins
		{"10D", "1D", SuitDoves, true},  // both trump (the lead suit is trump)
		{"1R", "2D", SuitDoves, false},  // follow trumps a lead of another suit
	}
	for _, tt := range tests {
		if got := leadWins(c(tt.lead), c(tt.follow), tt.trump); got != tt.want {
			t.Errorf("%s vs %s (trump %s) = %v, want %v", tt.lead, tt.follow, tt.trump, got, tt.want)
		}
	}
}

func TestTrickMovesTowardTheWinnerAndCollectsAGem(t *testing.T) {
	tests := []struct {
		name         string
		lead, follow string
		winner       int
		to           int
	}{
		{"leader alice wins: toward the left", "6D", "4D", 0, 2},
		{"follower bob wins: toward the right", "4D", "6D", 1, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newPlaying(t, 1)
			setup(g, tt.lead+" 2R", tt.follow+" 8S", "2S", 0, 1)
			play(t, g, alice, tt.lead)
			play(t, g, bob, tt.follow)
			r := g.LastTrick
			if r == nil || r.Winner != tt.winner || r.From != 5 || r.To != tt.to || !r.Collected || r.Beyond != EndNone {
				t.Fatalf("last trick %+v", r)
			}
			if g.Board.Traveller != tt.to || g.Board.Gems[tt.to] != 1 || g.Board.GemsLeft() != 11 {
				t.Fatalf("board %+v", g.Board)
			}
			if g.Step != StepLead || g.Trick.Number != 2 || g.Trick.Leader != tt.winner || g.Trick.Lead != nil {
				t.Fatalf("step %s trick %+v", g.Step, g.Trick)
			}
			if len(eventsOf[TrickWon](g)) != 1 || len(eventsOf[TravellerMoved](g)) != 1 || len(eventsOf[GemCollected](g)) != 1 {
				t.Fatalf("events %v", g.events)
			}
		})
	}
}

func TestGemCollectedWithoutMoving(t *testing.T) {
	// 9 and 1 move 0; a Musician with no movement leaves one result, so no choice is asked.
	g := newPlaying(t, 1)
	setup(g, "9D 2R", "1D 8S", "2S", 0, 1)
	g.Board.Traveller = 2
	play(t, g, alice, "9D")
	play(t, g, bob, "1D")
	if g.Step != StepLead || g.Board.Traveller != 2 || g.Board.Gems[2] != 1 || !g.LastTrick.Collected {
		t.Fatalf("step %s board %+v", g.Step, g.Board)
	}
}

func TestNoGemNoCollect(t *testing.T) {
	g := newPlaying(t, 1) // the start space of side A has no gems
	setup(g, "9D 2R", "1D 8S", "2S", 0, 1)
	play(t, g, alice, "9D")
	play(t, g, bob, "1D")
	if g.LastTrick.Collected || g.Board.GemsLeft() != 12 || len(eventsOf[GemCollected](g)) != 0 {
		t.Fatalf("collected from an empty location: %+v", g.LastTrick)
	}
}

func TestMoveOptions(t *testing.T) {
	tests := []struct {
		name         string
		traveller    int
		hand0, hand1 string
		lead, follow string
		winner       int
		want         []MoveOption
	}{
		{
			name: "plain: one option", traveller: 5, hand0: "6D", hand1: "4D", lead: "6D", follow: "4D", winner: 0,
			want: []MoveOption{{Delta: -3, To: 2, Gem: true}},
		},
		{
			name: "Musician: either direction", traveller: 5, hand0: "1D", hand1: "6D", lead: "1D", follow: "6D", winner: 1,
			want: []MoveOption{{Delta: 2, To: 7, Gem: true}, {Delta: -2, To: 3, Gem: true}},
		},
		{
			name: "Gazelle: ignore one card", traveller: 2, hand0: "5D", hand1: "10D", lead: "5D", follow: "10D", winner: 1,
			want: []MoveOption{{Delta: 4, To: 6, Gem: true}, {Delta: 3, To: 5}, {Delta: 1, To: 3, Gem: true}},
		},
		{
			name: "two Gazelles: ignore one, none or both", traveller: 5, hand0: "5D", hand1: "5R", lead: "5D", follow: "5R", winner: 0,
			want: []MoveOption{{Delta: -2, To: 3, Gem: true}, {Delta: -1, To: 4, Gem: true}, {Delta: 0, To: 5}},
		},
		{
			name: "Musician and Gazelle combine", traveller: 5, hand0: "1D", hand1: "5D", lead: "1D", follow: "5D", winner: 1,
			want: []MoveOption{{Delta: 1, To: 6, Gem: true}, {Delta: 0, To: 5}, {Delta: -1, To: 4, Gem: true}},
		},
		{
			name: "results beyond the same end are one option", traveller: 9, hand0: "5D", hand1: "10D", lead: "5D", follow: "10D", winner: 1,
			want: []MoveOption{{Delta: 4, To: 5, Beyond: EndRight}, {Delta: 1, To: 10, Gem: true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newPlaying(t, 1)
			setup(g, tt.hand0+" 2R", tt.hand1+" 8S", "2S", 0, 1)
			g.Board.Traveller = tt.traveller
			play(t, g, alice, tt.lead)
			play(t, g, bob, tt.follow)
			if len(tt.want) == 1 {
				if g.Step != StepLead || g.LastTrick.Delta != tt.want[0].Delta {
					t.Fatalf("step %s last trick %+v", g.Step, g.LastTrick)
				}
				return
			}
			if g.Step != StepMove || !slices.Equal(g.Actors(), []int{tt.winner}) {
				t.Fatalf("step %s actors %v", g.Step, g.Actors())
			}
			if got := g.MoveOptions(); !slices.Equal(got, tt.want) {
				t.Fatalf("options %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestChooseMove(t *testing.T) {
	g := newPlaying(t, 1)
	setup(g, "1D 2R", "6D 8S", "2S", 0, 1)
	play(t, g, alice, "1D")
	play(t, g, bob, "6D")
	mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.ChooseMove(alice, -2, noShuffle{}) })
	mustFail(t, g, ErrIllegalMove, func(g *Game) error { return g.ChooseMove(bob, 3, noShuffle{}) })
	mustFail(t, g, ErrWrongStep, func(g *Game) error { return g.PlayCard(bob, c("8S"), noShuffle{}) })
	must(t, g.ChooseMove(bob, -2, noShuffle{}))
	if g.Board.Traveller != 3 || g.LastTrick.Delta != -2 || g.Step != StepLead || g.Trick.Leader != 1 {
		t.Fatalf("traveller %d trick %+v step %s", g.Board.Traveller, g.LastTrick, g.Step)
	}

	// A move with the same result as an offered option is accepted too.
	g = newPlaying(t, 1)
	setup(g, "5D 2R", "5R 8S", "2S", 0, 1)
	g.Players[1].Hand = cs("5R 8S")
	play(t, g, alice, "5D")
	play(t, g, bob, "5R")
	must(t, g.ChooseMove(alice, -1, noShuffle{}))
	if g.Board.Traveller != 4 {
		t.Fatalf("traveller %d", g.Board.Traveller)
	}
}

func TestBeyondThePath(t *testing.T) {
	t.Run("left end: reset, cover, gems slide, no gem collected", func(t *testing.T) {
		g := newPlaying(t, 2) // side B: start 4 with 2 gems, gems [1 2 2 0 2 2 0 2 1 1]
		setup(g, "6D 2R", "4D 8S", "2S", 0, 1)
		g.Board.Traveller = 1
		play(t, g, alice, "6D")
		play(t, g, bob, "4D")
		b := g.Board
		if b.Traveller != 4 || b.CoveredLeft != 1 || b.ForestTokens != 2 || !slices.Equal(b.Gems, []int{0, 3, 2, 0, 2, 2, 0, 2, 1, 1}) {
			t.Fatalf("board %+v", b)
		}
		r := g.LastTrick
		if r.Beyond != EndLeft || r.To != 4 || r.Collected || r.From != 1 {
			t.Fatalf("last trick %+v", r)
		}
		placed := eventsOf[ForestTokenPlaced](g)
		if len(placed) != 1 || placed[0] != (ForestTokenPlaced{End: EndLeft, Space: 0, GemsMoved: 1, Left: 2}) {
			t.Fatalf("events %v", g.events)
		}
		if g.Step != StepLead || g.Phase != PhasePlaying {
			t.Fatalf("step %s phase %s", g.Step, g.Phase)
		}
	})
	t.Run("right end after an earlier cover", func(t *testing.T) {
		g := newPlaying(t, 2)
		setup(g, "4D 2R", "6D 8S", "2S", 0, 1)
		g.Board.CoveredRight = 1
		g.Board.Gems[9] = 0
		g.Board.ForestTokens = 2
		g.Board.Traveller = 7
		play(t, g, alice, "4D")
		play(t, g, bob, "6D") // +3 from 7: past space 8, the last uncovered on the right
		b := g.Board
		if b.Traveller != 4 || b.CoveredRight != 2 || b.ForestTokens != 1 || b.Gems[8] != 0 || b.Gems[7] != 3 {
			t.Fatalf("board %+v", b)
		}
	})
	t.Run("no forest token left: lost in the forest", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "6D 2R", "4D 8S", "2S", 0, 1)
		g.Board.Traveller = 1
		g.Board.ForestTokens = 0
		play(t, g, alice, "6D")
		play(t, g, bob, "4D")
		if g.Phase != PhaseLost || g.Result.Reason != LossLostInForest || g.Step != StepNone {
			t.Fatalf("phase %s result %+v", g.Phase, g.Result)
		}
		if g.Board.Traveller != 1 || g.LastTrick.Beyond != EndLeft || g.LastTrick.To != 1 {
			t.Fatalf("traveller %d last trick %+v", g.Board.Traveller, g.LastTrick)
		}
		mustFail(t, g, ErrWrongPhase, func(g *Game) error { return g.PlayCard(bob, c("8S"), noShuffle{}) })
	})
}

func TestVictory(t *testing.T) {
	g := newPlaying(t, 1)
	setup(g, "6D 2R", "4D 8S", "2S", 0, 5)
	g.Round = 2
	g.Board.Gems = []int{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}
	play(t, g, alice, "6D")
	play(t, g, bob, "4D")
	if g.Phase != PhaseWon || g.Step != StepNone || g.Result == nil || g.Result.Reason != "" {
		t.Fatalf("phase %s result %+v", g.Phase, g.Result)
	}
	// 10 (level 1) + 2 cards + 10 (won in round 2) + 4 tokens * 3.
	want := Score{Level: 10, Cards: 2, EarlyWin: 10, ForestTokens: 12, Total: 34}
	if got := g.Score(); got != want {
		t.Fatalf("score %+v, want %+v", got, want)
	}
	if won := eventsOf[GameWon](g); len(won) != 1 || won[0].Score != 34 {
		t.Fatalf("events %v", g.events)
	}
}

func TestScore(t *testing.T) {
	tests := []struct {
		level, round, tokens, cards int
		want                        Score
	}{
		{1, 3, 0, 0, Score{Level: 10, Total: 10}},
		{2, 2, 3, 4, Score{Level: 20, Cards: 4, EarlyWin: 10, ForestTokens: 9, Total: 43}},
		{3, 3, 1, 6, Score{Level: 30, Cards: 6, ForestTokens: 3, Total: 39}},
	}
	for _, tt := range tests {
		g := newPlaying(t, tt.level)
		g.Round = tt.round
		g.Board.ForestTokens = tt.tokens
		g.Players[0].Hand = NewDeck()[:tt.cards/2]
		g.Players[1].Hand = NewDeck()[10 : 10+tt.cards-tt.cards/2]
		if got := g.Score(); got != tt.want {
			t.Errorf("level %d round %d: %+v, want %+v", tt.level, tt.round, got, tt.want)
		}
	}
}

func TestLastTrickStaysUntilTheNextLead(t *testing.T) {
	g := newPlaying(t, 1)
	setup(g, "6D 2R", "4D 8S", "2S", 0, 1)
	play(t, g, alice, "6D")
	play(t, g, bob, "4D")
	if g.LastTrick == nil {
		t.Fatal("no last trick")
	}
	v, _ := g.ViewFor(bob)
	if v.LastTrick == nil || v.LastTrick.Lead != c("6D") || v.LastTrick.WinnerID != alice || v.Trick.Lead != nil {
		t.Fatalf("view %+v / %+v", v.LastTrick, v.Trick)
	}
	play(t, g, alice, "2R")
	if g.LastTrick != nil {
		t.Fatalf("last trick still shown: %+v", g.LastTrick)
	}
}

func TestFoxes(t *testing.T) {
	t.Run("led: the chosen player swaps, then the follower plays", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "3D 5S", "2D 8S", "2R", 0, 1)
		play(t, g, alice, "3D")
		if g.Step != StepFoxesChoose || !slices.Equal(g.Actors(), []int{0}) {
			t.Fatalf("step %s actors %v", g.Step, g.Actors())
		}
		mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.ChooseFoxesPlayer(bob, bob, noShuffle{}) })
		mustFail(t, g, ErrUnknownPlayer, func(g *Game) error { return g.ChooseFoxesPlayer(alice, "carol", noShuffle{}) })
		must(t, g.ChooseFoxesPlayer(alice, bob, noShuffle{}))
		if g.Step != StepFoxesSwap || !slices.Equal(g.Actors(), []int{1}) {
			t.Fatalf("step %s actors %v", g.Step, g.Actors())
		}
		mustFail(t, g, ErrNotYourTurn, func(g *Game) error { return g.KeepDecree(alice, noShuffle{}) })
		mustFail(t, g, ErrCardNotInHand, func(g *Game) error { return g.SwapDecree(bob, c("5S"), noShuffle{}) })
		must(t, g.SwapDecree(bob, c("8S"), noShuffle{}))
		if *g.Decree != c("8S") || !slices.Equal(g.Players[1].Hand, cs("2D 2R")) || g.Step != StepFollow {
			t.Fatalf("decree %v hand %v step %s", *g.Decree, g.Players[1].Hand, g.Step)
		}
		if ev := eventsOf[DecreeSwapped](g); len(ev) != 1 || ev[0].Old != c("2R") || ev[0].New != c("8S") {
			t.Fatalf("events %v", g.events)
		}
		// The lead suit stays doves.
		if !slices.Equal(g.LegalCards(1), cs("2D")) {
			t.Fatalf("legal %v", g.LegalCards(1))
		}
	})
	t.Run("played second: a new trump changes the winner", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "4D 10S", "3R 6R 8S", "2S", 0, 1)
		play(t, g, alice, "4D")
		play(t, g, bob, "3R") // void in doves
		if g.Step != StepFoxesChoose || !slices.Equal(g.Actors(), []int{1}) {
			t.Fatalf("step %s actors %v", g.Step, g.Actors())
		}
		must(t, g.ChooseFoxesPlayer(bob, bob, noShuffle{}))
		must(t, g.SwapDecree(bob, c("6R"), noShuffle{}))
		if g.LastTrick == nil || g.LastTrick.Winner != 1 || g.LastTrick.Delta != 3 {
			t.Fatalf("last trick %+v", g.LastTrick)
		}
	})
	t.Run("keeping the decree", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "4D 10S", "3R 6R 8S", "2S", 0, 1)
		play(t, g, alice, "4D")
		play(t, g, bob, "3R")
		must(t, g.ChooseFoxesPlayer(bob, alice, noShuffle{}))
		must(t, g.KeepDecree(alice, noShuffle{}))
		if *g.Decree != c("2S") || g.LastTrick.Winner != 0 || g.LastTrick.Delta != -3 || len(eventsOf[DecreeKept](g)) != 1 {
			t.Fatalf("decree %v last trick %+v", *g.Decree, g.LastTrick)
		}
	})
	t.Run("chosen player with an empty hand: nothing happens", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "3D", "2D", "2R", 0, 11)
		play(t, g, alice, "3D")
		if g.Step != StepFoxesChoose {
			t.Fatalf("step %s", g.Step)
		}
		must(t, g.ChooseFoxesPlayer(alice, alice, noShuffle{}))
		if g.Step != StepFollow || !eventsOf[FoxesChosen](g)[0].Skipped {
			t.Fatalf("step %s events %v", g.Step, g.events)
		}
	})
	t.Run("both hands empty: skipped", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "4D", "3D", "2R", 0, 11)
		play(t, g, alice, "4D")
		play(t, g, bob, "3D")
		if g.Step == StepFoxesChoose || g.LastTrick == nil || len(eventsOf[FoxesChosen](g)) != 0 {
			t.Fatalf("step %s last trick %+v", g.Step, g.LastTrick)
		}
	})
}

func TestGift(t *testing.T) {
	t.Run("led: both choose secretly, then the cards change hands", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "7D 2R 4S", "5D 8S 10R", "2S", 0, 1)
		play(t, g, alice, "7D")
		if g.Step != StepGift || !slices.Equal(g.Actors(), []int{0, 1}) {
			t.Fatalf("step %s actors %v", g.Step, g.Actors())
		}
		mustFail(t, g, ErrCardNotInHand, func(g *Game) error { return g.GiveGift(alice, c("7D"), noShuffle{}) })
		mustFail(t, g, ErrWrongStep, func(g *Game) error { return g.PlayCard(bob, c("5D"), noShuffle{}) })
		must(t, g.GiveGift(alice, c("2R"), noShuffle{}))
		mustFail(t, g, ErrGiftAlreadyChosen, func(g *Game) error { return g.GiveGift(alice, c("4S"), noShuffle{}) })
		if !slices.Equal(g.Actors(), []int{1}) {
			t.Fatalf("actors %v", g.Actors())
		}
		va, _ := g.ViewFor(alice)
		vb, _ := g.ViewFor(bob)
		if va.Pending.YourGift == nil || *va.Pending.YourGift != c("2R") || vb.Pending.YourGift != nil {
			t.Fatalf("gift views %+v / %+v", va.Pending, vb.Pending)
		}
		must(t, g.GiveGift(bob, c("8S"), noShuffle{}))
		if !slices.Equal(g.Players[0].Hand, cs("4S 8S")) || !slices.Equal(g.Players[1].Hand, cs("2R 5D 10R")) {
			t.Fatalf("hands %v / %v", g.Players[0].Hand, g.Players[1].Hand)
		}
		if g.Step != StepFollow || len(eventsOf[GiftExchanged](g)) != 1 {
			t.Fatalf("step %s events %v", g.Step, g.events)
		}
		va, _ = g.ViewFor(alice)
		if va.LastGift == nil || va.LastGift.Gave != c("2R") || va.LastGift.Received != c("8S") {
			t.Fatalf("last gift %+v", va.LastGift)
		}
		// Following uses the hand after the exchange.
		mustFail(t, g, ErrMustFollowSuit, func(g *Game) error { return g.PlayCard(bob, c("2R"), noShuffle{}) })
	})
	t.Run("a single card is given automatically", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "7D 2R", "5D 8S", "2S", 0, 10)
		play(t, g, alice, "7D")
		if g.Step != StepGift || !slices.Equal(g.Actors(), []int{1}) || !eventsOf[GiftChosen](g)[0].Auto {
			t.Fatalf("step %s actors %v events %v", g.Step, g.Actors(), g.events)
		}
		must(t, g.GiveGift(bob, c("8S"), noShuffle{}))
		if !slices.Equal(g.Players[0].Hand, cs("8S")) || !slices.Equal(g.Players[1].Hand, cs("2R 5D")) {
			t.Fatalf("hands %v / %v", g.Players[0].Hand, g.Players[1].Hand)
		}
	})
	t.Run("both single: exchanged automatically", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "2R 4S", "7R 9S", "2D", 0, 10)
		play(t, g, alice, "2R")
		play(t, g, bob, "7R")
		if !slices.Equal(g.Players[0].Hand, cs("9S")) || !slices.Equal(g.Players[1].Hand, cs("4S")) {
			t.Fatalf("hands %v / %v", g.Players[0].Hand, g.Players[1].Hand)
		}
		if g.Step != StepLead || g.LastTrick.Winner != 1 || g.LastTrick.Delta != 3 {
			t.Fatalf("step %s last trick %+v", g.Step, g.LastTrick)
		}
	})
	t.Run("a hand is empty: no exchange", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "7D", "5D", "2S", 0, 11)
		play(t, g, alice, "7D")
		if g.Step != StepFollow || len(eventsOf[GiftChosen](g)) != 0 {
			t.Fatalf("step %s events %v", g.Step, g.events)
		}
	})
}

func TestRoundEnd(t *testing.T) {
	// The last trick of round 1: bob wins 4D/6D and moves the traveller +3 from from.
	lastTrick := func(t *testing.T, from int) *Game {
		t.Helper()
		g := newPlaying(t, 1)
		setup(g, "4D", "6D", "2S", 0, 11)
		g.Board.Traveller = from
		play(t, g, alice, "4D")
		play(t, g, bob, "6D")
		return g
	}
	t.Run("the host chooses the end", func(t *testing.T) {
		g := lastTrick(t, 5) // lands on 8 and collects
		if g.Step != StepForestEnd || !slices.Equal(g.Actors(), []int{0}) || !slices.Equal(g.CoverableEnds(), []End{EndLeft, EndRight}) {
			t.Fatalf("step %s actors %v ends %v", g.Step, g.Actors(), g.CoverableEnds())
		}
		if added := eventsOf[GemsAdded](g); len(added) != 1 || added[0].Count != 5 {
			t.Fatalf("events %v", g.events)
		}
		if !slices.Equal(g.Board.Gems, []int{1, 2, 2, 2, 1, 1, 1, 2, 1, 2, 1}) {
			t.Fatalf("gems %v", g.Board.Gems)
		}
		mustFail(t, g, ErrNotHost, func(g *Game) error { return g.ChooseForestEnd(bob, EndLeft, noShuffle{}) })
		mustFail(t, g, ErrIllegalEnd, func(g *Game) error { return g.ChooseForestEnd(alice, "up", noShuffle{}) })
		must(t, g.ChooseForestEnd(alice, EndLeft, noShuffle{}))
		if g.Board.CoveredLeft != 1 || g.Board.Gems[0] != 0 || g.Board.Gems[1] != 3 || g.Board.ForestTokens != 3 {
			t.Fatalf("board %+v", g.Board)
		}
		if g.Round != 2 || g.Dealer != 1 || g.Trick.Leader != 0 || g.Trick.Number != 1 || g.Step != StepLead || totalCards(g) != 30 {
			t.Fatalf("round %d dealer %d trick %+v step %s cards %d", g.Round, g.Dealer, g.Trick, g.Step, totalCards(g))
		}
		if g.Board.Traveller != 8 || g.LastTrick == nil {
			t.Fatalf("traveller %d last trick %+v", g.Board.Traveller, g.LastTrick)
		}
	})
	t.Run("the traveller on an end forces the other end", func(t *testing.T) {
		g := lastTrick(t, 7) // lands on 10, the right end
		placed := eventsOf[ForestTokenPlaced](g)
		if len(placed) != 1 || placed[0].End != EndLeft || !placed[0].RoundEnd {
			t.Fatalf("events %v", g.events)
		}
		if g.Round != 2 || g.Step != StepLead || g.Board.CoveredLeft != 1 {
			t.Fatalf("round %d step %s board %+v", g.Round, g.Step, g.Board)
		}
	})
	t.Run("no forest token left: nothing is placed", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "4D", "6D", "2S", 0, 11)
		g.Board.ForestTokens = 0
		play(t, g, alice, "4D")
		play(t, g, bob, "6D")
		if g.Round != 2 || g.Step != StepLead || len(eventsOf[ForestTokenPlaced](g)) != 0 {
			t.Fatalf("round %d step %s events %v", g.Round, g.Step, g.events)
		}
	})
	t.Run("after round 3: out of time", func(t *testing.T) {
		g := newPlaying(t, 1)
		setup(g, "4D", "6D", "2S", 0, 11)
		g.Round = 3
		play(t, g, alice, "4D")
		play(t, g, bob, "6D")
		if g.Phase != PhaseLost || g.Result.Reason != LossOutOfTime || len(eventsOf[GemsAdded](g)) != 0 {
			t.Fatalf("phase %s result %+v", g.Phase, g.Result)
		}
	})
	t.Run("the deal alternates", func(t *testing.T) {
		g := lastTrick(t, 5)
		must(t, g.ChooseForestEnd(alice, EndRight, noShuffle{}))
		rounds := eventsOf[RoundStarted](g)
		if len(rounds) != 1 || rounds[0].Round != 2 || rounds[0].DealerID != bob {
			t.Fatalf("events %v", g.events)
		}
	})
}

func TestRevisionOncePerMutator(t *testing.T) {
	g := newPlaying(t, 1)
	setup(g, "4D", "6D", "2S", 0, 11)
	r := g.Revision
	play(t, g, alice, "4D")
	play(t, g, bob, "6D") // resolves the trick and ends the round in the same call
	if g.Revision != r+2 {
		t.Fatalf("revision %d, want %d", g.Revision, r+2)
	}
}
