package domain

import (
	"fmt"
	"slices"
)

// ScenarioID identifies a scenario (§9), e.g. "yul-green".
type ScenarioID string

// DefaultScenario is the lobby's scenario until the host picks another (§4.1).
const DefaultScenario ScenarioID = "yul-green"

// ScenarioColour is a scenario's difficulty colour (§9). It also picks the altitude side (§4.2).
type ScenarioColour string

const (
	ColourGreen  ScenarioColour = "green"
	ColourYellow ScenarioColour = "yellow"
	ColourRed    ScenarioColour = "red"
	ColourBlack  ScenarioColour = "black"
)

// redSide reports whether the scenario uses the red/black side of the altitude track.
func (c ScenarioColour) redSide() bool {
	return c == ColourRed || c == ColourBlack
}

// Module is an optional rules module (§12). The real-time module is not in the app (§12.4).
type Module string

const (
	ModuleKerosene     Module = "kerosene"
	ModuleKeroseneLeak Module = "keroseneLeak"
	ModuleIntern       Module = "intern"
	ModuleWind         Module = "wind"
	ModuleIceBrakes    Module = "iceBrakes"
)

// ScenarioDef is one scenario of §9. RealTimeDropped marks the scenarios that list the real-time
// module, which the app plays without it.
type ScenarioDef struct {
	ID              ScenarioID
	Airport         string // IATA code, e.g. "YUL"
	City            string // airport name as printed, e.g. "Montréal-Trudeau"
	Colour          ScenarioColour
	Track           TrackID
	Modules         []Module
	Abilities       int // number of special-ability cards (§13)
	RealTimeDropped bool
}

// Has reports whether the scenario plays with module m.
func (s ScenarioDef) Has(m Module) bool {
	return slices.Contains(s.Modules, m)
}

// hasKerosene reports whether a kerosene track is in play (§12.1 or §12.5).
func (s ScenarioDef) hasKerosene() bool {
	return s.Has(ModuleKerosene) || s.Has(ModuleKeroseneLeak)
}

// TrackID identifies an approach track face (§10), e.g. "YUL-G".
type TrackID string

// TrackDef is one approach track face. Spaces[0] is the start, the last space the airport.
type TrackDef struct {
	ID     TrackID
	Spaces []TrackSpace
}

// TrackSpace is one approach space: plane tokens at setup, traffic-die icons (§11.1) and the
// turn icon (§11.2).
type TrackSpace struct {
	Planes, Traffic int
	Turn            TurnMask
}

// Len returns N, the number of spaces (the airport's number).
func (t TrackDef) Len() int { return len(t.Spaces) }

// TurnMask is a turn icon (§11.2): bit i set means axis position i-2 is permitted. Zero means no
// icon (every icon permits at least one position).
type TurnMask uint8

// turn parses a 5-character mask for the axis positions -2..+2 ('o' permitted, 'x' forbidden);
// "" is no icon. It panics on bad input (data tables only).
func turn(s string) TurnMask {
	if s == "" {
		return 0
	}
	if len(s) != 5 {
		panic(fmt.Sprintf("turn mask %q: want 5 characters", s))
	}
	var m TurnMask
	for i := range 5 {
		switch s[i] {
		case 'o':
			m |= 1 << i
		case 'x':
		default:
			panic(fmt.Sprintf("turn mask %q: bad character %q", s, s[i]))
		}
	}
	if m == 0 {
		panic(fmt.Sprintf("turn mask %q permits nothing", s))
	}
	return m
}

// Present reports whether the space has a turn icon.
func (m TurnMask) Present() bool { return m != 0 }

// Permits reports whether axis position axis (-2..+2) is permitted. Without an icon every
// position is permitted.
func (m TurnMask) Permits(axis int) bool {
	if m == 0 {
		return true
	}
	if axis < -2 || axis > 2 {
		return false
	}
	return m&(1<<(axis+2)) != 0
}

