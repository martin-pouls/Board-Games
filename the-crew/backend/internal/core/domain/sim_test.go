package domain

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"testing"
)

// The simulation plays random legal actions, drawn only from the advertised options (Actors,
// LegalCards, Actions.Communicate, CanX == nil), for every mission, crew size and several seeds,
// until the attempt ends. After every action it checks the invariants of plan §9: no panic, the 40
// cards are conserved, the actors are never empty, the views hide what plan §5.7 says, the events
// carry no hidden card, and Revision grows by exactly 1 per successful mutator.

const simMaxActions = 500

// simAbort is the panic value that ends one seed after an error was reported.
type simAbort struct{}

// simStats collects the outcomes of all runs.
type simStats struct {
	mu       sync.Mutex
	attempts int
	won      int
	lost     map[Reason]int
	steps    map[Step]int
	actions  int
	distress int
	handover int
	comm     int
	followUp map[string]int
}

func (st *simStats) step(s Step) {
	st.mu.Lock()
	st.steps[s]++
	st.actions++
	st.mu.Unlock()
}

func (st *simStats) count(f func()) {
	st.mu.Lock()
	f()
	st.mu.Unlock()
}

// simRun is one simulated game.
type simRun struct {
	t       *testing.T
	st      *simStats
	g       *Game
	rng     *rand.Rand
	tag     string
	log     []string
	distres bool // the distress signal was sent in the running attempt
}

func (s *simRun) failf(format string, args ...any) {
	s.t.Helper()
	last := s.log
	if len(last) > 12 {
		last = last[len(last)-12:]
	}
	s.t.Errorf("%s: %s\n    last actions: %s", s.tag, fmt.Sprintf(format, args...), strings.Join(last, " | "))
	panic(simAbort{})
}

// try runs a mutator. On error it asserts that nothing changed and returns the error; on success
// it asserts Revision+1 and checks the events and the invariants.
func (s *simRun) try(name string, fn func() error) error {
	s.t.Helper()
	g := s.g
	before := g.Clone()
	rev := g.Revision
	err := fn()
	if err != nil {
		s.log = append(s.log, name+" -> "+err.Error())
		if !reflect.DeepEqual(before, g) {
			s.failf("%s failed (%v) but changed the state", name, err)
		}
		return err
	}
	s.log = append(s.log, name)
	if g.Revision != rev+1 {
		s.failf("%s: revision %d -> %d, want +1", name, rev, g.Revision)
	}
	s.checkEvents(name)
	s.checkInvariants(name)
	return nil
}

// act runs a mutator that the random policy chose from the advertised options: it must succeed.
func (s *simRun) act(name string, fn func() error) {
	s.t.Helper()
	if err := s.try(name, fn); err != nil {
		s.failf("advertised action %s failed: %v", name, err)
	}
}

// ---- invariants ----

var cardType = reflect.TypeFor[Card]()

// countCards returns the number of Card values reachable from v.
func countCards(v reflect.Value) int {
	if v.Type() == cardType {
		return 1
	}
	n := 0
	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			n += countCards(v.Field(i))
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			n += countCards(v.Index(i))
		}
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			n += countCards(v.Elem())
		}
	}
	return n
}

func (s *simRun) checkEvents(name string) {
	s.t.Helper()
	g := s.g
	for _, e := range g.PullEvents() {
		switch ev := e.(type) {
		case CardPlayed:
			if g.Attempt != nil {
				for i := range g.Attempt.Members {
					if slices.Contains(g.Attempt.Members[i].Cards(), ev.Card) {
						s.failf("%s: event %+v: the played card is still held by seat %d", name, ev, i)
					}
				}
			}
		case Communicated:
			if !slices.Contains(g.Attempt.Members[ev.Seat].Hand, ev.Card) {
				s.failf("%s: event %+v: the communicated card is not in seat %d's hand", name, ev, ev.Seat)
			}
		default:
			if n := countCards(reflect.ValueOf(e)); n > 0 {
				s.failf("%s: event %T %+v carries %d card(s)", name, e, e, n)
			}
		}
	}
}

