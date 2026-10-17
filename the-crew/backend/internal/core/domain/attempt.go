package domain

import (
	"fmt"
	"slices"
)

// Step is what the attempt waits for while playing (see Actors for who acts).
type Step string

const (
	StepNone         Step = ""             // not playing
	StepAdjustTokens Step = "adjustTokens" // 23, 40: the commander swaps/moves tokens or skips
	StepPoll         Step = "poll"         // every human but the commander answers
	StepPick         Step = "pick"         // the commander picks seats
	StepDraft        Step = "draft"        // open distribution: the seat whose turn it is takes a task
	StepAssign       Step = "assign"       // Deep Sea: the captain gives each task to a member
	StepPredict      Step = "predict"      // Deep Sea: the owner of a prediction task notes the number
	StepPass         Step = "pass"         // distress signal: every human chooses a card to pass
	StepPlay         Step = "play"         // tricks
)

// Reason says how an attempt ended.
type Reason string

const (
	ReasonCompleted       Reason = "completed"
	ReasonTaskFailed      Reason = "taskFailed"      // a task card was won by someone else, or not at all
	ReasonTokenOrder      Reason = "tokenOrder"      // a token's order condition broke
	ReasonObjectiveFailed Reason = "objectiveFailed" // a special objective failed
	ReasonGaveUp          Reason = "gaveUp"
)

// AttemptResult is set when the attempt is over.
type AttemptResult struct {
	Success   bool
	Reason    Reason
	Task      int // the failed task's ID, else -1
	Objective int // the failed objective's index, else -1
	Trick     int // tricks played when the attempt ended
	Seat      int // taskFailed: who won the task card (-1 if unplayed); gaveUp: who gave up; else -1
}

// Task is a task card: an objective with a token (rules §6, §7). Its owner is Objective.Owner.
type Task struct {
	ID        int // its index in Attempt.Tasks (layout order)
	Objective Objective
	Token     Token
	FaceDown  bool // drawn but not revealed yet (commander's decision and distribution); Q2: for good
	// Deep Sea: the task card and its difficulty for this crew size; 0 for Planet Nine.
	CardID     int
	Difficulty int
}

// Card is the task's card (Planet Nine tasks name one card; not for Deep Sea task cards).
func (t Task) Card() Card { return t.Objective.Cards[0] }

// Owner is the seat that owns the task, or OwnerUnassigned.
func (t Task) Owner() int { return t.Objective.Owner }

// Trick is the trick in progress.
type Trick struct {
	Leader int
	Plays  []Play
}

// lead returns the led card, or nil when no card is on the table.
func (t Trick) lead() *Card {
	if len(t.Plays) == 0 {
		return nil
	}
	return &t.Plays[0].Card
}

// Attempt is one deal of a mission. Starting a new attempt replaces it.
type Attempt struct {
	Number          int      // 1-based attempt of this mission
	Members         []Member // seat order; 2 players: [p0, p1, dummy]
	Commander       int      // seat holding trump 4
	Step            Step
	Tasks           []Task            // layout order
	Objectives      []Objective       // the mission's special objectives, in MissionDef order
	Poll            *Poll             // the open or just answered poll
	Silenced        int               // mission 11: the silenced seat, -1 until picked (and in other missions)
	Comms           [][]Communication // per seat, in order; rapture of the deep allows several
	DistressDecided bool              // the commander sent or skipped the distress signal
	Pass            *Pass             // StepPass
	Exchange        *Exchange         // the last pass or draw
	HandoverUsed    bool
	TokensAdjusted  bool // StepAdjustTokens is done (adjusted or skipped)
	Trick           Trick
	Tricks          []TrickRecord // completed
	Result          *AttemptResult

	// Deep Sea.
	Comm     CommMode // the communication mode of this attempt (unfamiliar terrain: drawn)
	Terrain  *Card    // unfamiliar terrain: the colour card drawn before the deal
	TaskDeck []int    // task cards not on the table, in draw order (replacing takes from it)
	Passes   int      // passes in the draft
	Chosen   []int    // the members the captain chose to take the tasks (6, 10, 13, 14–16, 26)
	Stage    int      // Q1: 0 = stage A, 1 = stage B
	Redeals  int      // automatic re-deals of this attempt (not attempts)
}

