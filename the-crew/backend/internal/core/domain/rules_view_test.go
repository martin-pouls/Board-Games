package domain

import (
	"reflect"
	"slices"
	"testing"
)

// bbCardsIn collects every Card value reachable from v (structs, slices, pointers, maps).
func bbCardsIn(v reflect.Value, out *[]Card) {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			bbCardsIn(v.Elem(), out)
		}
	case reflect.Struct:
		if v.Type() == cardType {
			*out = append(*out, v.Interface().(Card))
			return
		}
		for i := range v.NumField() {
			if v.Type().Field(i).IsExported() {
				bbCardsIn(v.Field(i), out)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			bbCardsIn(v.Index(i), out)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			bbCardsIn(v.MapIndex(k), out)
		}
	}
}

// bbViewCards returns the distinct cards appearing anywhere in the viewer's view.
func bbViewCards(t *testing.T, g *Game, id PlayerID) []Card {
	t.Helper()
	v, err := g.ViewFor(id)
	must(t, err)
	var all []Card
	bbCardsIn(reflect.ValueOf(v), &all)
	SortCards(all)
	return slices.Compact(all)
}

// bbAssertOnly fails if the view of id shows a card outside allowed.
func bbAssertOnly(t *testing.T, g *Game, id PlayerID, allowed []Card, what string) {
	t.Helper()
	for _, card := range bbViewCards(t, g, id) {
		if !slices.Contains(allowed, card) {
			t.Fatalf("%s: the view of %s shows %v", what, id, card)
		}
	}
}

// rules: a human's hand is never shown to another human while the attempt runs.
func TestRulesViewHidesOtherHands(t *testing.T) {
	g := bbCommGame(t, 1)
	hands := [][]Card{cs(bbH0), cs(bbH1), cs(bbH2), cs(bbH3)}
	for s := range 4 {
		allowed := slices.Concat(hands[s], cs("y9")) // y9 is the public task card
		bbAssertOnly(t, g, bbID(g, s), allowed, "fresh deal")
		v, _ := g.ViewFor(bbID(g, s))
		if !bbCardsEqual(v.Attempt.Hand, hands[s]) {
			t.Fatalf("seat %d: hand %v", s, v.Attempt.Hand)
		}
		for o := range 4 {
			m := v.Attempt.Members[o]
			if len(m.Cards) != 0 {
				t.Fatalf("seat %d sees the cards of %d while playing", s, o)
			}
			if m.CardCount != len(hands[o]) {
				t.Fatalf("card count of %d: %d", o, m.CardCount)
			}
		}
	}
	// legal cards belong to the one who plays
	v, _ := g.ViewFor(dave)
	if !bbCardsEqual(v.Attempt.Playable, cs(bbH3)) {
		t.Fatalf("the leader may play every card: %v", v.Attempt.Playable)
	}
	v, _ = g.ViewFor(alice)
	if len(v.Attempt.Playable) != 0 || v.Attempt.Turn != 3 {
		t.Fatalf("alice: playable %v turn %d", v.Attempt.Playable, v.Attempt.Turn)
	}
	// a card in the trick is public from then on
	must(t, g.PlayCard(dave, c("p4"), noShuffle{}))
	bbAssertOnly(t, g, bob, slices.Concat(cs(bbH1), cs("y9 p4")), "after the lead")
	// a communicated card is public
	must(t, g.PlayCard(alice, c("p5"), noShuffle{}))
	must(t, g.PlayCard(bob, c("p1"), noShuffle{}))
	must(t, g.PlayCard(carol, c("p3"), noShuffle{}))
	must(t, g.Communicate(dave, c("b5"), PosOnly))
	bbAssertOnly(t, g, bob, slices.Concat(cs("g1 g2 b1 b2 y1"), cs("y9 b5 p4 p5 p1 p3")), "communicated + last trick")
}

// rules §12: JARVIS's hidden cards are never shown; its face-up cards are public.
func TestRulesViewDummyHiddenCards(t *testing.T) {
	g := bbDummyGame(t, "t4 p5 b5", "p6 b6", bbUp, bbDown)
	for _, id := range []PlayerID{alice, bob} {
		hand := g.Attempt.Members[0].Hand
		if id == bob {
			hand = g.Attempt.Members[1].Hand
		}
		allowed := slices.Concat(hand, cs(bbUp), cs("y9"))
		bbAssertOnly(t, g, id, allowed, "dummy")
		v, _ := g.ViewFor(id)
		d := v.Attempt.Members[2]
		if !d.Dummy || len(d.Stacks) != 7 || d.CardCount != 14 || len(d.Cards) != 0 {
			t.Fatalf("dummy view %+v", d)
		}
		for k, s := range d.Stacks {
			if s.Up == nil || !s.Hidden {
				t.Fatalf("stack %d: %+v", k, s)
			}
		}
	}
	// the dummy's legal cards are shown to the one who plays for it (the commander)
	must(t, g.PlayCard(alice, c("p5"), noShuffle{}))
	must(t, g.PlayCard(bob, c("p6"), noShuffle{}))
	va, _ := g.ViewFor(alice)
	vb, _ := g.ViewFor(bob)
	if !bbCardsEqual(va.Attempt.Playable, cs("p1 p2")) || va.Attempt.Turn != 2 {
		t.Fatalf("commander: playable %v turn %d", va.Attempt.Playable, va.Attempt.Turn)
	}
	if len(vb.Attempt.Playable) != 0 {
		t.Fatalf("bob: playable %v", vb.Attempt.Playable)
	}
}

