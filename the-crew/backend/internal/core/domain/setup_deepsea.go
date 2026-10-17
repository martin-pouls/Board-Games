package domain

import (
	"fmt"
	"slices"
)

// The Deep Sea parts of an attempt's setup (deep-sea/rules.md §2–§4, §8): the unfamiliar-terrain
// draw, task cards by difficulty with the used-card pile, the automatic re-deal, replacing a
// task, the draft with passing, the captain's choices (one member, volunteers, free assignment)
// and the predictions. The order of the setup items is in setup.go.

// maxRedeals bounds the automatic re-deal; an alignment that forces a re-deal is rare.
const maxRedeals = 1000

// drawTerrain draws the unfamiliar-terrain card: a random colour card (rules §4; a submarine is
// redrawn, so the draw is among the colour cards).
func drawTerrain(rng RNG) Card {
	cs := ColourCards()
	return cs[rng.IntN(len(cs))]
}

// terrainMode is the communication mode the terrain card picks: 1–3 normal, 4–6 currents, 7–9
// rapture.
func terrainMode(c Card) CommMode {
	switch {
	case c.Rank <= 3:
		return CommNormal
	case c.Rank <= 6:
		return CommCurrents
	}
	return CommRapture
}

// newDeepSeaTask builds task id from task card card for the crew size. A face-down task (Q2)
// has no condition of its own: only its difficulty counts.
func newDeepSeaTask(id, card, members int, faceDown bool) Task {
	tc, _ := TaskCardFor(card)
	o := tc.objectiveFor(members)
	if faceDown {
		o = Objective{Kind: KindNone, Owner: OwnerUnassigned}
	}
	return Task{ID: id, Objective: o, FaceDown: faceDown, CardID: card, Difficulty: tc.DifficultyFor(members)}
}

// notForCaptain reports whether the captain may not own task t (cards 4–6).
func (t Task) notForCaptain() bool { return t.Objective.Kind == KindCmpCaptain }

// taskDeck returns the task cards not in exclude in draw order: the unused ones shuffled, then
// the used ones shuffled (rules §2 App decision).
func (g *Game) taskDeck(exclude []int, rng RNG) []int {
	var fresh, used []int
	for id := 1; id <= TaskCards; id++ {
		switch {
		case slices.Contains(exclude, id):
		case slices.Contains(g.UsedTasks, id):
			used = append(used, id)
		default:
			fresh = append(fresh, id)
		}
	}
	for _, s := range [][]int{fresh, used} {
		rng.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
	}
	return append(fresh, used...)
}

// markUsed adds ids to the used task cards; once every card was used, the pile starts again.
func (g *Game) markUsed(ids []int) {
	for _, id := range ids {
		if !slices.Contains(g.UsedTasks, id) {
			g.UsedTasks = append(g.UsedTasks, id)
		}
	}
	if len(g.UsedTasks) >= TaskCards {
		g.UsedTasks = slices.Clone(ids)
	}
	slices.Sort(g.UsedTasks)
}

// dealTaskCards lays out the Deep Sea tasks of the attempt: the mission's fixed tasks, the kept
// ones of the last attempt, or cards drawn until the total difficulty is reached exactly (a card
// that would overshoot is skipped).
func (g *Game) dealTaskCards(a *Attempt, m MissionDef, keep []int, rng RNG) {
	n := len(a.Members)
	ids := slices.Clone(m.FixedTasks)
	if ids == nil {
		ids = slices.Clone(keep)
		deck := g.taskDeck(ids, rng)
		if ids == nil {
			target, sum := m.difficultyFor(n), 0
			rest := []int{}
			for _, id := range deck {
				tc, _ := TaskCardFor(id)
				if d := tc.DifficultyFor(n); sum+d <= target && sum < target {
					ids = append(ids, id)
					sum += d
				} else {
					rest = append(rest, id)
				}
			}
			deck = rest
		}
		if !m.FaceDownTasks {
			a.TaskDeck = deck
		}
		g.markUsed(ids)
	}
	for i, id := range ids {
		a.Tasks = append(a.Tasks, newDeepSeaTask(i, id, n, m.FaceDownTasks))
	}
}

