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
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/martin-pouls/regicide/backend/internal/adapters/apidto"
	"github.com/martin-pouls/regicide/backend/internal/adapters/httpapi"
	"github.com/martin-pouls/regicide/backend/internal/adapters/sqlite"
	"github.com/martin-pouls/regicide/backend/internal/adapters/ws"
	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/service"
)

// app is one running instance of the backend on a sqlite file.
type app struct {
	srv *httptest.Server
	db  *sqlite.DB
}

// identityRNG never shuffles, so the deck stays ascending and deals are deterministic.
type identityRNG struct{}

func (identityRNG) Shuffle(int, func(i, j int)) {}

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

// wc parses a card such as "8C", "10S", "AH", "KD" or "X" (Jester).
func wc(s string) apidto.Card {
	if s == "X" {
		return apidto.Card{}
	}
	suits := map[byte]string{'H': "hearts", 'D': "diamonds", 'C': "clubs", 'S': "spades"}
	ranks := map[string]int{"A": 1, "J": 11, "Q": 12, "K": 13}
	r, ok := ranks[s[:len(s)-1]]
	if !ok {
		var err error
		if r, err = strconv.Atoi(s[:len(s)-1]); err != nil {
			panic("bad card " + s)
		}
	}
	return apidto.Card{Suit: suits[s[len(s)-1]], Rank: r}
}

func wcs(s string) []apidto.Card {
	out := []apidto.Card{}
	for _, f := range strings.Fields(s) {
		out = append(out, wc(f))
	}
	return out
}

// cardNames renders cards as sorted "8C"-style names.
func cardNames(cs []apidto.Card) []string {
	sym := map[string]string{"hearts": "H", "diamonds": "D", "clubs": "C", "spades": "S", "": ""}
	rank := map[int]string{0: "X", 1: "A", 11: "J", 12: "Q", 13: "K"}
	out := make([]string, 0, len(cs))
	for _, c := range cs {
		r, ok := rank[c.Rank]
		if !ok {
			r = strconv.Itoa(c.Rank)
		}
		if c.Rank == 0 {
			out = append(out, "X")
			continue
		}
		out = append(out, r+sym[c.Suit])
	}
	slices.Sort(out)
	return out
}

func sortedNames(s string) []string {
	out := strings.Fields(s)
	slices.Sort(out)
	return out
}

func wantHand(t *testing.T, who string, v apidto.PlayerView, want string) {
	t.Helper()
	if got := cardNames(v.Hand); !slices.Equal(got, sortedNames(want)) {
		t.Fatalf("%s hand = %v, want %v", who, got, sortedNames(want))
	}
}

func hasEvent(evs []apidto.Event, typ string) *apidto.Event {
	for i := range evs {
		if evs[i].Type == typ {
			return &evs[i]
		}
	}
	return nil
}

