package service

import (
	crand "crypto/rand"
	"encoding/base32"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/martin-pouls/the-game/backend/internal/core/domain"
)

// gameIDAlphabet has 32 symbols (no 0/O/1/I), so a masked random byte is uniform.
const (
	gameIDAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	gameIDLen      = 5
)

func mustRandom(b []byte) {
	if _, err := crand.Read(b); err != nil {
		panic(fmt.Sprintf("service: crypto/rand failed: %v", err))
	}
}

func defaultGameID() domain.GameID {
	var raw [gameIDLen]byte
	mustRandom(raw[:])
	var sb strings.Builder
	for _, b := range raw {
		sb.WriteByte(gameIDAlphabet[b&31])
	}
	return domain.GameID(sb.String())
}

func defaultPlayerID() domain.PlayerID {
	var raw [7]byte
	mustRandom(raw[:])
	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw[:])
	return domain.PlayerID("p_" + strings.ToLower(s[:10]))
}

func defaultRNG() domain.RNG {
	var seed [32]byte
	mustRandom(seed[:])
	return rand.New(rand.NewChaCha8(seed))
}

// normalizeID canonicalises a user-supplied join code.
func normalizeID(id domain.GameID) domain.GameID {
	return domain.GameID(strings.ToUpper(strings.TrimSpace(string(id))))
}
