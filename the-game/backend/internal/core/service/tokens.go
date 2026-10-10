package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func defaultToken() string {
	var raw [32]byte
	mustRandom(raw[:])
	return base64.RawURLEncoding.EncodeToString(raw[:])
}

// hashToken returns hex(sha256(token)); only hashes are ever stored.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
