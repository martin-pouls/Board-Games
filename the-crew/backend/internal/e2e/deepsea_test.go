package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-crew/backend/internal/adapters/gametest"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
)

const deepSea = domain.EditionDeepSea

// --- Deep Sea moves and leak checks (used by crewT) ---

// actorOf is the human who acts for seat s: the captain acts for the dummy.
func actorOf(a *domain.Attempt, s int) int {
	if s >= 0 && a.Members[s].Dummy {
		return a.Commander
	}
	return s
}

// deepSeaCandidates are the moves of the Deep Sea steps; false for the steps shared with Planet
// Nine.
func (c *crewT) deepSeaCandidates(v apidto.PlayerView) ([]apidto.ClientMessage, bool) {
	if c.edition != deepSea {
		return nil, false
	}
	at := v.Attempt
	n := len(at.Members)
	switch at.Step {
	case "draft":
		var out []apidto.ClientMessage
		if c.passFirst && at.Actions.Pass {
			out = append(out, apidto.ClientMessage{Type: apidto.CmdPassTask})
		}
		for _, id := range at.Takeable {
			out = append(out, apidto.ClientMessage{Type: apidto.CmdPickTask, TaskID: ptr(id)})
		}
		return append(out, apidto.ClientMessage{Type: apidto.CmdPassTask}), true
	case "assign":
		var out []apidto.ClientMessage
		for id, seats := range at.Assignable {
			for _, s := range seats {
				out = append(out, apidto.ClientMessage{Type: apidto.CmdAssignTask, TaskID: ptr(id), Seat: ptr(s)})
			}
		}
		return out, true
	case "predict":
		return []apidto.ClientMessage{{Type: apidto.CmdPredict, TaskID: ptr(at.Predict), Tricks: ptr(2)}}, true
	case "pick":
		if at.Pick != "volunteers" && at.Pick != "oneMember" {
			return nil, false
		}
		var out []apidto.ClientMessage
		for s := range n {
			out = append(out, apidto.ClientMessage{Type: apidto.CmdPick, Seats: []int{s}})
		}
		for s := range n {
			for u := range n {
				if s != u {
					out = append(out, apidto.ClientMessage{Type: apidto.CmdPick, Seats: []int{s, u}})
				}
			}
		}
		return out, true
	}
	return nil, false
}

// checkDeepSea checks one view against the server's game: the task cards, the hidden predictions,
// the face-down tasks of Q2 and the actions only the actor gets.
func (c *crewT) checkDeepSea(seat int, v apidto.PlayerView, g *domain.Game) {
	t := c.t
	t.Helper()
	a, at := g.Attempt, v.Attempt
	over := a.Result != nil
	if len(at.Tasks) != len(a.Tasks) {
		t.Fatalf("player %d sees %d tasks, server has %d", seat, len(at.Tasks), len(a.Tasks))
	}
	onTable := map[int]bool{}
	for _, st := range a.Tasks {
		onTable[st.CardID] = true
	}
	for i, st := range a.Tasks {
		w := at.Tasks[i]
		if w.Difficulty != st.Difficulty || w.Card != nil {
			t.Fatalf("player %d: task %d = %+v, server has difficulty %d and no Planet Nine card", seat, i, w, st.Difficulty)
		}
		if st.FaceDown && !over {
			// Only the difficulty of a face-down task is known.
			if w.CardID != 0 || w.Objective != nil || !w.FaceDown {
				t.Fatalf("player %d sees the face-down task %d: %+v", seat, i, w)
			}
			continue
		}
		if w.CardID != st.CardID || w.Objective == nil || !onTable[w.CardID] {
			t.Fatalf("player %d: task %d = %+v, server has card %d", seat, i, w, st.CardID)
		}
		if st.Objective.Kind != domain.KindPredict {
			continue
		}
		want := st.Objective.Count
		if st.Objective.Hidden && !over && (st.Owner() < 0 || actorOf(a, st.Owner()) != seat) {
			want = -1
		}
		if w.Objective.Count != want || w.Predicted != (st.Objective.Count >= 0) {
			t.Fatalf("player %d: prediction of task %d shows count %d (predicted %v), want %d; server has %d (hidden %v, revision %d)",
				seat, i, w.Objective.Count, w.Predicted, want, st.Objective.Count, st.Objective.Hidden, v.Revision)
		}
	}
	// Only the one who picks or assigns is offered tasks.
	if len(at.Takeable) > 0 && !slices.Contains(at.Actors, seat) {
		t.Fatalf("player %d is offered tasks %v but does not act", seat, at.Takeable)
	}
	for id, seats := range at.Assignable {
		if len(seats) > 0 && seat != a.Commander {
			t.Fatalf("player %d may assign task %d to %v but is no captain", seat, id, seats)
		}
	}
	if at.Actions.ReplaceTasks && seat != a.Commander {
		t.Fatalf("player %d may replace tasks but is no captain", seat)
	}
	// Nothing of the task deck or the used pile is on the wire.
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"taskDeck", "usedTasks", "deck"} {
		if strings.Contains(string(b), `"`+key+`"`) {
			t.Fatalf("player %d: the view carries %q", seat, key)
		}
	}
}

