package domain_test

// Expectation tables typed from docs/rules.md (§3.1, §3.3, §3.4, §4.2, §9, §10, §12.3, §12.6),
// compared with the domain's tables through the public accessors.

import (
	"fmt"
	"slices"
	"sort"
	"testing"

	"github.com/martin-pouls/sky-team/backend/internal/core/domain"
)

func TestDataConstants(t *testing.T) {
	// §1, §2, §3.5, §5.5, §12.1, §12.2, §12.3, §14
	checks := []struct {
		name      string
		got, want int
	}{
		{"HandSize §5.2", domain.HandSize, 4},
		{"InternTokens §12.2", domain.InternTokens, 6},
		{"MaxCoffee §3.5", domain.MaxCoffee, 3},
		{"PlaneTokens §2", domain.PlaneTokens, 12},
		{"Players §14", domain.Players, 2},
		{"RerollTokens §2", domain.RerollTokens, 2},
		{"Rounds §5.5", domain.Rounds, 7},
		{"StartFuel §12.1", domain.StartFuel, 20},
		{"WindRingSpaces §12.3", domain.WindRingSpaces, 20},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, c.got, c.want)
		}
	}
}

func TestDataTrafficDie(t *testing.T) {
	// §2: faces 2, 3, 3, 4, 4, 5
	got := domain.TrafficFaces()
	s := got[:]
	sort.Ints(s)
	if want := []int{2, 3, 3, 4, 4, 5}; !slices.Equal(s, want) {
		t.Errorf("traffic faces = %v, want %v", s, want)
	}
}

func TestDataAbilityList(t *testing.T) {
	// §13 order
	want := []domain.AbilityID{"adaptation", "anticipation", "control", "mastery", "synchronisation", "workingTogether"}
	if got := domain.Abilities(); !slices.Equal(got, want) {
		t.Errorf("Abilities() = %v, want %v", got, want)
	}
}

// --- §3.1 panel ---

type spaceExp struct {
	id       domain.SpaceID
	kind     domain.SpaceKind
	colour   domain.Colour
	min, max int
	index    int // checked only when idx is true
	idx      bool
}

func panelExpect() []spaceExp {
	b, o, both := domain.ColourBlue, domain.ColourOrange, domain.ColourBoth
	return []spaceExp{
		{id: "axisPilot", kind: domain.KindAxis, colour: b, min: 1, max: 6},
		{id: "axisCopilot", kind: domain.KindAxis, colour: o, min: 1, max: 6},
		{id: "enginesPilot", kind: domain.KindEngines, colour: b, min: 1, max: 6},
		{id: "enginesCopilot", kind: domain.KindEngines, colour: o, min: 1, max: 6},
		{id: "radioPilot", kind: domain.KindRadio, colour: b, min: 1, max: 6},
		{id: "radioCopilot1", kind: domain.KindRadio, colour: o, min: 1, max: 6},
		{id: "radioCopilot2", kind: domain.KindRadio, colour: o, min: 1, max: 6},
		{id: "gear12", kind: domain.KindGear, colour: b, min: 1, max: 2, index: 0, idx: true},
		{id: "gear34", kind: domain.KindGear, colour: b, min: 3, max: 4, index: 1, idx: true},
		{id: "gear56", kind: domain.KindGear, colour: b, min: 5, max: 6, index: 2, idx: true},
		{id: "flaps12", kind: domain.KindFlaps, colour: o, min: 1, max: 2, index: 0, idx: true},
		{id: "flaps23", kind: domain.KindFlaps, colour: o, min: 2, max: 3, index: 1, idx: true},
		{id: "flaps45", kind: domain.KindFlaps, colour: o, min: 4, max: 5, index: 2, idx: true},
		{id: "flaps56", kind: domain.KindFlaps, colour: o, min: 5, max: 6, index: 3, idx: true},
		{id: "brakes2", kind: domain.KindBrakes, colour: b, min: 2, max: 2, index: 0, idx: true},
		{id: "brakes4", kind: domain.KindBrakes, colour: b, min: 4, max: 4, index: 1, idx: true},
		{id: "brakes6", kind: domain.KindBrakes, colour: b, min: 6, max: 6, index: 2, idx: true},
		{id: "concentration1", kind: domain.KindConcentration, colour: both, min: 1, max: 6},
		{id: "concentration2", kind: domain.KindConcentration, colour: both, min: 1, max: 6},
		{id: "concentration3", kind: domain.KindConcentration, colour: both, min: 1, max: 6},
		// §12.1, §12.2
		{id: "kerosene", kind: domain.KindKerosene, colour: both, min: 1, max: 6},
		{id: "internPilot", kind: domain.KindIntern, colour: b, min: 1, max: 6},
		{id: "internCopilot", kind: domain.KindIntern, colour: o, min: 1, max: 6},
		// §12.6: tops blue, bottoms either; columns 2..5
		{id: "ice2Top", kind: domain.KindIce, colour: b, min: 2, max: 2, index: 0, idx: true},
		{id: "ice3Top", kind: domain.KindIce, colour: b, min: 3, max: 3, index: 1, idx: true},
		{id: "ice4Top", kind: domain.KindIce, colour: b, min: 4, max: 4, index: 2, idx: true},
		{id: "ice5Top", kind: domain.KindIce, colour: b, min: 5, max: 5, index: 3, idx: true},
		{id: "ice2Bottom", kind: domain.KindIce, colour: both, min: 2, max: 2, index: 0, idx: true},
		{id: "ice3Bottom", kind: domain.KindIce, colour: both, min: 3, max: 3, index: 1, idx: true},
		{id: "ice4Bottom", kind: domain.KindIce, colour: both, min: 4, max: 4, index: 2, idx: true},
		{id: "ice5Bottom", kind: domain.KindIce, colour: both, min: 5, max: 5, index: 3, idx: true},
	}
}

