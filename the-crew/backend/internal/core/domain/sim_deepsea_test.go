package domain

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime/debug"
	"slices"
	"strings"
	"testing"
	"time"
)

// The Deep Sea simulation extends sim_test.go: random legal actions, drawn only from what the
// views and Actors advertise, through every Deep Sea mission (1-41) with 2-5 players, plus
// illegal calls that must be rejected without a change. After every action the invariants of
// sim_test.go hold where they apply, and the Deep Sea ones below: ownership of the tasks, the
// difficulty, hiding of predictions, the communication modes, the logbook and the Clone.

const dsMaxActions = 1500

func newDeepSeaCrew(t *testing.T, n int) *Game {
	t.Helper()
	g, err := NewGame("GAME1", EditionDeepSea, alice, "Alice", Options{})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < n; i++ {
		id := crewIDs[i]
		if err := g.AddPlayer(id, strings.ToUpper(string(id[:1]))+string(id[1:])); err != nil {
			t.Fatal(err)
		}
	}
	g.PullEvents()
	return g
}

func (s *simRun) bump(key string) {
	s.st.count(func() { s.st.followUp[key]++ })
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// expectErr runs a mutator that must be rejected (with want, if set), leaving the state unchanged.
func (s *simRun) expectErr(name string, want error, fn func() error) {
	s.t.Helper()
	err := s.try(name, fn)
	if err == nil {
		s.failf("%s succeeded, want an error", name)
	}
	if want != nil && !errors.Is(err, want) {
		s.failf("%s: error %v, want %v", name, err, want)
	}
}

// otherHuman returns a human seat other than seat, or -1.
func (s *simRun) otherHuman(seat int) int {
	var out []int
	for _, h := range s.humanSeats() {
		if h != seat {
			out = append(out, h)
		}
	}
	if len(out) == 0 {
		return -1
	}
	return out[s.rng.IntN(len(out))]
}

// ---- invariants ----

func (s *simRun) checkEventDeepSea(name string, e Event) {
	s.t.Helper()
	a := s.g.Attempt
	switch ev := e.(type) {
	case Predicted:
		if ev.TaskID < 0 || ev.TaskID >= len(a.Tasks) || a.Tasks[ev.TaskID].Objective.Kind != KindPredict {
			s.failf("%s: event %+v: not a prediction task", name, ev)
		}
		t := a.Tasks[ev.TaskID]
		if ev.Seat != t.Owner() {
			s.failf("%s: event %+v: owner is %d", name, ev, t.Owner())
		}
		if t.Objective.Hidden && ev.Tricks != -1 {
			s.failf("%s: event %+v reveals a hidden prediction", name, ev)
		}
		if !t.Objective.Hidden && ev.Tricks != t.Objective.Count {
			s.failf("%s: event %+v: open prediction is %d", name, ev, t.Objective.Count)
		}
	case Communicated:
		if a.Comm == CommNone {
			s.failf("%s: event %+v: communication in a mission without", name, ev)
		}
		if a.Comm == CommCurrents {
			if ev.Position != PosHidden {
				s.failf("%s: event %+v: position shown in the currents", name, ev)
			}
		} else if truth, ok := StatementFor(a.Members[ev.Seat].Hand, ev.Card); !ok || ev.Position != truth {
			s.failf("%s: event %+v: statement is not true (%v, %v)", name, ev, truth, ok)
		}
	case TaskReplaced:
		if ev.TaskID < 0 || ev.TaskID >= len(a.Tasks) {
			s.failf("%s: event %+v: no such task", name, ev)
		}
	case TaskPicked:
		if ev.TaskID < 0 || ev.TaskID >= len(a.Tasks) || a.Tasks[ev.TaskID].Owner() != ev.Seat {
			s.failf("%s: event %+v: owner is not the seat", name, ev)
		}
	}
}

func (s *simRun) checkTaskViewDeepSea(name string, seat int, t Task, tv TaskView) {
	s.t.Helper()
	a := s.g.Attempt
	over := a.Result != nil
	if tv.ID != t.ID || tv.Owner != t.Owner() || tv.Difficulty != t.Difficulty || tv.FaceDown || tv.Card != nil {
		s.failf("%s: view of seat %d: task %d is %+v", name, seat, t.ID, tv)
	}
	if tv.CardID != t.CardID || tv.Objective == nil {
		s.failf("%s: view of seat %d: task %d card %d / %+v, want card %d", name, seat, t.ID, tv.CardID, tv.Objective, t.CardID)
	}
	o := *tv.Objective
	real := t.Objective
	if real.Kind != KindPredict {
		if !reflect.DeepEqual(o, real) {
			s.failf("%s: view of seat %d: task %d objective %+v, want %+v", name, seat, t.ID, o, real)
		}
		return
	}
	if o.Kind != KindPredict || tv.Predicted != (real.Count >= 0) {
		s.failf("%s: view of seat %d: prediction task %d: %+v predicted=%v", name, seat, t.ID, o, tv.Predicted)
	}
	// Hidden predictions: shown only to the owner (the captain for the dummy) before the debrief.
	owner := t.Owner()
	mine := owner >= 0 && (owner == seat || (a.Members[owner].Dummy && seat == a.Commander))
	want := real.Count
	if real.Hidden && !over && !mine {
		want = -1
	}
	if o.Count != want {
		s.failf("%s: view of seat %d: prediction task %d shows count %d, want %d (hidden=%v)", name, seat, t.ID, o.Count, want, real.Hidden)
	}
	o.Count, real.Count = 0, 0
	if !reflect.DeepEqual(o, real) {
		s.failf("%s: view of seat %d: prediction task %d objective %+v, want %+v", name, seat, t.ID, o, real)
	}
}

func (s *simRun) checkViewDeepSea(name string, seat int, v PlayerView) {
	s.t.Helper()
	g, a := s.g, s.g.Attempt
	av := v.Attempt
	over := a.Result != nil
	mdef := g.MissionDef()
	if av.Comm != a.Comm || av.TokensLeft != a.TokensLeft() || !sameCard(av.Terrain, a.Terrain) || !slices.Equal(av.Chosen, a.Chosen) {
		s.failf("%s: view of seat %d: comm %q tokens %d terrain %v chosen %v", name, seat, av.Comm, av.TokensLeft, av.Terrain, av.Chosen)
	}
	if (a.Comm == CommRapture) != (av.TokensLeft >= 0) {
		s.failf("%s: view of seat %d: tokens left %d in mode %q", name, seat, av.TokensLeft, a.Comm)
	}
	if av.Redeals != a.Redeals {
		s.failf("%s: view of seat %d: redeals %d", name, seat, av.Redeals)
	}
	first := g.Mission == 10 || g.Mission == 13
	wantFirst := first && len(a.Chosen) == 1 && a.Chosen[0] != a.Commander
	if av.CommFirstTrickOnly != wantFirst {
		s.failf("%s: view of seat %d: CommFirstTrickOnly %v, want %v", name, seat, av.CommFirstTrickOnly, wantFirst)
	}

	// Who sees which options.
	turnIsMe := false
	if a.Step == StepDraft {
		turnIsMe = a.actorFor(g.draftTurn()) == seat
		if av.Turn != g.draftTurn() {
			s.failf("%s: view of seat %d: turn %d, draft turn %d", name, seat, av.Turn, g.draftTurn())
		}
	}
	if turnIsMe {
		if !slices.Equal(av.Takeable, g.Takeable()) {
			s.failf("%s: view of seat %d: takeable %v, want %v", name, seat, av.Takeable, g.Takeable())
		}
	} else if len(av.Takeable) != 0 {
		s.failf("%s: view of seat %d outside its draft turn offers %v", name, seat, av.Takeable)
	}
	if want := !over && turnIsMe && g.CanPass() == nil; av.Actions.Pass != want {
		s.failf("%s: view of seat %d: Actions.Pass %v, want %v", name, seat, av.Actions.Pass, want)
	}
	if want := !over && seat == a.Commander && g.CanReplaceTasks(seat) == nil; av.Actions.ReplaceTasks != want {
		s.failf("%s: view of seat %d: Actions.ReplaceTasks %v, want %v", name, seat, av.Actions.ReplaceTasks, want)
	}
	if av.Actions.ReplaceTasks && (a.TaskDeck == nil || mdef.FixedTasks != nil) {
		s.failf("%s: view of seat %d may replace tasks of mission %d", name, seat, g.Mission)
	}
	if seat != a.Commander {
		if len(av.Pickable) != 0 {
			s.failf("%s: view of seat %d offers the picks %v", name, seat, av.Pickable)
		}
		for _, l := range av.Assignable {
			if len(l) != 0 {
				s.failf("%s: view of seat %d offers assignments %v", name, seat, av.Assignable)
			}
		}
	} else if a.Step == StepAssign && !reflect.DeepEqual(av.Assignable, g.Assignable()) {
		s.failf("%s: captain view: assignable %v, want %v", name, av.Assignable, g.Assignable())
	}
	wantPredict := -1
	if a.Step == StepPredict {
		wantPredict = g.pendingItem().index
	}
	if av.Predict != wantPredict {
		s.failf("%s: view of seat %d: predict %d, want %d", name, seat, av.Predict, wantPredict)
	}

	// Communication options by mode.
	if !over && a.Step == StepPlay {
		switch {
		case a.Comm == CommNone, a.Comm == CommRapture && a.TokensLeft() == 0, wantFirst && len(a.Tricks) > 0,
			len(a.Tricks)+1 < mdef.Comm.FromTrick:
			if len(av.Actions.Communicate) != 0 {
				s.failf("%s: view of seat %d offers communication %v (mode %q, tokens %d, trick %d)", name, seat, av.Actions.Communicate, a.Comm, a.TokensLeft(), len(a.Tricks)+1)
			}
		}
		if a.Comm == CommCurrents {
			for _, o := range av.Actions.Communicate {
				if o.Position != PosHidden {
					s.failf("%s: view of seat %d: position %q offered in the currents", name, seat, o.Position)
				}
			}
		}
	}
	for _, mv := range av.Members {
		for _, c := range mv.Communications {
			if a.Comm == CommCurrents && c.Position != PosHidden {
				s.failf("%s: view of seat %d: position %q of a shown communication in the currents", name, seat, c.Position)
			}
		}
	}
}

// checkDeepSea verifies the state itself.
func (s *simRun) checkDeepSea(name string) {
	s.t.Helper()
	g, a := s.g, s.g.Attempt
	base, _ := MissionFor(EditionDeepSea, g.Mission)
	m := g.MissionDef()
	n := len(a.Members)

	// Step consistency.
	if g.Phase == PhasePlaying {
		it := g.pendingItem()
		ok := false
		switch a.Step {
		case StepPlay, StepPass:
			ok = it.kind == itemDone
		case StepDraft:
			ok = it.kind == itemDraft && m.Distribution == DistDraft
		case StepAssign:
			ok = it.kind == itemAssign
		case StepPredict:
			ok = it.kind == itemPredict && a.Tasks[it.index].Objective.Kind == KindPredict && a.Tasks[it.index].Objective.Count < 0
		case StepPick:
			ok = it.kind == itemOneMember || it.kind == itemVolunteer
		case StepPoll:
			ok = it.kind == itemVolunteer && a.Poll != nil
		}
		if !ok {
			s.failf("%s: step %s with pending item %+v", name, a.Step, it)
		}
		if (a.Poll != nil) != (a.Step == StepPoll || (a.Step == StepPick && it.kind == itemVolunteer)) {
			s.failf("%s: poll %v in step %s", name, a.Poll != nil, a.Step)
		}
		if (a.Pass != nil) != (a.Step == StepPass) {
			s.failf("%s: pass %v in step %s", name, a.Pass != nil, a.Step)
		}
		if a.Step == StepPlay || a.Step == StepPass || a.Step == StepPredict {
			for _, t := range a.Tasks {
				if t.Owner() == OwnerUnassigned {
					s.failf("%s: task %d has no owner in step %s", name, t.ID, a.Step)
				}
			}
		}
	}

	// Ownership.
	for _, t := range a.Tasks {
		o := t.Owner()
		if o != OwnerUnassigned && (o < 0 || o >= n) {
			s.failf("%s: task %d owner %d", name, t.ID, o)
		}
		if o < 0 {
			continue
		}
		tc, _ := TaskCardFor(t.CardID)
		if o == a.Commander && tc.NotForCaptain() {
			s.failf("%s: the captain owns task %d (card %d, not for the captain)", name, t.ID, t.CardID)
		}
		if o == a.Commander && base.CaptainSkipped {
			s.failf("%s: the captain owns task %d in mission %d", name, t.ID, g.Mission)
		}
		if len(a.Chosen) > 0 && !slices.Contains(a.Chosen, o) {
			s.failf("%s: task %d belongs to seat %d, not to the chosen %v", name, t.ID, o, a.Chosen)
		}
	}
	if a.Step == StepPlay || a.Step == StepPass {
		if len(a.Chosen) > 0 {
			for _, v := range a.Chosen {
				if !slices.ContainsFunc(a.Tasks, func(t Task) bool { return t.Owner() == v }) {
					s.failf("%s: chosen seat %d has no task (%v)", name, v, a.Chosen)
				}
			}
		}
	}
	if len(a.Chosen) > 0 {
		want := max(base.Volunteers, 1)
		if base.Distribution == DistOneMember || base.Distribution == DistCaptainOrVolunteer {
			want = 1
		}
		if len(a.Chosen) != want || (want == 2 && a.Chosen[0] == a.Chosen[1]) {
			s.failf("%s: chosen %v, want %d seat(s)", name, a.Chosen, want)
		}
	}

	// Tasks and difficulty.
	ids := map[int]bool{}
	sum := 0
	for _, t := range a.Tasks {
		tc, ok := TaskCardFor(t.CardID)
		if !ok || ids[t.CardID] {
			s.failf("%s: task %d: card %d (duplicate or unknown)", name, t.ID, t.CardID)
		}
		ids[t.CardID] = true
		if t.ID >= len(a.Tasks) || a.Tasks[t.ID].CardID != t.CardID {
			s.failf("%s: task id %d out of place", name, t.ID)
		}
		if t.Difficulty != tc.DifficultyFor(n) || t.FaceDown {
			s.failf("%s: task %d: difficulty %d / faceDown %v for card %d", name, t.ID, t.Difficulty, t.FaceDown, t.CardID)
		}
		sum += t.Difficulty
	}
	switch {
	case base.FixedTasks != nil:
		var got []int
		for _, t := range a.Tasks {
			got = append(got, t.CardID)
		}
		if !slices.Equal(got, base.FixedTasks) {
			s.failf("%s: fixed tasks %v, want %v", name, got, base.FixedTasks)
		}
	case base.DrawsTasks():
		if sum != base.Difficulty {
			s.failf("%s: total difficulty %d, want %d (%d members)", name, sum, base.Difficulty, n)
		}
	default:
		if len(a.Tasks) != 0 {
			s.failf("%s: mission %d has %d tasks", name, g.Mission, len(a.Tasks))
		}
	}
	if base.DrawsTasks() {
		seen := map[int]bool{}
		for _, id := range a.TaskDeck {
			if ids[id] || seen[id] {
				s.failf("%s: task deck holds card %d twice or on the table", name, id)
			}
			seen[id] = true
		}
		if len(a.TaskDeck)+len(a.Tasks) != TaskCards {
			s.failf("%s: task deck %d + tasks %d != %d", name, len(a.TaskDeck), len(a.Tasks), TaskCards)
		}
	} else if a.TaskDeck != nil {
		s.failf("%s: task deck in mission %d", name, g.Mission)
	}
	if base.DrawsTasks() {
		if !slices.IsSorted(g.UsedTasks) || len(slices.Compact(slices.Clone(g.UsedTasks))) != len(g.UsedTasks) {
			s.failf("%s: used tasks %v not a sorted set", name, g.UsedTasks)
		}
		for id := range ids {
			if !slices.Contains(g.UsedTasks, id) {
				s.failf("%s: card %d on the table is not marked used", name, id)
			}
		}
	}

	// The captain and the re-deal, before the first card.
	if len(a.Tricks) == 0 && len(a.Trick.Plays) == 0 && a.Result == nil {
		if !slices.Contains(a.Members[a.Commander].Cards(), CommanderCard) {
			s.failf("%s: the captain (seat %d) does not hold %v", name, a.Commander, CommanderCard)
		}
		if a.needsRedeal() {
			s.failf("%s: an alignment that needs a re-deal is on the table", name)
		}
		if a.Members[a.Commander].Dummy {
			s.failf("%s: the dummy is the captain", name)
		}
	}
	// Mission objectives.
	if len(a.Objectives) != len(m.Objectives) {
		s.failf("%s: %d objectives, mission has %d", name, len(a.Objectives), len(m.Objectives))
	}

	// Communication modes.
	switch base.Comm.Mode {
	case CommUnfamiliar:
		if a.Terrain == nil || a.Terrain.IsTrump() || !a.Terrain.Valid() {
			s.failf("%s: terrain %v", name, a.Terrain)
		}
		want := CommNormal
		switch {
		case a.Terrain.Rank >= 7:
			want = CommRapture
		case a.Terrain.Rank >= 4:
			want = CommCurrents
		}
		if a.Comm != want {
			s.failf("%s: terrain %v gives mode %q, state says %q", name, a.Terrain, want, a.Comm)
		}
	default:
		if a.Terrain != nil || a.Comm != base.Comm.Mode {
			s.failf("%s: mode %q terrain %v in mission %d", name, a.Comm, a.Terrain, g.Mission)
		}
	}
	total := 0
	onlyFirst := (g.Mission == 10 || g.Mission == 13) && len(a.Chosen) == 1 && a.Chosen[0] != a.Commander
	for seat, cs := range a.Comms {
		total += len(cs)
		if len(cs) > 1 && a.Comm != CommRapture {
			s.failf("%s: seat %d communicated %d times in mode %q", name, seat, len(cs), a.Comm)
		}
		if a.Members[seat].Dummy && len(cs) > 0 {
			s.failf("%s: the dummy communicated", name)
		}
		for _, c := range cs {
			switch {
			case c.Card.IsTrump():
				s.failf("%s: seat %d communicated %v", name, seat, c.Card)
			case a.Comm == CommCurrents && c.Position != PosHidden:
				s.failf("%s: seat %d: position %q in the currents", name, seat, c.Position)
			case a.Comm != CommCurrents && c.Position == PosHidden:
				s.failf("%s: seat %d: no position in mode %q", name, seat, a.Comm)
			case onlyFirst && c.Trick != 1:
				s.failf("%s: seat %d communicated before trick %d in mission %d", name, seat, c.Trick, g.Mission)
			case c.Trick < max(base.Comm.FromTrick, 1):
				s.failf("%s: seat %d communicated before trick %d (mission %d)", name, seat, c.Trick, g.Mission)
			}
		}
	}
	switch a.Comm {
	case CommNone:
		if total != 0 {
			s.failf("%s: %d communications without communication", name, total)
		}
	case CommRapture:
		if total > n-2 || a.SonarTokens() != n-2 || a.TokensLeft() != n-2-total {
			s.failf("%s: %d communications, %d sonar tokens, %d left", name, total, a.SonarTokens(), a.TokensLeft())
		}
	}

	// Logbook: the epilogue never counts attempts.
	for mission := 33; mission <= 41; mission++ {
		if e := g.Logbook[mission]; e.Attempts != 0 {
			s.failf("%s: epilogue mission %d counted %d attempts", name, mission, e.Attempts)
		}
	}
	if e := g.Logbook[g.Mission]; e.Completed && e.Attempts < 1 && !base.Uncounted {
		s.failf("%s: mission %d completed without an attempt", name, g.Mission)
	}

	s.checkClone(name)
}

// ---- Clone: equal, and nothing shared ----

var timeType = reflect.TypeFor[time.Time]()

// collectAddrs adds the address of every pointer, map and non-empty slice reachable from v.
func collectAddrs(v reflect.Value, into map[uintptr]bool) {
	if v.Type() == timeType {
		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			into[v.Pointer()] = true
			collectAddrs(v.Elem(), into)
		}
	case reflect.Interface:
		if !v.IsNil() {
			collectAddrs(v.Elem(), into)
		}
	case reflect.Map:
		if !v.IsNil() {
			into[v.Pointer()] = true
			for it := v.MapRange(); it.Next(); {
				collectAddrs(it.Value(), into)
			}
		}
	case reflect.Slice:
		if v.Len() > 0 {
			into[v.Pointer()] = true
			for i := range v.Len() {
				collectAddrs(v.Index(i), into)
			}
		}
	case reflect.Array:
		for i := range v.Len() {
			collectAddrs(v.Index(i), into)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			collectAddrs(v.Field(i), into)
		}
	}
}

