package domain

// Player count limits.
const MinPlayers, MaxPlayers = 1, 4

// CastleSize is the number of enemies: 4 Jacks, 4 Queens, 4 Kings.
const CastleSize = 12

// Ruleset holds the values that depend on the player count (rules §3).
type Ruleset struct {
	MaxHandSize   int
	TavernJesters int // Jesters shuffled into the Tavern
	SoloJesters   int // Jester powers (solo only)
}

// rulesets is the ruleset table keyed by player count.
var rulesets = map[int]Ruleset{
	1: {MaxHandSize: 8, TavernJesters: 0, SoloJesters: 2},
	2: {MaxHandSize: 7, TavernJesters: 0, SoloJesters: 0},
	3: {MaxHandSize: 6, TavernJesters: 1, SoloJesters: 0},
	4: {MaxHandSize: 5, TavernJesters: 2, SoloJesters: 0},
}

// RulesFor returns the ruleset for the given player count; false outside MinPlayers..MaxPlayers.
func RulesFor(players int) (Ruleset, bool) {
	r, ok := rulesets[players]
	return r, ok
}
