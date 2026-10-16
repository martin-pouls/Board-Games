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

	"github.com/martin-pouls/the-game/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-game/backend/internal/adapters/httpapi"
	"github.com/martin-pouls/the-game/backend/internal/adapters/sqlite"
	"github.com/martin-pouls/the-game/backend/internal/adapters/ws"
	"github.com/martin-pouls/the-game/backend/internal/core/domain"
	"github.com/martin-pouls/the-game/backend/internal/core/service"
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

func (a *app) create(t *testing.T, name string, opts *apidto.GameOptions) apidto.SessionResponse {
	t.Helper()
	var s apidto.SessionResponse
	a.call(t, http.MethodPost, "/api/games", "", apidto.CreateGameRequest{Name: name, Options: opts}, &s, http.StatusCreated)
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
	reqID int

	// events and raw belong to the last view message read, so a test can inspect
	// the update that a command caused after send returned.
	events []apidto.Event
	raw    []byte
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
		cl.events, cl.raw = m.Events, data
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

// try sends a command and reads until its ack or error reply. It returns the
// error, or nil on an ack. The ack follows the update the command published,
// so cl.view is current when try returns.
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
	reply := cl.until("reply to "+cmd.RequestID, func(m wireMsg) bool {
		return (m.Type == apidto.MsgAck || m.Type == apidto.MsgError) && m.RequestID == cmd.RequestID
	})
	return reply.Error
}

// send sends a command that must be accepted.
func (cl *client) send(cmd apidto.ClientMessage) {
	cl.t.Helper()
	if e := cl.try(cmd); e != nil {
		cl.t.Fatalf("%s: %s rejected: %+v", cl.name, cmd.Type, e)
	}
}

// reject sends a command that must be refused with the given error code.
func (cl *client) reject(cmd apidto.ClientMessage, code string) {
	cl.t.Helper()
	if e := cl.try(cmd); e == nil || e.Code != code {
		cl.t.Fatalf("%s: %s: got %+v, want error %s", cl.name, cmd.Type, e, code)
	}
}

// syncAll reads on every client until each has seen the newest revision, so no
// socket's send buffer fills up while another player acts.
func syncAll(clients map[string]*client) {
	var rev int64
	for _, cl := range clients {
		rev = max(rev, cl.view.Revision)
	}
	for _, cl := range clients {
		if cl.view.Revision < rev {
			cl.t.Helper()
			cl.until(fmt.Sprintf("revision %d", rev), func(m wireMsg) bool { return m.Type == apidto.MsgView && m.View.Revision >= rev })
		}
	}
}

func intp(i int) *int { return &i }

// pickPlay returns the legal play with the smallest jump (a backwards trick counts as best).
func pickPlay(v apidto.PlayerView) (card, pile int, ok bool) {
	best := 1 << 30
	for _, hc := range v.Hand {
		for _, p := range hc.LegalPiles {
			gap := hc.Value - v.Piles[p].TopValue
			if v.Piles[p].Direction == "down" {
				gap = -gap
			}
			if gap < 0 {
				gap = -100
			}
			if gap < best {
				best, card, pile, ok = gap, hc.Value, p, true
			}
		}
	}
	return card, pile, ok
}

func hasRevealedHands(v apidto.PlayerView) bool {
	for _, p := range v.Players {
		if p.RevealedHand != nil {
			return true
		}
	}
	return false
}

func handValues(v apidto.PlayerView) []int {
	out := make([]int, 0, len(v.Hand))
	for _, c := range v.Hand {
		out = append(out, c.Value)
	}
	return out
}