func (s *simRun) checkClone(name string) {
	s.t.Helper()
	g := s.g
	c := g.Clone()
	if !reflect.DeepEqual(g, c) {
		s.failf("%s: Clone is not deeply equal to the game", name)
	}
	orig, cp := map[uintptr]bool{}, map[uintptr]bool{}
	collectAddrs(reflect.ValueOf(g), orig)
	collectAddrs(reflect.ValueOf(c), cp)
	for p := range cp {
		if orig[p] {
			s.failf("%s: Clone shares memory with the game", name)
		}
	}
}

// ---- the random policy ----

// stepDS performs one random legal action in the running attempt.
func (s *simRun) stepDS() {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	actors := g.Actors()
	if len(actors) == 0 {
		s.failf("no actors in step %s", a.Step)
	}
	s.st.step(a.Step)

	if rng.IntN(500) == 0 {
		seat := rng.IntN(len(g.Players))
		s.act(fmt.Sprintf("giveUp(%d)", seat), func() error { return g.GiveUp(s.pid(seat)) })
		return
	}
	if a.Step != StepPlay && a.Step != StepPass && s.replaceTask() {
		return
	}
	actor := actors[rng.IntN(len(actors))]
	id := s.pid(actor)
	if rng.IntN(10) == 0 {
		s.wrongStepCall(id)
	}

	switch a.Step {
	case StepPoll:
		if rng.IntN(4) == 0 {
			s.expectErr("answerPoll(captain)", ErrNotYourTurn, func() error { return g.AnswerPoll(s.pid(a.Commander), true) })
		}
		for sd := range a.Members {
			if a.Poll.Answers[sd] != nil && rng.IntN(4) == 0 {
				s.expectErr("answerPoll(again)", ErrAlreadyAnswered, func() error { return g.AnswerPoll(s.pid(sd), true) })
				break
			}
		}
		yes := rng.IntN(2) == 0
		s.act(fmt.Sprintf("answerPoll(%d,%v)", actor, yes), func() error { return g.AnswerPoll(id, yes) })
		s.bump("ds:poll")

	case StepPick:
		s.pickDS()

	case StepDraft:
		s.draftDS(actor, id)

	case StepAssign:
		s.assignDS()

	case StepPredict:
		s.predictDS(actor, id)

	case StepPass:
		var cands []Card
		trump := Card{}
		for _, c := range a.Members[actor].Hand {
			if c.IsTrump() {
				trump = c
			} else {
				cands = append(cands, c)
			}
		}
		if trump != (Card{}) && rng.IntN(4) == 0 {
			s.expectErr("choosePassCard(trump)", ErrCannotPassTrump, func() error { return g.ChoosePassCard(id, trump) })
		}
		if len(cands) == 0 {
			s.failf("seat %d has no card to pass", actor)
		}
		card := cands[rng.IntN(len(cands))]
		s.act(fmt.Sprintf("choosePassCard(%d,%v)", actor, card), func() error { return g.ChoosePassCard(id, card) })

	case StepPlay:
		s.playDS()

	default:
		s.failf("no action for step %q", a.Step)
	}
}

