package domain

import (
	"fmt"
	"slices"
)

// requireActive checks that by is seated, the game is running and it is by's turn.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn.
func (g *Game) requireActive(by PlayerID) (int, error) {
	seat, err := g.requireSeated(by)
	if err != nil {
		return -1, err
	}
	if g.Phase != PhasePlaying {
		return -1, fmt.Errorf("%w: the game is not running", ErrWrongPhase)
	}
	if seat != g.Turn.ActiveSeat {
		return -1, fmt.Errorf("%w: %s", ErrNotYourTurn, by)
	}
	return seat, nil
}

// requireStep checks the current step. Errors: ErrWrongStep.
func (g *Game) requireStep(steps ...Step) error {
	if !slices.Contains(steps, g.Step) {
		return fmt.Errorf("%w: step is %q", ErrWrongStep, g.Step)
	}
	return nil
}

// Play plays cards from by's hand against the enemy (rules §4 Steps 1–3, §5, §7).
// rng shuffles the Discard for a ♥ heal.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn, ErrWrongStep, ErrCardNotInHand, ErrIllegalPlay.
// Emits JesterPlayed, or CardsPlayed followed by the flow events (defeat, discard step, next turn).
func (g *Game) Play(by PlayerID, cards []Card, rng RNG) error {
	seat, err := g.requireActive(by)
	if err != nil {
		return err
	}
	if err := g.requireStep(StepPlay); err != nil {
		return err
	}
	rest, err := removeCards(g.Players[seat].Hand, cards)
	if err != nil {
		return err
	}
	kind, err := ClassifyPlay(cards)
	if err != nil {
		return err
	}

	// All validation done; mutate.
	played := slices.Clone(cards)
	SortCards(played)
	p := &g.Players[seat]
	p.Hand = rest
	p.YieldedLastTurn = false
	g.Played = append(g.Played, PlayedSet{By: by, Cards: played})

	if kind == PlayJester {
		g.emit(JesterPlayed{PlayerID: by})
		if g.IsSolo() {
			// Unreachable: solo has no Tavern Jesters. Kept so the flow can never stall.
			g.beginTurn(seat)
		} else {
			g.Step = StepChooseNext
			g.clearSignals()
		}
		g.commit()
		return nil
	}

	e := g.PlayEffectAt(len(g.Played) - 1)
	healed, drawn := 0, 0
	for _, s := range e.Suits { // Suits order: ♥ resolves before ♦
		if slices.Contains(e.BlockedSuits, s) {
			continue
		}
		switch s {
		case SuitHearts:
			healed = g.heal(e.Value, rng)
		case SuitDiamonds:
			drawn = g.drawRoundRobin(seat, e.Value)
		}
	}
	g.emit(CardsPlayed{
		PlayerID:     by,
		Cards:        slices.Clone(played),
		Value:        e.Value,
		Healed:       healed,
		Drawn:        drawn,
		Damage:       e.Damage,
		BlockedSuits: slices.Clone(e.BlockedSuits),
	})
	if g.Damage() >= EnemyHealth(g.Enemy.Rank) {
		g.defeatEnemy()
	} else {
		g.beginDiscard()
	}
	g.commit()
	return nil
}

// Yield skips Steps 2 and 3 and goes to Step 4 (rules §8).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn, ErrWrongStep, ErrYieldNotAllowed.
// Emits Yielded followed by the flow events.
func (g *Game) Yield(by PlayerID) error {
	seat, err := g.requireActive(by)
	if err != nil {
		return err
	}
	if err := g.requireStep(StepPlay); err != nil {
		return err
	}
	if !g.CanYield(seat) {
		return fmt.Errorf("%w: every other player yielded on their last turn", ErrYieldNotAllowed)
	}
	g.Players[seat].YieldedLastTurn = true
	g.emit(Yielded{PlayerID: by})
	g.beginDiscard()
	g.commit()
	return nil
}

// Discard discards cards to cover the enemy's effective attack (rules §4 Step 4).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn, ErrWrongStep, ErrCardNotInHand,
// ErrIllegalDiscard. Emits CardsDiscarded followed by the next turn's events.
func (g *Game) Discard(by PlayerID, cards []Card) error {
	seat, err := g.requireActive(by)
	if err != nil {
		return err
	}
	if err := g.requireStep(StepDiscard); err != nil {
		return err
	}
	rest, err := removeCards(g.Players[seat].Hand, cards)
	if err != nil {
		return err
	}
	attack := g.EffectiveAttack()
	if err := ValidateDiscard(cards, attack); err != nil {
		return err
	}

	discarded := slices.Clone(cards)
	SortCards(discarded)
	g.Players[seat].Hand = rest
	g.DiscardPile = append(g.DiscardPile, discarded...)
	g.emit(CardsDiscarded{PlayerID: by, Cards: discarded, Attack: attack})
	g.advance()
	g.commit()
	return nil
}