func TestDataPanelSpaces(t *testing.T) {
	for _, e := range panelExpect() {
		t.Run(string(e.id)+" §3.1", func(t *testing.T) {
			d, ok := domain.SpaceFor(e.id)
			if !ok {
				t.Fatalf("SpaceFor(%s) not found", e.id)
			}
			if d.ID != e.id || d.Kind != e.kind || d.Colour != e.colour || d.Min != e.min || d.Max != e.max {
				t.Errorf("got %+v, want kind=%s colour=%s %d..%d", d, e.kind, e.colour, e.min, e.max)
			}
			if e.idx && d.Index != e.index {
				t.Errorf("Index = %d, want %d", d.Index, e.index)
			}
		})
	}
	if _, ok := domain.SpaceFor("nope"); ok {
		t.Error("SpaceFor(unknown) found")
	}
}

func TestDataPanelTableContents(t *testing.T) {
	// the table lists exactly the spaces of §3.1, §12.1, §12.2, §12.6, each once
	want := panelExpect()
	got := domain.Spaces()
	if len(got) != len(want) {
		t.Errorf("Spaces() has %d rows, want %d", len(got), len(want))
	}
	seen := map[domain.SpaceID]int{}
	for _, d := range got {
		seen[d.ID]++
	}
	for _, e := range want {
		if seen[e.id] != 1 {
			t.Errorf("space %s appears %d times, want 1", e.id, seen[e.id])
		}
	}
	// the base spaces keep the §3.1 order
	var base []domain.SpaceID
	for _, d := range got {
		switch d.Kind {
		case domain.KindKerosene, domain.KindIntern, domain.KindIce:
		default:
			base = append(base, d.ID)
		}
	}
	var wantBase []domain.SpaceID
	for _, e := range want[:20] {
		wantBase = append(wantBase, e.id)
	}
	if !slices.Equal(base, wantBase) {
		t.Errorf("base order = %v, want %v", base, wantBase)
	}
}

func TestDataPerPlayerSpaceCounts(t *testing.T) {
	// §3.1: Pilot has 1 radio space, Co-Pilot 2; gear and brakes Pilot only, flaps Co-Pilot only
	count := func(k domain.SpaceKind, c domain.Colour) int {
		n := 0
		for _, d := range domain.Spaces() {
			if d.Kind == k && d.Colour == c {
				n++
			}
		}
		return n
	}
	cases := []struct {
		k    domain.SpaceKind
		c    domain.Colour
		want int
	}{
		{domain.KindRadio, domain.ColourBlue, 1}, {domain.KindRadio, domain.ColourOrange, 2},
		{domain.KindGear, domain.ColourBlue, 3}, {domain.KindGear, domain.ColourOrange, 0},
		{domain.KindFlaps, domain.ColourOrange, 4}, {domain.KindFlaps, domain.ColourBlue, 0},
		{domain.KindBrakes, domain.ColourBlue, 3}, {domain.KindBrakes, domain.ColourOrange, 0},
		{domain.KindConcentration, domain.ColourBoth, 3},
	}
	for _, c := range cases {
		if got := count(c.k, c.c); got != c.want {
			t.Errorf("%s/%s spaces = %d, want %d", c.k, c.c, got, c.want)
		}
	}
}

