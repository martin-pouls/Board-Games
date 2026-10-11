package domain

import (
	"fmt"
	"testing"
)

const (
	bbH0 = "p2 p5 p9 b4 g1 g3 t1 t2"
	bbH1 = "p1 b1 b2 g2 y1"
	bbH2 = "p3 b3 g4 y2 y3"
	bbH3 = "t4 p4 b5 g5 y4" // the commander (seat 3) leads
)

// bbCommGame is a 4-player game of the mission at the lead of trick 1 (the task y9 of seat 0 is
// never played).
func bbCommGame(t *testing.T, mission int) *Game {
	t.Helper()
	g := startMission(t, 4, mission)
	g.Attempt.Objectives = nil
	setHands(g, bbH0, bbH1, bbH2, bbH3)
	setTasks(g, "y9:0")
	g.PullEvents()
	return g
}

var errCommNotAllowed = []error{ErrCannotCommunicate, ErrWrongStep, ErrNotAllowed}

// rules §8: which statement is true for a card.
func TestRulesStatementFor(t *testing.T) {
	hand := cs(bbH0)
	for _, tc := range []struct {
		card string
		pos  Position
		ok   bool
	}{
		{"p2", PosLowest, true},
		{"p5", "", false}, // in the middle
		{"p9", PosHighest, true},
		{"b4", PosOnly, true},
		{"g1", PosLowest, true},
		{"g3", PosHighest, true},
		{"t1", "", false}, // trump
		{"t2", "", false},
		{"y5", "", false}, // not in the hand
	} {
		pos, ok := StatementFor(hand, c(tc.card))
		if ok != tc.ok || (ok && pos != tc.pos) {
			t.Errorf("StatementFor(%s) = %q, %v; want %q, %v", tc.card, pos, ok, tc.pos, tc.ok)
		}
	}
	// two cards of a colour: both ends; a singleton is "only"
	for _, tc := range []struct {
		hand, card string
		pos        Position
		ok         bool
	}{
		{"p3 p4", "p3", PosLowest, true},
		{"p3 p4", "p4", PosHighest, true},
		{"p3", "p3", PosOnly, true},
		{"p1 p2 p3", "p2", "", false},
		{"y9 t3", "y9", PosOnly, true},
		{"t1 t2 t3", "t2", "", false},
	} {
		pos, ok := StatementFor(cs(tc.hand), c(tc.card))
		if ok != tc.ok || (ok && pos != tc.pos) {
			t.Errorf("StatementFor(%q, %s) = %q, %v; want %q, %v", tc.hand, tc.card, pos, ok, tc.pos, tc.ok)
		}
	}
}

// rules §8: what may be communicated and how.
func TestRulesCommunicateEligibility(t *testing.T) {
	for _, tc := range []struct {
		name string
		card string
		pos  Position
		err  []error // nil = accepted
	}{
		{"highest", "p9", PosHighest, nil},
		{"lowest", "p2", PosLowest, nil},
		{"only", "b4", PosOnly, nil},
		{"lowest of two", "g1", PosLowest, nil},
		{"highest of two", "g3", PosHighest, nil},
		{"middle card as highest", "p5", PosHighest, []error{ErrInvalidPosition, ErrCannotCommunicate}},
		{"middle card as lowest", "p5", PosLowest, []error{ErrInvalidPosition, ErrCannotCommunicate}},
		{"middle card as only", "p5", PosOnly, []error{ErrInvalidPosition, ErrCannotCommunicate}},
		{"lowest stated as highest", "p2", PosHighest, []error{ErrInvalidPosition}},
		{"highest stated as lowest", "p9", PosLowest, []error{ErrInvalidPosition}},
		{"highest stated as only", "p9", PosOnly, []error{ErrInvalidPosition}},
		{"a singleton as highest", "b4", PosHighest, []error{ErrInvalidPosition}},
		{"a singleton as lowest", "b4", PosLowest, []error{ErrInvalidPosition}},
		{"trump as only", "t1", PosOnly, []error{ErrInvalidPosition, ErrCannotCommunicate}},
		{"trump as highest", "t2", PosHighest, []error{ErrInvalidPosition, ErrCannotCommunicate}},
		{"no position in a normal mission", "p9", PosHidden, []error{ErrInvalidPosition}},
		{"a card of somebody else", "p1", PosOnly, []error{ErrCardNotInHand}},
		{"a card nobody has", "y9", PosOnly, []error{ErrCardNotInHand}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := bbCommGame(t, 1)
			if tc.err != nil {
				bbFailAny(t, g, func(g *Game) error { return g.Communicate(alice, c(tc.card), tc.pos) }, tc.err...)
				return
			}
			must(t, g.Communicate(alice, c(tc.card), tc.pos))
			ev := eventsOf[Communicated](g)
			if len(ev) != 1 || ev[0].Seat != 0 || ev[0].Card != c(tc.card) || ev[0].Position != tc.pos {
				t.Fatalf("events %v", ev)
			}
			cm := g.Attempt.CommunicatedCard(0)
			if cm == nil || cm.Card != c(tc.card) || cm.Position != tc.pos {
				t.Fatalf("communication %+v", cm)
			}
			if !bbHas(g.Attempt.Members[0].Hand, tc.card) {
				t.Fatal("the card must stay in the hand")
			}
		})
	}
}

