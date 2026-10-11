package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-crew/backend/internal/adapters/gametest"
	"github.com/martin-pouls/the-crew/backend/internal/adapters/sqlite"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
)

// crewT plays a crew of humans over their WebSockets against a sqlite-backed server, always
// answering with the first legal option, and checks on every revision that no view and no event
// leaks a hidden card.
//
// The RNG is one seeded generator shared by every service call, so a run is reproducible: the
// seeds below were found by playing the same strategy on the domain (mission 1, first legal
// card): seed 1 with two players wins the first attempt; seed 2 with four players loses the
// first attempt and wins the second.
type crewT struct {
	t       *testing.T
	a       *app
	dbPath  string
	rng     domain.RNG
	code    string
	ps      []*player
	snaps   map[int64]*domain.Game // the server's game at each revision seen after a step
	checked []int                  // per player: how many logged views have been checked
	// distress makes the commander send the distress signal (to the left) when he may.
	distress bool
}

func newCrew(t *testing.T, n int, seed uint64) *crewT {
	t.Helper()
	c := &crewT{t: t, dbPath: filepath.Join(t.TempDir(), "crew.db"), rng: gametest.NewRNG(seed), snaps: map[int64]*domain.Game{}}
	c.a = startAppRNG(t, c.dbPath, func() domain.RNG { return c.rng })
	host := c.a.create(t, "Ann")
	c.code = host.GameID
	c.ps = []*player{{sess: host}}
	for _, name := range []string{"Ben", "Cleo", "Dan", "Eve"}[:n-1] {
		c.ps = append(c.ps, &player{sess: c.a.join(t, c.code, name)})
	}
	c.connect()
	return c
}

// connect (re)dials every player.
func (c *crewT) connect() {
	c.t.Helper()
	c.checked = make([]int, len(c.ps))
	for i, p := range c.ps {
		p.ws = c.a.dial(c.t, fmt.Sprintf("player%d", i), c.code, p.sess.Token)
	}
	for _, p := range c.ps {
		p.ws.presence(len(c.ps))
	}
}

func (c *crewT) host() *client { return c.ps[0].ws }

func (c *crewT) load() *domain.Game {
	c.t.Helper()
	rec, err := sqlite.NewGameRepository(c.a.db).Load(context.Background(), domain.GameID(c.code))
	if err != nil {
		c.t.Fatal(err)
	}
	return rec.Game
}

// sync stores the server's game, waits until every player has received that revision and checks
// every view they received since the last sync.
func (c *crewT) sync() *domain.Game {
	c.t.Helper()
	g := c.load()
	c.snaps[g.Revision] = g
	for k, p := range c.ps {
		p.ws.syncTo(g.Revision)
		for ; c.checked[k] < len(p.ws.log); c.checked[k]++ {
			m := p.ws.log[c.checked[k]]
			if snap := c.snaps[m.View.Revision]; snap != nil {
				c.checkMessage(k, m, snap)
			}
		}
	}
	return g
}

// run sends cmd from player k, expects it to be accepted and syncs.
func (c *crewT) run(k int, cmd apidto.ClientMessage) *domain.Game {
	c.t.Helper()
	c.ps[k].ws.send(cmd)
	return c.sync()
}

// reject sends cmd from player k, expects the error code and that nothing changed.
func (c *crewT) reject(k int, cmd apidto.ClientMessage, code string) {
	c.t.Helper()
	before := c.load().Revision
	if e := c.ps[k].ws.sendErr(cmd); e.Code != code {
		c.t.Fatalf("%s from player %d: code %q, want %q (%s)", cmd.Type, k, e.Code, code, e.Message)
	}
	if after := c.load().Revision; after != before {
		c.t.Fatalf("%s changed the game: revision %d -> %d", cmd.Type, before, after)
	}
}

func (c *crewT) rest(k int) apidto.PlayerView {
	return c.a.view(c.t, c.code, c.ps[k].sess.Token)
}

// --- leak checks ---

