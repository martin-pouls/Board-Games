package sqlite

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/gametest"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
)

// deepSeaGame starts a Deep Sea mission with the first seed (from 1 on) for which ok accepts the
// freshly dealt game; ok may be nil.
func deepSeaGame(t *testing.T, n, mission int, ok func(*domain.Game) bool) (*domain.Game, domain.RNG) {
	t.Helper()
	for seed := uint64(1); seed < 3000; seed++ {
		rng := gametest.NewRNG(seed)
		g := gametest.NewDeepSeaGame(t, n, mission, rng)
		if ok == nil || ok(g) {
			return g, rng
		}
	}
	t.Fatalf("no seed fits mission %d with %d players", mission, n)
	return nil, nil
}

func hasKind(g *domain.Game, kind domain.ObjectiveKind) bool {
	return slices.ContainsFunc(g.Attempt.Tasks, func(t domain.Task) bool { return t.Objective.Kind == kind })
}

// TestRoundTripSetupOfEveryDeepSeaMission snapshots the setup and the first tricks of every Deep
// Sea mission with every crew size: the drawn task cards, the deck, the used pile, selections,
// predictions and stages.
func TestRoundTripSetupOfEveryDeepSeaMission(t *testing.T) {
	for mission := 1; mission <= domain.Missions(domain.EditionDeepSea); mission++ {
		for n := 2; n <= 5; n++ {
			rng := gametest.NewRNG(uint64(mission*10 + n))
			g := gametest.NewDeepSeaGame(t, n, mission, rng)
			for range 300 {
				assertRoundTrip(t, g)
				if g.Phase != domain.PhasePlaying || len(g.Attempt.Tricks) >= 3 {
					break
				}
				gametest.Step(t, g, rng)
			}
		}
	}
}

// TestRoundTripDeepSeaDraftPassAssignPredict walks one game through a draft with a pass, the
// prediction (the hidden number is -1 until it is made) and the play, snapshotting every step.
func TestRoundTripDeepSeaDraftPassAssignPredict(t *testing.T) {
	// 5: one hidden prediction card and fewer tasks than members, so somebody may pass.
	g, rng := deepSeaGame(t, 3, 5, func(g *domain.Game) bool { return hasKind(g, domain.KindPredict) && len(g.Attempt.Tasks) < 3 })
	passed, predicted := false, false
	for range 300 {
		assertRoundTrip(t, g)
		a := g.Attempt
		if g.Phase != domain.PhasePlaying || len(a.Tricks) >= 2 {
			break
		}
		switch {
		case a.Step == domain.StepDraft && !passed && g.CanPass() == nil && len(g.Takeable()) > 0:
			if err := g.PassTask(g.Players[g.Actors()[0]].ID); err != nil {
				t.Fatal(err)
			}
			passed = true
			if got := roundTrip(t, g).Attempt.Passes; got != 1 {
				t.Fatalf("passes after the round trip = %d, want 1", got)
			}
		case a.Step == domain.StepPredict:
			for _, task := range a.Tasks {
				if task.Objective.Kind == domain.KindPredict && task.Objective.Count == -1 {
					if got := roundTrip(t, g).Attempt.Tasks[task.ID].Objective.Count; got != -1 {
						t.Fatalf("open prediction count = %d after the round trip, want -1", got)
					}
				}
			}
			gametest.Step(t, g, rng)
			predicted = true
		default:
			gametest.Step(t, g, rng)
		}
	}
	if !passed || !predicted {
		t.Fatalf("not the intended walk: passed %v predicted %v", passed, predicted)
	}
	for _, task := range roundTrip(t, g).Attempt.Tasks {
		if task.Objective.Kind == domain.KindPredict && task.Objective.Count != 2 {
			t.Fatalf("prediction after the round trip = %d, want 2", task.Objective.Count)
		}
	}
}