// checkDeepSeaEvents checks that a hidden prediction is announced without its number.
func (c *crewT) checkDeepSeaEvents(seat int, m wireMsg, g *domain.Game) {
	t := c.t
	t.Helper()
	for _, e := range m.Events {
		if e.Type != apidto.EventPredicted {
			continue
		}
		if e.TaskID == nil || e.Tricks == nil || e.Seat == nil {
			t.Fatalf("player %d: predicted event %+v is incomplete", seat, e)
		}
		if task := g.Attempt.Tasks[*e.TaskID]; task.Objective.Hidden && *e.Tricks != -1 {
			t.Fatalf("player %d: the hidden prediction of task %d was announced as %d", seat, *e.TaskID, *e.Tricks)
		}
	}
}

// --- helpers ---

// findDeepSeaSeed returns the first seed for which the freshly dealt game satisfies ok. The e2e
// server uses the same RNG sequence from the start of the mission on.
func findDeepSeaSeed(t *testing.T, n, mission int, ok func(*domain.Game) bool) uint64 {
	t.Helper()
	for seed := uint64(1); seed < 5000; seed++ {
		if ok(gametest.NewDeepSeaGame(t, n, mission, gametest.NewRNG(seed))) {
			return seed
		}
	}
	t.Fatalf("no seed for mission %d with %d players", mission, n)
	return 0
}

func cardIDs(v apidto.PlayerView) []int {
	var out []int
	for _, task := range v.Attempt.Tasks {
		out = append(out, task.CardID)
	}
	slices.Sort(out)
	return out
}

func hasKind(g *domain.Game, id int) bool {
	return slices.ContainsFunc(g.Attempt.Tasks, func(t domain.Task) bool { return t.CardID == id })
}

func event(c *crewT, k int, typ string) *apidto.Event {
	for _, m := range c.ps[k].ws.log {
		if e := hasEvent(m.Events, typ); e != nil {
			return e
		}
	}
	return nil
}

// --- tests ---