func TestDataRoleColours(t *testing.T) {
	// §1: Pilot blue, Co-Pilot orange
	if domain.RolePilot.Colour() != domain.ColourBlue || domain.RoleCopilot.Colour() != domain.ColourOrange {
		t.Error("role colours wrong")
	}
	if domain.RolePilot.Other() != domain.RoleCopilot || domain.RoleCopilot.Other() != domain.RolePilot {
		t.Error("Other wrong")
	}
}

// --- §3.3 / §3.4 / §12.6 gauges ---

func TestDataAerodynamics(t *testing.T) {
	// §3.3: blue 4 +1 per gear (max 7); orange 8 +1 per flap (max 12)
	for mask := 0; mask < 8; mask++ {
		gear := [3]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		n := 0
		for _, g := range gear {
			if g {
				n++
			}
		}
		t.Run(fmt.Sprintf("gear %v §3.3", gear), func(t *testing.T) {
			g := domain.NewFlightForTest(domain.FlightSetup{Gear: gear})
			if got := g.Flight.AeroBlue(); got != 4+n {
				t.Errorf("AeroBlue = %d, want %d", got, 4+n)
			}
			if got := g.Flight.GearDown(); got != n {
				t.Errorf("GearDown = %d, want %d", got, n)
			}
		})
	}
	for flaps := 0; flaps <= 4; flaps++ {
		t.Run(fmt.Sprintf("flaps %d §3.3", flaps), func(t *testing.T) {
			g := domain.NewFlightForTest(domain.FlightSetup{Flaps: flaps})
			if got := g.Flight.AeroOrange(); got != 8+flaps {
				t.Errorf("AeroOrange = %d, want %d", got, 8+flaps)
			}
		})
	}
}

func TestDataBrakeLimit(t *testing.T) {
	// §3.4: marker left of 2 / between 2-3 / between 4-5 / right of 6; speed must be lower than the
	// marker, so the limit (exclusive) is 2, 3, 5, 7.
	normal := []int{2, 3, 5, 7}
	for brakes, want := range normal {
		t.Run(fmt.Sprintf("brakes %d §3.4", brakes), func(t *testing.T) {
			g := domain.NewFlightForTest(domain.FlightSetup{Brakes: brakes})
			if got := g.Flight.BrakeLimit(false); got != want {
				t.Errorf("Flight.BrakeLimit(false) = %d, want %d", got, want)
			}
			if got := g.BrakeLimit(); got != want {
				t.Errorf("Game.BrakeLimit() = %d, want %d", got, want)
			}
		})
	}
	// §12.6: left of 2, then one slot after each column; past 5 means speed 5 or less lands.
	for slot := 0; slot <= 4; slot++ {
		t.Run(fmt.Sprintf("ice slot %d §12.6", slot), func(t *testing.T) {
			g := domain.NewFlightForTest(domain.FlightSetup{Scenario: "kef-yellow", Abilities: domain.Abilities()[:1], Ice: slot})
			want := 2 + slot
			if got := g.Flight.BrakeLimit(true); got != want {
				t.Errorf("Flight.BrakeLimit(true) = %d, want %d", got, want)
			}
			if got := g.BrakeLimit(); got != want {
				t.Errorf("Game.BrakeLimit() = %d, want %d", got, want)
			}
		})
	}
}

func TestDataPlaneSupplyAndAirport(t *testing.T) {
	// §2 12 tokens; §6.2 airport is the last space (YUL-G: N = 7)
	g := domain.NewFlightForTest(domain.FlightSetup{Planes: []int{0, 1, 2, 3, 0, 4, 0}, Position: 7})
	if got := g.Flight.PlaneSupply(); got != 2 {
		t.Errorf("PlaneSupply = %d, want 2", got)
	}
	if !g.Flight.AtAirport() {
		t.Error("position 7 of 7 should be at the airport")
	}
	g = domain.NewFlightForTest(domain.FlightSetup{Position: 6})
	if g.Flight.AtAirport() {
		t.Error("position 6 of 7 is not the airport")
	}
}

// --- §12.3 wind ---

func TestDataWindValue(t *testing.T) {
	// min(p, 20-p) -> wind
	byDist := []int{3, 3, 2, 2, 1, 0, -1, -2, -2, -3, -3}
	for off := -45; off <= 45; off++ {
		p := ((off % 20) + 20) % 20
		d := min(p, 20-p)
		want := byDist[d]
		if got := domain.WindValue(off); got != want {
			t.Errorf("WindValue(%d) = %d, want %d (§12.3, p=%d)", off, got, want, p)
		}
	}
}

