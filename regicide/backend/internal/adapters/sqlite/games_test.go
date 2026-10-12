package sqlite

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/martin-pouls/regicide/backend/internal/adapters/repotest"
	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/ports"
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
		Players:     []domain.Player{{ID: "p1", Name: "Ann", Hand: []domain.Card{}}, {ID: "p2", Name: "Bob", Hand: []domain.Card{}}},
		Castle:      []domain.Card{},
		Tavern:      []domain.Card{},
		DiscardPile: []domain.Card{},
	}
}

func cs(s ...domain.Card) []domain.Card { return s }

func card(s domain.Suit, r domain.Rank) domain.Card { return domain.NewCard(s, r) }

func TestRoundTripEachPhase(t *testing.T) {
	H, D, C, S := domain.SuitHearts, domain.SuitDiamonds, domain.SuitClubs, domain.SuitSpades

	// A running game with Enemy set, a Jester play, signals and yield flags.
	playing := func(id string, step domain.Step) *domain.Game {
		g := baseGame(id)
		g.Phase = domain.PhasePlaying
		g.Step = step
		g.Revision = 7
		g.Players[0].Hand = cs(card(H, 2), card(S, 9), domain.Jester())
		g.Players[0].YieldedLastTurn = true
		g.Players[1].Hand = cs(card(D, 5))
		g.Players[1].Signal = domain.SignalRatherNot
		g.Castle = cs(card(C, 13), card(D, 12))
		e := card(S, 11)
		g.Enemy = &e
		g.Tavern = cs(card(H, 1), card(C, 4))
		g.DiscardPile = cs(card(D, 3))
		g.Played = []domain.PlayedSet{
			{By: "p1", Cards: cs(card(C, 8), card(C, 1))},
			{By: "p2", Cards: cs(domain.Jester())},
		}
		g.Turn = domain.Turn{Number: 3, ActiveSeat: 1}
		jack := card(H, 11)
		g.LastTurn = &domain.LastTurn{By: "p1", Played: cs(card(C, 8)), Defeated: &jack, Exact: true, Discarded: cs(card(D, 9))}
		g.SoloJestersUsed = 1
		return g
	}
	// nil slices must stay nil.
	nilSlices := baseGame("NILSL")
	nilSlices.Phase = domain.PhasePlaying
	nilSlices.Step = domain.StepPlay
	nilSlices.Castle, nilSlices.Tavern, nilSlices.DiscardPile, nilSlices.Played = nil, nil, nil, nil
	nilSlices.Players[0].Hand = nil
	nilSlices.LastTurn = &domain.LastTurn{By: "p1"} // a yield, no discard yet

	won := baseGame("WONNN")
	won.Phase = domain.PhaseWon
	won.Result = &domain.Result{}
	won.NextGameID = "NEXT1"
	won.Players[0].Hand = cs(card(H, 4))
	lost := playing("LOSTT", domain.StepDiscard)
	lost.Phase = domain.PhaseLost
	lost.Step = domain.StepNone
	lost.Result = &domain.Result{Reason: domain.LossCannotCover, PlayerID: "p2"}

	tests := []struct {
		name   string
		g      *domain.Game
		status string
	}{
		{"lobby", baseGame("LOBBY"), "lobby"},
		{"playing play", playing("PLAYP", domain.StepPlay), "playing"},
		{"playing discard", playing("PLAYD", domain.StepDiscard), "playing"},
		{"playing chooseNext", playing("PLAYC", domain.StepChooseNext), "playing"},
		{"nil slices", nilSlices, "playing"},
		{"won", won, "won"},
		{"lost", lost, "lost"},
	}
	// Enemy nil is covered by lobby and won; Enemy set by the playing games.
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
