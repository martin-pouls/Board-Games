package domain

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Per-card rule tests for the Deep Sea task cards (deep-sea/rules.md §6 and §9) and the mission
// objectives (§8). The expectations come from the rules text, not from the implementation.
//
// Every case builds a History in a small DSL (tkBuild) and states the Status the rules force.
// Trick specs (seats are numbers, the dummy is the last of the members):
//
//	"0:p3 1:p5 2:p9=2"   explicit: seat:card in play order, "=seat" is the winner (checked)
//	"2|p3 b1"            the cards are won by seat 2 (best card leads); missing cards come from the pool
//	"~1:4"               4 filler tricks won (and led) by seat 1, taken from the unused cards
//	"~1"                 filler tricks won by seat 1 up to the end of the attempt
//
// Cards that no spec uses form the pool for the missing cards and fillers. A history never uses a
// card twice. `left` stays unplayed (3 members: the 40th card), `reserve` is used last.

const (
	tkF = StatusFulfilled
	tkO = StatusOpen
	tkX = StatusFailed
)

type tkCase struct {
	id      int    // task card, or 0
	mission string // mission objective (name prefix), with obj
	name    string
	obj     *Objective
	members int // 0 = 3 (the dummy included)
	owner   int
	captain int
	count   int // cards 90, 91: the prediction
	tricks  []string
	left    string
	reserve string
	noLead  string // suit letters fillers must not lead (then "~s" means: led by seat s)
	want    Status
}

func tkc(id int, name string, want Status, tricks ...string) tkCase {
	return tkCase{id: id, name: name, want: want, tricks: tricks}
}

func tkm(mission, name string, o Objective, want Status, tricks ...string) tkCase {
	return tkCase{mission: mission, name: name, obj: &o, want: want, tricks: tricks}
}

func (c tkCase) mem(n int) tkCase       { c.members = n; return c }
func (c tkCase) own(s int) tkCase       { c.owner = s; return c }
func (c tkCase) capt(s int) tkCase      { c.captain = s; return c }
func (c tkCase) cnt(n int) tkCase       { c.count = n; return c }
func (c tkCase) lft(s string) tkCase    { c.left = s; return c }
func (c tkCase) rsv(s string) tkCase    { c.reserve = s; return c }
func (c tkCase) nolead(s string) tkCase { c.noLead = s; return c }

// tkWins returns one filler trick per entry, won by that seat: "0 1 0".
func tkWins(s string) []string {
	var out []string
	for _, f := range strings.Fields(s) {
		out = append(out, "~"+f+":1")
	}
	return out
}

// tkCat joins strings and string slices into one trick list.
func tkCat(parts ...any) []string {
	var out []string
	for _, p := range parts {
		switch v := p.(type) {
		case string:
			out = append(out, v)
		case []string:
			out = append(out, v...)
		}
	}
	return out
}

func tkDeck() []Card {
	var out []Card
	for _, s := range []Suit{SuitPink, SuitBlue, SuitGreen, SuitYellow} {
		for r := Rank(1); r <= MaxColourRank; r++ {
			out = append(out, Card{s, r})
		}
	}
	for r := Rank(1); r <= MaxTrumpRank; r++ {
		out = append(out, Card{SuitTrump, r})
	}
	return out
}

func tkCards(s string) []Card {
	var out []Card
	for _, f := range strings.Fields(s) {
		out = append(out, c(f))
	}
	return out
}

// tkBestFirst builds a trick won and led by seat w: the best card (a submarine, else the highest
// rank) leads, so it wins whatever the others are.
func tkBestFirst(members, w int, cards []Card) TrickRecord {
	best := 0
	for i, x := range cards {
		b := cards[best]
		if (x.IsTrump() && !b.IsTrump()) || (x.IsTrump() == b.IsTrump() && x.Rank > b.Rank) {
			best = i
		}
	}
	order := []Card{cards[best]}
	for i, x := range cards {
		if i != best {
			order = append(order, x)
		}
	}
	r := TrickRecord{Leader: w, Winner: w}
	for i, x := range order {
		r.Plays = append(r.Plays, Play{Seat: (w + i) % members, Card: x})
	}
	return r
}

func tkAtoi(t *testing.T, s, spec string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("bad trick spec %q", spec)
	}
	return n
}

// tkBuild turns the case's trick specs into a History.
func tkBuild(t *testing.T, tc tkCase) History {
	t.Helper()
	members := tc.members
	if members == 0 {
		members = 3
	}
	total := TotalTricks(members)
	used := map[Card]bool{}
	use := func(cards ...Card) {
		for _, x := range cards {
			if used[x] {
				t.Fatalf("card %s used twice in the history", x.shortString())
			}
			used[x] = true
		}
	}
	left := tkCards(tc.left)
	use(left...)

	type spec struct {
		kind   byte // e explicit, a auto, f filler
		rec    TrickRecord
		winner int
		cards  []Card
		n      int // filler: -1 = up to the end
	}
	var specs []spec
	fixed, pads := 0, 0
	for _, s := range tc.tricks {
		switch {
		case strings.HasPrefix(s, "~"):
			w, n, hasN := strings.Cut(s[1:], ":")
			sp := spec{kind: 'f', winner: tkAtoi(t, w, s), n: -1}
			if hasN {
				sp.n = tkAtoi(t, n, s)
				fixed += sp.n
			} else {
				pads++
			}
			specs = append(specs, sp)
		case strings.Contains(s, "|"):
			w, rest, _ := strings.Cut(s, "|")
			sp := spec{kind: 'a', winner: tkAtoi(t, w, s), cards: tkCards(rest)}
			use(sp.cards...)
			if len(sp.cards) > members {
				t.Fatalf("trick %q has more cards than members", s)
			}
			fixed++
			specs = append(specs, sp)
		default:
			body, win, ok := strings.Cut(s, "=")
			if !ok {
				t.Fatalf("explicit trick %q needs =winner", s)
			}
			want := tkAtoi(t, win, s)
			rec := trick(strings.Fields(body)...)
			if len(rec.Plays) != members {
				t.Fatalf("trick %q has %d plays, want %d", s, len(rec.Plays), members)
			}
			seen := map[int]bool{}
			for _, p := range rec.Plays {
				if p.Seat >= members || seen[p.Seat] {
					t.Fatalf("trick %q: bad seats", s)
				}
				seen[p.Seat] = true
				use(p.Card)
			}
			if rec.Winner != want {
				t.Fatalf("trick %q: the winner is seat %d, not %d", s, rec.Winner, want)
			}
			fixed++
			specs = append(specs, spec{kind: 'e', rec: rec})
		}
	}
	if pads > 1 {
		t.Fatalf("more than one open filler marker")
	}
	pad := total - fixed
	if pad < 0 || (pads == 0 && fixed > total) {
		t.Fatalf("too many tricks: %d fixed, %d in total", fixed, total)
	}

	reserve := tkCards(tc.reserve)
	var pool, late []Card
	for _, x := range tkDeck() {
		switch {
		case used[x]:
		case slices.Contains(reserve, x):
			late = append(late, x)
		default:
			pool = append(pool, x)
		}
	}
	pool = append(pool, late...)
	take := func(n int) []Card {
		if len(pool) < n {
			t.Fatalf("the pool of unused cards is empty")
		}
		out := slices.Clone(pool[:n])
		pool = pool[n:]
		return out
	}
	forbidden := map[Suit]bool{}
	for _, l := range tc.noLead {
		forbidden[suitLetters[byte(l)]] = true
	}
	popWhere := func(pred func(Card) bool) (Card, bool) {
		for i, x := range pool {
			if pred(x) {
				pool = slices.Delete(slices.Clone(pool), i, i+1)
				return x, true
			}
		}
		return Card{}, false
	}
	filler := func(w int) TrickRecord {
		if len(forbidden) == 0 {
			return tkBestFirst(members, w, take(members))
		}
		lead, ok := popWhere(func(x Card) bool { return !forbidden[x.Suit] })
		if !ok {
			t.Fatalf("no card left that may lead")
		}
		cards := []Card{lead}
		for len(cards) < members {
			x, ok := popWhere(func(x Card) bool { return forbidden[x.Suit] })
			if !ok {
				x, ok = popWhere(func(Card) bool { return true })
			}
			if !ok {
				t.Fatalf("the pool of unused cards is empty")
			}
			cards = append(cards, x)
		}
		r := TrickRecord{Leader: w}
		for i, x := range cards {
			r.Plays = append(r.Plays, Play{Seat: (w + i) % members, Card: x})
		}
		r.Winner = r.Plays[trickWinner(cards)].Seat
		return r
	}

	var recs []TrickRecord
	for _, sp := range specs {
		switch sp.kind {
		case 'e':
			recs = append(recs, sp.rec)
		case 'a':
			cards := slices.Clone(sp.cards)
			cards = append(cards, take(members-len(cards))...)
			recs = append(recs, tkBestFirst(members, sp.winner, cards))
		case 'f':
			n := sp.n
			if n < 0 {
				n = pad
			}
			for range n {
				recs = append(recs, filler(sp.winner))
			}
		}
	}
	if len(recs) > total {
		t.Fatalf("%d tricks, but the attempt has %d", len(recs), total)
	}
	if len(recs) == total && len(pool)+len(left) != 40-total*members {
		t.Fatalf("complete history leaves %d cards unplayed, want %d", len(pool)+len(left), 40-total*members)
	}
	return History{Tricks: recs, Total: total, Seats: members, Captain: tc.captain}
}