// --- §4.2 altitude ---

func TestDataAltitudeTrack(t *testing.T) {
	rows := []struct {
		round, alt         int
		first              domain.Role
		rerollGY, rerollRB bool
	}{
		{1, 6000, domain.RolePilot, true, true},
		{2, 5000, domain.RoleCopilot, false, false},
		{3, 4000, domain.RolePilot, false, false},
		{4, 3000, domain.RoleCopilot, false, false},
		{5, 2000, domain.RolePilot, true, false},
		{6, 1000, domain.RoleCopilot, false, false},
		{7, 0, domain.RolePilot, false, false},
	}
	for _, r := range rows {
		t.Run(fmt.Sprintf("round %d §4.2", r.round), func(t *testing.T) {
			a := domain.AltitudeRow(r.round)
			if a.Round != r.round || a.Altitude != r.alt || a.First != r.first || a.RerollGreen != r.rerollGY || a.RerollRed != r.rerollRB {
				t.Errorf("got %+v", a)
			}
			for _, c := range []domain.ScenarioColour{domain.ColourGreen, domain.ColourYellow} {
				if a.Reroll(c) != r.rerollGY {
					t.Errorf("Reroll(%s) = %v, want %v", c, a.Reroll(c), r.rerollGY)
				}
			}
			for _, c := range []domain.ScenarioColour{domain.ColourRed, domain.ColourBlack} {
				if a.Reroll(c) != r.rerollRB {
					t.Errorf("Reroll(%s) = %v, want %v", c, a.Reroll(c), r.rerollRB)
				}
			}
		})
	}
	for _, bad := range []int{0, 8, -1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("AltitudeRow(%d) did not panic", bad)
				}
			}()
			domain.AltitudeRow(bad)
		}()
	}
}

func TestDataFirstPlayerPerRound(t *testing.T) {
	// §4.2 via the game
	want := []domain.Role{domain.RolePilot, domain.RoleCopilot, domain.RolePilot, domain.RoleCopilot, domain.RolePilot, domain.RoleCopilot, domain.RolePilot}
	for r := 1; r <= 7; r++ {
		g := domain.NewFlightForTest(domain.FlightSetup{Round: r})
		if got := g.FirstPlayer(); got != want[r-1] {
			t.Errorf("round %d first = %v, want %v", r, got, want[r-1])
		}
	}
}

func TestDataRerollRounds(t *testing.T) {
	// §4.2, §5.1: at round 1 the round-1 token is already picked up; green/yellow still hold round 5
	for _, sc := range domain.Scenarios() {
		t.Run(string(sc.ID)+" §4.2", func(t *testing.T) {
			g := startScenario(t, sc.ID)
			want := []int{}
			if sc.Colour == domain.ColourGreen || sc.Colour == domain.ColourYellow {
				want = []int{5}
			}
			got := g.RerollRounds()
			if len(got) != len(want) || (len(want) > 0 && !slices.Equal(got, want)) {
				t.Errorf("RerollRounds = %v, want %v", got, want)
			}
			if g.Flight.Rerolls != 1 {
				t.Errorf("Rerolls after round-1 pickup = %d, want 1", g.Flight.Rerolls)
			}
		})
	}
}

// --- §9 scenarios ---

type scenExp struct {
	id        domain.ScenarioID
	airport   string
	city      string
	colour    domain.ScenarioColour
	track     domain.TrackID
	modules   []domain.Module
	abilities int
	realTime  bool
}