func cardKey(c apidto.Card) string { return fmt.Sprintf("%s %d", c.Suit, c.Rank) }

func wireCard(c domain.Card) apidto.Card { return apidto.FromCard(c) }

// forbidden returns the cards that seat must not see in a running attempt: every other human's
// hand (except a communicated card, which is on the table) and the dummy's hidden cards.
func forbidden(g *domain.Game, seat int) map[string]bool {
	out := map[string]bool{}
	a := g.Attempt
	for o := range a.Members {
		m := &a.Members[o]
		switch {
		case m.Dummy:
			for _, s := range m.Stacks {
				if s.Down != nil {
					out[cardKey(wireCard(*s.Down))] = true
				}
			}
		case o != seat:
			public := a.CommunicatedCard(o)
			for _, h := range m.Hand {
				if public == nil || public.Card != h {
					out[cardKey(wireCard(h))] = true
				}
			}
		}
	}
	return out
}

// checkView checks one view of player seat against the server's game g at the same revision.
func (c *crewT) checkView(seat int, v apidto.PlayerView, g *domain.Game) {
	t := c.t
	t.Helper()
	if v.Attempt == nil || g.Attempt == nil {
		return
	}
	a, at := g.Attempt, v.Attempt
	if a.Result != nil {
		// Everything is revealed, and the reveal is the truth.
		for i, m := range at.Members {
			want := a.Members[i].Cards()
			got := make([]domain.Card, len(m.Cards))
			for j, w := range m.Cards {
				got[j] = domain.Card{Suit: domain.Suit(w.Suit), Rank: domain.Rank(w.Rank)}
			}
			domain.SortCards(got)
			if !slices.Equal(want, got) {
				t.Fatalf("player %d: reveal of seat %d = %v, server has %v", seat, i, got, want)
			}
		}
		return
	}
	for i, m := range at.Members {
		if m.Cards != nil {
			t.Fatalf("player %d sees the cards of seat %d before the debrief", seat, i)
		}
	}
	// The viewer's own hand is exactly the server's.
	own := make([]domain.Card, len(at.Hand))
	for j, w := range at.Hand {
		own[j] = domain.Card{Suit: domain.Suit(w.Suit), Rank: domain.Rank(w.Rank)}
	}
	if !slices.Equal(own, a.Members[seat].Hand) {
		t.Fatalf("player %d: hand %v, server has %v", seat, own, a.Members[seat].Hand)
	}
	bad := forbidden(g, seat)
	see := func(where string, cs ...apidto.Card) {
		for _, c := range cs {
			if bad[cardKey(c)] {
				t.Fatalf("player %d sees %s in %s (revision %d)", seat, cardKey(c), where, v.Revision)
			}
		}
	}
	see("hand", at.Hand...)
	see("playable", at.Playable...)
	for _, m := range at.Members {
		for _, s := range m.Stacks {
			if s.Up != nil {
				see("a dummy stack", *s.Up)
			}
		}
		if m.Communication != nil {
			see("a communication", m.Communication.Card)
		}
	}
	if at.Pass != nil && at.Pass.Chosen != nil {
		see("the pass", *at.Pass.Chosen)
	}
	// The exchange shows only the viewer's own side.
	if ex := at.Exchange; ex != nil {
		if a.Exchange == nil {
			t.Fatalf("player %d: exchange without a server exchange", seat)
		}
		if (ex.Gave == nil) != (a.Exchange.Gave[seat] == nil) || (ex.Got == nil) != (a.Exchange.Got[seat] == nil) {
			t.Fatalf("player %d: exchange %+v, server has gave %v got %v", seat, ex, a.Exchange.Gave[seat], a.Exchange.Got[seat])
		}
		if ex.Gave != nil && *ex.Gave != wireCard(*a.Exchange.Gave[seat]) || ex.Got != nil && *ex.Got != wireCard(*a.Exchange.Got[seat]) {
			t.Fatalf("player %d: exchange %+v differs from the server's", seat, ex)
		}
	}
	// The dummy's hidden cards are only counted.
	for i, m := range at.Members {
		if m.Dummy && (len(m.Stacks) != domain.DummyStacks || m.CardCount != a.Members[i].CardCount()) {
			t.Fatalf("dummy member = %+v", m)
		}
	}
}