// A hidden prediction (task card 91) over the WebSocket: a draft with passing, the number only
// for its owner (the captain for the dummy), the restart in between, the reveal in the debrief.
func TestDeepSeaHiddenPredictionAfterDraftWithPass(t *testing.T) {
	for _, n := range []int{2, 3} {
		t.Run(map[int]string{2: "two players with Tonoja", 3: "three players"}[n], func(t *testing.T) {
			seed := findDeepSeaSeed(t, n, 5, func(g *domain.Game) bool { return hasKind(g, 91) && len(g.Attempt.Tasks) < 3 })
			c := newCrewIn(t, deepSea, n, seed)
			c.passFirst = true

			// The lobby: the catalogue is Deep Sea's, a Planet Nine mission number is out of range.
			var cat apidto.MissionCatalogue
			c.a.call(t, http.MethodGet, "/api/missions", "", nil, &cat, http.StatusOK)
			if cat.Edition != "deepSea" || len(cat.Missions) != 43 || cat.Missions[41].Label != "Q1" {
				t.Fatalf("catalogue = %s, %d missions", cat.Edition, len(cat.Missions))
			}
			c.reject(0, apidto.ClientMessage{Type: apidto.CmdSelectMission, Mission: ptr(44)}, "invalid_mission")
			c.reject(0, apidto.ClientMessage{Type: apidto.CmdPassTask}, "wrong_phase")
			c.run(0, apidto.ClientMessage{Type: apidto.CmdSelectMission, Mission: ptr(5)})
			if v := c.rest(1); v.Edition != "deepSea" || v.MissionDef == nil || v.MissionDef.Difficulty != 5 || v.MissionDef.Distribution != "draft" {
				t.Fatalf("lobby view = %+v", v)
			}
			c.run(0, apidto.ClientMessage{Type: apidto.CmdStartMission})

			// The draft starts with the captain.
			g := c.sync()
			cmd := g.Attempt.Commander
			v := c.rest(cmd)
			if v.Attempt.Step != "draft" || !slices.Contains(v.Attempt.Actors, cmd) || !v.Attempt.Actions.ReplaceTasks || !v.Attempt.Actions.Pass || v.Attempt.Stages != 0 || v.Attempt.Comm != "" {
				t.Fatalf("captain's view = %+v", v.Attempt)
			}
			for k := range c.ps {
				if k == cmd {
					continue
				}
				c.reject(k, apidto.ClientMessage{Type: apidto.CmdPassTask}, "not_your_turn")
				c.reject(k, apidto.ClientMessage{Type: apidto.CmdReplaceTask, TaskID: ptr(0)}, "not_commander")
				c.reject(k, apidto.ClientMessage{Type: apidto.CmdPredict, TaskID: ptr(0), Tricks: ptr(1)}, "wrong_step")
				c.reject(k, apidto.ClientMessage{Type: apidto.CmdAssignTask, TaskID: ptr(0), Seat: ptr(0)}, "wrong_step")
			}
			c.reject(cmd, apidto.ClientMessage{Type: apidto.CmdReplaceTask}, "bad_request")
			c.reject(cmd, apidto.ClientMessage{Type: apidto.CmdReplaceTask, TaskID: ptr(9)}, "invalid_task")

			// Draft (somebody passes), up to the prediction.
			if !c.until(func(g *domain.Game) bool { return g.Attempt.Step == domain.StepPredict }) {
				t.Fatal("the attempt ended before the prediction")
			}
			if event(c, 0, apidto.EventTaskPassed) == nil {
				t.Fatal("nobody passed")
			}
			g = c.sync()
			var id int
			for _, task := range g.Attempt.Tasks {
				if task.CardID == 91 {
					id = task.ID
				}
			}
			owner := g.Attempt.Tasks[id].Owner()
			actor := actorOf(g.Attempt, owner)
			if owner == g.Attempt.Commander {
				t.Fatal("the captain passed, so the captain is not the owner")
			}
			ov := c.rest(actor)
			if ov.Attempt.Step != "predict" || ov.Attempt.Predict != id || !slices.Equal(ov.Attempt.Actors, []int{actor}) || ov.Attempt.Tasks[id].Predicted {
				t.Fatalf("owner's view = %+v", ov.Attempt)
			}
			// Only the owner predicts, within 0..tricks, for the task asked for.
			for k := range c.ps {
				if k != actor {
					c.reject(k, apidto.ClientMessage{Type: apidto.CmdPredict, TaskID: ptr(id), Tricks: ptr(1)}, "not_your_turn")
				}
			}
			c.reject(actor, apidto.ClientMessage{Type: apidto.CmdPredict, TaskID: ptr(id)}, "bad_request")
			c.reject(actor, apidto.ClientMessage{Type: apidto.CmdPredict, TaskID: ptr(id), Tricks: ptr(-1)}, "invalid_prediction")
			c.reject(actor, apidto.ClientMessage{Type: apidto.CmdPredict, TaskID: ptr(id), Tricks: ptr(ov.Attempt.TotalTricks + 1)}, "invalid_prediction")
			c.reject(actor, apidto.ClientMessage{Type: apidto.CmdPredict, TaskID: ptr(1 - id), Tricks: ptr(1)}, "invalid_task")
			c.reject(actor, apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: &apidto.Card{Suit: "pink", Rank: 1}}, "wrong_step")
			c.run(actor, apidto.ClientMessage{Type: apidto.CmdPredict, TaskID: ptr(id), Tricks: ptr(0)}) // 0 is a number

			// The number is private to its owner; the announcement carries none.
			for k, p := range c.ps {
				task := p.ws.view.Attempt.Tasks[id]
				want := -1
				if k == actor {
					want = 0
				}
				if !task.Predicted || task.Objective.Count != want || !task.Objective.Hidden {
					t.Fatalf("player %d sees the prediction as %+v, want count %d", k, task, want)
				}
				if e := hasEvent(p.ws.last.Events, apidto.EventPredicted); e == nil || *e.Tricks != -1 || *e.TaskID != id || *e.Seat != owner {
					t.Fatalf("player %d: events = %+v", k, p.ws.last.Events)
				}
			}

			// The prediction survives a restart, still hidden.
			c.restart()
			for k := range c.ps {
				want := -1
				if k == actor {
					want = 0
				}
				if got := c.rest(k).Attempt.Tasks[id].Objective.Count; got != want {
					t.Fatalf("player %d after the restart: count %d, want %d", k, got, want)
				}
			}

			// Play on; in the debrief everybody sees the number.
			c.playToDebrief()
			for k := range c.ps {
				task := c.rest(k).Attempt.Tasks[id]
				if task.Objective.Count != 0 || !task.Predicted {
					t.Fatalf("player %d in the debrief: %+v", k, task)
				}
			}
		})
	}
}

