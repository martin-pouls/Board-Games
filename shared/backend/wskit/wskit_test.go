package wskit

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

	"github.com/martin-pouls/shared/backend/wire"
)

const testGame = "ABCDE"

var testTokens = map[string]string{"tok-host": "p_host", "tok-guest": "p_guest"}

// testView and testEvent stand in for a game's view and events.
type testView struct {
	You      string `json:"you"`
	Hand     int    `json:"hand"`
	Revision int64  `json:"revision"`
}

type testEvent struct {
	Type string `json:"type"`
}

type message = wire.ServerMessage[testView, testEvent]

var (
	errUnauthorized = errors.New("unauthorized")
	errNotFound     = errors.New("game not found")
	errIllegal      = errors.New("illegal move")
	errBadRequest   = errors.New("bad request")
)

// handValue gives each player a distinct hand so per-player views differ.
func handValue(p string) int {
	switch p {
	case "p_host":
		return 3
	case "p_guest":
		return 7
	default:
		return 9
	}
}

// viewFor builds p's view message; players other than host and guest are gone (they left).
func viewFor(p string, rev int64, events []testEvent) (payload []byte, gone bool) {
	if p != "p_host" && p != "p_guest" {
		return nil, true
	}
	b, _ := json.Marshal(wire.ViewMessage(testView{You: p, Hand: handValue(p), Revision: rev}, events))
	return b, false
}

// fakeBackend is a canned game: "act" commands bump the revision and broadcast through the hub
// before returning, like a real service. err, when set, fails every "act".
type fakeBackend struct {
	hub *Hub

	mu      sync.Mutex
	err     error
	rev     int64
	calls   []string
	viewErr error
}

func (b *fakeBackend) Authenticate(_ context.Context, gameID, token string) (string, error) {
	if gameID != testGame {
		return "", errNotFound
	}
	if p, ok := testTokens[token]; ok {
		return p, nil
	}
	return "", errUnauthorized
}

func (b *fakeBackend) Hello(_ context.Context, _ string, p string) ([]byte, error) {
	b.mu.Lock()
	err, rev := b.viewErr, b.rev
	b.mu.Unlock()
	if err != nil {
		return nil, err
	}
	payload, _ := viewFor(p, rev, nil)
	return payload, nil
}

func (b *fakeBackend) Handle(_ context.Context, gameID, _ string, cmd wire.Command, data []byte) error {
	if cmd.Type != "act" {
		return fmt.Errorf("%w: unknown message type %q", errBadRequest, cmd.Type)
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("%w: %v", errBadRequest, err)
	}
	b.mu.Lock()
	b.calls = append(b.calls, body.Name)
	if b.err != nil {
		err := b.err
		b.mu.Unlock()
		return err
	}
	b.rev++
	rev := b.rev
	b.mu.Unlock()
	b.hub.Broadcast(gameID, func(p string) ([]byte, bool) {
		return viewFor(p, rev, []testEvent{{Type: body.Name}})
	})
	return nil
}

func (b *fakeBackend) MapError(err error) (int, wire.ErrorBody) {
	for _, e := range []struct {
		err    error
		status int
		code   string
	}{
		{errUnauthorized, http.StatusUnauthorized, wire.CodeUnauthorized},
		{errNotFound, http.StatusNotFound, wire.CodeGameNotFound},
		{errIllegal, http.StatusUnprocessableEntity, "illegal"},
		{errBadRequest, http.StatusBadRequest, wire.CodeBadRequest},
	} {
		if errors.Is(err, e.err) {
			return e.status, wire.ErrorBody{Code: e.code, Message: err.Error()}
		}
	}
	return http.StatusInternalServerError, wire.ErrorBody{Code: wire.CodeInternal, Message: wire.InternalMessage}
}

func (b *fakeBackend) callNames() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.calls)
}

// --- helpers ---

type env struct {
	srv     *httptest.Server
	hub     *Hub
	backend *fakeBackend
}

