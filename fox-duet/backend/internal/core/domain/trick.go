package domain

import (
	"fmt"
	"slices"
)

// dealRound starts the next round: the deal passes to the other seat, all 30 cards are
// shuffled, each player gets HandSize cards (one at a time, the non-dealer first), the next card
// becomes the decree, and the non-dealer leads (rules §5, §7).
func (g *Game) dealRound(rng RNG) {
	g.Round++
	g.Dealer = 1 - g.Dealer
	deck := NewDeck()
	rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	hands := [Players][]Card{}
	for range HandSize {
		for k := 1; k <= Players; k++ {
			seat := (g.Dealer + k) % Players
			top := len(deck) - 1
			hands[seat] = append(hands[seat], deck[top])
			deck = deck[:top]
		}
	}
	for seat := range g.Players {
		SortCards(hands[seat])
		g.Players[seat].Hand = hands[seat]
	}
	top := len(deck) - 1
	decree := deck[top]
	g.Decree = &decree
	g.Deck = deck[:top]
	g.Trick = Trick{Number: 1, Leader: 1 - g.Dealer}
	g.Step = StepLead
	g.emit(RoundStarted{Round: g.Round, DealerID: g.Players[g.Dealer].ID, Decree: decree})
}

// Actors returns the seats that must act now: one seat, both while choosing gifts, none when
// not playing.
func (g *Game) Actors() []int {
	if g.Phase != PhasePlaying {
		return nil
	}
	switch g.Step {
	case StepLead:
		return []int{g.Trick.Leader}
	case StepFollow:
		return []int{1 - g.Trick.Leader}
	case StepFoxesChoose:
		seat, _ := g.Trick.played()
		return []int{seat}
	case StepFoxesSwap:
		return []int{g.Trick.Chosen}
	case StepGift:
		var out []int
		for seat, c := range g.Trick.Gift {
			if c == nil {
				out = append(out, seat)
			}
		}
		return out
	case StepMove:
		return []int{g.TrickWinner()}
	case StepForestEnd:
		_, seat, _ := g.PlayerByID(g.HostID)
		return []int{seat}
	}
	return nil
}

// requireActor checks that by is seated, the game is playing in one of steps, and by must act.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn.
func (g *Game) requireActor(by PlayerID, steps ...Step) (int, error) {
	seat, err := g.requireSeated(by)
	if err != nil {
		return -1, err
	}
	if g.Phase != PhasePlaying {
		return -1, fmt.Errorf("%w: the game is not running", ErrWrongPhase)
	}
	if !slices.Contains(steps, g.Step) {
		return -1, fmt.Errorf("%w: the game waits for %s", ErrWrongStep, g.Step)
	}
	if !slices.Contains(g.Actors(), seat) {
		return -1, fmt.Errorf("%w: %s", ErrNotYourTurn, g.Step)
	}
	return seat, nil
}

// LegalCards returns the cards seat may play now: the whole hand when leading; when following,
// the cards of the lead suit if any, unless a Royal Heir was led (rules §6). Empty when seat
// does not lead or follow now.
func (g *Game) LegalCards(seat int) []Card {
	if g.Phase != PhasePlaying || seat < 0 || seat >= len(g.Players) {
		return []Card{}
	}
	hand := g.Players[seat].Hand
	switch {
	case g.Step == StepLead && seat == g.Trick.Leader:
		return cloneCards(hand)
	case g.Step == StepFollow && seat != g.Trick.Leader:
		lead := *g.Trick.Lead
		if lead.Rank != RankRoyalHeir {
			var same []Card
			for _, c := range hand {
				if c.Suit == lead.Suit {
					same = append(same, c)
				}
			}
			if len(same) > 0 {
				return same
			}
		}
		return cloneCards(hand)
	}
	return []Card{}
}

// PlayCard plays c from by's hand as the lead or follow card, then resolves its ability, and
// with the second card the trick.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn, ErrCardNotInHand,
// ErrMustFollowSuit. Emits CardPlayed and whatever the trick resolution emits.
func (g *Game) PlayCard(by PlayerID, c Card, rng RNG) error {
	seat, err := g.requireActor(by, StepLead, StepFollow)
	if err != nil {
		return err
	}
	hand := g.Players[seat].Hand
	idx := slices.Index(hand, c)
	if idx < 0 {
		return fmt.Errorf("%w: %s", ErrCardNotInHand, c)
	}
	if !slices.Contains(g.LegalCards(seat), c) {
		return fmt.Errorf("%w: %s", ErrMustFollowSuit, g.Trick.Lead.Suit)
	}

	// All validation done; mutate.
	g.Players[seat].Hand = slices.Delete(slices.Clone(hand), idx, idx+1)
	lead := g.Step == StepLead
	if lead {
		g.LastTrick = nil
		g.LastGift = nil
		g.Trick.Lead = &c
	} else {
		g.Trick.Follow = &c
	}
	g.emit(CardPlayed{PlayerID: by, Card: c, Lead: lead})
	g.resolveAbility(rng)
	g.commit()
	return nil
}