func TestTwoPlayerGameOverWebSocketSurvivesRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "e2e.db")
	a := startApp(t, dbPath, 7)

	ann := a.create(t, "Ann", nil)
	code := ann.GameID
	if len(code) != 5 || ann.View.Phase != "lobby" || !ann.View.Players[0].IsHost {
		t.Fatalf("create: unexpected session %+v", ann)
	}
	bob := a.join(t, strings.ToLower(code), "Bob") // codes are case-insensitive
	if bob.GameID != code || len(bob.View.Players) != 2 {
		t.Fatalf("join: unexpected session %+v", bob)
	}

	var sum apidto.GameSummary
	a.call(t, http.MethodGet, "/api/games/"+code, "", nil, &sum, http.StatusOK)
	if sum.HostName != "Ann" || sum.PlayerCount != 2 || !sum.Joinable {
		t.Fatalf("summary: %+v", sum)
	}
	// A bad token is rejected on REST and on WS.
	a.call(t, http.MethodGet, "/api/games/"+code+"/view", "nope", nil, nil, http.StatusUnauthorized)

	annWS := a.dial(t, "ann", code, ann.Token)
	annWS.presence(1)
	if annWS.view.You != ann.PlayerID || annWS.view.Phase != "lobby" || len(annWS.view.Players) != 2 {
		t.Fatalf("ann initial view: %+v", annWS.view)
	}
	bobWS := a.dial(t, "bob", code, bob.Token)
	bobWS.presence(2)
	annWS.presence(2)

	// Host toggles Pro and starts over WS; both see the dealt game.
	annWS.send(apidto.ClientMessage{Type: apidto.CmdSetOptions, Options: &apidto.GameOptions{Variant: "classic", Pro: true}})
	if !annWS.view.Options.Pro || annWS.view.Rules.MinCardsPerTurn != 3 {
		t.Fatalf("options not applied: %+v %+v", annWS.view.Options, annWS.view.Rules)
	}
	annWS.send(apidto.ClientMessage{Type: apidto.CmdStartGame})
	if annWS.view.Phase != "choosingFirstPlayer" || len(annWS.view.Hand) != 7 {
		t.Fatalf("after start: phase %s, hand %d", annWS.view.Phase, len(annWS.view.Hand))
	}
	bobWS.until("choosing view", func(m wireMsg) bool { return m.Type == apidto.MsgView && m.View.Phase == "choosingFirstPlayer" })
	if len(bobWS.view.Hand) != 7 || annWS.view.DrawPileCount != 98-14 {
		t.Fatalf("deal: bob hand %d, deck %d", len(bobWS.view.Hand), annWS.view.DrawPileCount)
	}
	if slices.ContainsFunc(handValues(annWS.view), func(v int) bool { return slices.Contains(handValues(bobWS.view), v) }) {
		t.Fatal("hands overlap")
	}

	// Only the host may choose; Bob's attempt is rejected without changing state.
	bobWS.reject(apidto.ClientMessage{Type: apidto.CmdChooseFirstPlayer, PlayerID: bob.PlayerID}, apidto.CodeNotHost)

	annWS.send(apidto.ClientMessage{Type: apidto.CmdChooseFirstPlayer, PlayerID: bob.PlayerID})
	bobWS.until("bob's turn", func(m wireMsg) bool {
		return m.Type == apidto.MsgView && m.View.Turn != nil && m.View.Turn.ActivePlayerID == bob.PlayerID
	})

	// Bob plays legal cards until he may end the turn, then ends it.
	for i := 0; !bobWS.view.Turn.CanEndTurn; i++ {
		if i > 10 {
			t.Fatal("bob never reached canEndTurn")
		}
		card, pile, ok := pickPlay(bobWS.view)
		if !ok {
			t.Fatalf("bob has no legal play: %+v", bobWS.view.Hand)
		}
		bobWS.send(apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: intp(card), Pile: intp(pile)})
		if slices.Contains(handValues(bobWS.view), card) {
			t.Fatalf("card %d still in hand after play", card)
		}
	}
	if bobWS.view.Turn.CardsPlayed != 3 {
		t.Fatalf("canEndTurn after %d cards, want 3 (Pro)", bobWS.view.Turn.CardsPlayed)
	}
	// Ann sees Bob's plays live, but never his hand.
	annWS.until("bob's 3rd play", func(m wireMsg) bool {
		return m.Type == apidto.MsgView && m.View.Turn != nil && m.View.Turn.CardsPlayed == 3
	})
	bobWS.send(apidto.ClientMessage{Type: apidto.CmdEndTurn})
	if len(bobWS.view.Hand) != 7 || bobWS.view.DrawPileCount != 98-14-3 {
		t.Fatalf("after end turn: bob hand %d, deck %d", len(bobWS.view.Hand), bobWS.view.DrawPileCount)
	}
	bobHand := handValues(bobWS.view)

	m := annWS.until("ann's turn", func(m wireMsg) bool {
		return m.Type == apidto.MsgView && m.View.Turn != nil && m.View.Turn.ActivePlayerID == ann.PlayerID
	})
	if !slices.ContainsFunc(m.Events, func(e apidto.Event) bool { return e.Type == apidto.EventTurnEnded }) {
		t.Fatalf("ann's update lacks turnEnded event: %+v", m.Events)
	}
	if m.View.You != ann.PlayerID || hasRevealedHands(*m.View) {
		t.Fatalf("ann's view leaks: you=%s revealed=%v", m.View.You, hasRevealedHands(*m.View))
	}
	for _, v := range bobHand {
		if slices.Contains(handValues(*m.View), v) {
			t.Fatalf("ann's hand contains bob's card %d", v)
		}
	}
	for _, v := range bobHand {
		if strings.Contains(string(m.raw), fmt.Sprintf(`"value":%d,`, v)) && !slices.Contains(handValues(*m.View), v) &&
			!slices.ContainsFunc(m.View.Piles, func(p apidto.Pile) bool { return p.TopValue == v }) {
			t.Fatalf("ann's raw update mentions bob's card %d", v)
		}
	}
	for _, p := range m.View.Players {
		if p.ID == bob.PlayerID && p.HandCount != 7 {
			t.Fatalf("bob's hand count as seen by ann = %d", p.HandCount)
		}
	}
	want := a.view(t, code, bob.Token)

	// Restart: reopen the same DB file with a fresh service and server.
	a.stop()
	a2 := startApp(t, dbPath, 99)
	got := a2.view(t, code, bob.Token)
	if got.You != bob.PlayerID || got.Revision != want.Revision || got.Phase != "playing" {
		t.Fatalf("restored view: you=%s rev=%d phase=%s, want %s rev=%d", got.You, got.Revision, got.Phase, bob.PlayerID, want.Revision)
	}
	gj, _ := json.Marshal(got)
	wj, _ := json.Marshal(want)
	if !bytes.Equal(gj, wj) {
		t.Fatalf("restored view differs:\n got %s\nwant %s", gj, wj)
	}
	// Ann reconnects over WS to the new server and can play.
	annWS2 := a2.dial(t, "ann2", code, ann.Token)
	if annWS2.view.Turn == nil || annWS2.view.Turn.ActivePlayerID != ann.PlayerID {
		t.Fatalf("ann reconnect view: %+v", annWS2.view.Turn)
	}
	card, pile, ok := pickPlay(annWS2.view)
	if !ok {
		t.Fatal("ann has no legal play after restart")
	}
	annWS2.send(apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: intp(card), Pile: intp(pile)})
	if annWS2.view.Turn.CardsPlayed != 1 {
		t.Fatalf("ann cardsPlayed = %d", annWS2.view.Turn.CardsPlayed)
	}
}

