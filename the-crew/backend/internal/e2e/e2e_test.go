// Package e2e drives the fully wired backend (sqlite + service + REST + WS)
// over real HTTP and WebSocket connections.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-crew/backend/internal/adapters/httpapi"
	"github.com/martin-pouls/the-crew/backend/internal/adapters/sqlite"
	"github.com/martin-pouls/the-crew/backend/internal/adapters/ws"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
	"github.com/martin-pouls/the-crew/backend/internal/core/service"
)

// app is one running instance of the backend on a sqlite file.
type app struct {
	srv *httptest.Server
	db  *sqlite.DB
}

func startApp(t *testing.T, dbPath string, seed uint64) *app {
	t.Helper()
	return startAppRNG(t, dbPath, func() domain.RNG { return rand.New(rand.NewPCG(seed, seed+1)) })
}

func startAppRNG(t *testing.T, dbPath string, newRNG func() domain.RNG) *app {
	t.Helper()
	db, err := sqlite.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	hub := ws.NewHub(nil)
	svc := service.New(sqlite.NewGameRepository(db), hub, service.Config{
		NewRNG: newRNG,
	})
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
	log   []wireMsg // every view message received, in order (leak checks)
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

// try sends a command and returns nil when it was acked, else the error body.
func (cl *client) try(cmd apidto.ClientMessage) *apidto.ErrorBody {
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
	return m.Error
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
		cl.log = append(cl.log, m)
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

func TestLobbyFlowSurvivesRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "lobby.db")
	a := startApp(t, dbPath, 1)
	ps := lobby(t, a)
	code := ps[0].sess.GameID

	// Up to five players fit; the sixth is turned away.
	var late []apidto.SessionResponse
	for _, name := range []string{"Cy", "Di", "Ed"} {
		late = append(late, a.join(t, code, name))
	}
	a.call(t, http.MethodPost, "/api/games/"+code+"/players", "", apidto.JoinGameRequest{Name: "Flo"}, nil, http.StatusConflict)
	var s apidto.GameSummary
	a.call(t, http.MethodGet, "/api/games/"+code, "", nil, &s, http.StatusOK)
	if s.PlayerCount != 5 || s.MaxPlayers != 5 || s.Joinable || s.Phase != "lobby" || s.HostName != "Ann" {
		t.Fatalf("summary = %+v", s)
	}

	// Both sockets see all five seats, each with their own "isYou".
	for _, p := range ps {
		p.ws.untilView("five players", func(m wireMsg) bool { return len(m.View.Players) == 5 })
		you := 0
		for _, pl := range p.ws.view.Players {
			if pl.IsYou {
				you++
				if pl.ID != p.sess.PlayerID {
					t.Fatalf("isYou on %s for %s", pl.ID, p.sess.PlayerID)
				}
			}
		}
		if you != 1 || p.ws.view.Phase != "lobby" || p.ws.view.HostID != ps[0].sess.PlayerID {
			t.Fatalf("view = %+v", p.ws.view)
		}
	}
	before := a.view(t, code, ps[0].sess.Token)

	// Restart the server on the same database: the lobby survives, tokens still work.
	a.stop()
	a = startApp(t, dbPath, 2)
	after := a.view(t, code, ps[0].sess.Token)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("state changed across restart:\n before %+v\n after  %+v", before, after)
	}
	for _, s := range late {
		if v := a.view(t, code, s.Token); v.You != s.PlayerID || len(v.Players) != 5 {
			t.Fatalf("view of %s after restart: %+v", s.PlayerID, v)
		}
	}
	reconnected := a.dial(t, "host-again", code, ps[0].sess.Token)
	if reconnected.view.Revision != before.Revision {
		t.Fatalf("revision after reconnect = %d, want %d", reconnected.view.Revision, before.Revision)
	}
}

func TestEndGameOverWebSocket(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "end.db"), 9)
	ps := lobby(t, a)
	code := ps[0].sess.GameID

	if e := ps[1].ws.sendErr(apidto.ClientMessage{Type: apidto.CmdEndGame}); e.Code != "not_host" {
		t.Fatalf("guest ends the game: %+v", e)
	}
	if e := ps[0].ws.sendErr(apidto.ClientMessage{Type: "startGame"}); e.Code != "bad_request" { // the command is called startMission
		t.Fatalf("unknown command: %+v", e)
	}
	ps[0].ws.send(apidto.ClientMessage{Type: apidto.CmdEndGame})
	m := ps[1].ws.untilView("ended", func(m wireMsg) bool { return m.View.Phase == "ended" })
	if hasEvent(m.Events, apidto.EventGameEnded) == nil {
		t.Fatalf("events = %+v", m.Events)
	}
	var s apidto.GameSummary
	a.call(t, http.MethodGet, "/api/games/"+code, "", nil, &s, http.StatusOK)
	if s.Joinable || s.Phase != "ended" {
		t.Fatalf("summary = %+v", s)
	}
	a.call(t, http.MethodPost, "/api/games/"+code+"/players", "", apidto.JoinGameRequest{Name: "Late"}, nil, http.StatusConflict)
}

func TestLeave(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "leave.db"), 3)
	host := a.create(t, "Ann")
	code := host.GameID
	guest := a.join(t, code, "Bob")
	hostWS := a.dial(t, "host", code, host.Token)
	guestWS := a.dial(t, "guest", code, guest.Token)
	hostWS.presence(2)

	// The guest leaves.
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

func TestHostLeavesPassesHostToSeatZero(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "host.db"), 4)
	host := a.create(t, "Ann")
	code := host.GameID
	bob := a.join(t, code, "Bob")
	cy := a.join(t, code, "Cy")
	a.call(t, http.MethodPost, "/api/games/"+code+"/leave", host.Token, nil, nil, http.StatusNoContent)
	for _, s := range []apidto.SessionResponse{bob, cy} {
		v := a.view(t, code, s.Token)
		if v.HostID != bob.PlayerID || len(v.Players) != 2 || v.Players[0].ID != bob.PlayerID || !v.Players[0].IsHost || v.Players[1].Seat != 1 {
			t.Fatalf("view for %s: %+v", s.PlayerID, v)
		}
	}
	// The new host can end the lobby over the WebSocket; the guest cannot.
	bobWS := a.dial(t, "bob", code, bob.Token)
	cyWS := a.dial(t, "cy", code, cy.Token)
	if e := cyWS.sendErr(apidto.ClientMessage{Type: apidto.CmdEndGame}); e.Code != "not_host" {
		t.Fatalf("guest ends the game: %+v", e)
	}
	bobWS.send(apidto.ClientMessage{Type: apidto.CmdEndGame})
	cyWS.untilView("ended", func(m wireMsg) bool { return m.View.Phase == "ended" })
	if e := bobWS.sendErr(apidto.ClientMessage{Type: apidto.CmdEndGame}); e.Code != "wrong_phase" {
		t.Fatalf("second end: %+v", e)
	}
}
