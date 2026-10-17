package domain

import (
	"fmt"
	"slices"
)

// StatementFor returns the true statement about card c in hand (rules §8): PosOnly if it is the
// only card of its colour, PosHighest / PosLowest if it is the highest / lowest of at least two.
// ok is false for a middle card, a trump or a card not in hand: those cannot be communicated.
func StatementFor(hand []Card, c Card) (pos Position, ok bool) {
	if c.IsTrump() || !slices.Contains(hand, c) {
		return "", false
	}
	lo, hi, n := c.Rank, c.Rank, 0
	for _, h := range hand {
		if h.Suit != c.Suit {
			continue
		}
		n++
		lo, hi = min(lo, h.Rank), max(hi, h.Rank)
	}
	switch {
	case n == 1:
		return PosOnly, true
	case c.Rank == hi:
		return PosHighest, true
	case c.Rank == lo:
		return PosLowest, true
	}
	return "", false
}

// CanCommunicate reports why seat may not communicate now, or nil. Communication is open to a
// human who has not communicated in this attempt and is not silenced, during StepPlay before any
// card of the current trick is played, from the mission's first allowed trick on (rules §8).
// Deep Sea: never without communication; in rapture of the deep while a shared token is left,
// as often as the member likes; in 10/13 only before the first trick once the tasks went to
// another member (deep-sea/rules.md §4).
func (g *Game) CanCommunicate(seat int) error {
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: no attempt is running", ErrWrongPhase)
	}
	a := g.Attempt
	switch {
	case a.Step != StepPlay:
		return fmt.Errorf("%w: the attempt waits for %s", ErrWrongStep, a.Step)
	case !a.validSeat(seat) || a.Members[seat].Dummy:
		return fmt.Errorf("%w: seat %d", ErrInvalidSeat, seat)
	case a.Comm == CommNone:
		return fmt.Errorf("%w: no communication in this mission", ErrCannotCommunicate)
	case a.Comm == CommRapture && a.TokensLeft() == 0:
		return fmt.Errorf("%w: no shared token is left", ErrCannotCommunicate)
	case a.Comm != CommRapture && len(a.Comms[seat]) > 0:
		return fmt.Errorf("%w: once per attempt", ErrAlreadyCommunicated)
	case a.Silenced == seat:
		return fmt.Errorf("%w: you are silenced in this attempt", ErrCannotCommunicate)
	case len(a.Trick.Plays) > 0:
		return fmt.Errorf("%w: only before a trick", ErrCannotCommunicate)
	case len(a.Tricks)+1 < g.MissionDef().Comm.FromTrick:
		return fmt.Errorf("%w: not before trick %d", ErrCannotCommunicate, g.MissionDef().Comm.FromTrick)
	case g.onlyBeforeFirstTrick() && len(a.Tricks) > 0:
		return fmt.Errorf("%w: only before the first trick", ErrCannotCommunicate)
	}
	return nil
}

// Communicate puts card face up with the statement pos (rules §8). In a dead zone pos is ignored,
// the card must still qualify and is shown without a position.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrAlreadyCommunicated,
// ErrCannotCommunicate (timing, silenced, trump), ErrCardNotInHand, ErrInvalidPosition.
// Emits Communicated.
func (g *Game) Communicate(by PlayerID, card Card, pos Position) error {
	seat, err := g.requireSeated(by)
	if err != nil {
		return err
	}
	if err := g.CanCommunicate(seat); err != nil {
		return err
	}
	a := g.Attempt
	hand := a.Members[seat].Hand
	if !slices.Contains(hand, card) {
		return fmt.Errorf("%w: %s", ErrCardNotInHand, card)
	}
	if card.IsTrump() {
		return fmt.Errorf("%w: trump cards cannot be communicated", ErrCannotCommunicate)
	}
	if a.communicates(seat, card) {
		return fmt.Errorf("%w: %s is already on the table", ErrCannotCommunicate, card)
	}
	truth, ok := StatementFor(hand, card)
	if !ok {
		return fmt.Errorf("%w: %s is neither your highest, lowest nor only %s card", ErrInvalidPosition, card, card.Suit)
	}
	if g.deadZone() {
		pos = PosHidden
	} else if pos != truth {
		return fmt.Errorf("%w: %s is not your %s %s card", ErrInvalidPosition, card, pos, card.Suit)
	}
	a.Comms[seat] = append(a.Comms[seat], Communication{Card: card, Position: pos, Trick: len(a.Tricks) + 1})
	g.emit(Communicated{Seat: seat, Card: card, Position: pos})
	g.commit()
	return nil
}