func newEnv(t *testing.T) *env {
	t.Helper()
	hub := NewHub(slog.New(slog.DiscardHandler))
	backend := &fakeBackend{hub: hub}
	mux := http.NewServeMux()
	mux.Handle("GET /ws/games/{code}", NewHandler(hub, backend, Options{}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &env{srv: srv, hub: hub, backend: backend}
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
	message
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
	var m message
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
	wantType(t, m, wire.MsgPresence)
	if !slices.Equal(m.Online, online) {
		t.Fatalf("online = %v, want %v", m.Online, online)
	}
}

// connect dials and consumes the initial view + presence.
func (e *env) connect(t *testing.T, token string, online ...string) *websocket.Conn {
	t.Helper()
	c := e.dial(t, "abcde", token)
	v := read(t, c)
	wantType(t, v, wire.MsgView)
	if v.View == nil || v.View.You != testTokens[token] {
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

// --- endpoint tests ---

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
		{"bad token", "ABCDE", "nope", CloseUnauthorized},
		{"missing token", "ABCDE", "", CloseUnauthorized},
		{"unknown game", "ZZZZZ", "tok-host", CloseGameNotFound},
		{"unknown game without token", "ZZZZZ", "", CloseGameNotFound},
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

func TestHelloErrorCloses(t *testing.T) {
	for _, tt := range []struct {
		err  error
		want websocket.StatusCode
	}{
		{errNotFound, CloseGameNotFound},
		{errors.New("boom"), websocket.StatusInternalError},
	} {
		t.Run(tt.err.Error(), func(t *testing.T) {
			e := newEnv(t)
			e.backend.mu.Lock()
			e.backend.viewErr = tt.err
			e.backend.mu.Unlock()
			c := e.dial(t, "ABCDE", "tok-host")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _, err := c.Read(ctx)
			if got := websocket.CloseStatus(err); got != tt.want {
				t.Fatalf("close status = %d, want %d", got, tt.want)
			}
			waitFor(t, func() bool { return len(e.hub.Online(testGame)) == 0 })
		})
	}
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

	send(t, host, `{"type":"act","requestId":"4","name":"played"}`)

	v := read(t, host)
	wantType(t, v, wire.MsgView)
	if v.View.You != "p_host" || v.View.Hand != 3 || v.View.Revision != 1 {
		t.Fatalf("host view = %s", v.raw)
	}
	if len(v.Events) != 1 || v.Events[0].Type != "played" {
		t.Fatalf("events = %s", v.raw)
	}
	ack := read(t, host)
	wantType(t, ack, wire.MsgAck)
	if ack.RequestID != "4" {
		t.Fatalf("ack = %s", ack.raw)
	}

	gv := read(t, guest)
	wantType(t, gv, wire.MsgView)
	if gv.View.You != "p_guest" || gv.View.Hand != 7 {
		t.Fatalf("guest view = %s", gv.raw)
	}
	if got := e.backend.callNames(); !slices.Equal(got, []string{"played"}) {
		t.Fatalf("calls = %v", got)
	}
}

func TestCommandErrors(t *testing.T) {
	tests := []struct {
		name, msg, reqID, code string
		err                    error
	}{
		{"game error", `{"type":"act","requestId":"7"}`, "7", "illegal", fmt.Errorf("%w: nope", errIllegal)},
		{"internal", `{"type":"act","requestId":"8"}`, "8", wire.CodeInternal, errors.New("boom")},
		{"unknown type", `{"type":"dance","requestId":"3"}`, "3", wire.CodeBadRequest, nil},
		{"malformed", `{"type":`, "", wire.CodeBadRequest, nil},
		{"wrong envelope types", `{"type":1,"requestId":"5"}`, "", wire.CodeBadRequest, nil},
		{"bad body", `{"type":"act","requestId":"5","name":7}`, "5", wire.CodeBadRequest, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			e.backend.mu.Lock()
			e.backend.err = tt.err
			e.backend.mu.Unlock()
			c := e.connect(t, "tok-host", "p_host")
			send(t, c, tt.msg)
			m := read(t, c)
			wantType(t, m, wire.MsgError)
			if m.RequestID != tt.reqID || m.Error == nil || m.Error.Code != tt.code {
				t.Fatalf("error = %s", m.raw)
			}
			if tt.code == wire.CodeInternal && m.Error.Message != wire.InternalMessage {
				t.Fatalf("internal message leaked: %s", m.raw)
			}
			// The connection stays usable.
			e.backend.mu.Lock()
			e.backend.err = nil
			e.backend.mu.Unlock()
			send(t, c, `{"type":"act","requestId":"x","name":"again"}`)
			wantType(t, read(t, c), wire.MsgView)
			wantType(t, read(t, c), wire.MsgAck)
		})
	}
}

func TestReadLimitClosesConnection(t *testing.T) {
	e := newEnv(t)
	c := e.connect(t, "tok-host", "p_host")
	send(t, c, `{"type":"act","requestId":"`+strings.Repeat("x", 5000)+`"}`)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := c.Read(ctx)
	if got := websocket.CloseStatus(err); got != websocket.StatusMessageTooBig {
		t.Fatalf("close status = %d (err %v), want 1009", got, err)
	}
	waitFor(t, func() bool { return len(e.hub.Online(testGame)) == 0 })
}

func TestSlowConsumerClosedOverNetwork(t *testing.T) {
	e := newEnv(t)
	c := e.connect(t, "tok-host", "p_host")
	conns := e.hub.snapshot(testGame)
	if len(conns) != 1 {
		t.Fatalf("conns = %d", len(conns))
	}
	conns[0].kick(websocket.StatusPolicyViolation, "slow consumer")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		_, _, err := c.Read(ctx)
		if err != nil {
			if got := websocket.CloseStatus(err); got != websocket.StatusPolicyViolation {
				t.Fatalf("close status = %d (err %v), want 1008", got, err)
			}
			break
		}
	}
	waitFor(t, func() bool { return len(e.hub.Online(testGame)) == 0 })
}

// --- hub unit tests (no network) ---

// fakeConn registers a conn without a socket or writer goroutine.
func fakeConn(h *Hub, game, p string) *conn {
	c := newConn(nil, game, p, h.logger)
	h.register(c)
	return c
}

func drain(c *conn) []message {
	var out []message
	for {
		select {
		case b := <-c.send:
			var m message
			_ = json.Unmarshal(b, &m)
			out = append(out, m)
		default:
			return out
		}
	}
}

func TestBroadcastBuildsOnePayloadPerPlayer(t *testing.T) {
	h := NewHub(nil)
	var calls atomic.Int32
	build := func(p string) ([]byte, bool) {
		calls.Add(1)
		return viewFor(p, 5, []testEvent{{Type: "started"}})
	}
	host1, host2, guest := fakeConn(h, testGame, "p_host"), fakeConn(h, testGame, "p_host"), fakeConn(h, testGame, "p_guest")
	other := fakeConn(h, "OTHER", "p_host")

	h.Broadcast(testGame, build)

	if n := calls.Load(); n != 2 {
		t.Fatalf("build called %d times, want 2", n)
	}
	for _, tc := range []struct {
		c    *conn
		hand int
	}{{host1, 3}, {host2, 3}, {guest, 7}} {
		ms := drain(tc.c)
		if len(ms) != 1 || ms[0].Type != wire.MsgView || ms[0].View.Hand != tc.hand || ms[0].View.Revision != 5 {
			t.Fatalf("conn %s got %+v", tc.c.playerID, ms)
		}
		if len(ms[0].Events) != 1 || ms[0].Events[0].Type != "started" {
			t.Fatalf("events = %+v", ms[0].Events)
		}
	}
	if ms := drain(other); len(ms) != 0 {
		t.Fatalf("other room got %d messages", len(ms))
	}
}

func TestBroadcastToEmptyRoomBuildsNothing(t *testing.T) {
	h := NewHub(nil)
	called := false
	h.Broadcast(testGame, func(string) ([]byte, bool) { called = true; return nil, false })
	if called {
		t.Fatal("build called for an empty room")
	}
}

func TestBroadcastNilPayloadSendsNothing(t *testing.T) {
	h := NewHub(nil)
	c := fakeConn(h, testGame, "p_host")
	h.Broadcast(testGame, func(string) ([]byte, bool) { return nil, false })
	if ms := drain(c); len(ms) != 0 {
		t.Fatalf("got %+v", ms)
	}
	select {
	case <-c.closing:
		t.Fatal("conn was kicked")
	default:
	}
}

func TestSlowConsumerDoesNotBlockBroadcast(t *testing.T) {
	h := NewHub(nil)
	slow := fakeConn(h, testGame, "p_guest") // no writer: its buffer never drains
	fast := fakeConn(h, testGame, "p_host")

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := range sendBuffer + 10 {
			h.Broadcast(testGame, func(p string) ([]byte, bool) { return viewFor(p, int64(i), nil) })
			drain(fast)
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Broadcast blocked on a slow consumer")
	}
	select {
	case <-slow.closing:
	default:
		t.Fatal("slow consumer was not kicked")
	}
	if slow.closeCode != websocket.StatusPolicyViolation || slow.closeReason != "slow consumer" {
		t.Fatalf("close = %d %q", slow.closeCode, slow.closeReason)
	}
	select {
	case <-fast.closing:
		t.Fatal("fast consumer was kicked")
	default:
	}
}

func TestBroadcastKicksPlayerWhoLeft(t *testing.T) {
	h := NewHub(nil)
	gone1, gone2, host := fakeConn(h, testGame, "p_stranger"), fakeConn(h, testGame, "p_stranger"), fakeConn(h, testGame, "p_host")

	h.Broadcast(testGame, func(p string) ([]byte, bool) { return viewFor(p, 2, nil) })

	for _, c := range []*conn{gone1, gone2} {
		select {
		case <-c.closing:
		default:
			t.Fatalf("conn of the player who left was not kicked")
		}
		if c.closeCode != CloseUnauthorized || c.closeReason != wire.CodeUnauthorized {
			t.Fatalf("close = %d %q, want 4401 unauthorized", c.closeCode, c.closeReason)
		}
		if ms := drain(c); len(ms) != 0 {
			t.Fatalf("leaver received %+v", ms)
		}
	}
	if ms := drain(host); len(ms) != 1 || ms[0].Type != wire.MsgView {
		t.Fatalf("host got %+v", ms)
	}
	select {
	case <-host.closing:
		t.Fatal("remaining player was kicked")
	default:
	}
}

func TestOnline(t *testing.T) {
	h := NewHub(nil)
	fakeConn(h, testGame, "p_b")
	fakeConn(h, testGame, "p_a")
	fakeConn(h, testGame, "p_b")
	fakeConn(h, "OTHER", "p_c")
	if got := h.Online(testGame); !slices.Equal(got, []string{"p_a", "p_b"}) {
		t.Fatalf("Online = %v", got)
	}
}

// recordingHub returns a hub and the text of everything it logs.
func recordingHub() (*Hub, *strings.Builder) {
	var sb strings.Builder
	return NewHub(slog.New(slog.NewTextHandler(&sb, nil))), &sb
}

func TestPublishViewsSendsEachPlayerTheirView(t *testing.T) {
	h := NewHub(nil)
	var calls atomic.Int32
	host1, host2, guest := fakeConn(h, testGame, "p_host"), fakeConn(h, testGame, "p_host"), fakeConn(h, testGame, "p_guest")
	other := fakeConn(h, "OTHER", "p_host")

	h.PublishViews(testGame, errIllegal, func(p string) (any, error) {
		calls.Add(1)
		return wire.ViewMessage(testView{You: p, Hand: handValue(p), Revision: 4}, []testEvent{{Type: "played"}}), nil
	})

	if n := calls.Load(); n != 2 {
		t.Fatalf("view called %d times, want 2", n)
	}
	for _, tc := range []struct {
		c    *conn
		hand int
	}{{host1, 3}, {host2, 3}, {guest, 7}} {
		ms := drain(tc.c)
		if len(ms) != 1 || ms[0].Type != wire.MsgView || ms[0].View.You != tc.c.playerID || ms[0].View.Hand != tc.hand || ms[0].View.Revision != 4 {
			t.Fatalf("conn %s got %+v", tc.c.playerID, ms)
		}
		if len(ms[0].Events) != 1 || ms[0].Events[0].Type != "played" {
			t.Fatalf("events = %+v", ms[0].Events)
		}
	}
	if ms := drain(other); len(ms) != 0 {
		t.Fatalf("other room got %d messages", len(ms))
	}
}

func TestPublishViewsKicksGonePlayer(t *testing.T) {
	h := NewHub(nil)
	left, host := fakeConn(h, testGame, "p_stranger"), fakeConn(h, testGame, "p_host")

	h.PublishViews(testGame, errNotFound, func(p string) (any, error) {
		if p == "p_stranger" {
			return nil, fmt.Errorf("view of %s: %w", p, errNotFound)
		}
		return wire.ViewMessage(testView{You: p}, []testEvent(nil)), nil
	})

	select {
	case <-left.closing:
	default:
		t.Fatal("conn of the player who left was not kicked")
	}
	if left.closeCode != CloseUnauthorized || left.closeReason != wire.CodeUnauthorized {
		t.Fatalf("close = %d %q, want 4401 unauthorized", left.closeCode, left.closeReason)
	}
	if ms := drain(left); len(ms) != 0 {
		t.Fatalf("leaver received %+v", ms)
	}
	if ms := drain(host); len(ms) != 1 || ms[0].Type != wire.MsgView {
		t.Fatalf("host got %+v", ms)
	}
	select {
	case <-host.closing:
		t.Fatal("remaining player was kicked")
	default:
	}
}

func TestPublishViewsLogsErrorsAndSkipsThatPlayer(t *testing.T) {
	for name, tc := range map[string]struct {
		gone error
		view func(string) (any, error)
		log  string
	}{
		"view error":    {errNotFound, func(string) (any, error) { return nil, errIllegal }, "build view"},
		"nil gone":      {nil, func(string) (any, error) { return nil, errNotFound }, "build view"},
		"marshal error": {errNotFound, func(string) (any, error) { return make(chan int), nil }, "marshal view"},
	} {
		t.Run(name, func(t *testing.T) {
			h, logs := recordingHub()
			c := fakeConn(h, testGame, "p_host")

			h.PublishViews(testGame, tc.gone, tc.view)

			if ms := drain(c); len(ms) != 0 {
				t.Fatalf("got %+v", ms)
			}
			select {
			case <-c.closing:
				t.Fatal("conn was kicked")
			default:
			}
			if !strings.Contains(logs.String(), tc.log) || !strings.Contains(logs.String(), "p_host") {
				t.Fatalf("log = %q, want %q for p_host", logs.String(), tc.log)
			}
		})
	}
}

type playerID string

func TestOnlineAs(t *testing.T) {
	h := NewHub(nil)
	fakeConn(h, testGame, "p_b")
	fakeConn(h, testGame, "p_a")
	if got := OnlineAs[playerID](h, testGame); !slices.Equal(got, []playerID{"p_a", "p_b"}) {
		t.Fatalf("OnlineAs = %v", got)
	}
	if got := OnlineAs[playerID](h, "EMPTY"); got != nil {
		t.Fatalf("OnlineAs of an empty room = %v", got)
	}
}
