package domain

import "slices"

// Objectives (rules §6 and §13): tasks and special objectives are evaluated the same way, as a
// pure function of the completed tricks. Statuses are never stored.
//
// The rule for every kind: an objective is fulfilled once its condition is met and can no longer
// fail; it fails as soon as it can no longer be met. When the last trick has been played, every
// objective that is still open fails (Status does that for all kinds).

// Status is the state of an objective.
type Status string

const (
	StatusOpen      Status = "open"
	StatusFulfilled Status = "fulfilled"
	StatusFailed    Status = "failed"
)

// ObjectiveKind is the kind of an objective. It is an open set: a later edition adds kinds.
type ObjectiveKind string

const (
	KindWinCards         ObjectiveKind = "winCards"         // a task: the owner wins Cards
	KindWinNoTricks      ObjectiveKind = "winNoTricks"      // the owner wins no trick
	KindWinExactlyTricks ObjectiveKind = "winExactlyTricks" // the owner wins exactly Count tricks
	KindWinOnlyTricks    ObjectiveKind = "winOnlyTricks"    // the owner wins exactly the Tricks
	KindWinTricks        ObjectiveKind = "winTricks"        // the owner wins each of the Tricks
	KindWinAllOfSuit     ObjectiveKind = "winAllOfSuit"     // the owner wins every card of Suit
	KindCrewWinWithAny   ObjectiveKind = "crewWinWithAny"   // Count tricks are won by one of Cards
	KindCrewWinWithEach  ObjectiveKind = "crewWinWithEach"  // each of Cards wins a trick
	KindCrewNeverWinWith ObjectiveKind = "crewNeverWinWith" // no trick is won by a colour of Rank
	KindCrewBalance      ObjectiveKind = "crewBalance"      // no member ever leads by more than Count tricks
	KindCardInLastTrick  ObjectiveKind = "cardInLastTrick"  // Cards are won in the last trick

	// Deep Sea (deep-sea/rules.md §9 and the mission objectives of §8).
	KindCmpTricks         ObjectiveKind = "cmpTricks"         // Cmp more | fewer than each other member, moreThanAll: than all others together
	KindCmpCaptain        ObjectiveKind = "cmpCaptain"        // Cmp more | fewer | equal tricks than the captain (never the captain's task)
	KindPredict           ObjectiveKind = "predict"           // win exactly Count tricks, predicted after the selection (-1 until then); Hidden
	KindAvoidTricks       ObjectiveKind = "avoidTricks"       // the owner wins none of the Tricks
	KindNoConsecutive     ObjectiveKind = "noConsecutive"     // the owner never wins two tricks in a row
	KindConsecutive       ObjectiveKind = "consecutive"       // Count tricks in a row; Exact: exactly Count tricks in total, all in a row
	KindWinWith           ObjectiveKind = "winWith"           // win a trick whose winning card matches With, with another card matching Target (if set)
	KindWinTrickWhere     ObjectiveKind = "winTrickWhere"     // win a trick without trump whose cards meet Where
	KindWinCount          ObjectiveKind = "winCount"          // the won cards matching Suits and Ranks number Cmp exactly | atLeast Count
	KindWinEachColour     ObjectiveKind = "winEachColour"     // at least one won card of each colour
	KindWinAllOfAnyColour ObjectiveKind = "winAllOfAnyColour" // every card of at least one colour
	KindNeverLead         ObjectiveKind = "neverLead"         // the owner (crew: nobody) never leads a card of Suits
	KindEqualSuits        ObjectiveKind = "equalSuits"        // as many won cards of Suits[0] as of Suits[1], more than 0; InOneTrick
	KindMoreSuit          ObjectiveKind = "moreSuit"          // more won cards of Suits[0] than of Suits[1]
	KindAll               ObjectiveKind = "all"               // every one of Parts (owned by the owner)
	KindLeaderStaysAhead  ObjectiveKind = "leaderStaysAhead"  // the first trick's winner always has more tricks than every other member
	KindLastCardPlayed    ObjectiveKind = "lastCardPlayed"    // Cards[0] is the last card played in the last trick
	KindTrickQuota        ObjectiveKind = "trickQuota"        // Q2: every seat wins exactly Targets[seat] tricks
	KindNone              ObjectiveKind = "none"              // no condition: Q2's face-down tasks
)

