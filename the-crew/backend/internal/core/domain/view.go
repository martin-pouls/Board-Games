package domain

import (
	"fmt"
	"maps"
	"slices"
	"time"
)

// PlayerView is the projection of a game for one player. Before the debrief it never exposes
// another human's hand, the dummy's hidden cards, the deck (nor the task deck), face-down tasks,
// another member's hidden prediction, other players' pass choices or another player's side of an
// exchange.
type PlayerView struct {
	GameID      GameID
	Revision    int64
	Phase       Phase
	You, HostID PlayerID
	Players     []PlayerSummary // humans, seat order
	Edition     Edition
	Options     Options
	Mission     int              // selected (lobby) or being played
	Logbook     map[int]LogEntry // per mission
	Attempt     *AttemptView     // nil in the lobby (and when ended from the lobby)
}

// PlayerSummary is the public information about a seated player.
type PlayerSummary struct {
	ID            PlayerID
	Name          string
	Seat          int
	IsHost, IsYou bool
}

// AttemptView is the running attempt (playing) or the one just finished (debrief).
type AttemptView struct {
	Number      int
	Seat        int // the viewer's seat
	Commander   int
	Step        Step
	Pick        SetupPick // StepPick: what the commander picks
	Actors      []int     // humans the attempt waits for
	Turn        int       // StepPlay: the seat to play; StepDraft: the seat to take a task; else -1
	TrickNumber int       // the trick in progress (1-based)
	TotalTricks int
	Members     []MemberView
	Hand        []Card // the viewer's hand, sorted
	Playable    []Card // the cards the viewer may play now: their own, or the dummy's when they act for it
	Tasks       []TaskView
	Objectives  []ObjectiveView
	Silenced    int   // mission 11: the silenced seat, else -1
	Pickable    []int // StepPick, for the commander: the seats they may pick
	Poll        *PollView
	Pass        *PassView
	Exchange    *ExchangeView // the viewer's side of the last pass or draw
	Trick       []Play        // the trick in progress, in play order
	TrickLeader int
	LastTrick   *TrickRecord // the last completed trick (public: everyone saw it played)
	Distress    bool         // the distress signal is active: used for this unfinished mission (rules §9)
	Actions     Actions      // what the viewer may do now besides the step's action
	Result      *AttemptResult

	// Deep Sea.
	Comm               CommMode // this attempt's communication mode
	Terrain            *Card    // unfamiliar terrain: the card that picked the mode
	TokensLeft         int      // rapture of the deep: shared sonar tokens left, else -1
	CommFirstTrickOnly bool     // 10, 13: communication ends with the first trick
	Chosen             []int    // the members the captain chose to take the tasks
	Stage, Stages      int      // Q1: the stage (0-based) and how many there are; Stages 0 without
	Redeals            int      // automatic re-deals before this attempt
	Takeable           []int    // StepDraft, for the member to take (or the captain for the dummy): the tasks they may take
	Assignable         [][]int  // StepAssign, for the captain: per task, the seats it may go to
	Predict            int      // StepPredict: the task being predicted, else -1
}

// MemberView is a seat in the attempt.
type MemberView struct {
	Seat           int
	PlayerID       PlayerID // "" for the dummy
	Name           string   // "" for the dummy
	Dummy          bool
	CardCount      int
	TricksWon      int
	Communication  *Communication  // the first communication shown (its card is still held)
	Communications []Communication // all shown (rapture of the deep allows several)
	Communicated   bool
	Stacks         []StackView // the dummy's stacks
	MustAct        bool
	Cards          []Card // debrief only: the cards still held (all hands are revealed)
}

// StackView is a dummy stack: the face-up card (nil once played) and whether a hidden card lies
// under it. The hidden card itself is never shown before the debrief.
type StackView struct {
	Up     *Card
	Hidden bool
}

// TaskView is a task. Card is nil while the task is face down (and for Deep Sea task cards).
type TaskView struct {
	ID       int
	Card     *Card
	Token    Token
	Owner    int // seat or OwnerUnassigned
	Status   Status
	FaceDown bool

	// Deep Sea: the task card (0 and nil while face down) and its difficulty for this crew size.
	// A hidden prediction (91) of another member shows Count -1 with Predicted true.
	CardID     int
	Difficulty int
	Objective  *Objective
	Predicted  bool
}

// ObjectiveView is a special objective with its status.
type ObjectiveView struct {
	Objective Objective
	Status    Status
}