// CommunicatedCards returns seat's communications whose card is still in the hand (on the table
// in front of the member), in order. Never nil.
func (a *Attempt) CommunicatedCards(seat int) []Communication {
	out := []Communication{}
	for _, c := range a.Comms[seat] {
		if slices.Contains(a.Members[seat].Hand, c.Card) {
			out = append(out, c)
		}
	}
	return out
}

// CommunicatedCard returns seat's first communication whose card is still in the hand, else nil.
func (a *Attempt) CommunicatedCard(seat int) *Communication {
	if cs := a.CommunicatedCards(seat); len(cs) > 0 {
		return &cs[0]
	}
	return nil
}

// communicates reports whether card is one of seat's communicated cards on the table.
func (a *Attempt) communicates(seat int, card Card) bool {
	return slices.ContainsFunc(a.CommunicatedCards(seat), func(c Communication) bool { return c.Card == card })
}

// CanSendDistress reports why the commander (seat) may not send or skip the distress signal now,
// or nil: only if the lobby option is on, before the first trick, before anyone has communicated,
// once per attempt (rules §9).
func (g *Game) CanSendDistress(seat int) error {
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: no attempt is running", ErrWrongPhase)
	}
	a := g.Attempt
	if err := a.requireCommander(seat); err != nil {
		return err
	}
	switch {
	case !g.Options.Distress:
		return fmt.Errorf("%w: the distress signal is off", ErrNotAllowed)
	case !a.prePlay():
		return fmt.Errorf("%w: only before the first trick", ErrNotAllowed)
	case a.DistressDecided:
		return fmt.Errorf("%w: already decided in this attempt", ErrNotAllowed)
	case a.anyCommunicated():
		return fmt.Errorf("%w: someone has communicated", ErrNotAllowed)
	}
	return nil
}

// SendDistress sends the distress signal (commander): every human passes one non-trump card in
// direction dir. The logbook notes the signal.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotCommander, ErrNotAllowed, ErrInvalidDirection.
// Emits DistressSignal.
func (g *Game) SendDistress(by PlayerID, dir Direction) error {
	seat, err := g.requireSeated(by)
	if err != nil {
		return err
	}
	if err := g.CanSendDistress(seat); err != nil {
		return err
	}
	if dir != DirLeft && dir != DirRight {
		return fmt.Errorf("%w: %q", ErrInvalidDirection, dir)
	}
	a := g.Attempt
	a.DistressDecided = true
	a.Pass = &Pass{Direction: dir, Chosen: make([]*Card, len(a.Members))}
	a.Step = StepPass
	if e := g.Logbook[g.Mission]; !e.Completed {
		e.Distress = true
		g.Logbook[g.Mission] = e
	}
	g.emit(DistressSignal{Direction: dir})
	g.commit()
	return nil
}

// SkipDistress lets the distress signal pass for this attempt (commander).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotCommander, ErrNotAllowed. Emits DistressSignal.
func (g *Game) SkipDistress(by PlayerID) error {
	seat, err := g.requireSeated(by)
	if err != nil {
		return err
	}
	if err := g.CanSendDistress(seat); err != nil {
		return err
	}
	g.Attempt.DistressDecided = true
	g.emit(DistressSignal{Skipped: true})
	g.commit()
	return nil
}

// ChoosePassCard chooses the card to pass in a distress pass (every human, once). When all have
// chosen, the cards move at once and play begins.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrAlreadyAnswered, ErrCardNotInHand,
// ErrCannotPassTrump. Emits PassCardChosen, then CardsPassed.
func (g *Game) ChoosePassCard(by PlayerID, card Card) error {
	seat, err := g.requireStep(by, StepPass)
	if err != nil {
		return err
	}
	a := g.Attempt
	if a.Pass.Chosen[seat] != nil {
		return fmt.Errorf("%w: card already chosen", ErrAlreadyAnswered)
	}
	if !slices.Contains(a.Members[seat].Hand, card) {
		return fmt.Errorf("%w: %s", ErrCardNotInHand, card)
	}
	if card.IsTrump() {
		return fmt.Errorf("%w: %s", ErrCannotPassTrump, card)
	}
	a.Pass.Chosen[seat] = &card
	g.emit(PassCardChosen{Seat: seat})
	if slices.ContainsFunc(a.humans(), func(s int) bool { return a.Pass.Chosen[s] == nil }) {
		g.commit()
		return nil
	}

	ex := &Exchange{Kind: ExchangePass, Gave: make([]*Card, len(a.Members)), Got: make([]*Card, len(a.Members))}
	for _, s := range a.humans() {
		c := *a.Pass.Chosen[s]
		ex.Gave[s] = &c
		ex.Got[a.neighbour(s, a.Pass.Direction)] = &c
	}
	a.applyExchange(ex)
	dir := a.Pass.Direction
	a.Pass = nil
	a.Step = StepPlay
	g.emit(CardsPassed{Direction: dir})
	g.commit()
	return nil
}