func TestTwoPlayerGameOverWebSocketSurvivesRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "two.db")
	a := startAppRNG(t, dbPath, func() domain.RNG { return identityRNG{} })

	host := a.create(t, "Ann")
	code := host.GameID
	guest := a.join(t, code, "Bob")

	hostWS := a.dial(t, "host", code, host.Token)
	guestWS := a.dial(t, "guest", code, guest.Token)
	hostWS.presence(2)
	guestWS.presence(2)

	hostWS.send(apidto.ClientMessage{Type: apidto.CmdStartGame, FirstPlayerID: host.PlayerID})
	guestWS.untilView("started", func(m wireMsg) bool { return m.View.Phase == "playing" })
	hv := *hostWS.last.View
	gv := guestWS.view

	// Deal: identityRNG, host first. The first enemy is the Jack of spades.
	if e := hv.Enemy; e == nil || e.Card != wc("JS") || e.Attack != 10 || e.Health != 20 || e.Damage != 0 {
		t.Fatalf("enemy = %+v", e)
	}
	wantHand(t, "host", hv, "2S 4S 6S 8S 10S 8C 10C")
	wantHand(t, "guest", gv, "AS 3S 5S 7S 9S 7C 9C")
	if hv.Turn.ActivePlayerID != host.PlayerID || hv.Turn.Step != "play" || hv.CastleCount != 11 || hv.TavernCount != 26 {
		t.Fatalf("turn/counts = %+v castle %d tavern %d", hv.Turn, hv.CastleCount, hv.TavernCount)
	}

	// Leak check: the guest never sees the host's cards, only a count.
	checkNoLeak := func(raw []byte, v apidto.PlayerView, hostCount int) {
		t.Helper()
		if strings.Contains(string(raw), "revealedHand") {
			t.Fatalf("hands revealed before the end: %s", raw)
		}
		for _, p := range v.Players {
			if p.ID == host.PlayerID && p.HandCount != hostCount {
				t.Fatalf("host hand count = %d, want %d", p.HandCount, hostCount)
			}
		}
		for _, n := range cardNames(v.Hand) {
			if slices.Contains(sortedNames("2S 4S 6S 8S 10S 8C 10C"), n) {
				t.Fatalf("guest sees host card %s", n)
			}
		}
	}
	var rawView []byte
	{
		var raw json.RawMessage
		a.call(t, http.MethodGet, "/api/games/"+code+"/view", guest.Token, nil, &raw, http.StatusOK)
		rawView = raw
	}
	checkNoLeak(rawView, gv, 7)

	// Host plays 8C: clubs double, 16 damage; the guest-facing step is discard.
	hostWS.send(apidto.ClientMessage{Type: apidto.CmdPlay, Cards: wcs("8C")})
	hv = *hostWS.last.View
	if hv.Enemy.Damage != 16 || hv.Turn.DamageToCover != 10 || len(hv.Enemy.Plays) != 1 || !hv.Enemy.Plays[0].Doubled {
		t.Fatalf("after 8C: enemy %+v turn %+v", hv.Enemy, hv.Turn)
	}
	// Host cannot yield in the discard step, nor play.
	if e := hostWS.sendErr(apidto.ClientMessage{Type: apidto.CmdPlay, Cards: wcs("2S")}); e.Code != "wrong_step" {
		t.Fatalf("play in discard step: %+v", e)
	}
	// An insufficient discard is rejected.
	if e := hostWS.sendErr(apidto.ClientMessage{Type: apidto.CmdDiscard, Cards: wcs("2S")}); e.Code != "illegal_discard" {
		t.Fatalf("short discard: %+v", e)
	}
	hostWS.send(apidto.ClientMessage{Type: apidto.CmdDiscard, Cards: wcs("10S")})
	gm := guestWS.untilView("guest turn", func(m wireMsg) bool { return m.View.Turn.ActivePlayerID == guest.PlayerID })
	if hasEvent(gm.Events, apidto.EventCardsDiscarded) == nil || gm.View.Turn.Step != "play" || !gm.View.Turn.CanYield {
		t.Fatalf("guest turn: events %+v turn %+v", gm.Events, gm.View.Turn)
	}

	// Guest yields and discards 9C + AS (10 damage to cover).
	guestWS.send(apidto.ClientMessage{Type: apidto.CmdYield})
	gv = *guestWS.last.View
	if gv.Turn.DamageToCover != 10 {
		t.Fatalf("guest damage to cover = %d", gv.Turn.DamageToCover)
	}
	guestWS.send(apidto.ClientMessage{Type: apidto.CmdDiscard, Cards: wcs("9C AS")})
	hm := hostWS.untilView("host turn", func(m wireMsg) bool {
		return m.View.Turn.ActivePlayerID == host.PlayerID && m.View.Turn.Step == "play"
	})
	if !hm.View.Players[1].YieldedLastTurn {
		t.Fatalf("guest yield flag missing: %+v", hm.View.Players)
	}
	if hm.View.Turn.CanYield {
		t.Fatalf("host must not be able to yield after the guest yielded")
	}

	// The host may not yield now.
	if e := hostWS.sendErr(apidto.ClientMessage{Type: apidto.CmdYield}); e.Code != "yield_not_allowed" {
		t.Fatalf("yield: %+v", e)
	}
	// 4S is immune (Jack of spades): no shield, but the 4 damage finishes the enemy exactly (16 + 4 = 20).
	tavernBefore := hm.View.TavernCount
	hostWS.send(apidto.ClientMessage{Type: apidto.CmdPlay, Cards: wcs("4S")})
	hm = hostWS.last
	if ev := hasEvent(hm.Events, apidto.EventEnemyDefeated); ev.Exact == nil || !*ev.Exact || ev.Card == nil || *ev.Card != wc("JS") {
		t.Fatalf("enemyDefeated = %+v", ev)
	}
	if ev := hasEvent(hm.Events, apidto.EventEnemyRevealed); ev == nil || ev.Card == nil || *ev.Card != wc("JC") {
		t.Fatalf("enemyRevealed = %+v", ev)
	}
	hv = *hm.View
	if hv.Enemy.Card != wc("JC") || hv.Enemy.Damage != 0 || hv.Enemy.Shield != 0 || hv.TavernCount != tavernBefore+1 ||
		hv.Turn.ActivePlayerID != host.PlayerID || hv.Turn.Step != "play" || hv.CastleCount != 10 || len(hv.Enemy.Plays) != 0 {
		t.Fatalf("after exact kill: enemy %+v turn %+v tavern %d (was %d) castle %d", hv.Enemy, hv.Turn, hv.TavernCount, tavernBefore, hv.CastleCount)
	}
	wantHand(t, "host", hv, "2S 6S 8S 10C")

	// Restart the server on the same database.
	rev := a.view(t, code, host.Token).Revision
	a.stop()
	a = startAppRNG(t, dbPath, func() domain.RNG { return identityRNG{} })
	hv2 := a.view(t, code, host.Token)
	if hv2.Revision != rev || hv2.Enemy.Card != wc("JC") || hv2.Turn.ActivePlayerID != host.PlayerID || hv2.TavernCount != hv.TavernCount {
		t.Fatalf("state changed across restart: rev %d vs %d, %+v", hv2.Revision, rev, hv2.Turn)
	}
	hostWS = a.dial(t, "host", code, host.Token)
	guestWS = a.dial(t, "guest", code, guest.Token)
	hostWS.presence(2)
	if hostWS.view.Revision != rev || guestWS.view.Revision != rev {
		t.Fatalf("revisions after reconnect: %d / %d, want %d", hostWS.view.Revision, guestWS.view.Revision, rev)
	}

	// Play on: 2S against the Jack of clubs is a 2-point shield; the guest-bound attack drops to 8.
	hostWS.send(apidto.ClientMessage{Type: apidto.CmdPlay, Cards: wcs("2S")})
	hv = *hostWS.last.View
	if hv.Enemy.Shield != 2 || hv.Enemy.EffectiveAttack != 8 || hv.Turn.DamageToCover != 8 || hv.Enemy.Damage != 2 {
		t.Fatalf("after 2S: %+v %+v", hv.Enemy, hv.Turn)
	}
	hostWS.send(apidto.ClientMessage{Type: apidto.CmdDiscard, Cards: wcs("8S")})
	gm = guestWS.untilView("guest turn again", func(m wireMsg) bool {
		return m.View.Turn.ActivePlayerID == guest.PlayerID && m.View.Turn.Step == "play"
	})
	if gm.View.Players[0].HandCount != 2 || len(gm.View.Hand) != 5 {
		t.Fatalf("guest sees %+v", gm.View.Players)
	}
	if hasRevealedHands(*gm.View) {
		t.Fatal("hands revealed mid-game")
	}
	checkNoLeak(gm.raw, *gm.View, 2)
}