// checkCards verifies that the 40 cards are all in the attempt exactly once.
func (s *simRun) checkCards(name string) {
	s.t.Helper()
	a := s.g.Attempt
	seen := map[Card]int{}
	for i := range a.Members {
		for _, c := range a.Members[i].Cards() {
			seen[c]++
		}
	}
	for _, p := range a.Trick.Plays {
		seen[p.Card]++
	}
	for _, tr := range a.Tricks {
		for _, p := range tr.Plays {
			seen[p.Card]++
		}
	}
	total := 0
	for c, n := range seen {
		total += n
		if !c.Valid() || n != 1 {
			s.failf("%s: card %v occurs %d times", name, c, n)
		}
	}
	if total != 40 || len(seen) != 40 {
		s.failf("%s: %d cards (%d distinct), want 40", name, total, len(seen))
	}
}

func (s *simRun) checkInvariants(name string) {
	s.t.Helper()
	g := s.g
	if g.Attempt == nil {
		return
	}
	if g.Phase == PhasePlaying || g.Phase == PhaseDebrief {
		s.checkCards(name)
	}
	if g.Phase == PhasePlaying {
		if len(g.Actors()) == 0 {
			s.failf("%s: no actors in step %s", name, g.Attempt.Step)
		}
		if g.Attempt.Result != nil {
			s.failf("%s: result set while playing", name)
		}
	}
	if g.Phase == PhaseDebrief {
		if g.Attempt.Result == nil {
			s.failf("%s: debrief without a result", name)
		}
		if len(g.Actors()) != 0 {
			s.failf("%s: actors in the debrief", name)
		}
	}
	for seat, p := range g.Players {
		v, err := g.ViewFor(p.ID)
		if err != nil {
			s.failf("%s: view for %s: %v", name, p.ID, err)
		}
		s.checkView(name, seat, v)
	}
}

func subset(xs, of []Card) bool {
	for _, x := range xs {
		if !slices.Contains(of, x) {
			return false
		}
	}
	return true
}

func intersects(xs, ys []Card) bool {
	for _, x := range xs {
		if slices.Contains(ys, x) {
			return true
		}
	}
	return false
}

func ptrCards(cs ...*Card) []Card {
	var out []Card
	for _, c := range cs {
		if c != nil {
			out = append(out, *c)
		}
	}
	return out
}

func sameCard(a, b *Card) bool {
	return (a == nil) == (b == nil) && (a == nil || *a == *b)
}