// playToEnd plays every turn over the players' WebSockets with a greedy
// smallest-jump strategy (ending the turn as soon as the minimum is met) until
// the game finishes. All clients must be connected and in sync.
func playToEnd(t *testing.T, clients map[string]*client) apidto.PlayerView {
	t.Helper()
	var anyone *client
	for _, cl := range clients {
		anyone = cl
		break
	}
	for range 2000 {
		if over(anyone.view) {
			return anyone.view
		}
		cl := clients[anyone.view.Turn.ActivePlayerID]
		v := cl.view
		if hasRevealedHands(v) {
			t.Fatal("hands revealed before the game finished")
		}
		if v.Turn.CanEndTurn {
			cl.send(apidto.ClientMessage{Type: apidto.CmdEndTurn})
			syncAll(clients)
			continue
		}
		card, pile, ok := pickPlay(v)
		if !ok {
			// Stuck mid-turn: the bot gives up instead of undoing.
			if !v.Turn.Stuck || v.Turn.CardsPlayed == 0 {
				t.Fatalf("no play, no end turn, but not stuck mid-turn: %+v", v)
			}
			cl.send(apidto.ClientMessage{Type: apidto.CmdGiveUp})
			syncAll(clients)
			continue
		}
		if v.Turn.Stuck {
			t.Fatalf("stuck although %d fits on pile %d", card, pile)
		}
		cl.send(apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: intp(card), Pile: intp(pile)})
		syncAll(clients)
	}
	t.Fatal("game did not finish")
	return apidto.PlayerView{}
}

// dialAll connects every session to game code over WebSocket.
func (a *app) dialAll(t *testing.T, code string, sessions ...apidto.SessionResponse) map[string]*client {
	t.Helper()
	clients := make(map[string]*client, len(sessions))
	for _, s := range sessions {
		clients[s.PlayerID] = a.dial(t, s.PlayerID, code, s.Token)
	}
	return clients
}