// PollView is the open poll.
type PollView struct {
	Kind      PollKind
	Answers   []*bool // per seat; nil = not answered (always nil for the commander and the dummy)
	Objective int     // the objective whose owner is chosen, else -1
	Task      int     // commander's distribution: the revealed task, else -1
}

// PassView is the distress pass in progress.
type PassView struct {
	Direction Direction
	Chosen    *Card // the viewer's chosen card
	Waiting   []int // humans who have not chosen yet
}

// ExchangeView is the viewer's side of the last pass or draw.
type ExchangeView struct {
	Kind      ExchangeKind
	Gave, Got *Card
}

// Actions are what the viewer may do now, besides the step's action.
type Actions struct {
	Communicate  []CommOption // empty unless the viewer may communicate now
	SendDistress bool         // the viewer is the commander and may send or skip the signal
	Handover     bool         // the viewer is the commander and may hand over a task
	Pass         bool         // Deep Sea: the viewer may pass in the draft now
	ReplaceTasks bool         // Deep Sea: the viewer is the captain and may replace displayed tasks
}

// CommOption is a card the viewer may communicate, with the true statement (PosHidden in a dead
// zone).
type CommOption struct {
	Card     Card
	Position Position
}

// GameSummary is the public, pre-join information about a game.
type GameSummary struct {
	GameID      GameID
	Phase       Phase
	HostName    string
	PlayerCount int
	MaxPlayers  int
	Joinable    bool // lobby && PlayerCount < MaxPlayers
	CreatedAt   time.Time
}

// ViewFor builds the projection for player id. Errors: ErrUnknownPlayer.
func (g *Game) ViewFor(id PlayerID) (PlayerView, error) {
	_, seat, ok := g.PlayerByID(id)
	if !ok {
		return PlayerView{}, fmt.Errorf("%w: %s", ErrUnknownPlayer, id)
	}
	v := PlayerView{
		GameID:   g.ID,
		Revision: g.Revision,
		Phase:    g.Phase,
		You:      id,
		HostID:   g.HostID,
		Players:  make([]PlayerSummary, len(g.Players)),
		Edition:  g.Edition,
		Options:  g.Options,
		Mission:  g.Mission,
		Logbook:  maps.Clone(g.Logbook),
	}
	for i, p := range g.Players {
		v.Players[i] = PlayerSummary{ID: p.ID, Name: p.Name, Seat: i, IsHost: p.ID == g.HostID, IsYou: p.ID == id}
	}
	if g.Attempt != nil && (g.Phase == PhasePlaying || g.Phase == PhaseDebrief || g.Phase == PhaseEnded) {
		v.Attempt = g.attemptView(seat)
	}
	return v, nil
}

