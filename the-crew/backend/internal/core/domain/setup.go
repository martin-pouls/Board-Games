package domain

import (
	"fmt"
	"slices"
)

// The setup of an attempt (rules §6, §7, §10, §16) is a sequence derived from the MissionDef and
// the attempt's state, in this order:
//
//  1. the token swap or move (23, 40): StepAdjustTokens;
//  2. the mission's picks: the silenced member (11) and the roles (50): StepPick; the chosen
//     member of an objective (5, 33, 41): StepPoll, then StepPick;
//  3. the task distribution: open: StepDraft; commander's decision: StepPoll, then StepPick;
//     commander's distribution: per task, reveal, StepPoll, StepPick; Deep Sea: the draft with
//     passing (StepDraft), one member (StepPick), volunteers (StepPoll, StepPick, and for two of
//     them StepAssign), free selection (StepAssign);
//  4. Deep Sea: the predictions, one task after the other: StepPredict;
//  5. StepPlay.
//
// Only Step and the open Poll are stored; which item is pending is derived by nextItem.

// itemKind is a pending setup item.
type itemKind int

const (
	itemDone itemKind = iota
	itemTokens
	itemSilence
	itemRoles
	itemChosen     // index: the objective
	itemDraft      //
	itemDecision   //
	itemDistribute // index: the task
	itemOneMember  // Deep Sea 6
	itemVolunteer  // Deep Sea 10, 13, 14–16, 26: poll, then the captain chooses
	itemAssign     // Deep Sea: free selection, the split of 26
	itemPredict    // index: the task
)

type setupItem struct {
	kind  itemKind
	index int
}

// SetupPick says what the commander picks in StepPick.
type SetupPick string

const (
	PickNone       SetupPick = ""
	PickSilence    SetupPick = "silence"    // one human, the commander included
	PickRoles      SetupPick = "roles"      // two different seats: role A, role B
	PickObjective  SetupPick = "objective"  // one seat other than the commander
	PickDecision   SetupPick = "decision"   // one seat other than the commander receives all tasks
	PickDistribute SetupPick = "distribute" // any seat receives the revealed task
	PickOneMember  SetupPick = "oneMember"  // Deep Sea 6: any seat receives all tasks
	PickVolunteers SetupPick = "volunteers" // Deep Sea: Volunteers seats (one: it receives all tasks)
)

// nextItem returns the first setup item still pending.
func (a *Attempt) nextItem(m MissionDef) setupItem {
	if (m.Has(SetupSwapTokens) || m.Has(SetupMoveToken)) && !a.TokensAdjusted {
		return setupItem{kind: itemTokens}
	}
	if m.Has(SetupSilenceMember) && a.Silenced < 0 {
		return setupItem{kind: itemSilence}
	}
	for i, s := range m.Objectives {
		if a.Objectives[i].Owner != OwnerUnassigned {
			continue
		}
		if s.Owner == OwnerRuleChosen {
			return setupItem{kind: itemChosen, index: i}
		}
		return setupItem{kind: itemRoles}
	}
	for i, t := range a.Tasks {
		if t.Owner() != OwnerUnassigned {
			continue
		}
		switch m.Distribution {
		case DistCommanderDecision:
			return setupItem{kind: itemDecision}
		case DistCommanderDistribution:
			return setupItem{kind: itemDistribute, index: i}
		case DistOneMember:
			return setupItem{kind: itemOneMember}
		case DistCaptainOrVolunteer, DistVolunteers:
			if a.Chosen == nil {
				return setupItem{kind: itemVolunteer}
			}
			return setupItem{kind: itemAssign}
		case DistFree:
			return setupItem{kind: itemAssign}
		}
		return setupItem{kind: itemDraft}
	}
	for i, t := range a.Tasks {
		if t.Objective.Kind == KindPredict && t.Objective.Count < 0 {
			return setupItem{kind: itemPredict, index: i}
		}
	}
	return setupItem{kind: itemDone}
}

// Pick returns what the commander picks in StepPick (PickNone in any other step).
func (g *Game) Pick() SetupPick {
	a := g.Attempt
	if g.Phase != PhasePlaying || a.Step != StepPick {
		return PickNone
	}
	switch a.nextItem(g.MissionDef()).kind {
	case itemSilence:
		return PickSilence
	case itemRoles:
		return PickRoles
	case itemChosen:
		return PickObjective
	case itemDecision:
		return PickDecision
	case itemDistribute:
		return PickDistribute
	case itemOneMember:
		return PickOneMember
	case itemVolunteer:
		return PickVolunteers
	}
	return PickNone
}

// pendingItem returns the first setup item still pending.
func (g *Game) pendingItem() setupItem {
	return g.Attempt.nextItem(g.MissionDef())
}

