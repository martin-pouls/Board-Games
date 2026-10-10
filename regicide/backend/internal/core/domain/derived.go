package domain

import "slices"

// PlayEffect is the derived effect of one play against the current enemy (rules §4 Step 2, §6).
type PlayEffect struct {
	Value  int    // Σ card values (0 for a Jester)
	Jester bool   // the play is the Jester
	Suits  []Suit // distinct suits in the play, Suits order (nil for a Jester)
	// BlockedSuits are the suits whose power is not in effect for this play: ♥ ♦ ♣ if the enemy was
	// immune at the time of the play; ♠ if the enemy is immune now (spades are retroactive, rules §6).
	BlockedSuits []Suit
	Doubled      bool // ♣ in the play and clubs not immune at the time of the play
	Damage       int  // Value ×2 if Doubled; 0 for a Jester
}

// isJesterPlay reports whether a played set is the Jester.
func isJesterPlay(ps PlayedSet) bool {
	return len(ps.Cards) == 1 && ps.Cards[0].IsJester()
}

// suitsOf returns the distinct suits of cards in Suits order (nil if none).
func suitsOf(cards []Card) []Suit {
	var out []Suit
	for _, s := range Suits {
		if slices.ContainsFunc(cards, func(c Card) bool { return c.Suit == s }) {
			out = append(out, s)
		}
	}
	return out
}

// PlayEffectAt derives the effect of Played[i]. The enemy was immune at the time of play i iff its
// suit is in the play and no earlier play against it was a Jester.
func (g *Game) PlayEffectAt(i int) PlayEffect {
	ps := g.Played[i]
	if isJesterPlay(ps) {
		return PlayEffect{Jester: true}
	}
	e := PlayEffect{Value: HandValue(ps.Cards), Suits: suitsOf(ps.Cards)}
	var enemySuit Suit
	if g.Enemy != nil {
		enemySuit = g.Enemy.Suit
	}
	jesterBefore := slices.ContainsFunc(g.Played[:i], isJesterPlay)
	for _, s := range e.Suits {
		if s != enemySuit {
			continue
		}
		if s == SuitSpades && g.ImmuneSuit() != SuitSpades {
			continue // retroactively unblocked by a later Jester
		}
		if s != SuitSpades && jesterBefore {
			continue
		}
		e.BlockedSuits = append(e.BlockedSuits, s)
	}
	e.Doubled = slices.Contains(e.Suits, SuitClubs) && !slices.Contains(e.BlockedSuits, SuitClubs)
	e.Damage = e.Value
	if e.Doubled {
		e.Damage *= 2
	}
	return e
}

// JesterPlayed reports whether a Jester was played against the current enemy.
func (g *Game) JesterPlayed() bool {
	return slices.ContainsFunc(g.Played, isJesterPlay)
}

// ImmuneSuit is the suit the current enemy is immune to right now: its own suit, unless a Jester
// was played against it. SuitNone without an enemy.
func (g *Game) ImmuneSuit() Suit {
	if g.Enemy == nil || g.JesterPlayed() {
		return SuitNone
	}
	return g.Enemy.Suit
}

// Shield is the total ♠ shield against the current enemy: Σ value of every play containing ♠,
// counted only if spades are not immune now (rules §6 App decision).
func (g *Game) Shield() int {
	if g.Enemy == nil || g.ImmuneSuit() == SuitSpades {
		return 0
	}
	sum := 0
	for _, ps := range g.Played {
		if slices.ContainsFunc(ps.Cards, func(c Card) bool { return c.Suit == SuitSpades }) {
			sum += HandValue(ps.Cards)
		}
	}
	return sum
}

// Damage is the damage accumulated against the current enemy.
func (g *Game) Damage() int {
	sum := 0
	for i := range g.Played {
		sum += g.PlayEffectAt(i).Damage
	}
	return sum
}

// EffectiveAttack is the enemy's attack minus the shield, never below 0; 0 without an enemy.
func (g *Game) EffectiveAttack() int {
	if g.Enemy == nil {
		return 0
	}
	return max(0, EnemyAttack(g.Enemy.Rank)-g.Shield())
}

// CanYield reports whether the player in seat may yield (rules §8): always in solo; otherwise iff
// some other player did not yield on their last turn.
func (g *Game) CanYield(seat int) bool {
	if g.IsSolo() {
		return true
	}
	for i, p := range g.Players {
		if i != seat && !p.YieldedLastTurn {
			return true
		}
	}
	return false
}

// Grade is the solo victory grade (rules §11): 0/1/2 Jesters used → gold/silver/bronze.
// GradeNone unless the game is a won solo game.
func (g *Game) Grade() Grade {
	if g.Phase != PhaseWon || !g.IsSolo() {
		return GradeNone
	}
	switch g.SoloJestersUsed {
	case 0:
		return GradeGold
	case 1:
		return GradeSilver
	default:
		return GradeBronze
	}
}

// EnemiesDefeated is the number of enemies defeated so far (0 before the game started).
func (g *Game) EnemiesDefeated() int {
	if g.Phase == PhaseLobby || g.Phase == PhaseEnded {
		return 0
	}
	n := CastleSize - len(g.Castle)
	if g.Enemy != nil {
		n--
	}
	return n
}
