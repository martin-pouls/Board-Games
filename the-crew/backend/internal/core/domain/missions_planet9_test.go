package domain

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// This table was transcribed independently from docs/rules.md section 14 (and 2, 7, 10, 11, 13,
// 16) without looking at missions_planet9.go. TestPlanetNineMissionTable compares every field.

type expMission struct {
	n       int
	tasks   int
	tokens  []Token
	dist    Distribution
	dead    bool
	disrupt int
	golden  bool
	setup   []SetupAction
	objs    []ObjectiveSpec
}

const (
	a1 = TokenArrow1
	a2 = TokenArrow2
	a3 = TokenArrow3
	a4 = TokenArrow4
	om = TokenOmega
	t1 = Token1
	t2 = Token2
	t3 = Token3
	t4 = Token4
	t5 = Token5
)

func toks(t ...Token) []Token { return t }

func colourOnes() []Card {
	return []Card{{SuitPink, 1}, {SuitBlue, 1}, {SuitGreen, 1}, {SuitYellow, 1}}
}

func rocketCards() []Card {
	return []Card{{SuitTrump, 1}, {SuitTrump, 2}, {SuitTrump, 3}, {SuitTrump, 4}}
}

var expectedPlanetNine = []expMission{
	{n: 1, tasks: 1, dist: DistOpen},
	{n: 2, tasks: 2, dist: DistOpen},
	{n: 3, tasks: 2, tokens: toks(t1, t2), dist: DistOpen},
	{n: 4, tasks: 3, dist: DistOpen},
	{n: 5, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindWinNoTricks}, Owner: OwnerRuleChosen}}},
	{n: 6, tasks: 3, tokens: toks(a1, a2), dist: DistOpen, dead: true},
	{n: 7, tasks: 3, tokens: toks(om), dist: DistOpen},
	{n: 8, tasks: 3, tokens: toks(t1, t2, t3), dist: DistOpen},
	{n: 9, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindCrewWinWithAny, Cards: colourOnes(), Count: 1}, Owner: OwnerRuleCrew}}},
	{n: 10, tasks: 4, dist: DistOpen},
	{n: 11, tasks: 4, tokens: toks(t1), dist: DistOpen, setup: []SetupAction{SetupSilenceMember}},
	{n: 12, tasks: 4, tokens: toks(om), dist: DistOpen, setup: []SetupAction{SetupDrawAfterFirstTrick}},
	{n: 13, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindCrewWinWithEach, Cards: rocketCards()}, Owner: OwnerRuleCrew}}},
	{n: 14, tasks: 4, tokens: toks(a1, a2, a3), dist: DistOpen, dead: true},
	{n: 15, tasks: 4, tokens: toks(t1, t2, t3, t4), dist: DistOpen},
	{n: 16, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindCrewNeverWinWith, Rank: 9}, Owner: OwnerRuleCrew}}},
	{n: 17, tasks: 2, dist: DistOpen, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindCrewNeverWinWith, Rank: 9}, Owner: OwnerRuleCrew}}},
	{n: 18, tasks: 5, dist: DistOpen, disrupt: 2},
	{n: 19, tasks: 5, tokens: toks(t1), dist: DistOpen, disrupt: 3},
	{n: 20, tasks: 2, dist: DistCommanderDecision},
	{n: 21, tasks: 5, tokens: toks(t1, t2), dist: DistOpen, dead: true},
	{n: 22, tasks: 5, tokens: toks(a1, a2, a3, a4), dist: DistOpen},
	{n: 23, tasks: 5, tokens: toks(t1, t2, t3, t4, t5), dist: DistOpen, setup: []SetupAction{SetupSwapTokens}},
	{n: 24, tasks: 6, dist: DistCommanderDistribution},
	{n: 25, tasks: 6, tokens: toks(a1, a2), dist: DistOpen, dead: true, golden: true},
	{n: 26, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindCrewWinWithAny, Cards: colourOnes(), Count: 2}, Owner: OwnerRuleCrew}}},
	{n: 27, tasks: 3, dist: DistCommanderDecision, golden: true},
	{n: 28, tasks: 6, tokens: toks(t1, om), dist: DistOpen, disrupt: 3, golden: true},
	{n: 29, dead: true, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindCrewBalance, Count: 1}, Owner: OwnerRuleCrew}}},
	{n: 30, tasks: 6, tokens: toks(a1, a2, a3), dist: DistOpen, disrupt: 2, golden: true},
	{n: 31, tasks: 6, tokens: toks(t1, t2, t3), dist: DistOpen, golden: true},
	{n: 32, tasks: 7, dist: DistCommanderDistribution, golden: true},
	{n: 33, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindWinExactlyTricks, Count: 1, NotWithTrump: true}, Owner: OwnerRuleChosen}}},
	{n: 34, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindCrewBalance, Count: 1}, Owner: OwnerRuleCrew},
		{Objective: Objective{Kind: KindWinTricks, Tricks: []int{1, LastTrick}}, Owner: OwnerRuleCommander}}},
	{n: 35, tasks: 7, tokens: toks(a1, a2, a3), dist: DistOpen, golden: true},
	{n: 36, tasks: 7, tokens: toks(t1, t2), dist: DistCommanderDistribution, golden: true},
	{n: 37, tasks: 4, dist: DistCommanderDecision, golden: true},
	{n: 38, tasks: 8, dist: DistOpen, disrupt: 3, golden: true},
	{n: 39, tasks: 8, tokens: toks(a1, a2, a3), dist: DistOpen, dead: true, golden: true},
	{n: 40, tasks: 8, tokens: toks(t1, t2, t3), dist: DistOpen, golden: true, setup: []SetupAction{SetupMoveToken}},
	{n: 41, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindWinOnlyTricks, Tricks: []int{1, LastTrick}, NotWithTrump: true}, Owner: OwnerRuleChosen}}},
	{n: 42, tasks: 9, dist: DistOpen, golden: true},
	{n: 43, tasks: 9, dist: DistCommanderDistribution, golden: true},
	{n: 44, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindCrewWinWithEach, Cards: rocketCards(), Ordered: true}, Owner: OwnerRuleCrew}}},
	{n: 45, tasks: 9, tokens: toks(a1, a2, a3), dist: DistOpen, golden: true},
	{n: 46, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindWinAllOfSuit, Suit: SuitPink}, Owner: OwnerRuleLeftOfHolder, Holder: Card{SuitPink, 9}}}},
	{n: 47, tasks: 10, dist: DistOpen, golden: true},
	{n: 48, tasks: 3, tokens: toks(om), dist: DistOpen, golden: true, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindCardInLastTrick}, Owner: OwnerRuleCrew, TaskToken: TokenOmega}}},
	{n: 49, tasks: 10, tokens: toks(a1, a2, a3), dist: DistOpen, golden: true},
	{n: 50, objs: []ObjectiveSpec{
		{Objective: Objective{Kind: KindWinOnlyTricks, Tricks: []int{1, 2, 3, 4}}, Owner: OwnerRuleRoleA},
		{Objective: Objective{Kind: KindWinOnlyTricks, Tricks: []int{LastTrick}}, Owner: OwnerRuleRoleB}}},
}