func scenarioExpect() []scenExp {
	K, L, I, W, IB := domain.ModuleKerosene, domain.ModuleKeroseneLeak, domain.ModuleIntern, domain.ModuleWind, domain.ModuleIceBrakes
	g, y, r, b := domain.ColourGreen, domain.ColourYellow, domain.ColourRed, domain.ColourBlack
	return []scenExp{
		{"yul-green", "YUL", "Montréal-Trudeau", g, "YUL-G", nil, 0, false},
		{"lhr-green", "LHR", "Heathrow", g, "LHR-G", nil, 0, false},
		{"hnd-green", "HND", "Haneda", g, "HND-G", nil, 0, false},
		{"osl-green", "OSL", "Gardermoen", g, "OSL-G", []domain.Module{K}, 0, false},
		{"atl-green", "ATL", "Hartsfield-Jackson", g, "ATL-G", []domain.Module{I}, 0, false},
		{"prg-green", "PRG", "Václav Havel", g, "PRG-G", []domain.Module{K}, 2, false},
		{"lhr-yellow", "LHR", "Heathrow", y, "LHR-Y", []domain.Module{I}, 0, false},
		{"tgu-yellow", "TGU", "Toncontín", y, "TGU-Y", []domain.Module{K}, 2, false},
		{"gig-yellow", "GIG", "Galeão", y, "GIG-Y", []domain.Module{W}, 1, false},
		{"kef-yellow", "KEF", "Keflavík", y, "KEF-Y", []domain.Module{IB}, 1, false},
		{"prg-yellow", "PRG", "Václav Havel", y, "PRG-Y", []domain.Module{L}, 2, false},
		{"kul-yellow", "KUL", "Kuala Lumpur", y, "KUL-Y", []domain.Module{K}, 1, false},
		{"atl-yellow", "ATL", "Hartsfield-Jackson", y, "ATL-Y", []domain.Module{L}, 1, false},
		{"pbh-red", "PBH", "Paro", r, "PBH-R", []domain.Module{K}, 2, true},
		{"hnd-red", "HND", "Haneda", r, "HND-R", []domain.Module{K, I}, 1, false},
		{"gig-red", "GIG", "Galeão", r, "GIG-R", []domain.Module{W, L}, 2, false},
		{"osl-red", "OSL", "Gardermoen", r, "OSL-R", []domain.Module{L, IB}, 2, false},
		{"tgu-red", "TGU", "Toncontín", r, "TGU-R", []domain.Module{K, W}, 2, false},
		{"kef-black", "KEF", "Keflavík", b, "KEF-B", []domain.Module{W, IB}, 2, false},
		{"kul-black", "KUL", "Kuala Lumpur", b, "KUL-B", []domain.Module{K}, 2, true},
		{"pbh-black", "PBH", "Paro", b, "PBH-B", []domain.Module{K}, 2, true},
		{"dus-yellow", "DUS", "Düsseldorf", y, "DUS-Y", []domain.Module{L}, 2, false},
		{"dus-black", "DUS", "Düsseldorf", b, "DUS-B", []domain.Module{L}, 1, true},
	}
}

func sortedModules(m []domain.Module) []string {
	s := make([]string, len(m))
	for i, x := range m {
		s[i] = string(x)
	}
	sort.Strings(s)
	return s
}

func TestDataScenarioTable(t *testing.T) {
	exp := scenarioExpect()
	all := domain.Scenarios()
	if len(all) != 23 {
		t.Fatalf("Scenarios() has %d entries, want 23 (§9)", len(all))
	}
	for i, e := range exp {
		t.Run(string(e.id)+" §9", func(t *testing.T) {
			if all[i].ID != e.id {
				t.Errorf("order: Scenarios()[%d] = %s, want %s", i, all[i].ID, e.id)
			}
			s, ok := domain.ScenarioFor(e.id)
			if !ok {
				t.Fatal("ScenarioFor not found")
			}
			if s.ID != e.id || s.Airport != e.airport || s.City != e.city || s.Colour != e.colour || s.Track != e.track ||
				s.Abilities != e.abilities || s.RealTimeDropped != e.realTime {
				t.Errorf("got %+v, want airport=%s city=%s colour=%s track=%s abilities=%d realTime=%v",
					s, e.airport, e.city, e.colour, e.track, e.abilities, e.realTime)
			}
			if !slices.Equal(sortedModules(s.Modules), sortedModules(e.modules)) {
				t.Errorf("modules = %v, want %v", s.Modules, e.modules)
			}
			for _, m := range []domain.Module{domain.ModuleKerosene, domain.ModuleKeroseneLeak, domain.ModuleIntern, domain.ModuleWind, domain.ModuleIceBrakes} {
				if s.Has(m) != slices.Contains(e.modules, m) {
					t.Errorf("Has(%s) = %v", m, s.Has(m))
				}
			}
			tr, ok := domain.TrackFor(e.track)
			if !ok {
				t.Fatalf("track %s missing", e.track)
			}
			info := domain.ScenarioInfoFor(s)
			if info.ID != e.id || info.Abilities != e.abilities || info.Spaces != tr.Len() || info.Colour != e.colour ||
				info.RealTimeDropped != e.realTime || info.Modules == nil {
				t.Errorf("info = %+v", info)
			}
		})
	}
	if _, ok := domain.ScenarioFor("nope"); ok {
		t.Error("unknown scenario found")
	}
	if domain.DefaultScenario != "yul-green" {
		t.Errorf("DefaultScenario = %s, want yul-green (§4.1)", domain.DefaultScenario)
	}
}

