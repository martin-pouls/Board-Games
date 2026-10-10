package domain

import (
	"math/rand/v2"
	"slices"
	"testing"
)

// subsets returns every non-empty subset of hand with at most maxLen cards.
func subsets(hand []Card, maxLen int) [][]Card {
	var out [][]Card
	n := len(hand)
	for mask := 1; mask < 1<<n; mask++ {
		var s []Card
		for i := range n {
			if mask&(1<<i) != 0 {
				s = append(s, hand[i])
			}
		}
		if len(s) <= maxLen {
			out = append(out, s)
		}
	}
	return out
}

// legalPlays enumerates every legal Step 1 play of hand (by brute force over subsets ≤ 4).
func legalPlays(hand []Card) [][]Card {
	var out [][]Card
	for _, s := range subsets(hand, 4) {
		if _, err := ClassifyPlay(s); err == nil {
			out = append(out, s)
		}
	}
	return out
}

// legalDiscards enumerates every legal discard set of hand against attack.
func legalDiscards(hand []Card, attack int) [][]Card {
	var out [][]Card
	for _, s := range subsets(hand, len(hand)) {
		if ValidateDiscard(s, attack) == nil {
			out = append(out, s)
		}
	}
	return out
}

// randomAction performs one random legal action and returns a short description.
func randomAction(t *testing.T, g *Game, rng *rand.Rand, greedy bool) string {
	t.Helper()
	seat := g.Turn.ActiveSeat
	p := g.Players[seat]
	soloJester := g.SoloJestersLeft() > 0 && rng.IntN(8) == 0
	switch g.Step {
	case StepPlay:
		if soloJester {
			mustOK(t, g.UseSoloJester(p.ID))
			return "solo jester (play)"
		}
		plays := legalPlays(p.Hand)
		if g.CanYield(seat) && (len(plays) == 0 || (!greedy && rng.IntN(6) == 0)) {
			mustOK(t, g.Yield(p.ID))
			return "yield"
		}
		if len(plays) == 0 {
			t.Fatalf("no action at Step 1, but no loss: %+v", g)
		}
		pick := plays[rng.IntN(len(plays))]
		if greedy {
			pick = bestPlay(g, plays)
		}
		mustOK(t, g.Play(p.ID, pick, rng))
		return "play"
	case StepDiscard:
		ds := legalDiscards(p.Hand, g.EffectiveAttack())
		if soloJester || len(ds) == 0 {
			if g.SoloJestersLeft() == 0 {
				t.Fatalf("cannot cover, no Jester, but no loss: %+v", g)
			}
			mustOK(t, g.UseSoloJester(p.ID))
			return "solo jester (discard)"
		}
		pick := ds[rng.IntN(len(ds))]
		if greedy {
			pick = slices.MinFunc(ds, func(a, b []Card) int { return HandValue(a) - HandValue(b) })
		}
		mustOK(t, g.Discard(p.ID, pick))
		return "discard"
	case StepChooseNext:
		for i, o := range g.Players {
			if i != seat && rng.IntN(2) == 0 {
				signals := []JesterSignal{SignalNone, SignalGoodPlay, SignalRatherNot}
				mustOK(t, g.SetJesterSignal(o.ID, signals[rng.IntN(3)]))
			}
		}
		next := g.Players[rng.IntN(len(g.Players))].ID
		mustOK(t, g.ChooseNext(p.ID, next))
		return "choose next"
	}
	t.Fatalf("unexpected step %q", g.Step)
	return ""
}

// checkViews checks that every player's view shows exactly their own hand.
func checkViews(t *testing.T, g *Game) {
	t.Helper()
	over := g.Phase == PhaseWon || g.Phase == PhaseLost
	for _, p := range g.Players {
		v, err := g.ViewFor(p.ID)
		mustOK(t, err)
		if !slices.Equal(v.Hand, p.Hand) || v.Revision != g.Revision {
			t.Fatalf("%s: view hand %v, hand %v", p.ID, v.Hand, p.Hand)
		}
		for _, o := range v.Players {
			if (o.RevealedHand != nil) != over {
				t.Fatalf("%s sees %s's hand before the end", p.ID, o.ID)
			}
		}
	}
}

func TestRandomGames(t *testing.T) {
	const seeds, maxActions = 50, 2000
	for n := MinPlayers; n <= MaxPlayers; n++ {
		won, lost := 0, 0
		for seed := range uint64(seeds) {
			rng := rand.New(rand.NewPCG(seed, uint64(n)))
			g := lobbyWith(t, n)
			mustOK(t, g.Start(pid(0), "", rng))
			checkInvariants(t, g)
			actions := 0
			for g.Phase == PhasePlaying {
				if actions++; actions > maxActions {
					t.Fatalf("%dp seed %d: no end after %d actions", n, seed, maxActions)
				}
				rev := g.Revision
				what := randomAction(t, g, rng, seed%2 == 1)
				if g.Revision <= rev {
					t.Fatalf("%dp seed %d: %s did not commit", n, seed, what)
				}
				checkInvariants(t, g)
				checkViews(t, g)
				g.PullEvents()
			}
			checkViews(t, g)
			if g.Phase == PhaseWon {
				won++
				if (g.Grade() != GradeNone) != g.IsSolo() {
					t.Fatalf("%dp seed %d: grade %q", n, seed, g.Grade())
				}
			} else {
				lost++
			}
		}
		t.Logf("%d players: %d won, %d lost", n, won, lost)
	}
}

// bestPlay picks the play with the best outcome one step ahead: win, then progress against the
// Castle, then damage dealt, then the cards kept; plays that lose come last.
func bestPlay(g *Game, plays [][]Card) []Card {
	score := func(cards []Card) int {
		h := g.Clone()
		if err := h.Play(h.Players[h.Turn.ActiveSeat].ID, cards, noShuffle{}); err != nil {
			panic(err)
		}
		switch h.Phase {
		case PhaseWon:
			return 1 << 30
		case PhaseLost:
			return -1 << 30
		}
		hand := 0
		for _, p := range h.Players {
			hand += HandValue(p.Hand)
		}
		return h.EnemiesDefeated()*10000 + h.Damage()*10 + h.Shield()*5 + hand - h.EffectiveAttack()*2
	}
	best, bestScore := plays[0], score(plays[0])
	for _, p := range plays[1:] {
		if s := score(p); s > bestScore {
			best, bestScore = p, s
		}
	}
	return best
}