// rules §8: once per attempt, per human.
func TestRulesCommunicateOncePerAttempt(t *testing.T) {
	g := bbCommGame(t, 1)
	must(t, g.Communicate(alice, c("p9"), PosHighest))
	mustFail(t, g, ErrAlreadyCommunicated, func(g *Game) error { return g.Communicate(alice, c("p2"), PosLowest) })
	mustFail(t, g, ErrAlreadyCommunicated, func(g *Game) error { return g.Communicate(alice, c("p9"), PosHighest) })
	if err := g.CanCommunicate(0); err == nil {
		t.Fatal("alice has used her communication")
	}
	// the others keep theirs
	if err := g.CanCommunicate(1); err != nil {
		t.Fatalf("bob: %v", err)
	}
	must(t, g.Communicate(bob, c("b1"), PosLowest))
	must(t, g.Communicate(carol, c("y3"), PosHighest))
	must(t, g.Communicate(dave, c("y4"), PosOnly))
	for seat := range 4 {
		if g.CanCommunicate(seat) == nil {
			t.Fatalf("seat %d can still communicate", seat)
		}
	}
	if len(eventsOf[Communicated](g)) != 4 {
		t.Fatal("events")
	}
}

// rules §8: the communicated card is public; it disappears from the display once it is played, but
// the communication stays used.
func TestRulesCommunicationDisplay(t *testing.T) {
	g := bbCommGame(t, 1)
	must(t, g.Communicate(alice, c("p2"), PosLowest))
	for _, id := range crewIDs[:4] {
		v, err := g.ViewFor(id)
		must(t, err)
		m := v.Attempt.Members[0]
		if m.Communication == nil || m.Communication.Card != c("p2") || m.Communication.Position != PosLowest || !m.Communicated {
			t.Fatalf("%s sees %+v", id, m)
		}
	}
	// dave (commander) leads p4, alice follows with her communicated p2 (must follow pink)
	must(t, g.PlayCard(dave, c("p4"), noShuffle{}))
	mustFail(t, g, ErrMustFollowSuit, func(g *Game) error { return g.PlayCard(alice, c("b4"), noShuffle{}) })
	must(t, g.PlayCard(alice, c("p2"), noShuffle{}))
	v, err := g.ViewFor(bob)
	must(t, err)
	if m := v.Attempt.Members[0]; m.Communication != nil || !m.Communicated {
		t.Fatalf("after the card is played: %+v", m)
	}
	mustFail(t, g, ErrAlreadyCommunicated, func(g *Game) error { return g.Communicate(alice, c("p9"), PosHighest) })
}

