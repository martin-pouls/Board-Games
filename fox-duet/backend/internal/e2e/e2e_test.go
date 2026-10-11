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
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/martin-pouls/fox-duet/backend/internal/adapters/apidto"
	"github.com/martin-pouls/fox-duet/backend/internal/adapters/httpapi"
	"github.com/martin-pouls/fox-duet/backend/internal/adapters/sqlite"
	"github.com/martin-pouls/fox-duet/backend/internal/adapters/ws"
	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
	"github.com/martin-pouls/fox-duet/backend/internal/core/service"
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

func hasRevealedHands(v apidto.PlayerView) bool {
	for _, p := range v.Players {
		if p.RevealedHand != nil {
			return true
		}
	}
	return false
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

// decide returns the command the player can send now, from their own view only.
// trick drives the choice between swapping the decree and keeping it.
func decide(v apidto.PlayerView) (apidto.ClientMessage, bool) {
	p := v.Pending
	if v.Phase != "playing" || p == nil || !slices.Contains(p.PlayerIDs, v.You) {
		return apidto.ClientMessage{}, false
	}
	other := ""
	for _, ps := range v.Players {
		if ps.ID != v.You {
			other = ps.ID
		}
	}
	switch p.Step {
	case "lead", "follow":
		c := v.Playable[0]
		return apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: &c}, true
	case "foxesChoose":
		// Alternate between choosing the teammate and oneself.
		id := other
		if v.Round%2 == 0 {
			id = v.You
		}
		return apidto.ClientMessage{Type: apidto.CmdChooseFoxesPlayer, PlayerID: id}, true
	case "foxesSwap":
		if v.Round%2 == 0 && len(v.Hand) > 0 {
			c := v.Hand[0]
			return apidto.ClientMessage{Type: apidto.CmdSwapDecree, Card: &c}, true
		}
		return apidto.ClientMessage{Type: apidto.CmdKeepDecree}, true
	case "gift":
		c := v.Hand[len(v.Hand)-1]
		return apidto.ClientMessage{Type: apidto.CmdGiveGift, Card: &c}, true
	case "move":
		d := p.MoveOptions[0].Delta
		return apidto.ClientMessage{Type: apidto.CmdChooseMove, Delta: &d}, true
	case "forestEnd":
		return apidto.ClientMessage{Type: apidto.CmdChooseForestEnd, End: p.Ends[0]}, true
	}
	return apidto.ClientMessage{}, false
}

// checkViews asserts the privacy invariants for two views at the same revision.
func checkViews(t *testing.T, a, b apidto.PlayerView, rawA []byte) {
	t.Helper()
	if a.Revision != b.Revision {
		return
	}
	for _, c := range a.Hand {
		if slices.Contains(b.Hand, c) {
			t.Fatalf("card %+v is in both hands", c)
		}
	}
	for _, ps := range a.Players {
		if ps.ID == b.You && ps.HandCount != len(b.Hand) {
			t.Fatalf("hand count of the teammate = %d, real hand %d", ps.HandCount, len(b.Hand))
		}
	}
	if a.Phase == "playing" && len(rawA) > 0 {
		if strings.Contains(string(rawA), "revealedHand") {
			t.Fatalf("hands revealed before the end: %s", rawA)
		}
		// None of the teammate's hand cards appears anywhere in the view except where the
		// cards are public (decree, trick, last trick, last gift) - check the hand-bearing fields.
		for _, c := range b.Hand {
			if slices.Contains(a.Hand, c) || slices.Contains(a.Playable, c) {
				t.Fatalf("teammate card %+v in own hand/playable", c)
			}
		}
	}
}