// ChooseNext picks the player who takes the next turn after a Jester (rules §7). The chooser may
// pick themselves. Errors: ErrUnknownPlayer (by or next), ErrWrongPhase, ErrNotYourTurn, ErrWrongStep.
// Emits NextPlayerChosen, TurnStarted (and GameLost if the chosen player cannot act).
func (g *Game) ChooseNext(by, next PlayerID) error {
	if _, err := g.requireActive(by); err != nil {
		return err
	}
	if err := g.requireStep(StepChooseNext); err != nil {
		return err
	}
	_, nextSeat, ok := g.PlayerByID(next)
	if !ok {
		return fmt.Errorf("%w: next player %s", ErrUnknownPlayer, next)
	}
	g.emit(NextPlayerChosen{ByPlayerID: by, PlayerID: next})
	g.beginTurn(nextSeat)
	g.commit()
	return nil
}

// SetJesterSignal sets by's public signal while the Jester player chooses (rules §7 App decision).
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrIllegalSignal (by is the chooser, or an
// unknown signal). Emits JesterSignalSet, also when the signal is unchanged.
func (g *Game) SetJesterSignal(by PlayerID, s JesterSignal) error {
	seat, err := g.requireSeated(by)
	if err != nil {
		return err
	}
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: the game is not running", ErrWrongPhase)
	}
	if err := g.requireStep(StepChooseNext); err != nil {
		return err
	}
	if seat == g.Turn.ActiveSeat {
		return fmt.Errorf("%w: the Jester player cannot signal", ErrIllegalSignal)
	}
	switch s {
	case SignalNone, SignalGoodPlay, SignalRatherNot:
	default:
		return fmt.Errorf("%w: %q", ErrIllegalSignal, s)
	}
	g.Players[seat].Signal = s
	g.emit(JesterSignalSet{PlayerID: by, Signal: s})
	g.commit()
	return nil
}

// UseSoloJester uses a solo Jester power: discard the hand and refill to the max hand size
// (rules §11). Allowed in Step 1 and Step 4. It does not cancel immunity.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrNotSolo, ErrNotYourTurn, ErrWrongStep, ErrNoSoloJester.
// Emits SoloJesterUsed (and GameLost if the player still cannot act / cover).
func (g *Game) UseSoloJester(by PlayerID) error {
	seat, err := g.requireSeated(by)
	if err != nil {
		return err
	}
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: the game is not running", ErrWrongPhase)
	}
	if !g.IsSolo() {
		return fmt.Errorf("%w: Jester powers exist only in solo games", ErrNotSolo)
	}
	if _, err := g.requireActive(by); err != nil {
		return err
	}
	if err := g.requireStep(StepPlay, StepDiscard); err != nil {
		return err
	}
	if g.SoloJestersLeft() == 0 {
		return fmt.Errorf("%w: all %d used", ErrNoSoloJester, g.SoloJestersUsed)
	}

	p := &g.Players[seat]
	g.DiscardPile = append(g.DiscardPile, p.Hand...)
	hand := make([]Card, 0, g.Rules().MaxHandSize)
	for len(hand) < g.Rules().MaxHandSize && len(g.Tavern) > 0 {
		hand = append(hand, g.popTavern())
	}
	SortCards(hand)
	p.Hand = hand
	g.SoloJestersUsed++
	g.emit(SoloJesterUsed{PlayerID: by, Drawn: len(hand), Left: g.SoloJestersLeft()})
	if g.Step == StepPlay {
		g.checkCannotAct()
	} else {
		g.checkCannotCover()
	}
	g.commit()
	return nil
}

// --- internal flow (no validation, no commit) ---

// popTavern removes and returns the top Tavern card. The Tavern must not be empty.
func (g *Game) popTavern() Card {
	top := len(g.Tavern) - 1
	c := g.Tavern[top]
	g.Tavern = g.Tavern[:top]
	return c
}

// heal resolves ♥: shuffle the Discard, move up to n cards under the Tavern. Returns the count.
func (g *Game) heal(n int, rng RNG) int {
	rng.Shuffle(len(g.DiscardPile), func(i, j int) { g.DiscardPile[i], g.DiscardPile[j] = g.DiscardPile[j], g.DiscardPile[i] })
	k := min(n, len(g.DiscardPile))
	cut := len(g.DiscardPile) - k
	moved := slices.Clone(g.DiscardPile[cut:])
	g.DiscardPile = g.DiscardPile[:cut]
	g.Tavern = append(moved, g.Tavern...)
	return k
}