func TestDataScenarioCatalogueInLobby(t *testing.T) {
	// §9 order in the lobby catalogue
	g, _ := domain.NewGame("ABCDE", "a", "Alice")
	v, err := g.ViewFor("a")
	if err != nil {
		t.Fatal(err)
	}
	exp := scenarioExpect()
	if len(v.Catalogue) != len(exp) {
		t.Fatalf("catalogue has %d, want %d", len(v.Catalogue), len(exp))
	}
	for i, e := range exp {
		if v.Catalogue[i].ID != e.id {
			t.Errorf("catalogue[%d] = %s, want %s", i, v.Catalogue[i].ID, e.id)
		}
	}
}

func startScenario(t *testing.T, id domain.ScenarioID) *domain.Game {
	t.Helper()
	s, _ := domain.ScenarioFor(id)
	g, err := domain.NewGame("TEST1", "a", "Alice")
	if err == nil {
		err = g.AddPlayer("b", "Bob")
	}
	if err == nil {
		err = g.Configure("a", id, domain.Abilities()[:s.Abilities])
	}
	if err == nil {
		err = g.Start("a", domain.ScriptedRNG(0, 0, 0, 0, 0, 0))
	}
	if err != nil {
		t.Fatalf("start %s: %v", id, err)
	}
	return g
}

func TestDataScenarioSetup(t *testing.T) {
	for _, e := range scenarioExpect() {
		t.Run(string(e.id)+" §4.1 §12", func(t *testing.T) {
			g := startScenario(t, e.id)
			f := g.Flight
			if g.Phase != domain.PhasePlaying || f.Round != 1 || f.Position != 1 || f.Axis != 0 || f.Gear != [3]bool{} ||
				f.Flaps != 0 || f.Brakes != 0 || f.Ice != 0 {
				t.Errorf("base setup wrong: %+v", f)
			}
			if f.R.Step != domain.StepDiscussion {
				t.Errorf("step = %s, want discussion", f.R.Step)
			}
			tr, _ := domain.TrackFor(e.track)
			if len(f.Planes) != tr.Len() {
				t.Errorf("planes len %d, want %d", len(f.Planes), tr.Len())
			}
			hasK := slices.Contains(e.modules, domain.ModuleKerosene) || slices.Contains(e.modules, domain.ModuleKeroseneLeak)
			if hasK && f.Fuel != 20 {
				t.Errorf("Fuel = %d, want 20 (§12.1)", f.Fuel)
			}
			if !hasK && f.Fuel != 0 {
				t.Errorf("Fuel = %d, want 0 without kerosene", f.Fuel)
			}
			if slices.Contains(e.modules, domain.ModuleWind) && f.Wind != 0 {
				t.Errorf("Wind offset = %d, want 0 (§12.3)", f.Wind)
			}
			if slices.Contains(e.modules, domain.ModuleIntern) {
				got := slices.Clone(f.Intern)
				sort.Ints(got)
				if !slices.Equal(got, []int{1, 2, 3, 4, 5, 6}) {
					t.Errorf("intern row = %v, want a permutation of 1..6 (§12.2)", f.Intern)
				}
			} else if len(f.Intern) != 0 {
				t.Errorf("intern row %v without the module", f.Intern)
			}
			v, err := g.ViewFor("a")
			if err != nil {
				t.Fatal(err)
			}
			fv := v.Flight
			if (fv.Fuel != nil) != hasK || fv.Leak != slices.Contains(e.modules, domain.ModuleKeroseneLeak) ||
				(fv.Wind != nil) != slices.Contains(e.modules, domain.ModuleWind) ||
				(fv.Ice != nil) != slices.Contains(e.modules, domain.ModuleIceBrakes) ||
				(fv.Intern != nil) != slices.Contains(e.modules, domain.ModuleIntern) {
				t.Errorf("view modules do not match scenario: fuel=%v leak=%v wind=%v ice=%v intern=%v", fv.Fuel, fv.Leak, fv.Wind, fv.Ice, fv.Intern)
			}
			if fv.Wind != nil && (fv.Wind.Offset != 0 || fv.Wind.Value != 3) {
				t.Errorf("wind view = %+v, want offset 0 value +3 (§12.3)", *fv.Wind)
			}
			if fv.AeroBlue != 4 || fv.AeroOrange != 8 {
				t.Errorf("aero = %d/%d, want 4/8 (§4.1)", fv.AeroBlue, fv.AeroOrange)
			}
			if fv.BrakeLimit != 2 {
				t.Errorf("BrakeLimit = %d, want 2 (§3.4/§12.6)", fv.BrakeLimit)
			}
		})
	}
}

