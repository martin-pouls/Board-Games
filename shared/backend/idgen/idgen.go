// Package idgen makes the identifiers and secrets of a game session: join codes, player IDs,
// bearer tokens (only their hashes are stored) and seeded RNGs. Games convert the strings to
// their own domain types.
package idgen

import (
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"strings"
)

// CodeAlphabet has 32 symbols (no 0/O/1/I), so a masked random byte is uniform.
// The frontend mirror is CODE_ALPHABET in shared/frontend.
const (
	CodeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	CodeLength   = 5
)

func mustRandom(b []byte) {
	if _, err := crand.Read(b); err != nil {
		panic(fmt.Sprintf("idgen: crypto/rand failed: %v", err))
	}
}

// GameCode returns a random join code of CodeLength symbols from CodeAlphabet.
func GameCode() string {
	var raw [CodeLength]byte
	mustRandom(raw[:])
	var sb strings.Builder
	for _, b := range raw {
		sb.WriteByte(CodeAlphabet[b&31])
	}
	return sb.String()
}

// PlayerID returns a random player ID like "p_abcdefghij".
func PlayerID() string {
	var raw [7]byte
	mustRandom(raw[:])
	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw[:])
	return "p_" + strings.ToLower(s[:10])
}

// Token returns a random 256-bit bearer token, base64url without padding.
func Token() string {
	var raw [32]byte
	mustRandom(raw[:])
	return base64.RawURLEncoding.EncodeToString(raw[:])
}

// HashToken returns hex(sha256(token)); only hashes are ever stored.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// NewRNG returns a ChaCha8 generator seeded from crypto/rand.
func NewRNG() *rand.Rand {
	var seed [32]byte
	mustRandom(seed[:])
	return rand.New(rand.NewChaCha8(seed))
}

// NormalizeCode canonicalises a user-supplied join code (trimmed, upper case).
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// GameCodeAs is GameCode as the game's own ID type.
func GameCodeAs[T ~string]() T { return T(GameCode()) }

// PlayerIDAs is PlayerID as the game's own ID type.
func PlayerIDAs[T ~string]() T { return T(PlayerID()) }

// NormalizeCodeAs is NormalizeCode on the game's own ID type.
func NormalizeCodeAs[T ~string](code T) T { return T(NormalizeCode(string(code))) }