func tkObjective(t *testing.T, tc tkCase) Objective {
	t.Helper()
	if tc.obj != nil {
		return tc.obj.clone()
	}
	card, ok := TaskCardFor(tc.id)
	if !ok {
		t.Fatalf("no task card %d", tc.id)
	}
	members := tc.members
	if members == 0 {
		members = 3
	}
	o := card.objectiveFor(members)
	o.Owner = tc.owner
	if o.Kind == KindPredict {
		o.Count = tc.count
	}
	return o
}

func tkFormat(h History) string {
	var sb strings.Builder
	for i, tr := range h.Tricks {
		fmt.Fprintf(&sb, "\n  %2d: lead %d:", i+1, tr.Leader)
		for _, p := range tr.Plays {
			fmt.Fprintf(&sb, " %d:%s", p.Seat, p.Card.shortString())
		}
		fmt.Fprintf(&sb, "  -> %d", tr.Winner)
	}
	return sb.String()
}

func tkName(tc tkCase) string {
	if tc.mission != "" {
		return fmt.Sprintf("%s/%s", tc.mission, tc.name)
	}
	return fmt.Sprintf("card%02d/%s", tc.id, tc.name)
}

func tkRun(t *testing.T, cases []tkCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tkName(tc), func(t *testing.T) {
			h := tkBuild(t, tc)
			o := tkObjective(t, tc)
			if got := o.Status(h); got != tc.want {
				t.Errorf("status = %s, want %s (%d members, owner %d, captain %d)%s", got, tc.want, h.Seats, tc.owner, tc.captain, tkFormat(h))
			}
			// A verdict never reverts (rules §6: fulfilled can no longer fail, failed can no
			// longer be met).
			verdict := tkO
			for n := 0; n <= len(h.Tricks); n++ {
				s := o.Status(h.prefix(n))
				if verdict != tkO && s != verdict {
					t.Errorf("after %d tricks the status is %s, but it was already %s%s", n, s, verdict, tkFormat(h))
					break
				}
				verdict = s
			}
		})
	}
}

// tkWinCards are the cards of the winCards tasks, from rules.md §9.
var tkWinCards = map[int]string{
	16: "p3", 17: "y1", 18: "b4", 19: "g6",
	27: "b1 b2 b3", 28: "b6 y7", 29: "p5 y6", 30: "g5 b8", 31: "b5 p8", 32: "p9 y8", 33: "p1 g7",
	34: "y9 b7", 35: "g3 y4 y5", 54: "t3",
}

// tkWinCardsCases generates the cases of the "win these cards" tasks.
func tkWinCardsCases() []tkCase {
	var out []tkCase
	for id := 1; id <= TaskCards; id++ {
		spec, ok := tkWinCards[id]
		if !ok {
			continue
		}
		cards := strings.Fields(spec)
		k := len(cards)
		each := func(winner int, cs []string) []string {
			var tr []string
			for _, x := range cs {
				tr = append(tr, fmt.Sprintf("%d|%s", winner, x))
			}
			return tr
		}
		out = append(out,
			tkc(id, "every card won in its own trick", tkF, each(0, cards)...).rsv(spec),
			tkc(id, "the dummy owns it and wins every card", tkF, each(2, cards)...).own(2).rsv(spec),
			tkc(id, "nothing played yet", tkO).rsv(spec),
			tkc(id, "another seat wins the first card", tkX, "1|"+cards[0]).rsv(spec),
			tkc(id, "another seat wins the last card", tkX, tkCat(each(0, cards[:k-1]), "1|"+cards[k-1])...).rsv(spec),
			tkc(id, "a card stays unplayed (the left-over card)", tkX, tkCat(each(0, cards[:k-1]), "~1")...).lft(cards[k-1]).rsv(spec),
		)
		if k > 1 {
			out = append(out,
				tkc(id, "all cards in one trick won by the owner", tkF, "0|"+spec).rsv(spec),
				tkc(id, "all cards in one trick won by another seat", tkX, "1|"+spec).rsv(spec),
				tkc(id, "one card won, the rest not played yet", tkO, "0|"+cards[0]).rsv(spec),
			)
		}
	}
	return out
}