// TestRoundTripDeepSeaFreeSelection snapshots the captain's assignments (mission 17).
func TestRoundTripDeepSeaFreeSelection(t *testing.T) {
	rng := gametest.NewRNG(4)
	g := gametest.NewDeepSeaGame(t, 4, 17, rng)
	if g.Attempt.Step != domain.StepAssign {
		t.Fatalf("step = %s, want assign", g.Attempt.Step)
	}
	for g.Attempt.Step == domain.StepAssign {
		assertRoundTrip(t, g)
		gametest.Step(t, g, rng)
	}
	assertRoundTrip(t, g)
}

// TestRoundTripDeepSeaRaptureCommunications keeps several communications of one seat.
func TestRoundTripDeepSeaRaptureCommunications(t *testing.T) {
	for seed := uint64(1); seed < 100; seed++ {
		rng := gametest.NewRNG(seed)
		g := gametest.NewDeepSeaGame(t, 5, 11, rng) // rapture of the deep: n-2 = 3 shared tokens
		if !gametest.TryUntil(t, g, rng, func(g *domain.Game) bool { return g.Attempt.Step == domain.StepPlay }) {
			continue
		}
		seat := g.Actors()[0]
		for range 2 {
			v, err := g.ViewFor(g.Players[seat].ID)
			if err != nil {
				t.Fatal(err)
			}
			if len(v.Attempt.Actions.Communicate) == 0 {
				break
			}
			o := v.Attempt.Actions.Communicate[0]
			if err := g.Communicate(g.Players[seat].ID, o.Card, o.Position); err != nil {
				t.Fatal(err)
			}
		}
		if len(g.Attempt.Comms[seat]) < 2 {
			continue
		}
		if g.Attempt.Comm != domain.CommRapture || g.Attempt.TokensLeft() != 1 {
			t.Fatalf("comm %q, tokens left %d", g.Attempt.Comm, g.Attempt.TokensLeft())
		}
		assertRoundTrip(t, g)
		got := roundTrip(t, g)
		if !reflect.DeepEqual(got.Attempt.Comms[seat], g.Attempt.Comms[seat]) || got.Attempt.TokensLeft() != 1 {
			t.Fatalf("communications after the round trip: %+v", got.Attempt.Comms)
		}
		return
	}
	t.Fatal("no game with two communications of one seat")
}

// TestRoundTripDeepSeaUnfamiliarTerrainAndDebrief keeps the terrain card, the mode, the used pile
// and the finished attempt.
func TestRoundTripDeepSeaUnfamiliarTerrainAndDebrief(t *testing.T) {
	rng := gametest.NewRNG(6)
	g := gametest.NewDeepSeaGame(t, 3, 20, rng)
	if g.Attempt.Terrain == nil || len(g.UsedTasks) == 0 || g.Attempt.TaskDeck == nil {
		t.Fatalf("terrain %v, used %v, deck %v", g.Attempt.Terrain, g.UsedTasks, g.Attempt.TaskDeck)
	}
	assertRoundTrip(t, g)
	gametest.ToEnd(t, g, rng)
	assertRoundTrip(t, g)
	got := roundTrip(t, g)
	if got.Attempt.Terrain == nil || *got.Attempt.Terrain != *g.Attempt.Terrain || got.Attempt.Comm != g.Attempt.Comm {
		t.Fatalf("terrain %v/%q after the round trip, want %v/%q", got.Attempt.Terrain, got.Attempt.Comm, g.Attempt.Terrain, g.Attempt.Comm)
	}
	if !slices.Equal(got.UsedTasks, g.UsedTasks) {
		t.Fatalf("used tasks %v, want %v", got.UsedTasks, g.UsedTasks)
	}
}

