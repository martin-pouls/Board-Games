package domain

import "slices"

// The Deep Sea objective kinds (deep-sea/rules.md §6, §9 and the mission objectives of §8). The
// same rule as for every kind: fulfilled once met and no longer able to fail, failed as soon as
// it can no longer be met. "Unplayed" cards include the card left over with 3 players, so a check
// that relies on them is decided at the latest after the last trick.

// wonCounts returns the tricks won per seat.
func (h History) wonCounts() []int {
	won := make([]int, h.Seats)
	for _, t := range h.Tricks {
		won[t.Winner]++
	}
	return won
}

// cardCounts returns how many cards matching pred the owner has won and how many have been played
// at all.
func (h History) cardCounts(owner int, pred func(Card) bool) (won, played int) {
	for _, t := range h.Tricks {
		for _, p := range t.Plays {
			if !pred(p.Card) {
				continue
			}
			played++
			if t.Winner == owner {
				won++
			}
		}
	}
	return won, played
}

// unplayed returns the cards of the deck not played in h.
func (h History) unplayed() []Card {
	var out []Card
	for _, c := range NewDeck() {
		if !slices.ContainsFunc(h.Tricks, func(t TrickRecord) bool { return t.contains(c) }) {
			out = append(out, c)
		}
	}
	return out
}

// deckCount is the number of cards of the deck matching pred.
func deckCount(pred func(Card) bool) int {
	n := 0
	for _, c := range NewDeck() {
		if pred(c) {
			n++
		}
	}
	return n
}

// isSuit selects the cards of suit s.
func isSuit(s Suit) func(Card) bool { return func(c Card) bool { return c.Suit == s } }

// matcher selects the cards of Suits (any) that are of Ranks (any; colour cards only).
func (o Objective) matcher() func(Card) bool {
	return func(c Card) bool {
		if len(o.Suits) > 0 && !slices.Contains(o.Suits, c.Suit) {
			return false
		}
		return len(o.Ranks) == 0 || (!c.IsTrump() && slices.Contains(o.Ranks, c.Rank))
	}
}

// others returns the counts of every seat but owner.
func others(won []int, owner int) []int {
	out := make([]int, 0, len(won)-1)
	for s, n := range won {
		if s != owner {
			out = append(out, n)
		}
	}
	return out
}

// decided returns fulfilled if ok, failed if fail, else open.
func decided(ok, fail bool) Status {
	switch {
	case fail:
		return StatusFailed
	case ok:
		return StatusFulfilled
	}
	return StatusOpen
}

// cmpTricks (1–3): more than each other member, fewer than each, or more than all others together
// (the dummy counts as a member).
func (o Objective) cmpTricks(h History) Status {
	won := h.wonCounts()
	mine, rem, rest := won[o.Owner], h.remaining(), others(won, o.Owner)
	switch o.Cmp {
	case CmpMore:
		top := slices.Max(rest)
		return decided(mine > top+rem, mine+rem <= top)
	case CmpFewer:
		low := slices.Min(rest)
		return decided(mine+rem < low, low+rem <= mine)
	case CmpMoreThanAll:
		sum := 0
		for _, n := range rest {
			sum += n
		}
		return decided(mine > sum+rem, mine+rem <= sum)
	}
	panic("domain: cmpTricks without a comparison")
}

// cmpCaptain (4–6): more, fewer or as many tricks as the captain.
func (o Objective) cmpCaptain(h History) Status {
	won := h.wonCounts()
	mine, capt, rem := won[o.Owner], won[h.Captain], h.remaining()
	switch o.Cmp {
	case CmpMore:
		return decided(mine > capt+rem, mine+rem <= capt)
	case CmpFewer:
		return decided(mine+rem < capt, capt+rem <= mine)
	case CmpEqual:
		d := mine - capt
		return decided(rem == 0 && d == 0, d > rem || -d > rem)
	}
	panic("domain: cmpCaptain without a comparison")
}

// predict (90, 91): exactly the predicted number of tricks; open until the prediction is made.
func (o Objective) predict(h History) Status {
	if o.Count < 0 {
		return StatusOpen
	}
	return Objective{Kind: KindWinExactlyTricks, Owner: o.Owner, Count: o.Count}.winExactlyTricks(h)
}