// Poll collects the yes/no (well/ill) answers of the humans other than the commander.
type Poll struct {
	Answers []*bool // per seat; nil = not answered (or not asked)
}

// Direction is the direction of a distress pass (rules §3: left is the next seat).
type Direction string

const (
	DirLeft  Direction = "left"
	DirRight Direction = "right"
)

// Pass is a distress pass in progress.
type Pass struct {
	Direction Direction
	Chosen    []*Card // per seat; nil = not chosen (always nil for the dummy)
}

// ExchangeKind is what moved cards between hands.
type ExchangeKind string

const (
	ExchangePass ExchangeKind = "pass" // distress signal
	ExchangeDraw ExchangeKind = "draw" // mission 12
)

// Exchange is the last pass or draw. Each player may see only their own side.
type Exchange struct {
	Kind ExchangeKind
	Gave []*Card // per seat: the card that left the seat
	Got  []*Card // per seat: the card that arrived
}

// Position is the statement of a communication (rules §8).
type Position string

const (
	PosHighest Position = "highest"
	PosOnly    Position = "only"
	PosLowest  Position = "lowest"
	PosHidden  Position = "" // dead zone
)

// Communication is a member's communicated card. It is shown while the card is in the hand.
type Communication struct {
	Card     Card
	Position Position
	Trick    int // the trick it was made before
}

func (a *Attempt) clone() *Attempt {
	if a == nil {
		return nil
	}
	c := *a
	c.Members = make([]Member, len(a.Members))
	for i, m := range a.Members {
		c.Members[i] = m.clone()
	}
	if a.Tasks != nil {
		c.Tasks = make([]Task, len(a.Tasks))
		for i, t := range a.Tasks {
			t.Objective = t.Objective.clone()
			c.Tasks[i] = t
		}
	}
	if a.Objectives != nil {
		c.Objectives = make([]Objective, len(a.Objectives))
		for i, o := range a.Objectives {
			c.Objectives[i] = o.clone()
		}
	}
	if a.Poll != nil {
		c.Poll = &Poll{Answers: cloneBools(a.Poll.Answers)}
	}
	c.Comms = make([][]Communication, len(a.Comms))
	for i, cm := range a.Comms {
		c.Comms[i] = slices.Clone(cm)
	}
	c.Terrain = cloneCard(a.Terrain)
	c.TaskDeck = slices.Clone(a.TaskDeck)
	c.Chosen = slices.Clone(a.Chosen)
	if a.Pass != nil {
		c.Pass = &Pass{Direction: a.Pass.Direction, Chosen: cloneCardPtrs(a.Pass.Chosen)}
	}
	if a.Exchange != nil {
		c.Exchange = &Exchange{Kind: a.Exchange.Kind, Gave: cloneCardPtrs(a.Exchange.Gave), Got: cloneCardPtrs(a.Exchange.Got)}
	}
	c.Trick.Plays = slices.Clone(a.Trick.Plays)
	if a.Tricks != nil {
		c.Tricks = make([]TrickRecord, len(a.Tricks))
		for i, t := range a.Tricks {
			t.Plays = slices.Clone(t.Plays)
			c.Tricks[i] = t
		}
	}
	if a.Result != nil {
		r := *a.Result
		c.Result = &r
	}
	return &c
}

func cloneBools(bs []*bool) []*bool {
	if bs == nil {
		return nil
	}
	out := make([]*bool, len(bs))
	for i, b := range bs {
		if b != nil {
			v := *b
			out[i] = &v
		}
	}
	return out
}

func cloneCardPtrs(cs []*Card) []*Card {
	if cs == nil {
		return nil
	}
	out := make([]*Card, len(cs))
	for i, c := range cs {
		out[i] = cloneCard(c)
	}
	return out
}

// TotalTricks is the number of tricks in an attempt with the given number of players (rules §4).
func TotalTricks(players int) int {
	switch players {
	case 4:
		return 10
	case 5:
		return 8
	}
	return 13
}

// history returns the evaluation input of the attempt.
func (a *Attempt) history() History {
	humans := 0
	for _, m := range a.Members {
		if !m.Dummy {
			humans++
		}
	}
	return History{Tricks: a.Tricks, Total: TotalTricks(humans), Seats: len(a.Members), Captain: a.Commander}
}

