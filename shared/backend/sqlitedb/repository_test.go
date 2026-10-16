package sqlitedb

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

type testGame struct {
	ID        string    `json:"id"`
	Phase     string    `json:"phase"`
	CreatedAt time.Time `json:"createdAt"`
}

type testPlayer string

var (
	errGameNotFound = errors.New("game not found")
	errConflict     = errors.New("version conflict")
	errDuplicate    = errors.New("duplicate game")
)

func testCodec() Codec[*testGame] {
	return Codec[*testGame]{
		ID:     func(g *testGame) string { return g.ID },
		Status: func(g *testGame) string { return "status-" + g.Phase },
		Encode: func(g *testGame) ([]byte, error) { return json.Marshal(g) },
		Decode: func(state []byte, createdAt time.Time) (*testGame, error) {
			var g testGame
			if err := json.Unmarshal(state, &g); err != nil {
				return nil, err
			}
			if g.CreatedAt.IsZero() {
				g.CreatedAt = createdAt
			}
			return &g, nil
		},
	}
}

func newTestRepo(t *testing.T) (*Repository[*testGame, testPlayer], *DB) {
	t.Helper()
	db := openTest(t)
	return NewRepository[*testGame, testPlayer](db, testCodec(), Errors{NotFound: errGameNotFound, Conflict: errConflict, Duplicate: errDuplicate}), db
}

func TestRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()
	r, db := newTestRepo(t)

	created := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	g := &testGame{ID: "ABCDE", Phase: "lobby", CreatedAt: created}
	v, err := r.Create(ctx, g, map[testPlayer]string{"p1": "h1"})
	if err != nil || v != 1 {
		t.Fatalf("create = %d, %v; want 1", v, err)
	}
	if _, err := r.Create(ctx, g, nil); err != errDuplicate {
		t.Fatalf("duplicate create: %v, want the Duplicate error", err)
	}
	var status string
	if err := db.SQL().QueryRowContext(ctx, `SELECT status FROM games WHERE id = 'ABCDE'`).Scan(&status); err != nil || status != "status-lobby" {
		t.Fatalf("status column = %q, %v", status, err)
	}

	got, tokens, version, err := r.Load(ctx, "ABCDE")
	if err != nil || version != 1 || got.Phase != "lobby" || !got.CreatedAt.Equal(created) {
		t.Fatalf("load = %+v, %v, %d, %v", got, tokens, version, err)
	}
	if len(tokens) != 1 || tokens["p1"] != "h1" {
		t.Fatalf("tokens = %v", tokens)
	}

	got.Phase = "playing"
	tokens["p2"] = "h2"
	nv, err := r.Update(ctx, got, tokens, version)
	if err != nil || nv != 2 {
		t.Fatalf("update = %d, %v; want 2", nv, err)
	}
	if _, err := r.Update(ctx, got, tokens, version); err != errConflict {
		t.Fatalf("stale update: %v, want the Conflict error", err)
	}
	if _, err := r.Update(ctx, &testGame{ID: "ZZZZZ"}, nil, 1); err != errGameNotFound {
		t.Fatalf("update missing: %v, want the NotFound error", err)
	}
	if _, _, _, err := r.Load(ctx, "ZZZZZ"); err != errGameNotFound {
		t.Fatalf("load missing: %v, want the NotFound error", err)
	}

	again, tokens, version, err := r.Load(ctx, "ABCDE")
	if err != nil || version != 2 || again.Phase != "playing" || len(tokens) != 2 {
		t.Fatalf("after update: %+v, %v, %d, %v", again, tokens, version, err)
	}
}

func TestRepositoryDecodeGetsRowCreationTime(t *testing.T) {
	ctx := context.Background()
	r, db := newTestRepo(t)
	if _, err := r.Create(ctx, &testGame{ID: "OLD01", Phase: "lobby"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL().ExecContext(ctx, `UPDATE games SET created_at = '2023-02-03 04:05:06' WHERE id = 'OLD01'`); err != nil {
		t.Fatal(err)
	}
	g, _, _, err := r.Load(ctx, "OLD01")
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2023, 2, 3, 4, 5, 6, 0, time.UTC); !g.CreatedAt.Equal(want) {
		t.Fatalf("CreatedAt = %v, want %v", g.CreatedAt, want)
	}
}

func TestRepositoryCodecErrors(t *testing.T) {
	ctx := context.Background()
	r, db := newTestRepo(t)

	boom := errors.New("boom")
	enc := testCodec()
	enc.Encode = func(*testGame) ([]byte, error) { return nil, boom }
	bad := NewRepository[*testGame, testPlayer](db, enc, Errors{})
	if _, err := bad.Create(ctx, &testGame{ID: "X"}, nil); !errors.Is(err, boom) {
		t.Fatalf("encode error: %v, want wrapped boom", err)
	}
	if _, err := bad.Update(ctx, &testGame{ID: "X"}, nil, 1); !errors.Is(err, boom) {
		t.Fatalf("encode error on update: %v, want wrapped boom", err)
	}

	if _, err := r.Create(ctx, &testGame{ID: "BROKE"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL().ExecContext(ctx, `UPDATE games SET state = 'not json' WHERE id = 'BROKE'`); err != nil {
		t.Fatal(err)
	}
	g, tokens, version, err := r.Load(ctx, "BROKE")
	if err == nil || g != nil || tokens != nil || version != 0 {
		t.Fatalf("load of a broken snapshot = %v, %v, %d, %v; want an error only", g, tokens, version, err)
	}
	if !strings.HasPrefix(err.Error(), "decode game BROKE: ") {
		t.Fatalf("error = %q, want prefix %q", err, "decode game BROKE: ")
	}
}

func TestRepositoryDeleteInactive(t *testing.T) {
	ctx := context.Background()
	r, db := newTestRepo(t)
	for _, id := range []string{"OLD01", "NEW01"} {
		if _, err := r.Create(ctx, &testGame{ID: id}, map[testPlayer]string{"p": "h"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.SQL().ExecContext(ctx, `UPDATE games SET updated_at = '2000-01-01 00:00:00' WHERE id = 'OLD01'`); err != nil {
		t.Fatal(err)
	}
	if n, err := r.DeleteInactive(ctx, time.Now().Add(-time.Hour)); err != nil || n != 1 {
		t.Fatalf("DeleteInactive = %d, %v; want 1", n, err)
	}
	if _, _, _, err := r.Load(ctx, "OLD01"); err != errGameNotFound {
		t.Fatalf("old game: %v", err)
	}
	if _, _, _, err := r.Load(ctx, "NEW01"); err != nil {
		t.Fatalf("new game: %v", err)
	}
}
