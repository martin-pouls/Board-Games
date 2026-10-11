package domain

// RNG is the injected source of randomness. *math/rand/v2.Rand satisfies it.
type RNG interface {
	Shuffle(n int, swap func(i, j int))
	// IntN returns a random int in [0, n). It panics if n <= 0.
	IntN(n int) int
}
