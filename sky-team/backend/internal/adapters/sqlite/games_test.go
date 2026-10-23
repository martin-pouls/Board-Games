package sqlite

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/martin-pouls/sky-team/backend/internal/adapters/repotest"
	"github.com/martin-pouls/sky-team/backend/internal/core/domain"
	"github.com/martin-pouls/sky-team/backend/internal/core/ports"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestContract(t *testing.T) {
	repotest.Run(t, func(t *testing.T) ports.GameRepository { return NewGameRepository(openTestDB(t)) })
}

func baseGame(id string) *domain.Game {
	return &domain.Game{
		ID: domain.GameID(id), CreatedAt: time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC),
		Phase: domain.PhaseLobby, HostID: "p1", Revision: 1,
		Players: []domain.Player{{ID: "p1", Name: "Ann"}, {ID: "p2", Name: "Bob"}},
	}
}

func TestFromDTORejectsUnknownSchema(t *testing.T) {
	d := toDTO(baseGame("X"))
	d.SchemaVersion = 99
	if _, err := fromDTO(d); err == nil {
		t.Fatal("expected error for unknown schema version")
	}
	d.SchemaVersion = 0
	if _, err := fromDTO(d); err == nil {
		t.Fatal("expected error for missing schema version")
	}
}

func TestDeleteInactiveBackdatedAndCascade(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	repo := NewGameRepository(db)
	for _, id := range []string{"OLD01", "NEW01"} {
		rec := &ports.GameRecord{Game: baseGame(id), TokenHashes: map[domain.PlayerID]string{"p1": "same-hash"}}
		if err := repo.Create(ctx, rec); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.SQL().Exec(`UPDATE games SET updated_at = '2000-01-01 00:00:00' WHERE id = 'OLD01'`); err != nil {
		t.Fatal(err)
	}
	n, err := repo.DeleteInactive(ctx, time.Now().Add(-24*time.Hour))
	if err != nil || n != 1 {
		t.Fatalf("n=%d err=%v, want 1", n, err)
	}
	var tokens int
	if err := db.SQL().QueryRow(`SELECT COUNT(*) FROM player_tokens WHERE game_id = 'OLD01'`).Scan(&tokens); err != nil || tokens != 0 {
		t.Errorf("tokens of deleted game = %d (err %v), want 0", tokens, err)
	}
	if err := db.SQL().QueryRow(`SELECT COUNT(*) FROM player_tokens WHERE game_id = 'NEW01'`).Scan(&tokens); err != nil || tokens != 1 {
		t.Errorf("tokens of kept game = %d (err %v), want 1", tokens, err)
	}
}

func TestMigrationsCreateTokensTableAndIndex(t *testing.T) {
	db := openTestDB(t)
	var n int
	if err := db.SQL().QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name IN ('player_tokens','games_updated_at')`).Scan(&n); err != nil || n != 2 {
		t.Fatalf("n=%d err=%v, want 2", n, err)
	}
}

func TestRoundTripEachPhase(t *testing.T) {
	// nil slices must stay nil.
	noPlayers := baseGame("NILSL")
	noPlayers.Players = nil
	ended := baseGame("ENDED")
	ended.Phase = domain.PhaseEnded

	tests := []struct {
		name   string
		g      *domain.Game
		status string
	}{
		{"lobby", baseGame("LOBBY"), "lobby"},
		{"nil players", noPlayers, "lobby"},
		{"ended", ended, "ended"},
	}
	ctx := context.Background()
	db := openTestDB(t)
	repo := NewGameRepository(db)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &ports.GameRecord{Game: tt.g, TokenHashes: map[domain.PlayerID]string{"p1": "h1", "p2": "h2"}}
			if err := repo.Create(ctx, rec); err != nil {
				t.Fatal(err)
			}
			got, err := repo.Load(ctx, tt.g.ID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tt.g, got.Game) {
				t.Errorf("round trip mismatch\nwant %+v\nhave %+v", tt.g, got.Game)
			}
			want, _ := json.Marshal(toDTO(tt.g))
			have, _ := json.Marshal(toDTO(got.Game))
			if string(want) != string(have) {
				t.Errorf("dto mismatch\nwant %s\nhave %s", want, have)
			}
			var status string
			if err := db.SQL().QueryRow(`SELECT status FROM games WHERE id = ?`, string(tt.g.ID)).Scan(&status); err != nil {
				t.Fatal(err)
			}
			if status != tt.status {
				t.Errorf("status = %s, want %s", status, tt.status)
			}
		})
	}
}
