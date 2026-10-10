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

	"github.com/martin-pouls/regicide/backend/internal/adapters/apidto"
	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/ports"
)

const testGame domain.GameID = "ABCDE"

var testTokens = map[string]domain.PlayerID{"tok-host": "p_host", "tok-guest": "p_guest"}

// handValue gives each player a distinct card so per-player views differ.
func handValue(p domain.PlayerID) int {
	switch p {
	case "p_host":
		return 3
	case "p_guest":
		return 7
	default:
		return 9
	}
}

func stubView(g *domain.Game, p domain.PlayerID) (domain.PlayerView, error) {
	if _, ok := map[domain.PlayerID]bool{"p_host": true, "p_guest": true}[p]; !ok {
		return domain.PlayerView{}, domain.ErrUnknownPlayer
	}
	return domain.PlayerView{
		GameID:   g.ID,
		Revision: g.Revision,
		Phase:    domain.PhasePlaying,
		You:      p,
		Hand:     []domain.Card{domain.NewCard(domain.SuitHearts, domain.Rank(handValue(p)))},
	}, nil
}

// fakeService is a canned ports.GameService. Successful mutations publish to
// the hub before returning, like the real service. err, when set, is
// returned by every mutating call.
type fakeService struct {
	hub *Hub

	mu         sync.Mutex
	err        error
	rev        int64
	calls      []string
	lastFirst  domain.PlayerID
	lastNext   domain.PlayerID
	lastCards  []domain.Card
	lastSignal domain.JesterSignal
	viewErr    error
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

func (f *fakeService) mutate(ctx context.Context, name string, id domain.GameID, p domain.PlayerID, evs ...domain.Event) (domain.PlayerView, error) {
	f.mu.Lock()
	f.calls = append(f.calls, name)
	if f.err != nil {
		err := f.err
		f.mu.Unlock()
		return domain.PlayerView{}, err
	}
	f.rev++
	g := &domain.Game{ID: id, Revision: f.rev}
	f.mu.Unlock()
	f.hub.Publish(ctx, ports.GameUpdate{Game: g, Events: evs})
	return stubView(g, p)
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
func (f *fakeService) StartGame(ctx context.Context, id domain.GameID, p, first domain.PlayerID) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastFirst = first
	f.mu.Unlock()
	return f.mutate(ctx, "StartGame", id, p, domain.GameStarted{FirstPlayerID: first})
}
func (f *fakeService) Play(ctx context.Context, id domain.GameID, p domain.PlayerID, cards []domain.Card) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastCards = cards
	f.mu.Unlock()
	return f.mutate(ctx, "Play", id, p, domain.CardsPlayed{PlayerID: p, Cards: cards})
}
func (f *fakeService) Yield(ctx context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return f.mutate(ctx, "Yield", id, p, domain.Yielded{PlayerID: p})
}
func (f *fakeService) Discard(ctx context.Context, id domain.GameID, p domain.PlayerID, cards []domain.Card) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastCards = cards
	f.mu.Unlock()
	return f.mutate(ctx, "Discard", id, p, domain.CardsDiscarded{PlayerID: p, Cards: cards})
}
func (f *fakeService) ChooseNext(ctx context.Context, id domain.GameID, p, next domain.PlayerID) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastNext = next
	f.mu.Unlock()
	return f.mutate(ctx, "ChooseNext", id, p, domain.NextPlayerChosen{ByPlayerID: p, PlayerID: next})
}
func (f *fakeService) UseSoloJester(ctx context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return f.mutate(ctx, "UseSoloJester", id, p, domain.SoloJesterUsed{PlayerID: p})
}
func (f *fakeService) SetJesterSignal(ctx context.Context, id domain.GameID, p domain.PlayerID, s domain.JesterSignal) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastSignal = s
	f.mu.Unlock()
	return f.mutate(ctx, "SetJesterSignal", id, p, domain.JesterSignalSet{PlayerID: p, Signal: s})
}
func (f *fakeService) GiveUp(ctx context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return f.mutate(ctx, "GiveUp", id, p, domain.GameLost{Reason: domain.LossGaveUp, PlayerID: p})
}
func (f *fakeService) EndGame(ctx context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return f.mutate(ctx, "EndGame", id, p, domain.GameEnded{ByPlayerID: p})
}
func (f *fakeService) Rematch(ctx context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return f.mutate(ctx, "Rematch", id, p, domain.RematchCreated{GameID: "NEXT2"})
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

	send(t, host, `{"type":"play","requestId":"4","cards":[{"suit":"clubs","rank":8},{"suit":"","rank":0}]}`)

	v := read(t, host)
	wantType(t, v, apidto.MsgView)
	if v.View.You != "p_host" || v.View.Hand[0].Rank != 3 || v.View.Revision != 1 {
		t.Fatalf("host view = %s", v.raw)
	}
	if len(v.Events) != 1 || v.Events[0].Type != apidto.EventCardsPlayed || len(v.Events[0].Cards) != 2 || v.Events[0].Cards[0].Rank != 8 {
		t.Fatalf("events = %s", v.raw)
	}
	ack := read(t, host)
	wantType(t, ack, apidto.MsgAck)
	if ack.RequestID != "4" {
		t.Fatalf("ack = %s", ack.raw)
	}

	gv := read(t, guest)
	wantType(t, gv, apidto.MsgView)
	if gv.View.You != "p_guest" || gv.View.Hand[0].Rank != 7 {
		t.Fatalf("guest view = %s", gv.raw)
	}
	e.svc.mu.Lock()
	cards := e.svc.lastCards
	e.svc.mu.Unlock()
	if len(cards) != 2 || cards[0] != domain.NewCard(domain.SuitClubs, 8) || !cards[1].IsJester() {
		t.Fatalf("service got cards %v", cards)
	}
}

