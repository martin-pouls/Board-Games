package domain

import (
	"math/rand/v2"
	"slices"
	"testing"
)

// greedyScore rates the state after a follow card or move: losing is worst, a gem is best.
func greedyScore(g *Game) int {
	switch {
	case g.Phase == PhaseLost:
		return -100
	case g.Phase == PhaseWon:
		return 100
	case g.LastTrick != nil && g.LastTrick.Beyond != EndNone:
		return -10
	case g.LastTrick != nil && g.LastTrick.Collected:
		return 1
	}
	return 0
}

// bestFollowScore returns the best score the follower can reach after the lead (0 when an
// ability decision comes first).
func bestFollowScore(g *Game) int {
	if g.Step != StepFollow {
		return 0
	}
	id := g.Players[1-g.Trick.Leader].ID
	best := -1000
	for _, card := range g.LegalCards(1 - g.Trick.Leader) {
		cl := g.Clone()
		if cl.PlayCard(id, card, noShuffle{}) == nil {
			best = max(best, greedyScore(cl))
		}
	}
	return best
}

// greedyPick returns the index of the candidate whose result scores best by score (ties: the
// first).
func greedyPick(g *Game, n int, score func(*Game) int, try func(*Game, int) error) int {
	best, bestScore := 0, -1000
	for i := range n {
		cl := g.Clone()
		if try(cl, i) != nil {
			continue
		}
		if s := score(cl); s > bestScore {
			best, bestScore = i, s
		}
	}
	return best
}

// randomAction performs one random legal action. A greedy player picks the follow card and the
// move that avoid getting lost and collect a gem when they can.
func randomAction(t *testing.T, g *Game, rng *rand.Rand, greedy bool) {
	t.Helper()
	actor := g.Actors()[rng.IntN(len(g.Actors()))]
	id := g.Players[actor].ID
	hand := g.Players[actor].Hand
	var err error
	switch g.Step {
	case StepLead, StepFollow:
		legal := g.LegalCards(actor)
		pick := rng.IntN(len(legal))
		if greedy {
			score := greedyScore
			if g.Step == StepLead {
				score = bestFollowScore
			}
			pick = greedyPick(g, len(legal), score, func(cl *Game, i int) error { return cl.PlayCard(id, legal[i], noShuffle{}) })
		}
		err = g.PlayCard(id, legal[pick], rng)
	case StepFoxesChoose:
		err = g.ChooseFoxesPlayer(id, g.Players[rng.IntN(Players)].ID, rng)
	case StepFoxesSwap:
		if rng.IntN(2) == 0 {
			err = g.KeepDecree(id, rng)
		} else {
			err = g.SwapDecree(id, hand[rng.IntN(len(hand))], rng)
		}
	case StepGift:
		err = g.GiveGift(id, hand[rng.IntN(len(hand))], rng)
	case StepMove:
		opts := g.MoveOptions()
		pick := rng.IntN(len(opts))
		if greedy {
			pick = greedyPick(g, len(opts), greedyScore, func(cl *Game, i int) error { return cl.ChooseMove(id, opts[i].Delta, noShuffle{}) })
		}
		err = g.ChooseMove(id, opts[pick].Delta, rng)
	case StepForestEnd:
		ends := g.CoverableEnds()
		err = g.ChooseForestEnd(id, ends[rng.IntN(len(ends))], rng)
	default:
		t.Fatalf("no action for step %q", g.Step)
	}
	if err != nil {
		t.Fatalf("step %s by %s: %v", g.Step, id, err)
	}
}

// checkInvariants verifies what must hold after every action of a running game.
func checkInvariants(t *testing.T, g *Game) {
	t.Helper()
	l := g.LevelDef()
	b := g.Board
	if b.ForestTokens < 0 || b.CoveredLeft+b.CoveredRight+b.ForestTokens != l.ForestTokens {
		t.Fatalf("forest tokens %+v", b)
	}
	if b.Covered(b.Traveller) || b.Covered(l.Start()) {
		t.Fatalf("traveller or start covered %+v", b)
	}
	for i, n := range b.Gems {
		if n < 0 || (n > 0 && b.Covered(i)) {
			t.Fatalf("gems %v on covered %+v", b.Gems, b)
		}
	}
	if g.Phase != PhasePlaying {
		return
	}
	if g.Round < 1 || g.Round > Rounds || len(g.Deck) != 7 || g.Decree == nil || b.GemsLeft() == 0 {
		t.Fatalf("round %d deck %d decree %v gems %d", g.Round, len(g.Deck), g.Decree, b.GemsLeft())
	}
	seen := map[Card]bool{*g.Decree: true}
	for _, card := range g.Deck {
		seen[card] = true
	}
	for _, p := range g.Players {
		for _, card := range p.Hand {
			if seen[card] {
				t.Fatalf("duplicate card %v", card)
			}
			seen[card] = true
		}
	}
	if g.Step != StepForestEnd {
		inPlay := 2 * (TricksPerRound - g.Trick.Number + 1)
		if got := totalCards(g) - 8; got != inPlay {
			t.Fatalf("trick %d: %d cards in play, want %d", g.Trick.Number, got, inPlay)
		}
	}
	if len(g.Actors()) == 0 {
		t.Fatalf("nobody can act in step %s", g.Step)
	}
}

func TestRandomGames(t *testing.T) {
	outcomes := map[string]int{}
	for seed := range 600 {
		rng := rand.New(rand.NewPCG(uint64(seed), 99))
		g := newLobbyGame(t)
		level := MinLevel + seed%MaxLevel
		must(t, g.Start(alice, level, rng))
		for steps := 0; g.Phase == PhasePlaying; steps++ {
			if steps > 500 {
				t.Fatalf("seed %d: game does not end", seed)
			}
			randomAction(t, g, rng, seed%2 == 0)
			checkInvariants(t, g)
		}
		switch g.Phase {
		case PhaseWon:
			outcomes["won"]++
		case PhaseLost:
			outcomes[string(g.Result.Reason)]++
		default:
			t.Fatalf("seed %d: phase %s", seed, g.Phase)
		}
	}
	for _, want := range []string{"won", string(LossLostInForest), string(LossOutOfTime)} {
		if outcomes[want] == 0 {
			t.Errorf("no game ended with %s: %v", want, outcomes)
		}
	}
	t.Logf("outcomes: %v", outcomes)
}

func TestCloneIsDeep(t *testing.T) {
	g := newPlaying(t, 1)
	setup(g, "7D 2R 4S", "5D 8S 10R", "2S", 0, 1)
	play(t, g, alice, "7D")
	must(t, g.GiveGift(alice, c("2R"), noShuffle{}))
	cl := g.Clone()
	g.Players[0].Hand[0] = c("1S")
	g.Board.Gems[0] = 9
	g.Trick.Lead.Rank = 1
	g.Trick.Gift[0].Rank = 1
	g.Decree.Rank = 1
	if cl.Players[0].Hand[0] == c("1S") || cl.Board.Gems[0] == 9 || cl.Trick.Lead.Rank == 1 || cl.Trick.Gift[0].Rank == 1 || cl.Decree.Rank == 1 {
		t.Fatal("clone shares state")
	}
	if !slices.Equal(cl.PullEvents(), g.PullEvents()) {
		t.Fatal("clone lost events")
	}
}