// applyExchange moves the cards of ex: every card in Gave leaves its seat, then every card in Got
// arrives.
func (a *Attempt) applyExchange(ex *Exchange) {
	for s, c := range ex.Gave {
		if c != nil {
			a.Members[s].take(*c)
		}
	}
	for s, c := range ex.Got {
		if c != nil {
			a.Members[s].give(*c)
		}
	}
	a.Exchange = ex
}

// draw is mission 12's draw after the first trick (rules §16): each human takes a random card
// from the right neighbour (the previous human seat), never a communicated card on the table. All
// draws are decided first, then the cards move at once.
func (a *Attempt) draw(rng RNG) {
	ex := &Exchange{Kind: ExchangeDraw, Gave: make([]*Card, len(a.Members)), Got: make([]*Card, len(a.Members))}
	for _, s := range a.humans() {
		from := a.neighbour(s, DirRight)
		var pool []Card
		for _, c := range a.Members[from].Hand {
			if !a.communicates(from, c) {
				pool = append(pool, c)
			}
		}
		if len(pool) == 0 {
			continue
		}
		c := pool[rng.IntN(len(pool))]
		ex.Gave[from], ex.Got[s] = &c, &c
	}
	a.applyExchange(ex)
}

// CanHandover reports why the commander (seat) may not hand over a task now, or nil: 5 players, a
// mission with a golden frame, before the first trick, once per attempt (rules §11).
func (g *Game) CanHandover(seat int) error {
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: no attempt is running", ErrWrongPhase)
	}
	a := g.Attempt
	if err := a.requireCommander(seat); err != nil {
		return err
	}
	switch {
	case len(a.Members) != MaxPlayers || !g.MissionDef().Golden:
		return fmt.Errorf("%w: only with 5 players in a golden mission", ErrNotAllowed)
	case !a.prePlay():
		return fmt.Errorf("%w: only before the first trick", ErrNotAllowed)
	case a.HandoverUsed:
		return fmt.Errorf("%w: once per attempt", ErrNotAllowed)
	}
	return nil
}

// Handover gives task taskID, with its token, to seat to (commander, rules §11).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotCommander, ErrNotAllowed, ErrInvalidTask,
// ErrInvalidSeat. Emits TaskHandedOver.
func (g *Game) Handover(by PlayerID, taskID, to int) error {
	seat, err := g.requireSeated(by)
	if err != nil {
		return err
	}
	if err := g.CanHandover(seat); err != nil {
		return err
	}
	a := g.Attempt
	if !a.validTask(taskID) {
		return fmt.Errorf("%w: %d", ErrInvalidTask, taskID)
	}
	from := a.Tasks[taskID].Owner()
	if !a.validSeat(to) || to == from {
		return fmt.Errorf("%w: %d", ErrInvalidSeat, to)
	}
	a.Tasks[taskID].Objective.Owner = to
	a.HandoverUsed = true
	g.emit(TaskHandedOver{TaskID: taskID, From: from, To: to})
	g.commit()
	return nil
}

// Actors returns the seats of the humans the attempt waits for (the commander for the dummy).
// While playing it is never empty; outside StepPlay/StepDraft several humans may be listed.
func (g *Game) Actors() []int {
	if g.Phase != PhasePlaying {
		return nil
	}
	a := g.Attempt
	switch a.Step {
	case StepAdjustTokens, StepPick, StepAssign:
		return []int{a.Commander}
	case StepPredict:
		return []int{a.actorFor(a.Tasks[g.pendingItem().index].Owner())}
	case StepPoll:
		var out []int
		for _, s := range a.pollers() {
			if a.Poll.Answers[s] == nil {
				out = append(out, s)
			}
		}
		return out
	case StepPass:
		var out []int
		for _, s := range a.humans() {
			if a.Pass.Chosen[s] == nil {
				out = append(out, s)
			}
		}
		return out
	case StepDraft:
		return []int{a.actorFor(g.draftTurn())}
	case StepPlay:
		return []int{a.actorFor(a.turn())}
	}
	return nil
}