// turn is the seat that plays next in the current trick.
func (a *Attempt) turn() int {
	return (a.Trick.Leader + len(a.Trick.Plays)) % len(a.Members)
}

// actorFor is the human who acts for seat s: the commander acts for the dummy.
func (a *Attempt) actorFor(s int) int {
	if a.Members[s].Dummy {
		return a.Commander
	}
	return s
}

// humans returns the seats of the human members.
func (a *Attempt) humans() []int {
	var out []int
	for i, m := range a.Members {
		if !m.Dummy {
			out = append(out, i)
		}
	}
	return out
}

// neighbour returns the next human seat from s in direction d, skipping the dummy.
func (a *Attempt) neighbour(s int, d Direction) int {
	n := len(a.Members)
	step := 1
	if d == DirRight {
		step = n - 1
	}
	for i := (s + step) % n; ; i = (i + step) % n {
		if !a.Members[i].Dummy {
			return i
		}
	}
}

// prePlay reports whether the first trick has not started yet: the window for the distress
// signal and the handover.
func (a *Attempt) prePlay() bool {
	return a.Step == StepPlay && len(a.Tricks) == 0 && len(a.Trick.Plays) == 0
}

// anyCommunicated reports whether a member has communicated in this attempt.
func (a *Attempt) anyCommunicated() bool {
	return slices.ContainsFunc(a.Comms, func(c []Communication) bool { return len(c) > 0 })
}

// startAttempt deals attempt number of the selected mission at stage (Q1), lays out the tasks
// (Deep Sea: keep, when set, are the task cards of the last attempt) and enters the setup. RNG
// use order: the terrain card (unfamiliar terrain), the deck (2 players: the dummy's deck, then
// the players' deck), the extra card's seat (3 players), the task deck, then any re-deals.
func (g *Game) startAttempt(number, stage int, keep []int, rng RNG) {
	base, _ := MissionFor(g.Edition, g.Mission)
	m := base.atStage(stage)
	a := &Attempt{Number: number, Silenced: -1, Stage: stage, Comm: m.Comm.Mode}
	if a.Comm == CommUnfamiliar {
		t := drawTerrain(rng)
		a.Terrain, a.Comm = &t, terrainMode(t)
	}
	a.Members = deal(g.Players, m.captainCard(), rng)
	a.setCommander(m)
	a.Comms = make([][]Communication, len(a.Members))

	if m.Tasks > 0 {
		deck := ColourCards()
		rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
		for i, c := range deck[:m.Tasks] {
			t := Task{ID: i, Objective: Objective{Kind: KindWinCards, Owner: OwnerUnassigned, Cards: []Card{c}}}
			if i < len(m.Tokens) {
				t.Token = m.Tokens[i]
			}
			t.FaceDown = m.Distribution != DistOpen
			a.Tasks = append(a.Tasks, t)
		}
	}
	if m.DrawsTasks() || m.FixedTasks != nil {
		g.dealTaskCards(a, m, keep, rng)
	}
	a.resolveObjectives(m)

	g.Phase = PhasePlaying
	g.Attempt = a
	g.emit(AttemptStarted{Mission: g.Mission, Number: number, Commander: a.Commander})
	g.redealIfNeeded(m, rng)
	g.advanceSetup()
}

// resolveObjectives lays out the mission's special objectives with the owners known at the deal.
func (a *Attempt) resolveObjectives(m MissionDef) {
	a.Objectives = nil
	for _, s := range m.Objectives {
		o := s.Objective.clone()
		switch s.Owner {
		case OwnerRuleCrew:
			o.Owner = OwnerCrew
		case OwnerRuleCommander:
			o.Owner = a.Commander
		case OwnerRuleLeftOfHolder:
			for i := range a.Members {
				if slices.Contains(a.Members[i].Cards(), s.Holder) {
					o.Owner = (i + 1) % len(a.Members)
				}
			}
		default:
			o.Owner = OwnerUnassigned
		}
		if s.TaskToken != TokenNone {
			for _, t := range a.Tasks {
				if t.Token == s.TaskToken {
					o.Cards = []Card{t.Card()}
				}
			}
		}
		a.Objectives = append(a.Objectives, o)
	}
}