func TestAllCommands(t *testing.T) {
	tests := []struct {
		msg, call string
	}{
		{`{"type":"startGame","requestId":"1"}`, "StartGame"},
		{`{"type":"startGame","requestId":"1","firstPlayerId":"p_guest"}`, "StartGame"},
		{`{"type":"play","requestId":"1","cards":[{"suit":"hearts","rank":5}]}`, "Play"},
		{`{"type":"yield","requestId":"1"}`, "Yield"},
		{`{"type":"discard","requestId":"1","cards":[{"suit":"spades","rank":10}]}`, "Discard"},
		{`{"type":"discard","requestId":"1","cards":[]}`, "Discard"},
		{`{"type":"chooseNext","requestId":"1","playerId":"p_guest"}`, "ChooseNext"},
		{`{"type":"useSoloJester","requestId":"1"}`, "UseSoloJester"},
		{`{"type":"setSignal","requestId":"1","signal":"goodPlay"}`, "SetJesterSignal"},
		{`{"type":"setSignal","requestId":"1","signal":"none"}`, "SetJesterSignal"},
		{`{"type":"giveUp","requestId":"1"}`, "GiveUp"},
		{`{"type":"endGame","requestId":"1"}`, "EndGame"},
		{`{"type":"rematch","requestId":"1"}`, "Rematch"},
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

func TestRematchEventOnWire(t *testing.T) {
	e := newEnv(t)
	c := e.connect(t, "tok-host", "p_host")
	send(t, c, `{"type":"rematch","requestId":"9"}`)
	v := read(t, c)
	if len(v.Events) != 1 || v.Events[0].Type != apidto.EventRematchCreated || v.Events[0].GameID != "NEXT2" {
		t.Fatalf("rematch view = %s", v.raw)
	}
}

func TestCommandErrors(t *testing.T) {
	tests := []struct {
		name, msg, reqID, code string
		svcErr                 error
	}{
		{"domain error", `{"type":"play","requestId":"7","cards":[{"suit":"hearts","rank":5}]}`, "7", apidto.CodeIllegalPlay, fmt.Errorf("%w: mixedRanks", domain.ErrIllegalPlay)},
		{"not host", `{"type":"startGame","requestId":"8"}`, "8", apidto.CodeNotHost, domain.ErrNotHost},
		{"rematch exists", `{"type":"rematch","requestId":"8"}`, "8", apidto.CodeRematchExists, domain.ErrRematchExists},
		{"yield not allowed", `{"type":"yield","requestId":"9"}`, "9", apidto.CodeYieldNotAllowed, domain.ErrYieldNotAllowed},
		{"not your turn", `{"type":"chooseNext","requestId":"9","playerId":"p_guest"}`, "9", apidto.CodeNotYourTurn, domain.ErrNotYourTurn},
		{"internal", `{"type":"yield","requestId":"8"}`, "8", apidto.CodeInternal, errors.New("boom")},
		{"unknown type", `{"type":"dance","requestId":"3"}`, "3", apidto.CodeBadRequest, nil},
		{"removed command", `{"type":"playCard","requestId":"3","card":1,"pile":0}`, "3", apidto.CodeBadRequest, nil},
		{"malformed", `{"type":`, "", apidto.CodeBadRequest, nil},
		{"play missing cards", `{"type":"play","requestId":"5"}`, "5", apidto.CodeBadRequest, nil},
		{"play null cards", `{"type":"play","requestId":"5","cards":null}`, "5", apidto.CodeBadRequest, nil},
		{"discard missing cards", `{"type":"discard","requestId":"5"}`, "5", apidto.CodeBadRequest, nil},
		{"invalid card", `{"type":"play","requestId":"5","cards":[{"suit":"hearts","rank":99}]}`, "5", apidto.CodeBadRequest, nil},
		{"missing signal", `{"type":"setSignal","requestId":"5"}`, "5", apidto.CodeBadRequest, nil},
		{"missing playerId", `{"type":"chooseNext","requestId":"5"}`, "5", apidto.CodeBadRequest, nil},
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
			send(t, c, `{"type":"yield","requestId":"x"}`)
			wantType(t, read(t, c), apidto.MsgView)
			wantType(t, read(t, c), apidto.MsgAck)
		})
	}
}