// advanceSetup enters the step of the next pending setup item. Called after the deal and after
// every completed setup action.
func (g *Game) advanceSetup() {
	a := g.Attempt
	it := a.nextItem(g.MissionDef())
	switch it.kind {
	case itemTokens:
		a.Step = StepAdjustTokens
	case itemSilence, itemRoles, itemOneMember:
		a.Step = StepPick
	case itemChosen, itemDecision, itemVolunteer:
		g.openPoll()
	case itemAssign:
		a.Step = StepAssign
	case itemPredict:
		a.Step = StepPredict
	case itemDistribute:
		a.Tasks[it.index].FaceDown = false
		g.openPoll()
	case itemDraft:
		a.Step = StepDraft
	case itemDone:
		a.setQuotas()
		a.Step = StepPlay
		a.Trick = Trick{Leader: a.Commander}
	}
}

// openPoll asks every human except the commander.
func (g *Game) openPoll() {
	a := g.Attempt
	a.Poll = &Poll{Answers: make([]*bool, len(a.Members))}
	a.Step = StepPoll
}

// pollers returns the seats that answer polls: every human except the commander.
func (a *Attempt) pollers() []int {
	var out []int
	for _, s := range a.humans() {
		if s != a.Commander {
			out = append(out, s)
		}
	}
	return out
}

// AdjustTokens swaps the tokens of tasks from and to (mission 23), or moves the token of task from
// to task to, which has none (mission 40). Commander only.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotCommander, ErrInvalidTask,
// ErrInvalidTokenMove. Emits TokensAdjusted.
func (g *Game) AdjustTokens(by PlayerID, from, to int) error {
	seat, err := g.requireStep(by, StepAdjustTokens)
	if err != nil {
		return err
	}
	a := g.Attempt
	if err := a.requireCommander(seat); err != nil {
		return err
	}
	if !a.validTask(from) || !a.validTask(to) || from == to {
		return fmt.Errorf("%w: tasks %d and %d", ErrInvalidTask, from, to)
	}
	tf, tt := a.Tasks[from].Token, a.Tasks[to].Token
	if g.MissionDef().Has(SetupSwapTokens) {
		if tf == TokenNone || tt == TokenNone {
			return fmt.Errorf("%w: both tasks need a token to swap", ErrInvalidTokenMove)
		}
	} else if tf == TokenNone || tt != TokenNone {
		return fmt.Errorf("%w: move a token to a task without one", ErrInvalidTokenMove)
	}
	a.Tasks[from].Token, a.Tasks[to].Token = tt, tf
	a.TokensAdjusted = true
	g.emit(TokensAdjusted{From: from, To: to})
	g.advanceSetup()
	g.commit()
	return nil
}

// SkipTokens leaves the tokens as they are (commander).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotCommander. Emits TokensAdjusted.
func (g *Game) SkipTokens(by PlayerID) error {
	seat, err := g.requireStep(by, StepAdjustTokens)
	if err != nil {
		return err
	}
	if err := g.Attempt.requireCommander(seat); err != nil {
		return err
	}
	g.Attempt.TokensAdjusted = true
	g.emit(TokensAdjusted{From: -1, To: -1})
	g.advanceSetup()
	g.commit()
	return nil
}

func (a *Attempt) validTask(id int) bool { return id >= 0 && id < len(a.Tasks) }

func (a *Attempt) validSeat(s int) bool { return s >= 0 && s < len(a.Members) }

// AnswerPoll answers the open poll (every human except the commander, once each; yes = "well" in
// mission 5). The last answer moves on to StepPick.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn (the commander),
// ErrAlreadyAnswered. Emits PollAnswered.
func (g *Game) AnswerPoll(by PlayerID, yes bool) error {
	seat, err := g.requireStep(by, StepPoll)
	if err != nil {
		return err
	}
	a := g.Attempt
	if seat == a.Commander {
		return fmt.Errorf("%w: the commander asks, the crew answers", ErrNotYourTurn)
	}
	if a.Poll.Answers[seat] != nil {
		return fmt.Errorf("%w: seat %d", ErrAlreadyAnswered, seat)
	}
	a.Poll.Answers[seat] = &yes
	g.emit(PollAnswered{Seat: seat, Yes: yes})
	if !slices.ContainsFunc(a.pollers(), func(s int) bool { return a.Poll.Answers[s] == nil }) {
		a.Step = StepPick
	}
	g.commit()
	return nil
}