// resolveAbility starts the ability of the card just played, if it needs a decision, or goes on.
func (g *Game) resolveAbility(rng RNG) {
	_, card := g.Trick.played()
	empty0, empty1 := len(g.Players[0].Hand) == 0, len(g.Players[1].Hand) == 0
	switch card.Ability() {
	case AbilityFoxes:
		if !empty0 || !empty1 {
			g.Step = StepFoxesChoose
			return
		}
	case AbilityGift:
		if !empty0 && !empty1 {
			g.Step = StepGift
			g.Trick.Gift = [Players]*Card{}
			for seat, p := range g.Players {
				if len(p.Hand) == 1 {
					c := p.Hand[0]
					g.Trick.Gift[seat] = &c
					g.emit(GiftChosen{PlayerID: p.ID, Auto: true})
				}
			}
			if g.Trick.Gift[0] == nil || g.Trick.Gift[1] == nil {
				return
			}
			g.exchangeGifts()
		}
	}
	g.afterAbility(rng)
}

// afterAbility continues the trick once the ability of the last card is done.
func (g *Game) afterAbility(rng RNG) {
	if g.Trick.Follow == nil {
		g.Step = StepFollow
		return
	}
	winner := g.TrickWinner()
	g.emit(TrickWon{PlayerID: g.Players[winner].ID, Number: g.Trick.Number})
	opts := g.MoveOptions()
	if len(opts) > 1 {
		g.Step = StepMove
		return
	}
	g.move(opts[0].Delta, rng)
}

// ChooseFoxesPlayer: whoever played the Foxes chooses the player who may swap the decree.
// If that player's hand is empty, nothing happens (app decision).
// Errors: ErrUnknownPlayer (by or chosen), ErrWrongPhase, ErrWrongStep, ErrNotYourTurn.
// Emits FoxesChosen.
func (g *Game) ChooseFoxesPlayer(by, chosen PlayerID, rng RNG) error {
	if _, err := g.requireActor(by, StepFoxesChoose); err != nil {
		return err
	}
	_, cs, ok := g.PlayerByID(chosen)
	if !ok {
		return fmt.Errorf("%w: chosen player %s", ErrUnknownPlayer, chosen)
	}

	skipped := len(g.Players[cs].Hand) == 0
	g.emit(FoxesChosen{ByPlayerID: by, PlayerID: chosen, Skipped: skipped})
	if skipped {
		g.afterAbility(rng)
	} else {
		g.Trick.Chosen = cs
		g.Step = StepFoxesSwap
	}
	g.commit()
	return nil
}

// SwapDecree: the player chosen by the Foxes puts c from their hand as the new decree and
// takes the old decree into their hand.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn, ErrCardNotInHand.
// Emits DecreeSwapped.
func (g *Game) SwapDecree(by PlayerID, c Card, rng RNG) error {
	seat, err := g.requireActor(by, StepFoxesSwap)
	if err != nil {
		return err
	}
	hand := g.Players[seat].Hand
	idx := slices.Index(hand, c)
	if idx < 0 {
		return fmt.Errorf("%w: %s", ErrCardNotInHand, c)
	}

	old := *g.Decree
	hand = slices.Clone(hand)
	hand[idx] = old
	SortCards(hand)
	g.Players[seat].Hand = hand
	g.Decree = &c
	g.emit(DecreeSwapped{PlayerID: by, Old: old, New: c})
	g.afterAbility(rng)
	g.commit()
	return nil
}

// KeepDecree: the player chosen by the Foxes leaves the decree as it is.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn. Emits DecreeKept.
func (g *Game) KeepDecree(by PlayerID, rng RNG) error {
	if _, err := g.requireActor(by, StepFoxesSwap); err != nil {
		return err
	}
	g.emit(DecreeKept{PlayerID: by})
	g.afterAbility(rng)
	g.commit()
	return nil
}