func checkFinished(t *testing.T, v apidto.PlayerView) {
	t.Helper()
	r := v.Result
	if !over(v) || r == nil {
		t.Fatalf("not finished: phase=%s result=%+v", v.Phase, r)
	}
	if r.Score != r.CardsLeftInHands+r.CardsLeftInDrawPile {
		t.Fatalf("score %d != %d + %d", r.Score, r.CardsLeftInHands, r.CardsLeftInDrawPile)
	}
	// A won game always has score 0. A lost game has score 0 only when the last
	// card played was an uncovered Skull (Extreme), i.e. a skull is on top of a pile.
	skullOnTop := slices.ContainsFunc(v.Piles, func(p apidto.Pile) bool { return p.TopCard != nil && p.TopCard.Command == "skull" })
	if v.Phase == "won" && r.Score != 0 || v.Phase == "lost" && r.Score == 0 && !skullOnTop || r.Excellent != (r.Score < 10) {
		t.Fatalf("inconsistent result %+v", r)
	}
	if v.Phase == "lost" && (r.PlayerID == "" || r.Reason != "stuck" && r.Reason != "gaveUp") || v.Phase == "won" && (r.PlayerID != "" || r.Reason != "") {
		t.Fatalf("inconsistent loss reason: %s %+v", v.Phase, r)
	}
	revealed := 0
	for _, p := range v.Players {
		if len(p.RevealedHand) != p.HandCount {
			t.Fatalf("%s: revealed %d cards, handCount %d", p.Name, len(p.RevealedHand), p.HandCount)
		}
		revealed += len(p.RevealedHand)
	}
	if revealed != r.CardsLeftInHands {
		t.Fatalf("revealed %d cards, result says %d in hands", revealed, r.CardsLeftInHands)
	}
}

func TestSoloGameToCompletion(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3} {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			a := startApp(t, filepath.Join(t.TempDir(), "solo.db"), seed)
			s := a.create(t, "Solo", nil)
			clients := a.dialAll(t, s.GameID, s)
			cl := clients[s.PlayerID]
			cl.send(apidto.ClientMessage{Type: apidto.CmdStartGame})
			v := cl.view
			if v.Phase != "playing" || len(v.Hand) != 8 || v.Turn.ActivePlayerID != s.PlayerID {
				t.Fatalf("solo start: phase %s hand %d", v.Phase, len(v.Hand))
			}
			v = playToEnd(t, clients)
			checkFinished(t, v)
			t.Logf("seed %d: %s, score %d", seed, v.Phase, v.Result.Score)
			// Further actions are rejected once finished.
			cl.reject(apidto.ClientMessage{Type: apidto.CmdEndTurn}, apidto.CodeWrongPhase)
		})
	}
}

