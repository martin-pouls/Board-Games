package apidto

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/gametest"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
	"github.com/martin-pouls/the-crew/backend/internal/core/ports"
)

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// reach starts mission with n players for seeds 1, 2, ... until done holds before the attempt
// ends, and returns that game and its RNG.
func reach(t *testing.T, n, mission int, done func(*domain.Game) bool) (*domain.Game, domain.RNG) {
	t.Helper()
	return reachWith(t, n, mission, domain.Options{}, done)
}

// reachWith is reach for a game with the lobby options opts.
func reachWith(t *testing.T, n, mission int, opts domain.Options, done func(*domain.Game) bool) (*domain.Game, domain.RNG) {
	t.Helper()
	for seed := uint64(1); seed < 300; seed++ {
		rng := gametest.NewRNG(seed)
		g := gametest.NewGameWith(t, n, mission, opts, rng)
		if gametest.TryUntil(t, g, rng, done) {
			return g, rng
		}
	}
	t.Fatalf("mission %d with %d players: no seed reaches the state", mission, n)
	return nil, nil
}

func atStep(step domain.Step) func(*domain.Game) bool {
	return func(g *domain.Game) bool { return g.Attempt.Step == step }
}

func viewOf(t *testing.T, g *domain.Game, seat int) PlayerView {
	t.Helper()
	v, err := g.ViewFor(g.Players[seat].ID)
	if err != nil {
		t.Fatal(err)
	}
	return FromView(v)
}

// nullKeys are the JSON keys whose value may be null; every other value must be a real array,
// object or scalar.
var nullKeys = map[string]bool{
	"missionDef": true, "attempt": true, "card": true, "up": true, "communication": true,
	"poll": true, "pass": true, "exchange": true, "lastTrick": true, "result": true,
	"cards": true, "chosen": true, "gave": true, "got": true, "holder": true,
}