// GiveGift: by secretly chooses c from their hand as their Gift card. Once both have chosen,
// the cards change hands.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrGiftAlreadyChosen,
// ErrCardNotInHand. Emits GiftChosen, then GiftExchanged once both have chosen.
func (g *Game) GiveGift(by PlayerID, c Card, rng RNG) error {
	seat, err := g.requireSeated(by)
	if err != nil {
		return err
	}
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: the game is not running", ErrWrongPhase)
	}
	if g.Step != StepGift {
		return fmt.Errorf("%w: the game waits for %s", ErrWrongStep, g.Step)
	}
	if g.Trick.Gift[seat] != nil {
		return fmt.Errorf("%w: wait for your teammate", ErrGiftAlreadyChosen)
	}
	if !slices.Contains(g.Players[seat].Hand, c) {
		return fmt.Errorf("%w: %s", ErrCardNotInHand, c)
	}

	g.Trick.Gift[seat] = &c
	g.emit(GiftChosen{PlayerID: by})
	if g.Trick.Gift[1-seat] != nil {
		g.exchangeGifts()
		g.afterAbility(rng)
	}
	g.commit()
	return nil
}

// exchangeGifts swaps the two chosen gift cards between the hands.
func (g *Game) exchangeGifts() {
	given := [Players]Card{*g.Trick.Gift[0], *g.Trick.Gift[1]}
	for seat := range g.Players {
		hand := slices.Clone(g.Players[seat].Hand)
		hand[slices.Index(hand, given[seat])] = given[1-seat]
		SortCards(hand)
		g.Players[seat].Hand = hand
	}
	g.LastGift = &GiftRecord{Given: given}
	g.Trick.Gift = [Players]*Card{}
	g.emit(GiftExchanged{})
}

// TrickWinner returns the seat that wins the trick; both cards must be on the table (rules §6).
func (g *Game) TrickWinner() int {
	if leadWins(*g.Trick.Lead, *g.Trick.Follow, g.Decree.Suit) {
		return g.Trick.Leader
	}
	return 1 - g.Trick.Leader
}

// leadWins reports whether the led card beats the follow card.
func leadWins(lead, follow Card, trump Suit) bool {
	switch {
	case follow.Suit == lead.Suit:
		return lead.Rank > follow.Rank
	case follow.Suit == trump:
		return false
	}
	return true
}

// MoveOption is one distinct result the trick winner can choose (rules §6, app decision).
type MoveOption struct {
	Delta  int  // spaces, negative = toward the left end
	To     int  // landing space; the start space after leaving the path; the old space when Lost
	Beyond End  // the end the traveller would leave the path at, EndNone if it stays on it
	Gem    bool // a gem would be collected
	Lost   bool // leaving the path with no forest token left: the game is lost
}

// moveDeltas returns every move the winner may choose, with repeats, the normal move first:
// toward the winner (and, with a Musician, the reverse), with each subset of cards whose
// movement may be ignored (one card per Gazelle).
func (g *Game) moveDeltas() []int {
	cards := [2]Card{*g.Trick.Lead, *g.Trick.Follow}
	musician, gazelles := false, 0
	for _, c := range cards {
		switch c.Rank {
		case RankMusician:
			musician = true
		case RankGazelle:
			gazelles++
		}
	}
	m0, m1 := cards[0].Movement(), cards[1].Movement()
	sums := []int{m0 + m1}
	if gazelles >= 1 {
		sums = append(sums, m1, m0)
	}
	if gazelles >= 2 {
		sums = append(sums, 0)
	}
	dir := direction(endOf(g.TrickWinner()))
	dirs := []int{dir}
	if musician {
		dirs = append(dirs, -dir)
	}
	var out []int
	for _, d := range dirs {
		for _, s := range sums {
			out = append(out, d*s)
		}
	}
	return out
}

// MoveOptions returns the distinct results the trick winner can choose from, the normal move
// first. Both cards must be on the table.
func (g *Game) MoveOptions() []MoveOption {
	var out []MoveOption
	for _, d := range g.moveDeltas() {
		o := g.moveOption(d)
		if !slices.ContainsFunc(out, func(x MoveOption) bool { return x.To == o.To && x.Beyond == o.Beyond }) {
			out = append(out, o)
		}
	}
	return out
}

// moveOption describes the result of moving the traveller by delta.
func (g *Game) moveOption(delta int) MoveOption {
	b := &g.Board
	o := MoveOption{Delta: delta, Beyond: b.beyond(delta)}
	switch {
	case o.Beyond != EndNone && b.ForestTokens == 0:
		o.To, o.Lost = b.Traveller, true
	case o.Beyond != EndNone:
		o.To = g.LevelDef().Start()
	default:
		o.To = b.Traveller + delta
		o.Gem = b.Gems[o.To] > 0
	}
	return o
}