// rules §8: only before a trick; once the leader plays, the window for this trick is closed.
func TestRulesCommunicateTiming(t *testing.T) {
	g := bbCommGame(t, 1)
	must(t, g.Communicate(bob, c("b1"), PosLowest)) // not the leader: fine before the lead
	must(t, g.PlayCard(dave, c("p4"), noShuffle{}))
	for seat := range 4 {
		if err := g.CanCommunicate(seat); err == nil {
			t.Errorf("seat %d may communicate during a trick", seat)
		}
	}
	bbFailAny(t, g, func(g *Game) error { return g.Communicate(alice, c("p9"), PosHighest) }, errCommNotAllowed...)
	bbFailAny(t, g, func(g *Game) error { return g.Communicate(carol, c("y3"), PosHighest) }, errCommNotAllowed...)
	must(t, g.PlayCard(alice, c("p5"), noShuffle{}))
	bbFailAny(t, g, func(g *Game) error { return g.Communicate(carol, c("y3"), PosHighest) }, errCommNotAllowed...)
	must(t, g.PlayCard(bob, c("p1"), noShuffle{}))
	must(t, g.PlayCard(carol, c("p3"), noShuffle{}))
	// trick 1 is over (alice won with p5): the window opens again
	if g.Attempt.Tricks[0].Winner != 0 {
		t.Fatalf("winner %d", g.Attempt.Tricks[0].Winner)
	}
	must(t, g.Communicate(carol, c("y3"), PosHighest))
	must(t, g.Communicate(dave, c("y4"), PosOnly))
	// once alice (the new leader) has played, it is closed again
	must(t, g.PlayCard(alice, c("p2"), noShuffle{}))
	bbFailAny(t, g, func(g *Game) error { return g.Communicate(alice, c("g1"), PosLowest) }, errCommNotAllowed...)
}

func TestRulesCommunicateNotInOtherSteps(t *testing.T) {
	// during the draft
	g := startMission(t, 4, 1)
	for seat := range 4 {
		if g.CanCommunicate(seat) == nil {
			t.Fatalf("seat %d may communicate during the task draft", seat)
		}
	}
	card := g.Attempt.Members[0].Hand[0]
	pos, ok := StatementFor(g.Attempt.Members[0].Hand, card)
	if ok {
		bbFailAny(t, g, func(g *Game) error { return g.Communicate(alice, card, pos) }, errCommNotAllowed...)
	}
	// after the attempt
	g = bbCommGame(t, 1)
	must(t, g.GiveUp(alice))
	bbFailAny(t, g, func(g *Game) error { return g.Communicate(alice, c("p9"), PosHighest) }, ErrWrongPhase, ErrCannotCommunicate)
	// unknown player
	g = bbCommGame(t, 1)
	bbFailAny(t, g, func(g *Game) error { return g.Communicate("nobody", c("p9"), PosHighest) }, ErrUnknownPlayer)
}

// rules §8: JARVIS never communicates.
func TestRulesDummyNeverCommunicates(t *testing.T) {
	g := bbDummyGame(t, "t4 p5 p9", "p6 b6", bbUp, bbDown)
	if g.CanCommunicate(2) == nil {
		t.Fatal("JARVIS must not communicate")
	}
	if g.CanCommunicate(0) != nil || g.CanCommunicate(1) != nil {
		t.Fatal("the humans may communicate")
	}
	must(t, g.Communicate(alice, c("p9"), PosHighest))
	v, err := g.ViewFor(bob)
	must(t, err)
	if len(v.Attempt.Actions.Communicate) == 0 {
		t.Fatal("bob should be offered cards")
	}
}

