package domain

import "testing"

func TestDeepSeaTasksSanity(t *testing.T) {
	kinds := map[ObjectiveKind]bool{}
	for _, k := range []ObjectiveKind{
		KindWinCards, KindWinNoTricks, KindWinExactlyTricks, KindWinOnlyTricks, KindWinTricks,
		KindWinAllOfSuit, KindCrewWinWithAny, KindCrewWinWithEach, KindCrewNeverWinWith,
		KindCrewBalance, KindCardInLastTrick, KindCmpTricks, KindCmpCaptain, KindPredict,
		KindAvoidTricks, KindNoConsecutive, KindConsecutive, KindWinWith, KindWinTrickWhere,
		KindWinCount, KindWinEachColour, KindWinAllOfAnyColour, KindNeverLead, KindEqualSuits,
		KindMoreSuit, KindAll, KindLeaderStaysAhead, KindLastCardPlayed, KindTrickQuota, KindNone,
	} {
		kinds[k] = true
	}
	if len(deepSeaTasks) != 96 {
		t.Fatalf("got %d task cards, want 96", len(deepSeaTasks))
	}
	for i, c := range deepSeaTasks {
		if c.ID != i+1 {
			t.Errorf("index %d: ID %d, want %d", i, c.ID, i+1)
		}
		for _, d := range c.Difficulty {
			if d < 1 || d > 6 {
				t.Errorf("card %d: difficulty %v out of range", c.ID, c.Difficulty)
			}
		}
		if !kinds[c.Objective.Kind] {
			t.Errorf("card %d: unknown kind %q", c.ID, c.Objective.Kind)
		}
		for _, p := range c.Objective.Parts {
			if !kinds[p.Kind] {
				t.Errorf("card %d: unknown part kind %q", c.ID, p.Kind)
			}
		}
	}
	for _, id := range []int{0, 97} {
		if _, ok := TaskCardFor(id); ok {
			t.Errorf("TaskCardFor(%d) succeeded", id)
		}
	}
	if c, ok := TaskCardFor(96); !ok || c.ID != 96 {
		t.Errorf("TaskCardFor(96) = %d, %v", c.ID, ok)
	}
}