// checkMessage checks a received view message and its events.
func (c *crewT) checkMessage(seat int, m wireMsg, g *domain.Game) {
	t := c.t
	t.Helper()
	c.checkView(seat, *m.View, g)
	if g.Attempt == nil || g.Attempt.Result != nil {
		return
	}
	raw, err := json.Marshal(m.Events)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Events {
		if e.Card != nil && e.Type != apidto.EventCardPlayed && e.Type != apidto.EventCommunicated {
			t.Fatalf("player %d: event %s carries a card", seat, e.Type)
		}
	}
	for key := range forbidden(g, seat) {
		var suit string
		var rank int
		_, _ = fmt.Sscanf(key, "%s %d", &suit, &rank)
		if strings.Contains(string(raw), fmt.Sprintf(`"suit":%q,"rank":%d`, suit, rank)) {
			t.Fatalf("player %d: events leak %s: %s", seat, key, raw)
		}
	}
}

// --- playing ---

// actor returns the player who must act now and their view, or -1 when nobody does.
func (c *crewT) actor(g *domain.Game) (int, apidto.PlayerView) {
	c.t.Helper()
	found := -1
	var fv apidto.PlayerView
	for k := range c.ps {
		v := c.rest(k)
		if v.Revision != g.Revision {
			c.t.Fatalf("REST view revision %d, server %d", v.Revision, g.Revision)
		}
		c.checkView(k, v, g)
		if found < 0 && v.Phase == "playing" && slices.Contains(v.Attempt.Actors, v.Attempt.Seat) {
			found, fv = k, v
		}
	}
	return found, fv
}

// candidates are the commands the actor tries, in order, for the step of their view.
func (c *crewT) candidates(v apidto.PlayerView) []apidto.ClientMessage {
	at := v.Attempt
	n := len(c.ps)
	msg := func(typ string) apidto.ClientMessage { return apidto.ClientMessage{Type: typ} }
	switch at.Step {
	case "adjustTokens":
		return []apidto.ClientMessage{msg(apidto.CmdSkipTokens)}
	case "poll":
		yes := true
		return []apidto.ClientMessage{{Type: apidto.CmdAnswerPoll, Yes: &yes}}
	case "pick":
		var out []apidto.ClientMessage
		if at.Pick == "roles" {
			for s := range n {
				for u := range n {
					if s != u {
						out = append(out, apidto.ClientMessage{Type: apidto.CmdPick, Seats: []int{s, u}})
					}
				}
			}
			return out
		}
		for s := range n {
			out = append(out, apidto.ClientMessage{Type: apidto.CmdPick, Seats: []int{s}})
		}
		return out
	case "draft":
		for _, task := range at.Tasks {
			if task.Owner == -2 {
				id := task.ID
				return []apidto.ClientMessage{{Type: apidto.CmdPickTask, TaskID: &id}}
			}
		}
	case "pass":
		for _, h := range at.Hand {
			if h.Suit != "trump" {
				h := h
				return []apidto.ClientMessage{{Type: apidto.CmdChoosePassCard, Card: &h}}
			}
		}
	case "play":
		if at.Actions.SendDistress {
			if c.distress {
				return []apidto.ClientMessage{{Type: apidto.CmdSendDistress, Direction: "left"}}
			}
			return []apidto.ClientMessage{msg(apidto.CmdSkipDistress)}
		}
		if len(at.Playable) > 0 {
			card := at.Playable[0]
			return []apidto.ClientMessage{{Type: apidto.CmdPlayCard, Card: &card}}
		}
	}
	c.t.Fatalf("no move for step %q: %+v", at.Step, at)
	return nil
}