// Pick makes the commander's choice in StepPick (see Pick for what is picked): one seat, or two
// for the roles of mission 50 (A, then B).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotCommander, ErrInvalidSeat,
// ErrUnevenDistribution. Emits Picked (and TaskPicked per task given).
func (g *Game) PickSeats(by PlayerID, seats []int) error {
	seat, err := g.requireStep(by, StepPick)
	if err != nil {
		return err
	}
	a := g.Attempt
	if err := a.requireCommander(seat); err != nil {
		return err
	}
	it := g.pendingItem()
	m := g.MissionDef()
	want := 1
	switch {
	case it.kind == itemRoles:
		want = 2
	case it.kind == itemVolunteer && m.Volunteers > 1:
		want = m.Volunteers
	}
	if len(seats) != want {
		return fmt.Errorf("%w: pick %d seat(s)", ErrInvalidSeat, want)
	}
	for _, s := range seats {
		if !a.validSeat(s) {
			return fmt.Errorf("%w: %d", ErrInvalidSeat, s)
		}
	}
	s := seats[0]
	if want == 2 && seats[0] == seats[1] {
		return fmt.Errorf("%w: pick two different members", ErrInvalidSeat)
	}
	for _, s := range seats {
		if err := a.pickError(m, it, s); err != nil {
			return err
		}
	}

	g.emit(Picked{Seats: slices.Clone(seats)})
	switch it.kind {
	case itemSilence:
		a.Silenced = s
	case itemRoles:
		for i, spec := range g.MissionDef().Objectives {
			switch spec.Owner {
			case OwnerRuleRoleA:
				a.Objectives[i].Owner = seats[0]
			case OwnerRuleRoleB:
				a.Objectives[i].Owner = seats[1]
			}
		}
	case itemChosen:
		a.Objectives[it.index].Owner = s
	case itemDecision:
		for i := range a.Tasks {
			a.Tasks[i].Objective.Owner = s
			a.Tasks[i].FaceDown = false
			g.emit(TaskPicked{TaskID: i, Seat: s})
		}
	case itemDistribute:
		a.Tasks[it.index].Objective.Owner = s
		g.emit(TaskPicked{TaskID: it.index, Seat: s})
	case itemOneMember:
		a.Chosen = []int{s}
		g.giveAll(s)
	case itemVolunteer:
		a.Chosen = slices.Clone(seats)
		if want == 1 {
			g.giveAll(s)
		}
	}
	a.Poll = nil
	g.advanceSetup()
	g.commit()
	return nil
}

// pickError reports why seat s may not be picked for the pending item, or nil.
func (a *Attempt) pickError(m MissionDef, it setupItem, s int) error {
	switch it.kind {
	case itemOneMember, itemVolunteer:
		return a.chooseError(m, s)
	case itemSilence:
		if a.Members[s].Dummy {
			return fmt.Errorf("%w: only a human can be silenced", ErrInvalidSeat)
		}
	case itemChosen, itemDecision:
		if s == a.Commander {
			return fmt.Errorf("%w: the commander must choose another member", ErrInvalidSeat)
		}
	case itemDistribute:
		if !a.evenReachable(s) {
			return fmt.Errorf("%w: an even split would no longer be possible", ErrUnevenDistribution)
		}
	}
	return nil
}

// Pickable returns the seats the commander may pick now in StepPick (for the roles of mission
// 50, each seat that may take a role); empty in any other step. Never nil.
func (g *Game) Pickable() []int {
	out := []int{}
	a := g.Attempt
	if g.Phase != PhasePlaying || a.Step != StepPick {
		return out
	}
	it, m := g.pendingItem(), g.MissionDef()
	for s := range a.Members {
		if a.pickError(m, it, s) == nil {
			out = append(out, s)
		}
	}
	return out
}

// evenReachable reports whether, after giving the next task to seat s, the commander's
// distribution can still end with no member having two tasks more than another (rules §10).
// With counts c and r tasks left, that holds iff raising every member to max(c)-1 needs at most
// r tasks.
func (a *Attempt) evenReachable(s int) bool {
	counts := make([]int, len(a.Members))
	left := -1 // the task being given
	for _, t := range a.Tasks {
		if o := t.Owner(); o >= 0 {
			counts[o]++
		} else {
			left++
		}
	}
	counts[s]++
	top := slices.Max(counts)
	need := 0
	for _, c := range counts {
		need += max(0, top-1-c)
	}
	return need <= left
}

// DraftTurn is the seat whose turn it is to take a task in StepDraft: the commander first, then
// clockwise (rules §6).
func (a *Attempt) DraftTurn() int {
	taken := 0
	for _, t := range a.Tasks {
		if t.Owner() != OwnerUnassigned {
			taken++
		}
	}
	return (a.Commander + taken) % len(a.Members)
}

// PickTask takes a task in the open distribution, for the seat whose turn it is (the commander
// acts for the dummy).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn, ErrInvalidTask.
// Emits TaskPicked.
func (g *Game) PickTask(by PlayerID, taskID int) error {
	seat, err := g.requireStep(by, StepDraft)
	if err != nil {
		return err
	}
	a := g.Attempt
	turn := g.draftTurn()
	if a.actorFor(turn) != seat {
		return fmt.Errorf("%w: seat %d picks", ErrNotYourTurn, turn)
	}
	if err := a.takeError(g.MissionDef(), turn, taskID); err != nil {
		return err
	}
	a.Tasks[taskID].Objective.Owner = turn
	g.emit(TaskPicked{TaskID: taskID, Seat: turn})
	g.advanceSetup()
	g.commit()
	return nil
}
