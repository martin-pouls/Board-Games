package sqlitedb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Store errors. A game's repository maps them to its port errors.
var (
	ErrNotFound  = errors.New("sqlitedb: game not found")
	ErrConflict  = errors.New("sqlitedb: version conflict")
	ErrDuplicate = errors.New("sqlitedb: duplicate game id")
)

// Record is one stored game: its serialised snapshot plus persistence metadata.
type Record struct {
	ID      string
	Status  string            // informational column (e.g. the phase); never read back by the app
	State   []byte            // JSON snapshot of the game aggregate
	Tokens  map[string]string // player ID -> hex(sha256(token)); never the raw token
	Version int64             // optimistic-concurrency version as loaded
}

// Store persists Records in the games and player_tokens tables.
type Store struct {
	db *sql.DB
}

// NewStore returns a store backed by db.
func NewStore(db *DB) *Store {
	return &Store{db: db.sql}
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func insertTokens(ctx context.Context, tx *sql.Tx, id string, tokens map[string]string) error {
	for pid, h := range tokens {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO player_tokens (game_id, player_id, token_hash) VALUES (?, ?, ?)`,
			id, pid, h); err != nil {
			return fmt.Errorf("insert token: %w", err)
		}
	}
	return nil
}

// Create inserts rec with version 1 and sets rec.Version to 1. ErrDuplicate if the ID exists.
func (s *Store) Create(ctx context.Context, rec *Record) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO games (id, status, state, version) VALUES (?, ?, ?, 1)`,
		rec.ID, rec.Status, string(rec.State)); err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("insert game: %w", err)
	}
	if err := insertTokens(ctx, tx, rec.ID, rec.Tokens); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	rec.Version = 1
	return nil
}

// Load returns the stored game and its token hashes. ErrNotFound.
func (s *Store) Load(ctx context.Context, id string) (*Record, error) {
	var status, state string
	var version int64
	err := s.db.QueryRowContext(ctx, `SELECT status, state, version FROM games WHERE id = ?`, id).Scan(&status, &state, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load game: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `SELECT player_id, token_hash FROM player_tokens WHERE game_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("load tokens: %w", err)
	}
	defer rows.Close()
	tokens := map[string]string{}
	for rows.Next() {
		var pid, h string
		if err := rows.Scan(&pid, &h); err != nil {
			return nil, fmt.Errorf("scan token: %w", err)
		}
		tokens[pid] = h
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load tokens: %w", err)
	}
	return &Record{ID: id, Status: status, State: []byte(state), Tokens: tokens, Version: version}, nil
}

// Update saves rec if the stored version still equals rec.Version, then increments
// rec.Version. Token hashes not stored yet are inserted. ErrConflict / ErrNotFound.
func (s *Store) Update(ctx context.Context, rec *Record) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE games SET state = ?, status = ?, version = version + 1, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND version = ?`,
		string(rec.State), rec.Status, rec.ID, rec.Version)
	if err != nil {
		return fmt.Errorf("update game: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		var one int
		err := tx.QueryRowContext(ctx, `SELECT 1 FROM games WHERE id = ?`, rec.ID).Scan(&one)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("check game: %w", err)
		}
		return ErrConflict
	}
	if err := insertTokens(ctx, tx, rec.ID, rec.Tokens); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	rec.Version++
	return nil
}

// DeleteInactive deletes games last updated before olderThan (their tokens cascade) and
// returns how many were deleted.
func (s *Store) DeleteInactive(ctx context.Context, olderThan time.Time) (int, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM games WHERE updated_at < ?`, olderThan.UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return 0, fmt.Errorf("delete inactive: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return int(n), nil
}