// nilIfEmpty makes nil and empty slices compare equal under reflect.DeepEqual.
func nilIfEmpty[S ~[]E, E any](s S) S {
	if len(s) == 0 {
		return nil
	}
	return s
}

// diffMission returns one line per mismatching field.
func diffMission(want expMission, got MissionDef) []string {
	var out []string
	chk := func(field string, w, g any) {
		if !reflect.DeepEqual(w, g) {
			out = append(out, fmt.Sprintf("mission %d, %s: expected %v, got %v", want.n, field, w, g))
		}
	}
	chk("Number", want.n, got.Number)
	chk("Tasks", want.tasks, got.Tasks)
	chk("Tokens", nilIfEmpty(want.tokens), nilIfEmpty(got.Tokens))
	chk("Distribution", want.dist, got.Distribution)
	chk("Comm.DeadZone", want.dead, got.Comm.DeadZone)
	chk("Comm.FromTrick", want.disrupt, got.Comm.FromTrick)
	chk("Golden", want.golden, got.Golden)
	chk("Setup", nilIfEmpty(want.setup), nilIfEmpty(got.Setup))
	wantPoll := PollYesNo
	if want.n == 5 {
		wantPoll = PollWellIll
	}
	chk("Poll", wantPoll, got.Poll)
	if len(want.objs) != len(got.Objectives) {
		chk("len(Objectives)", len(want.objs), len(got.Objectives))
		return out
	}
	for i, w := range want.objs {
		g := got.Objectives[i]
		f := func(name string) string { return fmt.Sprintf("Objectives[%d].%s", i, name) }
		chk(f("Owner(rule)"), w.Owner, g.Owner)
		chk(f("Holder"), w.Holder, g.Holder)
		chk(f("TaskToken"), w.TaskToken, g.TaskToken)
		chk(f("Objective.Kind"), w.Objective.Kind, g.Objective.Kind)
		chk(f("Objective.Cards"), nilIfEmpty(w.Objective.Cards), nilIfEmpty(g.Objective.Cards))
		chk(f("Objective.Count"), w.Objective.Count, g.Objective.Count)
		chk(f("Objective.Tricks"), nilIfEmpty(w.Objective.Tricks), nilIfEmpty(g.Objective.Tricks))
		chk(f("Objective.Suit"), w.Objective.Suit, g.Objective.Suit)
		chk(f("Objective.Rank"), w.Objective.Rank, g.Objective.Rank)
		chk(f("Objective.NotWithTrump"), w.Objective.NotWithTrump, g.Objective.NotWithTrump)
		chk(f("Objective.Ordered"), w.Objective.Ordered, g.Objective.Ordered)
		// Objective.Owner of a template is overwritten at the deal; any of 0, OwnerCrew and
		// OwnerUnassigned is acceptable there, so it is not compared.
		switch g.Objective.Owner {
		case 0, OwnerCrew, OwnerUnassigned:
		default:
			out = append(out, fmt.Sprintf("mission %d, %s: odd template owner %d", want.n, f("Objective.Owner"), g.Objective.Owner))
		}
	}
	return out
}

