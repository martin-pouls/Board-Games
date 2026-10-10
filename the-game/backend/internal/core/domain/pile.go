package domain

// Direction is the play direction of a pile.
type Direction string

const (
	DirectionUp   Direction = "up"   // ascending
	DirectionDown Direction = "down" // descending
)

// Pile is one of the play piles. The start card (never in hand) is implied by StartValue.
type Pile struct {
	Direction  Direction
	StartValue int
	Cards      []Card // played cards, last = top
}

// TopValue returns the value of the top card, or StartValue if no card was played.
func (p Pile) TopValue() int {
	if c, ok := p.TopCard(); ok {
		return c.Value
	}
	return p.StartValue
}

// TopCard returns the top played card; false if only the start card is showing.
func (p Pile) TopCard() (Card, bool) {
	if len(p.Cards) == 0 {
		return Card{}, false
	}
	return p.Cards[len(p.Cards)-1], true
}

// NewPiles returns the ascending piles first (StartValue = AscendingStart), then the
// descending piles (StartValue = DescendingStart). Classic: 0,1 up; 2,3 down.
// Cards is an empty, non-nil slice.
func NewPiles(r Ruleset) []Pile {
	piles := make([]Pile, 0, r.AscendingPiles+r.DescendingPiles)
	for range r.AscendingPiles {
		piles = append(piles, Pile{Direction: DirectionUp, StartValue: r.AscendingStart, Cards: []Card{}})
	}
	for range r.DescendingPiles {
		piles = append(piles, Pile{Direction: DirectionDown, StartValue: r.DescendingStart, Cards: []Card{}})
	}
	return piles
}
