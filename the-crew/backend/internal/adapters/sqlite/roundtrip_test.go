package sqlite

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/gametest"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
	"github.com/martin-pouls/the-crew/backend/internal/core/ports"
)

// roundTrip saves g through the DTO and JSON and returns the loaded copy.
func roundTrip(t *testing.T, g *domain.Game) *domain.Game {
	t.Helper()
	b, err := json.Marshal(toDTO(g))
	if err != nil {
		t.Fatal(err)
	}
	var d gameDTO
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	got, err := fromDTO(d)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func assertRoundTrip(t *testing.T, g *domain.Game) {
	t.Helper()
	g.PullEvents() // pending events are not persisted
	if got := roundTrip(t, g); !reflect.DeepEqual(g, got) {
		t.Fatalf("round trip mismatch\nwant %+v\nhave %+v\nwant attempt %+v\nhave attempt %+v", g, got, g.Attempt, got.Attempt)
	}
}

// midAttempt returns a game of n players playing mission 1 after a few tricks, with a
// communication on the table and (for 2 players) a dummy stack partially played.
func midAttempt(t *testing.T, id string, n int) *domain.Game {
	t.Helper()
	// The simple strategy often loses the task early; take the first seed that survives.
	for seed := uint64(1); seed < 200; seed++ {
		if g := tryMidAttempt(t, id, n, seed); g != nil {
			return g
		}
	}
	t.Fatal("no seed reaches a mid-attempt state")
	return nil
}

func tryMidAttempt(t *testing.T, id string, n int, seed uint64) *domain.Game {
	t.Helper()
	rng := gametest.NewRNG(seed)
	g := gametest.NewGame(t, n, 1, rng)
	g.ID = domain.GameID(id)
	ok := gametest.TryUntil(t, g, rng, func(g *domain.Game) bool {
		a := g.Attempt
		return a.Step == domain.StepPlay && len(a.Tricks) >= 2 && len(a.Trick.Plays) == 0
	})
	if !ok {
		return nil
	}
	// Someone communicates, if a card qualifies.
	for seat := range g.Players {
		v, err := g.ViewFor(g.Players[seat].ID)
		if err != nil {
			t.Fatal(err)
		}
		if opts := v.Attempt.Actions.Communicate; len(opts) > 0 {
			if err := g.Communicate(g.Players[seat].ID, opts[0].Card, opts[0].Position); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	// Play until a trick is in progress and, with the dummy, one of its stacks is half played.
	ok = gametest.TryUntil(t, g, rng, func(g *domain.Game) bool {
		a := g.Attempt
		if a.Step != domain.StepPlay || len(a.Trick.Plays) == 0 {
			return false
		}
		if n != 2 {
			return true
		}
		return halfPlayedStack(a)
	})
	if !ok || !communicated(g.Attempt) {
		return nil
	}
	return g
}

func halfPlayedStack(a *domain.Attempt) bool {
	for _, s := range a.Members[len(a.Members)-1].Stacks {
		if s.Up == nil && s.Down != nil {
			return true
		}
	}
	return false
}

func communicated(a *domain.Attempt) bool {
	for _, c := range a.Comms {
		if c != nil {
			return true
		}
	}
	return false
}

func TestRoundTripMidAttemptWithDummy(t *testing.T) {
	g := midAttempt(t, "MIDD2", 2)
	a := g.Attempt
	if !a.Members[2].Dummy || len(a.Tricks) < 2 || len(a.Trick.Plays) == 0 || !communicated(a) || !halfPlayedStack(a) {
		t.Fatalf("not the intended state: %+v", a)
	}
	assertRoundTrip(t, g)

	// Through the repository too.
	ctx := context.Background()
	repo := NewGameRepository(openTestDB(t))
	rec := &ports.GameRecord{Game: g, TokenHashes: map[domain.PlayerID]string{"p1": "h1", "p2": "h2"}}
	if err := repo.Create(ctx, rec); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.Load(ctx, g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(g, loaded.Game) {
		t.Fatalf("repository round trip mismatch\nwant %+v\nhave %+v", g.Attempt, loaded.Game.Attempt)
	}
	// The loaded game is playable to the end.
	gametest.ToEnd(t, loaded.Game, gametest.NewRNG(8))
}

func TestRoundTripMidAttemptFivePlayers(t *testing.T) {
	g := midAttempt(t, "MID05", 5)
	if len(g.Attempt.Members) != 5 || len(g.Attempt.Tricks) < 2 || len(g.Attempt.Trick.Plays) == 0 {
		t.Fatalf("not the intended state: %+v", g.Attempt)
	}
	assertRoundTrip(t, g)
}

func TestRoundTripDebrief(t *testing.T) {
	for _, n := range []int{2, 4} {
		rng := gametest.NewRNG(3)
		g := gametest.NewGame(t, n, 1, rng)
		gametest.ToEnd(t, g, rng)
		if g.Attempt.Result == nil || g.Phase != domain.PhaseDebrief || len(g.Logbook) == 0 {
			t.Fatalf("not a debrief: %+v", g)
		}
		assertRoundTrip(t, g)
	}
}

// TestRoundTripSetupOfEveryMission snapshots the setup steps and the first tricks of every
// mission with every crew size: polls, picks, drafts, silenced members, objectives with owners,
// card lists and trick lists.
func TestRoundTripSetupOfEveryMission(t *testing.T) {
	for mission := 1; mission <= domain.Missions(domain.EditionPlanetNine); mission++ {
		for n := 2; n <= 5; n++ {
			rng := gametest.NewRNG(uint64(mission*10 + n))
			g := gametest.NewGame(t, n, mission, rng)
			for range 200 {
				assertRoundTrip(t, g)
				if g.Phase != domain.PhasePlaying || len(g.Attempt.Tricks) >= 3 {
					break
				}
				gametest.Step(t, g, rng)
			}
		}
	}
}

// TestRoundTripDistressPass snapshots the pass in progress and the exchange afterwards.
func TestRoundTripDistressPass(t *testing.T) {
	rng := gametest.NewRNG(5)
	g := gametest.NewGame(t, 3, 1, rng)
	gametest.Until(t, g, rng, func(g *domain.Game) bool { return g.Attempt.Step == domain.StepPlay })
	if err := g.SendDistress(g.Players[g.Attempt.Commander].ID, domain.DirLeft); err != nil {
		t.Fatal(err)
	}
	assertRoundTrip(t, g)
	first := g.Actors()[0]
	var c domain.Card
	for _, h := range g.Attempt.Members[first].Hand {
		if !h.IsTrump() {
			c = h
			break
		}
	}
	if err := g.ChoosePassCard(g.Players[first].ID, c); err != nil {
		t.Fatal(err)
	}
	assertRoundTrip(t, g)
	for g.Attempt.Step == domain.StepPass {
		gametest.Step(t, g, rng)
	}
	if g.Attempt.Exchange == nil {
		t.Fatal("no exchange")
	}
	assertRoundTrip(t, g)
}

// TestRoundTripNilAndEmpty keeps nil and empty slices and maps apart.
func TestRoundTripNilAndEmpty(t *testing.T) {
	g := baseGame("NILEM")
	g.Edition = domain.EditionPlanetNine
	g.Logbook = map[int]domain.LogEntry{}
	assertRoundTrip(t, g)
	g.Logbook = map[int]domain.LogEntry{3: {Attempts: 2, Distress: true, Completed: true}}
	g.Attempt = &domain.Attempt{
		Members:  []domain.Member{{PlayerID: "p1", Hand: []domain.Card{}}, {Dummy: true}},
		Comms:    []*domain.Communication{nil, nil},
		Silenced: -1,
	}
	assertRoundTrip(t, g)
}