func TestThreePlayerJesterFlow(t *testing.T) {
	a := startAppRNG(t, filepath.Join(t.TempDir(), "three.db"), func() domain.RNG { return identityRNG{} })
	p0 := a.create(t, "Ann")
	code := p0.GameID
	p1 := a.join(t, code, "Bob")
	p2 := a.join(t, code, "Cy")
	w0 := a.dial(t, "p0", code, p0.Token)
	w1 := a.dial(t, "p1", code, p1.Token)
	w2 := a.dial(t, "p2", code, p2.Token)
	w0.presence(3)

	w0.send(apidto.ClientMessage{Type: apidto.CmdStartGame, FirstPlayerID: p0.PlayerID})
	v0 := *w0.last.View
	w1.untilView("started", func(m wireMsg) bool { return m.View.Phase == "playing" })
	w2.untilView("started", func(m wireMsg) bool { return m.View.Phase == "playing" })
	if !slices.Contains(cardNames(v0.Hand), "X") || len(v0.Hand) != 6 || v0.Rules.MaxHandSize != 6 || v0.Rules.TavernJesters != 1 {
		t.Fatalf("seat 0 hand = %v rules %+v", cardNames(v0.Hand), v0.Rules)
	}

	w0.send(apidto.ClientMessage{Type: apidto.CmdPlay, Cards: wcs("X")})
	m0 := w0.last
	if hasEvent(m0.Events, apidto.EventJesterPlayed) == nil || !m0.View.Enemy.ImmunityCancelled || m0.View.Enemy.Damage != 0 {
		t.Fatalf("jester play: events %+v enemy %+v", m0.Events, m0.View.Enemy)
	}
	if len(m0.View.Enemy.Plays) != 1 || !m0.View.Enemy.Plays[0].Jester {
		t.Fatalf("jester play not recorded: %+v", m0.View.Enemy.Plays)
	}
	if !m0.View.Turn.CanYield && m0.View.Turn.ActivePlayerID != p0.PlayerID {
		t.Fatalf("turn = %+v", m0.View.Turn)
	}

	// Seat 1 signals; seat 2 sees it. The chooser cannot signal.
	sig := "goodPlay"
	w1.send(apidto.ClientMessage{Type: apidto.CmdSetSignal, Signal: &sig})
	m2 := w2.untilView("signal visible", func(m wireMsg) bool { return m.View.Players[1].Signal == "goodPlay" })
	if ev := hasEvent(m2.Events, apidto.EventSignalSet); ev == nil || ev.Signal != "goodPlay" || ev.PlayerID != p1.PlayerID {
		t.Fatalf("signalSet = %+v", ev)
	}
	if m2.View.Players[0].Signal != "none" || m2.View.Players[2].Signal != "none" {
		t.Fatalf("signals = %+v", m2.View.Players)
	}
	if e := w0.sendErr(apidto.ClientMessage{Type: apidto.CmdSetSignal, Signal: &sig}); e.Code != "illegal_signal" {
		t.Fatalf("chooser signal: %+v", e)
	}
	// Seat 1 may not choose.
	if e := w1.sendErr(apidto.ClientMessage{Type: apidto.CmdChooseNext, PlayerID: p2.PlayerID}); e.Code != "not_your_turn" {
		t.Fatalf("non-chooser chooseNext: %+v", e)
	}
	// Seat 0 chooses seat 2.
	w0.send(apidto.ClientMessage{Type: apidto.CmdChooseNext, PlayerID: p2.PlayerID})
	m2 = w2.untilView("seat 2 active", func(m wireMsg) bool {
		return m.View.Turn.ActivePlayerID == p2.PlayerID && m.View.Turn.Step == "play"
	})
	if hasEvent(m2.Events, apidto.EventNextPlayerChosen) == nil || !m2.View.Enemy.ImmunityCancelled {
		t.Fatalf("events %+v enemy %+v", m2.Events, m2.View.Enemy)
	}
	for _, p := range m2.View.Players {
		if p.Signal != "none" {
			t.Fatalf("signals not cleared: %+v", m2.View.Players)
		}
	}
	if !m2.View.Players[2].IsActive || m2.View.Players[0].IsActive {
		t.Fatalf("active flags: %+v", m2.View.Players)
	}
}