// rules §9: other players' pass choices are secret; each sees their own side of the exchange.
func TestRulesViewPassSecrecy(t *testing.T) {
	g := bbReady(t, 4, 1)
	must(t, g.SendDistress(bbCmd(g), DirLeft))
	hands := make([][]Card, 4)
	chosen := make([]Card, 4)
	for s := range 4 {
		hands[s] = slices.Clone(g.Attempt.Members[s].Hand)
		chosen[s] = bbNonTrump(g, s)
	}
	pub := publicTaskCards(g)
	must(t, g.ChoosePassCard(bbID(g, 0), chosen[0]))
	must(t, g.ChoosePassCard(bbID(g, 1), chosen[1]))
	for s := range 4 {
		bbAssertOnly(t, g, bbID(g, s), slices.Concat(hands[s], pub), "during the pass")
	}
	v, _ := g.ViewFor(bbID(g, 2))
	if v.Attempt.Pass == nil || !slices.Equal(v.Attempt.Pass.Waiting, []int{2, 3}) {
		t.Fatalf("waiting %+v", v.Attempt.Pass)
	}
	must(t, g.ChoosePassCard(bbID(g, 2), chosen[2]))
	must(t, g.ChoosePassCard(bbID(g, 3), chosen[3]))
	for s := range 4 {
		from := (s + 3) % 4
		allowed := slices.Concat(hands[s], []Card{chosen[from]}, pub)
		bbAssertOnly(t, g, bbID(g, s), allowed, "after the pass")
	}
}

func publicTaskCards(g *Game) []Card {
	var out []Card
	for _, tk := range g.Attempt.Tasks {
		if !tk.FaceDown {
			out = append(out, tk.Card())
		}
	}
	return out
}

// rules §10: face-down tasks stay hidden until revealed.
func TestRulesViewFaceDownTasks(t *testing.T) {
	g := startMission(t, 4, 20)
	for s := range 4 {
		bbAssertOnly(t, g, bbID(g, s), g.Attempt.Members[s].Hand, "face-down tasks")
	}
	g = startMission(t, 4, 24) // only the first task is revealed
	for s := range 4 {
		bbAssertOnly(t, g, bbID(g, s), slices.Concat(g.Attempt.Members[s].Hand, []Card{g.Attempt.Tasks[0].Card()}), "distribution")
	}
}

// rules §5: the last completed trick stays visible until the next trick's first card.
func TestRulesViewLastTrick(t *testing.T) {
	g := bbTrickGame(t, 0, "p3 b1 b2", "p7 g1 g2", "p5 y1 y2", "p6 y3 y4")
	view := func(id PlayerID) *AttemptView {
		v, err := g.ViewFor(id)
		must(t, err)
		return v.Attempt
	}
	for _, id := range crewIDs[:4] {
		if view(id).LastTrick != nil {
			t.Fatal("no trick played yet")
		}
	}
	for i, f := range []string{"p3", "p7", "p5", "p6"} {
		must(t, g.PlayCard(g.Players[i].ID, c(f), noShuffle{}))
		if i < 3 {
			if view(bob).LastTrick != nil || len(view(bob).Trick) != i+1 {
				t.Fatalf("during trick 1: last %v trick %v", view(bob).LastTrick, view(bob).Trick)
			}
		}
	}
	for _, id := range crewIDs[:4] {
		av := view(id)
		lt := av.LastTrick
		if lt == nil || lt.Winner != 1 || lt.Leader != 0 || len(lt.Plays) != 4 || lt.Plays[1].Card != c("p7") {
			t.Fatalf("%s: last trick %+v", id, lt)
		}
		if len(av.Trick) != 0 || av.TrickNumber != 2 {
			t.Fatalf("trick %v number %d", av.Trick, av.TrickNumber)
		}
		if av.Members[1].TricksWon != 1 || av.Members[0].TricksWon != 0 {
			t.Fatalf("tricks won %+v", av.Members)
		}
	}
	// the first card of trick 2: the last trick stays visible (rules §16)
	must(t, g.PlayCard(bob, c("g1"), noShuffle{}))
	if lt := view(carol).LastTrick; lt == nil || lt.Winner != 1 || len(lt.Plays) != 4 {
		t.Fatalf("last trick during trick 2: %+v", lt)
	}
	if len(view(carol).Trick) != 1 || view(carol).TrickLeader != 1 {
		t.Fatalf("trick %v leader %d", view(carol).Trick, view(carol).TrickLeader)
	}
	// only trick 1 (public) remains of the past
	bbAssertOnly(t, g, carol, slices.Concat(cs("y1 y2"), cs("y9 g1"), cs("p3 p7 p5 p6")), "trick 2 in progress")
	must(t, g.PlayCard(carol, c("y1"), noShuffle{}))
	must(t, g.PlayCard(dave, c("y3"), noShuffle{}))
	must(t, g.PlayCard(alice, c("b1"), noShuffle{}))
	lt := view(dave).LastTrick
	if lt == nil || lt.Leader != 1 || len(lt.Plays) != 4 || lt.Plays[0].Card != c("g1") {
		t.Fatalf("last trick after trick 2: %+v", lt)
	}
}

