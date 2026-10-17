package domain

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// This table was transcribed independently from deep-sea/rules.md section 8 (and the conventions
// agreed for the mission table) without looking at missions_deepsea.go.
// TestDeepSeaMissionTable compares every field of every mission.

const (
	dsQ1 = 42
	dsQ2 = 43
)

type expDeepSeaMission struct {
	n          int
	diff       int // D; 0 = no task cards
	dist       Distribution
	mode       CommMode
	fromTrick  int
	volunteers int
	hardest    bool
	skipped    bool
	uncounted  bool
	fixed      []int
	objs       []ObjectiveSpec
}

func dsCrew(o Objective) []ObjectiveSpec {
	return []ObjectiveSpec{{Objective: o, Owner: OwnerRuleCrew}}
}

// expectedDeepSeaMissions lists missions 1..41; Q1 and Q2 are built in expectedDeepSeaMission.
var expectedDeepSeaMissions = func() []expDeepSeaMission {
	m := []expDeepSeaMission{
		{n: 1, diff: 1, dist: DistDraft},
		{n: 2, diff: 2, dist: DistDraft},
		{n: 3, diff: 3, dist: DistDraft},
		{n: 4, diff: 4, dist: DistDraft},
		{n: 5, diff: 5, dist: DistDraft},
		{n: 6, diff: 5, dist: DistOneMember},
		{n: 7, diff: 6, dist: DistDraft},
		{n: 8, objs: dsCrew(Objective{Kind: KindCrewBalance, Count: 1, Rank: 9})},
		{n: 9, diff: 7, dist: DistDraft, mode: CommCurrents},
		{n: 10, diff: 4, dist: DistCaptainOrVolunteer},
		{n: 11, diff: 8, dist: DistDraft, mode: CommRapture},
		{n: 12, objs: dsCrew(Objective{Kind: KindNeverLead, Suits: []Suit{SuitPink, SuitTrump}})},
		{n: 13, diff: 5, dist: DistCaptainOrVolunteer},
		{n: 14, diff: 6, dist: DistVolunteers, volunteers: 1, mode: CommCurrents},
		{n: 15, diff: 6, dist: DistVolunteers, volunteers: 1, mode: CommRapture},
		{n: 16, diff: 6, dist: DistVolunteers, volunteers: 1, mode: CommNone},
		{n: 17, diff: 9, dist: DistFree},
		{n: 18, diff: 9, dist: DistDraft},
		{n: 19, diff: 9, dist: DistDraft, hardest: true},
		{n: 20, diff: 10, dist: DistDraft, mode: CommUnfamiliar},
		{n: 21, mode: CommUnfamiliar, objs: dsCrew(Objective{Kind: KindCrewBalance, Count: 1, Rank: 1})},
		{n: 22, diff: 11, dist: DistDraft, mode: CommUnfamiliar},
		{n: 23, mode: CommUnfamiliar, fromTrick: 2, objs: dsCrew(Objective{Kind: KindLeaderStaysAhead})},
		{n: 24, diff: 12, dist: DistDraft, mode: CommUnfamiliar},
		{n: 25, diff: 12, dist: DistDraft, mode: CommUnfamiliar, skipped: true},
		{n: 26, diff: 12, dist: DistVolunteers, volunteers: 2, mode: CommUnfamiliar},
		{n: 27, diff: 5, dist: DistDraft, mode: CommUnfamiliar,
			objs: dsCrew(Objective{Kind: KindLastCardPlayed, Cards: []Card{{SuitYellow, 5}}})},
		{n: 28, diff: 14, dist: DistFree},
		{n: 29, diff: 15, dist: DistFree},
		{n: 30, diff: 16, dist: DistFree},
		{n: 31, diff: 17, dist: DistFree},
		{n: 32, dist: DistDraft, fixed: []int{74, 88, 85, 80}},
	}
	// Epilogue 33..41: D 18..26.
	for n := 33; n <= 41; n++ {
		m = append(m, expDeepSeaMission{n: n, diff: n - 15, dist: DistFree, uncounted: true})
	}
	return m
}()

// expectedDeepSeaMission builds the full expected MissionDef of mission n (1..43).
func expectedDeepSeaMission(n int) MissionDef {
	switch n {
	case dsQ1:
		stage := func(count int) []ObjectiveSpec {
			return []ObjectiveSpec{{
				Owner:     OwnerRuleCommander,
				Objective: Objective{Kind: KindWinExactlyTricks, Count: count, Without: []Rank{7, 8, 9}},
			}}
		}
		return MissionDef{
			Number: dsQ1, Label: "Q1", Poll: PollYesNo,
			CaptainCard: Card{SuitTrump, 1},
			Redeal:      [][]Card{{dsSub(1), dsSub(2), dsSub(3), dsSub(4)}},
			Stages:      [][]ObjectiveSpec{stage(2), stage(3)},
		}
	case dsQ2:
		return MissionDef{
			Number: dsQ2, Label: "Q2", Poll: PollYesNo,
			Distribution: DistDraft, DifficultyBy: [3]int{13, 10, 8}, FaceDownTasks: true,
			Objectives: dsCrew(Objective{Kind: KindTrickQuota}),
		}
	}
	e := expectedDeepSeaMissions[n-1]
	return MissionDef{
		Number: n, Poll: PollYesNo,
		Difficulty: e.diff, Distribution: e.dist,
		Comm:       CommRule{Mode: e.mode, FromTrick: e.fromTrick},
		Volunteers: e.volunteers, HardestFirst: e.hardest, CaptainSkipped: e.skipped,
		Uncounted: e.uncounted, FixedTasks: e.fixed, Objectives: e.objs,
	}
}

func TestDeepSeaMissionTable(t *testing.T) {
	if len(expectedDeepSeaMissions) != 41 {
		t.Fatalf("expectation table has %d rows, want 41", len(expectedDeepSeaMissions))
	}
	for i, e := range expectedDeepSeaMissions {
		if e.n != i+1 {
			t.Fatalf("expectation row %d has number %d", i+1, e.n)
		}
	}
	var all []string
	for n := 1; n <= 43; n++ {
		want := expectedDeepSeaMission(n)
		got, ok := MissionFor(EditionDeepSea, n)
		if !ok {
			all = append(all, fmt.Sprintf("mission %d: MissionFor returned !ok", n))
			continue
		}
		for _, d := range dsDiff("", reflect.ValueOf(want), reflect.ValueOf(got)) {
			all = append(all, fmt.Sprintf("mission %d, field %s", n, strings.TrimPrefix(d, ".")))
		}
	}
	if len(all) > 0 {
		t.Errorf("%d mismatches:\n%s", len(all), strings.Join(all, "\n"))
	}
}

func TestDeepSeaMissionCount(t *testing.T) {
	if got := Missions(EditionDeepSea); got != 43 {
		t.Errorf("Missions(DeepSea) = %d, want 43", got)
	}
	for _, n := range []int{-1, 0, 44, 100} {
		if _, ok := MissionFor(EditionDeepSea, n); ok {
			t.Errorf("MissionFor(DeepSea, %d) succeeded", n)
		}
	}
}