// avoidTricks (71–73): the owner wins none of the listed tricks.
func (o Objective) avoidTricks(h History) Status {
	done := true
	for _, n := range o.Tricks {
		switch n = h.trickNumber(n); {
		case n > len(h.Tricks):
			done = false
		case h.Tricks[n-1].Winner == o.Owner:
			return StatusFailed
		}
	}
	return decided(done, false)
}

// noConsecutive (75): never two tricks in a row.
func (o Objective) noConsecutive(h History) Status {
	for i := 1; i < len(h.Tricks); i++ {
		if h.Tricks[i].Winner == o.Owner && h.Tricks[i-1].Winner == o.Owner {
			return StatusFailed
		}
	}
	return decided(h.over(), false)
}

// consecutive (85, 86: at least Count in a row; 88, 89 and mission 32 with Exact: exactly Count
// tricks in total, all in a row).
func (o Objective) consecutive(h History) Status {
	run, best, total, gap := 0, 0, 0, false
	for _, t := range h.Tricks {
		if t.Winner != o.Owner {
			run = 0
			continue
		}
		if total > 0 && run == 0 {
			gap = true // a second run
		}
		run++
		total++
		best = max(best, run)
	}
	rem := h.remaining()
	if !o.Exact {
		if best >= o.Count {
			return StatusFulfilled
		}
		return decided(false, run+rem < o.Count)
	}
	fail := total > o.Count || gap ||
		(total > 0 && run == 0 && total < o.Count) || total+rem < o.Count
	return decided(h.over() && total == o.Count, fail)
}

// winWith (9–15, 58, 59): the owner wins a trick whose winning card matches With, with another card
// matching Target in it (if set). Fails once no unplayed cards can make such a trick.
func (o Objective) winWith(h History) Status {
	for _, t := range h.Tricks {
		if t.Winner != o.Owner {
			continue
		}
		w := t.WinningCard()
		if !o.With.Matches(w) {
			continue
		}
		if o.Target.IsZero() || slices.ContainsFunc(t.Plays, func(p Play) bool { return p.Card != w && o.Target.Matches(p.Card) }) {
			return StatusFulfilled
		}
	}
	left := h.unplayed()
	for _, w := range left {
		if !o.With.Matches(w) {
			continue
		}
		if o.Target.IsZero() || slices.ContainsFunc(left, func(c Card) bool { return c != w && o.Target.Matches(c) }) {
			return StatusOpen
		}
	}
	return StatusFailed
}

// winTrickWhere (7, 8, 46–50): the owner wins a trick without trump whose cards meet Where.
func (o Objective) winTrickWhere(h History) Status {
	for _, t := range h.Tricks {
		if t.Winner == o.Owner && o.trickMeets(t) {
			return StatusFulfilled
		}
	}
	return StatusOpen
}

// trickMeets reports whether t has no trump and its values meet o.Where.
func (o Objective) trickMeets(t TrickRecord) bool {
	sum := 0
	for _, p := range t.Plays {
		r := int(p.Card.Rank)
		if p.Card.IsTrump() {
			return false
		}
		sum += r
		switch {
		case o.Where == WhereBelow && r >= o.Count,
			o.Where == WhereAbove && r <= o.Count,
			o.Where == WhereEven && r%2 != 0,
			o.Where == WhereOdd && r%2 == 0:
			return false
		}
	}
	switch o.Where {
	case WhereSumAtLeast:
		return sum >= o.Count
	case WhereSumBelow:
		return sum < o.Count
	case WhereSumBetween:
		return sum >= o.Count && sum <= o.Max
	}
	return true
}

// winCount (20–26, 38–43, 51, 55–57, 62–70): the number of won cards matching Suits and Ranks is
// exactly or at least Count.
func (o Objective) winCount(h History) Status {
	pred := o.matcher()
	won, played := h.cardCounts(o.Owner, pred)
	left := deckCount(pred) - played
	switch o.Cmp {
	case CmpAtLeast:
		return decided(won >= o.Count, won+left < o.Count)
	case CmpExactly:
		return decided((left == 0 || h.over()) && won == o.Count, won > o.Count || won+left < o.Count)
	}
	panic("domain: winCount without a comparison")
}

// winEachColour (44): at least one won card of each colour.
func (o Objective) winEachColour(h History) Status {
	all := true
	for _, s := range Colours {
		won, played := h.cardCounts(o.Owner, isSuit(s))
		if won == 0 {
			all = false
			if played == suitSize(s) {
				return StatusFailed
			}
		}
	}
	return decided(all, false)
}

