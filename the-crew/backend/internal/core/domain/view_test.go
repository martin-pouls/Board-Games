package domain

import (
	"slices"
	"testing"
)

// viewCards collects every card a view shows anywhere.
func viewCards(v PlayerView) []Card {
	a := v.Attempt
	if a == nil {
		return nil
	}
	out := slices.Clone(a.Hand)
	out = append(out, a.Playable...)
	for _, m := range a.Members {
		out = append(out, m.Cards...)
		if m.Communication != nil {
			out = append(out, m.Communication.Card)
		}
		for _, s := range m.Stacks {
			if s.Up != nil {
				out = append(out, *s.Up)
			}
		}
	}
	for _, t := range a.Tasks {
		if t.Card != nil {
			out = append(out, *t.Card)
		}
	}
	for _, p := range a.Trick {
		out = append(out, p.Card)
	}
	if a.LastTrick != nil {
		for _, p := range a.LastTrick.Plays {
			out = append(out, p.Card)
		}
	}
	if a.Pass != nil && a.Pass.Chosen != nil {
		out = append(out, *a.Pass.Chosen)
	}
	if a.Exchange != nil {
		for _, c := range []*Card{a.Exchange.Gave, a.Exchange.Got} {
			if c != nil {
				out = append(out, *c)
			}
		}
	}
	for _, o := range a.Actions.Communicate {
		out = append(out, o.Card)
	}
	return out
}

func TestViewInLobby(t *testing.T) {
	g := newCrew(t, 2)
	v, err := g.ViewFor(bob)
	must(t, err)
	if v.Attempt != nil || v.Phase != PhaseLobby || v.Mission != 1 || len(v.Players) != 2 || !v.Players[1].IsYou || !v.Players[0].IsHost {
		t.Fatalf("view %+v", v)
	}
	if _, err := g.ViewFor("nobody"); err == nil {
		t.Fatal("unknown player")
	}
}

func TestViewHidesOtherHandsAndDummyHiddenCards(t *testing.T) {
	g := startMission(t, 2, 1)
	setTasks(g, "p8:2") // a face-up dummy card: task cards are public anyway
	must(t, g.PlayCard(bob, c("y1"), noShuffle{}))
	a := g.Attempt
	hidden := []Card{}
	for _, s := range a.Members[2].Stacks {
		if s.Down != nil {
			hidden = append(hidden, *s.Down)
		}
	}
	for _, viewer := range []int{0, 1} {
		v, _ := g.ViewFor(crewIDs[viewer])
		other := a.Members[1-viewer].Hand
		for _, card := range viewCards(v) {
			if slices.Contains(other, card) {
				t.Errorf("viewer %d sees %s from the other hand", viewer, card)
			}
			if slices.Contains(hidden, card) {
				t.Errorf("viewer %d sees the dummy's hidden %s", viewer, card)
			}
		}
		if !slices.Equal(v.Attempt.Hand, a.Members[viewer].Hand) {
			t.Errorf("viewer %d: own hand", viewer)
		}
		for _, m := range v.Attempt.Members {
			if m.Cards != nil {
				t.Errorf("viewer %d: revealed cards before the debrief", viewer)
			}
		}
		d := v.Attempt.Members[2]
		if !d.Dummy || len(d.Stacks) != DummyStacks || !d.Stacks[0].Hidden || d.CardCount != 14 {
			t.Errorf("dummy view %+v", d)
		}
	}
	// The dummy's turn: the commander (bob) sees its playable face-up cards, alice sees none.
	vb, _ := g.ViewFor(bob)
	va, _ := g.ViewFor(alice)
	if vb.Attempt.Turn != 2 || !slices.Equal(vb.Attempt.Playable, g.LegalCards(2)) || len(va.Attempt.Playable) != 0 {
		t.Fatalf("playable: bob %v alice %v", vb.Attempt.Playable, va.Attempt.Playable)
	}
	if !vb.Attempt.Members[1].MustAct || va.Attempt.Members[2].MustAct {
		t.Fatal("bob acts for the dummy")
	}
}

