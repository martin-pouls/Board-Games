package apidto

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/gametest"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
)

func TestDeepSeaCatalogue(t *testing.T) {
	c := Catalogue(domain.EditionDeepSea)
	if c.Edition != "deepSea" || len(c.Missions) != 43 {
		t.Fatalf("catalogue = %s %d", c.Edition, len(c.Missions))
	}
	assertNoNullsExcept(t, c, map[string]bool{"holder": true, "with": true, "target": true})
	m := func(n int) MissionDef { return c.Missions[n-1] }
	if x := m(3); x.Difficulty != 3 || x.Distribution != "draft" || x.CommMode != "" || x.Label != "" || x.CaptainCard != (Card{"trump", 4}) ||
		len(x.DifficultyBy) != 3 || x.FixedTasks == nil || x.StageObjectives == nil {
		t.Errorf("mission 3 = %+v", x)
	}
	if x := m(9); x.CommMode != "currents" {
		t.Errorf("mission 9 = %+v", x)
	}
	if x := m(11); x.CommMode != "rapture" {
		t.Errorf("mission 11 = %+v", x)
	}
	if x := m(16); x.CommMode != "none" || x.Volunteers != 1 || x.Distribution != "volunteers" {
		t.Errorf("mission 16 = %+v", x)
	}
	if x := m(20); x.CommMode != "unfamiliar" {
		t.Errorf("mission 20 = %+v", x)
	}
	if x := m(8); len(x.Objectives) != 1 || x.Objectives[0].Kind != "crewBalance" || x.Objectives[0].Rank != 9 || x.Objectives[0].OwnerRule != "crew" {
		t.Errorf("mission 8 = %+v", x)
	}
	if x := m(12); len(x.Objectives) != 1 || x.Objectives[0].Kind != "neverLead" || !slices.Equal(x.Objectives[0].Suits, []string{"pink", "trump"}) {
		t.Errorf("mission 12 = %+v", x)
	}
	if x := m(19); !x.HardestFirst {
		t.Errorf("mission 19 = %+v", x)
	}
	if x := m(25); !x.CaptainSkipped {
		t.Errorf("mission 25 = %+v", x)
	}
	if x := m(32); !slices.Equal(x.FixedTasks, []int{74, 88, 85, 80}) {
		t.Errorf("mission 32 = %+v", x)
	}
	if x := m(33); !x.Uncounted || x.Distribution != "free" {
		t.Errorf("mission 33 = %+v", x)
	}
	if x := m(42); x.Label != "Q1" || x.Stages != 2 || len(x.StageObjectives) != 2 || x.CaptainCard != (Card{"trump", 1}) ||
		x.StageObjectives[1][0].Count != 3 || !slices.Equal(x.StageObjectives[1][0].Without, []int{7, 8, 9}) || x.StageObjectives[0][0].OwnerRule != "commander" {
		t.Errorf("mission 42 = %+v", x)
	}
	if x := m(43); x.Label != "Q2" || !x.FaceDownTasks || !slices.Equal(x.DifficultyBy, []int{13, 10, 8}) {
		t.Errorf("mission 43 = %+v", x)
	}
	// A Planet Nine mission has the neutral Deep Sea values.
	if x := Catalogue(domain.EditionPlanetNine).Missions[0]; x.Difficulty != 0 || x.CommMode != "" || x.Stages != 0 || x.Uncounted || x.CaptainCard != (Card{"trump", 4}) {
		t.Errorf("planet nine mission 1 = %+v", x)
	}
}