// playToEnd drives both players from their views until the game is over and returns the final
// view of player 0 and the steps seen. onTrick is called once after the first completed trick.
func playToEnd(t *testing.T, ps [2]*player, onTrick func()) (apidto.PlayerView, map[string]int) {
	t.Helper()
	steps := map[string]int{}
	trickDone := false
	for i := 0; i < 2000; i++ {
		v0, v1 := ps[0].ws.view, ps[1].ws.view
		if v0.Phase != "playing" {
			return v0, steps
		}
		var actor *player
		var cmd apidto.ClientMessage
		for _, p := range ps {
			if c, ok := decide(p.ws.view); ok {
				actor, cmd = p, c
				break
			}
		}
		if actor == nil {
			t.Fatalf("nobody can act: pending %+v / %+v", v0.Pending, v1.Pending)
		}
		steps[actor.ws.view.Pending.Step]++
		actor.ws.send(cmd)
		other := ps[0]
		if actor == ps[0] {
			other = ps[1]
		}
		other.ws.syncTo(actor.ws.view.Revision)
		checkViews(t, ps[0].ws.view, ps[1].ws.view, ps[0].ws.last.raw)
		checkViews(t, ps[1].ws.view, ps[0].ws.view, ps[1].ws.last.raw)
		if !trickDone && ps[0].ws.view.LastTrick != nil {
			trickDone = true
			if onTrick != nil {
				onTrick()
			}
		}
	}
	t.Fatal("game did not finish")
	return apidto.PlayerView{}, nil
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

func TestTwoPlayerGameToCompletionSurvivesRestart(t *testing.T) {
	var won, lost int
	seen := map[string]int{}
	for _, tc := range []struct {
		level int
		seed  uint64
	}{{1, 1}, {1, 2}, {1, 3}, {2, 4}, {3, 5}} {
		t.Run(fmt.Sprintf("level%d_seed%d", tc.level, tc.seed), func(t *testing.T) {
			dbPath := filepath.Join(t.TempDir(), "two.db")
			newRNG := func() domain.RNG { return rand.New(rand.NewPCG(tc.seed, tc.seed+1)) }
			a := startAppRNG(t, dbPath, newRNG)
			ps := lobby(t, a)
			code := ps[0].sess.GameID

			// Start over REST; both sockets learn about it.
			var started apidto.PlayerView
			a.call(t, http.MethodPost, "/api/games/"+code+"/start", ps[0].sess.Token, apidto.StartGameRequest{Level: tc.level}, &started, http.StatusOK)
			for _, p := range ps {
				p.ws.untilView("started", func(m wireMsg) bool { return m.View.Phase == "playing" })
			}
			v0, v1 := ps[0].ws.view, ps[1].ws.view
			if v0.Level != tc.level || len(v0.Hand) != domain.HandSize || len(v1.Hand) != domain.HandSize || v0.Board == nil || v0.Decree == nil || v0.Round != 1 {
				t.Fatalf("start views: %+v / %+v", v0, v1)
			}
			if v0.Pending == nil || v0.Pending.Step != "lead" || len(v0.Pending.PlayerIDs) != 1 {
				t.Fatalf("pending = %+v", v0.Pending)
			}
			checkViews(t, v0, v1, ps[0].ws.last.raw)

			// Errors: acting out of turn, and playing a card the player does not hold.
			var leader, follower *player
			if v0.Pending.PlayerIDs[0] == ps[0].sess.PlayerID {
				leader, follower = ps[0], ps[1]
			} else {
				leader, follower = ps[1], ps[0]
			}
			fc := follower.ws.view.Hand[0]
			if e := follower.ws.sendErr(apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: &fc}); e.Code != "not_your_turn" {
				t.Fatalf("follower plays first: %+v", e)
			}
			if e := leader.ws.sendErr(apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: &fc}); e.Code != "card_not_in_hand" {
				t.Fatalf("leader plays the teammate's card: %+v", e)
			}
			if e := leader.ws.sendErr(apidto.ClientMessage{Type: apidto.CmdKeepDecree}); e.Code != "wrong_step" {
				t.Fatalf("keepDecree while leading: %+v", e)
			}
			if e := leader.ws.sendErr(apidto.ClientMessage{Type: apidto.CmdPlayCard}); e.Code != "bad_request" {
				t.Fatalf("playCard without a card: %+v", e)
			}
			if e := leader.ws.sendErr(apidto.ClientMessage{Type: apidto.CmdStartGame}); e.Code != "wrong_phase" && e.Code != "not_host" {
				t.Fatalf("startGame while playing: %+v", e)
			}

			restarted := false
			final, steps := playToEnd(t, ps, func() {
				// The first trick is complete: restart the server on the same database.
				rev := ps[0].ws.view.Revision
				before := a.view(t, code, ps[0].sess.Token)
				a.stop()
				a = startAppRNG(t, dbPath, newRNG)
				after := a.view(t, code, ps[0].sess.Token)
				if after.Revision != rev || !reflect.DeepEqual(after.Hand, before.Hand) || !reflect.DeepEqual(after.Board, before.Board) ||
					!reflect.DeepEqual(after.Pending, before.Pending) || !reflect.DeepEqual(after.Decree, before.Decree) {
					t.Fatalf("state changed across restart:\n before %+v\n after  %+v", before, after)
				}
				for _, p := range ps {
					p.ws = a.dial(t, p.ws.name, code, p.sess.Token)
				}
				if ps[0].ws.view.Revision != rev || ps[1].ws.view.Revision != rev {
					t.Fatalf("revisions after reconnect: %d / %d, want %d", ps[0].ws.view.Revision, ps[1].ws.view.Revision, rev)
				}
				restarted = true
			})
			if !restarted {
				t.Fatal("no trick completed before the game ended")
			}
			for k, n := range steps {
				seen[k] += n
			}

			// The game is over: hands are revealed, the result is set, nothing else is accepted.
			ps[1].ws.syncTo(final.Revision)
			fv := ps[1].ws.view
			if final.Phase != "won" && final.Phase != "lost" || final.Result == nil || final.Result.Outcome != final.Phase || final.Pending != nil {
				t.Fatalf("final view: %+v", final)
			}
			if final.Phase == "won" {
				won++
				s := final.Result.Score
				if s == nil || s.Total != s.Level+s.Cards+s.EarlyWin+s.ForestTokens || s.Level != map[int]int{1: 10, 2: 20, 3: 30}[tc.level] {
					t.Fatalf("score = %+v", s)
				}
			} else {
				lost++
				if final.Result.Score != nil || final.Result.Reason == "" {
					t.Fatalf("lost result = %+v", final.Result)
				}
			}
			if len(final.Players) != 2 || len(fv.Players) != 2 {
				t.Fatalf("players: %+v", final.Players)
			}
			for _, p := range final.Players {
				if p.RevealedHand == nil && p.HandCount > 0 {
					t.Fatalf("hand of %s not revealed after the end: %+v", p.ID, p)
				}
			}
			var e apidto.ErrorResponse
			a.call(t, http.MethodPost, "/api/games/"+code+"/give-up", ps[0].sess.Token, nil, &e, http.StatusConflict)
			if e.Error.Code != "wrong_phase" {
				t.Fatalf("give-up after the end: %+v", e)
			}
		})
	}
	t.Logf("results: %d won, %d lost; steps seen: %v", won, lost, seen)
	if seen["lead"] == 0 || seen["follow"] == 0 {
		t.Fatalf("no tricks were played: %v", seen)
	}
}

