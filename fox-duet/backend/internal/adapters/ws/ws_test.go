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

	"github.com/martin-pouls/fox-duet/backend/internal/adapters/apidto"
	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
	"github.com/martin-pouls/fox-duet/backend/internal/core/ports"
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
		Hand:     []domain.Card{domain.NewCard(domain.SuitDoves, domain.Rank(handValue(p)))},
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
	lastLevel  int
	lastCard   domain.Card
	lastChosen domain.PlayerID
	lastDelta  int
	lastEnd    domain.End
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
func (f *fakeService) StartGame(ctx context.Context, id domain.GameID, p domain.PlayerID, level int) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastLevel = level
	f.mu.Unlock()
	return f.mutate(ctx, "StartGame", id, p, domain.GameStarted{Level: level})
}
func (f *fakeService) PlayCard(ctx context.Context, id domain.GameID, p domain.PlayerID, card domain.Card) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastCard = card
	f.mu.Unlock()
	return f.mutate(ctx, "PlayCard", id, p, domain.CardPlayed{PlayerID: p, Card: card, Lead: true})
}
func (f *fakeService) ChooseFoxesPlayer(ctx context.Context, id domain.GameID, p, chosen domain.PlayerID) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastChosen = chosen
	f.mu.Unlock()
	return f.mutate(ctx, "ChooseFoxesPlayer", id, p, domain.FoxesChosen{ByPlayerID: p, PlayerID: chosen})
}
func (f *fakeService) SwapDecree(ctx context.Context, id domain.GameID, p domain.PlayerID, card domain.Card) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastCard = card
	f.mu.Unlock()
	return f.mutate(ctx, "SwapDecree", id, p, domain.DecreeSwapped{PlayerID: p, New: card})
}
func (f *fakeService) KeepDecree(ctx context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return f.mutate(ctx, "KeepDecree", id, p, domain.DecreeKept{PlayerID: p})
}
func (f *fakeService) GiveGift(ctx context.Context, id domain.GameID, p domain.PlayerID, card domain.Card) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastCard = card
	f.mu.Unlock()
	return f.mutate(ctx, "GiveGift", id, p, domain.GiftChosen{PlayerID: p})
}
func (f *fakeService) ChooseMove(ctx context.Context, id domain.GameID, p domain.PlayerID, delta int) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastDelta = delta
	f.mu.Unlock()
	return f.mutate(ctx, "ChooseMove", id, p, domain.TravellerMoved{Delta: delta})
}
func (f *fakeService) ChooseForestEnd(ctx context.Context, id domain.GameID, p domain.PlayerID, end domain.End) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastEnd = end
	f.mu.Unlock()
	return f.mutate(ctx, "ChooseForestEnd", id, p, domain.ForestTokenPlaced{End: end})
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

	send(t, host, `{"type":"playCard","requestId":"4","card":{"suit":"roses","rank":8}}`)

	v := read(t, host)
	wantType(t, v, apidto.MsgView)
	if v.View.You != "p_host" || v.View.Hand[0].Rank != 3 || v.View.Revision != 1 {
		t.Fatalf("host view = %s", v.raw)
	}
	if len(v.Events) != 1 || v.Events[0].Type != apidto.EventCardPlayed || v.Events[0].Card == nil || v.Events[0].Card.Rank != 8 {
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
	card := e.svc.lastCard
	e.svc.mu.Unlock()
	if card != domain.NewCard(domain.SuitRoses, 8) {
		t.Fatalf("service got card %v", card)
	}
}