func TestRematchKeepsGroupAndTokens(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "rematch.db"), 5)
	ann := a.create(t, "Ann", nil)
	code := ann.GameID
	bob := a.join(t, code, "Bob")
	clients := a.dialAll(t, code, ann, bob)
	annWS := clients[ann.PlayerID]
	annWS.reject(apidto.ClientMessage{Type: apidto.CmdRematch}, apidto.CodeWrongPhase) // not finished yet
	annWS.send(apidto.ClientMessage{Type: apidto.CmdStartGame})
	annWS.send(apidto.ClientMessage{Type: apidto.CmdChooseFirstPlayer, PlayerID: ann.PlayerID})
	syncAll(clients)

	fin := playToEnd(t, clients)
	checkFinished(t, fin)

	bobWS := clients[bob.PlayerID]
	bobWS.send(apidto.ClientMessage{Type: apidto.CmdRematch})
	old := bobWS.view
	next := old.NextGameID
	if next == "" || next == code || !over(old) {
		t.Fatalf("rematch: next=%q phase=%s", next, old.Phase)
	}
	// Ann's open socket learns about the next game.
	m := annWS.until("rematchCreated", func(m wireMsg) bool { return m.Type == apidto.MsgView && m.View.NextGameID != "" })
	if m.View.NextGameID != next || !slices.ContainsFunc(m.Events, func(e apidto.Event) bool { return e.Type == apidto.EventRematchCreated && e.GameID == next }) {
		t.Fatalf("ann's rematch update: next=%q events=%+v", m.View.NextGameID, m.Events)
	}
	// Idempotent: a second request is accepted and still points at the same next game.
	annWS.send(apidto.ClientMessage{Type: apidto.CmdRematch})
	if again := a.view(t, code, ann.Token); again.NextGameID != next {
		t.Fatalf("second rematch: %q, want %q", again.NextGameID, next)
	}

	// Same tokens work on the new game; it is a lobby with both players, Bob hosting.
	for _, s := range []apidto.SessionResponse{ann, bob} {
		v := a.view(t, next, s.Token)
		if v.You != s.PlayerID || v.Phase != "lobby" || v.HostID != bob.PlayerID || len(v.Players) != 2 || v.Revision != 1 {
			t.Fatalf("next game view for %s: %+v", s.PlayerID, v)
		}
		if len(v.Hand) != 0 || hasRevealedHands(v) || v.Result != nil {
			t.Fatalf("next game not fresh: %+v", v)
		}
	}
	nextWS := a.dial(t, "ann-next", next, ann.Token)
	if nextWS.view.GameID != next {
		t.Fatalf("ws on next game: %+v", nextWS.view)
	}
	nextWS.reject(apidto.ClientMessage{Type: apidto.CmdStartGame}, apidto.CodeNotHost) // Ann is no longer host
	bobNextWS := a.dial(t, "bob-next", next, bob.Token)
	bobNextWS.send(apidto.ClientMessage{Type: apidto.CmdStartGame})
	if bobNextWS.view.Phase != "choosingFirstPlayer" {
		t.Fatalf("next game start: %s", bobNextWS.view.Phase)
	}
}

func TestExtremeSoloDeterministic(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "extreme.db")
	a := startAppRNG(t, dbPath, func() domain.RNG { return identityRNG{} })
	s := a.create(t, "Solo", &apidto.GameOptions{Variant: "extreme"})
	code := s.GameID
	ws := a.dial(t, "solo", code, s.Token)
	ws.send(apidto.ClientMessage{Type: apidto.CmdStartGame})
	v := ws.view

	// Identity RNG: ascending deck, hand 92..99, draw pile top 91.
	if v.Phase != "playing" || v.Options.Variant != "extreme" || !slices.Equal(handValues(v), []int{92, 93, 94, 95, 96, 97, 98, 99}) {
		t.Fatalf("start: phase %s variant %s hand %v", v.Phase, v.Options.Variant, handValues(v))
	}
	for _, hc := range v.Hand {
		want := map[int]string{92: "noCommunication", 95: "stop"}[hc.Value]
		if hc.Command != want {
			t.Errorf("card %d command %q, want %q", hc.Value, hc.Command, want)
		}
	}
	if v.ActiveCommands == nil || len(v.ActiveCommands) != 0 || v.CommunicationRestricted {
		t.Fatalf("activeCommands %v restricted %v", v.ActiveCommands, v.CommunicationRestricted)
	}
	if c := v.Turn.Constraints; c.ExactCount != nil || c.SinglePile || c.BoundPile != nil || c.SkullPiles == nil || len(c.SkullPiles) != 0 || c.StopPlayed {
		t.Fatalf("constraints = %+v", c)
	}

	ws.reject(apidto.ClientMessage{Type: apidto.CmdEndTurn}, apidto.CodeMinimumNotReached)

	// STOP overrides the minimum and blocks every further card; the turn stays open.
	ws.send(apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: intp(95), Pile: intp(0)})
	v = ws.view
	if v.Turn.Number != 1 || v.Turn.CardsPlayed != 1 || !v.Turn.CanEndTurn || !v.Turn.Constraints.StopPlayed || v.Turn.Stuck {
		t.Fatalf("after STOP: turn %+v", v.Turn)
	}
	for _, hc := range v.Hand {
		if len(hc.LegalPiles) != 0 || hc.Blocked != apidto.CodeStopPlayed {
			t.Fatalf("after STOP: card %+v", hc)
		}
	}
	ws.reject(apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: intp(96), Pile: intp(0)}, apidto.CodeStopPlayed)

	// Undo takes the STOP back into the hand.
	ws.send(apidto.ClientMessage{Type: apidto.CmdUndoPlay})
	v = ws.view
	if v.Turn.CardsPlayed != 0 || v.Turn.Constraints.StopPlayed || v.Piles[0].TopCard != nil || !slices.Contains(handValues(v), 95) {
		t.Fatalf("after undo: turn %+v pile 0 %+v hand %v", v.Turn, v.Piles[0], handValues(v))
	}
	if len(ws.events) != 1 || ws.events[0].Type != apidto.EventPlayUndone || ws.events[0].Card == nil || ws.events[0].Card.Value != 95 ||
		ws.events[0].Pile == nil || *ws.events[0].Pile != 0 {
		t.Fatalf("undo events = %s", ws.raw)
	}
	ws.reject(apidto.ClientMessage{Type: apidto.CmdUndoPlay}, apidto.CodeNothingToUndo)

	// STOP again, then End turn: the draw refills the hand.
	ws.send(apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: intp(95), Pile: intp(0)})
	ws.send(apidto.ClientMessage{Type: apidto.CmdEndTurn})
	v = ws.view
	if v.Turn.Number != 2 || v.Turn.CardsPlayed != 0 || !slices.Equal(handValues(v), []int{91, 92, 93, 94, 96, 97, 98, 99}) {
		t.Fatalf("after STOP: turn %+v hand %v", v.Turn, handValues(v))
	}
	if top := v.Piles[0].TopCard; top == nil || top.Command != "stop" || v.Piles[0].CommandActive {
		t.Fatalf("pile 0 = %+v", v.Piles[0])
	}
	if len(ws.events) != 2 || ws.events[0].Type != apidto.EventTurnEnded || ws.events[0].CardsDrawn == nil || *ws.events[0].CardsDrawn != 1 ||
		ws.events[1].Type != apidto.EventTurnStarted {
		t.Fatalf("end turn events = %s", ws.raw)
	}

	// No communication on a descending pile is an Infinity card: active while visible.
	ws.send(apidto.ClientMessage{Type: apidto.CmdPlayCard, Card: intp(92), Pile: intp(2)})
	v = ws.view
	wantActive := []apidto.ActiveCommand{{Command: "noCommunication", Pile: 2}}
	if !v.CommunicationRestricted || !reflect.DeepEqual(v.ActiveCommands, wantActive) {
		t.Fatalf("after NC: restricted %v active %v", v.CommunicationRestricted, v.ActiveCommands)
	}

	// Restart on the same DB: commands survive.
	a.stop()
	a2 := startAppRNG(t, dbPath, func() domain.RNG { return identityRNG{} })
	got := a2.view(t, code, s.Token)
	if got.Options.Variant != "extreme" || !got.CommunicationRestricted || !reflect.DeepEqual(got.ActiveCommands, wantActive) ||
		got.Piles[0].TopCard == nil || got.Piles[0].TopCard.Command != "stop" ||
		got.Piles[2].TopCard == nil || got.Piles[2].TopCard.Command != "noCommunication" || !got.Piles[2].CommandActive {
		t.Fatalf("restored view: %+v", got)
	}
}