// checkView verifies the projection for the viewer in seat against the real state and against
// the hidden-information rules of plan §5.7.
func (s *simRun) checkView(name string, seat int, v PlayerView) {
	s.t.Helper()
	g, a := s.g, s.g.Attempt
	if v.Revision != g.Revision || v.Phase != g.Phase {
		s.failf("%s: view of seat %d: revision/phase mismatch", name, seat)
	}
	av := v.Attempt
	if av == nil {
		s.failf("%s: view of seat %d has no attempt", name, seat)
	}
	over := a.Result != nil
	mine := a.Members[seat].Hand
	if av.Seat != seat || !slices.Equal(av.Hand, mine) {
		s.failf("%s: view of seat %d: hand %v, want %v", name, seat, av.Hand, mine)
	}
	if !sameResult(av.Result, a.Result) {
		s.failf("%s: view of seat %d: result %+v, want %+v", name, seat, av.Result, a.Result)
	}

	// Playable: only the viewer's own cards, or the dummy's face-up cards when acting for it.
	allowed := slices.Clone(mine)
	dummyTurn := false
	if a.Step == StepPlay && a.Members[a.turn()].Dummy && a.actorFor(a.turn()) == seat {
		dummyTurn = true
		for i := range a.Members {
			if a.Members[i].Dummy {
				allowed = append(allowed, a.Members[i].FaceUp()...)
			}
		}
	}
	if !subset(av.Playable, allowed) {
		s.failf("%s: view of seat %d: playable %v outside %v", name, seat, av.Playable, allowed)
	}
	if a.Step == StepPlay && a.actorFor(a.turn()) == seat {
		if len(av.Playable) == 0 {
			s.failf("%s: view of seat %d: nothing playable on its turn", name, seat)
		}
	} else if len(av.Playable) != 0 {
		s.failf("%s: view of seat %d: playable cards %v outside its turn", name, seat, av.Playable)
	}

	// Members.
	var hidden []Card // the dummy's hidden cards
	var others []Card // the other humans' hands
	if len(av.Members) != len(a.Members) {
		s.failf("%s: view of seat %d: %d members, want %d", name, seat, len(av.Members), len(a.Members))
	}
	for i := range a.Members {
		mem, mv := &a.Members[i], av.Members[i]
		if mv.Seat != i || mv.Dummy != mem.Dummy || mv.CardCount != mem.CardCount() || mv.PlayerID != mem.PlayerID {
			s.failf("%s: view of seat %d: member %d mismatch %+v", name, seat, i, mv)
		}
		if over {
			if !slices.Equal(mv.Cards, mem.Cards()) {
				s.failf("%s: debrief view of seat %d: member %d cards %v, want %v", name, seat, i, mv.Cards, mem.Cards())
			}
		} else if len(mv.Cards) != 0 {
			s.failf("%s: view of seat %d reveals member %d's cards %v before the debrief", name, seat, i, mv.Cards)
		}
		cm := a.CommunicatedCard(i)
		if (cm == nil) != (mv.Communication == nil) || (cm != nil && *cm != *mv.Communication) {
			s.failf("%s: view of seat %d: member %d communication %+v, want %+v", name, seat, i, mv.Communication, cm)
		}
		if mem.Dummy {
			if len(mv.Stacks) != DummyStacks {
				s.failf("%s: view of seat %d: %d stacks", name, seat, len(mv.Stacks))
			}
			for k, st := range mem.Stacks {
				sv := mv.Stacks[k]
				if !sameCard(sv.Up, st.Up) || sv.Hidden != (st.Down != nil) {
					s.failf("%s: view of seat %d: stack %d is %+v, want up=%v hidden=%v", name, seat, k, sv, st.Up, st.Down != nil)
				}
				if st.Down != nil {
					hidden = append(hidden, *st.Down)
				}
			}
		} else {
			if len(mv.Stacks) != 0 {
				s.failf("%s: view of seat %d: human %d has stacks", name, seat, i)
			}
			if i != seat {
				others = append(others, mem.Hand...)
			}
		}
	}
	_ = dummyTurn

	// Tasks.
	if len(av.Tasks) != len(a.Tasks) {
		s.failf("%s: view of seat %d: %d tasks, want %d", name, seat, len(av.Tasks), len(a.Tasks))
	}
	for i, t := range a.Tasks {
		tv := av.Tasks[i]
		if t.FaceDown && !over {
			if tv.Card != nil {
				s.failf("%s: view of seat %d shows the face-down task %d (%v)", name, seat, i, *tv.Card)
			}
		} else if tv.Card == nil || *tv.Card != t.Card() {
			s.failf("%s: view of seat %d: task %d card %v, want %v", name, seat, i, tv.Card, t.Card())
		}
	}

	// Exchange: only the viewer's side.
	if ex := a.Exchange; ex != nil && (ex.Gave[seat] != nil || ex.Got[seat] != nil) {
		ev := av.Exchange
		if ev == nil || ev.Kind != ex.Kind || !sameCard(ev.Gave, ex.Gave[seat]) || !sameCard(ev.Got, ex.Got[seat]) {
			s.failf("%s: view of seat %d: exchange %+v, want gave %v got %v", name, seat, ev, ex.Gave[seat], ex.Got[seat])
		}
	} else if av.Exchange != nil {
		s.failf("%s: view of seat %d: exchange %+v without a side of its own", name, seat, av.Exchange)
	}

	// Pass: only the viewer's own choice.
	if a.Pass == nil {
		if av.Pass != nil {
			s.failf("%s: view of seat %d: pass without a pass", name, seat)
		}
	} else {
		if av.Pass == nil || !sameCard(av.Pass.Chosen, a.Pass.Chosen[seat]) || av.Pass.Direction != a.Pass.Direction {
			s.failf("%s: view of seat %d: pass %+v, want chosen %v", name, seat, av.Pass, a.Pass.Chosen[seat])
		}
	}

	// Nothing else of the viewer's reach may contain another human's card or the dummy's hidden
	// cards. (Tasks, the trick, communications and the exchange's Gave card are public or the
	// viewer's own.)
	reach := slices.Clone(av.Hand)
	reach = append(reach, av.Playable...)
	if av.Exchange != nil {
		reach = append(reach, ptrCards(av.Exchange.Got)...)
	}
	if av.Pass != nil {
		reach = append(reach, ptrCards(av.Pass.Chosen)...)
	}
	for _, m := range av.Members {
		reach = append(reach, m.Cards...)
		for _, st := range m.Stacks {
			reach = append(reach, ptrCards(st.Up)...)
		}
	}
	if !over {
		if intersects(reach, hidden) {
			s.failf("%s: view of seat %d reveals a hidden dummy card", name, seat)
		}
		if av.Exchange != nil && intersects(ptrCards(av.Exchange.Gave), hidden) {
			s.failf("%s: view of seat %d reveals a hidden dummy card in the exchange", name, seat)
		}
		if intersects(reach, others) {
			s.failf("%s: view of seat %d reveals another human's hand card", name, seat)
		}
	}

	// Poll: the commander and the dummy never answer.
	if av.Poll != nil {
		for i, ans := range av.Poll.Answers {
			if ans != nil && (i == a.Commander || a.Members[i].Dummy) {
				s.failf("%s: view of seat %d: poll answer of seat %d", name, seat, i)
			}
		}
	}

	// Actions.
	if v.Options != g.Options {
		s.failf("%s: view of seat %d: options %+v, game %+v", name, seat, v.Options, g.Options)
	}
	if !g.Options.Distress && av.Actions.SendDistress {
		s.failf("%s: view of seat %d offers the distress signal while it is off", name, seat)
	}
	if over {
		if len(av.Actions.Communicate) != 0 || av.Actions.SendDistress || av.Actions.Handover {
			s.failf("%s: debrief view of seat %d offers actions %+v", name, seat, av.Actions)
		}
	}
	for _, o := range av.Actions.Communicate {
		if !slices.Contains(mine, o.Card) {
			s.failf("%s: view of seat %d offers to communicate %v, not in hand", name, seat, o.Card)
		}
	}
	if len(av.Actions.Communicate) > 0 && g.CanCommunicate(seat) != nil {
		s.failf("%s: view of seat %d offers communication that is not allowed", name, seat)
	}
}

