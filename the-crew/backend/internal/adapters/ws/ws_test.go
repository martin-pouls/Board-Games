package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/martin-pouls/shared/backend/wskit"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
	"github.com/martin-pouls/the-crew/backend/internal/core/ports"
)

const testGame domain.GameID = "ABCDE"

var testTokens = map[string]domain.PlayerID{"tok-host": "p_host", "tok-guest": "p_guest"}

// nameOf gives each player a distinct name so per-player views differ.
func nameOf(p domain.PlayerID) string {
	switch p {
	case "p_host":
		return "Hosta"
	case "p_guest":
		return "Guesto"
	default:
		return "Other"
	}
}

func stubView(g *domain.Game, p domain.PlayerID) (domain.PlayerView, error) {
	if _, ok := map[domain.PlayerID]bool{"p_host": true, "p_guest": true}[p]; !ok {
		return domain.PlayerView{}, domain.ErrUnknownPlayer
	}
	return domain.PlayerView{
		GameID:   g.ID,
		Revision: g.Revision,
		Phase:    domain.PhaseLobby,
		You:      p,
		HostID:   "p_host",
		Players:  []domain.PlayerSummary{{ID: p, Name: nameOf(p), IsYou: true}},
	}, nil
}

// fakeService is a canned ports.GameService. Successful mutations publish to
// the hub before returning, like the real service. err, when set, is
// returned by every mutating call.
type fakeService struct {
	hub *Hub

	mu      sync.Mutex
	err     error
	rev     int64
	calls   []string
	viewErr error
}

var _ ports.GameService = (*fakeService)(nil)

func (f *fakeService) Authenticate(_ context.Context, id domain.GameID, token string) (domain.PlayerID, error) {
	if id != testGame {
		return "", ports.ErrGameNotFound
	}
	if p, ok := testTokens[token]; ok {
		return p, nil
	}
	return "", ports.ErrUnauthorized
}

func (f *fakeService) View(_ context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	f.mu.Lock()
	err, rev := f.viewErr, f.rev
	f.mu.Unlock()
	if err != nil {
		return domain.PlayerView{}, err
	}
	return stubView(&domain.Game{ID: id, Revision: rev}, p)
}

func (f *fakeService) mutate(ctx context.Context, name string, id domain.GameID, evs ...domain.Event) error {
	f.mu.Lock()
	f.calls = append(f.calls, name)
	if f.err != nil {
		err := f.err
		f.mu.Unlock()
		return err
	}
	f.rev++
	g := &domain.Game{ID: id, Revision: f.rev}
	f.mu.Unlock()
	f.hub.Publish(ctx, ports.GameUpdate{Game: g, Events: evs})
	return nil
}

func (f *fakeService) callNames() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.calls)
}

func (f *fakeService) CreateGame(context.Context, string) (ports.Session, error) {
	return ports.Session{}, errors.New("not used")
}
func (f *fakeService) JoinGame(context.Context, domain.GameID, string) (ports.Session, error) {
	return ports.Session{}, errors.New("not used")
}
func (f *fakeService) Summary(context.Context, domain.GameID) (domain.GameSummary, error) {
	return domain.GameSummary{}, errors.New("not used")
}
func (f *fakeService) LeaveGame(context.Context, domain.GameID, domain.PlayerID) error {
	return errors.New("not used")
}
func (f *fakeService) EndGame(ctx context.Context, id domain.GameID, p domain.PlayerID) error {
	return f.mutate(ctx, "EndGame", id, domain.GameEnded{ByPlayerID: p})
}

// The gameplay commands record the call with its arguments, e.g. "AdjustTokens(1,2)".
func (f *fakeService) rec(ctx context.Context, id domain.GameID, name string, args ...any) error {
	call := name
	if len(args) > 0 {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = fmt.Sprint(a)
		}
		call += "(" + strings.Join(parts, ",") + ")"
	}
	return f.mutate(ctx, call, id)
}

