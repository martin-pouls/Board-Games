package sqlitedb

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"
)

// schema is the shape every game's migrations create (see the package doc).
const schema = `
CREATE TABLE games (
    id         TEXT PRIMARY KEY,
    status     TEXT    NOT NULL,
    state      TEXT    NOT NULL,
    version    INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE player_tokens (
    game_id    TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id  TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (game_id, player_id)
);`

var migrations = fstest.MapFS{
	"0001_init.sql":  {Data: []byte(schema)},
	"0002_extra.sql": {Data: []byte(`CREATE TABLE extra (x INTEGER);`)},
	"README.md":      {Data: []byte(`not a migration`)},
}

func openTest(t *testing.T) *DB {
	t.Helper()
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "sub", "test.db"), migrations)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpenAppliesMigrationsOnceInOrder(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "test.db")
	for i := range 2 {
		db, err := Open(ctx, path, migrations)
		if err != nil {
			t.Fatalf("open #%d: %v", i, err)
		}
		var n int
		if err := db.SQL().QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 2 {
			t.Fatalf("open #%d: %d migrations recorded, want 2", i, n)
		}
		if err := db.SQL().QueryRowContext(ctx, `SELECT COUNT(*) FROM extra`).Scan(&n); err != nil {
			t.Fatalf("open #%d: extra table: %v", i, err)
		}
		if err := db.Ping(ctx); err != nil {
			t.Fatal(err)
		}
		db.Close()
	}
}

func TestOpenFailsOnBrokenMigration(t *testing.T) {
	bad := fstest.MapFS{"0001_bad.sql": {Data: []byte(`CREATE TABLE (`)}}
	if _, err := Open(context.Background(), filepath.Join(t.TempDir(), "x.db"), bad); err == nil {
		t.Fatal("want an error for a broken migration")
	}
}

func TestStoreLifecycle(t *testing.T) {
	ctx := context.Background()
	s := NewStore(openTest(t))

	rec := &Record{ID: "ABCDE", Status: "lobby", State: []byte(`{"a":1}`), Tokens: map[string]string{"p1": "h1"}}
	if err := s.Create(ctx, rec); err != nil {
		t.Fatal(err)
	}
	if rec.Version != 1 {
		t.Fatalf("version after create = %d, want 1", rec.Version)
	}
	if err := s.Create(ctx, &Record{ID: "ABCDE", Status: "lobby", State: []byte(`{}`)}); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate create: %v, want ErrDuplicate", err)
	}

	got, err := s.Load(ctx, "ABCDE")
	if err != nil {
		t.Fatal(err)
	}
	if string(got.State) != `{"a":1}` || got.Status != "lobby" || got.Version != 1 || got.Tokens["p1"] != "h1" {
		t.Fatalf("loaded %+v", got)
	}

	got.State = []byte(`{"a":2}`)
	got.Status = "playing"
	got.Tokens["p2"] = "h2"
	if err := s.Update(ctx, got); err != nil {
		t.Fatal(err)
	}
	if got.Version != 2 {
		t.Fatalf("version after update = %d, want 2", got.Version)
	}

	stale := *rec // still version 1
	stale.State = []byte(`{"a":3}`)
	if err := s.Update(ctx, &stale); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale update: %v, want ErrConflict", err)
	}
	if err := s.Update(ctx, &Record{ID: "ZZZZZ", Version: 1}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("update missing: %v, want ErrNotFound", err)
	}
	if _, err := s.Load(ctx, "ZZZZZ"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("load missing: %v, want ErrNotFound", err)
	}

	again, err := s.Load(ctx, "ABCDE")
	if err != nil {
		t.Fatal(err)
	}
	if string(again.State) != `{"a":2}` || again.Version != 2 || len(again.Tokens) != 2 {
		t.Fatalf("after update: %+v", again)
	}
}

func TestDeleteInactiveCascadesTokens(t *testing.T) {
	ctx := context.Background()
	db := openTest(t)
	s := NewStore(db)
	for _, id := range []string{"OLD01", "NEW01"} {
		if err := s.Create(ctx, &Record{ID: id, Status: "lobby", State: []byte(`{}`), Tokens: map[string]string{"p": "h"}}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.SQL().ExecContext(ctx, `UPDATE games SET updated_at = '2000-01-01 00:00:00' WHERE id = 'OLD01'`); err != nil {
		t.Fatal(err)
	}
	n, err := s.DeleteInactive(ctx, time.Now().Add(-time.Hour))
	if err != nil || n != 1 {
		t.Fatalf("DeleteInactive = %d, %v; want 1", n, err)
	}
	if _, err := s.Load(ctx, "OLD01"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("old game: %v", err)
	}
	var tokens int
	if err := db.SQL().QueryRowContext(ctx, `SELECT COUNT(*) FROM player_tokens WHERE game_id = 'OLD01'`).Scan(&tokens); err != nil || tokens != 0 {
		t.Fatalf("tokens of the deleted game: %d, %v", tokens, err)
	}
	if _, err := s.Load(ctx, "NEW01"); err != nil {
		t.Fatalf("new game: %v", err)
	}
}