// drawRoundRobin resolves ♦: starting at seat and going clockwise, players draw one card at a time
// until n cards are drawn, skipping players at the max hand size. Stops early when the Tavern is
// empty or every hand is full. Returns the number of cards drawn.
func (g *Game) drawRoundRobin(seat, n int) int {
	maxHand := g.Rules().MaxHandSize
	players := len(g.Players)
	changed := make([]bool, players)
	drawn := 0
	for drawn < n && len(g.Tavern) > 0 {
		progressed := false
		for i := 0; i < players && drawn < n && len(g.Tavern) > 0; i++ {
			s := (seat + i) % players
			p := &g.Players[s]
			if len(p.Hand) >= maxHand {
				continue
			}
			p.Hand = append(p.Hand, g.popTavern())
			changed[s] = true
			drawn++
			progressed = true
		}
		if !progressed {
			break // everyone is full
		}
	}
	for s, ok := range changed {
		if ok {
			SortCards(g.Players[s].Hand)
		}
	}
	return drawn
}

// defeatEnemy handles a defeated enemy (rules §4 Step 3): exact kills go on top of the Tavern,
// overkills to the Discard; the played cards go to the Discard; then the next enemy is revealed
// and the same player starts a new turn, or the game is won.
func (g *Game) defeatEnemy() {
	enemy := *g.Enemy
	exact := g.Damage() == EnemyHealth(enemy.Rank)
	for _, ps := range g.Played {
		g.DiscardPile = append(g.DiscardPile, ps.Cards...)
	}
	g.Played = nil
	if exact {
		g.Tavern = append(g.Tavern, enemy)
	} else {
		g.DiscardPile = append(g.DiscardPile, enemy)
	}
	g.emit(EnemyDefeated{Enemy: enemy, Exact: exact})
	if len(g.Castle) == 0 {
		g.Enemy = nil
		g.win()
		return
	}
	top := len(g.Castle) - 1
	next := g.Castle[top]
	g.Castle = g.Castle[:top]
	g.Enemy = &next
	g.emit(EnemyRevealed{Enemy: next})
	g.beginTurn(g.Turn.ActiveSeat)
}

// beginTurn starts Step 1 for seat and checks the "cannot act" loss.
func (g *Game) beginTurn(seat int) {
	g.Turn.Number++
	g.Turn.ActiveSeat = seat
	g.Step = StepPlay
	g.clearSignals()
	g.emit(TurnStarted{PlayerID: g.Players[seat].ID, Number: g.Turn.Number})
	g.checkCannotAct()
}

// clearSignals resets every Jester signal.
func (g *Game) clearSignals() {
	for i := range g.Players {
		g.Players[i].Signal = SignalNone
	}
}

// checkCannotAct loses the game if the active player cannot play or yield (rules §9 App decisions):
// empty hand, no solo Jester power left, and yielding is forbidden or the game is solo.
func (g *Game) checkCannotAct() {
	seat := g.Turn.ActiveSeat
	if len(g.Players[seat].Hand) > 0 || g.SoloJestersLeft() > 0 {
		return
	}
	if g.IsSolo() || !g.CanYield(seat) {
		g.lose(LossCannotAct, g.Players[seat].ID)
	}
}

// beginDiscard starts Step 4. With effective attack 0 it completes automatically.
func (g *Game) beginDiscard() {
	g.Step = StepDiscard
	if g.EffectiveAttack() == 0 {
		g.emit(AttackBlocked{PlayerID: g.Players[g.Turn.ActiveSeat].ID})
		g.advance()
		return
	}
	g.checkCannotCover()
}

// checkCannotCover loses the game if the active player's hand cannot cover the effective attack
// and no solo Jester power is left (rules §4 Step 4, §11 App decision).
func (g *Game) checkCannotCover() {
	p := g.Players[g.Turn.ActiveSeat]
	if HandValue(p.Hand) < g.EffectiveAttack() && g.SoloJestersLeft() == 0 {
		g.lose(LossCannotCover, p.ID)
	}
}

// advance passes the turn to the next player clockwise.
func (g *Game) advance() {
	g.beginTurn((g.Turn.ActiveSeat + 1) % len(g.Players))
}

// win ends the game as a win.
func (g *Game) win() {
	g.Phase = PhaseWon
	g.Step = StepNone
	g.Result = &Result{}
	g.emit(GameWon{Grade: g.Grade()})
}

// lose ends the game as a loss. The enemy stays visible.
func (g *Game) lose(r LossReason, p PlayerID) {
	g.Phase = PhaseLost
	g.Step = StepNone
	g.Result = &Result{Reason: r, PlayerID: p}
	g.emit(GameLost{Reason: r, PlayerID: p})
}
