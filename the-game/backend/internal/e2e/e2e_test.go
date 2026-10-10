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

func startApp(t *testing.T, dbPath string, seed uint64) *app {
	t.Helper()
	db, err := sqlite.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	hub := ws.NewHub(nil)
	svc := service.New(sqlite.NewGameRepository(db), hub, service.Config{
		NewRNG: func() domain.RNG { return rand.New(rand.NewPCG(seed, seed+1)) },
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
	bobWS.reqID++
	b, _ := json.Marshal(apidto.ClientMessage{Type: apidto.CmdChooseFirstPlayer, RequestID: strconv.Itoa(bobWS.reqID), PlayerID: bob.PlayerID})
	if err := bobWS.c.Write(context.Background(), websocket.MessageText, b); err != nil {
		t.Fatal(err)
	}
	if m := bobWS.until("error", func(m wireMsg) bool { return m.Type == apidto.MsgError }); m.Error.Code != apidto.CodeNotHost {
		t.Fatalf("bob choosing: got %+v, want not_host", m.Error)
	}

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

// playToEnd plays every turn over REST with a greedy smallest-jump strategy
// (ending the turn as soon as the minimum is met) until the game finishes.
func playToEnd(t *testing.T, a *app, code string, tokens map[string]string) apidto.PlayerView {
	t.Helper()
	var anyTok string
	for _, tok := range tokens {
		anyTok = tok
		break
	}
	for range 2000 {
		v := a.view(t, code, anyTok)
		if v.Phase == "finished" {
			return v
		}
		tok := tokens[v.Turn.ActivePlayerID]
		v = a.view(t, code, tok)
		if hasRevealedHands(v) {
			t.Fatal("hands revealed before the game finished")
		}
		if v.Turn.CanEndTurn {
			a.call(t, http.MethodPost, "/api/games/"+code+"/end-turn", tok, nil, nil, http.StatusOK)
			continue
		}
		card, pile, ok := pickPlay(v)
		if !ok {
			t.Fatalf("active player stuck but game not finished: %+v", v)
		}
		a.call(t, http.MethodPost, "/api/games/"+code+"/plays", tok, apidto.PlayCardRequest{Card: intp(card), Pile: intp(pile)}, nil, http.StatusOK)
	}
	t.Fatal("game did not finish")
	return apidto.PlayerView{}
}

func checkFinished(t *testing.T, v apidto.PlayerView) {
	t.Helper()
	r := v.Result
	if v.Phase != "finished" || r == nil {
		t.Fatalf("not finished: phase=%s result=%+v", v.Phase, r)
	}
	if r.Score != r.CardsLeftInHands+r.CardsLeftInDrawPile {
		t.Fatalf("score %d != %d + %d", r.Score, r.CardsLeftInHands, r.CardsLeftInDrawPile)
	}
	if (r.Outcome == "won") != (r.Score == 0) || r.Excellent != (r.Score < 10) {
		t.Fatalf("inconsistent result %+v", r)
	}
	if r.Outcome == "lost" && r.StuckPlayerID == "" {
		t.Fatalf("lost without stuck player: %+v", r)
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
			var v apidto.PlayerView
			a.call(t, http.MethodPost, "/api/games/"+s.GameID+"/start", s.Token, nil, &v, http.StatusOK)
			if v.Phase != "playing" || len(v.Hand) != 8 || v.Turn.ActivePlayerID != s.PlayerID {
				t.Fatalf("solo start: phase %s hand %d", v.Phase, len(v.Hand))
			}
			v = playToEnd(t, a, s.GameID, map[string]string{s.PlayerID: s.Token})
			checkFinished(t, v)
			t.Logf("seed %d: %s, score %d", seed, v.Result.Outcome, v.Result.Score)
			// Further actions are rejected once finished.
			a.call(t, http.MethodPost, "/api/games/"+s.GameID+"/end-turn", s.Token, nil, nil, http.StatusConflict)
		})
	}
}

func TestRematchKeepsGroupAndTokens(t *testing.T) {
	a := startApp(t, filepath.Join(t.TempDir(), "rematch.db"), 5)
	ann := a.create(t, "Ann", nil)
	code := ann.GameID
	bob := a.join(t, code, "Bob")
	a.call(t, http.MethodPost, "/api/games/"+code+"/rematch", ann.Token, nil, nil, http.StatusConflict) // not finished yet
	a.call(t, http.MethodPost, "/api/games/"+code+"/start", ann.Token, nil, nil, http.StatusOK)
	a.call(t, http.MethodPost, "/api/games/"+code+"/first-player", ann.Token, apidto.ChooseFirstPlayerRequest{PlayerID: ann.PlayerID}, nil, http.StatusOK)

	annWS := a.dial(t, "ann", code, ann.Token)
	fin := playToEnd(t, a, code, map[string]string{ann.PlayerID: ann.Token, bob.PlayerID: bob.Token})
	checkFinished(t, fin)

	var old apidto.PlayerView
	a.call(t, http.MethodPost, "/api/games/"+code+"/rematch", bob.Token, nil, &old, http.StatusOK)
	next := old.NextGameID
	if next == "" || next == code || old.Phase != "finished" {
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
		if len(v.Hand) != 0 || hasRevealedHands(v) || v.Result != nil {
			t.Fatalf("next game not fresh: %+v", v)
		}
	}
	a.call(t, http.MethodPost, "/api/games/"+next+"/start", ann.Token, nil, nil, http.StatusForbidden) // Ann is no longer host
	nextWS := a.dial(t, "ann-next", next, ann.Token)
	if nextWS.view.GameID != next {
		t.Fatalf("ws on next game: %+v", nextWS.view)
	}
	var started apidto.PlayerView
	a.call(t, http.MethodPost, "/api/games/"+next+"/start", bob.Token, nil, &started, http.StatusOK)
	if started.Phase != "choosingFirstPlayer" {
		t.Fatalf("next game start: %s", started.Phase)
	}
}
