package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/martin-pouls/the-game/backend/internal/core/domain"
	"github.com/martin-pouls/the-game/backend/internal/core/ports"
)

// GameRepository is the SQLite implementation of ports.GameRepository.
type GameRepository struct {
	db *sql.DB
}

var _ ports.GameRepository = (*GameRepository)(nil)

// NewGameRepository returns a repository backed by db.
func NewGameRepository(db *DB) *GameRepository {
	return &GameRepository{db: db.sql}
}

func statusOf(g *domain.Game) string {
	switch g.Phase {
	case domain.PhaseLobby:
		return "lobby"
	case domain.PhaseFinished:
		if g.Result != nil {
			return string(g.Result.Outcome)
		}
		return string(domain.OutcomeLost)
	default:
		return "running"
	}
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func insertTokens(ctx context.Context, tx *sql.Tx, id domain.GameID, hashes map[domain.PlayerID]string) error {
	for pid, h := range hashes {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO player_tokens (game_id, player_id, token_hash) VALUES (?, ?, ?)`,
			string(id), string(pid), h); err != nil {
			return fmt.Errorf("insert token: %w", err)
		}
	}
	return nil
}

// Create implements ports.GameRepository.
func (r *GameRepository) Create(ctx context.Context, rec *ports.GameRecord) error {
	state, err := json.Marshal(toDTO(rec.Game))
	if err != nil {
		return fmt.Errorf("marshal game: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO games (id, variant, status, state, version) VALUES (?, ?, ?, ?, 1)`,
		string(rec.Game.ID), string(rec.Game.Options.Variant), statusOf(rec.Game), string(state)); err != nil {
		if isUniqueViolation(err) {
			return ports.ErrDuplicateGameID
		}
		return fmt.Errorf("insert game: %w", err)
	}
	if err := insertTokens(ctx, tx, rec.Game.ID, rec.TokenHashes); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	rec.Version = 1
	return nil
}

// Load implements ports.GameRepository.
func (r *GameRepository) Load(ctx context.Context, id domain.GameID) (*ports.GameRecord, error) {
	var state string
	var version int64
	err := r.db.QueryRowContext(ctx, `SELECT state, version FROM games WHERE id = ?`, string(id)).Scan(&state, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ports.ErrGameNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load game: %w", err)
	}
	var d gameDTO
	if err := json.Unmarshal([]byte(state), &d); err != nil {
		return nil, fmt.Errorf("unmarshal game %s: %w", id, err)
	}
	g, err := fromDTO(d)
	if err != nil {
		return nil, fmt.Errorf("decode game %s: %w", id, err)
	}

	rows, err := r.db.QueryContext(ctx, `SELECT player_id, token_hash FROM player_tokens WHERE game_id = ?`, string(id))
	if err != nil {
		return nil, fmt.Errorf("load tokens: %w", err)
	}
	defer rows.Close()
	hashes := map[domain.PlayerID]string{}
	for rows.Next() {
		var pid, h string
		if err := rows.Scan(&pid, &h); err != nil {
			return nil, fmt.Errorf("scan token: %w", err)
		}
		hashes[domain.PlayerID(pid)] = h
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load tokens: %w", err)
	}
	return &ports.GameRecord{Game: g, TokenHashes: hashes, Version: version}, nil
}

// Update implements ports.GameRepository.
func (r *GameRepository) Update(ctx context.Context, rec *ports.GameRecord) error {
	state, err := json.Marshal(toDTO(rec.Game))
	if err != nil {
		return fmt.Errorf("marshal game: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE games SET state = ?, status = ?, variant = ?, version = version + 1, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND version = ?`,
		string(state), statusOf(rec.Game), string(rec.Game.Options.Variant), string(rec.Game.ID), rec.Version)
	if err != nil {
		return fmt.Errorf("update game: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		var one int
		err := tx.QueryRowContext(ctx, `SELECT 1 FROM games WHERE id = ?`, string(rec.Game.ID)).Scan(&one)
		if errors.Is(err, sql.ErrNoRows) {
			return ports.ErrGameNotFound
		}
		if err != nil {
			return fmt.Errorf("check game: %w", err)
		}
		return ports.ErrVersionConflict
	}
	if err := insertTokens(ctx, tx, rec.Game.ID, rec.TokenHashes); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	rec.Version++
	return nil
}

// DeleteInactive implements ports.GameRepository.
func (r *GameRepository) DeleteInactive(ctx context.Context, olderThan time.Time) (int, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM games WHERE updated_at < ?`, olderThan.UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return 0, fmt.Errorf("delete inactive: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return int(n), nil
}
