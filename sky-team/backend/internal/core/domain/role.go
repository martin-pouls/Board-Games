package domain

// Role is a player's seat in the cockpit (rules §1). It doubles as the index of every per-role
// array (dice, ready flags, …).
type Role int

const (
	RolePilot   Role = 0 // blue, left
	RoleCopilot Role = 1 // orange, right
)

// Roles lists both roles in index order.
var Roles = [2]Role{RolePilot, RoleCopilot}

// Other returns the partner's role.
func (r Role) Other() Role { return 1 - r }

// String returns the wire name of the role.
func (r Role) String() string {
	if r == RoleCopilot {
		return "copilot"
	}
	return "pilot"
}

// Valid reports whether r is one of the two roles.
func (r Role) Valid() bool { return r == RolePilot || r == RoleCopilot }

// ParseRole maps a wire name back to a Role.
func ParseRole(s string) (Role, bool) {
	switch s {
	case "pilot":
		return RolePilot, true
	case "copilot":
		return RoleCopilot, true
	}
	return 0, false
}

// Colour returns the role's die colour.
func (r Role) Colour() Colour {
	if r == RoleCopilot {
		return ColourOrange
	}
	return ColourBlue
}

// Piece returns the piece kind of the role's own dice.
func (r Role) Piece() PieceKind {
	if r == RoleCopilot {
		return PieceCopilot
	}
	return PiecePilot
}

// PieceKind is what lies on a space: a player's die, an intern token (§12.2) or the traffic die
// placed through Synchronisation (§13).
type PieceKind string

const (
	PiecePilot   PieceKind = "pilot"
	PieceCopilot PieceKind = "copilot"
	PieceIntern  PieceKind = "intern"
	PieceTraffic PieceKind = "traffic"
)

// Colour is the colour of a die or a panel space (§3.1).
type Colour string

const (
	ColourBlue   Colour = "blue"   // Pilot
	ColourOrange Colour = "orange" // Co-Pilot
	ColourBoth   Colour = "both"   // either player
)

// accepts reports whether a space of colour c takes a die of role r.
func (c Colour) accepts(r Role) bool {
	return c == ColourBoth || c == r.Colour()
}