// String returns the 5-character form ("" without an icon).
func (m TurnMask) String() string {
	if m == 0 {
		return ""
	}
	b := []byte("xxxxx")
	for i := range 5 {
		if m&(1<<i) != 0 {
			b[i] = 'o'
		}
	}
	return string(b)
}

// Scenarios returns the scenarios in §9 order (the flight log's 21, then the 2 promo faces).
func Scenarios() []ScenarioDef {
	out := make([]ScenarioDef, len(scenarios))
	for i, s := range scenarios {
		s.Modules = slices.Clone(s.Modules)
		out[i] = s
	}
	return out
}

// ScenarioFor looks up a scenario.
func ScenarioFor(id ScenarioID) (ScenarioDef, bool) {
	for _, s := range scenarios {
		if s.ID == id {
			s.Modules = slices.Clone(s.Modules)
			return s, true
		}
	}
	return ScenarioDef{}, false
}

// TrackFor looks up an approach track.
func TrackFor(id TrackID) (TrackDef, bool) {
	for _, t := range tracks {
		if t.ID == id {
			t.Spaces = slices.Clone(t.Spaces)
			return t, true
		}
	}
	return TrackDef{}, false
}

// Rounds is the number of rounds of every game (§5.5).
const Rounds = 7

// AltitudeSpace is one row of the altitude track (§4.2).
type AltitudeSpace struct {
	Round       int
	Altitude    int // feet; 0 = landing
	First       Role
	RerollGreen bool // reroll icon on the green/yellow side
	RerollRed   bool // reroll icon on the red/black side
}

// Reroll reports whether the row has a reroll icon on the side colour plays on.
func (a AltitudeSpace) Reroll(colour ScenarioColour) bool {
	if colour.redSide() {
		return a.RerollRed
	}
	return a.RerollGreen
}

// altitude is the altitude track, round 1 first.
var altitude = [Rounds]AltitudeSpace{
	{1, 6000, RolePilot, true, true},
	{2, 5000, RoleCopilot, false, false},
	{3, 4000, RolePilot, false, false},
	{4, 3000, RoleCopilot, false, false},
	{5, 2000, RolePilot, true, false},
	{6, 1000, RoleCopilot, false, false},
	{7, 0, RolePilot, false, false},
}

// AltitudeRow returns the altitude space of round (1..Rounds). It panics outside that range.
func AltitudeRow(round int) AltitudeSpace {
	return altitude[round-1]
}

// RerollTokens is the number of reroll tokens in the box (§2).
const RerollTokens = 2

// trafficFaces are the faces of the traffic die (§2).
var trafficFaces = [6]int{2, 3, 3, 4, 4, 5}

// TrafficFaces returns the faces of the traffic die.
func TrafficFaces() [6]int { return trafficFaces }

// windByDistance is the wind value by the ring distance min(p, 20-p) from the top space (§12.3).
var windByDistance = [11]int{3, 3, 2, 2, 1, 0, -1, -2, -2, -3, -3}

// WindRingSpaces is the number of spaces on the wind ring (§12.3).
const WindRingSpaces = 20

// WindValue returns the wind value the airplane token points at, offset spaces clockwise from
// the top space (negative = counter-clockwise).
func WindValue(offset int) int {
	p := ((offset % WindRingSpaces) + WindRingSpaces) % WindRingSpaces
	return windByDistance[min(p, WindRingSpaces-p)]
}

// AbilityID identifies a special-ability card (§13).
type AbilityID string

const (
	AbilityAdaptation      AbilityID = "adaptation"
	AbilityAnticipation    AbilityID = "anticipation"
	AbilityControl         AbilityID = "control"
	AbilityMastery         AbilityID = "mastery"
	AbilitySynchronisation AbilityID = "synchronisation"
	AbilityWorkingTogether AbilityID = "workingTogether"
)

// abilities are the six cards in §13 order.
var abilities = []AbilityID{
	AbilityAdaptation, AbilityAnticipation, AbilityControl,
	AbilityMastery, AbilitySynchronisation, AbilityWorkingTogether,
}

// Abilities returns the six ability cards in §13 order.
func Abilities() []AbilityID {
	return slices.Clone(abilities)
}