// deal shuffles and deals the 40 cards to the players (rules §4, §12). With 2 players the dummy
// takes seat 2 and the captain card (trump 4; Deep Sea Q1: trump 1) is set aside for a human.
func deal(players []Player, captain Card, rng RNG) []Member {
	shuffle := func(cs []Card) {
		rng.Shuffle(len(cs), func(i, j int) { cs[i], cs[j] = cs[j], cs[i] })
	}
	n := len(players)
	members := make([]Member, n)
	for i, p := range players {
		members[i] = Member{PlayerID: p.ID, Hand: []Card{}}
	}
	if n == 2 {
		deck, _ := removeCard(NewDeck(), captain)
		shuffle(deck)
		dummy := Member{Dummy: true, Stacks: make([]DummyStack, DummyStacks)}
		for i := range DummyStacks {
			down, up := deck[i], deck[DummyStacks+i]
			dummy.Stacks[i] = DummyStack{Down: &down, Up: &up}
		}
		rest := append(slices.Clone(deck[2*DummyStacks:]), captain)
		shuffle(rest)
		half := len(rest) / 2
		members[0].Hand, members[1].Hand = rest[:half], rest[half:]
		members = append(members, dummy)
	} else {
		deck := NewDeck()
		shuffle(deck)
		sizes := make([]int, n)
		for i := range sizes {
			sizes[i] = len(deck) / n
		}
		if n == 3 {
			sizes[rng.IntN(n)]++
		}
		at := 0
		for i := range members {
			members[i].Hand = slices.Clone(deck[at : at+sizes[i]])
			at += sizes[i]
		}
	}
	for i := range members {
		if !members[i].Dummy {
			members[i].Hand = slices.Clone(members[i].Hand)
			SortCards(members[i].Hand)
		}
	}
	return members
}

// requireAttempt checks the phase and step and returns by's seat.
func (g *Game) requireStep(by PlayerID, step Step) (int, error) {
	seat, err := g.requireSeated(by)
	if err != nil {
		return -1, err
	}
	if g.Phase != PhasePlaying {
		return -1, fmt.Errorf("%w: no attempt is running", ErrWrongPhase)
	}
	if g.Attempt.Step != step {
		return -1, fmt.Errorf("%w: the attempt waits for %s", ErrWrongStep, g.Attempt.Step)
	}
	return seat, nil
}

// requireCommander checks that seat is the commander.
func (a *Attempt) requireCommander(seat int) error {
	if seat != a.Commander {
		return fmt.Errorf("%w: only the commander may do this", ErrNotCommander)
	}
	return nil
}

// LegalCards returns the cards seat may play now: during StepPlay on seat's turn, its playable
// cards; otherwise none. The result is never nil.
func (g *Game) LegalCards(seat int) []Card {
	a := g.Attempt
	if g.Phase != PhasePlaying || a.Step != StepPlay || seat < 0 || seat >= len(a.Members) || a.turn() != seat {
		return []Card{}
	}
	return a.Members[seat].Playable(a.Trick.lead())
}

// PlayCard plays card for the seat whose turn it is: its player, or the commander for the dummy.
// Completing a trick evaluates the tasks and objectives, which may end the attempt; in mission 12
// the first trick triggers the draw (rng).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn, ErrCardNotInHand,
// ErrMustFollowSuit. Emits CardPlayed, then on a full trick TrickWon (CardsDrawn, AttemptEnded).
func (g *Game) PlayCard(by PlayerID, card Card, rng RNG) error {
	seat, err := g.requireStep(by, StepPlay)
	if err != nil {
		return err
	}
	a := g.Attempt
	turn := a.turn()
	if a.actorFor(turn) != seat {
		return fmt.Errorf("%w: seat %d plays", ErrNotYourTurn, turn)
	}
	m := &a.Members[turn]
	if !m.holds(card) {
		return fmt.Errorf("%w: %s", ErrCardNotInHand, card)
	}
	if !slices.Contains(m.Playable(a.Trick.lead()), card) {
		return fmt.Errorf("%w: %s", ErrMustFollowSuit, card)
	}

	m.take(card)
	a.Trick.Plays = append(a.Trick.Plays, Play{Seat: turn, Card: card})
	g.emit(CardPlayed{Seat: turn, Card: card})
	if len(a.Trick.Plays) == len(a.Members) {
		g.completeTrick(rng)
	}
	g.commit()
	return nil
}