func TestReadLimitClosesConnection(t *testing.T) {
	e := newEnv(t)
	c := e.connect(t, "tok-host", "p_host")
	send(t, c, `{"type":"yield","requestId":"`+strings.Repeat("x", 5000)+`"}`)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := c.Read(ctx)
	if got := websocket.CloseStatus(err); got != websocket.StatusMessageTooBig {
		t.Fatalf("close status = %d (err %v), want 1009", got, err)
	}
	waitFor(t, func() bool { return len(e.hub.Online(testGame)) == 0 })
}

// --- hub unit tests (no network) ---

// fakeConn registers a conn without a socket or writer goroutine.
func fakeConn(h *Hub, p domain.PlayerID) *conn {
	c := newConn(nil, testGame, p, h.logger)
	h.register(c)
	return c
}

func drain(c *conn) []apidto.ServerMessage {
	var out []apidto.ServerMessage
	for {
		select {
		case b := <-c.send:
			var m apidto.ServerMessage
			_ = json.Unmarshal(b, &m)
			out = append(out, m)
		default:
			return out
		}
	}
}

func TestPublishBuildsOneViewPerPlayer(t *testing.T) {
	h := NewHub(nil)
	var calls atomic.Int32
	h.viewFor = func(g *domain.Game, p domain.PlayerID) (domain.PlayerView, error) {
		calls.Add(1)
		return stubView(g, p)
	}
	host1, host2, guest := fakeConn(h, "p_host"), fakeConn(h, "p_host"), fakeConn(h, "p_guest")
	other := newConn(nil, "OTHER", "p_host", h.logger)
	h.register(other)

	h.Publish(context.Background(), ports.GameUpdate{Game: &domain.Game{ID: testGame, Revision: 5}, Events: []domain.Event{domain.GameStarted{}}})

	if n := calls.Load(); n != 2 {
		t.Fatalf("viewFor called %d times, want 2", n)
	}
	for _, tc := range []struct {
		c    *conn
		hand int
	}{{host1, 3}, {host2, 3}, {guest, 7}} {
		ms := drain(tc.c)
		if len(ms) != 1 || ms[0].Type != apidto.MsgView || ms[0].View.Hand[0].Rank != tc.hand || ms[0].View.Revision != 5 {
			t.Fatalf("conn %s got %+v", tc.c.playerID, ms)
		}
		if len(ms[0].Events) != 1 || ms[0].Events[0].Type != apidto.EventGameStarted {
			t.Fatalf("events = %+v", ms[0].Events)
		}
	}
	if ms := drain(other); len(ms) != 0 {
		t.Fatalf("other room got %d messages", len(ms))
	}
}