func TestViewFaceDownTasks(t *testing.T) {
	g := startMission(t, 4, 37) // commander's decision, 4 face-down tasks
	v, _ := g.ViewFor(alice)
	for _, task := range v.Attempt.Tasks {
		if task.Card != nil || !task.FaceDown {
			t.Fatalf("face-down task shows %+v", task)
		}
	}
	g = startMission(t, 4, 24) // commander's distribution: only the first is revealed
	v, _ = g.ViewFor(alice)
	if v.Attempt.Tasks[0].Card == nil || v.Attempt.Tasks[1].Card != nil || v.Attempt.Poll.Task != 0 {
		t.Fatalf("tasks %+v", v.Attempt.Tasks)
	}
}

func TestViewLastTrick(t *testing.T) {
	g := fourAtPlay(t, 1)
	playAll(t, g, "p1 p3 b9 y2")
	v, _ := g.ViewFor(alice)
	if v.Attempt.LastTrick == nil || v.Attempt.LastTrick.Winner != 1 || len(v.Attempt.Trick) != 0 || v.Attempt.TrickNumber != 2 {
		t.Fatalf("last trick %+v", v.Attempt.LastTrick)
	}
	if v.Attempt.Members[1].TricksWon != 1 {
		t.Fatal("tricks won")
	}
	playAll(t, g, "b2")
	v, _ = g.ViewFor(alice)
	if v.Attempt.LastTrick == nil || v.Attempt.LastTrick.Winner != 1 || len(v.Attempt.Trick) != 1 {
		t.Fatal("the last trick stays visible while the next trick is played")
	}
}

func TestViewDebriefRevealsAll(t *testing.T) {
	g := startMission(t, 2, 1)
	setTasks(g, "p1:2")
	lose(t, g)
	v, _ := g.ViewFor(alice)
	a := g.Attempt
	for i, m := range v.Attempt.Members {
		if !slices.Equal(m.Cards, a.Members[i].Cards()) {
			t.Errorf("seat %d revealed %v", i, m.Cards)
		}
	}
	if len(v.Attempt.Members[2].Cards) != 14 || v.Attempt.Result == nil || v.Attempt.Result.Reason != ReasonGaveUp {
		t.Fatalf("debrief %+v", v.Attempt)
	}
	if len(v.Attempt.Actions.Communicate) != 0 || v.Attempt.Actions.SendDistress {
		t.Fatal("no actions in the debrief")
	}
}

func TestViewActions(t *testing.T) {
	g := fourAtPlay(t, 1)
	v, _ := g.ViewFor(alice)
	want := []CommOption{{c("p1"), PosLowest}, {c("p5"), PosHighest}, {c("b3"), PosOnly}}
	if !slices.Equal(v.Attempt.Actions.Communicate, want) || !v.Attempt.Actions.SendDistress || v.Attempt.Actions.Handover {
		t.Fatalf("actions %+v", v.Attempt.Actions)
	}
	v, _ = g.ViewFor(bob)
	if v.Attempt.Actions.SendDistress {
		t.Fatal("only the commander")
	}
	if v.Attempt.Seat != 1 || v.Attempt.Commander != 0 || v.Attempt.TotalTricks != 10 {
		t.Fatalf("attempt %+v", v.Attempt)
	}
}

func TestEventsCarryNoHiddenCards(t *testing.T) {
	g := fourAtPlay(t, 12)
	must(t, g.SendDistress(alice, DirLeft))
	for i, card := range []string{"b3", "g2", "g4", "y3"} {
		must(t, g.ChoosePassCard(crewIDs[i], c(card)))
	}
	for _, e := range g.PullEvents() {
		switch e.(type) {
		case DistressSignal, PassCardChosen, CardsPassed:
		default:
			t.Fatalf("unexpected event %#v", e)
		}
	}
}
