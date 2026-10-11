package sqlite

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/martin-pouls/fox-duet/backend/internal/adapters/repotest"
	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
	"github.com/martin-pouls/fox-duet/backend/internal/core/ports"
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
		Players: []domain.Player{{ID: "p1", Name: "Ann", Hand: []domain.Card{}}, {ID: "p2", Name: "Bob", Hand: []domain.Card{}}},
	}
}

// playUntil starts a level-1 game with a fixed seed and plays the first legal option of every
// step until the game waits at step (or the game ends, which fails the test).
func playUntil(t *testing.T, id string, step domain.Step) *domain.Game {
	t.Helper()
	rng := rand.New(rand.NewPCG(7, 11))
	g, err := domain.NewGame(domain.GameID(id), "p1", "Ann")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Bob"); err != nil {
		t.Fatal(err)
	}
	if err := g.Start("p1", 1, rng); err != nil {
		t.Fatal(err)
	}
	for range 2000 {
		if g.Phase != domain.PhasePlaying {
			t.Fatalf("game over (%s) before reaching step %q", g.Phase, step)
		}
		if g.Step == step {
			g.PullEvents()
			return g
		}
		actor := g.Players[g.Actors()[0]].ID
		seat := g.Actors()[0]
		switch g.Step {
		case domain.StepLead, domain.StepFollow:
			err = g.PlayCard(actor, g.LegalCards(seat)[0], rng)
		case domain.StepFoxesChoose:
			err = g.ChooseFoxesPlayer(actor, g.Players[1-seat].ID, rng)
		case domain.StepFoxesSwap:
			err = g.KeepDecree(actor, rng)
		case domain.StepGift:
			err = g.GiveGift(actor, g.Players[seat].Hand[0], rng)
		case domain.StepMove:
			err = g.ChooseMove(actor, g.MoveOptions()[0].Delta, rng)
		case domain.StepForestEnd:
			err = g.ChooseForestEnd(actor, g.CoverableEnds()[0], rng)
		}
		if err != nil {
			t.Fatalf("step %s: %v", g.Step, err)
		}
	}
	t.Fatalf("never reached step %q", step)
	return nil
}

func TestRoundTripEachPhase(t *testing.T) {
	// nil slices must stay nil.
	nilSlices := baseGame("NILSL")
	nilSlices.Players[0].Hand = nil

	won := playUntil(t, "WONNN", domain.StepMove)
	won.Phase, won.Step, won.Result = domain.PhaseWon, domain.StepNone, &domain.Result{}
	won.NextGameID = "NEXT1"
	lost := playUntil(t, "LOSTT", domain.StepGift)
	lost.Phase, lost.Step = domain.PhaseLost, domain.StepNone
	lost.Result = &domain.Result{Reason: domain.LossGaveUp, PlayerID: "p2"}

	// A game in the middle of the gift step: one gift chosen, a last trick and a last gift
	// recorded, and a decree from the deal.
	gift := playUntil(t, "GIFTT", domain.StepGift)
	gift.Trick.Gift[0] = &domain.Card{Suit: domain.SuitRoses, Rank: 4}
	gift.LastGift = &domain.GiftRecord{Given: [2]domain.Card{{Suit: domain.SuitDoves, Rank: 2}, {Suit: domain.SuitStars, Rank: 6}}}
	gift.LastTrick = &domain.TrickRecord{
		Number: 2, Leader: 1, Lead: domain.Card{Suit: domain.SuitStars, Rank: 8}, Follow: domain.Card{Suit: domain.SuitStars, Rank: 3},
		Winner: 1, From: 5, Delta: -2, To: 3, Beyond: domain.EndNone, Collected: true,
	}
	// A game in the move step: both cards on the table, covered ends and a changed traveller.
	move := playUntil(t, "MOVEE", domain.StepMove)
	move.Board.CoveredLeft, move.Board.CoveredRight, move.Board.ForestTokens = 1, 2, 1
	move.Board.Gems[0], move.Board.Gems[1] = 0, 3
	move.Dealer, move.Round, move.Trick.Chosen = 1, 2, 1
	move.LastTrick = &domain.TrickRecord{Number: 1, Winner: 0, Beyond: domain.EndLeft}

	tests := []struct {
		name   string
		g      *domain.Game
		status string
	}{
		{"lobby", baseGame("LOBBY"), "lobby"},
		{"nil slices", nilSlices, "lobby"},
		{"gift step", gift, "playing"},
		{"move step", move, "playing"},
		{"playing lead", playUntil(t, "LEADD", domain.StepLead), "playing"},
		{"won", won, "won"},
		{"lost", lost, "lost"},
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
	if gift.Trick.Gift[0] == nil || move.Board.CoveredRight != 2 || gift.Decree == nil || len(gift.Deck) == 0 {
		t.Fatal("fixture games lack the state the round trip is meant to cover")
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
