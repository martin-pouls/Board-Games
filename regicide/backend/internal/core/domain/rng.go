package domain

// RNG is the injected source of randomness. *math/rand/v2.Rand satisfies it.
type RNG interface {
	Shuffle(n int, swap func(i, j int))
}