func (f *fakeService) SetOptions(ctx context.Context, id domain.GameID, p domain.PlayerID, o domain.Options) error {
	return f.rec(ctx, id, "SetOptions", o.Distress)
}
func (f *fakeService) SelectMission(ctx context.Context, id domain.GameID, p domain.PlayerID, m int) error {
	return f.rec(ctx, id, "SelectMission", m)
}
func (f *fakeService) StartMission(ctx context.Context, id domain.GameID, p domain.PlayerID) error {
	return f.rec(ctx, id, "StartMission")
}
func (f *fakeService) AdjustTokens(ctx context.Context, id domain.GameID, p domain.PlayerID, from, to int) error {
	return f.rec(ctx, id, "AdjustTokens", from, to)
}
func (f *fakeService) SkipTokens(ctx context.Context, id domain.GameID, p domain.PlayerID) error {
	return f.rec(ctx, id, "SkipTokens")
}
func (f *fakeService) AnswerPoll(ctx context.Context, id domain.GameID, p domain.PlayerID, yes bool) error {
	return f.rec(ctx, id, "AnswerPoll", yes)
}
func (f *fakeService) PickSeats(ctx context.Context, id domain.GameID, p domain.PlayerID, seats []int) error {
	return f.rec(ctx, id, "PickSeats", seats)
}
func (f *fakeService) PickTask(ctx context.Context, id domain.GameID, p domain.PlayerID, task int) error {
	return f.rec(ctx, id, "PickTask", task)
}
func (f *fakeService) SendDistress(ctx context.Context, id domain.GameID, p domain.PlayerID, d domain.Direction) error {
	return f.rec(ctx, id, "SendDistress", d)
}
func (f *fakeService) SkipDistress(ctx context.Context, id domain.GameID, p domain.PlayerID) error {
	return f.rec(ctx, id, "SkipDistress")
}
func (f *fakeService) ChoosePassCard(ctx context.Context, id domain.GameID, p domain.PlayerID, c domain.Card) error {
	return f.rec(ctx, id, "ChoosePassCard", c)
}
func (f *fakeService) Handover(ctx context.Context, id domain.GameID, p domain.PlayerID, task, seat int) error {
	return f.rec(ctx, id, "Handover", task, seat)
}
func (f *fakeService) Communicate(ctx context.Context, id domain.GameID, p domain.PlayerID, c domain.Card, pos domain.Position) error {
	return f.rec(ctx, id, "Communicate", c, "["+string(pos)+"]")
}
func (f *fakeService) PlayCard(ctx context.Context, id domain.GameID, p domain.PlayerID, c domain.Card) error {
	return f.rec(ctx, id, "PlayCard", c)
}
func (f *fakeService) GiveUp(ctx context.Context, id domain.GameID, p domain.PlayerID) error {
	return f.rec(ctx, id, "GiveUp")
}
func (f *fakeService) NextAttempt(ctx context.Context, id domain.GameID, p domain.PlayerID) error {
	return f.rec(ctx, id, "NextAttempt")
}
func (f *fakeService) NextMission(ctx context.Context, id domain.GameID, p domain.PlayerID) error {
	return f.rec(ctx, id, "NextMission")
}
func (f *fakeService) BackToLobby(ctx context.Context, id domain.GameID, p domain.PlayerID) error {
	return f.rec(ctx, id, "BackToLobby")
}

// --- helpers ---

type env struct {
	srv *httptest.Server
	hub *Hub
	svc *fakeService
}

func newEnv(t *testing.T) *env {
	t.Helper()
	hub := NewHub(slog.New(slog.DiscardHandler))
	hub.viewFor = stubView
	svc := &fakeService{hub: hub}
	mux := http.NewServeMux()
	mux.Handle("GET /ws/games/{code}", NewHandler(hub, svc, Options{}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &env{srv: srv, hub: hub, svc: svc}
}

func (e *env) dial(t *testing.T, code, token string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	u := strings.Replace(e.srv.URL, "http", "ws", 1) + "/ws/games/" + code + "?token=" + token
	c, _, err := websocket.Dial(ctx, u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = c.CloseNow() })
	return c
}

type received struct {
	apidto.ServerMessage
	raw string
}

func read(t *testing.T, c *websocket.Conn) received {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var m apidto.ServerMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal %s: %v", data, err)
	}
	return received{m, string(data)}
}

