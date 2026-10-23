package domain

import "slices"

// PlaneTokens is the size of the plane-token supply (§2).
const PlaneTokens = 12

// MaxCoffee is the most coffee tokens the panel holds (§3.5).
const MaxCoffee = 3

// StartFuel is the kerosene marker's start (§12.1).
const StartFuel = 20

// Flight is the state that persists across rounds. Nil in the lobby.
type Flight struct {
	Round    int   // 1..Rounds, the altitude space (§4.2)
	Position int   // the plane's approach space, 1..N
	Planes   []int // plane tokens per approach space; Planes[0] is space 1
	Axis     int   // -3..3 (±3 only once lost, §3.2)
	Gear     [3]bool
	Flaps    int // deployed flaps, in order (§6.5)
	Brakes   int // deployed brakes, in order (§6.6)
	Ice      int // ice-brake marker slot 0..4: 0 = left of 2, 4 = past 5 (§12.6)
	Coffee   int // on the panel, 0..MaxCoffee
	Rerolls  int // reroll supply (§3.5)
	Fuel     int // kerosene left (§12.1, §12.5); 0 without a kerosene track
	Wind     int // wind ring offset from the top space (§12.3); 0 without wind
	// Intern is the intern token row by value, left to right (§12.2); nil without the module.
	// The Pilot takes Intern[InternLeft] next, the Co-Pilot Intern[InternRight]; the row is used
	// up when InternLeft > InternRight.
	Intern                  []int
	InternLeft, InternRight int
	AdaptationUsed          [2]bool // by role, once per game (§13)
	R                       Round
}

// clone returns a deep copy; nil stays nil.
func (f *Flight) clone() *Flight {
	if f == nil {
		return nil
	}
	c := *f
	c.Planes = slices.Clone(f.Planes)
	c.Intern = slices.Clone(f.Intern)
	c.R = f.R.clone()
	return &c
}

// GearDown returns the number of deployed landing gear.
func (f *Flight) GearDown() int {
	n := 0
	for _, d := range f.Gear {
		if d {
			n++
		}
	}
	return n
}

// AeroBlue is the blue aerodynamics threshold: speeds up to it move 0 spaces (§3.3).
func (f *Flight) AeroBlue() int { return 4 + f.GearDown() }

// AeroOrange is the orange aerodynamics threshold: speeds above it move 2 spaces (§3.3).
func (f *Flight) AeroOrange() int { return 8 + f.Flaps }

// brakeLimits by deployed brakes: the final speed must be lower (§3.4).
var brakeLimits = [4]int{2, 3, 5, 7}

// BrakeLimit returns the speed the final round must stay below (§3.4, §12.6): the normal brakes,
// or the ice-brake marker (2 + slot) with that module.
func (f *Flight) BrakeLimit(iceBrakes bool) int {
	if iceBrakes {
		return 2 + f.Ice
	}
	return brakeLimits[f.Brakes]
}

// PlaneSupply returns the plane tokens not on the approach track.
func (f *Flight) PlaneSupply() int {
	n := PlaneTokens
	for _, p := range f.Planes {
		n -= p
	}
	return n
}

// AtAirport reports whether the plane is on the last approach space.
func (f *Flight) AtAirport() bool { return f.Position == len(f.Planes) }

// RerollRounds returns the altitude rounds still holding a reroll token. A round's token is
// picked up when it starts, so these are the later rounds with an icon on the scenario's side.
func (g *Game) RerollRounds() []int {
	out := []int{}
	if g.Flight == nil {
		return out
	}
	colour := g.ScenarioDef().Colour
	for r := g.Flight.Round + 1; r <= Rounds; r++ {
		if AltitudeRow(r).Reroll(colour) {
			out = append(out, r)
		}
	}
	return out
}

// rerollsInBox returns the reroll tokens neither on the altitude track nor in the supply: spent
// ones and, on the red side, the unused one. Mastery can bring them back (§13).
func (g *Game) rerollsInBox() int {
	return RerollTokens - g.Flight.Rerolls - len(g.RerollRounds())
}

// BrakeLimit returns the game's current brake limit (§3.4, §12.6).
func (g *Game) BrakeLimit() int {
	return g.Flight.BrakeLimit(g.ScenarioDef().Has(ModuleIceBrakes))
}

// FirstPlayer returns the first player of the current round (§4.2).
func (g *Game) FirstPlayer() Role {
	return AltitudeRow(g.Flight.Round).First
}