func sameResult(a, b *AttemptResult) bool {
	return (a == nil) == (b == nil) && (a == nil || *a == *b)
}

// ---- the random policy ----

func (s *simRun) humanSeats() []int { return s.g.Attempt.humans() }

func (s *simRun) pid(seat int) PlayerID { return s.g.Players[seat].ID }

// step performs one random legal action.
func (s *simRun) step() {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	actors := g.Actors()
	if len(actors) == 0 {
		s.failf("no actors in step %s", a.Step)
	}
	s.st.step(a.Step)

	if rng.IntN(400) == 0 {
		seat := rng.IntN(len(g.Players))
		s.act(fmt.Sprintf("giveUp(%d)", seat), func() error { return g.GiveUp(s.pid(seat)) })
		return
	}
	actor := actors[rng.IntN(len(actors))]
	id := s.pid(actor)

	switch a.Step {
	case StepAdjustTokens:
		if rng.IntN(3) > 0 && len(a.Tasks) > 1 {
			from, to := rng.IntN(len(a.Tasks)), rng.IntN(len(a.Tasks))
			if s.try(fmt.Sprintf("adjustTokens(%d,%d)", from, to), func() error { return g.AdjustTokens(id, from, to) }) == nil {
				return
			}
		}
		s.act("skipTokens", func() error { return g.SkipTokens(id) })

	case StepPoll:
		yes := rng.IntN(2) == 0
		s.act(fmt.Sprintf("answerPoll(%d,%v)", actor, yes), func() error { return g.AnswerPoll(id, yes) })

	case StepPick:
		s.pick(actor, id)

	case StepDraft:
		var open []int
		for i, t := range a.Tasks {
			if t.Owner() == OwnerUnassigned {
				open = append(open, i)
			}
		}
		if len(open) == 0 {
			s.failf("draft without an open task")
		}
		task := open[rng.IntN(len(open))]
		s.act(fmt.Sprintf("pickTask(%d by %d)", task, actor), func() error { return g.PickTask(id, task) })

	case StepPass:
		var cands []Card
		for _, c := range a.Members[actor].Hand {
			if !c.IsTrump() {
				cands = append(cands, c)
			}
		}
		if len(cands) == 0 {
			s.failf("seat %d has no card to pass", actor)
		}
		card := cands[rng.IntN(len(cands))]
		s.act(fmt.Sprintf("choosePassCard(%d,%v)", actor, card), func() error { return g.ChoosePassCard(id, card) })

	case StepPlay:
		s.play()

	default:
		s.failf("no action for step %q", a.Step)
	}
}