// wrongStepCall calls a mutator of another step: it must be rejected without a change.
func (s *simRun) wrongStepCall(id PlayerID) {
	s.t.Helper()
	g, rng := s.g, s.rng
	st := g.Attempt.Step
	calls := map[Step]func() error{
		StepPoll:    func() error { return g.AnswerPoll(id, true) },
		StepPick:    func() error { return g.PickSeats(id, []int{0}) },
		StepDraft:   func() error { return g.PickTask(id, 0) },
		StepAssign:  func() error { return g.AssignTask(id, 0, 0) },
		StepPredict: func() error { return g.Predict(id, 0, 0) },
		StepPass:    func() error { return g.ChoosePassCard(id, Card{SuitPink, 1}) },
		StepPlay:    func() error { return g.PlayCard(id, Card{SuitPink, 1}, rng) },
	}
	var steps []Step
	for k := range calls {
		if k != st {
			steps = append(steps, k)
		}
	}
	slices.Sort(steps)
	k := steps[rng.IntN(len(steps))]
	s.expectErr("wrongStep("+string(k)+")", ErrWrongStep, calls[k])
	if st != StepDraft {
		s.expectErr("passTask(wrong step)", ErrWrongStep, func() error { return g.PassTask(id) })
	}
}

// replaceTask lets the captain replace a task when the views advertise it.
func (s *simRun) replaceTask() bool {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	cmd := a.Commander
	v, err := g.ViewFor(s.pid(cmd))
	if err != nil {
		s.failf("view: %v", err)
	}
	adv := v.Attempt.Actions.ReplaceTasks
	if adv != (g.CanReplaceTasks(cmd) == nil) {
		s.failf("Actions.ReplaceTasks %v but CanReplaceTasks %v", adv, g.CanReplaceTasks(cmd))
	}
	if o := s.otherHuman(cmd); o >= 0 && rng.IntN(12) == 0 {
		s.expectErr("replaceTask(not the captain)", ErrNotCommander, func() error { return g.ReplaceTask(s.pid(o), 0, rng) })
	}
	if len(a.Tasks) == 0 || rng.IntN(5) != 0 {
		return false
	}
	task := rng.IntN(len(a.Tasks))
	if !adv {
		s.expectErr("replaceTask(unavailable)", nil, func() error { return g.ReplaceTask(s.pid(cmd), task, rng) })
		return false
	}
	before := a.Tasks[task].CardID
	diff := a.Tasks[task].Difficulty
	err = s.try(fmt.Sprintf("replaceTask(%d)", task), func() error { return g.ReplaceTask(s.pid(cmd), task, rng) })
	switch {
	case err == nil:
		a = g.Attempt
		if a.Tasks[task].CardID == before || a.Tasks[task].Difficulty != diff {
			s.failf("replaceTask(%d): card %d (difficulty %d) became card %d (difficulty %d)", task, before, diff, a.Tasks[task].CardID, a.Tasks[task].Difficulty)
		}
		s.bump("ds:replace")
	case errors.Is(err, ErrNotAllowed):
		s.bump("ds:replace:none")
	default:
		s.failf("advertised replaceTask failed: %v", err)
	}
	return true
}