func send(t *testing.T, c *websocket.Conn, msg string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func wantType(t *testing.T, m received, typ string) {
	t.Helper()
	if m.Type != typ {
		t.Fatalf("got %s, want type %q", m.raw, typ)
	}
}

func wantPresence(t *testing.T, m received, online ...string) {
	t.Helper()
	wantType(t, m, apidto.MsgPresence)
	if !slices.Equal(m.Online, online) {
		t.Fatalf("online = %v, want %v", m.Online, online)
	}
}

// connect dials and consumes the initial view + presence.
func (e *env) connect(t *testing.T, token string, online ...string) *websocket.Conn {
	t.Helper()
	c := e.dial(t, "abcde", token)
	v := read(t, c)
	wantType(t, v, apidto.MsgView)
	if v.View == nil || v.View.You != string(testTokens[token]) {
		t.Fatalf("initial view = %s", v.raw)
	}
	if !strings.Contains(v.raw, `"events":[]`) {
		t.Fatalf("initial view must carry events:[]: %s", v.raw)
	}
	wantPresence(t, read(t, c), online...)
	return c
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition not met in time")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// --- tests ---

func TestConnectSendsViewThenPresence(t *testing.T) {
	e := newEnv(t)
	e.connect(t, "tok-host", "p_host")
}

func TestAuthFailuresCloseWithCode(t *testing.T) {
	e := newEnv(t)
	tests := []struct {
		name, code, token string
		want              websocket.StatusCode
	}{
		{"bad token", "ABCDE", "nope", 4401},
		{"missing token", "ABCDE", "", 4401},
		{"unknown game", "ZZZZZ", "tok-host", 4404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := e.dial(t, tt.code, tt.token)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _, err := c.Read(ctx)
			if got := websocket.CloseStatus(err); got != tt.want {
				t.Fatalf("close status = %d (err %v), want %d", got, err, tt.want)
			}
		})
	}
}

func TestInitialViewErrorCloses(t *testing.T) {
	e := newEnv(t)
	e.svc.mu.Lock()
	e.svc.viewErr = ports.ErrGameNotFound
	e.svc.mu.Unlock()
	c := e.dial(t, "ABCDE", "tok-host")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := c.Read(ctx)
	if got := websocket.CloseStatus(err); got != 4404 {
		t.Fatalf("close status = %d, want 4404", got)
	}
	waitFor(t, func() bool { return len(e.hub.Online(testGame)) == 0 })
}

func TestPresenceOnConnectAndDisconnect(t *testing.T) {
	e := newEnv(t)
	host := e.connect(t, "tok-host", "p_host")
	guest := e.connect(t, "tok-guest", "p_guest", "p_host")
	wantPresence(t, read(t, host), "p_guest", "p_host")

	// A second tab of the same player: still two players online.
	tab := e.connect(t, "tok-guest", "p_guest", "p_host")
	wantPresence(t, read(t, host), "p_guest", "p_host")
	wantPresence(t, read(t, guest), "p_guest", "p_host")

	_ = tab.Close(websocket.StatusNormalClosure, "")
	wantPresence(t, read(t, host), "p_guest", "p_host")
	_ = guest.Close(websocket.StatusNormalClosure, "")
	wantPresence(t, read(t, host), "p_host")
}

func TestCommandBroadcastsViewThenAck(t *testing.T) {
	e := newEnv(t)
	host := e.connect(t, "tok-host", "p_host")
	guest := e.connect(t, "tok-guest", "p_guest", "p_host")
	wantPresence(t, read(t, host), "p_guest", "p_host")

	send(t, host, `{"type":"endGame","requestId":"4"}`)

	v := read(t, host)
	wantType(t, v, apidto.MsgView)
	if v.View.You != "p_host" || v.View.Players[0].Name != "Hosta" || v.View.Revision != 1 {
		t.Fatalf("host view = %s", v.raw)
	}
	if len(v.Events) != 1 || v.Events[0].Type != apidto.EventGameEnded || v.Events[0].ByPlayerID != "p_host" {
		t.Fatalf("events = %s", v.raw)
	}
	ack := read(t, host)
	wantType(t, ack, apidto.MsgAck)
	if ack.RequestID != "4" {
		t.Fatalf("ack = %s", ack.raw)
	}

	gv := read(t, guest)
	wantType(t, gv, apidto.MsgView)
	if gv.View.You != "p_guest" || gv.View.Players[0].Name != "Guesto" {
		t.Fatalf("guest view = %s", gv.raw)
	}
}

func TestAllCommands(t *testing.T) {
	tests := []struct {
		msg, call string
	}{
		{`{"type":"setOptions","requestId":"1","options":{"distress":true}}`, "SetOptions(true)"},
		{`{"type":"setOptions","requestId":"1","options":{"distress":false}}`, "SetOptions(false)"},
		{`{"type":"selectMission","requestId":"1","mission":7}`, "SelectMission(7)"},
		{`{"type":"selectMission","requestId":"1","mission":0}`, "SelectMission(0)"}, // 0 is present, the domain judges it
		{`{"type":"startMission","requestId":"1"}`, "StartMission"},
		{`{"type":"adjustTokens","requestId":"1","from":0,"to":2}`, "AdjustTokens(0,2)"},
		{`{"type":"skipTokens","requestId":"1"}`, "SkipTokens"},
		{`{"type":"answerPoll","requestId":"1","yes":false}`, "AnswerPoll(false)"},
		{`{"type":"answerPoll","requestId":"1","yes":true}`, "AnswerPoll(true)"},
		{`{"type":"pick","requestId":"1","seats":[2]}`, "PickSeats([2])"},
		{`{"type":"pick","requestId":"1","seats":[0,2]}`, "PickSeats([0 2])"},
		{`{"type":"pickTask","requestId":"1","taskId":0}`, "PickTask(0)"},
		{`{"type":"sendDistress","requestId":"1","direction":"left"}`, "SendDistress(left)"},
		{`{"type":"skipDistress","requestId":"1"}`, "SkipDistress"},
		{`{"type":"choosePassCard","requestId":"1","card":{"suit":"blue","rank":3}}`, "ChoosePassCard(blue 3)"},
		{`{"type":"handover","requestId":"1","taskId":1,"seat":0}`, "Handover(1,0)"},
		{`{"type":"communicate","requestId":"1","card":{"suit":"pink","rank":9},"position":"highest"}`, "Communicate(pink 9,[highest])"},
		{`{"type":"communicate","requestId":"1","card":{"suit":"pink","rank":9}}`, "Communicate(pink 9,[])"}, // dead zone
		{`{"type":"playCard","requestId":"1","card":{"suit":"trump","rank":4}}`, "PlayCard(trump 4)"},
		{`{"type":"giveUp","requestId":"1"}`, "GiveUp"},
		{`{"type":"nextAttempt","requestId":"1"}`, "NextAttempt"},
		{`{"type":"nextMission","requestId":"1"}`, "NextMission"},
		{`{"type":"backToLobby","requestId":"1"}`, "BackToLobby"},
		{`{"type":"endGame","requestId":"1"}`, "EndGame"},
	}
	for _, tt := range tests {
		t.Run(tt.call, func(t *testing.T) {
			e := newEnv(t)
			c := e.connect(t, "tok-host", "p_host")
			send(t, c, tt.msg)
			wantType(t, read(t, c), apidto.MsgView)
			ack := read(t, c)
			wantType(t, ack, apidto.MsgAck)
			if ack.RequestID != "1" {
				t.Fatalf("ack = %s", ack.raw)
			}
			if got := e.svc.callNames(); !slices.Equal(got, []string{tt.call}) {
				t.Fatalf("calls = %v", got)
			}
		})
	}
}

func TestCommandErrors(t *testing.T) {
	tests := []struct {
		name, msg, reqID, code string
		svcErr                 error
	}{
		{"not host", `{"type":"endGame","requestId":"8"}`, "8", apidto.CodeNotHost, domain.ErrNotHost},
		{"wrong phase", `{"type":"endGame","requestId":"9"}`, "9", apidto.CodeWrongPhase, fmt.Errorf("%w: only a lobby can be ended", domain.ErrWrongPhase)},
		{"internal", `{"type":"endGame","requestId":"8"}`, "8", apidto.CodeInternal, errors.New("boom")},
		{"unknown type", `{"type":"dance","requestId":"3"}`, "3", apidto.CodeBadRequest, nil},
		{"removed command", `{"type":"rematch","requestId":"3"}`, "3", apidto.CodeBadRequest, nil},
		{"malformed", `{"type":`, "", apidto.CodeBadRequest, nil},
		{"invalid card", `{"type":"playCard","requestId":"3","card":{"suit":"doves","rank":5}}`, "3", apidto.CodeBadRequest, nil},
		{"rank out of range", `{"type":"playCard","requestId":"3","card":{"suit":"trump","rank":5}}`, "3", apidto.CodeBadRequest, nil},
		{"wrong field type", `{"type":"selectMission","requestId":"3","mission":"7"}`, "3", apidto.CodeBadRequest, nil},
		{"must follow suit", `{"type":"playCard","requestId":"3","card":{"suit":"pink","rank":5}}`, "3", apidto.CodeMustFollowSuit, domain.ErrMustFollowSuit},
		{"not commander", `{"type":"skipTokens","requestId":"3"}`, "3", apidto.CodeNotCommander, domain.ErrNotCommander},
	}
	// A command without its required fields is bad_request and never reaches the service.
	for _, msg := range []string{
		`{"type":"setOptions","requestId":"3"}`,
		`{"type":"selectMission","requestId":"3"}`,
		`{"type":"adjustTokens","requestId":"3","from":1}`,
		`{"type":"adjustTokens","requestId":"3","to":1}`,
		`{"type":"answerPoll","requestId":"3"}`,
		`{"type":"pick","requestId":"3"}`,
		`{"type":"pick","requestId":"3","seats":[]}`,
		`{"type":"pickTask","requestId":"3"}`,
		`{"type":"sendDistress","requestId":"3"}`,
		`{"type":"choosePassCard","requestId":"3"}`,
		`{"type":"handover","requestId":"3","taskId":1}`,
		`{"type":"handover","requestId":"3","seat":1}`,
		`{"type":"communicate","requestId":"3","position":"only"}`,
		`{"type":"playCard","requestId":"3"}`,
	} {
		tests = append(tests, struct {
			name, msg, reqID, code string
			svcErr                 error
		}{"missing field " + msg, msg, "3", apidto.CodeBadRequest, nil})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			e.svc.mu.Lock()
			e.svc.err = tt.svcErr
			e.svc.mu.Unlock()
			c := e.connect(t, "tok-host", "p_host")
			send(t, c, tt.msg)
			m := read(t, c)
			wantType(t, m, apidto.MsgError)
			if m.RequestID != tt.reqID || m.Error == nil || m.Error.Code != tt.code {
				t.Fatalf("error = %s", m.raw)
			}
			if tt.code == apidto.CodeInternal && m.Error.Message != apidto.InternalMessage {
				t.Fatalf("internal message leaked: %s", m.raw)
			}
			// The connection stays usable.
			e.svc.mu.Lock()
			e.svc.err = nil
			e.svc.mu.Unlock()
			send(t, c, `{"type":"endGame","requestId":"x"}`)
			wantType(t, read(t, c), apidto.MsgView)
			wantType(t, read(t, c), apidto.MsgAck)
		})
	}
}

