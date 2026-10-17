package domain

import (
	"errors"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

// Rules tests of the Deep Sea setup and communication (deep-sea/rules.md §2–§5, §8).

// dssLobby returns a Deep Sea lobby with the first n of alice, bob, carol, dave and erin.
func dssLobby(t *testing.T, n int) *Game {
	t.Helper()
	g, err := NewGame("DEEP1", EditionDeepSea, alice, "Alice", Options{})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < n; i++ {
		id := crewIDs[i]
		must(t, g.AddPlayer(id, strings.ToUpper(string(id[:1]))+string(id[1:])))
	}
	g.PullEvents()
	return g
}

func dssRNG(seed uint64) RNG { return rand.New(rand.NewPCG(seed, 7)) }

// dssStart starts Deep Sea mission with n players (pending events cleared).
func dssStart(t *testing.T, n, mission int, seed uint64) *Game {
	t.Helper()
	g := dssLobby(t, n)
	must(t, g.SelectMission(alice, mission))
	must(t, g.StartMission(alice, dssRNG(seed)))
	g.PullEvents()
	return g
}

// dssTasks replaces the tasks by the given cards, all unassigned, and restarts the selection.
func dssTasks(g *Game, ids ...int) {
	a := g.Attempt
	a.Tasks = nil
	for i, id := range ids {
		a.Tasks = append(a.Tasks, newDeepSeaTask(i, id, len(a.Members), g.MissionDef().FaceDownTasks))
	}
	a.Passes, a.Chosen, a.Poll = 0, nil, nil
	a.Trick, a.Tricks = Trick{}, nil
	g.advanceSetup()
}

// dssID is the player acting for seat s.
func dssID(g *Game, s int) PlayerID {
	a := g.Attempt
	return a.Members[a.actorFor(s)].PlayerID
}

// dssActor is the first human the attempt waits for.
func dssActor(t *testing.T, g *Game) PlayerID {
	t.Helper()
	acts := g.Actors()
	if len(acts) == 0 {
		t.Fatalf("no actors in step %s", g.Attempt.Step)
	}
	return g.Attempt.Members[acts[0]].PlayerID
}

// owners returns the task owners in task order.
func dssOwners(g *Game) []int {
	var out []int
	for _, t := range g.Attempt.Tasks {
		out = append(out, t.Owner())
	}
	return out
}

func dssCards(g *Game) []int {
	var out []int
	for _, t := range g.Attempt.Tasks {
		out = append(out, t.CardID)
	}
	return out
}

// dssDraftRest takes the first takeable task (or passes) until the draft is over.
func dssDraftRest(t *testing.T, g *Game) {
	t.Helper()
	for g.Attempt.Step == StepDraft {
		if tk := g.Takeable(); len(tk) > 0 {
			must(t, g.PickTask(dssActor(t, g), tk[0]))
		} else {
			must(t, g.PassTask(dssActor(t, g)))
		}
	}
}

// dssPlayTrick plays a full trick, each seat its first legal card.
func dssPlayTrick(t *testing.T, g *Game) {
	t.Helper()
	for range g.Attempt.Members {
		turn := g.Attempt.turn()
		must(t, g.PlayCard(dssID(g, turn), g.LegalCards(turn)[0], noShuffle{}))
		if g.Phase != PhasePlaying {
			return
		}
	}
}

func TestDeepSeaDrawToDifficulty(t *testing.T) {
	for n := 2; n <= 5; n++ {
		for mission := 1; mission <= Missions(EditionDeepSea); mission++ {
			m, _ := MissionFor(EditionDeepSea, mission)
			if !m.DrawsTasks() {
				continue
			}
			for seed := uint64(1); seed <= 4; seed++ {
				g := dssStart(t, n, mission, seed)
				a := g.Attempt
				sum := 0
				for _, task := range a.Tasks {
					tc, _ := TaskCardFor(task.CardID)
					if task.Difficulty != tc.DifficultyFor(len(a.Members)) {
						t.Fatalf("m%d n%d: task %d difficulty %d", mission, n, task.CardID, task.Difficulty)
					}
					sum += task.Difficulty
					if !slices.Contains(g.UsedTasks, task.CardID) {
						t.Fatalf("m%d n%d: card %d not marked used", mission, n, task.CardID)
					}
				}
				if want := m.difficultyFor(len(a.Members)); sum != want {
					t.Fatalf("m%d n%d seed %d: difficulty %d, want %d", mission, n, seed, sum, want)
				}
				ids := dssCards(g)
				all := append(slices.Clone(ids), a.TaskDeck...)
				slices.Sort(all)
				if (!m.FaceDownTasks && len(slices.Compact(all)) != TaskCards) || (m.FaceDownTasks != (a.TaskDeck == nil)) {
					t.Fatalf("m%d n%d: tasks %v and deck (%d) are not the 96 cards", mission, n, ids, len(a.TaskDeck))
				}
			}
		}
	}
}

func TestDeepSeaFixedTasks(t *testing.T) {
	g := dssStart(t, 3, 32, 1)
	if got := dssCards(g); !slices.Equal(got, []int{74, 88, 85, 80}) {
		t.Fatalf("mission 32 tasks %v", got)
	}
	if g.Attempt.TaskDeck != nil || g.CanReplaceTasks(g.Attempt.Commander) == nil {
		t.Fatal("fixed tasks cannot be replaced")
	}
}

func TestDeepSeaUnusedCardsFirst(t *testing.T) {
	g := dssStart(t, 3, 31, 3)
	seen := map[int]bool{}
	for attempt := 0; attempt < 6; attempt++ {
		for _, id := range dssCards(g) {
			if seen[id] {
				t.Fatalf("attempt %d: card %d again before the deck ran out", attempt+1, id)
			}
			seen[id] = true
		}
		must(t, g.GiveUp(alice))
		must(t, g.NextAttempt(alice, NextAttemptChoice{NewTasks: true}, dssRNG(uint64(attempt))))
	}
}

func TestDeepSeaSameTasksAndUncounted(t *testing.T) {
	g := dssStart(t, 4, 5, 1)
	ids := dssCards(g)
	must(t, g.GiveUp(alice))
	must(t, g.NextAttempt(alice, NextAttemptChoice{}, dssRNG(2)))
	if got := dssCards(g); !slices.Equal(got, ids) || g.Attempt.Number != 2 || slices.Contains(dssOwners(g), 0) {
		t.Fatalf("same tasks: %v (want %v), attempt %d, owners %v", got, ids, g.Attempt.Number, dssOwners(g))
	}
	must(t, g.GiveUp(alice))
	if e := g.Logbook[5]; e.Attempts != 2 {
		t.Fatalf("logbook %+v", e)
	}
	must(t, g.NextAttempt(alice, NextAttemptChoice{NewTasks: true, Uncounted: true}, dssRNG(3)))
	if e := g.Logbook[5]; e.Attempts != 1 || g.Attempt.Number != 2 {
		t.Fatalf("uncounted: logbook %+v, attempt %d", e, g.Attempt.Number)
	}
	if slices.ContainsFunc(dssCards(g), func(id int) bool { return slices.Contains(ids, id) }) {
		t.Fatalf("new tasks %v share cards with %v", dssCards(g), ids)
	}
}

func TestDeepSeaChoicesNotInPlanetNine(t *testing.T) {
	g := startMission(t, 3, 1)
	must(t, g.GiveUp(alice))
	mustFail(t, g, ErrNotAllowed, func(g *Game) error {
		return g.NextAttempt(alice, NextAttemptChoice{Uncounted: true}, noShuffle{})
	})
}

func TestDeepSeaDraftPassing(t *testing.T) {
	g := dssStart(t, 4, 2, 1)
	dssTasks(g, 16, 17) // two tasks of difficulty 1, four members
	a := g.Attempt
	c := a.Commander
	if g.draftTurn() != c || g.CanPass() != nil {
		t.Fatalf("turn %d, pass %v", g.draftTurn(), g.CanPass())
	}
	must(t, g.PassTask(dssID(g, c)))
	must(t, g.PassTask(dssID(g, (c+1)%4)))
	// Two tasks left for the last two members: they must take them.
	if !errors.Is(g.CanPass(), ErrNotAllowed) {
		t.Fatalf("third pass allowed: %v", g.CanPass())
	}
	bbFailAny(t, g, func(g *Game) error { return g.PassTask(dssID(g, (c+2)%4)) }, ErrNotAllowed)
	must(t, g.PickTask(dssID(g, (c+2)%4), 0))
	must(t, g.PickTask(dssID(g, (c+3)%4), 1))
	if a.Step != StepPlay || !slices.Equal(dssOwners(g), []int{(c + 2) % 4, (c + 3) % 4}) {
		t.Fatalf("step %s owners %v", a.Step, dssOwners(g))
	}

	// As many tasks as members: no passing.
	g = dssStart(t, 3, 3, 1)
	dssTasks(g, 16, 17, 18)
	if !errors.Is(g.CanPass(), ErrNotAllowed) {
		t.Fatalf("pass with 3 tasks and 3 members: %v", g.CanPass())
	}
}

func TestDeepSeaDraftWithDummy(t *testing.T) {
	g := dssStart(t, 2, 3, 1)
	dssTasks(g, 16, 17, 18)
	a := g.Attempt
	dssDraftRest(t, g)
	got := dssOwners(g)
	slices.Sort(got)
	if !slices.Equal(got, []int{0, 1, 2}) || a.Step != StepPlay {
		t.Fatalf("owners %v (the dummy takes one, the captain acts for her), step %s", dssOwners(g), a.Step)
	}
}

func TestDeepSeaCaptainComparisonTasks(t *testing.T) {
	g := dssStart(t, 3, 5, 1)
	dssTasks(g, 4, 16, 17)
	a := g.Attempt
	c := a.Commander
	if tk := g.Takeable(); slices.Contains(tk, 0) {
		t.Fatalf("the captain may take card 4: %v", tk)
	}
	bbFailAny(t, g, func(g *Game) error { return g.PickTask(dssID(g, c), 0) }, ErrInvalidTask)

	// Only captain-comparison tasks: the captain passes, the others take them all.
	dssTasks(g, 4, 5, 6)
	dssDraftRest(t, g)
	if slices.Contains(dssOwners(g), c) || a.Step != StepPlay {
		t.Fatalf("owners %v, captain %d, step %s", dssOwners(g), c, a.Step)
	}
}

func TestDeepSeaHardestFirst(t *testing.T) {
	g := dssStart(t, 3, 19, 1)
	dssTasks(g, 16, 24, 27) // difficulties 1, 4, 2 with three members
	c := g.Attempt.Commander
	if tk := g.Takeable(); !slices.Equal(tk, []int{1}) {
		t.Fatalf("captain's first choice %v, want only the hardest", tk)
	}
	bbFailAny(t, g, func(g *Game) error { return g.PickTask(dssID(g, c), 0) }, ErrInvalidTask)
	must(t, g.PickTask(dssID(g, c), 1))
}

func TestDeepSeaCaptainSkipped(t *testing.T) {
	g := dssStart(t, 3, 25, 1)
	dssTasks(g, 16, 17, 18)
	a := g.Attempt
	c := a.Commander
	if g.draftTurn() != (c+1)%3 {
		t.Fatalf("first turn %d, want the captain's left %d", g.draftTurn(), (c+1)%3)
	}
	dssDraftRest(t, g)
	if slices.Contains(dssOwners(g), c) {
		t.Fatalf("the captain got a task: %v", dssOwners(g))
	}
}

func TestDeepSeaOneMember(t *testing.T) {
	g := dssStart(t, 3, 6, 1)
	dssTasks(g, 4, 16)
	a := g.Attempt
	c := a.Commander
	if a.Step != StepPick || g.Pick() != PickOneMember || slices.Contains(g.Pickable(), c) {
		t.Fatalf("step %s pick %s pickable %v", a.Step, g.Pick(), g.Pickable())
	}
	o := (c + 1) % 3
	must(t, g.PickSeats(dssID(g, c), []int{o}))
	if !slices.Equal(dssOwners(g), []int{o, o}) || !slices.Equal(a.Chosen, []int{o}) {
		t.Fatalf("owners %v chosen %v", dssOwners(g), a.Chosen)
	}
}

func TestDeepSeaVolunteer(t *testing.T) {
	g := dssStart(t, 3, 14, 1)
	dssTasks(g, 16, 17, 18)
	a := g.Attempt
	c := a.Commander
	if a.Step != StepPoll || a.Comm != CommCurrents {
		t.Fatalf("step %s comm %q", a.Step, a.Comm)
	}
	for _, s := range a.pollers() {
		must(t, g.AnswerPoll(dssID(g, s), s%2 == 0))
	}
	if a.Step != StepPick || g.Pick() != PickVolunteers {
		t.Fatalf("step %s pick %s", a.Step, g.Pick())
	}
	must(t, g.PickSeats(dssID(g, c), []int{c})) // the captain may take them themselves
	for _, o := range dssOwners(g) {
		if o != c {
			t.Fatalf("owners %v, want all %d", dssOwners(g), c)
		}
	}
}

func TestDeepSeaTwoVolunteers(t *testing.T) {
	g := dssStart(t, 4, 26, 1)
	dssTasks(g, 16, 17, 18)
	a := g.Attempt
	c := a.Commander
	for _, s := range a.pollers() {
		must(t, g.AnswerPoll(dssID(g, s), true))
	}
	bbFailAny(t, g, func(g *Game) error { return g.PickSeats(dssID(g, c), []int{c}) }, ErrInvalidSeat)
	bbFailAny(t, g, func(g *Game) error { return g.PickSeats(dssID(g, c), []int{c, c}) }, ErrInvalidSeat)
	o := (c + 2) % 4
	must(t, g.PickSeats(dssID(g, c), []int{o, c}))
	if a.Step != StepAssign || !slices.Equal(g.Assignable()[0], []int{min(c, o), max(c, o)}) {
		t.Fatalf("step %s assignable %v", a.Step, g.Assignable())
	}
	bbFailAny(t, g, func(g *Game) error { return g.AssignTask(dssID(g, c), 0, (c+1)%4) }, ErrInvalidSeat)
	must(t, g.AssignTask(dssID(g, c), 0, o))
	must(t, g.AssignTask(dssID(g, c), 1, o))
	// The captain has none yet: the last task must go to the captain.
	bbFailAny(t, g, func(g *Game) error { return g.AssignTask(dssID(g, c), 2, o) }, ErrUnevenDistribution)
	must(t, g.AssignTask(dssID(g, c), 2, c))
	if a.Step != StepPlay {
		t.Fatalf("step %s", a.Step)
	}
}

func TestDeepSeaCaptainOrVolunteer(t *testing.T) {
	for _, toCaptain := range []bool{false, true} {
		g := withDistress(dssStart(t, 3, 10, 1))
		dssTasks(g, 16, 17)
		a := g.Attempt
		c := a.Commander
		for _, s := range a.pollers() {
			must(t, g.AnswerPoll(dssID(g, s), true))
		}
		to := (c + 1) % 3
		if toCaptain {
			to = c
		}
		must(t, g.PickSeats(dssID(g, c), []int{to}))
		if a.Step != StepPlay || g.onlyBeforeFirstTrick() == toCaptain {
			t.Fatalf("to captain %v: step %s, first trick only %v", toCaptain, a.Step, g.onlyBeforeFirstTrick())
		}
		must(t, g.SkipDistress(dssID(g, c)))
		dssPlayTrick(t, g)
		if g.Phase != PhasePlaying {
			continue
		}
		for _, s := range a.humans() {
			err := g.CanCommunicate(s)
			if !toCaptain && !errors.Is(err, ErrCannotCommunicate) {
				t.Fatalf("seat %d may communicate after trick 1: %v", s, err)
			}
		}
	}
}

func TestDeepSeaFreeSelection(t *testing.T) {
	g := dssStart(t, 3, 17, 1)
	a := g.Attempt
	c := a.Commander
	if a.Step != StepAssign || g.Actors()[0] != c {
		t.Fatalf("step %s actors %v", a.Step, g.Actors())
	}
	for i := range a.Tasks {
		s := (c + 1 + i) % 3
		if a.Tasks[i].notForCaptain() && s == c {
			s = (c + 1) % 3
		}
		must(t, g.AssignTask(dssID(g, c), i, s))
	}
	if a.Step != StepPlay {
		t.Fatalf("step %s", a.Step)
	}
}

func TestDeepSeaPredictions(t *testing.T) {
	g := dssStart(t, 3, 5, 1)
	dssTasks(g, 91, 90) // hidden, open
	a := g.Attempt
	c, o := a.Commander, (a.Commander+1)%3
	must(t, g.PickTask(dssID(g, c), 0))
	must(t, g.PickTask(dssID(g, o), 1))
	if a.Step != StepPredict || !slices.Equal(g.Actors(), []int{c}) {
		t.Fatalf("step %s actors %v", a.Step, g.Actors())
	}
	bbFailAny(t, g, func(g *Game) error { return g.Predict(dssID(g, o), 0, 2) }, ErrNotYourTurn)
	bbFailAny(t, g, func(g *Game) error { return g.Predict(dssID(g, c), 0, 14) }, ErrInvalidPrediction)
	bbFailAny(t, g, func(g *Game) error { return g.Predict(dssID(g, c), 1, 2) }, ErrInvalidTask)
	g.PullEvents()
	must(t, g.Predict(dssID(g, c), 0, 3))
	if ev := eventsOf[Predicted](g); len(ev) != 1 || ev[0].Tricks != -1 {
		t.Fatalf("hidden prediction event %+v", ev)
	}
	must(t, g.Predict(dssID(g, o), 1, 4))
	if a.Step != StepPlay {
		t.Fatalf("step %s", a.Step)
	}
	mine, _ := g.ViewFor(dssID(g, c))
	theirs, _ := g.ViewFor(dssID(g, o))
	if mine.Attempt.Tasks[0].Objective.Count != 3 || theirs.Attempt.Tasks[0].Objective.Count != -1 ||
		!theirs.Attempt.Tasks[0].Predicted || theirs.Attempt.Tasks[1].Objective.Count != 4 {
		t.Fatalf("views: mine %+v theirs %+v / %+v", mine.Attempt.Tasks[0].Objective, theirs.Attempt.Tasks[0], theirs.Attempt.Tasks[1].Objective)
	}
	if a.Tasks[0].Objective.Count != 3 {
		t.Fatal("the state keeps the hidden prediction")
	}
}

func TestDeepSeaReplaceTask(t *testing.T) {
	g := dssStart(t, 3, 5, 1)
	a := g.Attempt
	c := a.Commander
	o := (c + 1) % 3
	bbFailAny(t, g, func(g *Game) error { return g.ReplaceTask(dssID(g, o), 0, noShuffle{}) }, ErrNotCommander)
	before := a.Tasks[0]
	must(t, g.ReplaceTask(dssID(g, c), 0, noShuffle{}))
	after := a.Tasks[0]
	if after.CardID == before.CardID || after.Difficulty != before.Difficulty || slices.Contains(dssCards(g)[1:], after.CardID) {
		t.Fatalf("replaced %+v by %+v", before, after)
	}
	if a.TaskDeck[len(a.TaskDeck)-1] != before.CardID || slices.Contains(a.TaskDeck, after.CardID) {
		t.Fatal("the replaced card goes under the deck")
	}
	dssDraftRest(t, g)
	bbFailAny(t, g, func(g *Game) error { return g.ReplaceTask(dssID(g, c), 0, noShuffle{}) }, ErrNotAllowed, ErrWrongPhase)
}

func TestDeepSeaRedealQ1(t *testing.T) {
	redealt := false
	for seed := uint64(1); seed <= 1500; seed++ {
		g := dssStart(t, 4, 42, seed)
		a := g.Attempt
		if a.needsRedeal(g.MissionDef()) {
			t.Fatalf("seed %d: a hand holds all four submarines", seed)
		}
		if !slices.Contains(a.Members[a.Commander].Cards(), Card{SuitTrump, 1}) {
			t.Fatalf("seed %d: the captain does not hold submarine 1", seed)
		}
		if o := a.Objectives[0]; o.Owner != a.Commander || o.Count != 2 {
			t.Fatalf("seed %d: stage A objective %+v", seed, o)
		}
		redealt = redealt || a.Redeals > 0
	}
	if !redealt {
		t.Fatal("no seed needed a re-deal")
	}
	// Two players: submarine 1 is set aside for a human, so the captain is never the dummy.
	for seed := uint64(1); seed <= 50; seed++ {
		g := dssStart(t, 2, 42, seed)
		if g.Attempt.Members[g.Attempt.Commander].Dummy {
			t.Fatalf("seed %d: the dummy is captain", seed)
		}
	}
}

func TestDeepSeaStagesQ1(t *testing.T) {
	g := dssStart(t, 3, 42, 1)
	won := &AttemptResult{Success: true, Reason: ReasonCompleted, Task: -1, Objective: -1, Seat: -1}
	g.endAttempt(won)
	if e := g.Logbook[42]; e.Completed || e.Attempts != 1 {
		t.Fatalf("after stage A: %+v", e)
	}
	mustFail(t, g, ErrInvalidResultChoice, func(g *Game) error { return g.NextMission(alice, noShuffle{}) })
	must(t, g.NextAttempt(alice, NextAttemptChoice{}, dssRNG(2)))
	if a := g.Attempt; a.Stage != 1 || a.Number != 2 || a.Objectives[0].Count != 3 {
		t.Fatalf("stage B: stage %d attempt %d objective %+v", a.Stage, a.Number, a.Objectives[0])
	}
	g.endAttempt(&AttemptResult{Reason: ReasonGaveUp, Task: -1, Objective: -1, Seat: 0})
	must(t, g.NextAttempt(alice, NextAttemptChoice{}, dssRNG(3)))
	if g.Attempt.Stage != 1 {
		t.Fatal("a lost stage B is played again")
	}
	g.endAttempt(won)
	if e := g.Logbook[42]; !e.Completed || e.Attempts != 3 {
		t.Fatalf("after stage B: %+v", e)
	}
	mustFail(t, g, ErrInvalidResultChoice, func(g *Game) error { return g.NextAttempt(alice, NextAttemptChoice{}, noShuffle{}) })
}

func TestDeepSeaEpilogueUncounted(t *testing.T) {
	g := dssStart(t, 3, 33, 1)
	must(t, g.GiveUp(alice))
	if e := g.Logbook[33]; e.Attempts != 0 {
		t.Fatalf("epilogue logbook %+v", e)
	}
}

func TestDeepSeaFaceDownQ2(t *testing.T) {
	g := dssStart(t, 4, 43, 1)
	a := g.Attempt
	sum := 0
	for _, task := range a.Tasks {
		sum += task.Difficulty
		if !task.FaceDown {
			t.Fatal("Q2 tasks are face down")
		}
	}
	if sum != 10 || g.CanReplaceTasks(a.Commander) == nil {
		t.Fatalf("difficulty %d, replace allowed %v", sum, g.CanReplaceTasks(a.Commander) == nil)
	}
	v, _ := g.ViewFor(bob)
	for _, tv := range v.Attempt.Tasks {
		if tv.CardID != 0 || tv.Objective != nil || tv.Difficulty == 0 {
			t.Fatalf("face-down task view %+v", tv)
		}
	}
	dssDraftRest(t, g)
	targets := a.Objectives[0].Targets
	total := 0
	for _, n := range targets {
		total += n
	}
	if total != 10 || len(targets) != 4 {
		t.Fatalf("quotas %v", targets)
	}
}

func TestDeepSeaTerrain(t *testing.T) {
	for r := Rank(1); r <= 9; r++ {
		want := CommNormal
		if r >= 4 {
			want = CommCurrents
		}
		if r >= 7 {
			want = CommRapture
		}
		if got := terrainMode(Card{SuitBlue, r}); got != want {
			t.Fatalf("terrain %d: %q, want %q", r, got, want)
		}
	}
	for seed := uint64(1); seed <= 20; seed++ {
		g := dssStart(t, 3, 20, seed)
		a := g.Attempt
		if a.Terrain == nil || a.Terrain.IsTrump() || a.Comm != terrainMode(*a.Terrain) {
			t.Fatalf("seed %d: terrain %v comm %q", seed, a.Terrain, a.Comm)
		}
	}
}

// dssPlayReady skips the selection: every task to seat owner, hands as given, play begins.
func dssPlayReady(g *Game, owner int, hands ...string) {
	setHands(g, hands...)
	a := g.Attempt
	for i := range a.Tasks {
		a.Tasks[i].Objective.Owner = owner
	}
	toPlay(g)
}

func TestDeepSeaRapture(t *testing.T) {
	g := dssStart(t, 4, 11, 1)
	dssPlayReady(g, 1,
		"t4 p1 p5 p9 b1 b2 b3 b4 b5 b6",
		"t1 t2 t3 p2 p3 p4 b7 b8 b9 g1",
		"p6 p7 p8 g2 g3 g4 g5 g6 g7 g8",
		"g9 y1 y2 y3 y4 y5 y6 y7 y8 y9")
	a := g.Attempt
	if a.SonarTokens() != 2 || a.TokensLeft() != 2 {
		t.Fatalf("tokens %d left %d", a.SonarTokens(), a.TokensLeft())
	}
	must(t, g.Communicate(alice, c("p1"), PosLowest))
	bbFailAny(t, g, func(g *Game) error { return g.Communicate(alice, c("p1"), PosLowest) }, ErrCannotCommunicate)
	must(t, g.Communicate(alice, c("p9"), PosHighest)) // the same member again
	if a.TokensLeft() != 0 || len(a.CommunicatedCards(0)) != 2 {
		t.Fatalf("left %d, shown %v", a.TokensLeft(), a.CommunicatedCards(0))
	}
	bbFailAny(t, g, func(g *Game) error { return g.Communicate(dave, c("y9"), PosHighest) }, ErrCannotCommunicate)
	v, _ := g.ViewFor(bob)
	if len(v.Attempt.Members[0].Communications) != 2 || v.Attempt.TokensLeft != 0 {
		t.Fatalf("view %+v", v.Attempt.Members[0])
	}
}

func TestDeepSeaNoCommunication(t *testing.T) {
	g := dssStart(t, 3, 16, 1)
	dssPlayReady(g, 1, "t4 p1 p2 p3 p4 p5 p6 p7 p8 p9 b1 b2 b3 b4", "", "")
	if g.Attempt.Comm != CommNone || !errors.Is(g.CanCommunicate(0), ErrCannotCommunicate) {
		t.Fatalf("comm %q: %v", g.Attempt.Comm, g.CanCommunicate(0))
	}
}

func TestDeepSeaCurrentsHidePosition(t *testing.T) {
	g := dssStart(t, 3, 9, 1)
	dssPlayReady(g, 1, "t4 p1 p2 p3 p4 p5 p6 p7 p8 p9 b1 b2 b3 b4", "", "")
	must(t, g.Communicate(alice, c("p9"), PosHighest))
	if cm := g.Attempt.Comms[0][0]; cm.Position != PosHidden {
		t.Fatalf("currents: %+v", cm)
	}
}