func (s *simRun) pickDS() {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	m := g.MissionDef()
	cid := s.pid(a.Commander)
	pick := g.Pick()
	want := 1
	switch pick {
	case PickOneMember:
	case PickVolunteers:
		if m.Volunteers > 1 {
			want = m.Volunteers
		}
	default:
		s.failf("pick step with pick %q", pick)
	}
	pickable := g.Pickable()
	if len(pickable) < want {
		s.failf("pick %s: only %v pickable, want %d", pick, pickable, want)
	}
	if o := s.otherHuman(a.Commander); o >= 0 && rng.IntN(4) == 0 {
		s.expectErr("pick(not the captain)", ErrNotCommander, func() error { return g.PickSeats(s.pid(o), pickable[:want]) })
	}
	if rng.IntN(4) == 0 {
		s.expectErr("pick(wrong count)", ErrInvalidSeat, func() error { return g.PickSeats(cid, append(slices.Clone(pickable[:want]), 0)) })
	}
	if rng.IntN(4) == 0 {
		s.expectErr("pick(no such seat)", ErrInvalidSeat, func() error { return g.PickSeats(cid, slices.Repeat([]int{len(a.Members)}, want)) })
	}
	if len(pickable) < len(a.Members) && rng.IntN(2) == 0 {
		for sd := range a.Members {
			if !slices.Contains(pickable, sd) {
				seats := slices.Repeat([]int{sd}, want)
				if want == 2 {
					seats[1] = pickable[0]
				}
				s.expectErr(fmt.Sprintf("pick(unpickable %d)", sd), ErrInvalidSeat, func() error { return g.PickSeats(cid, seats) })
				break
			}
		}
	}
	seats := slices.Clone(pickable)
	rng.Shuffle(len(seats), func(i, j int) { seats[i], seats[j] = seats[j], seats[i] })
	seats = seats[:want]
	s.act(fmt.Sprintf("pick(%s %v)", pick, seats), func() error { return g.PickSeats(cid, seats) })
	s.bump("ds:pick:" + string(pick))
}