// step lets the player who must act make the first accepted move. It reports false when the
// attempt is not running.
func (c *crewT) step() bool {
	c.t.Helper()
	g := c.sync()
	if g.Phase != domain.PhasePlaying {
		return false
	}
	k, v := c.actor(g)
	if k < 0 {
		c.t.Fatalf("nobody must act: %+v", v)
	}
	for _, cmd := range c.candidates(v) {
		if e := c.ps[k].ws.try(cmd); e == nil {
			c.sync()
			return true
		}
	}
	c.t.Fatalf("player %d: every candidate was rejected in step %q", k, v.Attempt.Step)
	return false
}

// until steps until done holds for the server's game; false if the attempt ended first.
func (c *crewT) until(done func(*domain.Game) bool) bool {
	c.t.Helper()
	for range 2000 {
		if done(c.load()) {
			return true
		}
		if !c.step() {
			return false
		}
	}
	c.t.Fatal("never reached the state")
	return false
}

func (c *crewT) playToDebrief() *apidto.AttemptResult {
	c.t.Helper()
	for c.step() {
	}
	g := c.sync()
	if g.Phase != domain.PhaseDebrief {
		c.t.Fatalf("phase = %s, want debrief", g.Phase)
	}
	// Every player's view carries the result.
	var res *apidto.AttemptResult
	for k := range c.ps {
		v := c.rest(k)
		if v.Phase != "debrief" || v.Attempt == nil || v.Attempt.Result == nil || v.Attempt.Step != "" {
			c.t.Fatalf("player %d debrief view = %+v", k, v)
		}
		if res != nil && !reflect.DeepEqual(res, v.Attempt.Result) {
			c.t.Fatalf("players disagree on the result")
		}
		res = v.Attempt.Result
		c.checkView(k, v, g)
	}
	return res
}

func (c *crewT) restart() {
	c.t.Helper()
	before := make([]apidto.PlayerView, len(c.ps))
	for k := range c.ps {
		before[k] = c.rest(k)
	}
	c.a.stop()
	c.a = startAppRNG(c.t, c.dbPath, func() domain.RNG { return c.rng })
	for k := range c.ps {
		if after := c.rest(k); !reflect.DeepEqual(before[k], after) {
			c.t.Fatalf("player %d: view changed across the restart:\n before %+v\n after  %+v", k, before[k], after)
		}
	}
	c.connect()
	c.sync()
}

func logbookOf(v apidto.PlayerView, mission int) *apidto.LogEntry {
	for i := range v.Logbook {
		if v.Logbook[i].Mission == mission {
			return &v.Logbook[i]
		}
	}
	return nil
}

func TestMissionCatalogueOverHTTP(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "m.db"), 1)
	var c apidto.MissionCatalogue
	a.call(t, http.MethodGet, "/api/missions", "", nil, &c, http.StatusOK)
	if c.Edition != "planetNine" || len(c.Missions) != 50 || c.Missions[23].Number != 24 || c.Missions[23].Distribution != "commanderDistribution" {
		t.Fatalf("catalogue: %s, %d missions", c.Edition, len(c.Missions))
	}
}