func TestPlanetNineMissionTable(t *testing.T) {
	if len(expectedPlanetNine) != 50 {
		t.Fatalf("expectation table has %d rows, want 50", len(expectedPlanetNine))
	}
	var all []string
	for i, want := range expectedPlanetNine {
		if want.n != i+1 {
			t.Fatalf("expectation row %d has number %d", i+1, want.n)
		}
		got, ok := MissionFor(EditionPlanetNine, want.n)
		if !ok {
			all = append(all, fmt.Sprintf("mission %d: MissionFor returned !ok", want.n))
			continue
		}
		all = append(all, diffMission(want, got)...)
	}
	if len(all) > 0 {
		t.Errorf("%d mismatches:\n%s", len(all), strings.Join(all, "\n"))
	}
}

func TestPlanetNineMissionCount(t *testing.T) {
	if got := Missions(EditionPlanetNine); got != 50 {
		t.Errorf("Missions(PlanetNine) = %d, want 50", got)
	}
}

func TestMissionForBounds(t *testing.T) {
	for _, n := range []int{-1, 0, 51, 100} {
		if _, ok := MissionFor(EditionPlanetNine, n); ok {
			t.Errorf("MissionFor(PlanetNine, %d) succeeded", n)
		}
	}
	if _, ok := MissionFor(Edition("unknown"), 1); ok {
		t.Error("MissionFor(unknown edition, 1) succeeded")
	}
	if got := Missions(Edition("unknown")); got != 0 {
		t.Errorf("Missions(unknown edition) = %d, want 0", got)
	}
	for e, want := range map[Edition]bool{EditionPlanetNine: true, EditionDeepSea: true, "": false, "x": false} {
		if got := e.Valid(); got != want {
			t.Errorf("Edition(%q).Valid() = %v, want %v", e, got, want)
		}
	}
}

func TestPlanetNineMissionInvariants(t *testing.T) {
	for n := 1; n <= Missions(EditionPlanetNine); n++ {
		m, ok := MissionFor(EditionPlanetNine, n)
		if !ok {
			t.Fatalf("mission %d missing", n)
		}
		if m.Number != n {
			t.Errorf("mission %d: Number = %d", n, m.Number)
		}
		if len(m.Tokens) > m.Tasks {
			t.Errorf("mission %d: %d tokens for %d tasks", n, len(m.Tokens), m.Tasks)
		}
		if m.Golden && n < 25 {
			t.Errorf("mission %d: golden below 25", n)
		}
		if m.Distribution != DistOpen && m.Distribution != DistNone && m.Tasks == 0 {
			t.Errorf("mission %d: distribution %q without tasks", n, m.Distribution)
		}
		if m.Tasks == 0 && m.Distribution != DistNone {
			t.Errorf("mission %d: no tasks but distribution %q", n, m.Distribution)
		}
		if m.Tasks > 0 && m.Distribution == DistNone {
			t.Errorf("mission %d: tasks but no distribution", n)
		}
		for _, tok := range m.Tokens {
			if tok == TokenNone || !tok.Valid() {
				t.Errorf("mission %d: bad token %q", n, tok)
			}
		}
	}
}
