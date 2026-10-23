// Package e2e drives the fully wired backend (sqlite + service + REST + WS)
// over real HTTP and WebSocket connections.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/martin-pouls/sky-team/backend/internal/adapters/apidto"
	"github.com/martin-pouls/sky-team/backend/internal/adapters/httpapi"
	"github.com/martin-pouls/sky-team/backend/internal/adapters/sqlite"
	"github.com/martin-pouls/sky-team/backend/internal/adapters/ws"
	"github.com/martin-pouls/sky-team/backend/internal/core/service"
)

// app is one running instance of the backend on a sqlite file.
type app struct {
	srv *httptest.Server
	db  *sqlite.DB
}

func startApp(t *testing.T, dbPath string) *app {
	t.Helper()
	db, err := sqlite.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	hub := ws.NewHub(nil)
	svc := service.New(sqlite.NewGameRepository(db), hub, service.Config{})
	srv := httptest.NewServer(httpapi.NewRouter(httpapi.Deps{
		Health: db.Ping,
		Games:  svc,
		WS:     ws.NewHandler(hub, svc, ws.Options{}),
	}))
	a := &app{srv: srv, db: db}
	t.Cleanup(a.stop)
	return a
}

func (a *app) stop() {
	if a.srv != nil {
		a.srv.CloseClientConnections()
		a.srv.Close()
		a.srv = nil
	}
	if a.db != nil {
		_ = a.db.Close()
		a.db = nil
	}
}

// call performs a REST request and decodes the 2xx body into out.
func (a *app) call(t *testing.T, method, path, token string, body, out any, wantStatus int) {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, a.srv.URL+path, r)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := a.srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != wantStatus {
		t.Fatalf("%s %s: status %d, want %d: %s", method, path, resp.StatusCode, wantStatus, data)
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			t.Fatalf("%s %s: decode: %v: %s", method, path, err, data)
		}
	}
}

func (a *app) create(t *testing.T, name string) apidto.SessionResponse {
	t.Helper()
	var s apidto.SessionResponse
	a.call(t, http.MethodPost, "/api/games", "", apidto.CreateGameRequest{Name: name}, &s, http.StatusCreated)
	return s
}

func (a *app) join(t *testing.T, code, name string) apidto.SessionResponse {
	t.Helper()
	var s apidto.SessionResponse
	a.call(t, http.MethodPost, "/api/games/"+code+"/players", "", apidto.JoinGameRequest{Name: name}, &s, http.StatusCreated)
	return s
}

func (a *app) view(t *testing.T, code, token string) apidto.PlayerView {
	t.Helper()
	var v apidto.PlayerView
	a.call(t, http.MethodGet, "/api/games/"+code+"/view", token, nil, &v, http.StatusOK)
	return v
}

// client is a WebSocket connection for one player.
type client struct {
	t     *testing.T
	name  string
	c     *websocket.Conn
	view  apidto.PlayerView
	last  wireMsg // most recent view message
	reqID int
}

// wireMsg decodes any server message; View is kept raw too for leak checks.
type wireMsg struct {
	Type      string             `json:"type"`
	RequestID string             `json:"requestId"`
	View      *apidto.PlayerView `json:"view"`
	Events    []apidto.Event     `json:"events"`
	Online    []string           `json:"online"`
	Error     *apidto.ErrorBody  `json:"error"`
	raw       []byte
}

func (a *app) dial(t *testing.T, name, code, token string) *client {
	t.Helper()
	u := "ws" + strings.TrimPrefix(a.srv.URL, "http") + "/ws/games/" + code + "?token=" + url.QueryEscape(token)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, u, nil)
	if err != nil {
		t.Fatalf("%s: dial: %v", name, err)
	}
	t.Cleanup(func() { _ = c.CloseNow() })
	cl := &client{t: t, name: name, c: c}
	m := cl.read()
	if m.Type != apidto.MsgView || m.Events == nil {
		t.Fatalf("%s: first message %q (events nil=%v), want view with events", name, m.Type, m.Events == nil)
	}
	return cl
}

func (cl *client) read() wireMsg {
	cl.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, data, err := cl.c.Read(ctx)
	if err != nil {
		cl.t.Fatalf("%s: read: %v", cl.name, err)
	}
	var m wireMsg
	if err := json.Unmarshal(data, &m); err != nil {
		cl.t.Fatalf("%s: decode %s: %v", cl.name, data, err)
	}
	m.raw = data
	if m.Type == apidto.MsgView {
		if m.View.Revision < cl.view.Revision {
			cl.t.Fatalf("%s: view revision went backwards: %d < %d", cl.name, m.View.Revision, cl.view.Revision)
		}
		cl.view = *m.View
		cl.last = m
	}
	return m
}