func TestGiveUpAndRematchKeepGroupAndTokens(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "rematch.db"), 5)
	ps := lobby(t, a)
	ann, bob := ps[0].sess, ps[1].sess
	code := ann.GameID
	a.call(t, http.MethodPost, "/api/games/"+code+"/rematch", ann.Token, nil, nil, http.StatusConflict) // not finished yet
	// A missing body starts level 1.
	var started apidto.PlayerView
	a.call(t, http.MethodPost, "/api/games/"+code+"/start", ann.Token, nil, &started, http.StatusOK)
	if started.Phase != "playing" || started.Level != 1 || started.Round != 1 {
		t.Fatalf("start: %+v", started)
	}
	for _, p := range ps {
		p.ws.untilView("started", func(m wireMsg) bool { return m.View.Phase == "playing" })
	}

	// Bob gives up over WebSocket; both players see the revealed hands.
	ps[1].ws.send(apidto.ClientMessage{Type: apidto.CmdGiveUp})
	m := ps[0].ws.untilView("gave up", func(m wireMsg) bool { return m.View.Phase == "lost" })
	fin := *m.View
	if fin.Result == nil || fin.Result.Reason != "gaveUp" || fin.Result.PlayerID != bob.PlayerID || !hasRevealedHands(fin) || hasEvent(m.Events, apidto.EventGameLost) == nil {
		t.Fatalf("give up: %+v events %+v", fin.Result, m.Events)
	}

	var old apidto.PlayerView
	a.call(t, http.MethodPost, "/api/games/"+code+"/rematch", bob.Token, nil, &old, http.StatusOK)
	next := old.NextGameID
	if next == "" || next == code || old.Phase != "lost" {
		t.Fatalf("rematch: next=%q phase=%s", next, old.Phase)
	}
	// Ann's open socket learns about the next game.
	m = ps[0].ws.until("rematchCreated", func(m wireMsg) bool { return m.Type == apidto.MsgView && m.View.NextGameID != "" })
	if m.View.NextGameID != next || !slices.ContainsFunc(m.Events, func(e apidto.Event) bool { return e.Type == apidto.EventRematchCreated && e.GameID == next }) {
		t.Fatalf("ann's rematch update: next=%q events=%+v", m.View.NextGameID, m.Events)
	}
	// Idempotent: a second request returns the same next game.
	var again apidto.PlayerView
	a.call(t, http.MethodPost, "/api/games/"+code+"/rematch", ann.Token, nil, &again, http.StatusOK)
	if again.NextGameID != next {
		t.Fatalf("second rematch: %q, want %q", again.NextGameID, next)
	}

	// Same tokens work on the new game; it is a lobby with both players, Bob hosting.
	for _, s := range []apidto.SessionResponse{ann, bob} {
		v := a.view(t, next, s.Token)
		if v.You != s.PlayerID || v.Phase != "lobby" || v.HostID != bob.PlayerID || len(v.Players) != 2 || v.Revision != 1 {
			t.Fatalf("next game view for %s: %+v", s.PlayerID, v)
		}
		if len(v.Hand) != 0 || hasRevealedHands(v) || v.Result != nil || v.Board != nil || v.Pending != nil {
			t.Fatalf("next game not fresh: %+v", v)
		}
	}
	a.call(t, http.MethodPost, "/api/games/"+next+"/start", ann.Token, nil, nil, http.StatusForbidden) // Ann is no longer host
	nextWS := a.dial(t, "ann-next", next, ann.Token)
	if nextWS.view.GameID != next {
		t.Fatalf("ws on next game: %+v", nextWS.view)
	}
	// The rematch is started over WebSocket by its host, at level 2.
	bobNext := a.dial(t, "bob-next", next, bob.Token)
	bobNext.send(apidto.ClientMessage{Type: apidto.CmdStartGame, Level: 2})
	if v := bobNext.view; v.Phase != "playing" || v.Level != 2 || v.Board == nil || v.Board.Side != "B" {
		t.Fatalf("next game start: %+v", v)
	}
}

