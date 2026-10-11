package service

import (
	"github.com/martin-pouls/shared/backend/idgen"

	"github.com/martin-pouls/regicide/backend/internal/core/domain"
)

func defaultGameID() domain.GameID     { return domain.GameID(idgen.GameCode()) }
func defaultPlayerID() domain.PlayerID { return domain.PlayerID(idgen.PlayerID()) }
func defaultToken() string             { return idgen.Token() }
func defaultRNG() domain.RNG           { return idgen.NewRNG() }

// hashToken returns hex(sha256(token)); only hashes are ever stored.
func hashToken(token string) string { return idgen.HashToken(token) }

// normalizeID canonicalises a user-supplied join code.
func normalizeID(id domain.GameID) domain.GameID {
	return domain.GameID(idgen.NormalizeCode(string(id)))
}
