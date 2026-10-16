package sqlitedb

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Codec turns a game into a stored snapshot and back.
type Codec[G any] struct {
	ID     func(G) string                                     // the game's ID
	Status func(G) string                                     // informational status column
	Encode func(G) ([]byte, error)                            // JSON snapshot
	Decode func(state []byte, createdAt time.Time) (G, error) // createdAt: the row's creation time
}

// Errors are the port sentinels a Repository returns instead of the store's.
type Errors struct{ NotFound, Conflict, Duplicate error }

// Repository stores games of type G, whose players are identified by P, as JSON snapshots in a
// Store. A game's repository adapter wraps it and converts its port types.
type Repository[G any, P ~string] struct {
	store *Store
	codec Codec[G]
	errs  Errors
}

// NewRepository returns a repository backed by db.
func NewRepository[G any, P ~string](db *DB, c Codec[G], e Errors) *Repository[G, P] {
	return &Repository[G, P]{store: NewStore(db), codec: c, errs: e}
}

// mapErr replaces the store's sentinel errors with the repository's.
func (r *Repository[G, P]) mapErr(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return r.errs.NotFound
	case errors.Is(err, ErrConflict):
		return r.errs.Conflict
	case errors.Is(err, ErrDuplicate):
		return r.errs.Duplicate
	default:
		return err
	}
}

func (r *Repository[G, P]) record(g G, tokens map[P]string, version int64) (*Record, error) {
	state, err := r.codec.Encode(g)
	if err != nil {
		return nil, fmt.Errorf("marshal game: %w", err)
	}
	hashes := make(map[string]string, len(tokens))
	for pid, h := range tokens {
		hashes[string(pid)] = h
	}
	return &Record{ID: r.codec.ID(g), Status: r.codec.Status(g), State: state, Tokens: hashes, Version: version}, nil
}

// Create stores a new game with its token hashes and returns its version (1).
// The Duplicate error if the ID exists.
func (r *Repository[G, P]) Create(ctx context.Context, g G, tokens map[P]string) (version int64, err error) {
	rec, err := r.record(g, tokens, 0)
	if err != nil {
		return 0, err
	}
	if err := r.store.Create(ctx, rec); err != nil {
		return 0, r.mapErr(err)
	}
	return rec.Version, nil
}

// Load returns the game, its token hashes and its version. The NotFound error if there is none.
func (r *Repository[G, P]) Load(ctx context.Context, id string) (g G, tokens map[P]string, version int64, err error) {
	var zero G
	rec, err := r.store.Load(ctx, id)
	if err != nil {
		return zero, nil, 0, r.mapErr(err)
	}
	g, err = r.codec.Decode(rec.State, rec.CreatedAt)
	if err != nil {
		return zero, nil, 0, fmt.Errorf("decode game %s: %w", id, err)
	}
	tokens = make(map[P]string, len(rec.Tokens))
	for pid, h := range rec.Tokens {
		tokens[P(pid)] = h
	}
	return g, tokens, rec.Version, nil
}

// Update saves the game if the stored version still equals version and returns the new one.
// Token hashes not stored yet are inserted. The Conflict or NotFound error otherwise.
func (r *Repository[G, P]) Update(ctx context.Context, g G, tokens map[P]string, version int64) (newVersion int64, err error) {
	rec, err := r.record(g, tokens, version)
	if err != nil {
		return 0, err
	}
	if err := r.store.Update(ctx, rec); err != nil {
		return 0, r.mapErr(err)
	}
	return rec.Version, nil
}

// DeleteInactive deletes games last updated before olderThan and returns how many.
func (r *Repository[G, P]) DeleteInactive(ctx context.Context, olderThan time.Time) (int, error) {
	return r.store.DeleteInactive(ctx, olderThan)
}