func TestDataAbilityCountEnforced(t *testing.T) {
	// §13: exactly the scenario's number at Start, never more in Configure
	for _, e := range scenarioExpect() {
		t.Run(string(e.id)+" §13", func(t *testing.T) {
			g, _ := domain.NewGame("TEST1", "a", "Alice")
			_ = g.AddPlayer("b", "Bob")
			if err := g.Configure("a", e.id, domain.Abilities()[:e.abilities+1]); err == nil {
				t.Errorf("Configure with %d cards accepted, want ErrInvalidAbilities", e.abilities+1)
			}
			if e.abilities > 0 {
				if err := g.Configure("a", e.id, domain.Abilities()[:e.abilities-1]); err != nil {
					t.Fatalf("partial selection rejected: %v", err)
				}
				if err := g.Start("a", domain.ScriptedRNG(0, 0, 0, 0, 0)); err == nil {
					t.Error("Start with too few cards accepted")
				}
			}
		})
	}
}

// --- §10 approach tracks ---

type trackExp struct {
	id      domain.TrackID
	n       int
	planes  []int
	traffic []int
	turns   map[int]string // space number -> mask over axis -2..+2, o permitted / x forbidden
}

func trackExpect() []trackExp {
	return []trackExp{
		{"YUL-G", 7, []int{0, 0, 1, 2, 1, 3, 2}, nil, nil},
		{"LHR-G", 6, []int{0, 1, 1, 2, 2, 2}, []int{1, 0, 1, 0, 1, 0}, nil},
		{"LHR-Y", 6, []int{1, 1, 1, 2, 1, 2}, []int{2, 0, 1, 1, 1, 0}, nil},
		{"HND-G", 8, []int{0, 1, 1, 2, 1, 0, 2, 1}, []int{2, 0, 0, 0, 0, 0, 0, 0}, map[int]string{3: "xooxx", 5: "ooxxx", 6: "oooxx"}},
		{"HND-R", 8, []int{1, 0, 1, 1, 1, 2, 1, 1}, []int{3, 0, 0, 1, 1, 1, 0, 0}, map[int]string{2: "xooxx", 4: "ooxxx", 5: "xooxx", 7: "ooxxx"}},
		{"OSL-G", 8, []int{0, 0, 1, 0, 1, 1, 1, 0}, []int{2, 0, 0, 1, 0, 0, 0, 0}, nil},
		{"OSL-R", 8, []int{1, 0, 1, 0, 0, 1, 0, 0}, []int{3, 0, 1, 0, 1, 0, 1, 0}, nil},
		{"ATL-G", 8, []int{0, 0, 0, 0, 1, 2, 1, 2}, []int{4, 0, 1, 0, 1, 1, 0, 0}, nil},
		{"ATL-Y", 8, []int{0, 1, 0, 2, 2, 1, 3, 1}, []int{1, 1, 0, 2, 1, 0, 0, 0}, map[int]string{3: "xooox", 7: "xxxoo"}},
		{"PRG-G", 8, []int{0, 0, 1, 1, 0, 1, 1, 1}, []int{1, 0, 1, 0, 0, 1, 0, 0}, map[int]string{2: "xooxx", 4: "xxxoo"}},
		{"PRG-Y", 8, []int{0, 0, 1, 3, 0, 3, 2, 3}, []int{0, 0, 0, 1, 0, 1, 0, 0}, nil},
		{"TGU-Y", 5, []int{0, 1, 1, 1, 1}, []int{3, 0, 0, 0, 0}, map[int]string{2: "ooxxx", 3: "xooxx", 4: "ooxxx"}},
		{"TGU-R", 5, []int{0, 1, 1, 1, 2}, []int{3, 0, 2, 0, 0}, map[int]string{2: "ooxxx", 3: "xooxx", 4: "ooxxx"}},
		{"GIG-Y", 7, []int{0, 0, 2, 1, 2, 3, 1}, []int{2, 0, 0, 1, 1, 0, 0}, nil},
		{"GIG-R", 7, []int{0, 1, 2, 2, 1, 1, 2}, []int{3, 0, 1, 0, 1, 0, 0}, nil},
		{"KEF-Y", 6, []int{0, 0, 1, 1, 1, 0}, []int{2, 0, 1, 0, 1, 0}, nil},
		{"KEF-B", 6, []int{0, 0, 2, 1, 1, 0}, []int{2, 0, 1, 0, 0, 0}, map[int]string{2: "oooxx", 4: "xxooo", 5: "xooox"}},
		{"KUL-Y", 8, []int{0, 0, 1, 1, 0, 1, 1, 1}, []int{2, 0, 0, 1, 0, 0, 0, 0}, map[int]string{2: "xxxoo", 4: "xxoox", 5: "xxxoo", 7: "xxxoo"}},
		{"KUL-B", 8, []int{0, 0, 1, 0, 1, 1, 1, 1}, []int{3, 0, 1, 0, 0, 1, 0, 0}, map[int]string{2: "xxoox", 4: "oxxxx", 5: "xooxx", 6: "xxoox", 7: "xxxoo"}},
		{"PBH-R", 6, []int{0, 1, 1, 1, 1, 1}, []int{3, 0, 0, 1, 0, 0}, map[int]string{3: "ooxxx", 4: "oooxx", 5: "xxxoo"}},
		{"PBH-B", 6, []int{1, 0, 1, 1, 1, 1}, []int{3, 0, 1, 1, 1, 0}, map[int]string{3: "ooxxx", 4: "ooxxx", 5: "xxxxo"}},
		{"DUS-Y", 8, []int{0, 1, 1, 2, 1, 2, 1, 1}, []int{3, 0, 0, 0, 0, 2, 0, 0}, map[int]string{4: "xooox", 6: "xooox"}},
		{"DUS-B", 8, []int{1, 0, 0, 1, 2, 1, 3, 2}, []int{3, 0, 1, 1, 0, 0, 0, 0}, map[int]string{2: "xxoxx", 4: "ooxxx", 6: "xxxoo"}},
	}
}