// completeTrick records the full trick, turns up the dummy's cards, and then ends the attempt or
// sets up the next trick.
func (g *Game) completeTrick(rng RNG) {
	a := g.Attempt
	cards := make([]Card, len(a.Trick.Plays))
	for i, p := range a.Trick.Plays {
		cards[i] = p.Card
	}
	winner := a.Trick.Plays[trickWinner(cards)].Seat
	a.Tricks = append(a.Tricks, TrickRecord{Leader: a.Trick.Leader, Plays: a.Trick.Plays, Winner: winner})
	a.Trick = Trick{Leader: winner}
	for i := range a.Members {
		a.Members[i].flip()
	}
	g.emit(TrickWon{Number: len(a.Tricks), Seat: winner})

	if r := a.evaluate(); r != nil {
		g.endAttempt(r)
		return
	}
	if len(a.Tricks) == 1 && g.MissionDef().Has(SetupDrawAfterFirstTrick) {
		a.draw(rng)
		g.emit(CardsDrawn{})
	}
}

// Statuses returns the status of every task (by ID) and every objective.
func (a *Attempt) Statuses() (tasks, objectives []Status) {
	h := a.history()
	tasks = make([]Status, len(a.Tasks))
	for i, t := range a.Tasks {
		tasks[i] = t.Objective.Status(h)
	}
	objectives = make([]Status, len(a.Objectives))
	for i, o := range a.Objectives {
		objectives[i] = o.Status(h)
	}
	return tasks, objectives
}

// evaluate checks the tasks, the token order and the objectives after a trick. It returns the
// result when the attempt is over (lost or completed), else nil.
func (a *Attempt) evaluate() *AttemptResult {
	h := a.history()
	lost := func(r Reason, task, obj, seat int) *AttemptResult {
		return &AttemptResult{Reason: r, Task: task, Objective: obj, Trick: len(a.Tricks), Seat: seat}
	}
	tasks, objectives := a.Statuses()
	for i, s := range tasks {
		if s == StatusFailed {
			seat := -1
			if t := a.Tasks[i]; t.CardID == 0 {
				seat = a.winnerOf(t.Card())
			}
			return lost(ReasonTaskFailed, i, -1, seat)
		}
	}
	tokens := make([]Token, len(a.Tasks))
	at := make([]int, len(a.Tasks))
	for i, t := range a.Tasks {
		tokens[i] = t.Token
		at[i] = t.Objective.FulfilledAt(h)
	}
	if i := orderViolation(tokens, at); i >= 0 {
		return lost(ReasonTokenOrder, i, -1, -1)
	}
	for i, s := range objectives {
		if s == StatusFailed {
			return lost(ReasonObjectiveFailed, -1, i, -1)
		}
	}
	if slices.ContainsFunc(tasks, func(s Status) bool { return s != StatusFulfilled }) ||
		slices.ContainsFunc(objectives, func(s Status) bool { return s != StatusFulfilled }) {
		return nil
	}
	return &AttemptResult{Success: true, Reason: ReasonCompleted, Task: -1, Objective: -1, Trick: len(a.Tricks), Seat: -1}
}

// winnerOf returns the seat that won the trick containing c, or -1 if c was not played.
func (a *Attempt) winnerOf(c Card) int {
	for _, t := range a.Tricks {
		if t.contains(c) {
			return t.Winner
		}
	}
	return -1
}

// endAttempt moves to the debrief and writes the logbook (attempts of a replay of a completed
// mission are not counted, nor those of the Deep Sea epilogue). With stages (Q1) the mission is
// completed when the last stage is won.
func (g *Game) endAttempt(r *AttemptResult) {
	a := g.Attempt
	a.Result = r
	a.Step = StepNone
	a.Poll, a.Pass = nil, nil
	g.Phase = PhaseDebrief
	m := g.MissionDef()
	if e := g.Logbook[g.Mission]; !e.Completed {
		if !m.Uncounted {
			e.Attempts++
		}
		e.Completed = r.Success && !a.hasNextStage(m)
		g.Logbook[g.Mission] = e
	}
	g.emit(AttemptEnded{Success: r.Success, Reason: r.Reason})
}