// attemptView builds the attempt for the viewer in seat.
func (g *Game) attemptView(seat int) *AttemptView {
	a := g.Attempt
	m := g.MissionDef()
	h := a.history()
	over := a.Result != nil
	actors := g.Actors()
	v := &AttemptView{
		Number:      a.Number,
		Seat:        seat,
		Commander:   a.Commander,
		Step:        a.Step,
		Pick:        g.Pick(),
		Actors:      actors,
		Turn:        -1,
		TrickNumber: len(a.Tricks) + 1,
		TotalTricks: h.Total,
		Hand:        cloneCards(a.Members[seat].Hand),
		Playable:    []Card{},
		Silenced:    a.Silenced,
		Pickable:    []int{},
		Trick:       slices.Clone(a.Trick.Plays),
		TrickLeader: a.Trick.Leader,
		Distress:    g.distressActive(),
		Result:      cloneResult(a.Result),
		Comm:        a.Comm,
		Terrain:     cloneCard(a.Terrain),
		TokensLeft:  a.TokensLeft(),
		Chosen:      slices.Clone(a.Chosen),
		Stage:       a.Stage,
		Redeals:     a.Redeals,
		Takeable:    []int{},
		Assignable:  [][]int{},
		Predict:     -1,
	}
	if base, ok := MissionFor(g.Edition, g.Mission); ok {
		v.Stages = len(base.Stages)
	}
	v.CommFirstTrickOnly = g.onlyBeforeFirstTrick()
	if v.Trick == nil {
		v.Trick = []Play{}
	}
	if seat == a.Commander {
		v.Pickable = g.Pickable()
	}
	switch a.Step {
	case StepPlay:
		v.Turn = a.turn()
		if a.actorFor(v.Turn) == seat {
			v.Playable = g.LegalCards(v.Turn)
		}
	case StepDraft:
		v.Turn = g.draftTurn()
		if a.actorFor(v.Turn) == seat {
			v.Takeable = g.Takeable()
		}
	case StepAssign:
		if seat == a.Commander {
			v.Assignable = g.Assignable()
		}
	case StepPredict:
		v.Predict = g.pendingItem().index
	}
	if n := len(a.Tricks); n > 0 {
		last := a.Tricks[n-1]
		last.Plays = slices.Clone(last.Plays)
		v.LastTrick = &last
	}

	won := make([]int, len(a.Members))
	for _, t := range a.Tricks {
		won[t.Winner]++
	}
	v.Members = make([]MemberView, len(a.Members))
	for i := range a.Members {
		mem := &a.Members[i]
		mv := MemberView{
			Seat:         i,
			PlayerID:     mem.PlayerID,
			Dummy:        mem.Dummy,
			CardCount:    mem.CardCount(),
			TricksWon:    won[i],
			Communicated: len(a.Comms[i]) > 0,
			MustAct:      slices.Contains(actors, i),
		}
		if p, _, ok := g.PlayerByID(mem.PlayerID); ok {
			mv.Name = p.Name
		}
		mv.Communications = a.CommunicatedCards(i)
		if len(mv.Communications) > 0 {
			c := mv.Communications[0]
			mv.Communication = &c
		}
		if mem.Dummy {
			mv.Stacks = make([]StackView, len(mem.Stacks))
			for k, s := range mem.Stacks {
				mv.Stacks[k] = StackView{Up: cloneCard(s.Up), Hidden: s.Down != nil}
			}
		}
		if over {
			mv.Cards = mem.Cards()
		}
		v.Members[i] = mv
	}

	statuses, objStatuses := a.Statuses()
	v.Tasks = make([]TaskView, len(a.Tasks))
	for i, t := range a.Tasks {
		tv := TaskView{ID: t.ID, Token: t.Token, Owner: t.Owner(), Status: statuses[i], FaceDown: t.FaceDown, Difficulty: t.Difficulty}
		switch {
		case t.FaceDown && !over:
		case t.CardID == 0:
			c := t.Card()
			tv.Card = &c
		default:
			o := t.Objective.clone()
			tv.CardID, tv.Objective = t.CardID, &o
			if o.Kind == KindPredict {
				tv.Predicted = o.Count >= 0
				if o.Hidden && !over && a.actorFor(t.Owner()) != seat {
					o.Count = -1
				}
			}
		}
		v.Tasks[i] = tv
	}
	v.Objectives = make([]ObjectiveView, len(a.Objectives))
	for i, o := range a.Objectives {
		v.Objectives[i] = ObjectiveView{Objective: o.clone(), Status: objStatuses[i]}
	}

	if a.Poll != nil {
		pv := &PollView{Kind: m.Poll, Answers: cloneBools(a.Poll.Answers), Objective: -1, Task: -1}
		switch it := a.nextItem(m); it.kind {
		case itemChosen:
			pv.Objective = it.index
		case itemDistribute:
			pv.Task = it.index
		}
		v.Poll = pv
	}
	if a.Pass != nil {
		pv := &PassView{Direction: a.Pass.Direction, Chosen: cloneCard(a.Pass.Chosen[seat])}
		for _, s := range a.humans() {
			if a.Pass.Chosen[s] == nil {
				pv.Waiting = append(pv.Waiting, s)
			}
		}
		v.Pass = pv
	}
	if ex := a.Exchange; ex != nil && (ex.Gave[seat] != nil || ex.Got[seat] != nil) {
		v.Exchange = &ExchangeView{Kind: ex.Kind, Gave: cloneCard(ex.Gave[seat]), Got: cloneCard(ex.Got[seat])}
	}

	if !over {
		if g.CanCommunicate(seat) == nil {
			for _, c := range a.Members[seat].Hand {
				if pos, ok := StatementFor(a.Members[seat].Hand, c); ok && !a.communicates(seat, c) {
					if g.deadZone() {
						pos = PosHidden
					}
					v.Actions.Communicate = append(v.Actions.Communicate, CommOption{Card: c, Position: pos})
				}
			}
		}
		v.Actions.SendDistress = g.CanSendDistress(seat) == nil
		v.Actions.Handover = g.CanHandover(seat) == nil
		v.Actions.ReplaceTasks = g.CanReplaceTasks(seat) == nil
		v.Actions.Pass = a.Step == StepDraft && a.actorFor(g.draftTurn()) == seat && g.CanPass() == nil
	}
	return v
}

func cloneResult(r *AttemptResult) *AttemptResult {
	if r == nil {
		return nil
	}
	c := *r
	return &c
}