// Comparison is the relation of a Deep Sea objective.
type Comparison string

const (
	CmpMore        Comparison = "more"
	CmpFewer       Comparison = "fewer"
	CmpEqual       Comparison = "equal"
	CmpMoreThanAll Comparison = "moreThanAll" // more than all others together
	CmpExactly     Comparison = "exactly"
	CmpAtLeast     Comparison = "atLeast"
)

// TrickCondition is what every card of a trick (winTrickWhere) must meet.
type TrickCondition string

const (
	WhereBelow      TrickCondition = "below"      // every card lower than Count
	WhereAbove      TrickCondition = "above"      // every card higher than Count
	WhereEven       TrickCondition = "even"       // only even values
	WhereOdd        TrickCondition = "odd"        // only odd values
	WhereSumAtLeast TrickCondition = "sumAtLeast" // the values add up to at least Count
	WhereSumBelow   TrickCondition = "sumBelow"   // the values add up to less than Count
	WhereSumBetween TrickCondition = "sumBetween" // the values add up to Count..Max
)

// CardMatch selects cards: Suit (trump for submarines) and/or Rank; the zero value matches nothing.
// A rank alone matches colour cards only.
type CardMatch struct {
	Suit Suit
	Rank Rank
}

// IsZero reports whether m is unset.
func (m CardMatch) IsZero() bool { return m == CardMatch{} }

// Matches reports whether c is selected by m.
func (m CardMatch) Matches(c Card) bool {
	switch {
	case m.IsZero():
		return false
	case m.Suit == "" && c.IsTrump():
		return false
	case m.Suit != "" && c.Suit != m.Suit:
		return false
	}
	return m.Rank == 0 || c.Rank == m.Rank
}

// Owner values besides a seat.
const (
	OwnerCrew       = -1 // the objective belongs to the whole crew
	OwnerUnassigned = -2 // the setup has not assigned the owner yet
)

// LastTrick in Objective.Tricks stands for the attempt's last trick (13, 10 or 8).
const LastTrick = -1

// Objective is a task or a special objective. Only the fields its Kind uses are set.
type Objective struct {
	Kind         ObjectiveKind
	Owner        int    // a seat, OwnerCrew, or OwnerUnassigned during the setup
	Cards        []Card // winCards, crewWinWithAny, crewWinWithEach (in order), cardInLastTrick
	Count        int    // winExactlyTricks: tricks; crewWinWithAny: tricks; crewBalance: max lead
	Tricks       []int  // winOnlyTricks, winTricks: 1-based trick numbers or LastTrick
	Suit         Suit   // winAllOfSuit
	Rank         Rank   // crewNeverWinWith
	NotWithTrump bool   // winExactlyTricks, winOnlyTricks: no counted trick may be won by trump
	Ordered      bool   // crewWinWithEach: Cards win in their order

	// Deep Sea. crewBalance with Rank counts the won colour cards of that rank instead of tricks;
	// winExactlyTricks with Without: no counted trick may contain a colour card of these ranks.
	Cmp        Comparison     // cmpTricks, cmpCaptain, winCount
	Suits      []Suit         // winCount, neverLead, equalSuits, moreSuit
	Ranks      []Rank         // winCount: colour cards of these ranks
	With       CardMatch      // winWith: the winning card
	Target     CardMatch      // winWith: another card in the trick
	Where      TrickCondition // winTrickWhere, with Count (and Max)
	Max        int            // winTrickWhere sumBetween
	Exact      bool           // consecutive
	InOneTrick bool           // equalSuits
	Hidden     bool           // predict: the prediction is secret until the end
	Parts      []Objective    // all
	Targets    []int          // trickQuota: per seat
	Without    []Rank         // winExactlyTricks
}

