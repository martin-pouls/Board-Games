// Package sqlitedb opens a game's embedded SQLite database, applies the game's migrations and
// stores games as JSON snapshots with optimistic concurrency (Store). It uses the pure-Go
// modernc.org/sqlite driver, so no CGO or C toolchain is required.
//
// The migrations belong to each game (its database, its history). Store needs the schema every
// game shares:
//
//	games(id TEXT PRIMARY KEY, status TEXT, state TEXT, version INTEGER, updated_at TEXT, ...)
//	player_tokens(game_id REFERENCES games(id) ON DELETE CASCADE, player_id, token_hash,
//	              PRIMARY KEY (game_id, player_id))
package sqlitedb

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	_ "modernc.org/sqlite"
)

// DB wraps the SQL connection pool.
type DB struct {
	sql *sql.DB
}

// Open opens (creating if needed) the database file and applies pending migrations: every
// *.sql file at the root of migrations (use fs.Sub for an embedded subdirectory).
func Open(ctx context.Context, path string, migrations fs.FS) (*DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// SQLite allows a single writer; serialising through one connection avoids SQLITE_BUSY.
	conn.SetMaxOpenConns(1)

	db := &DB{sql: conn}
	if err := db.migrate(ctx, migrations); err != nil {
		conn.Close()
		return nil, err
	}
	return db, nil
}

// SQL returns the connection pool, for queries a game runs itself (tests, extra tables).
func (db *DB) SQL() *sql.DB { return db.sql }

// Ping reports whether the database is reachable.
func (db *DB) Ping(ctx context.Context) error {
	return db.sql.PingContext(ctx)
}

// Close closes the underlying connection pool.
func (db *DB) Close() error {
	return db.sql.Close()
}

// migrate applies every *.sql file in migrations not yet recorded in schema_migrations, in
// lexical order, each inside its own transaction. A migration is recorded by its file name.
func (db *DB) migrate(ctx context.Context, migrations fs.FS) error {
	if _, err := db.sql.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
	); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	names, err := fs.Glob(migrations, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)

	for _, name := range names {
		var exists int
		if err := db.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		body, err := fs.ReadFile(migrations, name)
		if err != nil {
			return err
		}
		tx, err := db.sql.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (name) VALUES (?)`, name); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
