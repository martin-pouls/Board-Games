package domain

// CardKind distinguishes plain number cards from future special kinds
// (Extreme command cards keep the number kind; mini-expansions add joker, tens, ...).
type CardKind string

// CardKindNumber is a plain number card (2..99).
const CardKindNumber CardKind = "number"

// Command is the command printed on an Extreme card. Phase 1 only uses CommandNone.
type Command string

// CommandNone means the card carries no command.
const CommandNone Command = ""

// Card is a single playable card. In classic and extreme the Value is unique
// within the deck and identifies the card.
type Card struct {
	Value   int
	Kind    CardKind
	Command Command
}

// NumberCard returns a plain number card with value v.
func NumberCard(v int) Card {
	return Card{Value: v, Kind: CardKindNumber}
}
