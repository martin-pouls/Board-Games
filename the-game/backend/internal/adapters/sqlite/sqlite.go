// Package sqlite is the driven adapter that persists game state in an embedded
// SQLite database. It uses the pure-Go modernc.org/sqlite driver, so no CGO or
// C toolchain is required.
package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB wraps the SQL connection pool. Repository implementations hang off it.
type DB struct {
	sql *sql.DB
}

// Open opens (creating if needed) the database file and applies pending migrations.
func Open(ctx context.Context, path string) (*DB, error) {
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
	if err := db.migrate(ctx); err != nil {
		conn.Close()
		return nil, err
	}
	return db, nil
}

// Ping reports whether the database is reachable.
func (db *DB) Ping(ctx context.Context) error {
	return db.sql.PingContext(ctx)
}

// Close closes the underlying connection pool.
func (db *DB) Close() error {
	return db.sql.Close()
}

// migrate applies every embedded migrations/*.sql file not yet recorded in
// schema_migrations, in lexical order, each inside its own transaction.
func (db *DB) migrate(ctx context.Context) error {
	if _, err := db.sql.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
	); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	names, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)

	for _, name := range names {
		id := strings.TrimPrefix(name, "migrations/")

		var exists int
		if err := db.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, id).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		body, err := migrationsFS.ReadFile(name)
		if err != nil {
			return err
		}
		tx, err := db.sql.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s: %w", id, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (name) VALUES (?)`, id); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