// Two players with JARVIS: mission 1 played to a win over the WebSocket, a restart in the middle
// of the attempt, then the debrief, the way back to the lobby and a replay given up.
func TestTwoPlayersWithJarvisWinMission1(t *testing.T) {
	c := newCrew(t, 2, 1)

	// Lobby commands: only the host, only valid missions.
	c.reject(1, apidto.ClientMessage{Type: apidto.CmdStartMission}, "not_host")
	c.reject(1, apidto.ClientMessage{Type: apidto.CmdSelectMission, Mission: ptr(3)}, "not_host")
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdSelectMission}, "bad_request")
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdSelectMission, Mission: ptr(51)}, "invalid_mission")
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: &apidto.Card{Suit: "pink", Rank: 1}}, "wrong_phase")
	c.run(0, apidto.ClientMessage{Type: apidto.CmdSelectMission, Mission: ptr(3)})
	m := c.ps[1].ws.last // run synced the guest to the new revision
	if e := hasEvent(m.Events, apidto.EventMissionSelected); e == nil || *e.Mission != 3 || m.View.Mission != 3 ||
		m.View.MissionDef == nil || m.View.MissionDef.Number != 3 {
		t.Fatalf("guest after selectMission: %+v", m.View)
	}
	c.run(0, apidto.ClientMessage{Type: apidto.CmdSelectMission, Mission: ptr(1)})

	c.run(0, apidto.ClientMessage{Type: apidto.CmdStartMission})
	for k := range c.ps {
		v := c.rest(k)
		if v.Phase != "playing" || v.Attempt == nil || len(v.Attempt.Members) != 3 || !v.Attempt.Members[2].Dummy ||
			v.Attempt.Number != 1 || v.Attempt.TotalTricks != 13 || len(v.Attempt.Hand) != 13 {
			t.Fatalf("player %d view after start: %+v", k, v.Attempt)
		}
	}

	// Setup (the draft), then a few tricks.
	if !c.until(func(g *domain.Game) bool {
		a := g.Attempt
		return a.Step == domain.StepPlay && len(a.Tricks) >= 3 && len(a.Trick.Plays) == 0
	}) {
		t.Fatal("the attempt ended before trick 4")
	}

	// Illegal moves are rejected, not punished.
	g := c.sync()
	k, v := c.actor(g)
	other := 1 - k
	if v.Attempt.Seat == 2 {
		t.Fatal("the dummy is not a player")
	}
	c.reject(other, apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: &c.rest(other).Attempt.Hand[0]}, "not_your_turn")
	c.reject(g.Attempt.Commander, apidto.ClientMessage{Type: apidto.CmdSendDistress, Direction: "left"}, "not_allowed")
	c.reject(other, apidto.ClientMessage{Type: apidto.CmdSendDistress, Direction: "left"}, "not_commander")
	c.reject(other, apidto.ClientMessage{Type: apidto.CmdSkipTokens}, "wrong_step")
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt}, "wrong_phase")
	c.reject(1, apidto.ClientMessage{Type: apidto.CmdNextAttempt}, "not_host")
	c.reject(k, apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: &apidto.Card{Suit: "pink", Rank: 10}}, "bad_request")
	c.reject(k, apidto.ClientMessage{Type: apidto.CmdPlayCard}, "bad_request")

	// Somebody communicates (public, with the position), then the server restarts.
	comm := -1
	for kk := range c.ps {
		if opts := c.rest(kk).Attempt.Actions.Communicate; len(opts) > 0 {
			comm = kk
			o := opts[0]
			c.run(kk, apidto.ClientMessage{Type: apidto.CmdCommunicate, Card: &o.Card, Position: o.Position})
			if o.Position == "" {
				t.Fatal("mission 1 is no dead zone")
			}
			for _, p := range c.ps {
				mem := p.ws.view.Attempt.Members[kk]
				if mem.Communication == nil || mem.Communication.Card != o.Card || mem.Communication.Position != o.Position || !mem.Communicated {
					t.Fatalf("communication seen as %+v", mem)
				}
			}
			c.reject(kk, apidto.ClientMessage{Type: apidto.CmdCommunicate, Card: &o.Card, Position: o.Position}, "already_communicated")
			break
		}
	}
	if comm < 0 {
		t.Fatal("nobody could communicate")
	}
	tricks := len(c.load().Attempt.Tricks)
	c.restart()
	if got := len(c.load().Attempt.Tricks); got != tricks {
		t.Fatalf("tricks after restart = %d, want %d", got, tricks)
	}
	if mem := c.rest(1 - comm).Attempt.Members[comm]; mem.Communication == nil {
		t.Fatal("the communication was lost in the restart")
	}

	// Play on over the new connections.
	res := c.playToDebrief()
	if !res.Success || res.Reason != "completed" {
		t.Fatalf("result = %+v, want a win with this seed", res)
	}
	if e := logbookOf(c.rest(0), 1); e == nil || e.Attempts != 1 || !e.Completed || e.Total != 1 {
		t.Fatalf("logbook = %+v", e)
	}
	for k, p := range c.ps {
		var ended *apidto.Event
		for _, m := range p.ws.log {
			if e := hasEvent(m.Events, apidto.EventAttemptEnded); e != nil {
				ended = e
			}
		}
		if ended == nil || ended.Success == nil || !*ended.Success || ended.Reason != "completed" {
			t.Fatalf("player %d never saw a winning attemptEnded: %+v", k, ended)
		}
	}

	// The mission is won: next attempt is refused, back to the lobby offers the next mission.
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt}, "invalid_result_choice")
	c.reject(1, apidto.ClientMessage{Type: apidto.CmdBackToLobby}, "not_host")
	c.run(0, apidto.ClientMessage{Type: apidto.CmdBackToLobby})
	v0 := c.rest(0)
	if v0.Phase != "lobby" || v0.Attempt != nil || v0.Mission != 2 || logbookOf(v0, 1) == nil {
		t.Fatalf("lobby after the debrief: %+v", v0)
	}

	// A replay of mission 1 given up by the guest; the logbook keeps its first result.
	c.run(0, apidto.ClientMessage{Type: apidto.CmdSelectMission, Mission: ptr(1)})
	c.run(0, apidto.ClientMessage{Type: apidto.CmdStartMission})
	c.run(1, apidto.ClientMessage{Type: apidto.CmdGiveUp})
	if r := c.rest(0).Attempt.Result; r == nil || r.Success || r.Reason != "gaveUp" || r.Seat != 1 {
		t.Fatalf("given-up result = %+v", r)
	}
	if e := logbookOf(c.rest(0), 1); e == nil || e.Attempts != 1 || !e.Completed {
		t.Fatalf("logbook after a replay = %+v", e)
	}
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdGiveUp}, "wrong_phase")
	c.run(0, apidto.ClientMessage{Type: apidto.CmdEndGame})
	if v := c.rest(1); v.Phase != "ended" {
		t.Fatalf("phase = %s", v.Phase)
	}
}