func (o Objective) clone() Objective {
	c := o
	c.Cards = slices.Clone(o.Cards)
	c.Tricks = slices.Clone(o.Tricks)
	c.Suits = slices.Clone(o.Suits)
	c.Ranks = slices.Clone(o.Ranks)
	c.Targets = slices.Clone(o.Targets)
	c.Without = slices.Clone(o.Without)
	if o.Parts != nil {
		c.Parts = make([]Objective, len(o.Parts))
		for i, p := range o.Parts {
			c.Parts[i] = p.clone()
		}
	}
	return c
}

// Play is one card played to a trick by a seat.
type Play struct {
	Seat int
	Card Card
}

// TrickRecord is a completed trick.
type TrickRecord struct {
	Leader int
	Plays  []Play // in play order; Plays[0] is the lead
	Winner int    // seat
}

// WinningCard is the card that won the trick.
func (r TrickRecord) WinningCard() Card {
	for _, p := range r.Plays {
		if p.Seat == r.Winner {
			return p.Card
		}
	}
	panic("domain: trick winner did not play")
}

// contains reports whether c was played to the trick.
func (r TrickRecord) contains(c Card) bool {
	for _, p := range r.Plays {
		if p.Card == c {
			return true
		}
	}
	return false
}

// History is what objectives are evaluated over.
type History struct {
	Tricks []TrickRecord // completed so far, in order
	Total  int           // tricks in the attempt (13, 13, 10, 8)
	Seats  int           // members, the dummy included
	// Captain is the commander's seat (cmpCaptain).
	Captain int
}

// over reports whether the last trick has been played.
func (h History) over() bool { return len(h.Tricks) >= h.Total }

// remaining is the number of tricks still to play.
func (h History) remaining() int { return h.Total - len(h.Tricks) }

// prefix returns the history after the first n tricks.
func (h History) prefix(n int) History {
	h.Tricks = h.Tricks[:n]
	return h
}

// trickNumber resolves an entry of Objective.Tricks to a 1-based trick number.
func (h History) trickNumber(t int) int {
	if t == LastTrick {
		return h.Total
	}
	return t
}

// Status evaluates o over h. Every objective still open after the last trick has failed.
func (o Objective) Status(h History) Status {
	s := o.status(h)
	if s == StatusOpen && h.over() {
		return StatusFailed
	}
	return s
}

// FulfilledAt returns the 1-based number of the trick in which o became fulfilled, or 0.
func (o Objective) FulfilledAt(h History) int {
	if o.Status(h) != StatusFulfilled {
		return 0
	}
	for n := 1; n <= len(h.Tricks); n++ {
		if o.Status(h.prefix(n)) == StatusFulfilled {
			return n
		}
	}
	return 0
}

// status evaluates o without the end-of-attempt rule.
func (o Objective) status(h History) Status {
	switch o.Kind {
	case KindWinCards:
		return o.winCards(h)
	case KindWinNoTricks:
		return o.winNoTricks(h)
	case KindWinExactlyTricks:
		return o.winExactlyTricks(h)
	case KindWinOnlyTricks:
		return o.winOnlyTricks(h)
	case KindWinTricks:
		return o.winTricks(h)
	case KindWinAllOfSuit:
		return o.winAllOfSuit(h)
	case KindCrewWinWithAny:
		return o.crewWinWithAny(h)
	case KindCrewWinWithEach:
		return o.crewWinWithEach(h)
	case KindCrewNeverWinWith:
		return o.crewNeverWinWith(h)
	case KindCrewBalance:
		return o.crewBalance(h)
	case KindCardInLastTrick:
		return o.cardInLastTrick(h)
	case KindCmpTricks:
		return o.cmpTricks(h)
	case KindCmpCaptain:
		return o.cmpCaptain(h)
	case KindPredict:
		return o.predict(h)
	case KindAvoidTricks:
		return o.avoidTricks(h)
	case KindNoConsecutive:
		return o.noConsecutive(h)
	case KindConsecutive:
		return o.consecutive(h)
	case KindWinWith:
		return o.winWith(h)
	case KindWinTrickWhere:
		return o.winTrickWhere(h)
	case KindWinCount:
		return o.winCount(h)
	case KindWinEachColour:
		return o.winEachColour(h)
	case KindWinAllOfAnyColour:
		return o.winAllOfAnyColour(h)
	case KindNeverLead:
		return o.neverLead(h)
	case KindEqualSuits:
		return o.equalSuits(h)
	case KindMoreSuit:
		return o.moreSuit(h)
	case KindAll:
		return o.all(h)
	case KindLeaderStaysAhead:
		return o.leaderStaysAhead(h)
	case KindLastCardPlayed:
		return o.lastCardPlayed(h)
	case KindTrickQuota:
		return o.trickQuota(h)
	case KindNone:
		return StatusFulfilled
	}
	panic("domain: unknown objective kind " + string(o.Kind))
}

