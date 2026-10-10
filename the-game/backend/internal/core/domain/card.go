package domain

// CardKind distinguishes plain number cards from future special kinds
// (Extreme command cards keep the number kind; mini-expansions add joker, tens, ...).
type CardKind string

// CardKindNumber is a plain number card (2..99).
const CardKindNumber CardKind = "number"

// Command is the command printed on an Extreme card (rules.md §2). Classic cards carry
// CommandNone. The values double as the camelCase wire strings.
type Command string

const (
	// CommandNone means the card carries no command.
	CommandNone Command = ""

	CommandStop             Command = "stop"             // ⚡ the turn ends immediately
	CommandSkull            Command = "skull"            // ⚡ must be covered in the same turn
	CommandExactlyThree     Command = "exactlyThree"     // ⚡ 3!: exactly 3 cards this turn
	CommandNoCommunication  Command = "noCommunication"  // ∞ no talking while visible
	CommandNoBackwardsTrick Command = "noBackwardsTrick" // ∞ no backwards trick while visible
	CommandSinglePile       Command = "singlePile"       // ∞ all cards of a turn on one pile while visible
	CommandDrawOnlyOne      Command = "drawOnlyOne"      // ∞ draw only 1 card at turn end while visible
)

// CommandClass says how long a command is in effect.
type CommandClass string

const (
	CommandClassNone      CommandClass = ""
	CommandClassLightning CommandClass = "lightning" // effect for the rest of the turn of the player who played it
	CommandClassInfinity  CommandClass = "infinity"  // effect for everyone while visible on top of a pile
)

// Class returns the command's class; CommandClassNone for CommandNone and unknown values.
func (c Command) Class() CommandClass {
	switch c {
	case CommandStop, CommandSkull, CommandExactlyThree:
		return CommandClassLightning
	case CommandNoCommunication, CommandNoBackwardsTrick, CommandSinglePile, CommandDrawOnlyOne:
		return CommandClassInfinity
	default:
		return CommandClassNone
	}
}

// IsLightning reports whether c is a lightning command.
func (c Command) IsLightning() bool { return c.Class() == CommandClassLightning }

// IsInfinity reports whether c is an Infinity command.
func (c Command) IsInfinity() bool { return c.Class() == CommandClassInfinity }

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