// redealRules returns the alignments that force a re-deal: the mission's and those of the tasks.
func (a *Attempt) redealRules(m MissionDef) [][]Card {
	rules := slices.Clone(m.Redeal)
	for _, t := range a.Tasks {
		if tc, ok := TaskCardFor(t.CardID); ok {
			rules = append(rules, tc.Redeal...)
		}
	}
	return rules
}

// needsRedeal reports whether one hand holds all cards of a re-deal rule (rules §2, §6).
func (a *Attempt) needsRedeal(m MissionDef) bool {
	rules := a.redealRules(m)
	for i := range a.Members {
		held := a.Members[i].Cards()
		for _, r := range rules {
			if !slices.ContainsFunc(r, func(c Card) bool { return !slices.Contains(held, c) }) {
				return true
			}
		}
	}
	return false
}

// redealIfNeeded deals the playing cards again while an alignment makes a task impossible. It is
// not an attempt. Emits Redealt.
func (g *Game) redealIfNeeded(m MissionDef, rng RNG) {
	a := g.Attempt
	n := 0
	for ; n < maxRedeals && a.needsRedeal(m); n++ {
		a.Members = deal(g.Players, m.captainCard(), rng)
		a.setCommander(m)
	}
	if n > 0 {
		a.Redeals += n
		a.resolveObjectives(m)
		g.emit(Redealt{Times: n})
	}
}

// setCommander makes the holder of the mission's captain card the commander.
func (a *Attempt) setCommander(m MissionDef) {
	for i, mem := range a.Members {
		if mem.holds(m.captainCard()) || (mem.Dummy && slices.Contains(mem.Cards(), m.captainCard())) {
			a.Commander = i
		}
	}
}

// --- Replacing a task (rules §3 App decision) ---

// CanReplaceTasks reports why seat may not replace a displayed task now, or nil: the captain,
// before the first choice of the selection, in a mission whose tasks are drawn face up.
func (g *Game) CanReplaceTasks(seat int) error {
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: no attempt is running", ErrWrongPhase)
	}
	a := g.Attempt
	if err := a.requireCommander(seat); err != nil {
		return err
	}
	if a.TaskDeck == nil || !a.selectionUntouched() {
		return fmt.Errorf("%w: only before the first choice of the selection", ErrNotAllowed)
	}
	return nil
}

// selectionUntouched reports whether the selection has not started: no task has an owner, nobody
// passed, chose or answered.
func (a *Attempt) selectionUntouched() bool {
	switch a.Step {
	case StepDraft, StepPoll, StepPick, StepAssign:
	default:
		return false
	}
	if a.Passes > 0 || a.Chosen != nil || slices.ContainsFunc(a.Tasks, func(t Task) bool { return t.Owner() != OwnerUnassigned }) {
		return false
	}
	return a.Poll == nil || !slices.ContainsFunc(a.Poll.Answers, func(b *bool) bool { return b != nil })
}

// ReplaceTask puts task taskID back under the deck and lays out the next card of the same
// difficulty in its place (captain). A new alignment may force a re-deal (rng).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotCommander, ErrNotAllowed, ErrInvalidTask.
// Emits TaskReplaced (Redealt).
func (g *Game) ReplaceTask(by PlayerID, taskID int, rng RNG) error {
	seat, err := g.requireSeated(by)
	if err != nil {
		return err
	}
	if err := g.CanReplaceTasks(seat); err != nil {
		return err
	}
	a := g.Attempt
	if !a.validTask(taskID) {
		return fmt.Errorf("%w: %d", ErrInvalidTask, taskID)
	}
	n := len(a.Members)
	old := a.Tasks[taskID]
	k := slices.IndexFunc(a.TaskDeck, func(id int) bool {
		tc, _ := TaskCardFor(id)
		return tc.DifficultyFor(n) == old.Difficulty
	})
	if k < 0 {
		return fmt.Errorf("%w: no other task card of difficulty %d", ErrNotAllowed, old.Difficulty)
	}
	id := a.TaskDeck[k]
	a.TaskDeck = append(slices.Delete(slices.Clone(a.TaskDeck), k, k+1), old.CardID)
	a.Tasks[taskID] = newDeepSeaTask(taskID, id, n, false)
	g.markUsed([]int{id})
	g.emit(TaskReplaced{TaskID: taskID})
	m := g.MissionDef()
	g.redealIfNeeded(m, rng)
	a.Poll = nil
	g.advanceSetup()
	g.commit()
	return nil
}

