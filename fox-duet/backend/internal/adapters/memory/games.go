// Package memory provides in-memory driven-port implementations for tests and local runs.
package memory

import (
	"context"
	"maps"
	"sync"
	"time"

	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
	"github.com/martin-pouls/fox-duet/backend/internal/core/ports"
)

type entry struct {
	game      *domain.Game
	hashes    map[domain.PlayerID]string
	version   int64
	updatedAt time.Time
}

// GameRepository is an in-memory ports.GameRepository.
type GameRepository struct {
	mu    sync.Mutex
	games map[domain.GameID]*entry
	now   func() time.Time
}

var _ ports.GameRepository = (*GameRepository)(nil)

// Option configures a GameRepository.
type Option func(*GameRepository)

// WithNow injects the clock used for updatedAt.
func WithNow(now func() time.Time) Option {
	return func(r *GameRepository) { r.now = now }
}

// NewGameRepository returns an empty repository.
func NewGameRepository(opts ...Option) *GameRepository {
	r := &GameRepository{games: map[domain.GameID]*entry{}, now: time.Now}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Create implements ports.GameRepository.
func (r *GameRepository) Create(_ context.Context, rec *ports.GameRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := rec.Game.ID
	if _, ok := r.games[id]; ok {
		return ports.ErrDuplicateGameID
	}
	hashes := maps.Clone(rec.TokenHashes)
	if hashes == nil {
		hashes = map[domain.PlayerID]string{}
	}
	r.games[id] = &entry{game: rec.Game.Clone(), hashes: hashes, version: 1, updatedAt: r.now()}
	rec.Version = 1
	return nil
}

// Load implements ports.GameRepository.
func (r *GameRepository) Load(_ context.Context, id domain.GameID) (*ports.GameRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.games[id]
	if !ok {
		return nil, ports.ErrGameNotFound
	}
	return &ports.GameRecord{Game: e.game.Clone(), TokenHashes: maps.Clone(e.hashes), Version: e.version}, nil
}

// Update implements ports.GameRepository. Already stored token hashes are never overwritten.
func (r *GameRepository) Update(_ context.Context, rec *ports.GameRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.games[rec.Game.ID]
	if !ok {
		return ports.ErrGameNotFound
	}
	if e.version != rec.Version {
		return ports.ErrVersionConflict
	}
	e.game = rec.Game.Clone()
	for pid, h := range rec.TokenHashes {
		if _, exists := e.hashes[pid]; !exists {
			e.hashes[pid] = h
		}
	}
	e.version++
	e.updatedAt = r.now()
	rec.Version++
	return nil
}

// DeleteInactive implements ports.GameRepository.
func (r *GameRepository) DeleteInactive(_ context.Context, olderThan time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for id, e := range r.games {
		if e.updatedAt.Before(olderThan) {
			delete(r.games, id)
			n++
		}
	}
	return n, nil
}