// draftOrder is the clockwise order of the draft: from the captain, without them if skipped.
func (s *simRun) draftOrder() []int {
	a := s.g.Attempt
	var out []int
	for i := range a.Members {
		sd := (a.Commander + i) % len(a.Members)
		if sd == a.Commander && s.baseMission().CaptainSkipped {
			continue
		}
		out = append(out, sd)
	}
	return out
}

func (s *simRun) baseMission() MissionDef {
	m, _ := MissionFor(EditionDeepSea, s.g.Mission)
	return m
}

func (s *simRun) draftDS(actor int, id PlayerID) {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	base := s.baseMission()
	order := s.draftOrder()
	turn := order[s.dsPos%len(order)]
	if g.draftTurn() != turn || actor != a.actorFor(turn) {
		s.failf("draft turn is seat %d (actor %d), want seat %d", g.draftTurn(), actor, turn)
	}
	isCaptain := turn == a.Commander
	hasTask := slices.ContainsFunc(a.Tasks, func(t Task) bool { return t.Owner() == turn })
	forbidden := func(t Task) bool {
		tc, _ := TaskCardFor(t.CardID)
		return isCaptain && tc.NotForCaptain()
	}
	top := 0
	for _, t := range a.Tasks {
		if t.Owner() == OwnerUnassigned && !forbidden(t) {
			top = max(top, t.Difficulty)
		}
	}
	tk := g.Takeable()
	for i, t := range a.Tasks {
		want := t.Owner() == OwnerUnassigned && !forbidden(t)
		if want && base.HardestFirst && isCaptain && !hasTask && t.Difficulty != top {
			want = false
		}
		if want != slices.Contains(tk, i) {
			s.failf("takeable %v: task %d (card %d, difficulty %d, owner %d) should be takeable=%v; top %d", tk, i, t.CardID, t.Difficulty, t.Owner(), want, top)
		}
	}
	// Passing: fewer tasks than members, and the members after this one can still take the rest.
	remaining := len(a.Tasks) - a.tasksTaken()
	wantPass := len(tk) == 0 || (len(a.Tasks) < len(order) && s.dsPos < len(order) && remaining <= len(order)-s.dsPos-1)
	// dsPos counts picks and passes; positions beyond the first round never allow a pass.
	canPass := g.CanPass() == nil
	if canPass != wantPass {
		s.failf("CanPass %v, want %v (tasks %d, members %d, pos %d, taken %d, takeable %v)", canPass, wantPass, len(a.Tasks), len(order), s.dsPos, a.tasksTaken(), tk)
	}

	// Illegal calls.
	if rng.IntN(4) == 0 {
		var bad []int
		for i := -1; i <= len(a.Tasks); i++ {
			if !slices.Contains(tk, i) {
				bad = append(bad, i)
			}
		}
		t := bad[rng.IntN(len(bad))]
		s.expectErr(fmt.Sprintf("pickTask(not takeable %d)", t), ErrInvalidTask, func() error { return g.PickTask(id, t) })
	}
	if o := s.otherHuman(actor); o >= 0 && rng.IntN(5) == 0 {
		s.expectErr("pickTask(not your turn)", ErrNotYourTurn, func() error { return g.PickTask(s.pid(o), 0) })
		s.expectErr("passTask(not your turn)", ErrNotYourTurn, func() error { return g.PassTask(s.pid(o)) })
	}
	if !canPass && rng.IntN(3) == 0 {
		s.expectErr("passTask(not allowed)", ErrNotAllowed, func() error { return g.PassTask(id) })
	}

	if canPass && (len(tk) == 0 || rng.IntN(3) == 0) {
		s.act(fmt.Sprintf("passTask(seat %d)", turn), func() error { return g.PassTask(id) })
		s.dsPos++
		s.bump("ds:pass")
		return
	}
	if len(tk) == 0 {
		s.failf("seat %d can neither take a task nor pass", turn)
	}
	task := tk[rng.IntN(len(tk))]
	s.act(fmt.Sprintf("pickTask(%d by %d)", task, actor), func() error { return g.PickTask(id, task) })
	s.dsPos++
	if base.HardestFirst && isCaptain && !hasTask && a.Tasks[task].Difficulty != top {
		s.failf("the captain's first task has difficulty %d, the highest is %d", a.Tasks[task].Difficulty, top)
	}
	s.bump("ds:draft")
}