// assertNoNulls fails for a null in a key that is not in nullKeys, and for a null inside an
// array other than in a poll's answers.
func assertNoNulls(t *testing.T, v any) {
	t.Helper()
	var walk func(path string, key string, v any)
	walk = func(path, key string, v any) {
		switch x := v.(type) {
		case nil:
			if !nullKeys[key] {
				t.Errorf("unexpected null at %s", path)
			}
		case map[string]any:
			for k, e := range x {
				walk(path+"."+k, k, e)
			}
		case []any:
			for i, e := range x {
				if e == nil && key == "answers" {
					continue
				}
				walk(fmt.Sprintf("%s[%d]", path, i), "", e)
			}
		}
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var any_ any
	if err := json.Unmarshal(b, &any_); err != nil {
		t.Fatal(err)
	}
	walk("$", "", any_)
}

func TestFromViewEmptySlicesMarshalAsArrays(t *testing.T) {
	s := mustJSON(t, FromView(domain.PlayerView{}))
	if !strings.Contains(s, `"players":[]`) || !strings.Contains(s, `"logbook":[]`) {
		t.Errorf("missing empty arrays in %s", s)
	}
	if got := mustJSON(t, FromEvents(nil)); got != "[]" {
		t.Errorf("FromEvents(nil) = %s", got)
	}
}

func TestFromViewLobby(t *testing.T) {
	g, err := domain.NewGame("K7MQ2", domain.EditionPlanetNine, "p1", "Anna", domain.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Ben"); err != nil {
		t.Fatal(err)
	}
	v, err := g.ViewFor("p2")
	if err != nil {
		t.Fatal(err)
	}
	got := FromView(v)
	mission1 := FromMission(domain.MissionDef{Number: 1, Tasks: 1, Distribution: domain.DistOpen, Poll: domain.PollYesNo})
	want := PlayerView{
		GameID: "K7MQ2", Revision: 2, Phase: "lobby", You: "p2", HostID: "p1",
		Players: []PlayerSummary{
			{ID: "p1", Name: "Anna", Seat: 0, IsHost: true},
			{ID: "p2", Name: "Ben", Seat: 1, IsYou: true},
		},
		Edition: "planetNine", Mission: 1, MissionDef: &mission1, Logbook: []LogEntry{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FromView mismatch:\n got %+v\nwant %+v", got, want)
	}
	assertNoNulls(t, got)
	if got.Attempt != nil {
		t.Error("a lobby has no attempt")
	}
}

func TestOptionsOnTheWire(t *testing.T) {
	g, err := domain.NewGame("K7MQ2", domain.EditionPlanetNine, "p1", "Anna", domain.Options{})
	if err != nil {
		t.Fatal(err)
	}
	view := func() PlayerView {
		v, err := g.ViewFor("p1")
		if err != nil {
			t.Fatal(err)
		}
		return FromView(v)
	}
	if got := mustJSON(t, view().Options); got != `{"distress":false}` {
		t.Fatalf("default options = %s", got)
	}
	if err := g.SetOptions("p1", GameOptions{Distress: true}.ToDomain()); err != nil {
		t.Fatal(err)
	}
	if got := view().Options; got != (GameOptions{Distress: true}) || FromOptions(g.Options) != got {
		t.Fatalf("options = %+v", got)
	}
	if !strings.Contains(mustJSON(t, view()), `"options":{"distress":true}`) {
		t.Fatalf("view = %s", mustJSON(t, view()))
	}
	// An absent flag in a setOptions command means off.
	var m ClientMessage
	if err := json.Unmarshal([]byte(`{"type":"setOptions","requestId":"1","options":{}}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Options == nil || m.Options.ToDomain().Distress {
		t.Fatalf("decoded %+v", m.Options)
	}
}

func TestLogbookSortedByMission(t *testing.T) {
	v := domain.PlayerView{Logbook: map[int]domain.LogEntry{
		12: {Attempts: 3, Completed: true},
		2:  {Attempts: 1, Distress: true, Completed: true},
		40: {Attempts: 2, Distress: true},
	}}
	want := []LogEntry{
		{Mission: 2, Attempts: 1, Distress: true, Completed: true, Total: 2},
		{Mission: 12, Attempts: 3, Completed: true, Total: 3},
		{Mission: 40, Attempts: 2, Distress: true, Total: 2}, // the +1 counts only on completion,
	}
	if got := FromView(v).Logbook; !reflect.DeepEqual(got, want) {
		t.Fatalf("logbook = %+v, want %+v", got, want)
	}
}

// A 2-player mid-attempt view: the dummy's stacks, the viewer's hand and playable cards, tasks,
// the trick in progress and the last trick.
func TestFromViewMidAttemptWithDummy(t *testing.T) {
	g, _ := reach(t, 2, 1, func(g *domain.Game) bool {
		a := g.Attempt
		return a.Step == domain.StepPlay && len(a.Tricks) >= 1 && len(a.Trick.Plays) == 1 &&
			a.Members[0].CardCount() > 0
	})
	a := g.Attempt
	for seat := range 2 {
		v := viewOf(t, g, seat)
		assertNoNulls(t, v)
		at := v.Attempt
		if at == nil || at.Step != "play" || at.Seat != seat || len(at.Members) != 3 {
			t.Fatalf("attempt = %+v", at)
		}
		if at.TotalTricks != 13 || at.TrickNumber != len(a.Tricks)+1 || len(at.Trick) != 1 || at.LastTrick == nil {
			// the last trick stays visible while the next one is played
			t.Fatalf("tricks: %+v", at)
		}
		d := at.Members[2]
		if !d.Dummy || d.PlayerID != "" || d.Name != "" || len(d.Stacks) != 7 || d.Cards != nil || d.CardCount != a.Members[2].CardCount() {
			t.Fatalf("dummy = %+v", d)
		}
		for i, s := range a.Members[2].Stacks {
			if (s.Up == nil) != (d.Stacks[i].Up == nil) || d.Stacks[i].Hidden != (s.Down != nil) {
				t.Errorf("stack %d = %+v, domain %+v", i, d.Stacks[i], s)
			}
		}
		for _, m := range at.Members[:2] {
			if m.Dummy || m.PlayerID == "" || m.Name == "" || m.Stacks == nil || len(m.Stacks) != 0 || m.Cards != nil {
				t.Errorf("human member = %+v", m)
			}
		}
		if len(at.Hand) != a.Members[seat].CardCount() {
			t.Errorf("hand = %v", at.Hand)
		}
		// The dummy's hidden cards and the other hand appear nowhere in the view.
		for _, s := range a.Members[2].Stacks {
			if s.Down != nil && inView(v, *s.Down) {
				t.Errorf("hidden dummy card %v leaked to seat %d", *s.Down, seat)
			}
		}
		for _, c := range a.Members[1-seat].Hand {
			if inView(v, c) {
				t.Errorf("seat %d sees %v of the other hand", seat, c)
			}
		}
	}
	// The commander acts for the dummy: on its turn the playable cards are the dummy's.
	cmd := a.Commander
	turn := (a.Trick.Leader + len(a.Trick.Plays)) % 3
	v := viewOf(t, g, cmd)
	if turn == 2 {
		want := FromView(domainView(t, g, cmd)).Attempt.Playable
		if len(v.Attempt.Playable) == 0 || !reflect.DeepEqual(v.Attempt.Playable, want) {
			t.Errorf("commander's playable on JARVIS's turn = %v", v.Attempt.Playable)
		}
		if !v.Attempt.Members[2].MustAct || v.Attempt.Turn != 2 || !slices.Equal(v.Attempt.Actors, []int{cmd}) {
			t.Errorf("turn = %d actors = %v", v.Attempt.Turn, v.Attempt.Actors)
		}
	}
}

func domainView(t *testing.T, g *domain.Game, seat int) domain.PlayerView {
	t.Helper()
	v, err := g.ViewFor(g.Players[seat].ID)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// inView reports whether card c occurs anywhere in the view's hands, dummy stacks, plays or tasks.
func inView(v PlayerView, c domain.Card) bool {
	w := FromCard(c)
	at := v.Attempt
	if at == nil {
		return false
	}
	if slices.Contains(at.Hand, w) {
		return true
	}
	for _, m := range at.Members {
		if slices.Contains(m.Cards, w) {
			return true
		}
		for _, s := range m.Stacks {
			if s.Up != nil && *s.Up == w {
				return true
			}
		}
		if m.Communication != nil && m.Communication.Card == w {
			return true
		}
	}
	for _, p := range at.Trick {
		if p.Card == w {
			return true
		}
	}
	if at.LastTrick != nil {
		for _, p := range at.LastTrick.Plays {
			if p.Card == w {
				return true
			}
		}
	}
	return false
}

// A communication in a dead zone has no position, in the member, in the options and in the event.
func TestDeadZoneCommunication(t *testing.T) {
	g, _ := reach(t, 2, 6, func(g *domain.Game) bool {
		if g.Attempt.Step != domain.StepPlay {
			return false
		}
		for seat := range 2 {
			if g.CanCommunicate(seat) == nil && len(domainView(t, g, seat).Attempt.Actions.Communicate) > 0 {
				return true
			}
		}
		return false
	})
	seat := 0
	if len(domainView(t, g, 0).Attempt.Actions.Communicate) == 0 {
		seat = 1
	}
	v := viewOf(t, g, seat)
	opts := v.Attempt.Actions.Communicate
	if len(opts) == 0 {
		t.Fatal("no options")
	}
	for _, o := range opts {
		if o.Position != "" {
			t.Errorf("dead zone option has position %q", o.Position)
		}
		if o.Card.Suit == "trump" {
			t.Errorf("trump offered: %+v", o)
		}
	}
	if err := g.Communicate(g.Players[seat].ID, domain.Card{Suit: domain.Suit(opts[0].Card.Suit), Rank: domain.Rank(opts[0].Card.Rank)}, ""); err != nil {
		t.Fatal(err)
	}
	evs := FromEvents(g.PullEvents())
	if len(evs) != 1 || evs[0].Type != EventCommunicated || evs[0].Card == nil || evs[0].Position != "" || evs[0].Seat == nil || *evs[0].Seat != seat {
		t.Fatalf("event = %+v", evs)
	}
	for s := range 2 {
		m := viewOf(t, g, s).Attempt.Members[seat]
		if !m.Communicated || m.Communication == nil || m.Communication.Position != "" || m.Communication.Card != opts[0].Card {
			t.Errorf("seat %d sees member %+v", s, m)
		}
	}
	if got := viewOf(t, g, seat).Attempt.Actions.Communicate; len(got) != 0 {
		t.Errorf("options after communicating: %+v", got)
	}
	// A normal mission names the position.
	g2, _ := reach(t, 2, 1, func(g *domain.Game) bool {
		return g.Attempt.Step == domain.StepPlay && len(domainView(t, g, 0).Attempt.Actions.Communicate) > 0
	})
	for _, o := range viewOf(t, g2, 0).Attempt.Actions.Communicate {
		if o.Position != "highest" && o.Position != "lowest" && o.Position != "only" {
			t.Errorf("option %+v has no position", o)
		}
	}
}

// Face-down tasks have no card; an open poll shows answers, null for those who have not answered.
func TestFaceDownTasksAndPoll(t *testing.T) {
	g, _ := reach(t, 3, 20, atStep(domain.StepPoll))
	for seat := range 3 {
		v := viewOf(t, g, seat)
		assertNoNulls(t, v)
		at := v.Attempt
		if len(at.Tasks) != 2 {
			t.Fatalf("tasks = %+v", at.Tasks)
		}
		for _, task := range at.Tasks {
			if task.Card != nil || !task.FaceDown || task.Owner != -2 {
				t.Errorf("face-down task = %+v", task)
			}
		}
		if !strings.Contains(mustJSON(t, at.Tasks[0]), `"card":null`) {
			t.Errorf("face-down card must be null: %s", mustJSON(t, at.Tasks[0]))
		}
		p := at.Poll
		if p == nil || p.Kind != "yesNo" || len(p.Answers) != 3 || p.Objective != -1 || p.Task != -1 {
			t.Fatalf("poll = %+v", p)
		}
		if at.Step != "poll" || at.Pick != "" || at.Turn != -1 {
			t.Errorf("step = %q pick = %q turn = %d", at.Step, at.Pick, at.Turn)
		}
		if seat == at.Commander && len(at.Actors) != 2 {
			t.Errorf("actors = %v", at.Actors)
		}
	}
	// One answer: public, with its value; the others stay null.
	cmd := g.Attempt.Commander
	voter := (cmd + 1) % 3
	if err := g.AnswerPoll(g.Players[voter].ID, false); err != nil {
		t.Fatal(err)
	}
	evs := FromEvents(g.PullEvents())
	if len(evs) != 1 || evs[0].Type != EventPollAnswered || evs[0].Yes == nil || *evs[0].Yes || *evs[0].Seat != voter {
		t.Fatalf("event = %s", mustJSON(t, evs))
	}
	if s := mustJSON(t, evs[0]); !strings.Contains(s, `"yes":false`) {
		t.Errorf("a false answer must survive omitempty: %s", s)
	}
	p := viewOf(t, g, cmd).Attempt.Poll
	if p.Answers[voter] == nil || *p.Answers[voter] || p.Answers[cmd] != nil {
		t.Errorf("answers = %+v", p.Answers)
	}
	if !strings.Contains(mustJSON(t, p), "null") {
		t.Errorf("unanswered entries must be null: %s", mustJSON(t, p))
	}
	// After the last answer the commander picks; the tasks are revealed after the pick.
	for g.Attempt.Step == domain.StepPoll {
		gametest.Step(t, g, gametest.NewRNG(1))
	}
	v := viewOf(t, g, cmd)
	if v.Attempt.Step != "pick" || v.Attempt.Pick != "decision" || v.Attempt.Tasks[0].Card != nil {
		t.Fatalf("after polling: %+v", v.Attempt)
	}
	gametest.Step(t, g, gametest.NewRNG(1))
	if c := viewOf(t, g, 0).Attempt.Tasks[0].Card; c == nil {
		t.Error("tasks should be revealed after the decision")
	}

	// Mission 5 asks "well / ill".
	g5, _ := reach(t, 3, 5, atStep(domain.StepPoll))
	v5 := viewOf(t, g5, 0)
	if v5.Attempt.Poll.Kind != "wellIll" || len(v5.Attempt.Tasks) != 0 || len(v5.Attempt.Objectives) != 1 {
		t.Fatalf("mission 5 poll: %+v", v5.Attempt)
	}
	if o := v5.Attempt.Objectives[0]; o.Kind != "winNoTricks" || o.Owner != -2 || o.Status != "open" || o.Cards == nil || o.Tricks == nil {
		t.Errorf("objective = %+v", o)
	}
	if v5.Attempt.Poll.Objective != 0 {
		t.Errorf("poll objective = %d", v5.Attempt.Poll.Objective)
	}
}

// The distress pass: each player sees only their own choice; the events name no card; afterwards
// the exchange shows the viewer's side only.
func TestPassAndExchange(t *testing.T) {
	g, rng := reachWith(t, 3, 1, domain.Options{Distress: true}, func(g *domain.Game) bool {
		return g.Attempt.Step == domain.StepPlay && g.CanSendDistress(g.Attempt.Commander) == nil
	})
	cmd := g.Attempt.Commander
	if v := viewOf(t, g, cmd); !v.Attempt.Actions.SendDistress {
		t.Fatal("the commander may send the signal")
	}
	if v := viewOf(t, g, (cmd+1)%3); v.Attempt.Actions.SendDistress {
		t.Fatal("only the commander may")
	}
	if err := g.SendDistress(g.Players[cmd].ID, domain.DirRight); err != nil {
		t.Fatal(err)
	}
	evs := FromEvents(g.PullEvents())
	if len(evs) != 1 || evs[0].Type != EventDistressSignal || evs[0].Direction != "right" || evs[0].Skipped == nil || *evs[0].Skipped {
		t.Fatalf("event = %s", mustJSON(t, evs))
	}
	first := g.Actors()[0]
	var chosen domain.Card
	for _, c := range g.Attempt.Members[first].Hand {
		if !c.IsTrump() {
			chosen = c
			break
		}
	}
	if err := g.ChoosePassCard(g.Players[first].ID, chosen); err != nil {
		t.Fatal(err)
	}
	evs = FromEvents(g.PullEvents())
	if len(evs) != 1 || evs[0].Type != EventPassCardChosen || *evs[0].Seat != first || evs[0].Card != nil {
		t.Fatalf("event = %s", mustJSON(t, evs))
	}
	for seat := range 3 {
		v := viewOf(t, g, seat)
		assertNoNulls(t, v)
		p := v.Attempt.Pass
		if p == nil || p.Direction != "right" || len(p.Waiting) != 2 || slices.Contains(p.Waiting, first) {
			t.Fatalf("pass = %+v", p)
		}
		if (seat == first) != (p.Chosen != nil) {
			t.Errorf("seat %d sees chosen %v", seat, p.Chosen)
		}
		if seat != first && inView(v, chosen) {
			// the card is in the chooser's hand, never in another view
			t.Errorf("seat %d sees the chosen card", seat)
		}
	}
	for g.Attempt.Step == domain.StepPass {
		gametest.Step(t, g, rng)
	}
	evs = FromEvents(g.PullEvents())
	last := evs[len(evs)-1]
	if last.Type != EventCardsPassed || last.Direction != "right" || last.Card != nil {
		t.Fatalf("events = %s", mustJSON(t, evs))
	}
	for _, e := range evs {
		if e.Card != nil {
			t.Errorf("pass event carries a card: %s", mustJSON(t, e))
		}
	}
	for seat := range 3 {
		v := viewOf(t, g, seat)
		ex := v.Attempt.Exchange
		if ex == nil || ex.Kind != "pass" || ex.Gave == nil || ex.Got == nil || v.Attempt.Pass != nil {
			t.Errorf("seat %d exchange = %+v", seat, ex)
		}
	}
	if g.Logbook[1].Distress != true {
		t.Error("logbook should note the signal")
	}
	if l := viewOf(t, g, 0).Logbook; len(l) != 1 || !l[0].Distress || l[0].Total != 0 || !viewOf(t, g, 0).Attempt.Distress {
		t.Errorf("logbook = %+v", l)
	}
}

// After a debrief every hand is revealed; the result explains the outcome.
func TestDebriefView(t *testing.T) {
	for _, n := range []int{2, 4} {
		rng := gametest.NewRNG(2)
		g := gametest.NewGame(t, n, 1, rng)
		gametest.ToEnd(t, g, rng)
		for seat := range n {
			v := viewOf(t, g, seat)
			assertNoNulls(t, v)
			at := v.Attempt
			if v.Phase != "debrief" || at.Step != "" || at.Result == nil || at.Poll != nil || at.Pass != nil {
				t.Fatalf("debrief view = %+v", at)
			}
			if at.Result.Reason == "" || at.Result.Trick < 1 {
				t.Errorf("result = %+v", at.Result)
			}
			for _, m := range at.Members {
				if m.Cards == nil {
					t.Errorf("seat %d: member %d cards must be revealed", seat, m.Seat)
				}
			}
			if len(at.Playable) != 0 || len(at.Actions.Communicate) != 0 || at.Actions.SendDistress {
				t.Errorf("no actions in the debrief: %+v", at)
			}
			if at.LastTrick == nil {
				t.Error("the last trick stays visible")
			}
			if len(v.Logbook) != 1 || v.Logbook[0].Mission != 1 || v.Logbook[0].Completed != at.Result.Success {
				t.Errorf("logbook = %+v", v.Logbook)
			}
		}
	}
	// A given-up attempt: gaveUp, with the seat that gave up.
	rng := gametest.NewRNG(2)
	g := gametest.NewGame(t, 3, 1, rng)
	if err := g.GiveUp("p2"); err != nil {
		t.Fatal(err)
	}
	r := viewOf(t, g, 0).Attempt.Result
	if r == nil || r.Success || r.Reason != "gaveUp" || r.Seat != 1 || r.Task != -1 || r.Objective != -1 {
		t.Fatalf("result = %+v", r)
	}
	ev := FromEvents(g.PullEvents())
	if len(ev) != 1 || ev[0].Type != EventAttemptEnded || ev[0].Success == nil || *ev[0].Success || ev[0].Reason != "gaveUp" {
		t.Fatalf("event = %s", mustJSON(t, ev))
	}
}

func TestFromMissionAndCatalogue(t *testing.T) {
	c := Catalogue(domain.EditionPlanetNine)
	if c.Edition != "planetNine" || len(c.Missions) != 50 {
		t.Fatalf("catalogue = %s %d", c.Edition, len(c.Missions))
	}
	s := mustJSON(t, c)
	if strings.Contains(s, "null") && !strings.Contains(s, `"holder":null`) {
		t.Errorf("unexpected null in the catalogue")
	}
	assertNoNullsExcept(t, c, map[string]bool{"holder": true})
	for i, m := range c.Missions {
		if m.Number != i+1 {
			t.Fatalf("mission %d at index %d", m.Number, i)
		}
	}
	m := func(n int) MissionDef { return c.Missions[n-1] }
	if x := m(6); !x.DeadZone || !reflect.DeepEqual(x.Tokens, []string{"arrow1", "arrow2"}) || x.Tasks != 3 || x.Distribution != "open" {
		t.Errorf("mission 6 = %+v", x)
	}
	if x := m(18); x.FromTrick != 2 || x.Golden || x.Tasks != 5 {
		t.Errorf("mission 18 = %+v", x)
	}
	if x := m(20); x.Distribution != "commanderDecision" || x.Tasks != 2 {
		t.Errorf("mission 20 = %+v", x)
	}
	if x := m(24); x.Distribution != "commanderDistribution" {
		t.Errorf("mission 24 = %+v", x)
	}
	if x := m(23); !reflect.DeepEqual(x.Setup, []string{"swapTokens"}) {
		t.Errorf("mission 23 = %+v", x)
	}
	if x := m(12); !reflect.DeepEqual(x.Setup, []string{"drawAfterFirstTrick"}) || !reflect.DeepEqual(x.Tokens, []string{"omega"}) {
		t.Errorf("mission 12 = %+v", x)
	}
	if x := m(28); !x.Golden || x.FromTrick != 3 {
		t.Errorf("mission 28 = %+v", x)
	}
	if x := m(5); x.Poll != "wellIll" || x.Tasks != 0 || len(x.Objectives) != 1 || x.Objectives[0].OwnerRule != "chosen" || x.Objectives[0].Kind != "winNoTricks" {
		t.Errorf("mission 5 = %+v", x)
	}
	if x := m(46); len(x.Objectives) != 1 || x.Objectives[0].OwnerRule != "leftOfHolder" || x.Objectives[0].Holder == nil ||
		*x.Objectives[0].Holder != (Card{"pink", 9}) || x.Objectives[0].Suit != "pink" {
		t.Errorf("mission 46 = %+v", x)
	}
	if x := m(48); len(x.Objectives) != 1 || x.Objectives[0].TaskToken != "omega" || x.Objectives[0].Kind != "cardInLastTrick" {
		t.Errorf("mission 48 = %+v", x)
	}
	if x := m(50); len(x.Objectives) != 2 || x.Objectives[0].OwnerRule != "roleA" || x.Objectives[1].OwnerRule != "roleB" ||
		!reflect.DeepEqual(x.Objectives[1].Tricks, []int{-1}) || !reflect.DeepEqual(x.Objectives[0].Tricks, []int{1, 2, 3, 4}) {
		t.Errorf("mission 50 = %+v", x)
	}
	if x := m(34); len(x.Objectives) != 2 || x.Objectives[0].Kind != "crewBalance" || x.Objectives[0].Count != 1 || x.Objectives[0].OwnerRule != "crew" ||
		x.Objectives[1].OwnerRule != "commander" || !reflect.DeepEqual(x.Objectives[1].Tricks, []int{1, -1}) {
		t.Errorf("mission 34 = %+v", x)
	}
	if x := m(13); len(x.Objectives[0].Cards) != 4 || x.Objectives[0].Cards[0] != (Card{"trump", 1}) || x.Objectives[0].Ordered {
		t.Errorf("mission 13 = %+v", x)
	}
	if x := m(44); !x.Objectives[0].Ordered {
		t.Errorf("mission 44 = %+v", x)
	}
	if x := m(41); !x.Objectives[0].NotWithTrump || x.Poll != "yesNo" {
		t.Errorf("mission 41 = %+v", x)
	}
	// The mission def in a view is the same row.
	g, err := domain.NewGame("K7MQ2", domain.EditionPlanetNine, "p1", "Anna", domain.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.SelectMission("p1", 34); err != nil {
		t.Fatal(err)
	}
	v, _ := g.ViewFor("p1")
	if got := FromView(v); got.MissionDef == nil || !reflect.DeepEqual(*got.MissionDef, m(34)) || got.Mission != 34 {
		t.Errorf("view mission def = %+v", got.MissionDef)
	}
	// An unknown mission or edition has no def.
	if got := FromView(domain.PlayerView{Edition: "x", Mission: 3}); got.MissionDef != nil {
		t.Errorf("unknown edition has a def: %+v", got.MissionDef)
	}
}

func assertNoNullsExcept(t *testing.T, v any, allowed map[string]bool) {
	t.Helper()
	saved := nullKeys
	nullKeys = allowed
	defer func() { nullKeys = saved }()
	assertNoNulls(t, v)
}

// The views of every mission, crew size and step carry no unexpected null.
func TestViewsNeverHaveUnexpectedNulls(t *testing.T) {
	for _, mission := range []int{1, 5, 6, 11, 12, 20, 23, 24, 27, 33, 34, 40, 41, 46, 48, 50} {
		for _, n := range []int{2, 3, 5} {
			rng := gametest.NewRNG(uint64(mission*7 + n))
			g := gametest.NewGame(t, n, mission, rng)
			for range 60 {
				for seat := range n {
					assertNoNulls(t, viewOf(t, g, seat))
				}
				if g.Phase != domain.PhasePlaying {
					break
				}
				gametest.Step(t, g, rng)
			}
		}
	}
}

func TestFromSummary(t *testing.T) {
	created := time.Date(2026, 10, 10, 12, 30, 5, 0, time.UTC)
	tests := []struct {
		name string
		in   domain.GameSummary
		want GameSummary
	}{
		{
			"lobby",
			domain.GameSummary{GameID: "ABCDE", Phase: domain.PhaseLobby, HostName: "Anna", PlayerCount: 1, MaxPlayers: 5, Joinable: true},
			GameSummary{GameID: "ABCDE", Phase: "lobby", HostName: "Anna", PlayerCount: 1, MaxPlayers: 5, Joinable: true},
		},
		{
			"created and ended",
			domain.GameSummary{GameID: "ABCDE", Phase: domain.PhaseEnded, PlayerCount: 2, MaxPlayers: 5, CreatedAt: created.In(time.FixedZone("x", 3600))},
			GameSummary{GameID: "ABCDE", Phase: "ended", PlayerCount: 2, MaxPlayers: 5, CreatedAt: "2026-10-10T12:30:05Z"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromSummary(tt.in); got != tt.want {
				t.Fatalf("FromSummary = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCardMapping(t *testing.T) {
	c, err := Card{Suit: "pink", Rank: 9}.ToDomain()
	if err != nil || c != (domain.Card{Suit: domain.SuitPink, Rank: 9}) {
		t.Fatalf("ToDomain = %v %v", c, err)
	}
	if got := FromCard(domain.Card{Suit: domain.SuitTrump, Rank: 4}); got != (Card{"trump", 4}) {
		t.Fatalf("FromCard = %+v", got)
	}
	for _, bad := range []Card{{"doves", 1}, {"pink", 0}, {"pink", 10}, {"trump", 5}, {"", 1}} {
		if _, err := bad.ToDomain(); !errors.Is(err, ErrBadRequest) {
			t.Errorf("%+v: err = %v, want bad request", bad, err)
		}
	}
}

func TestFromEvents(t *testing.T) {
	tests := []struct {
		in   domain.Event
		want string
	}{
		{domain.PlayerJoined{PlayerID: "p2", Name: "Ben"}, `{"type":"playerJoined","playerId":"p2","name":"Ben"}`},
		{domain.PlayerLeft{PlayerID: "p2", HostID: "p1"}, `{"type":"playerLeft","playerId":"p2","hostId":"p1"}`},
		{domain.MissionSelected{Mission: 7}, `{"type":"missionSelected","mission":7}`},
		{domain.OptionsChanged{Options: domain.Options{Distress: true}}, `{"type":"optionsChanged"}`},
		{domain.AttemptStarted{Mission: 7, Number: 2, Commander: 0}, `{"type":"attemptStarted","mission":7,"number":2,"commander":0}`},
		{domain.TokensAdjusted{From: 0, To: 2}, `{"type":"tokensAdjusted","from":0,"to":2}`},
		{domain.TokensAdjusted{From: -1, To: -1}, `{"type":"tokensAdjusted","from":-1,"to":-1}`},
		{domain.PollAnswered{Seat: 0, Yes: false}, `{"type":"pollAnswered","seat":0,"yes":false}`},
		{domain.PollAnswered{Seat: 2, Yes: true}, `{"type":"pollAnswered","seat":2,"yes":true}`},
		{domain.Picked{Seats: []int{1, 3}}, `{"type":"picked","seats":[1,3]}`},
		{domain.TaskPicked{TaskID: 0, Seat: 0}, `{"type":"taskPicked","taskId":0,"seat":0}`},
		{domain.DistressSignal{Direction: domain.DirLeft}, `{"type":"distressSignal","direction":"left","skipped":false}`},
		{domain.DistressSignal{Skipped: true}, `{"type":"distressSignal","skipped":true}`},
		{domain.PassCardChosen{Seat: 1}, `{"type":"passCardChosen","seat":1}`},
		{domain.CardsPassed{Direction: domain.DirRight}, `{"type":"cardsPassed","direction":"right"}`},
		{domain.TaskHandedOver{TaskID: 1, From: 0, To: 4}, `{"type":"taskHandedOver","taskId":1,"from":0,"to":4}`},
		{domain.Communicated{Seat: 1, Card: domain.Card{Suit: domain.SuitBlue, Rank: 7}, Position: domain.PosHighest}, `{"type":"communicated","seat":1,"card":{"suit":"blue","rank":7},"position":"highest"}`},
		{domain.Communicated{Seat: 0, Card: domain.Card{Suit: domain.SuitBlue, Rank: 7}}, `{"type":"communicated","seat":0,"card":{"suit":"blue","rank":7}}`},
		{domain.CardPlayed{Seat: 0, Card: domain.Card{Suit: domain.SuitBlue, Rank: 7}}, `{"type":"cardPlayed","seat":0,"card":{"suit":"blue","rank":7}}`},
		{domain.TrickWon{Number: 3, Seat: 2}, `{"type":"trickWon","number":3,"seat":2}`},
		{domain.CardsDrawn{}, `{"type":"cardsDrawn"}`},
		{domain.AttemptEnded{Success: false, Reason: domain.ReasonTaskFailed}, `{"type":"attemptEnded","success":false,"reason":"taskFailed"}`},
		{domain.AttemptEnded{Success: true, Reason: domain.ReasonCompleted}, `{"type":"attemptEnded","success":true,"reason":"completed"}`},
		{domain.BackInLobby{}, `{"type":"backInLobby"}`},
		{domain.GameEnded{ByPlayerID: "p1"}, `{"type":"gameEnded","byPlayerId":"p1"}`},
	}
	var evs []domain.Event
	for _, tt := range tests {
		evs = append(evs, tt.in)
	}
	got := FromEvents(evs)
	if len(got) != len(tests) {
		t.Fatalf("len = %d, want %d", len(got), len(tests))
	}
	for i, tt := range tests {
		if s := mustJSON(t, got[i]); !sameJSON(s, tt.want) {
			t.Errorf("event %d = %s, want %s", i, s, tt.want)
		}
	}
	// Every domain event type has a case: the list above covers all of events.go.
	seen := map[string]bool{}
	for _, e := range got {
		seen[e.Type] = true
	}
	for _, typ := range []string{
		EventPlayerJoined, EventPlayerLeft, EventMissionSelected, EventOptionsChanged, EventAttemptStarted, EventTokensAdjusted, EventPollAnswered,
		EventPicked, EventTaskPicked, EventDistressSignal, EventPassCardChosen, EventCardsPassed, EventTaskHandedOver,
		EventCommunicated, EventCardPlayed, EventTrickWon, EventCardsDrawn, EventAttemptEnded, EventBackInLobby, EventGameEnded,
	} {
		if !seen[typ] {
			t.Errorf("event type %s is not mapped", typ)
		}
	}
	// Events that concern hidden cards carry none.
	for _, typ := range []string{EventPassCardChosen, EventCardsPassed, EventCardsDrawn} {
		for _, e := range got {
			if e.Type == typ && e.Card != nil {
				t.Errorf("%s carries a card", typ)
			}
		}
	}
	if _, ok := FromEvent(nil); ok {
		t.Error("a nil event is not mapped")
	}
}

// sameJSON compares two JSON objects regardless of key order.
func sameJSON(a, b string) bool {
	var x, y any
	return json.Unmarshal([]byte(a), &x) == nil && json.Unmarshal([]byte(b), &y) == nil && reflect.DeepEqual(x, y)
}

func TestServerMessageJSON(t *testing.T) {
	tests := []struct {
		name string
		msg  ServerMessage
		want string
	}{
		{"view nil events", ServerMessage{Type: MsgView, View: &PlayerView{}}, `"events":[]`},
		{"view helper", ViewMessage(PlayerView{}, nil), `"events":[]`},
		{"view with events", ViewMessage(PlayerView{}, []Event{{Type: EventGameEnded}}), `"events":[{"type":"gameEnded"}]`},
		// presence, ack and error messages are tested in shared/backend/wire.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mustJSON(t, tt.msg)
			if !strings.Contains(s, tt.want) {
				t.Fatalf("got %s, want it to contain %s", s, tt.want)
			}
			if strings.Count(s, `"events"`) > 1 || strings.Count(s, `"online"`) > 1 {
				t.Fatalf("duplicate key in %s", s)
			}
		})
	}
	// Pointer marshal uses the same method.
	if s := mustJSON(t, &ServerMessage{Type: MsgView}); !strings.Contains(s, `"events":[]`) {
		t.Fatalf("pointer marshal = %s", s)
	}
	// Round trip.
	var back ServerMessage
	if err := json.Unmarshal([]byte(mustJSON(t, ViewMessage(PlayerView{GameID: "X"}, nil))), &back); err != nil {
		t.Fatal(err)
	}
	if back.Type != MsgView || back.View == nil || back.View.GameID != "X" || back.Events == nil {
		t.Fatalf("round trip = %+v", back)
	}
}

func TestClientMessageDecode(t *testing.T) {
	var m ClientMessage
	in := `{"type":"x","requestId":"4","mission":0,"seats":[1,2],"taskId":0,"card":{"suit":"pink","rank":3},` +
		`"position":"only","direction":"left","yes":false,"from":0,"to":2,"seat":0}`
	if err := json.Unmarshal([]byte(in), &m); err != nil {
		t.Fatal(err)
	}
	if m.Type != "x" || m.RequestID != "4" || m.Mission == nil || *m.Mission != 0 || !slices.Equal(m.Seats, []int{1, 2}) ||
		m.TaskID == nil || *m.TaskID != 0 || m.Card == nil || *m.Card != (Card{"pink", 3}) || m.Position != "only" ||
		m.Direction != "left" || m.Yes == nil || *m.Yes || m.From == nil || *m.From != 0 || m.To == nil || *m.To != 2 ||
		m.Seat == nil || *m.Seat != 0 {
		t.Fatalf("decoded %+v", m)
	}
	// Absent fields stay nil.
	var e ClientMessage
	if err := json.Unmarshal([]byte(`{"type":"endGame","requestId":"4"}`), &e); err != nil {
		t.Fatal(err)
	}
	if e.Type != CmdEndGame || e.Mission != nil || e.TaskID != nil || e.Card != nil || e.Yes != nil || e.From != nil || e.To != nil || e.Seat != nil || e.Seats != nil {
		t.Fatalf("decoded %+v", e)
	}
	// The command names are the ones of the protocol.
	want := []string{
		"setOptions", "selectMission", "startMission", "adjustTokens", "skipTokens", "answerPoll", "pick", "pickTask", "sendDistress",
		"skipDistress", "choosePassCard", "handover", "communicate", "playCard", "giveUp", "nextAttempt", "nextMission",
		"backToLobby", "endGame",
	}
	got := []string{
		CmdSetOptions, CmdSelectMission, CmdStartMission, CmdAdjustTokens, CmdSkipTokens, CmdAnswerPoll, CmdPick, CmdPickTask, CmdSendDistress,
		CmdSkipDistress, CmdChoosePassCard, CmdHandover, CmdCommunicate, CmdPlayCard, CmdGiveUp, CmdNextAttempt, CmdNextMission,
		CmdBackToLobby, CmdEndGame,
	}
	if !slices.Equal(got, want) {
		t.Fatalf("commands = %v", got)
	}
}

func TestErrorFor(t *testing.T) {
	var syntaxErr error
	{
		var v any
		syntaxErr = json.Unmarshal([]byte("{x"), &v)
	}
	var typeErr error
	{
		var v CreateGameRequest
		typeErr = json.Unmarshal([]byte(`{"name":5}`), &v)
	}
	tests := []struct {
		name       string
		err        error
		status     int
		code       string
		messageHas string
	}{
		{"bad request", BadRequest("missing %s", "name"), 400, "bad_request", "missing name"},
		{"json syntax", fmt.Errorf("decode: %w", syntaxErr), 400, "bad_request", ""},
		{"json type", typeErr, 400, "bad_request", ""},
		{"max bytes", &http.MaxBytesError{Limit: 4096}, 400, "bad_request", ""},
		{"rate limited", ErrRateLimited, 429, "rate_limited", ""},
		{"unauthorized", ports.ErrUnauthorized, 401, "unauthorized", ""},
		{"not host", domain.ErrNotHost, 403, "not_host", ""},
		{"not commander", fmt.Errorf("%w: only the commander", domain.ErrNotCommander), 403, "not_commander", "commander"},
		{"game not found", fmt.Errorf("load: %w", ports.ErrGameNotFound), 404, "game_not_found", ""},
		{"wrong phase", domain.ErrWrongPhase, 409, "wrong_phase", ""},
		{"wrong step", domain.ErrWrongStep, 409, "wrong_step", ""},
		{"game full", domain.ErrGameFull, 409, "game_full", ""},
		{"name taken", fmt.Errorf("%w: Anna", domain.ErrNameTaken), 409, "name_taken", "Anna"},
		{"duplicate player", domain.ErrDuplicatePlayer, 409, "duplicate_player", ""},
		{"not your turn", domain.ErrNotYourTurn, 409, "not_your_turn", ""},
		{"cannot leave", domain.ErrCannotLeave, 409, "cannot_leave", ""},
		{"not enough players", domain.ErrNotEnoughPlayers, 409, "not_enough_players", ""},
		{"cannot communicate", domain.ErrCannotCommunicate, 409, "cannot_communicate", ""},
		{"already communicated", domain.ErrAlreadyCommunicated, 409, "already_communicated", ""},
		{"already answered", domain.ErrAlreadyAnswered, 409, "already_answered", ""},
		{"not allowed", domain.ErrNotAllowed, 409, "not_allowed", ""},
		{"invalid result choice", domain.ErrInvalidResultChoice, 409, "invalid_result_choice", ""},
		{"conflict", ports.ErrVersionConflict, 409, "conflict", ""},
		{"invalid name", domain.ErrInvalidName, 422, "invalid_name", ""},
		{"unknown player", domain.ErrUnknownPlayer, 422, "unknown_player", ""},
		{"invalid mission", domain.ErrInvalidMission, 422, "invalid_mission", ""},
		{"card not in hand", domain.ErrCardNotInHand, 422, "card_not_in_hand", ""},
		{"must follow suit", domain.ErrMustFollowSuit, 422, "must_follow_suit", ""},
		{"invalid position", domain.ErrInvalidPosition, 422, "invalid_position", ""},
		{"invalid seat", domain.ErrInvalidSeat, 422, "invalid_seat", ""},
		{"invalid task", domain.ErrInvalidTask, 422, "invalid_task", ""},
		{"uneven distribution", domain.ErrUnevenDistribution, 422, "uneven_distribution", ""},
		{"cannot pass trump", domain.ErrCannotPassTrump, 422, "cannot_pass_trump", ""},
		{"invalid direction", domain.ErrInvalidDirection, 422, "invalid_direction", ""},
		{"invalid token move", domain.ErrInvalidTokenMove, 422, "invalid_token_move", ""},
		{"duplicate game id", ports.ErrDuplicateGameID, 500, "internal", "internal error"},
		{"other", errors.New("disk on fire"), 500, "internal", "internal error"},
		{"nil", nil, 500, "internal", "internal error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := ErrorFor(tt.err)
			if status != tt.status || body.Code != tt.code {
				t.Fatalf("ErrorFor(%v) = %d %+v, want %d %s", tt.err, status, body, tt.status, tt.code)
			}
			if body.Message == "" || !strings.Contains(body.Message, tt.messageHas) {
				t.Fatalf("message %q, want it to contain %q", body.Message, tt.messageHas)
			}
			if tt.code == "internal" && body.Message != InternalMessage {
				t.Fatalf("internal message leaked: %q", body.Message)
			}
		})
	}
}

// Every sentinel error of the domain has its own wire code (none falls through to 500).
func TestEveryDomainErrorIsMapped(t *testing.T) {
	sentinels := []error{
		domain.ErrInvalidName, domain.ErrNameTaken, domain.ErrGameFull, domain.ErrDuplicatePlayer, domain.ErrWrongPhase,
		domain.ErrWrongStep, domain.ErrNotHost, domain.ErrUnknownPlayer, domain.ErrNotYourTurn, domain.ErrCannotLeave,
		domain.ErrNotEnoughPlayers, domain.ErrInvalidMission, domain.ErrNotCommander, domain.ErrCardNotInHand,
		domain.ErrMustFollowSuit, domain.ErrCannotCommunicate, domain.ErrAlreadyCommunicated, domain.ErrInvalidPosition,
		domain.ErrInvalidSeat, domain.ErrInvalidTask, domain.ErrUnevenDistribution, domain.ErrCannotPassTrump,
		domain.ErrAlreadyAnswered, domain.ErrNotAllowed, domain.ErrInvalidDirection, domain.ErrInvalidTokenMove,
		domain.ErrInvalidResultChoice,
	}
	codes := map[string]error{}
	for _, e := range sentinels {
		status, body := ErrorFor(fmt.Errorf("wrapped: %w", e))
		if status >= 500 || body.Code == CodeInternal {
			t.Errorf("%v is not mapped", e)
		}
		if prev, dup := codes[body.Code]; dup {
			t.Errorf("code %s used by %v and %v", body.Code, prev, e)
		}
		codes[body.Code] = e
	}
}