// winCards (a task, rules §6). Fulfilled: the owner has won every card. Fails: anyone else wins
// a trick with one of the cards (and, at the end, a card that stays unplayed).
func (o Objective) winCards(h History) Status {
	won := 0
	for _, t := range h.Tricks {
		for _, c := range o.Cards {
			if t.contains(c) {
				if t.Winner != o.Owner {
					return StatusFailed
				}
				won++
			}
		}
	}
	if won == len(o.Cards) {
		return StatusFulfilled
	}
	return StatusOpen
}

// winNoTricks (mission 5). Fulfilled: after the last trick, the owner won none. Fails: the owner
// wins a trick.
func (o Objective) winNoTricks(h History) Status {
	for _, t := range h.Tricks {
		if t.Winner == o.Owner {
			return StatusFailed
		}
	}
	if h.over() {
		return StatusFulfilled
	}
	return StatusOpen
}

// winExactlyTricks (mission 33; Deep Sea 83, 84, 87, Q1). Fulfilled: after the last trick, the
// owner won exactly Count, none by trump if NotWithTrump, none containing a colour card of a rank
// in Without. Fails: a trick too many, a forbidden trick, or too few tricks left to reach Count.
func (o Objective) winExactlyTricks(h History) Status {
	won := 0
	for _, t := range h.Tricks {
		if t.Winner != o.Owner {
			continue
		}
		if o.NotWithTrump && t.WinningCard().IsTrump() {
			return StatusFailed
		}
		if slices.ContainsFunc(t.Plays, func(p Play) bool { return !p.Card.IsTrump() && slices.Contains(o.Without, p.Card.Rank) }) {
			return StatusFailed
		}
		won++
	}
	if won > o.Count || won+h.remaining() < o.Count {
		return StatusFailed
	}
	if h.over() {
		return StatusFulfilled
	}
	return StatusOpen
}

// winOnlyTricks (missions 41, 50). Fulfilled: after the last trick, the owner won exactly the
// listed tricks, none by trump if NotWithTrump. Fails: a listed trick goes to someone else, the
// owner wins an unlisted trick, or wins with trump.
func (o Objective) winOnlyTricks(h History) Status {
	for i, t := range h.Tricks {
		listed := o.lists(h, i+1)
		mine := t.Winner == o.Owner
		if listed != mine || (mine && o.NotWithTrump && t.WinningCard().IsTrump()) {
			return StatusFailed
		}
	}
	if h.over() {
		return StatusFulfilled
	}
	return StatusOpen
}

// winTricks (mission 34). Fulfilled: the owner won every listed trick. Fails: someone else wins a
// listed trick.
func (o Objective) winTricks(h History) Status {
	s := StatusFulfilled
	for _, n := range o.Tricks {
		switch n = h.trickNumber(n); {
		case n > len(h.Tricks):
			s = StatusOpen
		case h.Tricks[n-1].Winner != o.Owner:
			return StatusFailed
		}
	}
	return s
}

// lists reports whether the 1-based trick n is one of o.Tricks.
func (o Objective) lists(h History, n int) bool {
	for _, t := range o.Tricks {
		if h.trickNumber(t) == n {
			return true
		}
	}
	return false
}