// rules §8: the dead zone — the card must qualify, the position is not stated or shown.
func TestRulesDeadZone(t *testing.T) {
	for _, m := range []int{6, 14, 21, 25, 29, 39} {
		t.Run(fmt.Sprintf("mission %d", m), func(t *testing.T) {
			if !MissionDefOf(m).Comm.DeadZone {
				t.Skip("not a dead zone mission in the table")
			}
			g := bbCommGame(t, m)
			// the card must still qualify
			bbFailAny(t, g, func(g *Game) error { return g.Communicate(alice, c("p5"), PosHidden) }, ErrInvalidPosition, ErrCannotCommunicate)
			bbFailAny(t, g, func(g *Game) error { return g.Communicate(alice, c("t1"), PosHidden) }, ErrInvalidPosition, ErrCannotCommunicate)
			bbFailAny(t, g, func(g *Game) error { return g.Communicate(alice, c("p1"), PosHidden) }, ErrCardNotInHand)
			// the offered cards carry no position
			v, err := g.ViewFor(alice)
			must(t, err)
			offers := v.Attempt.Actions.Communicate
			if len(offers) != 5 {
				t.Fatalf("offers %v, want p2 p9 b4 g1 g3", offers)
			}
			for _, o := range offers {
				if o.Position != PosHidden {
					t.Fatalf("offer %+v shows a position", o)
				}
			}
			must(t, g.Communicate(alice, c("p9"), PosHidden))
			ev := eventsOf[Communicated](g)
			if len(ev) != 1 || ev[0].Position != PosHidden || ev[0].Card != c("p9") {
				t.Fatalf("events %v", ev)
			}
			for _, id := range crewIDs[:4] {
				v, _ := g.ViewFor(id)
				cm := v.Attempt.Members[0].Communication
				if cm == nil || cm.Card != c("p9") || cm.Position != PosHidden {
					t.Fatalf("%s sees %+v", id, cm)
				}
			}
		})
	}
}

// MissionDefOf is a test shortcut for the planet nine mission table.
func MissionDefOf(n int) MissionDef {
	m, _ := MissionFor(EditionPlanetNine, n)
	return m
}

// rules §8: disruption — nobody communicates before trick N, directly before trick N is fine.
func TestRulesDisruption(t *testing.T) {
	for _, tc := range []struct{ mission, from int }{{18, 2}, {19, 3}, {28, 3}, {30, 2}, {38, 3}, {1, 0}, {6, 0}} {
		t.Run(fmt.Sprintf("mission %d from trick %d", tc.mission, tc.from), func(t *testing.T) {
			if got := MissionDefOf(tc.mission).Comm.FromTrick; got != tc.from {
				t.Fatalf("table says disruption %d", got)
			}
			for trickNo := 1; trickNo <= 4; trickNo++ {
				g := bbCommGame(t, tc.mission)
				if trickNo > 1 {
					bbHistory(g, bbFillers(4, make([]int, trickNo-1)...)...)
					g.Attempt.Commander = 0
				}
				want := trickNo >= tc.from
				for seat := range 4 {
					if got := g.CanCommunicate(seat) == nil; got != want {
						t.Fatalf("before trick %d: seat %d may communicate = %v, want %v", trickNo, seat, got, want)
					}
				}
				if want {
					must(t, g.Communicate(alice, c("p9"), g.stmt(0, "p9")))
				} else {
					bbFailAny(t, g, func(g *Game) error { return g.Communicate(alice, c("p9"), PosHighest) }, errCommNotAllowed...)
					v, _ := g.ViewFor(alice)
					if len(v.Attempt.Actions.Communicate) != 0 {
						t.Fatalf("offers during the disruption: %v", v.Attempt.Actions.Communicate)
					}
				}
			}
		})
	}
}

// stmt is the true statement for the card in the seat's hand (test shortcut).
func (g *Game) stmt(seat int, card string) Position {
	p, _ := StatementFor(g.Attempt.Members[seat].Hand, c(card))
	return p
}

// rules §8: the offered cards of a view are exactly the communicable ones with their statement.
func TestRulesCommunicateOffers(t *testing.T) {
	g := bbCommGame(t, 1)
	v, err := g.ViewFor(alice)
	must(t, err)
	got := map[Card]Position{}
	for _, o := range v.Attempt.Actions.Communicate {
		got[o.Card] = o.Position
	}
	want := map[Card]Position{c("p2"): PosLowest, c("p9"): PosHighest, c("b4"): PosOnly, c("g1"): PosLowest, c("g3"): PosHighest}
	if len(got) != len(want) {
		t.Fatalf("offers %v", got)
	}
	for k, p := range want {
		if got[k] != p {
			t.Fatalf("offers %v, want %v", got, want)
		}
	}
	must(t, g.Communicate(alice, c("p9"), PosHighest))
	v, _ = g.ViewFor(alice)
	if len(v.Attempt.Actions.Communicate) != 0 {
		t.Fatal("no offers after communicating")
	}
}