func TestAllCommands(t *testing.T) {
	tests := []struct {
		msg, call string
	}{
		{`{"type":"startGame","requestId":"1"}`, "StartGame"},
		{`{"type":"startGame","requestId":"1","level":3}`, "StartGame"},
		{`{"type":"playCard","requestId":"1","card":{"suit":"doves","rank":5}}`, "PlayCard"},
		{`{"type":"chooseFoxesPlayer","requestId":"1","playerId":"p_guest"}`, "ChooseFoxesPlayer"},
		{`{"type":"swapDecree","requestId":"1","card":{"suit":"stars","rank":10}}`, "SwapDecree"},
		{`{"type":"keepDecree","requestId":"1"}`, "KeepDecree"},
		{`{"type":"giveGift","requestId":"1","card":{"suit":"stars","rank":10}}`, "GiveGift"},
		{`{"type":"chooseMove","requestId":"1","delta":-2}`, "ChooseMove"},
		{`{"type":"chooseMove","requestId":"1","delta":0}`, "ChooseMove"},
		{`{"type":"chooseForestEnd","requestId":"1","end":"left"}`, "ChooseForestEnd"},
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

func TestCommandArguments(t *testing.T) {
	e := newEnv(t)
	c := e.connect(t, "tok-host", "p_host")
	run := func(msg string) {
		t.Helper()
		send(t, c, msg)
		wantType(t, read(t, c), apidto.MsgView)
		wantType(t, read(t, c), apidto.MsgAck)
	}
	get := func() *fakeService {
		e.svc.mu.Lock()
		defer e.svc.mu.Unlock()
		return &fakeService{lastLevel: e.svc.lastLevel, lastCard: e.svc.lastCard, lastChosen: e.svc.lastChosen, lastDelta: e.svc.lastDelta, lastEnd: e.svc.lastEnd}
	}

	run(`{"type":"startGame","requestId":"1"}`)
	if l := get().lastLevel; l != 1 {
		t.Errorf("missing level = %d, want 1", l)
	}
	run(`{"type":"startGame","requestId":"1","level":0}`)
	if l := get().lastLevel; l != 1 {
		t.Errorf("level 0 = %d, want 1", l)
	}
	run(`{"type":"startGame","requestId":"1","level":3}`)
	if l := get().lastLevel; l != 3 {
		t.Errorf("level = %d, want 3", l)
	}
	run(`{"type":"chooseFoxesPlayer","requestId":"1","playerId":"p_guest"}`)
	if p := get().lastChosen; p != "p_guest" {
		t.Errorf("chosen = %q", p)
	}
	run(`{"type":"chooseMove","requestId":"1","delta":3}`)
	if d := get().lastDelta; d != 3 {
		t.Errorf("delta = %d", d)
	}
	run(`{"type":"chooseMove","requestId":"1","delta":0}`)
	if d := get().lastDelta; d != 0 {
		t.Errorf("delta = %d", d)
	}
	run(`{"type":"chooseForestEnd","requestId":"1","end":"right"}`)
	if end := get().lastEnd; end != domain.EndRight {
		t.Errorf("end = %q", end)
	}
	// The end is passed through; the domain rejects unknown values.
	run(`{"type":"chooseForestEnd","requestId":"1","end":"up"}`)
	if end := get().lastEnd; end != "up" {
		t.Errorf("end = %q", end)
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
		{"domain error", `{"type":"playCard","requestId":"7","card":{"suit":"doves","rank":5}}`, "7", apidto.CodeMustFollowSuit, fmt.Errorf("%w: roses", domain.ErrMustFollowSuit)},
		{"not host", `{"type":"startGame","requestId":"8"}`, "8", apidto.CodeNotHost, domain.ErrNotHost},
		{"rematch exists", `{"type":"rematch","requestId":"8"}`, "8", apidto.CodeRematchExists, domain.ErrRematchExists},
		{"illegal move", `{"type":"chooseMove","requestId":"9","delta":9}`, "9", apidto.CodeIllegalMove, domain.ErrIllegalMove},
		{"illegal end", `{"type":"chooseForestEnd","requestId":"9","end":"left"}`, "9", apidto.CodeIllegalEnd, domain.ErrIllegalEnd},
		{"invalid level", `{"type":"startGame","requestId":"9","level":4}`, "9", apidto.CodeInvalidLevel, domain.ErrInvalidLevel},
		{"not your turn", `{"type":"keepDecree","requestId":"9"}`, "9", apidto.CodeNotYourTurn, domain.ErrNotYourTurn},
		{"internal", `{"type":"keepDecree","requestId":"8"}`, "8", apidto.CodeInternal, errors.New("boom")},
		{"unknown type", `{"type":"dance","requestId":"3"}`, "3", apidto.CodeBadRequest, nil},
		{"removed command", `{"type":"play","requestId":"3","cards":[]}`, "3", apidto.CodeBadRequest, nil},
		{"malformed", `{"type":`, "", apidto.CodeBadRequest, nil},
		{"playCard missing card", `{"type":"playCard","requestId":"5"}`, "5", apidto.CodeBadRequest, nil},
		{"playCard null card", `{"type":"playCard","requestId":"5","card":null}`, "5", apidto.CodeBadRequest, nil},
		{"swapDecree missing card", `{"type":"swapDecree","requestId":"5"}`, "5", apidto.CodeBadRequest, nil},
		{"giveGift missing card", `{"type":"giveGift","requestId":"5"}`, "5", apidto.CodeBadRequest, nil},
		{"invalid card", `{"type":"playCard","requestId":"5","card":{"suit":"doves","rank":99}}`, "5", apidto.CodeBadRequest, nil},
		{"missing delta", `{"type":"chooseMove","requestId":"5"}`, "5", apidto.CodeBadRequest, nil},
		{"missing end", `{"type":"chooseForestEnd","requestId":"5"}`, "5", apidto.CodeBadRequest, nil},
		{"missing playerId", `{"type":"chooseFoxesPlayer","requestId":"5"}`, "5", apidto.CodeBadRequest, nil},
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
			send(t, c, `{"type":"keepDecree","requestId":"x"}`)
			wantType(t, read(t, c), apidto.MsgView)
			wantType(t, read(t, c), apidto.MsgAck)
		})
	}
}

func TestReadLimitClosesConnection(t *testing.T) {
	e := newEnv(t)
	c := e.connect(t, "tok-host", "p_host")
	send(t, c, `{"type":"keepDecree","requestId":"`+strings.Repeat("x", 5000)+`"}`)
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
	e.hub.Publish(context.Background(), ports.GameUpdate{Game: &domain.Game{ID: testGame, Revision: 5}, Events: []domain.Event{domain.GameStarted{}}})

	for _, tc := range []struct {
		c    *websocket.Conn
		hand int
	}{{host1, 3}, {host2, 3}, {guest, 7}} {
		m := read(t, tc.c)
		wantType(t, m, apidto.MsgView)
		if m.View.Hand[0].Rank != tc.hand || m.View.Revision != 5 {
			t.Fatalf("view = %s", m.raw)
		}
		if len(m.Events) != 1 || m.Events[0].Type != apidto.EventGameStarted {
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