func tkTaskCases() []tkCase {
	cs := []tkCase{
		// 1: more tricks than each other member (the dummy is a member)
		tkc(1, "seven of 13 can no longer be caught", tkF, "~0:7"),
		tkc(1, "most tricks at the end", tkF, "~0:5", "~1:4", "~2:4"),
		tkc(1, "a tie for the most fails", tkX, "~0:6", "~1:6", "~2:1"),
		tkc(1, "the dummy counts as a member", tkX, "~0:6", "~2:6", "~1:1"),
		tkc(1, "another member takes seven early", tkX, "~1:7"),
		tkc(1, "open", tkO, "~0:2", "~1:2"),
		tkc(1, "4 members: six of ten is safe", tkF, "~0:6").mem(4),
		tkc(1, "4 members: a tie at 4/4/2/0", tkX, "~0:4", "~1:4", "~2:2").mem(4),
		tkc(1, "5 members: five of eight is safe", tkF, "~0:5").mem(5),

		// 2: more tricks than all others together
		tkc(2, "seven of 13 beat the others together", tkF, "~0:7"),
		tkc(2, "4 members: six of ten", tkF, "~0:6").mem(4),
		tkc(2, "4 members: five equal the other five", tkX, "~0:5", "~1:3", "~2:2").mem(4),
		tkc(2, "the others together win seven", tkX, "~1:4", "~2:3"),
		tkc(2, "the dummy is among the others", tkX, "~0:6", "~1:3", "~2:4"),
		tkc(2, "open", tkO, "~0:2", "~1:2", "~2:1"),

		// 3: fewer tricks than each other member
		tkc(3, "far behind both others with one trick left", tkF, "~1:6", "~2:6"),
		tkc(3, "fewest at the end", tkF, "~0:1", "~1:6", "~2:6"),
		tkc(3, "no trick at all", tkF, "~1:7", "~2:6"),
		tkc(3, "a tie with one other fails", tkX, "~0:4", "~1:4", "~2:5"),
		tkc(3, "fewer than one but not the other", tkX, "~0:3", "~1:2", "~2:8"),
		tkc(3, "seven tricks can no longer be fewer", tkX, "~0:7"),
		tkc(3, "the dummy has fewer", tkX, "~0:2", "~1:10", "~2:1"),
		tkc(3, "open", tkO, "~0:1", "~1:1"),

		// 4: more tricks than the captain (seat 1), never the captain's own task
		tkc(4, "more than the captain", tkF, "~0:4", "~1:3", "~2:6").capt(1),
		tkc(4, "equal fails", tkX, "~0:3", "~1:3", "~2:7").capt(1),
		tkc(4, "fewer fails", tkX, "~0:2", "~1:5", "~2:6").capt(1),
		tkc(4, "seven tricks beat any captain", tkF, "~0:7").capt(1),
		tkc(4, "the captain takes seven", tkX, "~1:7").capt(1),
		tkc(4, "open", tkO, "~0:1", "~1:1").capt(1),
		tkc(4, "the dummy as owner (2 players) is more", tkF, "~2:5", "~0:4", "~1:4").own(2).capt(0),
		tkc(4, "the dummy as owner ties", tkX, "~2:4", "~0:4", "~1:5").own(2).capt(0),

		// 5: fewer tricks than the captain
		tkc(5, "fewer than the captain", tkF, "~0:1", "~1:5", "~2:7").capt(1),
		tkc(5, "equal fails", tkX, "~0:4", "~1:4", "~2:5").capt(1),
		tkc(5, "more fails", tkX, "~0:5", "~1:3", "~2:5").capt(1),
		tkc(5, "the captain takes seven", tkF, "~1:7").capt(1),
		tkc(5, "the owner takes seven", tkX, "~0:7").capt(1),
		tkc(5, "open", tkO, "~0:1", "~1:1").capt(1),
		tkc(5, "the dummy as owner has fewer", tkF, "~2:1", "~0:5", "~1:7").own(2).capt(0),

		// 6: the same number of tricks as the captain
		tkc(6, "equal at the end", tkF, "~0:4", "~1:4", "~2:5").capt(1),
		tkc(6, "unequal", tkX, "~0:4", "~1:5", "~2:4").capt(1),
		tkc(6, "both without a trick", tkF, "~2:13").capt(1),
		tkc(6, "seven tricks can no longer be matched", tkX, "~0:7").capt(1),
		tkc(6, "open", tkO, "~0:2", "~1:2").capt(1),
		tkc(6, "4 members equal", tkF, "~0:3", "~1:3", "~2:2", "~3:2").mem(4).capt(1),

		// 7: a trick with all cards lower than 7, no submarine in it
		tkc(7, "all lower than 7", tkF, "0|p6 b2 g3"),
		tkc(7, "4 members all lower than 7", tkF, "0|p6 b2 g3 y4").mem(4),
		tkc(7, "a 7 spoils it", tkX, "0|p7 b2 g3", "~1"),
		tkc(7, "a submarine spoils it", tkX, "0|t1 b2 g3", "~1"),
		tkc(7, "won by another member", tkX, "1|p6 b2 g3", "~1"),
		tkc(7, "open", tkO, "0|p8 b2 g3"),

		// 8: a trick with all cards greater than 5
		tkc(8, "all greater than 5", tkF, "0|p9 b6 g7"),
		tkc(8, "6 is greater than 5", tkF, "0|p6 b7 g8"),
		tkc(8, "a 5 spoils it", tkX, "0|p9 b5 g8", "~1"),
		tkc(8, "a submarine spoils it", tkX, "0|t1 b9 g8", "~1"),
		tkc(8, "won by another member", tkX, "1|p9 b6 g7", "~1"),
		tkc(8, "open", tkO, "0|p9 b5 g8"),

		// 9, 10, 11, 15: win a trick with a 6, 5, 3, 2 (the winning card is that colour card)
		tkc(9, "wins with the 6", tkF, "0:p6 1:b2 2:g3=0"),
		tkc(9, "higher off-colour cards do not matter", tkF, "0:p6 1:b9 2:g9=0"),
		tkc(9, "wins with a 7 although a 6 is in the trick", tkX, "0:p7 1:p6 2:b1=0", "~1"),
		tkc(9, "another member wins with the 6", tkX, "1:p6 2:b2 0:g3=1", "~1"),
		tkc(9, "a submarine wins", tkX, "0:t1 1:p6 2:b1=0", "~1"),
		tkc(9, "open", tkO, "0:p7 1:b2 2:g3=0"),

		tkc(10, "wins with the 5", tkF, "0:p5 1:b1 2:g2=0"),
		tkc(10, "wins with a 6 although a 5 is in the trick", tkX, "0:p6 1:b5 2:g1=0", "~1"),
		tkc(10, "another member wins with the 5", tkX, "1:p5 2:b1 0:g2=1", "~1"),
		tkc(10, "a submarine wins", tkX, "0:t2 1:p5 2:b1=0", "~1"),
		tkc(10, "open", tkO, "0:p6 1:b2 2:g1=0"),

		tkc(11, "wins with the 3", tkF, "0:p3 1:b9 2:g1=0"),
		tkc(11, "submarine 3 is not a 3", tkX, "0:t3 1:p1 2:b2=0", "~1"),
		tkc(11, "wins with a 4 although a 3 is in the trick", tkX, "0:p4 1:b3 2:g1=0", "~1"),
		tkc(11, "another member wins with the 3", tkX, "1:p3 2:b1 0:g2=1", "~1"),
		tkc(11, "open", tkO, "0:p4 1:b1 2:g2=0"),

		tkc(15, "wins with the 2", tkF, "0:p2 1:b3 2:g1=0"),
		tkc(15, "submarine 2 is not a 2", tkX, "0:t2 1:p1 2:b3=0", "~1"),
		tkc(15, "wins with a 3 although a 2 is in the trick", tkX, "0:p3 1:b2 2:g1=0", "~1"),
		tkc(15, "another member wins with the 2", tkX, "1:p2 2:b1 0:g3=1", "~1"),
		tkc(15, "open", tkO, "0:p3 1:b1 2:g1=0"),

		// 12: win a 5 with a 7
		tkc(12, "7 wins a trick with a 5", tkF, "0:p7 1:p5 2:b1=0"),
		tkc(12, "the 5 may be off-colour", tkF, "0:p7 1:b5 2:g1=0"),
		tkc(12, "no 5 in the trick", tkX, "0:p7 1:p4 2:b1=0", "~1"),
		tkc(12, "a 9 wins the trick with 7 and 5", tkX, "0:p9 1:p7 2:p5=0", "~1"),
		tkc(12, "the 5 wins, the 7 is off-colour", tkX, "0:p5 1:b7 2:g1=0", "~1"),
		tkc(12, "another member wins with the 7", tkX, "0:p1 1:p7 2:p5=1", "~1"),
		tkc(12, "open", tkO, "0:p7 1:p4 2:b1=0"),

		// 13: win an 8 with a 4
		tkc(13, "4 wins a trick with an 8", tkF, "0:p4 1:b8 2:g1=0"),
		tkc(13, "submarine 4 is not a 4", tkX, "0:t4 1:p8 2:b1=0", "~1"),
		tkc(13, "the 8 wins", tkX, "0:p8 1:b4 2:g1=0", "~1"),
		tkc(13, "no 8 in the trick", tkX, "0:p4 1:b7 2:g1=0", "~1"),
		tkc(13, "another member wins with the 4", tkX, "1:p4 2:b8 0:g1=1", "~1"),
		tkc(13, "open", tkO, "0:p4 1:b7 2:g1=0"),

		// 14: win any 6 with another 6
		tkc(14, "a 6 wins a trick with another 6", tkF, "0:p6 1:b6 2:g1=0"),
		tkc(14, "three 6s", tkF, "0:p6 1:b6 2:g6=0"),
		tkc(14, "only one 6 in the trick", tkX, "0:p6 1:b2 2:g1=0", "~1"),
		tkc(14, "a 9 wins the trick with two 6s", tkX, "0:p9 1:p6 2:b6=0", "~1"),
		tkc(14, "a submarine wins", tkX, "0:t1 1:p6 2:b6=0", "~1"),
		tkc(14, "another member wins with a 6", tkX, "1:p6 2:b6 0:g1=1", "~1"),
		tkc(14, "open", tkO, "0:p6 1:b2 2:g1=0"),

		// 58, 59: win a colour card with a submarine (owner is seat 1)
		tkc(58, "a submarine wins the pink 7", tkF, "0:p7 1:t1 2:b1=1").own(1),
		tkc(58, "a higher submarine of the owner", tkF, "0:p7 1:t4 2:t1=1").own(1),
		tkc(58, "pink 7 won without a submarine", tkX, "0:p7 1:p9 2:b1=1", "~0").own(1),
		tkc(58, "a submarine wins another 7", tkX, "0:b7 1:t1 2:p2=1", "~0").own(1),
		tkc(58, "another seat wins with a higher submarine", tkX, "0:p7 1:t1 2:t4=2", "~0").own(1),
		tkc(58, "open", tkO, "0:b2 1:p3 2:g1=0").own(1),

		tkc(59, "a submarine wins the green 9", tkF, "0:g9 1:t2 2:b1=1").own(1),
		tkc(59, "a higher submarine of the owner", tkF, "0:g9 1:t4 2:t1=1").own(1),
		tkc(59, "green 9 won without a submarine", tkX, "0:g8 1:g9 2:b1=1", "~0").own(1),
		tkc(59, "a submarine wins another card", tkX, "0:g8 1:t1 2:p1=1", "~0").own(1),
		tkc(59, "another seat wins with a higher submarine", tkX, "0:g9 1:t1 2:t2=2", "~0").own(1),
		tkc(59, "open", tkO, "0:b2 1:p3 2:g1=0").own(1),

		// 36: win the green 2 in the last trick
		tkc(36, "wins the last trick with the green 2", tkF, "~1:12", "0:g9 1:g2 2:b1=0"),
		tkc(36, "4 members", tkF, "~1:9", "0:g9 1:g2 2:b1 3:y1=0").mem(4),
		tkc(36, "the last trick goes to another seat", tkX, "~0:12", "1:g9 0:g2 2:b1=1"),
		tkc(36, "the green 2 is played earlier", tkX, "0:g9 1:g2 2:b1=0", "~0:12"),
		tkc(36, "the green 2 stays unplayed", tkX, "~1").lft("g2"),
		tkc(36, "open", tkO, "~1:5"),
		tkc(36, "owner wins the last trick without the green 2", tkX, "~1:12", "0:g9 1:b1 2:y1=0"),

		// 37: exactly one pink and one green card
		tkc(37, "one pink and one green", tkF, "0:p5 1:b1 2:y1=0", "0:g5 1:b2 2:y2=0", "~1"),
		tkc(37, "both in one trick", tkF, "0:p5 1:g3 2:b1=0", "~1"),
		tkc(37, "open", tkO, "0:p5 1:b1 2:y1=0"),
		tkc(37, "two pink", tkX, "0:p5 1:p3 2:b1=0"),
		tkc(37, "two green", tkX, "0:g5 1:g3 2:b1=0"),
		tkc(37, "no green", tkX, "0:p5 1:b1 2:y1=0", "~1"),
		tkc(37, "no pink", tkX, "0:g5 1:b1 2:y1=0", "~1"),

		// 38: at least seven yellow cards
		tkc(38, "exactly seven", tkF, "0:y9 1:y8 2:y7=0", "0:y6 1:y5 2:y4=0", "0:y3 1:b1 2:p1=0"),
		tkc(38, "all nine", tkF, "0:y9 1:y8 2:y7=0", "0:y6 1:y5 2:y4=0", "0:y3 1:y2 2:y1=0"),
		tkc(38, "six at the end", tkX, "0:y9 1:y8 2:y7=0", "0:y6 1:y5 2:y4=0", "~1"),
		tkc(38, "three lost early", tkX, "1:y9 2:y8 0:y7=1"),
		tkc(38, "open", tkO, "0:y9 1:y8 2:y7=0"),

		// 39: at least five pink cards
		tkc(39, "five", tkF, "0:p9 1:p8 2:p7=0", "0:p6 1:p5 2:b1=0"),
		tkc(39, "six", tkF, "0:p9 1:p8 2:p7=0", "0:p6 1:p5 2:p4=0"),
		tkc(39, "four are open", tkO, "0:p9 1:p8 2:p7=0", "0:p6 1:b1 2:g1=0"),
		tkc(39, "four at the end", tkX, "0:p9 1:p8 2:p7=0", "0:p6 1:b1 2:g1=0", "~1"),
		tkc(39, "five lost early", tkX, "1:p9 2:p8 0:p7=1", "1:p6 2:p5 0:b1=1"),

		// 40: exactly two green cards
		tkc(40, "two at the end", tkF, "0:g9 1:g8 2:b1=0", "~1"),
		tkc(40, "two are open until the end", tkO, "0:g9 1:g8 2:b1=0"),
		tkc(40, "three", tkX, "0:g9 1:g8 2:g7=0"),
		tkc(40, "one at the end", tkX, "0:g9 1:b1 2:y1=0", "~1"),
		tkc(40, "none at the end", tkX, "~1"),
		tkc(40, "only one green left", tkX, "1|g9 g8 g7", "1|g6 g5 g4", "1|g3 g2 b1"),
		tkc(40, "two green left", tkO, "1|g9 g8 g7", "1|g6 g5 g4", "1|g3 b1 b2"),

		// 41: exactly two blue cards
		tkc(41, "two at the end", tkF, "0:b9 1:b8 2:g1=0", "~1"),
		tkc(41, "two are open until the end", tkO, "0:b9 1:b8 2:g1=0"),
		tkc(41, "three", tkX, "0:b9 1:b8 2:b7=0"),
		tkc(41, "none at the end", tkX, "~1"),
		tkc(41, "only one blue left", tkX, "1|b9 b8 b7", "1|b6 b5 b4", "1|b3 b2 g1"),

		// 42: exactly one pink card
		tkc(42, "one at the end", tkF, "0:p9 1:b1 2:y1=0", "~1"),
		tkc(42, "one is open until the end", tkO, "0:p9 1:b1 2:y1=0"),
		tkc(42, "two", tkX, "0:p9 1:p8 2:b1=0"),
		tkc(42, "none at the end", tkX, "~1"),
		tkc(42, "all pink lost", tkX, "1|p9 p8 p7", "1|p6 p5 p4", "1|p3 p2 p1"),
		tkc(42, "one pink left", tkO, "1|p9 p8 p7", "1|p6 p5 p4", "1|p3 p2 b1"),

		// 43, 62, 63, 64, 65: no cards of a colour
		tkc(43, "no pink at the end", tkF, "0:b9 1:b8 2:y1=0", "~1"),
		tkc(43, "pink won", tkX, "0:p9 1:b1 2:y1=0"),
		tkc(43, "open", tkO, "0:b9 1:b8 2:y1=0"),

		tkc(62, "no green at the end", tkF, "0:b9 1:y8 2:p1=0", "~1"),
		tkc(62, "green won", tkX, "0:g9 1:b1 2:y1=0"),
		tkc(62, "a green card played by another member", tkX, "0:b9 1:g1 2:y1=0"),
		tkc(62, "open", tkO, "0:b9 1:y8 2:p1=0"),

		tkc(63, "no yellow at the end", tkF, "0:b9 1:g8 2:p1=0", "~1"),
		tkc(63, "yellow won", tkX, "0:y9 1:b1 2:g1=0"),
		tkc(63, "a yellow card played by another member", tkX, "0:b9 1:y1 2:g1=0"),
		tkc(63, "open", tkO, "0:b9 1:g8 2:p1=0"),

		tkc(64, "no pink or blue at the end", tkF, "0:g9 1:y1 2:y2=0", "~1"),
		tkc(64, "pink won", tkX, "0:g9 1:p1 2:y1=0"),
		tkc(64, "blue won", tkX, "0:g9 1:b1 2:y1=0"),
		tkc(64, "open", tkO, "0:g9 1:y1 2:y2=0"),

		tkc(65, "no yellow or green at the end", tkF, "0:p9 1:b1 2:p1=0", "~1"),
		tkc(65, "green won", tkX, "0:p9 1:g1 2:b1=0"),
		tkc(65, "yellow won", tkX, "0:p9 1:y1 2:b1=0"),
		tkc(65, "open", tkO, "0:p9 1:b1 2:p1=0"),

		// 66, 67, 68, 69, 70: no cards of some ranks (colour cards only)
		tkc(66, "no 8 or 9 at the end", tkF, "0:p7 1:b6 2:g5=0", "~1"),
		tkc(66, "a submarine is neither", tkF, "0:t4 1:p7 2:b6=0", "~1"),
		tkc(66, "an 8 won", tkX, "0:p7 1:b8 2:g1=0"),
		tkc(66, "a 9 won", tkX, "0:p7 1:b9 2:g1=0"),
		tkc(66, "open", tkO, "0:p7 1:b6 2:g5=0"),

		tkc(67, "no 9 at the end", tkF, "0:p8 1:b7 2:g1=0", "~1"),
		tkc(67, "a 9 won", tkX, "0:p9 1:b1 2:g1=0"),
		tkc(67, "an off-colour 9 played by another member", tkX, "0:p8 1:b9 2:g1=0"),
		tkc(67, "open", tkO, "0:p8 1:b7 2:g1=0"),

		tkc(68, "no 5 at the end", tkF, "0:p6 1:b4 2:g1=0", "~1"),
		tkc(68, "a 5 won", tkX, "0:p6 1:b5 2:g1=0"),
		tkc(68, "open", tkO, "0:p6 1:b4 2:g1=0"),

		tkc(69, "no 1 at the end; submarine 1 is not a 1", tkF, "0:t1 1:p2 2:b3=0", "~1"),
		tkc(69, "a 1 won", tkX, "0:p2 1:b1 2:g3=0"),
		tkc(69, "open", tkO, "0:p2 1:b3 2:g4=0"),

		tkc(70, "no 1, 2 or 3 at the end; submarine 3 is not a 3", tkF, "0:t3 1:p9 2:b8=0", "~1"),
		tkc(70, "a 1 won", tkX, "0:p9 1:b1 2:g8=0"),
		tkc(70, "a 2 won", tkX, "0:p9 1:b2 2:g8=0"),
		tkc(70, "a 3 won", tkX, "0:p9 1:b3 2:g8=0"),
		tkc(70, "open", tkO, "0:p9 1:b4 2:g8=0"),

		// 20-26: counts of a rank
		tkc(20, "all four 3s", tkF, "0|p3 b3 g3", "0|y3"),
		tkc(20, "another member wins a 3", tkX, "0|p3 b3 g3", "1|y3"),
		tkc(20, "three are open", tkO, "0|p3 b3 g3"),
		tkc(20, "the fourth 3 stays unplayed", tkX, "0|p3 b3 g3", "~1").lft("y3"),
		tkc(20, "submarine 3 is not a 3", tkO, "0|p3 b3 g3", "0|t3"),

		tkc(21, "three 5s", tkF, "0|p5 b5 g5"),
		tkc(21, "three 5s in separate tricks", tkF, "0|p5", "0|b5", "0|g5"),
		tkc(21, "more than three is fine", tkF, "0|p5 b5 g5", "0|y5"),
		tkc(21, "two lost early", tkX, "1|p5 b5"),
		tkc(21, "two at the end", tkX, "0|p5 b5", "~1"),
		tkc(21, "open", tkO, "0|p5 b5"),

		tkc(22, "three 9s", tkF, "0|p9 b9 g9"),
		tkc(22, "four 9s", tkF, "0|p9 b9 g9", "0|y9"),
		tkc(22, "two lost early", tkX, "1|p9 b9"),
		tkc(22, "two at the end", tkX, "0|p9 b9", "~1"),
		tkc(22, "open", tkO, "0|p9 b9"),

		tkc(23, "two 7s", tkF, "0|p7 b7"),
		tkc(23, "two 7s in separate tricks", tkF, "0|p7", "0|b7"),
		tkc(23, "three lost early", tkX, "1|p7 b7 g7"),
		tkc(23, "one at the end", tkX, "0|p7", "~1"),
		tkc(23, "open", tkO, "0|p7"),

		tkc(24, "all four 9s", tkF, "0|p9 b9 g9", "0|y9"),
		tkc(24, "another member wins one", tkX, "0|p9 b9 g9", "1|y9"),
		tkc(24, "three are open", tkO, "0|p9 b9 g9"),
		tkc(24, "the fourth stays unplayed", tkX, "0|p9 b9 g9", "~1").lft("y9"),

		tkc(25, "three are open until the fourth is played", tkO, "0|p6 b6 g6"),
		tkc(25, "three, the fourth to another member", tkF, "0|p6 b6 g6", "1|y6"),
		tkc(25, "three at the end", tkF, "0|p6 b6 g6", "~1"),
		tkc(25, "four", tkX, "0|p6 b6 g6", "0|y6"),
		tkc(25, "two at the end", tkX, "0|p6 b6", "~1"),
		tkc(25, "two lost early", tkX, "1|p6 b6"),

		tkc(26, "two at the end", tkF, "0|p9 b9", "~1"),
		tkc(26, "two, the others to another member", tkF, "0|p9 b9", "1|g9 y9"),
		tkc(26, "two are open until the end", tkO, "0|p9 b9"),
		tkc(26, "three", tkX, "0|p9 b9 g9"),
		tkc(26, "one at the end", tkX, "0|p9", "~1"),
		tkc(26, "three lost early", tkX, "1|p9 b9 g9"),

		// 44, 45
		tkc(44, "all four colours across tricks", tkF, "0:p9 1:b1 2:g1=0", "0:y9 1:b2 2:g2=0"),
		tkc(44, "a submarine is no colour", tkX, "0:t4 1:p1 2:b1=0", "0:g9 1:g1 2:p2=0", "~1"),
		tkc(44, "a colour is missing at the end", tkX, "0:p9 1:b1 2:g1=0", "~1"),
		tkc(44, "open", tkO, "0:p9 1:b1 2:g1=0"),

		tkc(45, "all nine pink", tkF, "0:p9 1:p8 2:p7=0", "0:p6 1:p5 2:p4=0", "0:p3 1:p2 2:p1=0"),
		tkc(45, "all four submarines are no colour", tkX, "0:t3 1:t1 2:t2=0", "0:t4 1:p1 2:b1=0", "~1"),
		tkc(45, "a card of every colour lost", tkX, "1:p9 2:b9 0:g9=1", "1:y9 2:b1 0:g1=1"),
		tkc(45, "eight of nine pink at the end", tkX, "0:p9 1:p8 2:p7=0", "0:p6 1:p5 2:p4=0", "0:p3 1:p2 2:b1=0", "~1"),
		tkc(45, "open", tkO, "0:p9 1:p8 2:p7=0"),

		// 46, 47: only even / odd numbers
		tkc(46, "all even", tkF, "0|p8 b2 g4"),
		tkc(46, "4 members all even", tkF, "0|p8 b6 g4 y2").mem(4),
		tkc(46, "an odd card spoils it", tkX, "0|p8 b3 g4", "~1"),
		tkc(46, "a submarine spoils it", tkX, "0|t2 b4 g6", "~1"),
		tkc(46, "won by another member", tkX, "1|p8 b2 g4", "~1"),
		tkc(46, "open", tkO, "0|p8 b3 g4"),

		tkc(47, "all odd", tkF, "0|p9 b1 g3"),
		tkc(47, "5 members all odd", tkF, "0|p9 b7 g5 y3 p1").mem(5),
		tkc(47, "an even card spoils it", tkX, "0|p9 b2 g3", "~1"),
		tkc(47, "a submarine spoils it", tkX, "0|t3 b1 g5", "~1"),
		tkc(47, "won by another member", tkX, "1|p9 b1 g3", "~1"),
		tkc(47, "open", tkO, "0|p9 b2 g3"),

		// 48: total value at least 23 / 28 / 31 (3, 4, 5 members; 2 players: 3 members)
		tkc(48, "3 members: 23", tkF, "0|p9 b8 g6"),
		tkc(48, "3 members: 22 is not enough", tkX, "0|p9 b8 g5", "~1"),
		tkc(48, "3 members: won by another member", tkX, "1|p9 b9 g8", "~1"),
		tkc(48, "3 members: open", tkO, "0|p9 b8 g5"),
		tkc(48, "4 members: 28", tkF, "0|p9 b8 g8 y3").mem(4),
		tkc(48, "4 members: 27 is not enough", tkX, "0|p9 b8 g8 y2", "~1").mem(4),
		tkc(48, "4 members: a submarine spoils it", tkX, "0|t4 p9 b9 g9", "~1").mem(4),
		tkc(48, "5 members: 31", tkF, "0|p9 b9 g8 y4 p1").mem(5),
		tkc(48, "5 members: 30 is not enough", tkX, "0|p9 b9 g8 y3 p1", "~1").mem(5),

		// 49: total value lower than 8 / 12 / 16
		tkc(49, "3 members: 7", tkF, "0|p4 b2 g1"),
		tkc(49, "3 members: 8 is too much", tkX, "0|p5 b2 g1", "~1"),
		tkc(49, "3 members: a submarine spoils it", tkX, "0|t1 b2 g1", "~1"),
		tkc(49, "3 members: won by another member", tkX, "1|p4 b2 g1", "~1"),
		tkc(49, "3 members: open", tkO, "0|p6 b2 g1"),
		tkc(49, "4 members: 11", tkF, "0|p5 b3 g2 y1").mem(4),
		tkc(49, "4 members: 12 is too much", tkX, "0|p6 b3 g2 y1", "~1").mem(4),
		tkc(49, "5 members: 15", tkF, "0|p5 b4 g3 y2 p1").mem(5),
		tkc(49, "5 members: 16 is too much", tkX, "0|p6 b4 g3 y2 p1", "~1").mem(5),

		// 50: total value 22 or 23
		tkc(50, "22", tkF, "0|p9 b8 g5"),
		tkc(50, "23", tkF, "0|p9 b8 g6"),
		tkc(50, "21", tkX, "0|p9 b8 g4", "~1"),
		tkc(50, "24", tkX, "0|p9 b8 g7", "~1"),
		tkc(50, "a submarine spoils it", tkX, "0|t4 p9 b9", "~1"),
		tkc(50, "won by another member", tkX, "1|p9 b8 g5", "~1"),
		tkc(50, "4 members: 22", tkF, "0|p9 b5 g4 y4").mem(4),
		tkc(50, "open", tkO, "0|p9 b8 g4"),

		// 51-57: submarines
		tkc(51, "one submarine at the end", tkF, "0:t4 1:p1 2:b1=0", "~1"),
		tkc(51, "one is open until the end", tkO, "0:t4 1:p1 2:b1=0"),
		tkc(51, "two in one trick", tkX, "0:t4 1:t1 2:p1=0"),
		tkc(51, "none at the end", tkX, "~1"),
		tkc(51, "all four lost", tkX, "1|t1 t2 t3", "1|t4"),
		tkc(51, "three lost, one left", tkO, "1|t1 t2 t3"),

		tkc(52, "submarine 1 only", tkF, "0:t1 1:p1 2:b1=0", "~1"),
		tkc(52, "submarine 1 and 2", tkX, "0:t2 1:t1 2:p1=0"),
		tkc(52, "another submarine than 1", tkX, "0:t3 1:p1 2:b1=0"),
		tkc(52, "another seat wins submarine 1", tkX, "1|t1"),
		tkc(52, "no submarine at the end", tkX, "~1"),
		tkc(52, "open", tkO, "0:p9 1:b1 2:g1=0"),

		tkc(53, "submarine 2 only", tkF, "0:t2 1:p1 2:b1=0", "~1"),
		tkc(53, "submarine 2 and 3", tkX, "0:t3 1:t2 2:p1=0"),
		tkc(53, "another submarine than 2", tkX, "0:t1 1:p1 2:b1=0"),
		tkc(53, "another seat wins submarine 2", tkX, "1|t2"),
		tkc(53, "open", tkO, "0:p9 1:b1 2:g1=0"),

		tkc(55, "two submarines in one trick", tkF, "0:t4 1:t3 2:p1=0", "~1"),
		tkc(55, "two in separate tricks", tkF, "0|t4", "0|t2", "~1"),
		tkc(55, "three", tkX, "0:t4 1:t3 2:t2=0"),
		tkc(55, "one at the end", tkX, "0|t4", "~1"),
		tkc(55, "three lost early", tkX, "1|t1 t2 t3"),
		tkc(55, "open", tkO, "0|t4"),

		tkc(56, "three submarines", tkF, "0|t4 t3 t2", "~1"),
		tkc(56, "four", tkX, "0|t4 t3 t2", "0|t1"),
		tkc(56, "two at the end", tkX, "0|t4 t3", "~1"),
		tkc(56, "two lost early", tkX, "1|t1 t2"),
		tkc(56, "open", tkO, "0|t4 t3 t2"),

		tkc(57, "no submarine at the end", tkF, "0:p9 1:b1 2:g1=0", "~1"),
		tkc(57, "a submarine won", tkX, "0:t1 1:p1 2:b1=0"),
		tkc(57, "open", tkO, "0:p9 1:b1 2:g1=0"),

		// 60, 61: never open a trick with the colours shown (the owner's own leads)
		tkc(60, "leads green only", tkF, "0:g5 1:g1 2:g2=0", "~1"),
		tkc(60, "leading a submarine is allowed", tkF, "0:t1 1:p1 2:b1=0", "~1"),
		tkc(60, "following pink is no lead", tkF, "1:p5 2:p1 0:p2=1", "~1"),
		tkc(60, "leads pink", tkX, "0:p5 1:p1 2:p2=0"),
		tkc(60, "leads yellow", tkX, "0:y5 1:y1 2:y2=0"),
		tkc(60, "leads blue", tkX, "0:b5 1:b1 2:b2=0"),
		tkc(60, "open", tkO, "0:g5 1:g1 2:g2=0"),

		tkc(61, "leads yellow only", tkF, "0:y5 1:y1 2:y2=0", "~1"),
		tkc(61, "leads blue", tkF, "0:b5 1:b1 2:b2=0", "~1"),
		tkc(61, "leading a submarine is allowed", tkF, "0:t2 1:y1 2:b1=0", "~1"),
		tkc(61, "following pink is no lead", tkF, "1:p5 2:p1 0:g2=1", "~1"),
		tkc(61, "leads pink", tkX, "0:p5 1:p1 2:p2=0"),
		tkc(61, "leads green", tkX, "0:g5 1:g1 2:g2=0"),
		tkc(61, "open", tkO, "0:y5 1:y1 2:y2=0"),

		// 71-73: don't win any of the first four / three / five tricks
		tkc(71, "four lost", tkF, "~1:4"),
		tkc(71, "wins from the fifth on", tkF, "~1:4", "~0:9"),
		tkc(71, "wins the third", tkX, "~1:2", "~0:1"),
		tkc(71, "wins the fourth", tkX, "~1:3", "~0:1"),
		tkc(71, "wins the first", tkX, "~0:1"),
		tkc(71, "open before the fourth", tkO, "~1:3"),

		tkc(72, "three lost", tkF, "~1:3"),
		tkc(72, "wins the fourth", tkF, "~1:3", "~0:1"),
		tkc(72, "wins the third", tkX, "~1:2", "~0:1"),
		tkc(72, "open", tkO, "~1:2"),

		tkc(73, "five lost", tkF, "~1:5"),
		tkc(73, "wins from the sixth on", tkF, "~1:5", "~0:8"),
		tkc(73, "wins the fifth", tkX, "~1:4", "~0:1"),
		tkc(73, "open", tkO, "~1:4"),

		// 74: don't win any tricks
		tkc(74, "no trick at the end", tkF, "~1:13"),
		tkc(74, "no trick, shared", tkF, "~1:6", "~2:7"),
		tkc(74, "wins the fourth", tkX, "~1:3", "~0:1"),
		tkc(74, "wins the last", tkX, "~1:12", "~0:1"),
		tkc(74, "open", tkO, "~1:5"),

		// 75: don't win two tricks in a row
		tkc(75, "alternating to the end", tkF, tkWins("0 1 0 1 0 1 0 1 0 1 0 1 0")...),
		tkc(75, "others win in a row", tkF, tkWins("1 1 1 0 2 2 0 1 1 0 2 1 0")...),
		tkc(75, "two in a row at the start", tkX, tkWins("0 0")...),
		tkc(75, "two in a row later", tkX, tkWins("1 0 1 0 0")...),
		tkc(75, "two in a row at the end", tkX, tkWins("1 0 1 0 1 0 1 0 1 0 1 0 0")...),
		tkc(75, "open", tkO, tkWins("0 1 0")...),

		// 76: win the last trick
		tkc(76, "wins the last", tkF, "~1:12", "~0:1"),
		tkc(76, "4 members: the last is trick 10", tkF, "~1:9", "~0:1").mem(4),
		tkc(76, "another seat wins the last", tkX, "~0:12", "~1:1"),
		tkc(76, "open before the last", tkO, "~0:12"),

		// 77, 78, 79: win the first three / two / one tricks
		tkc(77, "three won", tkF, "~0:3"),
		tkc(77, "three won, more lost", tkF, "~0:3", "~1:10"),
		tkc(77, "third lost", tkX, "~0:2", "~1:1"),
		tkc(77, "first lost", tkX, "~1:1"),
		tkc(77, "open", tkO, "~0:2"),

		tkc(78, "two won", tkF, "~0:2"),
		tkc(78, "second lost", tkX, "~0:1", "~1:1"),
		tkc(78, "open", tkO, "~0:1"),

		tkc(79, "first won", tkF, "~0:1"),
		tkc(79, "first lost", tkX, "~1:1"),
		tkc(79, "open", tkO),

		// 80: win the first and the last trick
		tkc(80, "first and last", tkF, "~0:1", "~1:11", "~0:1"),
		tkc(80, "4 members", tkF, "~0:1", "~1:8", "~0:1").mem(4),
		tkc(80, "first won, last pending", tkO, "~0:1", "~1:5"),
		tkc(80, "first lost", tkX, "~1:1", "~0:12"),
		tkc(80, "last lost", tkX, "~0:1", "~0:11", "~1:1"),

		// 81: win only the last trick
		tkc(81, "only the last", tkF, "~1:12", "~0:1"),
		tkc(81, "4 members", tkF, "~1:9", "~0:1").mem(4),
		tkc(81, "wins another trick", tkX, "~0:1", "~1:11", "~0:1"),
		tkc(81, "does not win the last", tkX, "~1:13"),
		tkc(81, "open", tkO, "~1:5"),

		// 82: win only the first trick
		tkc(82, "only the first", tkF, "~0:1", "~1:12"),
		tkc(82, "wins another trick", tkX, "~0:1", "~1:5", "~0:1"),
		tkc(82, "first lost", tkX, "~1:1"),
		tkc(82, "open", tkO, "~0:1", "~1:3"),

		// 83, 84, 87: win exactly one / two / four tricks
		tkc(83, "one at the end", tkF, "~0:1", "~1:12"),
		tkc(83, "one is open", tkO, "~0:1"),
		tkc(83, "two", tkX, "~0:2"),
		tkc(83, "none at the end", tkX, "~1:13"),
		tkc(83, "12 lost, one left", tkO, "~1:12"),

		tkc(84, "two at the end", tkF, "~0:2", "~1:11"),
		tkc(84, "two are open", tkO, "~0:2"),
		tkc(84, "three", tkX, "~0:3"),
		tkc(84, "one at the end", tkX, "~0:1", "~1:12"),

		tkc(87, "four at the end", tkF, "~0:4", "~1:9"),
		tkc(87, "five", tkX, "~0:5"),
		tkc(87, "three at the end", tkX, "~0:3", "~1:10"),
		tkc(87, "4 members: too few tricks left", tkX, "~1:7").mem(4),
		tkc(87, "4 members: four at the end", tkF, "~0:4", "~1:6").mem(4),
		tkc(87, "open", tkO, "~0:4"),

		// 85, 86: win two / three tricks in a row
		tkc(85, "two in a row", tkF, "~0:2"),
		tkc(85, "more than two in a row", tkF, "~0:5"),
		tkc(85, "two in a row after a gap", tkF, tkWins("0 1 0 0")...),
		tkc(85, "never in a row, one trick left", tkX, "~1:12"),
		tkc(85, "never in a row at the end", tkX, tkWins("0 1 0 1 0 1 0 1 0 1 0 1 0")...),
		tkc(85, "two tricks left", tkO, "~1:11"),
		tkc(85, "open", tkO, "~0:1"),

		tkc(86, "three in a row", tkF, "~0:3"),
		tkc(86, "three in a row after a loss", tkF, tkWins("1 0 0 0")...),
		tkc(86, "more than three", tkF, "~0:5"),
		tkc(86, "pairs only", tkX, tkWins("0 0 1 0 0 1 0 0 1 0 0 1 0")...),
		tkc(86, "two tricks left", tkX, "~1:11"),
		tkc(86, "three tricks left", tkO, "~1:10"),
		tkc(86, "open", tkO, "~0:2"),

		// 88, 89: exactly three / two tricks in total, all in a row
		tkc(88, "three in a row", tkF, "~1:4", "~0:3", "~1:6"),
		tkc(88, "three in a row are open until the end", tkO, "~0:3", "~1:2"),
		tkc(88, "three, not in a row", tkX, "~0:1", "~1:1", "~0:2"),
		tkc(88, "four in a row", tkX, "~0:4"),
		tkc(88, "two at the end", tkX, "~0:2", "~1:11"),

		tkc(89, "two in a row", tkF, "~1:5", "~0:2", "~1:6"),
		tkc(89, "two are open until the end", tkO, "~0:2"),
		tkc(89, "three in a row", tkX, "~0:3"),
		tkc(89, "two, not in a row", tkX, "~0:1", "~1:1", "~0:1"),
		tkc(89, "one at the end", tkX, "~0:1", "~1:12"),

		// 90, 91: predict the number of tricks (open and hidden behave the same)
		tkc(90, "exactly as predicted", tkF, "~0:3", "~1:10").cnt(3),
		tkc(90, "predicted none", tkF, "~1:13").cnt(0),
		tkc(90, "one trick too many", tkX, "~0:4").cnt(3),
		tkc(90, "one trick too few", tkX, "~0:2", "~1:11").cnt(3),
		tkc(90, "won one although none predicted", tkX, "~0:1").cnt(0),
		tkc(90, "too few tricks left", tkX, "~1:6").mem(4).cnt(5),
		tkc(90, "open", tkO, "~0:3").cnt(3),

		tkc(91, "exactly as predicted", tkF, "~0:5", "~1:8").cnt(5),
		tkc(91, "predicted none", tkF, "~1:13").cnt(0),
		tkc(91, "one trick too many", tkX, "~0:6").cnt(5),
		tkc(91, "one trick too few", tkX, "~0:4", "~1:9").cnt(5),
		tkc(91, "4 members as predicted", tkF, "~0:2", "~1:8").mem(4).cnt(2),
		tkc(91, "open", tkO, "~0:2").cnt(5),

		// 92: the same number of pink and yellow cards overall (more than 0)
		tkc(92, "one each", tkF, "0:p9 1:y9 2:b1=0", "~1"),
		tkc(92, "one each across tricks", tkF, "0:p9 1:b1 2:g1=0", "0:y9 1:b2 2:g2=0", "~1"),
		tkc(92, "none of either", tkX, "0:b9 1:b8 2:g1=0", "~1"),
		tkc(92, "two pink, one yellow", tkX, "0:p9 1:p8 2:y1=0", "~1"),
		tkc(92, "unequal for now", tkO, "0:p9 1:p8 2:y1=0"),
		tkc(92, "no trick for the owner", tkX, "~1"),

		// 93, 94: the same number of two colours in one trick (more than 0)
		tkc(93, "one green and one yellow", tkF, "0:g9 1:y1 2:b1=0"),
		tkc(93, "4 members: two each", tkF, "0:g9 1:g8 2:y1 3:y2=0").mem(4),
		tkc(93, "in different tricks", tkX, "0:g9 1:b1 2:p1=0", "0:y9 1:b2 2:p2=0", "~1"),
		tkc(93, "unequal in the trick", tkX, "0:g9 1:g8 2:y1=0", "~1"),
		tkc(93, "none of either", tkX, "0:b9 1:b8 2:p1=0", "~1"),
		tkc(93, "won by another member", tkX, "1:g9 2:y1 0:b1=1", "~1"),
		tkc(93, "open", tkO, "0:g9 1:b1 2:p1=0"),

		tkc(94, "one pink and one blue", tkF, "0:p9 1:b1 2:g1=0"),
		tkc(94, "4 members: two each", tkF, "0:p9 1:p8 2:b1 3:b2=0").mem(4),
		tkc(94, "in different tricks", tkX, "0:p9 1:g1 2:y1=0", "0:b9 1:g2 2:y2=0", "~1"),
		tkc(94, "unequal in the trick", tkX, "0:p9 1:p8 2:b1=0", "~1"),
		tkc(94, "none of either", tkX, "0:g9 1:g8 2:y1=0", "~1"),
		tkc(94, "won by another member", tkX, "1:p9 2:b1 0:g1=1", "~1"),
		tkc(94, "open", tkO, "0:p9 1:g1 2:y1=0"),

		// 95, 96: more of one colour than of another (zero of the other allowed)
		tkc(95, "yellow and no blue", tkF, "0:y9 1:g1 2:p1=0", "~1"),
		tkc(95, "equal fails", tkX, "0:y9 1:b1 2:g1=0", "~1"),
		tkc(95, "more blue", tkX, "0:b9 1:g1 2:p1=0", "~1"),
		tkc(95, "none of either", tkX, "~1"),
		tkc(95, "open", tkO, "0:y9 1:g1 2:p1=0"),

		tkc(96, "pink and no green", tkF, "0:p9 1:b1 2:y1=0", "~1"),
		tkc(96, "equal fails", tkX, "0:p9 1:g1 2:b1=0", "~1"),
		tkc(96, "more green", tkX, "0:g9 1:b1 2:y1=0", "~1"),
		tkc(96, "none of either", tkX, "~1"),
		tkc(96, "open", tkO, "0:p9 1:b1 2:y1=0"),
	}
	return append(cs, tkWinCardsCases()...)
}