func (s *simRun) pick(actor int, id PlayerID) {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	seats := len(a.Members)
	pick := g.Pick()
	switch pick {
	case PickSilence:
		hs := s.humanSeats()
		seat := hs[rng.IntN(len(hs))]
		s.act(fmt.Sprintf("pick(silence %d)", seat), func() error { return g.PickSeats(id, []int{seat}) })
	case PickRoles:
		x := rng.IntN(seats)
		y := (x + 1 + rng.IntN(seats-1)) % seats
		s.act(fmt.Sprintf("pick(roles %d,%d)", x, y), func() error { return g.PickSeats(id, []int{x, y}) })
	case PickObjective, PickDecision:
		seat := (a.Commander + 1 + rng.IntN(seats-1)) % seats
		s.act(fmt.Sprintf("pick(%s %d)", pick, seat), func() error { return g.PickSeats(id, []int{seat}) })
	case PickDistribute:
		for _, seat := range rng.Perm(seats) {
			if s.try(fmt.Sprintf("pick(distribute %d)", seat), func() error { return g.PickSeats(id, []int{seat}) }) == nil {
				return
			}
		}
		s.failf("no seat can receive the task: PickSeats failed for all %d seats", seats)
	default:
		s.failf("pick step with pick %q", pick)
	}
}

func (s *simRun) play() {
	s.t.Helper()
	g, rng := s.g, s.rng
	a := g.Attempt
	cmd := a.Commander
	cid := s.pid(cmd)

	if a.prePlay() {
		if !g.Options.Distress && a.Step == StepPlay && rng.IntN(4) == 0 {
			// Off: the commander's attempts to send or skip are rejected and change nothing.
			if err := s.try("sendDistress(off)", func() error { return g.SendDistress(cid, DirLeft) }); !errors.Is(err, ErrNotAllowed) {
				s.failf("sendDistress while off: %v, want ErrNotAllowed", err)
			}
			if err := s.try("skipDistress(off)", func() error { return g.SkipDistress(cid) }); !errors.Is(err, ErrNotAllowed) {
				s.failf("skipDistress while off: %v, want ErrNotAllowed", err)
			}
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
				s.st.count(func() { s.st.distress++ })
				return
			case 2:
				s.act("skipDistress", func() error { return g.SkipDistress(cid) })
			}
		}
		if g.CanHandover(cmd) == nil && rng.IntN(3) == 0 && len(a.Tasks) > 0 {
			task := rng.IntN(len(a.Tasks))
			from := a.Tasks[task].Owner()
			to := rng.IntN(len(a.Members))
			if to == from {
				to = (to + 1) % len(a.Members)
			}
			s.act(fmt.Sprintf("handover(%d->%d)", task, to), func() error { return g.Handover(cid, task, to) })
			s.st.count(func() { s.st.handover++ })
		}
	}

	if len(a.Trick.Plays) == 0 && rng.IntN(3) == 0 {
		type cand struct {
			seat int
			opt  CommOption
		}
		var cands []cand
		for _, seat := range s.humanSeats() {
			v, err := g.ViewFor(s.pid(seat))
			if err != nil {
				s.failf("view: %v", err)
			}
			for _, o := range v.Attempt.Actions.Communicate {
				cands = append(cands, cand{seat, o})
			}
		}
		if len(cands) > 0 {
			c := cands[rng.IntN(len(cands))]
			s.act(fmt.Sprintf("communicate(%d,%v,%q)", c.seat, c.opt.Card, c.opt.Position),
				func() error { return g.Communicate(s.pid(c.seat), c.opt.Card, c.opt.Position) })
			s.st.count(func() { s.st.comm++ })
		}
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
	card := legal[rng.IntN(len(legal))]
	s.act(fmt.Sprintf("play(seat %d by %d: %v)", turn, actor, card), func() error { return g.PlayCard(s.pid(actor), card, rng) })
}