// Free selection (mission 17): the captain replaces a task, assigns every task, the crew gives up
// and the host chooses between the same tasks, new tasks and an attempt that does not count.
func TestDeepSeaFreeSelectionReplaceAndNextAttemptChoices(t *testing.T) {
	c := newCrewIn(t, deepSea, 4, 11)
	c.run(0, apidto.ClientMessage{Type: apidto.CmdSelectMission, Mission: ptr(17)})
	c.run(0, apidto.ClientMessage{Type: apidto.CmdStartMission})
	g := c.sync()
	cmd := g.Attempt.Commander
	v := c.rest(cmd)
	if v.Attempt.Step != "assign" || !slices.Equal(v.Attempt.Actors, []int{cmd}) || v.MissionDef.Distribution != "free" || !v.Attempt.Actions.ReplaceTasks {
		t.Fatalf("captain's view: %+v", v.Attempt)
	}
	first := cardIDs(v)
	for k := range c.ps {
		if k == cmd {
			continue
		}
		c.reject(k, apidto.ClientMessage{Type: apidto.CmdAssignTask, TaskID: ptr(0), Seat: ptr(k)}, "not_commander")
		c.reject(k, apidto.ClientMessage{Type: apidto.CmdReplaceTask, TaskID: ptr(0)}, "not_commander")
	}
	c.reject(cmd, apidto.ClientMessage{Type: apidto.CmdAssignTask, TaskID: ptr(0)}, "bad_request")
	c.reject(cmd, apidto.ClientMessage{Type: apidto.CmdAssignTask, TaskID: ptr(0), Seat: ptr(7)}, "invalid_seat")
	c.reject(cmd, apidto.ClientMessage{Type: apidto.CmdAssignTask, TaskID: ptr(50), Seat: ptr(0)}, "invalid_task")
	c.reject(cmd, apidto.ClientMessage{Type: apidto.CmdPickTask, TaskID: ptr(0)}, "wrong_step")

	// Replace task 0: another card of the same difficulty, announced without it.
	before := v.Attempt.Tasks[0]
	c.run(cmd, apidto.ClientMessage{Type: apidto.CmdReplaceTask, TaskID: ptr(0)})
	for k, p := range c.ps {
		after := p.ws.view.Attempt.Tasks[0]
		if after.CardID == before.CardID || after.Difficulty != before.Difficulty {
			t.Fatalf("player %d: task 0 %+v -> %+v", k, before, after)
		}
		if e := hasEvent(p.ws.last.Events, apidto.EventTaskReplaced); e == nil || *e.TaskID != 0 {
			t.Fatalf("player %d: events = %+v", k, p.ws.last.Events)
		}
	}
	first = cardIDs(c.rest(cmd))
	if first[0] == 0 {
		t.Fatalf("cards = %v", first)
	}

	// Assign everything, then give up.
	if !c.until(func(g *domain.Game) bool {
		return g.Attempt.Step == domain.StepPlay || g.Attempt.Step == domain.StepPredict
	}) {
		t.Fatal("the attempt ended during the selection")
	}
	// Replacing is over once the selection has begun.
	c.reject(cmd, apidto.ClientMessage{Type: apidto.CmdReplaceTask, TaskID: ptr(0)}, "not_allowed")
	for c.load().Attempt.Step != domain.StepPlay {
		c.step()
	}
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt}, "wrong_phase")
	c.run(1, apidto.ClientMessage{Type: apidto.CmdGiveUp})
	attempts := func() int {
		e := logbookOf(c.rest(0), 17)
		if e == nil {
			t.Fatal("no logbook entry")
		}
		return e.Attempts
	}
	if attempts() != 1 {
		t.Fatalf("attempts = %d", attempts())
	}
	c.reject(1, apidto.ClientMessage{Type: apidto.CmdNextAttempt, NewTasks: true}, "not_host")
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdNextMission}, "invalid_result_choice")

	// The same tasks again.
	c.run(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt})
	if v := c.rest(2); v.Attempt.Number != 2 || !slices.Equal(cardIDs(v), first) {
		t.Fatalf("same tasks: attempt %d, cards %v, want %v", v.Attempt.Number, cardIDs(v), first)
	}
	for c.load().Attempt.Step != domain.StepPlay {
		c.step()
	}
	c.run(2, apidto.ClientMessage{Type: apidto.CmdGiveUp})
	if attempts() != 2 {
		t.Fatalf("attempts = %d", attempts())
	}

	// New tasks: the unused cards come first, so none of the last ones.
	c.run(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt, NewTasks: true})
	v = c.rest(3)
	if v.Attempt.Number != 3 || v.Attempt.Step != "assign" {
		t.Fatalf("new tasks: %+v", v.Attempt)
	}
	for _, id := range cardIDs(v) {
		if slices.Contains(first, id) {
			t.Fatalf("new tasks %v repeat a card of %v", cardIDs(v), first)
		}
	}
	for c.load().Attempt.Step != domain.StepPlay {
		c.step()
	}
	c.run(3, apidto.ClientMessage{Type: apidto.CmdGiveUp})
	if attempts() != 3 {
		t.Fatalf("attempts = %d", attempts())
	}

	// A failure that does not count: same attempt number, the logbook takes the loss back.
	c.run(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt, Uncounted: true})
	if v := c.rest(0); v.Attempt.Number != 3 || attempts() != 2 {
		t.Fatalf("uncounted: attempt %d, logbook %d", v.Attempt.Number, attempts())
	}
}