func TestDataApproachTracks(t *testing.T) {
	exp := trackExpect()
	if len(exp) != 23 {
		t.Fatalf("test table has %d tracks", len(exp))
	}
	for _, e := range exp {
		t.Run(string(e.id)+" §10", func(t *testing.T) {
			tr, ok := domain.TrackFor(e.id)
			if !ok {
				t.Fatal("track missing")
			}
			if tr.ID != e.id || tr.Len() != e.n || len(tr.Spaces) != e.n {
				t.Fatalf("N = %d (len %d), want %d", tr.Len(), len(tr.Spaces), e.n)
			}
			total := 0
			for i, sp := range tr.Spaces {
				n := i + 1
				wantTraffic := 0
				if e.traffic != nil {
					wantTraffic = e.traffic[i]
				}
				if sp.Planes != e.planes[i] {
					t.Errorf("space %d planes = %d, want %d", n, sp.Planes, e.planes[i])
				}
				if sp.Traffic != wantTraffic {
					t.Errorf("space %d traffic = %d, want %d", n, sp.Traffic, wantTraffic)
				}
				total += sp.Planes
				mask, has := e.turns[n]
				if sp.Turn.Present() != has {
					t.Errorf("space %d turn present = %v, want %v", n, sp.Turn.Present(), has)
				}
				for axis := -2; axis <= 2; axis++ {
					want := true
					if has {
						want = mask[axis+2] == 'o'
					}
					if sp.Turn.Permits(axis) != want {
						t.Errorf("space %d turn permits(%d) = %v, want %v (mask %q)", n, axis, sp.Turn.Permits(axis), want, mask)
					}
				}
			}
			if total > domain.PlaneTokens {
				t.Errorf("setup planes %d exceed supply", total)
			}
			// §11.2: neither the start nor the airport has a turn icon
			if tr.Spaces[0].Turn.Present() || tr.Spaces[e.n-1].Turn.Present() {
				t.Error("start or airport has a turn icon")
			}
		})
	}
	if _, ok := domain.TrackFor("XXX-G"); ok {
		t.Error("unknown track found")
	}
	// §10: PRG-Y uses the whole supply
	tr, _ := domain.TrackFor("PRG-Y")
	sum := 0
	for _, s := range tr.Spaces {
		sum += s.Planes
	}
	if sum != 12 {
		t.Errorf("PRG-Y setup planes = %d, want 12", sum)
	}
}

func TestDataTurnMaskString(t *testing.T) {
	// "5-character form" of a turn icon; "" without an icon (o permitted, x forbidden as in §10)
	tr, _ := domain.TrackFor("HND-G")
	if got := tr.Spaces[2].Turn.String(); got != "xooxx" {
		t.Errorf("HND-G space 3 Turn.String() = %q, want xooxx", got)
	}
	if got := tr.Spaces[0].Turn.String(); got != "" {
		t.Errorf("no icon Turn.String() = %q, want empty", got)
	}
}