// winAllOfSuit (mission 46). Fulfilled: the owner has won every card of the suit. Fails: anyone
// else wins a trick containing a card of the suit (and, at the end, one stays unplayed).
func (o Objective) winAllOfSuit(h History) Status {
	won := 0
	for _, t := range h.Tricks {
		for _, p := range t.Plays {
			if p.Card.Suit != o.Suit {
				continue
			}
			if t.Winner != o.Owner {
				return StatusFailed
			}
			won++
		}
	}
	if won == suitSize(o.Suit) {
		return StatusFulfilled
	}
	return StatusOpen
}

// suitSize is the number of cards of suit s in the deck.
func suitSize(s Suit) int {
	if s == SuitTrump {
		return int(MaxTrumpRank)
	}
	return int(MaxColourRank)
}

// crewWinWithAny (missions 9, 26). Fulfilled: Count tricks have one of the cards as their winning
// card. Fails: the unplayed cards (at most one per remaining trick) can no longer make up Count.
func (o Objective) crewWinWithAny(h History) Status {
	wins, played := 0, 0
	for _, t := range h.Tricks {
		if slices.Contains(o.Cards, t.WinningCard()) {
			wins++
		}
		for _, c := range o.Cards {
			if t.contains(c) {
				played++
			}
		}
	}
	if wins >= o.Count {
		return StatusFulfilled
	}
	if wins+min(len(o.Cards)-played, h.remaining()) < o.Count {
		return StatusFailed
	}
	return StatusOpen
}

// crewWinWithEach (missions 13, 44). Fulfilled: each card has been the winning card of a trick.
// Fails: one of the cards is played and does not win, too few tricks are left for the cards still
// to win, or with Ordered, a card wins before the card listed before it.
func (o Objective) crewWinWithEach(h History) Status {
	wonAt := make([]int, len(o.Cards)) // 1-based trick, 0 = not won yet
	for i, t := range h.Tricks {
		w := t.WinningCard()
		for k, c := range o.Cards {
			if !t.contains(c) {
				continue
			}
			if c != w {
				return StatusFailed
			}
			wonAt[k] = i + 1
		}
	}
	open := 0
	for k, at := range wonAt {
		if at == 0 {
			open++
			continue
		}
		if o.Ordered && k > 0 && (wonAt[k-1] == 0 || wonAt[k-1] > at) {
			return StatusFailed
		}
	}
	if open == 0 {
		return StatusFulfilled
	}
	if open > h.remaining() {
		return StatusFailed
	}
	return StatusOpen
}

// crewNeverWinWith (missions 16, 17). Fulfilled: after the last trick, no trick was won by a
// colour card of Rank. Fails: such a trick.
func (o Objective) crewNeverWinWith(h History) Status {
	for _, t := range h.Tricks {
		if w := t.WinningCard(); !w.IsTrump() && w.Rank == o.Rank {
			return StatusFailed
		}
	}
	if h.over() {
		return StatusFulfilled
	}
	return StatusOpen
}

// crewBalance (missions 29, 34; Deep Sea 8, 21 with Rank). Fulfilled: after the last trick, the
// condition always held. Fails: after some trick, a member has won more than Count tricks (with
// Rank: colour cards of that rank) more than another member (the dummy counts as a member).
func (o Objective) crewBalance(h History) Status {
	won := make([]int, h.Seats)
	for _, t := range h.Tricks {
		if o.Rank == 0 {
			won[t.Winner]++
		}
		for _, p := range t.Plays {
			if o.Rank > 0 && !p.Card.IsTrump() && p.Card.Rank == o.Rank {
				won[t.Winner]++
			}
		}
		if slices.Max(won)-slices.Min(won) > o.Count {
			return StatusFailed
		}
	}
	if h.over() {
		return StatusFulfilled
	}
	return StatusOpen
}

// cardInLastTrick (mission 48, the Ω task's card). Fulfilled: the cards are won in the last
// trick. Fails: one of them is played earlier. Who wins it is the task's business.
func (o Objective) cardInLastTrick(h History) Status {
	in := 0
	for i, t := range h.Tricks {
		for _, c := range o.Cards {
			if !t.contains(c) {
				continue
			}
			if i+1 < h.Total {
				return StatusFailed
			}
			in++
		}
	}
	if in == len(o.Cards) {
		return StatusFulfilled
	}
	return StatusOpen
}
