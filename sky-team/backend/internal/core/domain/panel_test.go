package domain

import (
	"slices"
	"testing"
)

func TestSpacesTable(t *testing.T) {
	all := Spaces()
	if len(all) != 31 {
		t.Fatalf("%d spaces, want 20 base + 11 module", len(all))
	}
	seen := map[SpaceID]bool{}
	for _, s := range all {
		if seen[s.ID] {
			t.Fatalf("duplicate %s", s.ID)
		}
		seen[s.ID] = true
		if s.Min < 1 || s.Max > 6 || s.Min > s.Max {
			t.Fatalf("%s: values %d..%d", s.ID, s.Min, s.Max)
		}
		if got, ok := SpaceFor(s.ID); !ok || got != s {
			t.Fatalf("SpaceFor(%s) = %+v", s.ID, got)
		}
	}
	if _, ok := SpaceFor("wing"); ok {
		t.Fatal("unknown space found")
	}
	all[0].Kind = KindFlaps
	if Spaces()[0].Kind != KindAxis {
		t.Fatal("Spaces must return a copy")
	}
}

func TestSpaceConstraints(t *testing.T) {
	// §3.1 per space: colour and the values that fit.
	tests := []struct {
		id     SpaceID
		colour Colour
		fit    []int
	}{
		{SpaceAxisPilot, ColourBlue, []int{1, 2, 3, 4, 5, 6}},
		{SpaceEnginesCopilot, ColourOrange, []int{1, 2, 3, 4, 5, 6}},
		{SpaceRadioPilot, ColourBlue, []int{1, 2, 3, 4, 5, 6}},
		{SpaceRadioCopilot2, ColourOrange, []int{1, 2, 3, 4, 5, 6}},
		{SpaceGear12, ColourBlue, []int{1, 2}},
		{SpaceGear34, ColourBlue, []int{3, 4}},
		{SpaceGear56, ColourBlue, []int{5, 6}},
		{SpaceFlaps12, ColourOrange, []int{1, 2}},
		{SpaceFlaps23, ColourOrange, []int{2, 3}},
		{SpaceFlaps45, ColourOrange, []int{4, 5}},
		{SpaceFlaps56, ColourOrange, []int{5, 6}},
		{SpaceBrakes2, ColourBlue, []int{2}},
		{SpaceBrakes4, ColourBlue, []int{4}},
		{SpaceBrakes6, ColourBlue, []int{6}},
		{SpaceConcentration3, ColourBoth, []int{1, 2, 3, 4, 5, 6}},
		{SpaceKerosene, ColourBoth, []int{1, 2, 3, 4, 5, 6}},
		{SpaceInternPilot, ColourBlue, []int{1, 2, 3, 4, 5, 6}},
		{SpaceInternCopilot, ColourOrange, []int{1, 2, 3, 4, 5, 6}},
		{SpaceIce2Top, ColourBlue, []int{2}},
		{SpaceIce5Top, ColourBlue, []int{5}},
		{SpaceIce3Bottom, ColourBoth, []int{3}},
		{SpaceIce4Bottom, ColourBoth, []int{4}},
	}
	for _, tt := range tests {
		s, _ := SpaceFor(tt.id)
		if s.Colour != tt.colour {
			t.Errorf("%s: colour %s, want %s", tt.id, s.Colour, tt.colour)
		}
		var fit []int
		for v := 0; v <= 7; v++ {
			if fits(s, v) {
				fit = append(fit, v)
			}
		}
		if !slices.Equal(fit, tt.fit) {
			t.Errorf("%s: fits %v, want %v", tt.id, fit, tt.fit)
		}
	}
	// Order indices of the ordered kinds.
	for kind, ids := range map[SpaceKind][]SpaceID{
		KindGear:   {SpaceGear12, SpaceGear34, SpaceGear56},
		KindFlaps:  {SpaceFlaps12, SpaceFlaps23, SpaceFlaps45, SpaceFlaps56},
		KindBrakes: {SpaceBrakes2, SpaceBrakes4, SpaceBrakes6},
		KindIce:    {SpaceIce2Top, SpaceIce3Top, SpaceIce4Top, SpaceIce5Top},
	} {
		for i, id := range ids {
			if s, _ := SpaceFor(id); s.Kind != kind || s.Index != i {
				t.Errorf("%s: kind %s index %d, want %s %d", id, s.Kind, s.Index, kind, i)
			}
		}
	}
}

func TestColourAccepts(t *testing.T) {
	if !ColourBlue.accepts(RolePilot) || ColourBlue.accepts(RoleCopilot) ||
		!ColourOrange.accepts(RoleCopilot) || ColourOrange.accepts(RolePilot) ||
		!ColourBoth.accepts(RolePilot) || !ColourBoth.accepts(RoleCopilot) {
		t.Fatal("colour rules (§3.1)")
	}
}

func TestActiveSpaces(t *testing.T) {
	ids := func(modules ...Module) []SpaceID {
		var out []SpaceID
		for _, s := range activeSpaces(modules) {
			out = append(out, s.ID)
		}
		return out
	}
	base := ids()
	if len(base) != 20 || slices.Contains(base, SpaceKerosene) || !slices.Contains(base, SpaceBrakes6) {
		t.Fatalf("base %v", base)
	}
	if k := ids(ModuleKerosene); len(k) != 21 || !slices.Contains(k, SpaceKerosene) {
		t.Fatalf("kerosene %v", k)
	}
	// The leak closes the kerosene space (§12.5).
	if k := ids(ModuleKeroseneLeak); slices.Contains(k, SpaceKerosene) || len(k) != 20 {
		t.Fatalf("leak %v", k)
	}
	if k := ids(ModuleIntern); len(k) != 22 || !slices.Contains(k, SpaceInternPilot) || !slices.Contains(k, SpaceInternCopilot) {
		t.Fatalf("intern %v", k)
	}
	// The ice-brake board covers the brakes (§12.6).
	ice := ids(ModuleIceBrakes)
	if len(ice) != 25 || slices.Contains(ice, SpaceBrakes2) || slices.Contains(ice, SpaceBrakes4) ||
		slices.Contains(ice, SpaceBrakes6) || !slices.Contains(ice, SpaceIce5Bottom) {
		t.Fatalf("ice %v", ice)
	}
	if w := ids(ModuleWind); !slices.Equal(w, base) {
		t.Fatalf("wind %v", w)
	}
}