// --- The draft with passing (rules §3) ---

// drafters are the seats taking part in the draft, clockwise from the commander (25: without
// the commander).
func (a *Attempt) drafters(m MissionDef) []int {
	var out []int
	for i := range a.Members {
		s := (a.Commander + i) % len(a.Members)
		if s == a.Commander && m.CaptainSkipped {
			continue
		}
		out = append(out, s)
	}
	return out
}

// tasksTaken is the number of tasks with an owner.
func (a *Attempt) tasksTaken() int {
	n := 0
	for _, t := range a.Tasks {
		if t.Owner() != OwnerUnassigned {
			n++
		}
	}
	return n
}

// draftTurn is the seat whose turn it is to take a task in StepDraft.
func (g *Game) draftTurn() int {
	a := g.Attempt
	m := g.MissionDef()
	if m.Distribution != DistDraft {
		return a.DraftTurn()
	}
	ds := a.drafters(m)
	return ds[(a.tasksTaken()+a.Passes)%len(ds)]
}

// takeError reports why seat may not take task id in the draft, or nil.
func (a *Attempt) takeError(m MissionDef, seat, id int) error {
	if !a.validTask(id) || a.Tasks[id].Owner() != OwnerUnassigned {
		return fmt.Errorf("%w: %d", ErrInvalidTask, id)
	}
	t := a.Tasks[id]
	if seat == a.Commander && t.notForCaptain() {
		return fmt.Errorf("%w: the captain may not take this task", ErrInvalidTask)
	}
	if m.HardestFirst && seat == a.Commander && !slices.ContainsFunc(a.Tasks, func(t Task) bool { return t.Owner() == seat }) {
		top := 0
		for _, o := range a.Tasks {
			if o.Owner() == OwnerUnassigned && !o.notForCaptain() {
				top = max(top, o.Difficulty)
			}
		}
		if t.Difficulty < top {
			return fmt.Errorf("%w: the captain's first task must have the highest difficulty", ErrInvalidTask)
		}
	}
	return nil
}

// Takeable returns the tasks the seat whose turn it is may take in StepDraft (never nil).
func (g *Game) Takeable() []int {
	out := []int{}
	a := g.Attempt
	if g.Phase != PhasePlaying || a.Step != StepDraft {
		return out
	}
	m, turn := g.MissionDef(), g.draftTurn()
	for i := range a.Tasks {
		if a.takeError(m, turn, i) == nil {
			out = append(out, i)
		}
	}
	return out
}

// CanPass reports why the seat whose turn it is may not pass in the draft, or nil. Passing is
// allowed only if at the start there were fewer tasks than drafters, and only while the members
// after this one in the round can still take the remaining tasks; a member who may take no task
// passes in any case.
func (g *Game) CanPass() error {
	if g.Phase != PhasePlaying || g.Attempt.Step != StepDraft {
		return fmt.Errorf("%w: only in the draft", ErrWrongStep)
	}
	a := g.Attempt
	m := g.MissionDef()
	if m.Distribution != DistDraft {
		return fmt.Errorf("%w: no passing in this mission", ErrNotAllowed)
	}
	if len(g.Takeable()) == 0 {
		return nil
	}
	ds := a.drafters(m)
	pos := a.tasksTaken() + a.Passes
	switch {
	case len(a.Tasks) >= len(ds):
		return fmt.Errorf("%w: passing needs fewer tasks than members", ErrNotAllowed)
	case pos >= len(ds) || len(a.Tasks)-a.tasksTaken() > len(ds)-pos-1:
		return fmt.Errorf("%w: every task must be taken in this round", ErrNotAllowed)
	}
	return nil
}