// ChooseMove: the trick winner moves the traveller by delta, one of the moves the Musician
// and Gazelle allow.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotYourTurn, ErrIllegalMove.
// Emits what the move emits.
func (g *Game) ChooseMove(by PlayerID, delta int, rng RNG) error {
	if _, err := g.requireActor(by, StepMove); err != nil {
		return err
	}
	if !slices.Contains(g.moveDeltas(), delta) {
		return fmt.Errorf("%w: %d spaces", ErrIllegalMove, delta)
	}
	g.move(delta, rng)
	g.commit()
	return nil
}

// move moves the traveller by delta at the end of a trick, then collects a gem, ends the game,
// or goes on with the next trick or the round end (rules §6).
func (g *Game) move(delta int, rng RNG) {
	b := &g.Board
	o := g.moveOption(delta)
	winner := g.TrickWinner()
	rec := &TrickRecord{
		Number: g.Trick.Number, Leader: g.Trick.Leader, Lead: *g.Trick.Lead, Follow: *g.Trick.Follow,
		Winner: winner, From: b.Traveller, Delta: delta, To: o.To, Beyond: o.Beyond,
	}
	g.LastTrick = rec
	number := g.Trick.Number
	g.Trick = Trick{}
	if o.Lost {
		g.lose(LossLostInForest, "")
		return
	}
	g.emit(TravellerMoved{From: rec.From, To: o.To, Delta: delta, Beyond: o.Beyond})
	b.Traveller = o.To
	if o.Beyond != EndNone {
		space, moved := b.cover(o.Beyond)
		g.emit(ForestTokenPlaced{End: o.Beyond, Space: space, GemsMoved: moved, Left: b.ForestTokens})
	} else if o.Gem {
		b.Gems[o.To]--
		rec.Collected = true
		g.emit(GemCollected{Space: o.To, Left: b.GemsLeft()})
		if b.GemsLeft() == 0 {
			g.win()
			return
		}
	}
	if number == TricksPerRound {
		g.endRound(rng)
		return
	}
	g.Trick = Trick{Number: number + 1, Leader: winner}
	g.Step = StepLead
}

// endRound ends the game after the last round, or adds the round-end gems, places a forest
// token (asking the host when both ends are allowed) and deals the next round (rules §7).
func (g *Game) endRound(rng RNG) {
	if g.Round == Rounds {
		g.lose(LossOutOfTime, "")
		return
	}
	g.emit(GemsAdded{Count: g.Board.addRoundGems(g.LevelDef().Spaces())})
	switch ends := g.Board.coverableEnds(); len(ends) {
	case 0:
		g.dealRound(rng)
	case 1:
		g.placeRoundEndToken(ends[0])
		g.dealRound(rng)
	default:
		g.Step = StepForestEnd
	}
}

// placeRoundEndToken covers end e at the end of a round.
func (g *Game) placeRoundEndToken(e End) {
	space, moved := g.Board.cover(e)
	g.emit(ForestTokenPlaced{End: e, Space: space, GemsMoved: moved, RoundEnd: true, Left: g.Board.ForestTokens})
}

// ChooseForestEnd: at the end of a round the host chooses the end the forest token covers,
// then the next round is dealt.
// Errors: ErrUnknownPlayer, ErrWrongPhase, ErrWrongStep, ErrNotHost, ErrIllegalEnd.
// Emits ForestTokenPlaced, RoundStarted.
func (g *Game) ChooseForestEnd(by PlayerID, e End, rng RNG) error {
	if _, err := g.requireSeated(by); err != nil {
		return err
	}
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: the game is not running", ErrWrongPhase)
	}
	if g.Step != StepForestEnd {
		return fmt.Errorf("%w: the game waits for %s", ErrWrongStep, g.Step)
	}
	if !g.IsHost(by) {
		return fmt.Errorf("%w: the host chooses the end", ErrNotHost)
	}
	if !slices.Contains(g.Board.coverableEnds(), e) {
		return fmt.Errorf("%w: %q", ErrIllegalEnd, e)
	}
	g.placeRoundEndToken(e)
	g.dealRound(rng)
	g.commit()
	return nil
}

// CoverableEnds returns the ends the round-end forest token may cover (StepForestEnd only).
func (g *Game) CoverableEnds() []End {
	if g.Step != StepForestEnd {
		return nil
	}
	return g.Board.coverableEnds()
}