// TestRoundTripDeepSeaStagesFaceDownAndRedeals covers the states that are hard to reach by
// playing: Q1's stage B, Q2's face-down tasks, a re-dealt attempt and the captain's choices.
func TestRoundTripDeepSeaStagesFaceDownAndRedeals(t *testing.T) {
	q1 := gametest.NewDeepSeaGame(t, 3, 42, gametest.NewRNG(1))
	q1.Attempt.Stage, q1.Attempt.Redeals = 1, 3
	q1.Attempt.Chosen = []int{2, 0}
	assertRoundTrip(t, q1)
	if got := roundTrip(t, q1).Attempt; got.Stage != 1 || got.Redeals != 3 || !slices.Equal(got.Chosen, []int{2, 0}) {
		t.Fatalf("stage %d redeals %d chosen %v", got.Stage, got.Redeals, got.Chosen)
	}

	q2 := gametest.NewDeepSeaGame(t, 4, 43, gametest.NewRNG(2))
	for _, task := range q2.Attempt.Tasks {
		if !task.FaceDown || task.CardID == 0 || task.Difficulty == 0 {
			t.Fatalf("Q2 task %+v", task)
		}
	}
	if q2.Attempt.TaskDeck != nil {
		t.Fatalf("Q2 keeps a task deck: %v", q2.Attempt.TaskDeck)
	}
	assertRoundTrip(t, q2)
}

// TestRoundTripNestedObjectives keeps every Deep Sea objective field, the recursive parts of an
// "all" objective included, and the unset card matches (nil, not an empty match).
func TestRoundTripNestedObjectives(t *testing.T) {
	g := baseGame("NESTD")
	g.Edition = domain.EditionDeepSea
	g.UsedTasks = []int{4, 8, 15}
	all := domain.Objective{Kind: domain.KindAll, Owner: 1, Parts: []domain.Objective{
		{Kind: domain.KindWinWith, Owner: 1, With: domain.CardMatch{Suit: domain.SuitTrump, Rank: 2}, Target: domain.CardMatch{Rank: 9}},
		{Kind: domain.KindAll, Owner: 1, Parts: []domain.Objective{{Kind: domain.KindWinTrickWhere, Where: domain.WhereSumBetween, Count: 8, Max: 12}}},
	}}
	g.Attempt = &domain.Attempt{
		Members:  []domain.Member{{PlayerID: "p1", Hand: []domain.Card{}}, {PlayerID: "p2", Hand: []domain.Card{}}},
		Comms:    [][]domain.Communication{nil, nil},
		Silenced: -1,
		Tasks: []domain.Task{
			{ID: 0, Objective: all, CardID: 12, Difficulty: 4},
			{ID: 1, Objective: domain.Objective{Kind: domain.KindWinCount, Cmp: domain.CmpAtLeast, Count: 3, Suits: []domain.Suit{domain.SuitPink, domain.SuitBlue}, Ranks: []domain.Rank{9}}, CardID: 3, Difficulty: 2},
			{ID: 2, Objective: domain.Objective{Kind: domain.KindEqualSuits, Suits: []domain.Suit{domain.SuitGreen, domain.SuitYellow}, InOneTrick: true, Exact: true, Hidden: true, Without: []domain.Rank{7, 8}}},
			{ID: 3, Objective: domain.Objective{Kind: domain.KindPredict, Count: -1, Hidden: true}, CardID: 91, Difficulty: 4},
		},
		Objectives: []domain.Objective{{Kind: domain.KindTrickQuota, Owner: domain.OwnerCrew, Targets: []int{1, 0, 2}}},
		Comm:       domain.CommCurrents,
		Terrain:    &domain.Card{Suit: domain.SuitBlue, Rank: 5},
		TaskDeck:   []int{},
		Passes:     2,
	}
	assertRoundTrip(t, g)
	if got := roundTrip(t, g).Attempt.Tasks[3].Objective.Count; got != -1 {
		t.Fatalf("open prediction count = %d, want -1", got)
	}
}

// oldKeys are the snapshot keys added for Deep Sea, per JSON object level.
var oldKeys = map[string][]string{
	"game":      {"usedTasks"},
	"attempt":   {"comm", "terrain", "taskDeck", "passes", "chosen", "stage", "redeals"},
	"task":      {"cardId", "difficulty"},
	"objective": {"cmp", "suits", "ranks", "with", "target", "where", "max", "exact", "inOneTrick", "hidden", "parts", "targets", "without"},
}

func dropKeys(m map[string]any, level string) {
	for _, k := range oldKeys[level] {
		delete(m, k)
	}
}

