package domain

import "math/rand/v2"

// *rand.Rand is the production RNG (see shared/backend/idgen); it must satisfy RNG.
var _ RNG = (*rand.Rand)(nil)
