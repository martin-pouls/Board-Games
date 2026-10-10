package domain

// NewTavern returns the unshuffled Tavern: for s in Suits: ranks A, 2..10 of s; then `jesters` Jesters.
// The last element is the top.
func NewTavern(jesters int) []Card {
	out := make([]Card, 0, 40+max(jesters, 0))
	for _, s := range Suits {
		for r := RankAce; r <= 10; r++ {
			out = append(out, NewCard(s, r))
		}
	}
	for range jesters {
		out = append(out, Jester())
	}
	return out
}

// NewCastle returns the Castle with the last element on top: Kings (bottom), then Queens, then
// Jacks (top). Each group is built in Suits order and shuffled separately, in this order:
// Jacks, Queens, Kings.
func NewCastle(rng RNG) []Card {
	group := func(r Rank) []Card {
		g := make([]Card, 0, len(Suits))
		for _, s := range Suits {
			g = append(g, NewCard(s, r))
		}
		rng.Shuffle(len(g), func(i, j int) { g[i], g[j] = g[j], g[i] })
		return g
	}
	jacks := group(RankJack)
	queens := group(RankQueen)
	kings := group(RankKing)
	out := make([]Card, 0, CastleSize)
	out = append(out, kings...)
	out = append(out, queens...)
	return append(out, jacks...)
}