func ptr[T any](v T) *T { return &v }

// Four players: a lost attempt, the second attempt won, the next mission, back to the lobby.
func TestFourPlayersLoseThenWinThenNextMission(t *testing.T) {
	c := newCrew(t, 4, 2)
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt}, "wrong_phase")
	c.run(0, apidto.ClientMessage{Type: apidto.CmdStartMission})
	for k := range c.ps {
		v := c.rest(k)
		if len(v.Attempt.Members) != 4 || v.Attempt.TotalTricks != 10 || len(v.Attempt.Hand) != 10 {
			t.Fatalf("player %d: %+v", k, v.Attempt)
		}
		for _, m := range v.Attempt.Members {
			if m.Dummy {
				t.Fatal("four players have no dummy")
			}
		}
	}

	res := c.playToDebrief()
	if res.Success || res.Reason != "taskFailed" || res.Task != 0 || res.Seat < 0 {
		t.Fatalf("first attempt result = %+v, want a lost task", res)
	}
	if e := logbookOf(c.rest(1), 1); e == nil || e.Attempts != 1 || e.Completed {
		t.Fatalf("logbook after a loss = %+v", e)
	}
	// After a loss only a new attempt is possible, and only for the host.
	c.reject(1, apidto.ClientMessage{Type: apidto.CmdNextAttempt}, "not_host")
	c.reject(0, apidto.ClientMessage{Type: apidto.CmdNextMission}, "invalid_result_choice")
	c.run(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt})
	if v := c.rest(2); v.Phase != "playing" || v.Attempt.Number != 2 || v.Mission != 1 {
		t.Fatalf("second attempt view: %+v", v)
	}

	res = c.playToDebrief()
	if !res.Success {
		t.Fatalf("second attempt result = %+v, want a win", res)
	}
	if e := logbookOf(c.rest(3), 1); e == nil || e.Attempts != 2 || !e.Completed || e.Total != 2 {
		t.Fatalf("logbook after the win = %+v", e)
	}

	c.reject(0, apidto.ClientMessage{Type: apidto.CmdNextAttempt}, "invalid_result_choice")
	c.run(0, apidto.ClientMessage{Type: apidto.CmdNextMission})
	v := c.rest(0)
	if v.Phase != "playing" || v.Mission != 2 || v.Attempt.Number != 1 || v.MissionDef == nil || v.MissionDef.Tasks != 2 || len(v.Attempt.Tasks) != 2 {
		t.Fatalf("mission 2 view: %+v", v)
	}
	res = c.playToDebrief() // either outcome
	c.run(0, apidto.ClientMessage{Type: apidto.CmdBackToLobby})
	v = c.rest(0)
	if v.Phase != "lobby" || v.Attempt != nil {
		t.Fatalf("lobby: %+v", v)
	}
	if e := logbookOf(v, 2); e == nil || e.Attempts != 1 || e.Completed != res.Success {
		t.Fatalf("logbook of mission 2 = %+v (result %+v)", e, res)
	}
	want := 2
	if res.Success {
		want = 3
	}
	if v.Mission != want {
		t.Fatalf("selected mission = %d, want %d", v.Mission, want)
	}
	// The next mission can be started from the lobby again.
	c.run(0, apidto.ClientMessage{Type: apidto.CmdStartMission})
	if v := c.rest(1); v.Phase != "playing" || v.Mission != want {
		t.Fatalf("restarted: %+v", v)
	}
}