// until reads messages until pred matches one and returns it.
func (cl *client) until(what string, pred func(wireMsg) bool) wireMsg {
	cl.t.Helper()
	for range 200 {
		if m := cl.read(); pred(m) {
			return m
		}
	}
	cl.t.Fatalf("%s: never got %s", cl.name, what)
	return wireMsg{}
}

func (cl *client) presence(n int) {
	cl.t.Helper()
	cl.until(fmt.Sprintf("presence with %d online", n), func(m wireMsg) bool {
		return m.Type == apidto.MsgPresence && len(m.Online) == n
	})
}

// send sends a command and reads until its ack, failing on an error reply.
func (cl *client) send(cmd apidto.ClientMessage) {
	cl.t.Helper()
	cl.reqID++
	cmd.RequestID = strconv.Itoa(cl.reqID)
	b, _ := json.Marshal(cmd)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cl.c.Write(ctx, websocket.MessageText, b); err != nil {
		cl.t.Fatalf("%s: write: %v", cl.name, err)
	}
	cl.until("ack "+cmd.RequestID, func(m wireMsg) bool {
		if m.Type == apidto.MsgError && m.RequestID == cmd.RequestID {
			cl.t.Fatalf("%s: %s rejected: %+v", cl.name, cmd.Type, m.Error)
		}
		return m.Type == apidto.MsgAck && m.RequestID == cmd.RequestID
	})
}

// sendErr sends a command that must be rejected and returns the error body.
func (cl *client) sendErr(cmd apidto.ClientMessage) apidto.ErrorBody {
	cl.t.Helper()
	cl.reqID++
	cmd.RequestID = strconv.Itoa(cl.reqID)
	b, _ := json.Marshal(cmd)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cl.c.Write(ctx, websocket.MessageText, b); err != nil {
		cl.t.Fatalf("%s: write: %v", cl.name, err)
	}
	m := cl.until("reply to "+cmd.RequestID, func(m wireMsg) bool {
		return (m.Type == apidto.MsgError || m.Type == apidto.MsgAck) && m.RequestID == cmd.RequestID
	})
	if m.Type != apidto.MsgError {
		cl.t.Fatalf("%s: %s was accepted, want an error", cl.name, cmd.Type)
	}
	return *m.Error
}

// untilView reads until a view message arrives (after any pending ones) that satisfies pred.
func (cl *client) untilView(what string, pred func(wireMsg) bool) wireMsg {
	cl.t.Helper()
	return cl.until(what, func(m wireMsg) bool { return m.Type == apidto.MsgView && pred(m) })
}

func hasEvent(evs []apidto.Event, typ string) *apidto.Event {
	for i := range evs {
		if evs[i].Type == typ {
			return &evs[i]
		}
	}
	return nil
}

// player is one seated player with a live WebSocket.
type player struct {
	sess apidto.SessionResponse
	ws   *client
}

// syncTo reads cl until its view reaches revision rev.
func (cl *client) syncTo(rev int64) {
	cl.t.Helper()
	if cl.view.Revision >= rev {
		return
	}
	cl.untilView(fmt.Sprintf("revision %d", rev), func(m wireMsg) bool { return m.View.Revision >= rev })
}

// lobby creates a game, seats two players and connects both over WebSocket.
func lobby(t *testing.T, a *app) [2]*player {
	t.Helper()
	host := a.create(t, "Ann")
	guest := a.join(t, host.GameID, "Bob")
	ps := [2]*player{{sess: host}, {sess: guest}}
	ps[0].ws = a.dial(t, "host", host.GameID, host.Token)
	ps[1].ws = a.dial(t, "guest", host.GameID, guest.Token)
	ps[0].ws.presence(2)
	ps[1].ws.presence(2)
	return ps
}

func TestLobbyJoinAndRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "restart.db")
	a := startApp(t, dbPath)
	ps := lobby(t, a)
	code := ps[0].sess.GameID

	v0, v1 := ps[0].ws.view, ps[1].ws.view
	if v0.Phase != "lobby" || v1.Phase != "lobby" || len(v0.Players) != 2 || v0.HostID != ps[0].sess.PlayerID {
		t.Fatalf("lobby views: %+v / %+v", v0, v1)
	}
	if v0.You != ps[0].sess.PlayerID || v1.You != ps[1].sess.PlayerID || !v0.Players[0].IsHost || !v1.Players[1].IsYou {
		t.Fatalf("views not per player: %+v / %+v", v0, v1)
	}
	var s apidto.GameSummary
	a.call(t, http.MethodGet, "/api/games/"+code, "", nil, &s, http.StatusOK)
	if s.Joinable || s.PlayerCount != 2 || s.HostName != "Ann" {
		t.Fatalf("summary = %+v", s)
	}

	// The game survives a restart: same views, same tokens.
	a.stop()
	a2 := startApp(t, dbPath)
	got := a2.view(t, code, ps[1].sess.Token)
	if got.Phase != "lobby" || got.Revision != v1.Revision || got.You != ps[1].sess.PlayerID || len(got.Players) != 2 {
		t.Fatalf("after restart: %+v, want revision %d", got, v1.Revision)
	}
}

func TestLeave(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "leave.db"))
	host := a.create(t, "Ann")
	code := host.GameID
	guest := a.join(t, code, "Bob")
	hostWS := a.dial(t, "host", code, host.Token)
	guestWS := a.dial(t, "guest", code, guest.Token)
	hostWS.presence(2)

	// A third player is turned away.
	a.call(t, http.MethodPost, "/api/games/"+code+"/players", "", apidto.JoinGameRequest{Name: "Cy"}, nil, http.StatusConflict)
	// The guest can leave.
	a.call(t, http.MethodPost, "/api/games/"+code+"/leave", guest.Token, nil, nil, http.StatusNoContent)

	// The guest's socket is closed with 4401.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		if _, _, err := guestWS.c.Read(ctx); err != nil {
			if got := websocket.CloseStatus(err); got != 4401 {
				t.Fatalf("guest close status = %d (err %v), want 4401", got, err)
			}
			break
		}
	}
	// The guest's token no longer works.
	a.call(t, http.MethodGet, "/api/games/"+code+"/view", guest.Token, nil, nil, http.StatusUnauthorized)
	a.call(t, http.MethodPost, "/api/games/"+code+"/leave", guest.Token, nil, nil, http.StatusUnauthorized)

	// The host sees one player and the playerLeft event.
	m := hostWS.untilView("playerLeft", func(m wireMsg) bool { return hasEvent(m.Events, apidto.EventPlayerLeft) != nil })
	if ev := hasEvent(m.Events, apidto.EventPlayerLeft); ev.PlayerID != guest.PlayerID || ev.HostID != host.PlayerID {
		t.Fatalf("playerLeft = %+v", ev)
	}
	if len(m.View.Players) != 1 || m.View.Players[0].ID != host.PlayerID {
		t.Fatalf("host players = %+v", m.View.Players)
	}
	// The sole remaining player cannot leave.
	var e apidto.ErrorResponse
	a.call(t, http.MethodPost, "/api/games/"+code+"/leave", host.Token, nil, &e, http.StatusConflict)
	if e.Error.Code != "cannot_leave" {
		t.Fatalf("sole leave: %+v", e)
	}
	// A new guest can take the free seat.
	if s := a.join(t, code, "Bob"); s.PlayerID == guest.PlayerID {
		t.Fatalf("rejoin reused the old player id")
	}
}

func TestEndGameClosesLobby(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "end.db"))
	host := a.create(t, "Ann")
	guest := a.join(t, host.GameID, "Bob")
	hostWS := a.dial(t, "host", host.GameID, host.Token)
	guestWS := a.dial(t, "guest", host.GameID, guest.Token)
	if e := guestWS.sendErr(apidto.ClientMessage{Type: apidto.CmdEndGame}); e.Code != "not_host" {
		t.Fatalf("end by the guest: %+v", e)
	}
	hostWS.send(apidto.ClientMessage{Type: apidto.CmdEndGame})
	if v := hostWS.view; v.Phase != "ended" {
		t.Fatalf("phase = %s", v.Phase)
	}
	guestWS.untilView("ended", func(m wireMsg) bool { return m.View.Phase == "ended" })
	var s apidto.GameSummary
	a.call(t, http.MethodGet, "/api/games/"+host.GameID, "", nil, &s, http.StatusOK)
	if s.Joinable || s.Phase != "ended" {
		t.Fatalf("summary = %+v", s)
	}
}