// winAllOfAnyColour (45): every card of at least one colour.
func (o Objective) winAllOfAnyColour(h History) Status {
	dead := 0
	for _, s := range Colours {
		won, played := h.cardCounts(o.Owner, isSuit(s))
		if won == suitSize(s) {
			return StatusFulfilled
		}
		if won < played {
			dead++
		}
	}
	return decided(false, dead == len(Colours))
}

// neverLead (60, 61; mission 12 for the crew): the owner (crew: anybody) never leads a card of
// Suits.
func (o Objective) neverLead(h History) Status {
	for _, t := range h.Tricks {
		if (o.Owner == OwnerCrew || t.Leader == o.Owner) && slices.Contains(o.Suits, t.Plays[0].Card.Suit) {
			return StatusFailed
		}
	}
	return decided(h.over(), false)
}

// equalSuits (92: overall; 93, 94: in one trick): as many cards of Suits[0] as of Suits[1], at
// least one each.
func (o Objective) equalSuits(h History) Status {
	a, b := isSuit(o.Suits[0]), isSuit(o.Suits[1])
	if o.InOneTrick {
		for _, t := range h.Tricks {
			if t.Winner != o.Owner {
				continue
			}
			na, nb := 0, 0
			for _, p := range t.Plays {
				if a(p.Card) {
					na++
				}
				if b(p.Card) {
					nb++
				}
			}
			if na > 0 && na == nb {
				return StatusFulfilled
			}
		}
		left := h.unplayed()
		return decided(false, !slices.ContainsFunc(left, a) || !slices.ContainsFunc(left, b))
	}
	wa, pa := h.cardCounts(o.Owner, a)
	wb, pb := h.cardCounts(o.Owner, b)
	la, lb := deckCount(a)-pa, deckCount(b)-pb
	settled := h.over() || (la == 0 && lb == 0)
	return decided(settled && wa > 0 && wa == wb, max(wa, wb, 1) > min(wa+la, wb+lb))
}

// moreSuit (95, 96): more won cards of Suits[0] than of Suits[1] (zero of Suits[1] allowed).
func (o Objective) moreSuit(h History) Status {
	wa, pa := h.cardCounts(o.Owner, isSuit(o.Suits[0]))
	wb, pb := h.cardCounts(o.Owner, isSuit(o.Suits[1]))
	if h.over() {
		return decided(wa > wb, wa <= wb)
	}
	la, lb := suitSize(o.Suits[0])-pa, suitSize(o.Suits[1])-pb
	return decided(wa > wb+lb, wa+la <= wb)
}

// all (36, 37, 52, 53): every part, owned by the owner.
func (o Objective) all(h History) Status {
	s := StatusFulfilled
	for _, p := range o.Parts {
		p.Owner = o.Owner
		switch p.status(h) {
		case StatusFailed:
			return StatusFailed
		case StatusOpen:
			s = StatusOpen
		}
	}
	return s
}

// leaderStaysAhead (mission 23): whoever wins the first trick has more tricks than every other
// member after every trick.
func (o Objective) leaderStaysAhead(h History) Status {
	if len(h.Tricks) == 0 {
		return StatusOpen
	}
	lead := h.Tricks[0].Winner
	won := make([]int, h.Seats)
	for _, t := range h.Tricks {
		won[t.Winner]++
		for s, n := range won {
			if s != lead && n >= won[lead] {
				return StatusFailed
			}
		}
	}
	return decided(h.over(), false)
}

// lastCardPlayed (mission 27): Cards[0] is the last card played in the last trick (with 3 players
// it fails when it is the card left over).
func (o Objective) lastCardPlayed(h History) Status {
	c := o.Cards[0]
	for i, t := range h.Tricks {
		if !t.contains(c) {
			continue
		}
		return decided(true, i+1 < h.Total || t.Plays[len(t.Plays)-1].Card != c)
	}
	return StatusOpen
}

// trickQuota (Q2): every seat wins exactly Targets[seat] tricks; open until the targets are set.
func (o Objective) trickQuota(h History) Status {
	if o.Targets == nil {
		return StatusOpen
	}
	won, rem := h.wonCounts(), h.remaining()
	for s, n := range won {
		if n > o.Targets[s] || n+rem < o.Targets[s] {
			return StatusFailed
		}
	}
	return decided(h.over(), false)
}