// The distress signal over the WebSocket: the commander sends it, everybody chooses a card, the
// cards move, and nobody learns another player's card.
func TestDistressSignalOverWebSocket(t *testing.T) {
	c := newCrew(t, 3, 5)
	c.distress = true
	c.run(0, apidto.ClientMessage{Type: apidto.CmdStartMission})
	if !c.until(func(g *domain.Game) bool { return g.Attempt.Step == domain.StepPass }) {
		t.Fatal("no pass step")
	}
	g := c.sync()
	cmd := g.Attempt.Commander
	for k := range c.ps {
		v := c.rest(k)
		if v.Attempt.Pass == nil || v.Attempt.Pass.Direction != "left" || len(v.Attempt.Pass.Waiting) != 3 || !v.Attempt.Distress {
			t.Fatalf("player %d pass view: %+v", k, v.Attempt)
		}
		if k != cmd && v.Attempt.Actions.SendDistress {
			t.Fatalf("only the commander sends the signal")
		}
	}
	// A trump cannot be passed.
	for _, h := range c.rest(0).Attempt.Hand {
		if h.Suit == "trump" {
			h := h
			c.reject(0, apidto.ClientMessage{Type: apidto.CmdChoosePassCard, Card: &h}, "cannot_pass_trump")
			break
		}
	}
	for g.Attempt.Step == domain.StepPass {
		c.step()
		g = c.load()
	}
	for k := range c.ps {
		ex := c.rest(k).Attempt.Exchange
		if ex == nil || ex.Kind != "pass" || ex.Gave == nil || ex.Got == nil {
			t.Fatalf("player %d exchange = %+v", k, ex)
		}
	}
	// Only the pass-card events were public, and they named no card.
	for k, p := range c.ps {
		for _, m := range p.ws.log {
			for _, e := range m.Events {
				if (e.Type == apidto.EventPassCardChosen || e.Type == apidto.EventCardsPassed) && e.Card != nil {
					t.Fatalf("player %d: %s names a card", k, e.Type)
				}
			}
		}
	}
	if e := logbookOf(c.rest(0), 1); e == nil || !e.Distress {
		t.Fatalf("logbook = %+v", e)
	}
	c.playToDebrief()
}