// PassTask passes in the draft, for the seat whose turn it is (the captain acts for the dummy).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn, ErrNotAllowed.
// Emits TaskPassed.
func (g *Game) PassTask(by PlayerID) error {
	seat, err := g.requireStep(by, StepDraft)
	if err != nil {
		return err
	}
	a := g.Attempt
	turn := g.draftTurn()
	if a.actorFor(turn) != seat {
		return fmt.Errorf("%w: seat %d picks", ErrNotYourTurn, turn)
	}
	if err := g.CanPass(); err != nil {
		return err
	}
	a.Passes++
	g.emit(TaskPassed{Seat: turn})
	g.commit()
	return nil
}

// --- The captain's choices (6, 10, 13, 14–16, 26, free selection) ---

// chooseError reports why seat may not be chosen to take every task (6, 10, 13, 14–16) or, with
// two volunteers (26), a share of them.
func (a *Attempt) chooseError(m MissionDef, s int) error {
	if s != a.Commander {
		return nil
	}
	forbidden := 0
	for _, t := range a.Tasks {
		if t.notForCaptain() {
			forbidden++
		}
	}
	if forbidden > 0 && (m.Volunteers < 2 || forbidden == len(a.Tasks)) {
		return fmt.Errorf("%w: the captain may not take one of the tasks", ErrInvalidSeat)
	}
	return nil
}

// giveAll gives every task to seat s.
func (g *Game) giveAll(s int) {
	a := g.Attempt
	for i := range a.Tasks {
		a.Tasks[i].Objective.Owner = s
		g.emit(TaskPicked{TaskID: i, Seat: s})
	}
}

// assignError reports why the captain may not give task id to seat s in StepAssign, or nil. With
// volunteers (26) only they receive tasks, each at least one.
func (a *Attempt) assignError(id, s int) error {
	if !a.validTask(id) || a.Tasks[id].Owner() != OwnerUnassigned {
		return fmt.Errorf("%w: %d", ErrInvalidTask, id)
	}
	if !a.validSeat(s) {
		return fmt.Errorf("%w: %d", ErrInvalidSeat, s)
	}
	if s == a.Commander && a.Tasks[id].notForCaptain() {
		return fmt.Errorf("%w: the captain may not take this task", ErrInvalidSeat)
	}
	if a.Chosen == nil {
		return nil
	}
	if !slices.Contains(a.Chosen, s) {
		return fmt.Errorf("%w: only the volunteers receive tasks", ErrInvalidSeat)
	}
	var left []Task // still unassigned after this one
	for i, t := range a.Tasks {
		if i != id && t.Owner() == OwnerUnassigned {
			left = append(left, t)
		}
	}
	for _, v := range a.Chosen {
		if v == s || slices.ContainsFunc(a.Tasks, func(t Task) bool { return t.Owner() == v }) {
			continue
		}
		// v still needs a task: one must be left that v may take.
		ok := slices.ContainsFunc(left, func(t Task) bool { return v != a.Commander || !t.notForCaptain() })
		if !ok || len(left) < a.emptyVolunteers(s) {
			return fmt.Errorf("%w: each volunteer takes at least one task", ErrUnevenDistribution)
		}
	}
	return nil
}

// emptyVolunteers is the number of chosen members other than s without a task.
func (a *Attempt) emptyVolunteers(s int) int {
	n := 0
	for _, v := range a.Chosen {
		if v != s && !slices.ContainsFunc(a.Tasks, func(t Task) bool { return t.Owner() == v }) {
			n++
		}
	}
	return n
}

// Assignable returns, per task, the seats the captain may give it to now (StepAssign; else
// empty lists). Never nil.
func (g *Game) Assignable() [][]int {
	a := g.Attempt
	out := make([][]int, len(a.Tasks))
	for i := range a.Tasks {
		out[i] = []int{}
		if g.Phase != PhasePlaying || a.Step != StepAssign {
			continue
		}
		for s := range a.Members {
			if a.assignError(i, s) == nil {
				out[i] = append(out[i], s)
			}
		}
	}
	return out
}

