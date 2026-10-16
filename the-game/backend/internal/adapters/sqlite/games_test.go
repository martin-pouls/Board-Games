package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
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
		ID: domain.GameID(id), Options: domain.Options{Variant: domain.VariantClassic}, Rules: rules,
		CreatedAt: time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC),
		Phase:     domain.PhaseLobby, HostID: "p1", Revision: 1,
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
	finished.Phase = domain.PhaseLost
	finished.Revision = 40
	finished.NextGameID = "NEXT1"
	finished.Players[1].Hand = cards(50, 51)
	finished.DrawPile = cards(70)
	finished.Turn = domain.Turn{Number: 12, ActiveSeat: 1, StartHandSize: 6, Plays: nil}
	finished.Result = &domain.Result{Score: 3, CardsLeftInHands: 2, CardsLeftInDrawPile: 1, Reason: domain.LossStuck, PlayerID: "p2"}

	won := baseGame("WONNN")
	won.Phase = domain.PhaseWon
	won.Result = &domain.Result{}

	ended := baseGame("ENDED")
	ended.Phase = domain.PhaseEnded

	tests := []struct {
		name   string
		g      *domain.Game
		status string
	}{
		{"lobby", baseGame("LOBBY"), "lobby"},
		{"choosing", choosing, "choosingFirstPlayer"},
		{"playing", playing, "playing"},
		{"lost", finished, "lost"},
		{"won", won, "won"},
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
			want, _ := json.Marshal(toDTO(tt.g))
			have, _ := json.Marshal(toDTO(got.Game))
			if string(want) != string(have) {
				t.Errorf("round trip mismatch\nwant %s\nhave %s", want, have)
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

// extremeCard returns the card with value v as the Extreme deck defines it.
func extremeCard(t *testing.T, rules domain.Ruleset, v int) domain.Card {
	t.Helper()
	deck, err := domain.NewDeck(rules)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range deck {
		if c.Value == v {
			return c
		}
	}
	t.Fatalf("card %d not in deck", v)
	return domain.Card{}
}

func TestRoundTripExtreme(t *testing.T) {
	g, err := domain.NewGame("XTREM", "p1", "Ann", domain.Options{Variant: domain.VariantExtreme, Pro: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Bob"); err != nil {
		t.Fatal(err)
	}
	ec := func(vs ...int) []domain.Card {
		out := make([]domain.Card, len(vs))
		for i, v := range vs {
			out[i] = extremeCard(t, g.Rules, v)
		}
		return out
	}
	g.Phase = domain.PhasePlaying
	g.Revision = 9
	g.Players[0].Hand = ec(95, 92, 30, 31) // STOP, No communication, plain
	g.Players[1].Hand = ec(9, 70, 8)       // Single pile, NBT, plain
	g.DrawPile = ec(87, 53, 61, 99)        // Skull, 3!, Skull, plain
	g.Piles[0].Cards = ec(28)              // 3! played this turn
	g.Piles[1].Cards = ec(41)              // uncovered skull played this turn
	g.Piles[2].Cards = ec(59)              // No backwards trick: Infinity top
	g.Turn = domain.Turn{
		Number: 4, ActiveSeat: 0, StartHandSize: 6,
		Plays: []domain.Play{{Card: g.Piles[0].Cards[0], Pile: 0}, {Card: g.Piles[1].Cards[0], Pile: 1}},
	}
	before, err := g.ViewFor("p1")
	if err != nil {
		t.Fatal(err)
	}
	if before.Turn == nil || before.Turn.Constraints.ExactCount != 3 || len(before.Turn.Constraints.SkullPiles) != 1 || len(before.ActiveCommands) != 1 {
		t.Fatalf("test setup: unexpected view %+v", before)
	}

	ctx := context.Background()
	db := openTestDB(t)
	repo := NewGameRepository(db)
	rec := &ports.GameRecord{Game: g, TokenHashes: map[domain.PlayerID]string{"p1": "h1", "p2": "h2"}}
	if err := repo.Create(ctx, rec); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Load(ctx, g.ID)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := json.Marshal(toDTO(g))
	have, _ := json.Marshal(toDTO(got.Game))
	if string(want) != string(have) {
		t.Errorf("round trip mismatch\nwant %s\nhave %s", want, have)
	}
	var status string
	if err := db.SQL().QueryRow(`SELECT status FROM games WHERE id = ?`, string(g.ID)).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "playing" {
		t.Errorf("status = %s, want playing", status)
	}
	after, err := got.Game.ViewFor("p1")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before.ActiveCommands, after.ActiveCommands) {
		t.Errorf("ActiveCommands %+v != %+v", before.ActiveCommands, after.ActiveCommands)
	}
	if !reflect.DeepEqual(before.Turn.Constraints, after.Turn.Constraints) {
		t.Errorf("Constraints %+v != %+v", before.Turn.Constraints, after.Turn.Constraints)
	}
	if !reflect.DeepEqual(before.Piles, after.Piles) || after.Hand[0].Card.Command != domain.CommandStop {
		t.Errorf("piles/hand differ after load: %+v", after)
	}
}

func TestLoadSnapshotWithoutCreatedAt(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	repo := NewGameRepository(db)

	g := baseGame("LEGAC")
	g.CreatedAt = time.Time{}
	if err := repo.Create(ctx, &ports.GameRecord{Game: g, TokenHashes: map[domain.PlayerID]string{"p1": "h"}}); err != nil {
		t.Fatal(err)
	}
	// Strip the field to mimic a snapshot written before it existed, and backdate the row.
	var raw map[string]any
	var state string
	if err := db.SQL().QueryRow(`SELECT state FROM games WHERE id = 'LEGAC'`).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(state), &raw); err != nil {
		t.Fatal(err)
	}
	delete(raw, "createdAt")
	stripped, _ := json.Marshal(raw)
	if _, err := db.SQL().Exec(`UPDATE games SET state = ?, created_at = '2026-05-04 03:02:01' WHERE id = 'LEGAC'`, string(stripped)); err != nil {
		t.Fatal(err)
	}

	got, err := repo.Load(ctx, "LEGAC")
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 5, 4, 3, 2, 1, 0, time.UTC); !got.Game.CreatedAt.Equal(want) {
		t.Errorf("CreatedAt = %v, want fallback to row timestamp %v", got.Game.CreatedAt, want)
	}

	// A snapshot that carries the field wins over the row timestamp.
	fresh := baseGame("FRESH")
	if err := repo.Create(ctx, &ports.GameRecord{Game: fresh, TokenHashes: map[domain.PlayerID]string{"p1": "h"}}); err != nil {
		t.Fatal(err)
	}
	got, err = repo.Load(ctx, "FRESH")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Game.CreatedAt.Equal(fresh.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.Game.CreatedAt, fresh.CreatedAt)
	}
}

