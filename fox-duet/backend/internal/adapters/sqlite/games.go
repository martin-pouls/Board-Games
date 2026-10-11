package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/martin-pouls/shared/backend/sqlitedb"

	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
	"github.com/martin-pouls/fox-duet/backend/internal/core/ports"
)

// GameRepository is the SQLite implementation of ports.GameRepository.
type GameRepository struct {
	store *sqlitedb.Store
}

var _ ports.GameRepository = (*GameRepository)(nil)

// NewGameRepository returns a repository backed by db.
func NewGameRepository(db *DB) *GameRepository {
	return &GameRepository{store: sqlitedb.NewStore(db)}
}

func statusOf(g *domain.Game) string { return string(g.Phase) }

// portErr maps store errors to the port's sentinel errors.
func portErr(err error) error {
	switch {
	case errors.Is(err, sqlitedb.ErrNotFound):
		return ports.ErrGameNotFound
	case errors.Is(err, sqlitedb.ErrConflict):
		return ports.ErrVersionConflict
	case errors.Is(err, sqlitedb.ErrDuplicate):
		return ports.ErrDuplicateGameID
	default:
		return err
	}
}

func toRecord(rec *ports.GameRecord) (*sqlitedb.Record, error) {
	state, err := json.Marshal(toDTO(rec.Game))
	if err != nil {
		return nil, fmt.Errorf("marshal game: %w", err)
	}
	tokens := make(map[string]string, len(rec.TokenHashes))
	for pid, h := range rec.TokenHashes {
		tokens[string(pid)] = h
	}
	return &sqlitedb.Record{
		ID:      string(rec.Game.ID),
		Status:  statusOf(rec.Game),
		State:   state,
		Tokens:  tokens,
		Version: rec.Version,
	}, nil
}

// Create implements ports.GameRepository.
func (r *GameRepository) Create(ctx context.Context, rec *ports.GameRecord) error {
	row, err := toRecord(rec)
	if err != nil {
		return err
	}
	if err := r.store.Create(ctx, row); err != nil {
		return portErr(err)
	}
	rec.Version = row.Version
	return nil
}

// Load implements ports.GameRepository.
func (r *GameRepository) Load(ctx context.Context, id domain.GameID) (*ports.GameRecord, error) {
	row, err := r.store.Load(ctx, string(id))
	if err != nil {
		return nil, portErr(err)
	}
	var d gameDTO
	if err := json.Unmarshal(row.State, &d); err != nil {
		return nil, fmt.Errorf("unmarshal game %s: %w", id, err)
	}
	g, err := fromDTO(d)
	if err != nil {
		return nil, fmt.Errorf("decode game %s: %w", id, err)
	}
	hashes := make(map[domain.PlayerID]string, len(row.Tokens))
	for pid, h := range row.Tokens {
		hashes[domain.PlayerID(pid)] = h
	}
	return &ports.GameRecord{Game: g, TokenHashes: hashes, Version: row.Version}, nil
}

// Update implements ports.GameRepository.
func (r *GameRepository) Update(ctx context.Context, rec *ports.GameRecord) error {
	row, err := toRecord(rec)
	if err != nil {
		return err
	}
	if err := r.store.Update(ctx, row); err != nil {
		return portErr(err)
	}
	rec.Version = row.Version
	return nil
}

// DeleteInactive implements ports.GameRepository.
func (r *GameRepository) DeleteInactive(ctx context.Context, olderThan time.Time) (int, error) {
	return r.store.DeleteInactive(ctx, olderThan)
}