func TestPublishSkipsUnknownPlayerAndEmptyRoom(t *testing.T) {
	h := NewHub(nil)
	var calls atomic.Int32
	h.viewFor = func(g *domain.Game, p domain.PlayerID) (domain.PlayerView, error) {
		calls.Add(1)
		return stubView(g, p)
	}
	h.Publish(context.Background(), ports.GameUpdate{Game: &domain.Game{ID: testGame}})
	h.Publish(context.Background(), ports.GameUpdate{})
	if calls.Load() != 0 {
		t.Fatal("viewFor called for empty room")
	}
	stranger, host := fakeConn(h, "p_stranger"), fakeConn(h, "p_host")
	h.Publish(context.Background(), ports.GameUpdate{Game: &domain.Game{ID: testGame}})
	if ms := drain(stranger); len(ms) != 0 {
		t.Fatalf("stranger got %+v", ms)
	}
	if ms := drain(host); len(ms) != 1 {
		t.Fatalf("host got %d messages", len(ms))
	}
}

func TestSlowConsumerDoesNotBlockPublish(t *testing.T) {
	h := NewHub(nil)
	h.viewFor = stubView
	slow := fakeConn(h, "p_guest") // no writer: its buffer never drains
	fast := fakeConn(h, "p_host")

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := range sendBuffer + 10 {
			h.Publish(context.Background(), ports.GameUpdate{Game: &domain.Game{ID: testGame, Revision: int64(i)}})
			drain(fast)
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Publish blocked on a slow consumer")
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

func TestSlowConsumerClosedOverNetwork(t *testing.T) {
	e := newEnv(t)
	c := e.connect(t, "tok-host", "p_host")
	// Find the server-side conn and fill its buffer so the next enqueue kicks it.
	conns := e.hub.snapshot(testGame)
	if len(conns) != 1 {
		t.Fatalf("conns = %d", len(conns))
	}
	sc := conns[0]
	sc.kick(websocket.StatusPolicyViolation, "slow consumer")
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

func TestPublishKicksPlayerWhoLeft(t *testing.T) {
	h := NewHub(nil)
	h.viewFor = stubView
	gone1, gone2, host := fakeConn(h, "p_stranger"), fakeConn(h, "p_stranger"), fakeConn(h, "p_host")

	h.Publish(context.Background(), ports.GameUpdate{Game: &domain.Game{ID: testGame, Revision: 2}, Events: []domain.Event{domain.PlayerLeft{PlayerID: "p_stranger", HostID: "p_host"}}})

	for _, c := range []*conn{gone1, gone2} {
		select {
		case <-c.closing:
		default:
			t.Fatalf("conn of the player who left was not kicked")
		}
		if c.closeCode != closeUnauthorized || c.closeReason != apidto.CodeUnauthorized {
			t.Fatalf("close = %d %q, want 4401 unauthorized", c.closeCode, c.closeReason)
		}
		if ms := drain(c); len(ms) != 0 {
			t.Fatalf("leaver received %+v", ms)
		}
	}
	if ms := drain(host); len(ms) != 1 || ms[0].Type != apidto.MsgView {
		t.Fatalf("host got %+v", ms)
	}
	select {
	case <-host.closing:
		t.Fatal("remaining player was kicked")
	default:
	}
}
