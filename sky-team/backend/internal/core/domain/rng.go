package domain

// RNG is the injected source of randomness. *math/rand/v2.Rand satisfies it.
// A player die is 1 + IntN(6), the traffic die TrafficFaces()[IntN(6)], the intern row a
// Shuffle of its 6 tokens.
type RNG interface {
	Shuffle(n int, swap func(i, j int))
	// IntN returns a random int in [0, n). It panics if n <= 0.
	IntN(n int) int
}

// rollDie rolls one player die.
func rollDie(rng RNG) int {
	return 1 + rng.IntN(6)
}

// rollTraffic rolls the traffic die.
func rollTraffic(rng RNG) int {
	return trafficFaces[rng.IntN(len(trafficFaces))]
}