// Planet Nine knows no choice for the next attempt.
func TestPlanetNineRejectsDeepSeaChoices(t *testing.T) {
	c := newCrew(t, 2, 1)
	c.run(0, apidto.ClientMessage{Type: apidto.CmdStartMission})
	c.run(1, apidto.ClientMessage{Type: apidto.CmdGiveUp})
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt, NewTasks: true}, "not_allowed")
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt, Uncounted: true}, "not_allowed")
	for _, typ := range []string{apidto.CmdPassTask, apidto.CmdReplaceTask, apidto.CmdPredict, apidto.CmdAssignTask} {
		c.reject(0, apidto.ClientMessage{Type: typ, TaskID: ptr(0), Seat: ptr(0), Tricks: ptr(0)}, "wrong_phase")
	}
	c.run(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt})
	if v := c.rest(1); v.Attempt.Number != 2 {
		t.Fatalf("attempt = %d", v.Attempt.Number)
	}
}

// Rapture of the deep (mission 11): n-2 shared sonar tokens, one member may use several.
func TestDeepSeaRaptureCommunication(t *testing.T) {
	c := newCrewIn(t, deepSea, 4, 5)
	c.run(0, apidto.ClientMessage{Type: apidto.CmdSelectMission, Mission: ptr(11)})
	c.run(0, apidto.ClientMessage{Type: apidto.CmdStartMission})
	if !c.until(func(g *domain.Game) bool { return g.Attempt.Step == domain.StepPlay }) {
		t.Fatal("the attempt ended in the setup")
	}
	if a := c.rest(0).Attempt; a.Comm != "rapture" || a.TokensLeft != 2 {
		t.Fatalf("rapture: %q, %d tokens", a.Comm, a.TokensLeft)
	}
	user := -1
	for range 2 {
		k := -1
		for kk := range c.ps {
			if len(c.rest(kk).Attempt.Actions.Communicate) > 0 && (user < 0 || kk == user) {
				k = kk
				break
			}
		}
		if k < 0 {
			t.Fatal("nobody can communicate")
		}
		user = k
		o := c.rest(k).Attempt.Actions.Communicate[0]
		c.run(k, apidto.ClientMessage{Type: apidto.CmdCommunicate, Card: &o.Card, Position: o.Position})
	}
	for k, p := range c.ps {
		a := p.ws.view.Attempt
		if a.TokensLeft != 0 || len(a.Members[user].Communications) != 2 || a.Members[user].Communication == nil ||
			a.Members[user].Communication.Card != a.Members[user].Communications[0].Card || len(a.Actions.Communicate) != 0 {
			t.Fatalf("player %d after two communications: tokens %d, %+v", k, a.TokensLeft, a.Members[user])
		}
	}
	// The pool is empty: nobody communicates any more.
	for k := range c.ps {
		for _, h := range c.rest(k).Attempt.Hand {
			if h.Suit != "trump" {
				h := h
				c.reject(k, apidto.ClientMessage{Type: apidto.CmdCommunicate, Card: &h, Position: "highest"}, "cannot_communicate")
				break
			}
		}
	}
	c.playToDebrief()
}