// soloPolicy decides the next REST action for the solo player from the view.
func soloStep(t *testing.T, a *app, code, token string, v apidto.PlayerView) apidto.PlayerView {
	t.Helper()
	base := "/api/games/" + code
	post := func(path string, body any) apidto.PlayerView {
		var out apidto.PlayerView
		a.call(t, http.MethodPost, base+path, token, body, &out, http.StatusOK)
		return out
	}
	value := func(c apidto.Card) int {
		switch c.Rank {
		case 11:
			return 10
		case 12:
			return 15
		case 13:
			return 20
		}
		return c.Rank
	}
	switch v.Turn.Step {
	case "play":
		if len(v.Hand) == 0 {
			if v.SoloJestersLeft > 0 {
				return post("/solo-jester", nil)
			}
			return post("/yield", nil)
		}
		e := v.Enemy
		score := func(c apidto.Card) float64 {
			dmg := float64(value(c))
			if c.Suit == "clubs" && (e.Immunity != "clubs" || e.ImmunityCancelled) {
				dmg *= 2
			}
			s := dmg
			if c.Suit == "spades" && (e.Immunity != "spades" || e.ImmunityCancelled) {
				s += 1.5 * float64(value(c))
			}
			if c.Suit == "diamonds" && len(v.Hand) <= 5 && (e.Immunity != "diamonds" || e.ImmunityCancelled) {
				s += 2 * float64(value(c))
			}
			if c.Suit == "hearts" && v.DiscardCount > 6 && (e.Immunity != "hearts" || e.ImmunityCancelled) {
				s += float64(value(c))
			}
			if int(dmg) >= e.Health-e.Damage { // a kill: prefer the cheapest
				s = 1000 - float64(value(c))
			}
			return s
		}
		best := v.Hand[0]
		for _, c := range v.Hand {
			if score(c) > score(best) {
				best = c
			}
		}
		// Use a Jester power when the hand is nearly spent.
		if len(v.Hand) <= 2 && v.SoloJestersLeft > 0 {
			return post("/solo-jester", nil)
		}
		return post("/play", apidto.CardsRequest{Cards: []apidto.Card{best}})
	case "discard":
		attack := v.Turn.DamageToCover
		total := 0
		for _, c := range v.Hand {
			total += value(c)
		}
		if total < attack {
			if v.SoloJestersLeft == 0 {
				t.Fatalf("hand total %d < attack %d with no Jester, yet the game is still running", total, attack)
			}
			return post("/solo-jester", nil)
		}
		// Smallest sum (then fewest cards) of a legal discard set, by brute force.
		var bestSet []apidto.Card
		bestSum := 1 << 30
		for mask := 1; mask < 1<<len(v.Hand); mask++ {
			var set []apidto.Card
			sum, minV := 0, 1<<30
			for i, c := range v.Hand {
				if mask&(1<<i) != 0 {
					set = append(set, c)
					sum += value(c)
					minV = min(minV, value(c))
				}
			}
			if sum < attack || sum-minV >= attack {
				continue
			}
			if sum < bestSum || (sum == bestSum && len(set) < len(bestSet)) {
				bestSet, bestSum = set, sum
			}
		}
		if bestSet == nil {
			t.Fatalf("no legal discard for hand %v attack %d", cardNames(v.Hand), attack)
		}
		return post("/discard", apidto.CardsRequest{Cards: bestSet})
	}
	t.Fatalf("unexpected step %q in phase %s", v.Turn.Step, v.Phase)
	return v
}