func TestExtremeGamesToCompletion(t *testing.T) {
	for _, players := range []int{1, 3} {
		for _, pro := range []bool{false, true} {
			for _, seed := range []uint64{1, 2, 3} {
				if pro && seed != 1 {
					continue
				}
				t.Run(fmt.Sprintf("%dp/pro=%v/seed%d", players, pro, seed), func(t *testing.T) {
					a := startApp(t, filepath.Join(t.TempDir(), "x.db"), seed)
					host := a.create(t, "P1", &apidto.GameOptions{Variant: "extreme", Pro: pro})
					code := host.GameID
					sessions := []apidto.SessionResponse{host}
					for i := 2; i <= players; i++ {
						sessions = append(sessions, a.join(t, code, fmt.Sprintf("P%d", i)))
					}
					clients := a.dialAll(t, code, sessions...)
					hostWS := clients[host.PlayerID]
					hostWS.send(apidto.ClientMessage{Type: apidto.CmdStartGame})
					if players > 1 {
						hostWS.send(apidto.ClientMessage{Type: apidto.CmdChooseFirstPlayer, PlayerID: host.PlayerID})
					}
					syncAll(clients)
					v := playToEnd(t, clients)
					checkFinished(t, v)
					t.Logf("%s, score %d", v.Phase, v.Result.Score)
				})
			}
		}
	}
}

// over reports whether the game was won or lost.
func over(v apidto.PlayerView) bool { return v.Phase == "won" || v.Phase == "lost" }