// ---- driving a game ----

// runAttempt plays the running attempt to the debrief and checks the logbook. replay is the
// expectation that the mission was completed before.
func (s *simRun) runAttempt(wantNumber int) {
	s.t.Helper()
	g := s.g
	mission := g.Mission
	e0 := g.Logbook[mission]
	if g.Phase != PhasePlaying || g.Attempt == nil {
		s.failf("attempt did not start: phase %s", g.Phase)
	}
	if g.Attempt.Number != wantNumber {
		s.failf("attempt number %d, want %d", g.Attempt.Number, wantNumber)
	}
	s.distres = false
	s.checkEvents("start")
	s.checkInvariants("start")

	for n := 0; g.Phase == PhasePlaying; n++ {
		if n >= simMaxActions {
			s.failf("the attempt does not end after %d actions (step %s)", n, g.Attempt.Step)
		}
		s.step()
	}
	if g.Phase != PhaseDebrief {
		s.failf("attempt ended in phase %s", g.Phase)
	}
	r := g.Attempt.Result
	if r == nil {
		s.failf("debrief without a result")
	}
	if r.Success != (r.Reason == ReasonCompleted) {
		s.failf("result %+v: success and reason disagree", r)
	}
	if r.Success && len(g.Attempt.Tricks) == 0 {
		s.failf("won without a trick")
	}
	s.st.count(func() {
		s.st.attempts++
		if r.Success {
			s.st.won++
		} else {
			s.st.lost[r.Reason]++
		}
	})

	e1 := g.Logbook[mission]
	if e0.Completed {
		if e1 != e0 {
			s.failf("a replay changed the logbook %+v -> %+v", e0, e1)
		}
	} else {
		want := LogEntry{Attempts: e0.Attempts + 1, Distress: e0.Distress || s.distres, Completed: r.Success}
		if e1 != want {
			s.failf("logbook %+v, want %+v", e1, want)
		}
	}
}

// followUp exercises the debrief transitions once, depending on the outcome and the variant.
func (s *simRun) followUp(variant int) {
	s.t.Helper()
	g, rng := s.g, s.rng
	r := g.Attempt.Result
	host := alice
	mission := g.Mission
	key := func(k string) { s.st.count(func() { s.st.followUp[k]++ }) }

	// Wrong-outcome and non-host calls are rejected without a change.
	if r.Success {
		s.try("nextAttempt(won)", func() error { return g.NextAttempt(host, rng) })
		if g.Phase != PhaseDebrief {
			s.failf("NextAttempt after a win left the debrief")
		}
	} else {
		s.try("nextMission(lost)", func() error { return g.NextMission(host, rng) })
		if g.Phase != PhaseDebrief {
			s.failf("NextMission after a loss left the debrief")
		}
	}
	if err := s.try("backToLobby(bob)", func() error { return g.BackToLobby(bob) }); err == nil {
		s.failf("a non-host sent the crew back to the lobby")
	}

	startNumber := func() int {
		e := g.Logbook[g.Mission]
		if e.Completed {
			return 1
		}
		return e.Attempts + 1
	}
	if !r.Success {
		switch variant {
		case 0:
			key("lost:nextAttempt")
			want := g.Attempt.Number + 1
			s.act("nextAttempt", func() error { return g.NextAttempt(host, rng) })
			s.runAttempt(want)
		case 1:
			key("lost:backToLobby")
			s.act("backToLobby", func() error { return g.BackToLobby(host) })
			if g.Phase != PhaseLobby || g.Attempt != nil || g.Mission != mission {
				s.failf("back in the lobby: phase %s mission %d attempt %v", g.Phase, g.Mission, g.Attempt)
			}
			s.act("startMission", func() error { return g.StartMission(host, rng) })
			s.runAttempt(startNumber())
		case 2:
			key("lost:endGame")
			s.act("endGame", func() error { return g.EndGame(host) })
			s.checkEnded()
		}
		return
	}
	switch variant {
	case 0:
		if mission < Missions(g.Edition) {
			key("won:nextMission")
			s.act("nextMission", func() error { return g.NextMission(host, rng) })
			if g.Mission != mission+1 {
				s.failf("next mission is %d, want %d", g.Mission, mission+1)
			}
			s.runAttempt(startNumber())
		} else {
			key("won:backToLobby(last)")
			s.act("backToLobby", func() error { return g.BackToLobby(host) })
			if g.Phase != PhaseLobby || g.Mission < 1 || g.Mission > Missions(g.Edition) {
				s.failf("back in the lobby: phase %s mission %d", g.Phase, g.Mission)
			}
		}
	case 1:
		key("won:replay")
		s.act("backToLobby", func() error { return g.BackToLobby(host) })
		if want := g.nextUnfinished(); g.Mission != want {
			s.failf("lobby mission %d, want the next unfinished one %d", g.Mission, want)
		}
		s.act("selectMission", func() error { return g.SelectMission(host, mission) })
		s.act("startMission", func() error { return g.StartMission(host, rng) })
		s.runAttempt(1)
	case 2:
		key("won:endGame")
		s.act("endGame", func() error { return g.EndGame(host) })
		s.checkEnded()
	}
}

