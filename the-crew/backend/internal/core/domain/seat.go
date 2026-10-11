package domain

import "slices"

// DummyStacks is the number of stacks the dummy gets (rules §12).
const DummyStacks = 7

// DummyStack is one of the dummy's stacks: a face-up card on a hidden one. A played face-up card
// leaves Up nil until the trick ends; then the hidden card turns up (Down moves to Up).
type DummyStack struct {
	Down, Up *Card
}

// Member is a seat in an attempt: a human player or the dummy (Planet Nine: JARVIS). Only the
// legal cards and who acts know the difference; tricks and objectives treat all seats alike.
type Member struct {
	PlayerID PlayerID     // "" for the dummy
	Dummy    bool         //
	Hand     []Card       // humans: sorted, never nil; the dummy: nil
	Stacks   []DummyStack // the dummy: DummyStacks stacks; humans: nil
}

// FaceUp returns the dummy's face-up cards in stack order (empty for a human).
func (m *Member) FaceUp() []Card {
	out := []Card{}
	for _, s := range m.Stacks {
		if s.Up != nil {
			out = append(out, *s.Up)
		}
	}
	return out
}

// Cards returns every card the member still holds: the hand, or all of the dummy's cards, face
// up and hidden. The result is sorted and never nil.
func (m *Member) Cards() []Card {
	if !m.Dummy {
		return cloneCards(m.Hand)
	}
	out := []Card{}
	for _, s := range m.Stacks {
		if s.Up != nil {
			out = append(out, *s.Up)
		}
		if s.Down != nil {
			out = append(out, *s.Down)
		}
	}
	SortCards(out)
	return out
}

// CardCount is the number of cards the member still holds.
func (m *Member) CardCount() int {
	if !m.Dummy {
		return len(m.Hand)
	}
	n := 0
	for _, s := range m.Stacks {
		if s.Up != nil {
			n++
		}
		if s.Down != nil {
			n++
		}
	}
	return n
}

// Playable returns the cards the member may play to a trick led by lead (nil when leading). A
// human follows suit from the hand; the dummy plays only face-up cards and follows suit among
// them, even if a hidden card would match (rules §12).
func (m *Member) Playable(lead *Card) []Card {
	if m.Dummy {
		return legalCards(m.FaceUp(), lead)
	}
	return legalCards(m.Hand, lead)
}

// holds reports whether c is a card the member can currently play from (the hand, or the dummy's
// face-up cards).
func (m *Member) holds(c Card) bool {
	if m.Dummy {
		return slices.Contains(m.FaceUp(), c)
	}
	return slices.Contains(m.Hand, c)
}

// take removes c from the hand or from the dummy's face-up cards. The caller has checked holds.
func (m *Member) take(c Card) {
	if !m.Dummy {
		m.Hand, _ = removeCard(m.Hand, c)
		return
	}
	for i, s := range m.Stacks {
		if s.Up != nil && *s.Up == c {
			m.Stacks[i].Up = nil
			return
		}
	}
}

// give adds c to a human's hand, keeping it sorted.
func (m *Member) give(c Card) {
	m.Hand = append(m.Hand, c)
	SortCards(m.Hand)
}

// flip turns up the hidden card of every stack whose face-up card was played (end of a trick).
func (m *Member) flip() {
	for i, s := range m.Stacks {
		if s.Up == nil && s.Down != nil {
			m.Stacks[i].Up, m.Stacks[i].Down = s.Down, nil
		}
	}
}

func (m Member) clone() Member {
	c := m
	if m.Hand != nil {
		c.Hand = cloneCards(m.Hand)
	}
	if m.Stacks != nil {
		c.Stacks = make([]DummyStack, len(m.Stacks))
		for i, s := range m.Stacks {
			c.Stacks[i] = DummyStack{Down: cloneCard(s.Down), Up: cloneCard(s.Up)}
		}
	}
	return c
}