// TestLoadLegacyFinishedSnapshot: snapshots written while every game over had phase "finished"
// plus an outcome still load, as won or lost.
func TestLoadLegacyFinishedSnapshot(t *testing.T) {
	tests := []struct {
		name    string
		result  string // the stored result JSON
		phase   domain.Phase
		want    domain.Result
		summary domain.Phase
	}{
		{"won", `{"outcome":"won","score":0,"cardsLeftInHands":0,"cardsLeftInDrawPile":0,"stuckPlayerId":""}`,
			domain.PhaseWon, domain.Result{}, domain.PhaseWon},
		{"lost", `{"outcome":"lost","score":3,"cardsLeftInHands":2,"cardsLeftInDrawPile":1,"stuckPlayerId":"p2"}`,
			domain.PhaseLost, domain.Result{Score: 3, CardsLeftInHands: 2, CardsLeftInDrawPile: 1, Reason: domain.LossStuck, PlayerID: "p2"}, domain.PhaseLost},
		{"aborted", `{"outcome":"aborted","score":7,"cardsLeftInHands":5,"cardsLeftInDrawPile":2,"stuckPlayerId":""}`,
			domain.PhaseLost, domain.Result{Score: 7, CardsLeftInHands: 5, CardsLeftInDrawPile: 2, Reason: domain.LossGaveUp}, domain.PhaseLost},
	}
	ctx := context.Background()
	db := openTestDB(t)
	repo := NewGameRepository(db)
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := domain.GameID(fmt.Sprintf("OLD%02d", i))
			g := baseGame(string(id))
			g.Phase = domain.PhasePlaying
			if err := repo.Create(ctx, &ports.GameRecord{Game: g, TokenHashes: map[domain.PlayerID]string{"p1": "h"}}); err != nil {
				t.Fatal(err)
			}
			var raw map[string]json.RawMessage
			var state string
			if err := db.SQL().QueryRow(`SELECT state FROM games WHERE id = ?`, string(id)).Scan(&state); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(state), &raw); err != nil {
				t.Fatal(err)
			}
			raw["phase"] = json.RawMessage(`"finished"`)
			raw["result"] = json.RawMessage(tt.result)
			old, _ := json.Marshal(raw)
			if _, err := db.SQL().Exec(`UPDATE games SET state = ?, status = 'finished' WHERE id = ?`, string(old), string(id)); err != nil {
				t.Fatal(err)
			}

			got, err := repo.Load(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if got.Game.Phase != tt.phase || got.Game.Result == nil || *got.Game.Result != tt.want {
				t.Fatalf("phase %s result %+v, want %s %+v", got.Game.Phase, got.Game.Result, tt.phase, tt.want)
			}
			if s := got.Game.Summary(); s.Phase != tt.summary {
				t.Fatalf("summary phase %s", s.Phase)
			}
			// Saving writes the new shape: no legacy fields, the status column is the phase.
			if err := repo.Update(ctx, got); err != nil {
				t.Fatal(err)
			}
			var status string
			if err := db.SQL().QueryRow(`SELECT state, status FROM games WHERE id = ?`, string(id)).Scan(&state, &status); err != nil {
				t.Fatal(err)
			}
			if status != string(tt.phase) || strings.Contains(state, `"outcome"`) || strings.Contains(state, `"finished"`) {
				t.Fatalf("saved status %s state %s", status, state)
			}
		})
	}
}