// TestSnapshotBeforeDeepSeaStillLoads: a Planet Nine snapshot saved before Deep Sea has none of the
// new keys and loads to the same game.
func TestSnapshotBeforeDeepSeaStillLoads(t *testing.T) {
	for _, mission := range []int{1, 5, 24, 50} {
		rng := gametest.NewRNG(uint64(mission))
		g := gametest.NewGame(t, 3, mission, rng)
		for range 40 {
			if g.Phase != domain.PhasePlaying || len(g.Attempt.Tricks) >= 2 {
				break
			}
			gametest.Step(t, g, rng)
		}
		g.PullEvents()
		b, err := json.Marshal(toDTO(g))
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]any
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}
		dropKeys(raw, "game")
		att, _ := raw["attempt"].(map[string]any)
		dropKeys(att, "attempt")
		list, _ := att["objectives"].([]any) // null without special objectives
		for _, o := range list {
			dropKeys(o.(map[string]any), "objective")
		}
		tasks, _ := att["tasks"].([]any)
		for _, task := range tasks {
			tm := task.(map[string]any)
			dropKeys(tm, "task")
			dropKeys(tm["objective"].(map[string]any), "objective")
		}
		old, err := json.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		var d gameDTO
		if err := json.Unmarshal(old, &d); err != nil {
			t.Fatal(err)
		}
		got, err := fromDTO(d)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(g, got) {
			t.Fatalf("mission %d: an old snapshot loads differently\nwant %+v\nhave %+v", mission, g.Attempt, got.Attempt)
		}
	}
}

// TestHandWrittenOldSnapshotLoads: a literal snapshot in the format before Deep Sea (one
// communication per seat in "comms", no Deep Sea key anywhere) loads with zero values.
func TestHandWrittenOldSnapshotLoads(t *testing.T) {
	const old = `{"schemaVersion":1,"id":"OLD01","phase":"playing","hostId":"p1","players":[{"id":"p1","name":"Ann"},{"id":"p2","name":"Ben"}],
"edition":"planetNine","options":{"distress":false},"mission":1,"logbook":{"1":{"attempts":1,"distress":false,"completed":false}},
"attempt":{"number":1,"members":[{"playerId":"p1","dummy":false,"hand":[{"suit":"pink","rank":1}],"stacks":null},{"playerId":"p2","dummy":false,"hand":[{"suit":"trump","rank":4}],"stacks":null}],
"commander":1,"step":"play","tasks":[{"id":0,"objective":{"kind":"winCards","owner":0,"cards":[{"suit":"pink","rank":1}],"count":0,"tricks":null,"suit":"","rank":0,"notWithTrump":false,"ordered":false},"token":"","faceDown":false}],
"objectives":[],"poll":null,"silenced":-1,"comms":[{"card":{"suit":"pink","rank":1},"position":"only","trick":1},null],"distressDecided":true,"pass":null,"exchange":null,
"handoverUsed":false,"tokensAdjusted":false,"trick":{"leader":1,"plays":[]},"tricks":[],"result":null},
"revision":7,"createdAt":"2026-09-01T10:00:00Z"}`
	var d gameDTO
	if err := json.Unmarshal([]byte(old), &d); err != nil {
		t.Fatal(err)
	}
	g, err := fromDTO(d)
	if err != nil {
		t.Fatal(err)
	}
	a := g.Attempt
	if g.UsedTasks != nil || a.Comm != domain.CommNormal || a.Terrain != nil || a.TaskDeck != nil || a.Passes != 0 || a.Chosen != nil || a.Stage != 0 || a.Redeals != 0 {
		t.Fatalf("Deep Sea fields are not zero: %+v", a)
	}
	if o := a.Tasks[0].Objective; o.Cmp != "" || o.Suits != nil || o.With != (domain.CardMatch{}) || o.Parts != nil || a.Tasks[0].CardID != 0 {
		t.Fatalf("objective = %+v", o)
	}
	if len(a.Comms[0]) != 1 || a.Comms[0][0].Position != domain.PosOnly || a.Comms[1] != nil {
		t.Fatalf("comms = %+v", a.Comms)
	}
	if _, err := g.ViewFor("p1"); err != nil {
		t.Fatal(err)
	}
	assertRoundTrip(t, g)
}