func TestSoloGameToCompletion(t *testing.T) {
	var won, lost int
	for seed := uint64(1); seed <= 12; seed++ {
		t.Run(fmt.Sprintf("seed%d", seed), func(t *testing.T) {
			a := startApp(t, filepath.Join(t.TempDir(), "solo.db"), seed)
			s := a.create(t, "Solo")
			code := s.GameID
			var v apidto.PlayerView
			a.call(t, http.MethodPost, "/api/games/"+code+"/start", s.Token, nil, &v, http.StatusOK)
			if v.Rules.MaxHandSize != 8 || len(v.Hand) != 8 || v.SoloJestersLeft != 2 || v.Rules.SoloJesters != 2 || v.TavernCount != 32 {
				t.Fatalf("solo setup: %+v hand %d tavern %d", v.Rules, len(v.Hand), v.TavernCount)
			}
			for i := 0; i < 2000 && v.Phase == "playing"; i++ {
				v = soloStep(t, a, code, s.Token, v)
				if got := a.view(t, code, s.Token); got.Revision != v.Revision {
					t.Fatalf("REST view revision %d != action view %d", got.Revision, v.Revision)
				}
			}
			if v.Phase != "won" && v.Phase != "lost" {
				t.Fatalf("game did not finish: phase %s", v.Phase)
			}
			r := v.Result
			if r == nil || r.Outcome != v.Phase || v.Turn == nil || v.Turn.Step != "" {
				t.Fatalf("result = %+v turn %+v", r, v.Turn)
			}
			switch v.Phase {
			case "won":
				won++
				want := map[int]string{0: "gold", 1: "silver", 2: "bronze"}[r.SoloJestersUsed]
				if r.Grade != want || r.EnemiesDefeated != 12 || r.Reason != "" || v.Enemy != nil {
					t.Fatalf("won result = %+v enemy %+v", r, v.Enemy)
				}
			case "lost":
				lost++
				if r.Grade != "" || (r.Reason != "cannotCoverDamage" && r.Reason != "cannotAct") || r.PlayerID != s.PlayerID || r.EnemiesDefeated >= 12 {
					t.Fatalf("lost result = %+v", r)
				}
			}
			if len(v.Hand) > 0 && !hasRevealedHands(v) { // an empty hand is omitted from the wire
				t.Fatal("hand not revealed after the end")
			}
			// Everything is rejected after the end, except a rematch.
			a.call(t, http.MethodPost, "/api/games/"+code+"/yield", s.Token, nil, nil, http.StatusConflict)
		})
	}
	t.Logf("solo results: %d won, %d lost", won, lost)
}

