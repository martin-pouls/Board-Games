// Package repotest holds the shared contract test suite for ports.GameRepository implementations.
package repotest

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/martin-pouls/sky-team/backend/internal/core/domain"
	"github.com/martin-pouls/sky-team/backend/internal/core/ports"
)

func newGame(id string) *domain.Game {
	return &domain.Game{
		ID:      domain.GameID(id),
		Phase:   domain.PhaseLobby,
		HostID:  "p1",
		Players: []domain.Player{{ID: "p1", Name: "Ann"}},
		// Revision starts at 1 like domain.NewGame.
		Revision:  1,
		CreatedAt: time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC),
	}
}

func newRec(id string) *ports.GameRecord {
	return &ports.GameRecord{Game: newGame(id), TokenHashes: map[domain.PlayerID]string{"p1": "hash1"}}
}

func mustCreate(t *testing.T, repo ports.GameRepository, rec *ports.GameRecord) {
	t.Helper()
	if err := repo.Create(context.Background(), rec); err != nil {
		t.Fatalf("create: %v", err)
	}
}

// Run executes the contract suite against fresh repositories from newRepo.
func Run(t *testing.T, newRepo func(t *testing.T) ports.GameRepository) {
	ctx := context.Background()

	t.Run("create and load round trip", func(t *testing.T) {
		repo := newRepo(t)
		rec := newRec("AAAAA")
		rec.Game.Players = append(rec.Game.Players, domain.Player{ID: "p2", Name: "Bob"})
		rec.TokenHashes["p2"] = "hash2"
		mustCreate(t, repo, rec)
		if rec.Version != 1 {
			t.Fatalf("version after create = %d, want 1", rec.Version)
		}
		got, err := repo.Load(ctx, "AAAAA")
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		if got.Version != 1 {
			t.Errorf("loaded version = %d, want 1", got.Version)
		}
		if got.Game.ID != "AAAAA" || got.Game.Revision != 1 {
			t.Errorf("unexpected game: %+v", got.Game)
		}
		if len(got.Game.Players) != 2 || got.Game.Players[1].Name != "Bob" {
			t.Errorf("players not round-tripped: %+v", got.Game.Players)
		}
		if !reflect.DeepEqual(got.TokenHashes, map[domain.PlayerID]string{"p1": "hash1", "p2": "hash2"}) {
			t.Errorf("tokens = %v", got.TokenHashes)
		}
	})

	t.Run("duplicate id", func(t *testing.T) {
		repo := newRepo(t)
		mustCreate(t, repo, newRec("AAAAA"))
		if err := repo.Create(ctx, newRec("AAAAA")); !errors.Is(err, ports.ErrDuplicateGameID) {
			t.Fatalf("err = %v, want ErrDuplicateGameID", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := newRepo(t)
		if _, err := repo.Load(ctx, "NOPE1"); !errors.Is(err, ports.ErrGameNotFound) {
			t.Fatalf("load err = %v, want ErrGameNotFound", err)
		}
		rec := newRec("NOPE1")
		rec.Version = 1
		if err := repo.Update(ctx, rec); !errors.Is(err, ports.ErrGameNotFound) {
			t.Fatalf("update err = %v, want ErrGameNotFound", err)
		}
	})

	t.Run("update increments version and persists state", func(t *testing.T) {
		repo := newRepo(t)
		rec := newRec("AAAAA")
		mustCreate(t, repo, rec)
		rec.Game.Revision = 2
		rec.Game.Phase = domain.PhasePlaying
		if err := repo.Update(ctx, rec); err != nil {
			t.Fatalf("update: %v", err)
		}
		if rec.Version != 2 {
			t.Fatalf("rec.Version = %d, want 2", rec.Version)
		}
		got, err := repo.Load(ctx, "AAAAA")
		if err != nil {
			t.Fatal(err)
		}
		if got.Version != 2 || got.Game.Revision != 2 || got.Game.Phase != domain.PhasePlaying {
			t.Errorf("loaded: version=%d revision=%d phase=%s", got.Version, got.Game.Revision, got.Game.Phase)
		}
		// Second update on the refreshed record works too.
		if err := repo.Update(ctx, rec); err != nil {
			t.Fatalf("second update: %v", err)
		}
		if rec.Version != 3 {
			t.Errorf("rec.Version = %d, want 3", rec.Version)
		}
	})

	t.Run("version conflict", func(t *testing.T) {
		repo := newRepo(t)
		mustCreate(t, repo, newRec("AAAAA"))
		a, _ := repo.Load(ctx, "AAAAA")
		b, _ := repo.Load(ctx, "AAAAA")
		if err := repo.Update(ctx, a); err != nil {
			t.Fatalf("first update: %v", err)
		}
		if err := repo.Update(ctx, b); !errors.Is(err, ports.ErrVersionConflict) {
			t.Fatalf("stale update err = %v, want ErrVersionConflict", err)
		}
		if b.Version != 1 {
			t.Errorf("stale rec.Version changed to %d", b.Version)
		}
	})

	t.Run("new tokens added on update, existing kept", func(t *testing.T) {
		repo := newRepo(t)
		rec := newRec("AAAAA")
		mustCreate(t, repo, rec)
		rec.Game.Players = append(rec.Game.Players, domain.Player{ID: "p2", Name: "Bob"})
		rec.TokenHashes["p2"] = "hash2"
		rec.TokenHashes["p1"] = "changed" // must be ignored
		if err := repo.Update(ctx, rec); err != nil {
			t.Fatal(err)
		}
		got, _ := repo.Load(ctx, "AAAAA")
		want := map[domain.PlayerID]string{"p1": "hash1", "p2": "hash2"}
		if !reflect.DeepEqual(got.TokenHashes, want) {
			t.Errorf("tokens = %v, want %v", got.TokenHashes, want)
		}
	})

	t.Run("same token hash allowed in different games", func(t *testing.T) {
		repo := newRepo(t)
		mustCreate(t, repo, newRec("AAAAA"))
		mustCreate(t, repo, newRec("BBBBB"))
		got, err := repo.Load(ctx, "BBBBB")
		if err != nil || got.TokenHashes["p1"] != "hash1" {
			t.Fatalf("load: %v %v", err, got)
		}
	})

	t.Run("load returns independent objects", func(t *testing.T) {
		repo := newRepo(t)
		rec := newRec("AAAAA")
		mustCreate(t, repo, rec)

		// Mutating the record passed to Create must not affect storage.
		rec.Game.Players[0].Name = "Changed"
		rec.TokenHashes["zz"] = "x"

		a, _ := repo.Load(ctx, "AAAAA")
		if a.Game.Players[0].Name != "Ann" || len(a.TokenHashes) != 1 {
			t.Fatalf("storage aliased the created record: %+v %v", a.Game.Players, a.TokenHashes)
		}
		a.Game.Players[0].Name = "Mutated"
		a.TokenHashes["p1"] = "mutated"

		b, _ := repo.Load(ctx, "AAAAA")
		if b == a || b.Game == a.Game {
			t.Fatal("Load returned the same object twice")
		}
		if b.Game.Players[0].Name != "Ann" ||
			b.TokenHashes["p1"] != "hash1" {
			t.Errorf("mutation of a loaded record leaked into storage: %+v", b)
		}
	})

	t.Run("delete inactive", func(t *testing.T) {
		repo := newRepo(t)
		mustCreate(t, repo, newRec("AAAAA"))
		mustCreate(t, repo, newRec("BBBBB"))

		n, err := repo.DeleteInactive(ctx, time.Now().Add(-time.Hour))
		if err != nil || n != 0 {
			t.Fatalf("past cutoff: n=%d err=%v, want 0", n, err)
		}
		if _, err := repo.Load(ctx, "AAAAA"); err != nil {
			t.Fatalf("game deleted by past cutoff: %v", err)
		}

		n, err = repo.DeleteInactive(ctx, time.Now().Add(time.Hour))
		if err != nil || n != 2 {
			t.Fatalf("future cutoff: n=%d err=%v, want 2", n, err)
		}
		if _, err := repo.Load(ctx, "AAAAA"); !errors.Is(err, ports.ErrGameNotFound) {
			t.Fatalf("load after delete err = %v", err)
		}
		// The id can be reused once deleted.
		mustCreate(t, repo, newRec("AAAAA"))
	})
}
