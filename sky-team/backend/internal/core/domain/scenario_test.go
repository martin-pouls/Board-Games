package domain

import (
	"slices"
	"testing"
)

func TestTurnMask(t *testing.T) {
	if m := turn(""); m.Present() || m.String() != "" || !m.Permits(-2) || !m.Permits(2) {
		t.Fatalf("no icon: %v", m)
	}
	m := turn("xooxx")
	if !m.Present() || m.String() != "xooxx" {
		t.Fatalf("mask %q", m.String())
	}
	for axis, want := range map[int]bool{-3: false, -2: false, -1: true, 0: true, 1: false, 2: false, 3: false} {
		if m.Permits(axis) != want {
			t.Errorf("xooxx permits %d = %v", axis, !want)
		}
	}
	if !turn("xxxxo").Permits(2) || turn("xxxxo").Permits(1) || !turn("oxxxx").Permits(-2) {
		t.Fatal("mask ends")
	}
	for _, bad := range []string{"xxxxx", "oooo", "ooooo!", "ooxxy"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("turn(%q) did not panic", bad)
				}
			}()
			turn(bad)
		}()
	}
}

func TestAltitudeTrack(t *testing.T) {
	// §4.2: first player alternates from the Pilot; rerolls on rounds 1 and 5 (green/yellow)
	// or round 1 only (red/black).
	for r := 1; r <= Rounds; r++ {
		a := AltitudeRow(r)
		wantFirst := RolePilot
		if r%2 == 0 {
			wantFirst = RoleCopilot
		}
		if a.Round != r || a.First != wantFirst || a.Altitude != (Rounds-r)*1000 {
			t.Errorf("round %d: %+v", r, a)
		}
		if a.Reroll(ColourGreen) != (r == 1 || r == 5) || a.Reroll(ColourYellow) != a.Reroll(ColourGreen) {
			t.Errorf("round %d: green-side reroll %v", r, a.Reroll(ColourGreen))
		}
		if a.Reroll(ColourRed) != (r == 1) || a.Reroll(ColourBlack) != a.Reroll(ColourRed) {
			t.Errorf("round %d: red-side reroll %v", r, a.Reroll(ColourRed))
		}
	}
}

func TestWindValue(t *testing.T) {
	// §12.3 by min(p, 20-p), and around the ring in both directions.
	want := []int{3, 3, 2, 2, 1, 0, -1, -2, -2, -3, -3, -3, -2, -2, -1, 0, 1, 2, 2, 3}
	for p, w := range want {
		if got := WindValue(p); got != w {
			t.Errorf("WindValue(%d) = %d, want %d", p, got, w)
		}
		if got := WindValue(p + 20); got != w {
			t.Errorf("WindValue(%d) = %d, want %d", p+20, got, w)
		}
		if got := WindValue(p - 40); got != w {
			t.Errorf("WindValue(%d) = %d, want %d", p-40, got, w)
		}
	}
	if WindValue(-1) != 3 || WindValue(-5) != 0 || WindValue(-10) != -3 {
		t.Fatal("counter-clockwise offsets")
	}
}

func TestSmallTables(t *testing.T) {
	if TrafficFaces() != [6]int{2, 3, 3, 4, 4, 5} {
		t.Fatalf("traffic faces %v", TrafficFaces())
	}
	want := []AbilityID{"adaptation", "anticipation", "control", "mastery", "synchronisation", "workingTogether"}
	if a := Abilities(); !slices.Equal(a, want) {
		t.Fatalf("abilities %v", a)
	}
	Abilities()[0] = "x"
	if Abilities()[0] != AbilityAdaptation {
		t.Fatal("Abilities must return a copy")
	}
}

func TestYULData(t *testing.T) {
	s, ok := ScenarioFor(DefaultScenario)
	if !ok || s.Airport != "YUL" || s.Colour != ColourGreen || s.Track != "YUL-G" || len(s.Modules) != 0 || s.Abilities != 0 || s.RealTimeDropped {
		t.Fatalf("yul-green %+v", s)
	}
	tr, ok := TrackFor("YUL-G")
	if !ok || tr.Len() != 7 {
		t.Fatalf("YUL-G %+v", tr)
	}
	var planes []int
	for _, sp := range tr.Spaces {
		planes = append(planes, sp.Planes)
		if sp.Traffic != 0 || sp.Turn.Present() {
			t.Fatalf("YUL-G has no traffic or turns: %+v", sp)
		}
	}
	if !slices.Equal(planes, []int{0, 0, 1, 2, 1, 3, 2}) {
		t.Fatalf("YUL-G planes %v", planes)
	}
	if _, ok := ScenarioFor("xyz"); ok {
		t.Fatal("unknown scenario found")
	}
	if _, ok := TrackFor("XYZ"); ok {
		t.Fatal("unknown track found")
	}
}

// TestDataConsistency checks every row of the data tables for shape, not content (the content
// is compared with independent expectation tables elsewhere).
func TestDataConsistency(t *testing.T) {
	ids := map[ScenarioID]bool{}
	for _, s := range Scenarios() {
		if ids[s.ID] {
			t.Errorf("duplicate scenario %s", s.ID)
		}
		ids[s.ID] = true
		if _, ok := TrackFor(s.Track); !ok {
			t.Errorf("%s: no track %s", s.ID, s.Track)
		}
		if s.Abilities < 0 || s.Abilities > 2 {
			t.Errorf("%s: %d abilities", s.ID, s.Abilities)
		}
		switch s.Colour {
		case ColourGreen, ColourYellow, ColourRed, ColourBlack:
		default:
			t.Errorf("%s: colour %q", s.ID, s.Colour)
		}
		for i, m := range s.Modules {
			switch m {
			case ModuleKerosene, ModuleKeroseneLeak, ModuleIntern, ModuleWind, ModuleIceBrakes:
			default:
				t.Errorf("%s: module %q", s.ID, m)
			}
			if slices.Contains(s.Modules[:i], m) {
				t.Errorf("%s: module %q twice", s.ID, m)
			}
		}
	}
	if Scenarios()[0].ID != DefaultScenario {
		t.Errorf("first scenario %s, want %s", Scenarios()[0].ID, DefaultScenario)
	}
	for _, tr := range tracks {
		n := tr.Len()
		if n < 2 {
			t.Errorf("%s: %d spaces", tr.ID, n)
			continue
		}
		total := 0
		for i, sp := range tr.Spaces {
			total += sp.Planes
			if sp.Planes < 0 || sp.Traffic < 0 {
				t.Errorf("%s space %d: %+v", tr.ID, i+1, sp)
			}
		}
		// §11.2: neither the start nor the airport has a turn icon.
		if tr.Spaces[0].Turn.Present() || tr.Spaces[n-1].Turn.Present() {
			t.Errorf("%s: turn icon on the start or the airport", tr.ID)
		}
		if total > PlaneTokens {
			t.Errorf("%s: %d plane tokens", tr.ID, total)
		}
	}
}

func TestLookupsReturnCopies(t *testing.T) {
	withScenario(t, testRed, testRedTrack)
	s, _ := ScenarioFor("tst-red")
	s.Modules[0] = ModuleWind
	Scenarios()[len(scenarios)-1].Modules[0] = ModuleWind
	tr, _ := TrackFor("TST-R")
	tr.Spaces[0].Planes = 9
	if s, _ := ScenarioFor("tst-red"); s.Modules[0] != ModuleKerosene {
		t.Fatal("scenario modules shared")
	}
	if tr, _ := TrackFor("TST-R"); tr.Spaces[0].Planes != 0 {
		t.Fatal("track spaces shared")
	}
}
