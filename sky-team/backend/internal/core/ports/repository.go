package ports

import (
	"context"
	"time"

	"github.com/martin-pouls/sky-team/backend/internal/core/domain"
)

// GameRecord is a loaded game plus its persistence metadata.
type GameRecord struct {
	Game        *domain.Game
	TokenHashes map[domain.PlayerID]string // hex(sha256(token)); never the raw token
	Version     int64                      // OCC version as loaded
}

// GameRepository persists games (driven port).
type GameRepository interface {
	// Create inserts a new game with Version=1 (sets rec.Version=1). ErrDuplicateGameID if the id exists.
	Create(ctx context.Context, rec *GameRecord) error
	// Load returns a fresh, independent *domain.Game each call. ErrGameNotFound.
	Load(ctx context.Context, id domain.GameID) (*GameRecord, error)
	// Update persists rec if the stored version == rec.Version, then sets rec.Version++.
	// Inserts token hashes not stored yet. ErrVersionConflict / ErrGameNotFound.
	Update(ctx context.Context, rec *GameRecord) error
	// DeleteInactive deletes games whose last update is before olderThan (tokens cascade)
	// and returns how many were deleted.
	DeleteInactive(ctx context.Context, olderThan time.Time) (int, error)
}
