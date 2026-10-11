package sqlite

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenAppliesMigrationsIdempotently(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "test.db")

	for i := range 2 {
		db, err := Open(ctx, path)
		if err != nil {
			t.Fatalf("open #%d: %v", i, err)
		}
		var n int
		if err := db.SQL().QueryRowContext(ctx, `SELECT COUNT(*) FROM games`).Scan(&n); err != nil {
			t.Fatalf("query games table: %v", err)
		}
		db.Close()
	}
}