func TestRematchKeepsGroupAndTokens(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "rematch.db"), 5)
	ann := a.create(t, "Ann")
	code := ann.GameID
	bob := a.join(t, code, "Bob")
	a.call(t, http.MethodPost, "/api/games/"+code+"/rematch", ann.Token, nil, nil, http.StatusConflict) // not finished yet
	// An empty body means a random first player.
	var started apidto.PlayerView
	a.call(t, http.MethodPost, "/api/games/"+code+"/start", ann.Token, nil, &started, http.StatusOK)
	if started.Phase != "playing" || started.Turn == nil || started.Turn.Number != 1 {
		t.Fatalf("start: %+v", started)
	}

	annWS := a.dial(t, "ann", code, ann.Token)
	var fin apidto.PlayerView
	a.call(t, http.MethodPost, "/api/games/"+code+"/give-up", bob.Token, nil, &fin, http.StatusOK)
	if fin.Phase != "lost" || fin.Result == nil || fin.Result.Reason != "gaveUp" || fin.Result.PlayerID != bob.PlayerID || !hasRevealedHands(fin) {
		t.Fatalf("give up: %+v", fin)
	}

	var old apidto.PlayerView
	a.call(t, http.MethodPost, "/api/games/"+code+"/rematch", bob.Token, nil, &old, http.StatusOK)
	next := old.NextGameID
	if next == "" || next == code || old.Phase != "lost" {
		t.Fatalf("rematch: next=%q phase=%s", next, old.Phase)
	}
	// Ann's open socket learns about the next game.
	m := annWS.until("rematchCreated", func(m wireMsg) bool { return m.Type == apidto.MsgView && m.View.NextGameID != "" })
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
		if len(v.Hand) != 0 || hasRevealedHands(v) || v.Result != nil || v.Enemy != nil || v.Turn != nil {
			t.Fatalf("next game not fresh: %+v", v)
		}
	}
	a.call(t, http.MethodPost, "/api/games/"+next+"/start", ann.Token, nil, nil, http.StatusForbidden) // Ann is no longer host
	nextWS := a.dial(t, "ann-next", next, ann.Token)
	if nextWS.view.GameID != next {
		t.Fatalf("ws on next game: %+v", nextWS.view)
	}
	var again2 apidto.PlayerView
	a.call(t, http.MethodPost, "/api/games/"+next+"/start", bob.Token, apidto.StartGameRequest{FirstPlayerID: ann.PlayerID}, &again2, http.StatusOK)
	if again2.Phase != "playing" || again2.Turn.ActivePlayerID != ann.PlayerID {
		t.Fatalf("next game start: %s %+v", again2.Phase, again2.Turn)
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

	// The host cannot leave a lobby it is alone in, but the guest can leave.
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
	hv := a.view(t, code, host.Token)
	if len(hv.Players) != 1 || hv.Phase != "lobby" {
		t.Fatalf("host view = %+v", hv)
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

func hasRevealedHands(v apidto.PlayerView) bool {
	for _, p := range v.Players {
		if p.RevealedHand != nil {
			return true
		}
	}
	return false
}