func TestReadLimitClosesConnection(t *testing.T) {
	e := newEnv(t)
	c := e.connect(t, "tok-host", "p_host")
	send(t, c, `{"type":"endGame","requestId":"`+strings.Repeat("x", 5000)+`"}`)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := c.Read(ctx)
	if got := websocket.CloseStatus(err); got != websocket.StatusMessageTooBig {
		t.Fatalf("close status = %d (err %v), want 1009", got, err)
	}
	waitFor(t, func() bool { return len(e.hub.Online(testGame)) == 0 })
}

// --- Publish (the transport itself is tested in shared/backend/wskit) ---

func TestPublishBuildsOneViewPerPlayer(t *testing.T) {
	e := newEnv(t)
	var calls atomic.Int32
	e.hub.viewFor = func(g *domain.Game, p domain.PlayerID) (domain.PlayerView, error) {
		calls.Add(1)
		return stubView(g, p)
	}
	host1 := e.connect(t, "tok-host", "p_host")
	host2 := e.connect(t, "tok-host", "p_host")
	wantPresence(t, read(t, host1), "p_host")
	guest := e.connect(t, "tok-guest", "p_guest", "p_host")
	wantPresence(t, read(t, host1), "p_guest", "p_host")
	wantPresence(t, read(t, host2), "p_guest", "p_host")

	calls.Store(0)
	e.hub.Publish(context.Background(), ports.GameUpdate{Game: &domain.Game{ID: testGame, Revision: 5}, Events: []domain.Event{domain.GameEnded{}}})

	for _, tc := range []struct {
		c    *websocket.Conn
		name string
	}{{host1, "Hosta"}, {host2, "Hosta"}, {guest, "Guesto"}} {
		m := read(t, tc.c)
		wantType(t, m, apidto.MsgView)
		if m.View.Players[0].Name != tc.name || m.View.Revision != 5 {
			t.Fatalf("view = %s", m.raw)
		}
		if len(m.Events) != 1 || m.Events[0].Type != apidto.EventGameEnded {
			t.Fatalf("events = %s", m.raw)
		}
	}
	if n := calls.Load(); n != 2 {
		t.Fatalf("viewFor called %d times, want 2", n)
	}
}