func TestLeave(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "leave.db"), 3)
	host := a.create(t, "Ann")
	code := host.GameID
	guest := a.join(t, code, "Bob")
	hostWS := a.dial(t, "host", code, host.Token)
	guestWS := a.dial(t, "guest", code, guest.Token)
	hostWS.presence(2)

	// A third player is turned away.
	a.call(t, http.MethodPost, "/api/games/"+code+"/players", "", apidto.JoinGameRequest{Name: "Cy"}, nil, http.StatusConflict)
	// The host cannot start alone, but the guest can leave.
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
	var e apidto.ErrorResponse
	a.call(t, http.MethodPost, "/api/games/"+code+"/start", host.Token, nil, &e, http.StatusConflict)
	if e.Error.Code != "not_enough_players" {
		t.Fatalf("start alone: %+v", e)
	}
	// The sole remaining player cannot leave.
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
	a := startApp(t, filepath.Join(t.TempDir(), "end.db"), 9)
	host := a.create(t, "Ann")
	guest := a.join(t, host.GameID, "Bob")
	a.call(t, http.MethodPost, "/api/games/"+host.GameID+"/end-game", guest.Token, nil, nil, http.StatusForbidden)
	var v apidto.PlayerView
	a.call(t, http.MethodPost, "/api/games/"+host.GameID+"/end-game", host.Token, nil, &v, http.StatusOK)
	if v.Phase != "ended" {
		t.Fatalf("phase = %s", v.Phase)
	}
	var s apidto.GameSummary
	a.call(t, http.MethodGet, "/api/games/"+host.GameID, "", nil, &s, http.StatusOK)
	if s.Joinable || s.Phase != "ended" {
		t.Fatalf("summary = %+v", s)
	}
}
