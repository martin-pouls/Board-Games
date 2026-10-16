package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/martin-pouls/shared/backend/sqlitedb"

	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
	"github.com/martin-pouls/fox-duet/backend/internal/core/ports"
)

// GameRepository is the SQLite implementation of ports.GameRepository.
type GameRepository struct {
	repo *sqlitedb.Repository[*domain.Game, domain.PlayerID]
}

var _ ports.GameRepository = (*GameRepository)(nil)

// NewGameRepository returns a repository backed by db.
func NewGameRepository(db *DB) *GameRepository {
	return &GameRepository{repo: sqlitedb.NewRepository[*domain.Game, domain.PlayerID](db, codec,
		sqlitedb.Errors{NotFound: ports.ErrGameNotFound, Conflict: ports.ErrVersionConflict, Duplicate: ports.ErrDuplicateGameID})}
}

// statusOf is the informational status column: the game's phase.
func statusOf(g *domain.Game) string { return string(g.Phase) }

// codec stores a game as its JSON snapshot (gameDTO).
var codec = sqlitedb.Codec[*domain.Game]{
	ID:     func(g *domain.Game) string { return string(g.ID) },
	Status: statusOf,
	Encode: func(g *domain.Game) ([]byte, error) { return json.Marshal(toDTO(g)) },
	Decode: func(state []byte, createdAt time.Time) (*domain.Game, error) {
		var d gameDTO
		if err := json.Unmarshal(state, &d); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
		return fromDTO(d)
	},
}

// Create implements ports.GameRepository.
func (r *GameRepository) Create(ctx context.Context, rec *ports.GameRecord) error {
	v, err := r.repo.Create(ctx, rec.Game, rec.TokenHashes)
	if err != nil {
		return err
	}
	rec.Version = v
	return nil
}

// Load implements ports.GameRepository.
func (r *GameRepository) Load(ctx context.Context, id domain.GameID) (*ports.GameRecord, error) {
	g, hashes, version, err := r.repo.Load(ctx, string(id))
	if err != nil {
		return nil, err
	}
	return &ports.GameRecord{Game: g, TokenHashes: hashes, Version: version}, nil
}

// Update implements ports.GameRepository.
func (r *GameRepository) Update(ctx context.Context, rec *ports.GameRecord) error {
	v, err := r.repo.Update(ctx, rec.Game, rec.TokenHashes, rec.Version)
	if err != nil {
		return err
	}
	rec.Version = v
	return nil
}

// DeleteInactive implements ports.GameRepository.
func (r *GameRepository) DeleteInactive(ctx context.Context, olderThan time.Time) (int, error) {
	return r.repo.DeleteInactive(ctx, olderThan)
}
