package sqlite

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/martin-pouls/the-game/backend/internal/adapters/repotest"
	"github.com/martin-pouls/the-game/backend/internal/core/domain"
	"github.com/martin-pouls/the-game/backend/internal/core/ports"
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
	rules := domain.Ruleset{
		Variant: domain.VariantClassic, MinPlayers: 1, MaxPlayers: 5,
		MinCardsPerTurn: 2, MinCardsDrawPileEmpty: 1,
		HandSizes:          map[int]int{1: 8, 2: 7, 3: 6, 4: 6, 5: 6},
		BackwardsTrickStep: 10, AscendingStart: 1, DescendingStart: 100,
		LowestCard: 2, HighestCard: 99, AscendingPiles: 2, DescendingPiles: 2,
	}
	return &domain.Game{
		ID: domain.GameID(id), Options: domain.DefaultOptions(), Rules: rules,
		Phase: domain.PhaseLobby, HostID: "p1", Revision: 1,
		Players: []domain.Player{{ID: "p1", Name: "Ann", Hand: []domain.Card{}}, {ID: "p2", Name: "Bob", Hand: []domain.Card{}}},
		Piles:   domain.NewPiles(rules),
	}
}

func cards(vs ...int) []domain.Card {
	out := make([]domain.Card, len(vs))
	for i, v := range vs {
		out[i] = domain.NumberCard(v)
	}
	return out
}

func TestRoundTripEachPhase(t *testing.T) {
	choosing := baseGame("CHOOS")
	choosing.Phase = domain.PhaseChoosingFirstPlayer
	choosing.Revision = 2
	choosing.Players[0].Hand = cards(3, 20, 44)
	choosing.Players[1].Hand = cards(5, 6, 99)
	choosing.DrawPile = cards(10, 11, 12)

	playing := baseGame("PLAYG")
	playing.Phase = domain.PhasePlaying
	playing.Options = domain.Options{Variant: domain.VariantClassic, Pro: true, ReducedHandSize: true}
	playing.Rules.MinCardsPerTurn = 3
	playing.Rules.HandSizeReduction = 1
	playing.Revision = 7
	playing.Players[0].Hand = cards(30, 31)
	playing.Players[1].Hand = cards(8)
	playing.DrawPile = cards(60, 61)
	playing.Piles[0].Cards = cards(4, 9)
	playing.Piles[3].Cards = cards(90)
	playing.Turn = domain.Turn{
		Number: 3, ActiveSeat: 1, StartHandSize: 3,
		Plays: []domain.Play{{Card: domain.NumberCard(4), Pile: 0}, {Card: domain.NumberCard(90), Pile: 3}},
	}

	finished := baseGame("FINSH")
	finished.Phase = domain.PhaseFinished
	finished.Revision = 40
	finished.NextGameID = "NEXT1"
	finished.Players[1].Hand = cards(50, 51)
	finished.DrawPile = cards(70)
	finished.Turn = domain.Turn{Number: 12, ActiveSeat: 1, StartHandSize: 6, Plays: nil}
	finished.Result = &domain.Result{Outcome: domain.OutcomeLost, Score: 3, CardsLeftInHands: 2, CardsLeftInDrawPile: 1, StuckPlayerID: "p2"}

	won := baseGame("WONNN")
	won.Phase = domain.PhaseFinished
	won.Result = &domain.Result{Outcome: domain.OutcomeWon}

	tests := []struct {
		name   string
		g      *domain.Game
		status string
	}{
		{"lobby", baseGame("LOBBY"), "lobby"},
		{"choosing", choosing, "running"},
		{"playing", playing, "running"},
		{"finished lost", finished, "lost"},
		{"finished won", won, "won"},
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
			want, _ := json.Marshal(toDTO(tt.g))
			have, _ := json.Marshal(toDTO(got.Game))
			if string(want) != string(have) {
				t.Errorf("round trip mismatch\nwant %s\nhave %s", want, have)
			}
			var status, variant string
			if err := db.sql.QueryRow(`SELECT status, variant FROM games WHERE id = ?`, string(tt.g.ID)).Scan(&status, &variant); err != nil {
				t.Fatal(err)
			}
			if status != tt.status || variant != "classic" {
				t.Errorf("status/variant = %s/%s, want %s/classic", status, variant, tt.status)
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
	if _, err := db.sql.Exec(`UPDATE games SET updated_at = '2000-01-01 00:00:00' WHERE id = 'OLD01'`); err != nil {
		t.Fatal(err)
	}
	n, err := repo.DeleteInactive(ctx, time.Now().Add(-24*time.Hour))
	if err != nil || n != 1 {
		t.Fatalf("n=%d err=%v, want 1", n, err)
	}
	var tokens int
	if err := db.sql.QueryRow(`SELECT COUNT(*) FROM player_tokens WHERE game_id = 'OLD01'`).Scan(&tokens); err != nil || tokens != 0 {
		t.Errorf("tokens of deleted game = %d (err %v), want 0", tokens, err)
	}
	if err := db.sql.QueryRow(`SELECT COUNT(*) FROM player_tokens WHERE game_id = 'NEW01'`).Scan(&tokens); err != nil || tokens != 1 {
		t.Errorf("tokens of kept game = %d (err %v), want 1", tokens, err)
	}
}

func TestMigrationsCreateTokensTableAndIndex(t *testing.T) {
	db := openTestDB(t)
	var n int
	if err := db.sql.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name IN ('player_tokens','games_updated_at')`).Scan(&n); err != nil || n != 2 {
		t.Fatalf("n=%d err=%v, want 2", n, err)
	}
}