// rules §16: after an attempt all hands are revealed, the unplayed card and JARVIS's hidden cards
// included; the attempt result is shown.
func TestRulesViewDebriefReveals(t *testing.T) {
	g := bbCommGame(t, 1)
	hands := [][]Card{cs(bbH0), cs(bbH1), cs(bbH2), cs(bbH3)}
	must(t, g.GiveUp(bob))
	for s := range 4 {
		v, _ := g.ViewFor(bbID(g, s))
		if v.Phase != PhaseDebrief || v.Attempt.Result == nil || v.Attempt.Result.Reason != ReasonGaveUp {
			t.Fatalf("phase %s result %+v", v.Phase, v.Attempt.Result)
		}
		for o := range 4 {
			if !bbCardsEqual(v.Attempt.Members[o].Cards, hands[o]) {
				t.Fatalf("viewer %d sees seat %d: %v, want %v", s, o, v.Attempt.Members[o].Cards, hands[o])
			}
		}
		if len(v.Attempt.Playable) != 0 {
			t.Fatal("nothing to play in the debrief")
		}
		if len(v.Attempt.Actions.Communicate) != 0 || v.Attempt.Actions.SendDistress || v.Attempt.Actions.Handover {
			t.Fatalf("actions %+v", v.Attempt.Actions)
		}
	}
	// JARVIS's hidden cards and the face-down tasks too
	g = startMission(t, 2, 20)
	must(t, g.GiveUp(alice))
	v, _ := g.ViewFor(bob)
	d := v.Attempt.Members[2]
	if len(d.Cards) != 14 {
		t.Fatalf("JARVIS cards in the debrief: %v", d.Cards)
	}
	for _, tv := range v.Attempt.Tasks {
		if tv.Card == nil {
			t.Fatal("face-down task still hidden in the debrief")
		}
	}
	for _, mem := range v.Attempt.Members {
		if mem.Dummy {
			continue
		}
		if len(mem.Cards) != 13 {
			t.Fatalf("hand in the debrief: %v", mem.Cards)
		}
	}
}

func TestRulesViewBasics(t *testing.T) {
	g := newCrew(t, 3)
	v, err := g.ViewFor(bob)
	must(t, err)
	if v.Attempt != nil || v.Phase != PhaseLobby || v.You != bob || v.HostID != alice || len(v.Players) != 3 {
		t.Fatalf("lobby view %+v", v)
	}
	if !v.Players[0].IsHost || !v.Players[1].IsYou || v.Players[2].Seat != 2 {
		t.Fatalf("players %+v", v.Players)
	}
	if _, err := g.ViewFor("nobody"); err == nil {
		t.Fatal("unknown player")
	}
	must(t, g.StartMission(alice, noShuffle{}))
	bbDraft(t, g)
	for s := range 3 {
		v, _ := g.ViewFor(bbID(g, s))
		av := v.Attempt
		if av.Seat != s || av.Commander != g.Attempt.Commander || av.TotalTricks != 13 || av.TrickNumber != 1 {
			t.Fatalf("view %+v", av)
		}
		if !slices.Equal(av.Actors, g.Actors()) {
			t.Fatalf("actors %v vs %v", av.Actors, g.Actors())
		}
		for o, m := range av.Members {
			if m.MustAct != slices.Contains(g.Actors(), o) {
				t.Fatalf("MustAct of %d: %v", o, m.MustAct)
			}
		}
		acting := s == g.Attempt.Commander
		if acting != (len(av.Playable) > 0) {
			t.Fatalf("seat %d playable %v", s, av.Playable)
		}
	}
}