// AssignTask gives task taskID to seat to (captain, StepAssign: free selection, or the split
// between two volunteers).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotCommander, ErrInvalidTask,
// ErrInvalidSeat, ErrUnevenDistribution. Emits TaskPicked.
func (g *Game) AssignTask(by PlayerID, taskID, to int) error {
	seat, err := g.requireStep(by, StepAssign)
	if err != nil {
		return err
	}
	a := g.Attempt
	if err := a.requireCommander(seat); err != nil {
		return err
	}
	if err := a.assignError(taskID, to); err != nil {
		return err
	}
	a.Tasks[taskID].Objective.Owner = to
	g.emit(TaskPicked{TaskID: taskID, Seat: to})
	g.advanceSetup()
	g.commit()
	return nil
}

// --- Predictions (cards 90, 91) ---

// Predict notes the number of tricks the owner of prediction task taskID will win (its owner, the
// captain for the dummy), 0 to the number of tricks.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn, ErrInvalidTask,
// ErrInvalidPrediction. Emits Predicted (without the number for a hidden prediction).
func (g *Game) Predict(by PlayerID, taskID, tricks int) error {
	seat, err := g.requireStep(by, StepPredict)
	if err != nil {
		return err
	}
	a := g.Attempt
	it := g.pendingItem()
	if taskID != it.index {
		return fmt.Errorf("%w: task %d is predicted now", ErrInvalidTask, it.index)
	}
	t := &a.Tasks[taskID]
	if a.actorFor(t.Owner()) != seat {
		return fmt.Errorf("%w: seat %d predicts", ErrNotYourTurn, t.Owner())
	}
	if tricks < 0 || tricks > a.history().Total {
		return fmt.Errorf("%w: %d", ErrInvalidPrediction, tricks)
	}
	t.Objective.Count = tricks
	e := Predicted{TaskID: taskID, Seat: t.Owner(), Tricks: tricks}
	if t.Objective.Hidden {
		e.Tricks = -1
	}
	g.emit(e)
	g.advanceSetup()
	g.commit()
	return nil
}

// setQuotas fills the targets of Q2's trick quota: each seat's tasks' total difficulty.
func (a *Attempt) setQuotas() {
	for i, o := range a.Objectives {
		if o.Kind != KindTrickQuota {
			continue
		}
		targets := make([]int, len(a.Members))
		for _, t := range a.Tasks {
			if s := t.Owner(); s >= 0 {
				targets[s] += t.Difficulty
			}
		}
		a.Objectives[i].Targets = targets
	}
}

// --- Communication modes (rules §4) ---

// deadZone reports whether communications show no position (Planet Nine's dead zone, Deep Sea's
// currents).
func (g *Game) deadZone() bool {
	return g.MissionDef().Comm.DeadZone || g.Attempt.Comm == CommCurrents
}

// SonarTokens is the number of shared tokens of rapture of the deep: the crew size (the dummy
// included) minus 2.
func (a *Attempt) SonarTokens() int { return len(a.Members) - 2 }

// TokensLeft is the number of shared sonar tokens left in rapture of the deep, else -1.
func (a *Attempt) TokensLeft() int {
	if a.Comm != CommRapture {
		return -1
	}
	n := a.SonarTokens()
	for _, cs := range a.Comms {
		n -= len(cs)
	}
	return max(n, 0)
}

// onlyBeforeFirstTrick reports whether communication ends with the first trick (10, 13: the tasks
// went to another member than the captain).
func (g *Game) onlyBeforeFirstTrick() bool {
	a := g.Attempt
	return g.MissionDef().Distribution == DistCaptainOrVolunteer && len(a.Chosen) == 1 && a.Chosen[0] != a.Commander
}

// hasNextStage reports whether the mission has a stage after the attempt's (Q1).
func (a *Attempt) hasNextStage(m MissionDef) bool { return a.Stage+1 < len(m.Stages) }

// atStage returns m with the objectives of stage appended (Q1).
func (m MissionDef) atStage(stage int) MissionDef {
	if stage < len(m.Stages) {
		m.Objectives = append(slices.Clone(m.Objectives), m.Stages[stage]...)
	}
	return m
}