func (s *simRun) assignDS() {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	cid := s.pid(a.Commander)
	as := g.Assignable()
	type pair struct{ task, seat int }
	var ok, bad []pair
	for t := range a.Tasks {
		for sd := range a.Members {
			p := pair{t, sd}
			if slices.Contains(as[t], sd) {
				ok = append(ok, p)
				if tc, _ := TaskCardFor(a.Tasks[t].CardID); sd == a.Commander && !a.Tasks[t].FaceDown && tc.NotForCaptain() {
					s.failf("task %d (card %d) may go to the captain", t, a.Tasks[t].CardID)
				}
				if a.Tasks[t].Owner() != OwnerUnassigned {
					s.failf("task %d is owned but assignable", t)
				}
			} else {
				bad = append(bad, p)
			}
		}
	}
	if len(ok) == 0 {
		s.failf("nothing can be assigned (assignable %v, chosen %v)", as, a.Chosen)
	}
	if len(bad) > 0 && rng.IntN(3) == 0 {
		p := bad[rng.IntN(len(bad))]
		s.expectErr(fmt.Sprintf("assignTask(%d->%d, not allowed)", p.task, p.seat), nil, func() error { return g.AssignTask(cid, p.task, p.seat) })
	}
	if o := s.otherHuman(a.Commander); o >= 0 && rng.IntN(5) == 0 {
		s.expectErr("assignTask(not the captain)", ErrNotCommander, func() error { return g.AssignTask(s.pid(o), ok[0].task, ok[0].seat) })
	}
	if rng.IntN(6) == 0 {
		s.expectErr("assignTask(no such task)", ErrInvalidTask, func() error { return g.AssignTask(cid, len(a.Tasks), 0) })
	}
	p := ok[rng.IntN(len(ok))]
	s.act(fmt.Sprintf("assignTask(%d->%d)", p.task, p.seat), func() error { return g.AssignTask(cid, p.task, p.seat) })
	s.bump("ds:assign")
}

func (s *simRun) predictDS(actor int, id PlayerID) {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	v, err := g.ViewFor(id)
	if err != nil {
		s.failf("view: %v", err)
	}
	task := v.Attempt.Predict
	if task < 0 || task >= len(a.Tasks) || a.Tasks[task].Objective.Kind != KindPredict || a.Tasks[task].Objective.Count >= 0 {
		s.failf("predict step with task %d", task)
	}
	if actor != a.actorFor(a.Tasks[task].Owner()) {
		s.failf("actor %d, owner %d", actor, a.Tasks[task].Owner())
	}
	total := v.Attempt.TotalTricks
	if rng.IntN(3) == 0 {
		s.expectErr("predict(-1)", ErrInvalidPrediction, func() error { return g.Predict(id, task, -1) })
		s.expectErr("predict(too many)", ErrInvalidPrediction, func() error { return g.Predict(id, task, total+1) })
	}
	if len(a.Tasks) > 1 && rng.IntN(4) == 0 {
		s.expectErr("predict(other task)", ErrInvalidTask, func() error { return g.Predict(id, (task+1)%len(a.Tasks), 0) })
	}
	if o := s.otherHuman(actor); o >= 0 && rng.IntN(4) == 0 {
		s.expectErr("predict(not the owner)", ErrNotYourTurn, func() error { return g.Predict(s.pid(o), task, 0) })
	}
	tricks := rng.IntN(total + 1)
	s.act(fmt.Sprintf("predict(task %d by %d: %d)", task, actor, tricks), func() error { return g.Predict(id, task, tricks) })
	if a.Tasks[task].Objective.Count != tricks {
		s.failf("prediction %d was stored as %d", tricks, a.Tasks[task].Objective.Count)
	}
	if a.Tasks[task].Objective.Hidden {
		s.bump("ds:predict:hidden")
	} else {
		s.bump("ds:predict:open")
	}
}

type commCand struct {
	seat int
	opt  CommOption
}

func (s *simRun) commCandidates() []commCand {
	var out []commCand
	for _, seat := range s.humanSeats() {
		v, err := s.g.ViewFor(s.pid(seat))
		if err != nil {
			s.failf("view: %v", err)
		}
		for _, o := range v.Attempt.Actions.Communicate {
			out = append(out, commCand{seat, o})
		}
	}
	return out
}

// badCommunicate asks a random human to communicate a random card: it must succeed exactly when
// the views offer it.
func (s *simRun) badCommunicate() {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	seat := s.humanSeats()[rng.IntN(len(s.humanSeats()))]
	hand := a.Members[seat].Hand
	if len(hand) == 0 {
		return
	}
	card := hand[rng.IntN(len(hand))]
	pos := []Position{PosHighest, PosOnly, PosLowest}[rng.IntN(3)]
	v, _ := g.ViewFor(s.pid(seat))
	offered := false
	for _, o := range v.Attempt.Actions.Communicate {
		if o.Card == card && (a.Comm == CommCurrents || o.Position == pos) {
			offered = true
		}
	}
	name := fmt.Sprintf("communicate?(%d,%v,%q)", seat, card, pos)
	err := s.try(name, func() error { return g.Communicate(s.pid(seat), card, pos) })
	if offered && err != nil {
		s.failf("%s offered but rejected: %v", name, err)
	}
	if !offered && err == nil {
		s.failf("%s accepted but not offered", name)
	}
	if err == nil {
		s.bump("ds:comm:extra")
	}
}

func (s *simRun) playDS() {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	cmd := a.Commander
	cid := s.pid(cmd)

	if a.prePlay() {
		if !g.Options.Distress && rng.IntN(4) == 0 {
			if err := s.try("sendDistress(off)", func() error { return g.SendDistress(cid, DirLeft) }); !errors.Is(err, ErrNotAllowed) {
				s.failf("sendDistress while off: %v, want ErrNotAllowed", err)
			}
			if err := s.try("skipDistress(off)", func() error { return g.SkipDistress(cid) }); !errors.Is(err, ErrNotAllowed) {
				s.failf("skipDistress while off: %v, want ErrNotAllowed", err)
			}
		}
		if v, _ := g.ViewFor(cid); v.Attempt.Actions.SendDistress != (g.CanSendDistress(cmd) == nil) {
			s.failf("Actions.SendDistress disagrees with CanSendDistress")
		}
		if g.CanSendDistress(cmd) == nil {
			switch rng.IntN(7) {
			case 0, 1:
				dir := DirLeft
				if rng.IntN(2) == 0 {
					dir = DirRight
				}
				s.act("sendDistress("+string(dir)+")", func() error { return g.SendDistress(cid, dir) })
				s.distres = true
				s.bump("ds:distress")
				return
			case 2:
				s.act("skipDistress", func() error { return g.SkipDistress(cid) })
			}
		}
	}

	if len(a.Trick.Plays) == 0 {
		// Communications: with rapture of the deep several, while a sonar token is left.
		for k := 0; ; k++ {
			stop := 2
			if k == 0 {
				stop = 3
				if a.Comm == CommRapture {
					stop = 2
				}
			}
			if rng.IntN(stop) != 0 {
				break
			}
			cands := s.commCandidates()
			if len(cands) == 0 {
				break
			}
			c := cands[rng.IntN(len(cands))]
			s.act(fmt.Sprintf("communicate(%d,%v,%q)", c.seat, c.opt.Card, c.opt.Position),
				func() error { return g.Communicate(s.pid(c.seat), c.opt.Card, c.opt.Position) })
			s.bump("ds:comm")
			if len(a.Comms[c.seat]) > 1 {
				s.bump("ds:comm:repeat")
			}
		}
	}
	if rng.IntN(5) == 0 {
		s.badCommunicate()
	}

	turn := a.turn()
	legal := g.LegalCards(turn)
	if len(legal) == 0 {
		s.failf("seat %d has no legal card", turn)
	}
	actor := a.actorFor(turn)
	if !slices.Contains(g.Actors(), actor) {
		s.failf("actor %d for seat %d is not in Actors %v", actor, turn, g.Actors())
	}
	card := s.chooseCard(legal)
	s.act(fmt.Sprintf("play(seat %d by %d: %v)", turn, actor, card), func() error { return g.PlayCard(s.pid(actor), card, rng) })
}