// Rules §6, §9: every task card is covered by cases of each outcome.
func TestRulesDeepSeaTaskCards(t *testing.T) {
	cases := tkTaskCases()
	tkRun(t, cases)
	t.Logf("%d task card cases", len(cases))
}

func TestRulesDeepSeaTaskCardCoverage(t *testing.T) {
	seen := map[int]map[Status]bool{}
	for _, tc := range tkTaskCases() {
		if seen[tc.id] == nil {
			seen[tc.id] = map[Status]bool{}
		}
		seen[tc.id][tc.want] = true
	}
	for id := 1; id <= TaskCards; id++ {
		if !seen[id][tkF] || !seen[id][tkX] {
			t.Errorf("card %d has no fulfilled and failed case", id)
		}
	}
}

// Rules §3, §9: the captain comparison tasks (4-6) are never the captain's.
func TestRulesDeepSeaCaptainComparisonNotForCaptain(t *testing.T) {
	for id := 1; id <= TaskCards; id++ {
		card, _ := TaskCardFor(id)
		if got, want := card.NotForCaptain(), id >= 4 && id <= 6; got != want {
			t.Errorf("card %d NotForCaptain = %v, want %v", id, got, want)
		}
	}
}

// Rules §9: cards 48 and 49 depend on the crew size; 2 players (with the dummy) use 3 members.
func TestRulesDeepSeaTrickSumByCrewSize(t *testing.T) {
	for _, tc := range []struct {
		id, members, count int
	}{{48, 3, 23}, {48, 4, 28}, {48, 5, 31}, {49, 3, 8}, {49, 4, 12}, {49, 5, 16}} {
		card, _ := TaskCardFor(tc.id)
		if got := card.objectiveFor(tc.members).Count; got != tc.count {
			t.Errorf("card %d with %d members: %d, want %d", tc.id, tc.members, got, tc.count)
		}
	}
}