// Side quest 2: the tasks stay face down for the crew until the debrief.
func TestDeepSeaFaceDownTasksStayHidden(t *testing.T) {
	c := newCrewIn(t, deepSea, 3, 3)
	c.run(0, apidto.ClientMessage{Type: apidto.CmdSelectMission, Mission: ptr(43)})
	c.run(0, apidto.ClientMessage{Type: apidto.CmdStartMission})
	v := c.rest(1)
	if v.Attempt.Stage != 0 || v.MissionDef.Label != "Q2" || !v.MissionDef.FaceDownTasks {
		t.Fatalf("Q2 view: %+v", v.MissionDef)
	}
	total := 0
	for _, task := range v.Attempt.Tasks {
		if !task.FaceDown || task.CardID != 0 || task.Objective != nil || task.Difficulty == 0 {
			t.Fatalf("task = %+v", task)
		}
		total += task.Difficulty
	}
	if total != 13 {
		t.Fatalf("difficulty of the face-down tasks = %d, want 13 for three players", total)
	}
	c.playToDebrief()
	for k := range c.ps {
		for _, task := range c.rest(k).Attempt.Tasks {
			if task.CardID == 0 || task.Objective == nil {
				t.Fatalf("player %d: task not revealed in the debrief: %+v", k, task)
			}
		}
	}
}

// A server serves the games of its own edition only: the other's are not found, over REST and
// over the WebSocket.
func TestEditionsDoNotSeeEachOthersGames(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "both.db")
	ds := startAppEdition(t, dbPath, deepSea, func() domain.RNG { return gametest.NewRNG(1) })
	dsGame := ds.create(t, "Ann")
	ds.stop()
	pn := startAppEdition(t, dbPath, domain.EditionPlanetNine, func() domain.RNG { return gametest.NewRNG(1) })
	pnGame := pn.create(t, "Pia")

	// Planet Nine does not find the Deep Sea game.
	pn.call(t, http.MethodGet, "/api/games/"+dsGame.GameID, "", nil, nil, http.StatusNotFound)
	pn.call(t, http.MethodGet, "/api/games/"+dsGame.GameID+"/view", dsGame.Token, nil, nil, http.StatusNotFound)
	pn.call(t, http.MethodPost, "/api/games/"+dsGame.GameID+"/players", "", apidto.JoinGameRequest{Name: "Bob"}, nil, http.StatusNotFound)
	pn.call(t, http.MethodPost, "/api/games/"+dsGame.GameID+"/leave", dsGame.Token, nil, nil, http.StatusNotFound)
	wsNotFound(t, pn, dsGame)
	pn.call(t, http.MethodGet, "/api/games/"+pnGame.GameID, "", nil, nil, http.StatusOK)
	var cat apidto.MissionCatalogue
	pn.call(t, http.MethodGet, "/api/missions", "", nil, &cat, http.StatusOK)
	if cat.Edition != "planetNine" || len(cat.Missions) != 50 {
		t.Fatalf("catalogue = %s %d", cat.Edition, len(cat.Missions))
	}
	pn.stop()

	// And Deep Sea does not find the Planet Nine game.
	ds = startAppEdition(t, dbPath, deepSea, func() domain.RNG { return gametest.NewRNG(1) })
	ds.call(t, http.MethodGet, "/api/games/"+pnGame.GameID, "", nil, nil, http.StatusNotFound)
	wsNotFound(t, ds, pnGame)
	ds.call(t, http.MethodGet, "/api/games/"+dsGame.GameID, "", nil, nil, http.StatusOK)
}

// wsNotFound dials the game and expects the socket to be closed with 4404 (game not found).
func wsNotFound(t *testing.T, a *app, s apidto.SessionResponse) {
	t.Helper()
	u := "ws" + strings.TrimPrefix(a.srv.URL, "http") + "/ws/games/" + s.GameID + "?token=" + s.Token
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.CloseNow()
	if _, _, err := c.Read(ctx); websocket.CloseStatus(err) != 4404 {
		t.Fatalf("close status = %d (err %v), want 4404", websocket.CloseStatus(err), err)
	}
}