func (s *simRun) checkEnded() {
	s.t.Helper()
	if s.g.Phase != PhaseEnded {
		s.failf("phase %s after EndGame", s.g.Phase)
	}
	for _, p := range s.g.Players {
		if _, err := s.g.ViewFor(p.ID); err != nil {
			s.failf("view after EndGame: %v", err)
		}
	}
}

// runGame simulates one game: select and start the mission, play to the debrief, then exercise
// the debrief transitions for a subset of the seeds.
func runGame(t *testing.T, st *simStats, mission, players int, seed uint64) {
	t.Helper()
	s := &simRun{
		t:   t,
		st:  st,
		rng: rand.New(rand.NewPCG(seed, uint64(mission*10+players))),
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

	s.g = newCrew(t, players)
	g := s.g
	// A third of the games play with the distress signal on, the rest with the default (off).
	s.act("setOptions", func() error { return g.SetOptions(alice, Options{Distress: seed%3 == 1}) })
	s.act("selectMission", func() error { return g.SelectMission(alice, mission) })
	s.act("startMission", func() error { return g.StartMission(alice, s.rng) })
	s.runAttempt(1)

	if variant := int(seed % 4); variant < 3 {
		s.followUp(variant)
	}
}

func TestSimulation(t *testing.T) {
	seeds := 20
	if testing.Short() {
		seeds = 3
	}
	st := &simStats{lost: map[Reason]int{}, steps: map[Step]int{}, followUp: map[string]int{}}
	t.Run("games", func(t *testing.T) {
		for mission := 1; mission <= Missions(EditionPlanetNine); mission++ {
			for players := MinPlayers; players <= MaxPlayers; players++ {
				t.Run(fmt.Sprintf("m%02d_p%d", mission, players), func(t *testing.T) {
					t.Parallel()
					for seed := range uint64(seeds) {
						runGame(t, st, mission, players, seed)
					}
				})
			}
		}
	})
	if t.Failed() {
		return
	}
	t.Logf("attempts %d, won %d, lost %v, steps %v, followUps %v, distress %d, handovers %d, communications %d",
		st.attempts, st.won, st.lost, st.steps, st.followUp, st.distress, st.handover, st.comm)
	if st.won == 0 {
		t.Error("no simulated attempt was won")
	}
	if !testing.Short() {
		for _, step := range []Step{StepAdjustTokens, StepPoll, StepPick, StepDraft, StepPass, StepPlay} {
			if st.steps[step] == 0 {
				t.Errorf("step %s was never played", step)
			}
		}
		if st.distress == 0 || st.handover == 0 || st.comm == 0 {
			t.Errorf("distress %d, handovers %d, communications %d: want all > 0", st.distress, st.handover, st.comm)
		}
	}
}