// chooseCard picks uniformly, or, with rollouts on, the card whose random playouts won most.
func (s *simRun) chooseCard(legal []Card) Card {
	rng := s.rng
	if s.rollouts == 0 || len(legal) == 1 {
		return legal[rng.IntN(len(legal))]
	}
	pr := rand.New(rand.NewPCG(rng.Uint64(), rng.Uint64()))
	best, bestWins := []Card(nil), -1
	for _, c := range legal {
		wins := 0
		for range s.rollouts {
			if s.playout(c, pr) {
				wins++
			}
		}
		switch {
		case wins > bestWins:
			best, bestWins = []Card{c}, wins
		case wins == bestWins:
			best = append(best, c)
		}
	}
	return best[rng.IntN(len(best))]
}

// playout plays first and then random cards on a copy; it reports whether the attempt was won.
func (s *simRun) playout(first Card, pr *rand.Rand) bool {
	c := s.g.Clone()
	a := c.Attempt
	if c.PlayCard(c.Players[a.actorFor(a.turn())].ID, first, pr) != nil {
		return false
	}
	for c.Phase == PhasePlaying && c.Attempt.Step == StepPlay {
		a := c.Attempt
		turn := a.turn()
		legal := c.LegalCards(turn)
		if len(legal) == 0 {
			return false
		}
		if c.PlayCard(c.Players[a.actorFor(turn)].ID, legal[pr.IntN(len(legal))], pr) != nil {
			return false
		}
	}
	return c.Phase == PhaseDebrief && c.Attempt.Result.Success
}

// ---- driving a game ----

// runAttemptDS plays the running attempt to the debrief and checks the result and the logbook.
func (s *simRun) runAttemptDS(wantNumber int) {
	s.t.Helper()
	g := s.g
	mission := g.Mission
	base := s.baseMission()
	e0 := g.Logbook[mission]
	if g.Phase != PhasePlaying || g.Attempt == nil {
		s.failf("attempt did not start: phase %s", g.Phase)
	}
	if g.Attempt.Number != wantNumber {
		s.failf("attempt number %d, want %d", g.Attempt.Number, wantNumber)
	}
	s.distres, s.dsPos = false, 0
	s.checkEvents("start")
	s.checkInvariants("start")
	if g.Attempt.Redeals > 0 {
		s.bump("ds:redeal")
	}
	if g.Attempt.Comm != CommNormal {
		s.bump("ds:mode:" + string(g.Attempt.Comm))
	}

	for n := 0; g.Phase == PhasePlaying; n++ {
		if n >= dsMaxActions {
			s.failf("the attempt does not end after %d actions (step %s)", n, g.Attempt.Step)
		}
		s.stepDS()
	}
	a := g.Attempt
	if g.Phase != PhaseDebrief || a.Result == nil {
		s.failf("attempt ended in phase %s", g.Phase)
	}
	r := a.Result
	if r.Success != (r.Reason == ReasonCompleted) {
		s.failf("result %+v: success and reason disagree", r)
	}
	if r.Success && len(a.Tricks) == 0 {
		s.failf("won without a trick")
	}
	if r.Reason == ReasonTokenOrder {
		s.failf("token order in Deep Sea: %+v", r)
	}
	if a.Tasks != nil && r.Success {
		for _, t := range a.Tasks {
			if t.Owner() == OwnerUnassigned {
				s.failf("won with an unassigned task")
			}
		}
	}
	s.st.count(func() {
		s.st.attempts++
		if r.Success {
			s.st.won++
			s.st.followUp[fmt.Sprintf("ds:won:%02d", mission)]++
		} else {
			s.st.lost[r.Reason]++
		}
	})
	for _, cs := range a.Comms {
		if len(cs) > 1 {
			s.bump("ds:rapture:multi")
		}
	}

	e1 := g.Logbook[mission]
	if e0.Completed {
		if e1 != e0 {
			s.failf("a replay changed the logbook %+v -> %+v", e0, e1)
		}
	} else {
		want := LogEntry{Attempts: e0.Attempts + b2i(!base.Uncounted), Distress: e0.Distress || s.distres, Completed: r.Success}
		if e1 != want {
			s.failf("logbook %+v, want %+v", e1, want)
		}
	}
}

// startNumber is the number of the first attempt started from the lobby or as the next mission.
func (s *simRun) startNumber() int {
	e := s.g.Logbook[s.g.Mission]
	if e.Completed {
		return 1
	}
	return e.Attempts + 1
}