func TestFromObjectiveDeepSeaFields(t *testing.T) {
	o := domain.Objective{
		Kind: domain.KindAll, Owner: 2,
		Parts: []domain.Objective{
			{Kind: domain.KindWinWith, With: domain.CardMatch{Suit: domain.SuitTrump, Rank: 3}, Target: domain.CardMatch{Rank: 9}},
			{Kind: domain.KindWinCount, Cmp: domain.CmpAtLeast, Count: 2, Suits: []domain.Suit{domain.SuitPink}, Ranks: []domain.Rank{8, 9}},
		},
		Where: domain.WhereSumBetween, Max: 12, Exact: true, InOneTrick: true, Hidden: true,
		Targets: []int{1, 2}, Without: []domain.Rank{7},
	}
	got := fromObjective(o)
	want := Objective{
		Kind: "all", Owner: 2, Cards: []Card{}, Tricks: []int{}, Suits: []string{}, Ranks: []int{}, Where: "sumBetween", Max: 12,
		Exact: true, InOneTrick: true, Hidden: true, Targets: []int{1, 2}, Without: []int{7},
		Parts: []Objective{
			{Kind: "winWith", Cards: []Card{}, Tricks: []int{}, Suits: []string{}, Ranks: []int{}, With: &CardMatch{"trump", 3}, Target: &CardMatch{"", 9}, Parts: []Objective{}, Targets: []int{}, Without: []int{}},
			{Kind: "winCount", Cmp: "atLeast", Count: 2, Cards: []Card{}, Tricks: []int{}, Suits: []string{"pink"}, Ranks: []int{8, 9}, Parts: []Objective{}, Targets: []int{}, Without: []int{}},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("objective = %+v\nwant %+v", got, want)
	}
	// An unset card match is null, not an empty match; the slices are never null.
	s := mustJSON(t, fromObjective(domain.Objective{Kind: domain.KindWinWith}))
	for _, frag := range []string{`"with":null`, `"target":null`, `"suits":[]`, `"ranks":[]`, `"parts":[]`, `"targets":[]`, `"without":[]`} {
		if !strings.Contains(s, frag) {
			t.Errorf("%s is missing %s", s, frag)
		}
	}
}

// The views of every Deep Sea mission, crew size and step carry no unexpected null.
func TestDeepSeaViewsNeverHaveUnexpectedNulls(t *testing.T) {
	for mission := 1; mission <= 43; mission++ {
		for _, n := range []int{2, 3, 5} {
			rng := gametest.NewRNG(uint64(mission*7 + n))
			g := gametest.NewDeepSeaGame(t, n, mission, rng)
			for range 80 {
				for seat := range n {
					assertNoNullsExcept(t, viewOf(t, g, seat), map[string]bool{
						"missionDef": true, "attempt": true, "card": true, "up": true, "communication": true, "poll": true, "pass": true,
						"exchange": true, "lastTrick": true, "result": true, "cards": true, "chosen": true, "gave": true, "got": true,
						"holder": true, "with": true, "target": true, "terrain": true, "objective": true,
					})
				}
				if g.Phase != domain.PhasePlaying {
					break
				}
				gametest.Step(t, g, rng)
			}
		}
	}
}

// A hidden prediction (card 91) of another member shows count -1 and predicted; the owner sees
// the number; in the debrief everybody does. The mission def and Deep Sea attempt fields are set.
func TestDeepSeaPredictionOnTheWire(t *testing.T) {
	var g *domain.Game
	var rng domain.RNG
	for seed := uint64(1); seed < 3000 && g == nil; seed++ {
		r := gametest.NewRNG(seed)
		c := gametest.NewDeepSeaGame(t, 3, 7, r)
		for _, task := range c.Attempt.Tasks {
			if task.CardID == 91 {
				g, rng = c, r
			}
		}
	}
	if g == nil {
		t.Fatal("no seed with card 91")
	}
	pending := func() (int, bool) {
		for _, task := range g.Attempt.Tasks {
			if task.CardID == 91 {
				return task.ID, true
			}
		}
		return 0, false
	}
	id, _ := pending()
	if v := viewOf(t, g, 0).Attempt; v.Predict != -1 || v.Takeable == nil || v.Assignable == nil || v.Chosen == nil || v.TokensLeft != -1 || v.Stages != 0 {
		t.Fatalf("setup view = %+v", v)
	}
	for g.Attempt.Step != domain.StepPredict {
		gametest.Step(t, g, rng)
	}
	owner := g.Attempt.Tasks[id].Owner()
	other := (owner + 1) % 3
	if v := viewOf(t, g, owner).Attempt; v.Step != "predict" || v.Predict != id || v.Tasks[id].Predicted || v.Tasks[id].Objective.Count != -1 {
		t.Fatalf("before the prediction: %+v", v.Tasks[id])
	}
	gametest.Step(t, g, rng)
	events := FromEvents(g.PullEvents())
	if e := events[len(events)-1]; e.Type != EventPredicted || *e.Tricks != -1 || *e.TaskID != id {
		t.Fatalf("hidden prediction event = %s", mustJSON(t, events))
	}
	ov, xv := viewOf(t, g, owner).Attempt.Tasks[id], viewOf(t, g, other).Attempt.Tasks[id]
	if !ov.Predicted || ov.Objective.Count != 2 || !ov.Objective.Hidden || ov.CardID != 91 || ov.Difficulty != 4 || ov.Card != nil {
		t.Fatalf("owner's task = %+v", ov)
	}
	if !xv.Predicted || xv.Objective.Count != -1 || xv.CardID != 91 {
		t.Fatalf("other member's task = %+v / %s", xv, mustJSON(t, xv))
	}
	gametest.ToEnd(t, g, rng)
	if got := viewOf(t, g, other).Attempt.Tasks[id].Objective.Count; got != 2 {
		t.Fatalf("debrief count = %d, want the prediction 2", got)
	}
}

// Q2: the face-down tasks show only their difficulty before the debrief, then the whole card.
func TestDeepSeaFaceDownTasksOnTheWire(t *testing.T) {
	rng := gametest.NewRNG(3)
	g := gametest.NewDeepSeaGame(t, 3, 43, rng)
	check := func(debrief bool) {
		t.Helper()
		for seat := range 3 {
			for _, task := range viewOf(t, g, seat).Attempt.Tasks {
				if debrief != (task.CardID != 0 && task.Objective != nil) || task.Difficulty == 0 {
					t.Fatalf("seat %d task = %s (debrief %v)", seat, mustJSON(t, task), debrief)
				}
				if !debrief && !task.FaceDown {
					t.Fatalf("task not face down: %+v", task)
				}
			}
		}
	}
	check(false)
	gametest.ToEnd(t, g, rng)
	check(true)
}

func TestDeepSeaActionsAndStages(t *testing.T) {
	rng := gametest.NewRNG(2)
	g := gametest.NewDeepSeaGame(t, 3, 5, rng)
	cmd := g.Attempt.Commander
	v := viewOf(t, g, cmd).Attempt
	if !v.Actions.ReplaceTasks || v.Step != "draft" || v.Turn != cmd {
		t.Fatalf("captain at the start: %+v", v)
	}
	for seat := range 3 {
		if seat != cmd && viewOf(t, g, seat).Attempt.Actions.ReplaceTasks {
			t.Fatalf("seat %d may replace tasks", seat)
		}
	}
	if len(v.Takeable) == 0 {
		t.Fatalf("takeable = %v", v.Takeable)
	}
	// Q1 reports its stages.
	q1 := gametest.NewDeepSeaGame(t, 3, 42, gametest.NewRNG(1))
	if a := viewOf(t, q1, 0).Attempt; a.Stages != 2 || a.Stage != 0 {
		t.Fatalf("Q1 stages = %d/%d", a.Stage, a.Stages)
	}
	// Rapture of the deep: tokens left and the mode.
	r := gametest.NewDeepSeaGame(t, 5, 11, gametest.NewRNG(1))
	if a := viewOf(t, r, 0).Attempt; a.Comm != "rapture" || a.TokensLeft != 3 {
		t.Fatalf("rapture = %q %d", a.Comm, a.TokensLeft)
	}
	// Unfamiliar terrain shows the card and the mode it picked, never "unfamiliar".
	u := gametest.NewDeepSeaGame(t, 3, 20, gametest.NewRNG(1))
	if a := viewOf(t, u, 0).Attempt; a.Terrain == nil || a.Comm == "unfamiliar" {
		t.Fatalf("terrain = %v %q", a.Terrain, a.Comm)
	}
}
