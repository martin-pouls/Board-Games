package memory

import (
	"context"
	"testing"
	"time"

	"github.com/martin-pouls/fox-duet/backend/internal/adapters/repotest"
	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
	"github.com/martin-pouls/fox-duet/backend/internal/core/ports"
)

func TestContract(t *testing.T) {
	repotest.Run(t, func(t *testing.T) ports.GameRepository { return NewGameRepository() })
}

func TestDeleteInactiveWithInjectedClock(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	repo := NewGameRepository(WithNow(func() time.Time { return now }))

	mk := func(id string) *ports.GameRecord {
		return &ports.GameRecord{Game: &domain.Game{ID: domain.GameID(id), Phase: domain.PhaseLobby, Revision: 1}}
	}
	old, fresh := mk("OLD01"), mk("NEW01")
	if err := repo.Create(ctx, old); err != nil {
		t.Fatal(err)
	}
	now = now.Add(48 * time.Hour)
	if err := repo.Create(ctx, fresh); err != nil {
		t.Fatal(err)
	}
	n, err := repo.DeleteInactive(ctx, now.Add(-24*time.Hour))
	if err != nil || n != 1 {
		t.Fatalf("n=%d err=%v, want 1", n, err)
	}
	if _, err := repo.Load(ctx, "OLD01"); err == nil {
		t.Error("old game still present")
	}
	if _, err := repo.Load(ctx, "NEW01"); err != nil {
		t.Errorf("fresh game deleted: %v", err)
	}
}
