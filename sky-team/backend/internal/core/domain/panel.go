package domain

import "slices"

// SpaceID names a space of the control panel (§3.1) or of a module board (§12).
type SpaceID string

const (
	SpaceAxisPilot      SpaceID = "axisPilot"
	SpaceAxisCopilot    SpaceID = "axisCopilot"
	SpaceEnginesPilot   SpaceID = "enginesPilot"
	SpaceEnginesCopilot SpaceID = "enginesCopilot"
	SpaceRadioPilot     SpaceID = "radioPilot"
	SpaceRadioCopilot1  SpaceID = "radioCopilot1"
	SpaceRadioCopilot2  SpaceID = "radioCopilot2"
	SpaceGear12         SpaceID = "gear12"
	SpaceGear34         SpaceID = "gear34"
	SpaceGear56         SpaceID = "gear56"
	SpaceFlaps12        SpaceID = "flaps12"
	SpaceFlaps23        SpaceID = "flaps23"
	SpaceFlaps45        SpaceID = "flaps45"
	SpaceFlaps56        SpaceID = "flaps56"
	SpaceBrakes2        SpaceID = "brakes2"
	SpaceBrakes4        SpaceID = "brakes4"
	SpaceBrakes6        SpaceID = "brakes6"
	SpaceConcentration1 SpaceID = "concentration1"
	SpaceConcentration2 SpaceID = "concentration2"
	SpaceConcentration3 SpaceID = "concentration3"
	SpaceKerosene       SpaceID = "kerosene"      // §12.1
	SpaceInternPilot    SpaceID = "internPilot"   // §12.2
	SpaceInternCopilot  SpaceID = "internCopilot" // §12.2
	SpaceIce2Top        SpaceID = "ice2Top"       // §12.6, blue
	SpaceIce3Top        SpaceID = "ice3Top"
	SpaceIce4Top        SpaceID = "ice4Top"
	SpaceIce5Top        SpaceID = "ice5Top"
	SpaceIce2Bottom     SpaceID = "ice2Bottom" // §12.6, either colour
	SpaceIce3Bottom     SpaceID = "ice3Bottom"
	SpaceIce4Bottom     SpaceID = "ice4Bottom"
	SpaceIce5Bottom     SpaceID = "ice5Bottom"
)

// SpaceKind groups spaces with the same effect.
type SpaceKind string

const (
	KindAxis          SpaceKind = "axis"
	KindEngines       SpaceKind = "engines"
	KindRadio         SpaceKind = "radio"
	KindGear          SpaceKind = "gear"
	KindFlaps         SpaceKind = "flaps"
	KindBrakes        SpaceKind = "brakes"
	KindConcentration SpaceKind = "concentration"
	KindKerosene      SpaceKind = "kerosene"
	KindIntern        SpaceKind = "intern"
	KindIce           SpaceKind = "ice"
)

// SpaceDef is one row of the panel table. A die fits when its value lies in Min..Max. Index is
// the space's position within its kind: the gear switch (0..2), the flap or brake order (0..3,
// 0..2), the ice column (0..3 = values 2..5), the concentration or radio space; 0 otherwise.
type SpaceDef struct {
	ID       SpaceID
	Kind     SpaceKind
	Colour   Colour
	Min, Max int
	Index    int
}

// spaces is the panel table: the 20 base spaces of §3.1, then the module spaces of §12.
var spaces = []SpaceDef{
	{SpaceAxisPilot, KindAxis, ColourBlue, 1, 6, 0},
	{SpaceAxisCopilot, KindAxis, ColourOrange, 1, 6, 1},
	{SpaceEnginesPilot, KindEngines, ColourBlue, 1, 6, 0},
	{SpaceEnginesCopilot, KindEngines, ColourOrange, 1, 6, 1},
	{SpaceRadioPilot, KindRadio, ColourBlue, 1, 6, 0},
	{SpaceRadioCopilot1, KindRadio, ColourOrange, 1, 6, 1},
	{SpaceRadioCopilot2, KindRadio, ColourOrange, 1, 6, 2},
	{SpaceGear12, KindGear, ColourBlue, 1, 2, 0},
	{SpaceGear34, KindGear, ColourBlue, 3, 4, 1},
	{SpaceGear56, KindGear, ColourBlue, 5, 6, 2},
	{SpaceFlaps12, KindFlaps, ColourOrange, 1, 2, 0},
	{SpaceFlaps23, KindFlaps, ColourOrange, 2, 3, 1},
	{SpaceFlaps45, KindFlaps, ColourOrange, 4, 5, 2},
	{SpaceFlaps56, KindFlaps, ColourOrange, 5, 6, 3},
	{SpaceBrakes2, KindBrakes, ColourBlue, 2, 2, 0},
	{SpaceBrakes4, KindBrakes, ColourBlue, 4, 4, 1},
	{SpaceBrakes6, KindBrakes, ColourBlue, 6, 6, 2},
	{SpaceConcentration1, KindConcentration, ColourBoth, 1, 6, 0},
	{SpaceConcentration2, KindConcentration, ColourBoth, 1, 6, 1},
	{SpaceConcentration3, KindConcentration, ColourBoth, 1, 6, 2},
	{SpaceKerosene, KindKerosene, ColourBoth, 1, 6, 0},
	{SpaceInternPilot, KindIntern, ColourBlue, 1, 6, 0},
	{SpaceInternCopilot, KindIntern, ColourOrange, 1, 6, 1},
	{SpaceIce2Top, KindIce, ColourBlue, 2, 2, 0},
	{SpaceIce3Top, KindIce, ColourBlue, 3, 3, 1},
	{SpaceIce4Top, KindIce, ColourBlue, 4, 4, 2},
	{SpaceIce5Top, KindIce, ColourBlue, 5, 5, 3},
	{SpaceIce2Bottom, KindIce, ColourBoth, 2, 2, 0},
	{SpaceIce3Bottom, KindIce, ColourBoth, 3, 3, 1},
	{SpaceIce4Bottom, KindIce, ColourBoth, 4, 4, 2},
	{SpaceIce5Bottom, KindIce, ColourBoth, 5, 5, 3},
}

// Spaces returns a copy of the panel table in table order.
func Spaces() []SpaceDef {
	return slices.Clone(spaces)
}

// SpaceFor looks up a space by ID.
func SpaceFor(id SpaceID) (SpaceDef, bool) {
	for _, s := range spaces {
		if s.ID == id {
			return s, true
		}
	}
	return SpaceDef{}, false
}

// activeSpaces returns the spaces in play with modules, in table order: the base spaces (without
// the brakes under the ice-brake board), the kerosene space (closed by the leak, §12.5), the
// intern spaces and the ice-brake spaces.
func activeSpaces(modules []Module) []SpaceDef {
	ice := slices.Contains(modules, ModuleIceBrakes)
	var out []SpaceDef
	for _, s := range spaces {
		var on bool
		switch s.Kind {
		case KindBrakes:
			on = !ice
		case KindKerosene:
			on = slices.Contains(modules, ModuleKerosene) && !slices.Contains(modules, ModuleKeroseneLeak)
		case KindIntern:
			on = slices.Contains(modules, ModuleIntern)
		case KindIce:
			on = ice
		default:
			on = true
		}
		if on {
			out = append(out, s)
		}
	}
	return out
}

// fits reports whether value meets the space's value constraint.
func fits(s SpaceDef, value int) bool {
	return value >= s.Min && value <= s.Max
}