// followUpDS exercises the debrief once. With more it usually starts another attempt, plays it
// and returns true.
func (s *simRun) followUpDS(more bool) bool {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	r := a.Result
	host := alice
	mission := g.Mission
	completed := r.Success

	// Calls the outcome does not allow, and a guest's calls.
	switch {
	case completed:
		s.expectErr("nextAttempt(completed)", ErrInvalidResultChoice, func() error { return g.NextAttempt(host, NextAttemptChoice{}, rng) })
		s.expectErr("backToLobby(won, uncounted)", ErrInvalidResultChoice, func() error { return g.BackToLobby(host, true) })
	default:
		s.expectErr("nextMission(lost)", ErrInvalidResultChoice, func() error { return g.NextMission(host, rng) })
	}
	s.expectErr("nextAttempt(guest)", ErrNotHost, func() error { return g.NextAttempt(bob, NextAttemptChoice{}, rng) })
	s.expectErr("nextMission(guest)", ErrNotHost, func() error { return g.NextMission(bob, rng) })
	s.expectErr("backToLobby(guest)", ErrNotHost, func() error { return g.BackToLobby(bob, false) })
	s.expectErr("endGame(guest)", ErrNotHost, func() error { return g.EndGame(bob) })
	s.expectErr("startMission(debrief)", ErrWrongPhase, func() error { return g.StartMission(host, rng) })
	s.expectErr("giveUp(debrief)", ErrWrongPhase, func() error { return g.GiveUp(host) })

	lobbyStart := func() {
		prev := g.Mission
		s.act("backToLobby", func() error { return g.BackToLobby(host, false) })
		if g.Phase != PhaseLobby || g.Attempt != nil {
			s.failf("back in the lobby: phase %s attempt %v", g.Phase, g.Attempt)
		}
		if completed {
			want := Missions(EditionDeepSea)
			for i := 1; i <= Missions(EditionDeepSea); i++ {
				if !g.Logbook[i].Completed {
					want = i
					break
				}
			}
			if g.Mission != want {
				s.failf("lobby mission %d, want the next unfinished one %d", g.Mission, want)
			}
		} else if g.Mission != prev {
			s.failf("lobby mission %d after an unfinished attempt, want %d", g.Mission, prev)
		}
		if rng.IntN(2) == 0 {
			next := 1 + rng.IntN(Missions(EditionDeepSea))
			s.act("selectMission", func() error { return g.SelectMission(host, next) })
		}
		s.act("startMission", func() error { return g.StartMission(host, rng) })
		s.runAttemptDS(s.startNumber())
	}

	if !more {
		switch rng.IntN(3) {
		case 0:
			s.act("endGame", func() error { return g.EndGame(host) })
			s.checkEnded()
		case 1:
			s.act("backToLobby", func() error { return g.BackToLobby(host, false) })
			s.act("endGame", func() error { return g.EndGame(host) })
			s.checkEnded()
		}
		return false
	}

	switch {
	case completed:
		switch rng.IntN(4) {
		case 0, 1:
			if _, ok := MissionFor(EditionDeepSea, mission+1); ok {
				s.bump("ds:nextMission")
				s.act("nextMission", func() error { return g.NextMission(host, rng) })
				if g.Mission != mission+1 {
					s.failf("next mission is %d, want %d", g.Mission, mission+1)
				}
				s.runAttemptDS(s.startNumber())
				return true
			}
			s.expectErr("nextMission(last)", ErrInvalidMission, func() error { return g.NextMission(host, rng) })
			lobbyStart()
		default:
			s.bump("ds:replay")
			lobbyStart()
		}
		return true
	}

	// Lost.
	if rng.IntN(5) == 0 {
		lobbyStart()
		return true
	}
	choice := NextAttemptChoice{NewTasks: rng.IntN(2) == 0, Uncounted: rng.IntN(3) == 0}
	var oldCards []int
	for _, t := range a.Tasks {
		oldCards = append(oldCards, t.CardID)
	}
	before := g.Logbook[mission]
	base := s.baseMission()
	want := a.Number + 1
	if choice.Uncounted {
		want = a.Number
	}
	s.act(fmt.Sprintf("nextAttempt(%+v)", choice), func() error { return g.NextAttempt(host, choice, rng) })
	after := g.Logbook[mission]
	switch {
	case choice.Uncounted && !before.Completed && !base.Uncounted:
		if after.Attempts != before.Attempts-1 {
			s.failf("uncounted attempt: logbook attempts %d -> %d", before.Attempts, after.Attempts)
		}
		s.bump("ds:uncounted")
	case after != before:
		s.failf("nextAttempt changed the logbook %+v -> %+v", before, after)
	}
	var newCards []int
	for _, t := range g.Attempt.Tasks {
		newCards = append(newCards, t.CardID)
	}
	switch {
	case base.FixedTasks != nil || len(oldCards) == 0:
	case !choice.NewTasks:
		s.bump("ds:sameTasks")
		if !slices.Equal(oldCards, newCards) {
			s.failf("same tasks: %v -> %v", oldCards, newCards)
		}
	default:
		s.bump("ds:newTasks")
	}
	s.runAttemptDS(want)
	return true
}

// runGameDS simulates one game: the mission, then up to a few attempts and debrief transitions.
func runGameDS(t *testing.T, st *simStats, mission, players int, seed uint64) {
	t.Helper()
	s := &simRun{
		t:   t,
		st:  st,
		rng: rand.New(rand.NewPCG(seed, uint64(mission*10+players)+1000)),
		tag: fmt.Sprintf("seed %d mission %d players %d", seed, mission, players),
	}
	defer func() {
		switch r := recover().(type) {
		case nil:
		case simAbort:
		default:
			t.Errorf("%s: panic: %v\n%s\n    last actions: %s", s.tag, r, debug.Stack(), strings.Join(s.log, " | "))
		}
	}()

	s.g = newDeepSeaCrew(t, players)
	g := s.g
	s.rollouts = 0
	if seed%5 == 2 {
		s.rollouts = 3
	}
	s.act("setOptions", func() error { return g.SetOptions(alice, Options{Distress: seed%2 == 1}) })
	s.act("selectMission", func() error { return g.SelectMission(alice, mission) })
	s.act("startMission", func() error { return g.StartMission(alice, s.rng) })

	limit := 1 + int(seed%3)
	number := 1
	s.runAttemptDS(number)
	for n := 1; ; n++ {
		if !s.followUpDS(n < limit) || g.Phase != PhaseDebrief {
			break
		}
	}
}

func TestSimulationDeepSea(t *testing.T) {
	seeds := 20
	if testing.Short() {
		seeds = 3
	}
	st := &simStats{lost: map[Reason]int{}, steps: map[Step]int{}, followUp: map[string]int{}}
	t.Run("games", func(t *testing.T) {
		for mission := 1; mission <= Missions(EditionDeepSea); mission++ {
			for players := MinPlayers; players <= MaxPlayers; players++ {
				t.Run(fmt.Sprintf("m%02d_p%d", mission, players), func(t *testing.T) {
					t.Parallel()
					for seed := range uint64(seeds) {
						runGameDS(t, st, mission, players, seed)
						if t.Failed() {
							return
						}
					}
				})
			}
		}
	})
	if t.Failed() {
		return
	}
	t.Logf("attempts %d, won %d, lost %v, steps %v", st.attempts, st.won, st.lost, st.steps)
	t.Logf("counters %v", st.followUp)
	if testing.Short() {
		return
	}
	if st.won == 0 {
		t.Error("no simulated attempt was won")
	}
	for _, step := range []Step{StepPoll, StepPick, StepDraft, StepAssign, StepPredict, StepPass, StepPlay} {
		if st.steps[step] == 0 {
			t.Errorf("step %s was never played", step)
		}
	}
	for _, key := range []string{
		"ds:replace", "ds:pass", "ds:draft", "ds:assign", "ds:poll", "ds:predict:hidden", "ds:predict:open",
		"ds:pick:oneMember", "ds:pick:volunteers", "ds:distress", "ds:comm", "ds:comm:repeat", "ds:rapture:multi",
		"ds:mode:currents", "ds:mode:rapture", "ds:mode:none", "ds:redeal", "ds:uncounted", "ds:sameTasks", "ds:newTasks",
		"ds:nextMission", "ds:replay",
	} {
		if st.followUp[key] == 0 {
			t.Errorf("coverage: %s never happened", key)
		}
	}
}