// eventOf returns the first event of type typ, or nil.
func eventOf(evs []apidto.Event, typ string) *apidto.Event {
	for i := range evs {
		if evs[i].Type == typ {
			return &evs[i]
		}
	}
	return nil
}

func TestLeaveAndEndLobby(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "leave.db"), 3)
	host := a.create(t, "Ann", nil)
	code := host.GameID
	guest := a.join(t, code, "Bob")
	hostWS := a.dial(t, "host", code, host.Token)
	guestWS := a.dial(t, "guest", code, guest.Token)
	hostWS.presence(2)

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
	m := hostWS.until("playerLeft", func(m wireMsg) bool {
		return m.Type == apidto.MsgView && eventOf(m.Events, apidto.EventPlayerLeft) != nil
	})
	if ev := eventOf(m.Events, apidto.EventPlayerLeft); ev.PlayerID != guest.PlayerID || ev.HostID != host.PlayerID {
		t.Fatalf("playerLeft = %+v", ev)
	}
	if len(m.View.Players) != 1 || m.View.Players[0].ID != host.PlayerID {
		t.Fatalf("host players = %+v", m.View.Players)
	}
	// The sole remaining player cannot leave, but can close the lobby.
	var e apidto.ErrorResponse
	a.call(t, http.MethodPost, "/api/games/"+code+"/leave", host.Token, nil, &e, http.StatusConflict)
	if e.Error.Code != apidto.CodeCannotLeave {
		t.Fatalf("sole leave: %+v", e)
	}
	hostWS.send(apidto.ClientMessage{Type: apidto.CmdEndGame})
	if hostWS.view.Phase != "ended" || hostWS.view.Result != nil || eventOf(hostWS.events, apidto.EventGameEnded) == nil {
		t.Fatalf("after end game: phase %s events %+v", hostWS.view.Phase, hostWS.events)
	}
	var sum apidto.GameSummary
	a.call(t, http.MethodGet, "/api/games/"+code, "", nil, &sum, http.StatusOK)
	if sum.Phase != "ended" || sum.Joinable {
		t.Fatalf("summary after end: %+v", sum)
	}
	hostWS.reject(apidto.ClientMessage{Type: apidto.CmdStartGame}, apidto.CodeWrongPhase)
	hostWS.reject(apidto.ClientMessage{Type: apidto.CmdRematch}, apidto.CodeWrongPhase)
}

// TestGiveUpByAnyPlayer: a guest gives up while the host is still choosing who starts.
func TestGiveUpByAnyPlayer(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "giveup.db"), 4)
	ann := a.create(t, "Ann", nil)
	code := ann.GameID
	bob := a.join(t, code, "Bob")
	clients := a.dialAll(t, code, ann, bob)
	annWS, bobWS := clients[ann.PlayerID], clients[bob.PlayerID]
	bobWS.reject(apidto.ClientMessage{Type: apidto.CmdGiveUp}, apidto.CodeWrongPhase) // nothing to give up in the lobby
	annWS.send(apidto.ClientMessage{Type: apidto.CmdStartGame})
	syncAll(clients)

	bobWS.send(apidto.ClientMessage{Type: apidto.CmdGiveUp})
	v := bobWS.view
	if v.Phase != "lost" || v.Result == nil || v.Result.Reason != "gaveUp" || v.Result.PlayerID != bob.PlayerID ||
		v.Result.Score != 98 || v.Turn != nil || !hasRevealedHands(v) {
		t.Fatalf("after give up: phase %s result %+v turn %+v", v.Phase, v.Result, v.Turn)
	}
	m := annWS.until("gameLost", func(m wireMsg) bool {
		return m.Type == apidto.MsgView && eventOf(m.Events, apidto.EventGameLost) != nil
	})
	if ev := eventOf(m.Events, apidto.EventGameLost); ev.PlayerID != bob.PlayerID || ev.Reason != "gaveUp" || ev.Score == nil || *ev.Score != 98 {
		t.Fatalf("gameLost = %+v", ev)
	}
	checkFinished(t, v)
	annWS.reject(apidto.ClientMessage{Type: apidto.CmdEndGame}, apidto.CodeWrongPhase)
	annWS.reject(apidto.ClientMessage{Type: apidto.CmdGiveUp}, apidto.CodeWrongPhase)
	annWS.send(apidto.ClientMessage{Type: apidto.CmdRematch})
	if annWS.view.NextGameID == "" {
		t.Fatal("no rematch after giving up")
	}
}
