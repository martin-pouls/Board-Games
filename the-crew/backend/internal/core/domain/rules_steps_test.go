package domain

import "testing"

// rules §4 / §6: cards are played only once the setup steps are over.
func TestRulesPlayCardWrongStep(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mission int
		pass    bool
	}{
		{"draft", 1, false},
		{"poll", 5, false},
		{"pick", 11, false},
		{"token step", 23, false},
		{"pass", 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := withDistress(startMission(t, 4, tc.mission))
			if tc.pass {
				bbDraft(t, g)
				must(t, g.SendDistress(bbCmd(g), DirLeft))
			}
			card := g.Attempt.Members[g.Attempt.Commander].Hand[0]
			bbFailAny(t, g, func(g *Game) error { return g.PlayCard(bbCmd(g), card, noShuffle{}) }, ErrWrongStep)
		})
	}
}

// rules §4 / §6: with 3 players one card stays unplayed after the last trick; a task on it fails.
func TestRulesThreePlayersUnplayedCard(t *testing.T) {
	for _, tc := range []struct {
		name string
		task string
		ok   bool
		why  Reason
	}{
		{"task card left over", "p9:0", false, ReasonTaskFailed},
		{"task card played in the last trick", "b9:0", true, ReasonCompleted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := startMission(t, 3, 1)
			setTasks(g, tc.task)
			bbHistory(g, bbFillers(3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)...)
			a := g.Attempt
			a.Commander = 0
			a.Members[0].Hand = cs("b9 p9") // the 14-card player
			a.Members[1].Hand = cs("b1")
			a.Members[2].Hand = cs("b2")
			a.Trick = Trick{Leader: 0}
			must(t, g.PlayCard(alice, c("b9"), noShuffle{}))
			must(t, g.PlayCard(bob, c("b1"), noShuffle{}))
			must(t, g.PlayCard(carol, c("b2"), noShuffle{}))
			bbExpectEnd(t, g, tc.ok, tc.why, 13)
			if len(a.Members[0].Hand) != 1 {
				t.Fatalf("the unplayed card: %v", a.Members[0].Hand)
			}
		})
	}
}