func TestPublishWithoutGameIsIgnored(t *testing.T) {
	e := newEnv(t)
	c := e.connect(t, "tok-host", "p_host")
	e.hub.Publish(context.Background(), ports.GameUpdate{})
	e.hub.Publish(context.Background(), ports.GameUpdate{Game: &domain.Game{ID: testGame, Revision: 1}})
	m := read(t, c)
	wantType(t, m, apidto.MsgView)
	if m.View.Revision != 1 {
		t.Fatalf("first message after publishes = %s", m.raw)
	}
}

func TestPublishKicksPlayerWhoLeft(t *testing.T) {
	e := newEnv(t)
	host := e.connect(t, "tok-host", "p_host")
	guest := e.connect(t, "tok-guest", "p_guest", "p_host")
	wantPresence(t, read(t, host), "p_guest", "p_host")

	// The guest is no longer seated: their view fails with ErrUnknownPlayer.
	e.hub.viewFor = func(g *domain.Game, p domain.PlayerID) (domain.PlayerView, error) {
		if p == "p_guest" {
			return domain.PlayerView{}, domain.ErrUnknownPlayer
		}
		return stubView(g, p)
	}
	e.hub.Publish(context.Background(), ports.GameUpdate{Game: &domain.Game{ID: testGame, Revision: 2}, Events: []domain.Event{domain.PlayerLeft{PlayerID: "p_guest", HostID: "p_host"}}})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := guest.Read(ctx)
	if got := websocket.CloseStatus(err); got != wskit.CloseUnauthorized {
		t.Fatalf("guest close status = %d (err %v), want 4401", got, err)
	}
	wantType(t, read(t, host), apidto.MsgView)
	wantPresence(t, read(t, host), "p_host")
}
