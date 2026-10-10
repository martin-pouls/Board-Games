package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/martin-pouls/regicide/backend/internal/adapters/apidto"
	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/ports"
)

// fakeService is a canned ports.GameService. Game "ABCDE" exists; token
// "tok-host" is p_host and "tok-guest" is p_guest. err, when set, is
// returned by every call except Authenticate.
type fakeService struct {
	mu    sync.Mutex
	err   error
	calls []string

	lastName   string
	lastFirst  domain.PlayerID
	lastNext   domain.PlayerID
	lastCards  []domain.Card
	lastSignal domain.JesterSignal
	lastID     domain.GameID
}

var _ ports.GameService = (*fakeService)(nil)

const testGame domain.GameID = "ABCDE"

var testTokens = map[string]domain.PlayerID{"tok-host": "p_host", "tok-guest": "p_guest"}

func (f *fakeService) record(name string, id domain.GameID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, name)
	f.lastID = id
	return f.err
}

func (f *fakeService) lastCall() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.calls) == 0 {
		return ""
	}
	return f.calls[len(f.calls)-1]
}

func fakeView(id domain.GameID, p domain.PlayerID) domain.PlayerView {
	return domain.PlayerView{
		GameID:   id,
		Revision: 7,
		Phase:    domain.PhaseLobby,
		You:      p,
		HostID:   "p_host",
		Players:  []domain.PlayerSummary{{ID: p, Name: "Ann", IsYou: true}},
	}
}

func (f *fakeService) session(id domain.GameID, name string) (ports.Session, error) {
	f.mu.Lock()
	f.lastName = name
	f.mu.Unlock()
	return ports.Session{GameID: id, PlayerID: "p_new", Token: "tok-new", View: fakeView(id, "p_new")}, nil
}

func (f *fakeService) CreateGame(_ context.Context, name string) (ports.Session, error) {
	if err := f.record("CreateGame", testGame); err != nil {
		return ports.Session{}, err
	}
	return f.session(testGame, name)
}

func (f *fakeService) JoinGame(_ context.Context, id domain.GameID, name string) (ports.Session, error) {
	if err := f.record("JoinGame", id); err != nil {
		return ports.Session{}, err
	}
	if id != testGame {
		return ports.Session{}, ports.ErrGameNotFound
	}
	return f.session(id, name)
}

func (f *fakeService) Authenticate(_ context.Context, id domain.GameID, token string) (domain.PlayerID, error) {
	if id != testGame {
		return "", ports.ErrGameNotFound
	}
	if p, ok := testTokens[token]; ok {
		return p, nil
	}
	return "", ports.ErrUnauthorized
}

func (f *fakeService) Summary(_ context.Context, id domain.GameID) (domain.GameSummary, error) {
	if err := f.record("Summary", id); err != nil {
		return domain.GameSummary{}, err
	}
	if id != testGame {
		return domain.GameSummary{}, ports.ErrGameNotFound
	}
	return domain.GameSummary{GameID: id, Phase: domain.PhaseLobby, HostName: "Ann", PlayerCount: 1, MaxPlayers: 4, Joinable: true}, nil
}

func (f *fakeService) viewCall(name string, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	if err := f.record(name, id); err != nil {
		return domain.PlayerView{}, err
	}
	return fakeView(id, p), nil
}

func (f *fakeService) View(_ context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return f.viewCall("View", id, p)
}

func (f *fakeService) LeaveGame(_ context.Context, id domain.GameID, _ domain.PlayerID) error {
	return f.record("LeaveGame", id)
}

func (f *fakeService) StartGame(_ context.Context, id domain.GameID, p, first domain.PlayerID) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastFirst = first
	f.mu.Unlock()
	return f.viewCall("StartGame", id, p)
}

func (f *fakeService) Play(_ context.Context, id domain.GameID, p domain.PlayerID, cards []domain.Card) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastCards = cards
	f.mu.Unlock()
	return f.viewCall("Play", id, p)
}

func (f *fakeService) Yield(_ context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return f.viewCall("Yield", id, p)
}

func (f *fakeService) Discard(_ context.Context, id domain.GameID, p domain.PlayerID, cards []domain.Card) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastCards = cards
	f.mu.Unlock()
	return f.viewCall("Discard", id, p)
}

func (f *fakeService) ChooseNext(_ context.Context, id domain.GameID, p, next domain.PlayerID) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastNext = next
	f.mu.Unlock()
	return f.viewCall("ChooseNext", id, p)
}

func (f *fakeService) UseSoloJester(_ context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return f.viewCall("UseSoloJester", id, p)
}

func (f *fakeService) SetJesterSignal(_ context.Context, id domain.GameID, p domain.PlayerID, s domain.JesterSignal) (domain.PlayerView, error) {
	f.mu.Lock()
	f.lastSignal = s
	f.mu.Unlock()
	return f.viewCall("SetJesterSignal", id, p)
}

func (f *fakeService) GiveUp(_ context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	return f.viewCall("GiveUp", id, p)
}

func (f *fakeService) Rematch(_ context.Context, id domain.GameID, p domain.PlayerID) (domain.PlayerView, error) {
	v, err := f.viewCall("Rematch", id, p)
	v.NextGameID = "NEXT2"
	return v, err
}

// --- helpers ---

func newTestRouter(t *testing.T, svc *fakeService) http.Handler {
	t.Helper()
	return NewRouter(Deps{
		Logger: slog.New(slog.DiscardHandler),
		Health: func(context.Context) error { return nil },
		Games:  svc,
	})
}

func do(t *testing.T, h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func decode[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	var v T
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %s: %v", w.Body.String(), err)
	}
	return v
}

func wantError(t *testing.T, w *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("status = %d, want %d (body %s)", w.Code, status, w.Body.String())
	}
	e := decode[apidto.ErrorResponse](t, w)
	if e.Error.Code != code {
		t.Fatalf("error code = %q, want %q (body %s)", e.Error.Code, code, w.Body.String())
	}
}

// --- tests ---

func TestHealth(t *testing.T) {
	h := newTestRouter(t, &fakeService{})
	w := do(t, h, "GET", "/api/health", "", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"ok"`) {
		t.Fatalf("health = %d %s", w.Code, w.Body.String())
	}
}

func TestCreateGame(t *testing.T) {
	svc := &fakeService{}
	h := newTestRouter(t, svc)

	w := do(t, h, "POST", "/api/games", "", `{"name":"Ann"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", w.Code, w.Body.String())
	}
	s := decode[apidto.SessionResponse](t, w)
	if s.GameID != "ABCDE" || s.PlayerID != "p_new" || s.Token != "tok-new" || s.View.You != "p_new" {
		t.Fatalf("session = %+v", s)
	}
	if svc.lastName != "Ann" {
		t.Fatalf("service got name %q", svc.lastName)
	}
	// Slices are never null on the wire.
	if !strings.Contains(w.Body.String(), `"hand":[]`) {
		t.Fatalf("hand should be [] in %s", w.Body.String())
	}
}

func TestBadBodies(t *testing.T) {
	h := newTestRouter(t, &fakeService{})
	big := `{"name":"` + strings.Repeat("x", 5000) + `"}`
	tests := []struct {
		name, method, path, token, body string
	}{
		{"create malformed", "POST", "/api/games", "", `{"name":`},
		{"create empty", "POST", "/api/games", "", ""},
		{"create wrong type", "POST", "/api/games", "", `{"name":5}`},
		{"create too large", "POST", "/api/games", "", big},
		{"join malformed", "POST", "/api/games/ABCDE/players", "", `nope`},
		{"start malformed", "POST", "/api/games/ABCDE/start", "tok-host", `[`},
		{"play empty body", "POST", "/api/games/ABCDE/play", "tok-host", ""},
		{"play null cards", "POST", "/api/games/ABCDE/play", "tok-host", `{"cards":null}`},
		{"play missing cards", "POST", "/api/games/ABCDE/play", "tok-host", `{}`},
		{"play invalid card", "POST", "/api/games/ABCDE/play", "tok-host", `{"cards":[{"suit":"hearts","rank":14}]}`},
		{"play jester with suit", "POST", "/api/games/ABCDE/play", "tok-host", `{"cards":[{"suit":"clubs","rank":0}]}`},
		{"play too large", "POST", "/api/games/ABCDE/play", "tok-host", big},
		{"discard null cards", "POST", "/api/games/ABCDE/discard", "tok-host", `{"cards":null}`},
		{"next-player missing id", "POST", "/api/games/ABCDE/next-player", "tok-host", `{}`},
		{"signal null", "PUT", "/api/games/ABCDE/signal", "tok-host", `{"signal":null}`},
		{"signal missing", "PUT", "/api/games/ABCDE/signal", "tok-host", `{}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantError(t, do(t, h, tt.method, tt.path, tt.token, tt.body), http.StatusBadRequest, apidto.CodeBadRequest)
		})
	}
}

func TestSummaryAndJoin(t *testing.T) {
	svc := &fakeService{}
	h := newTestRouter(t, svc)

	w := do(t, h, "GET", "/api/games/abcde", "", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if s := decode[apidto.GameSummary](t, w); s.GameID != "ABCDE" || !s.Joinable || s.MaxPlayers != 4 {
		t.Fatalf("summary = %+v", s)
	}
	wantError(t, do(t, h, "GET", "/api/games/ZZZZZ", "", ""), http.StatusNotFound, apidto.CodeGameNotFound)

	w = do(t, h, "POST", "/api/games/%20abcde%20/players", "", `{"name":"Bob"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("join status = %d %s", w.Code, w.Body.String())
	}
	if svc.lastID != "ABCDE" || svc.lastName != "Bob" {
		t.Fatalf("join got id %q name %q", svc.lastID, svc.lastName)
	}
	wantError(t, do(t, h, "POST", "/api/games/ZZZZZ/players", "", `{"name":"Bob"}`), http.StatusNotFound, apidto.CodeGameNotFound)
}

func TestAuthedRoutes(t *testing.T) {
	routes := []struct {
		method, path, body, call string
	}{
		{"GET", "/api/games/%s/view", "", "View"},
		{"POST", "/api/games/%s/start", `{"firstPlayerId":"p_guest"}`, "StartGame"},
		{"POST", "/api/games/%s/play", `{"cards":[{"suit":"hearts","rank":5}]}`, "Play"},
		{"POST", "/api/games/%s/yield", "", "Yield"},
		{"POST", "/api/games/%s/discard", `{"cards":[{"suit":"clubs","rank":10}]}`, "Discard"},
		{"POST", "/api/games/%s/next-player", `{"playerId":"p_guest"}`, "ChooseNext"},
		{"POST", "/api/games/%s/solo-jester", "", "UseSoloJester"},
		{"PUT", "/api/games/%s/signal", `{"signal":"goodPlay"}`, "SetJesterSignal"},
		{"POST", "/api/games/%s/give-up", "", "GiveUp"},
		{"POST", "/api/games/%s/rematch", "", "Rematch"},
	}
	for _, rt := range routes {
		t.Run(rt.call, func(t *testing.T) {
			svc := &fakeService{}
			h := newTestRouter(t, svc)
			path := fmt.Sprintf(rt.path, "abcde")

			wantError(t, do(t, h, rt.method, path, "", rt.body), http.StatusUnauthorized, apidto.CodeUnauthorized)
			wantError(t, do(t, h, rt.method, path, "wrong", rt.body), http.StatusUnauthorized, apidto.CodeUnauthorized)
			wantError(t, do(t, h, rt.method, fmt.Sprintf(rt.path, "ZZZZZ"), "tok-host", rt.body), http.StatusNotFound, apidto.CodeGameNotFound)

			// Non-bearer scheme is rejected.
			r := httptest.NewRequest(rt.method, path, strings.NewReader(rt.body))
			r.Header.Set("Authorization", "Basic tok-host")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			wantError(t, w, http.StatusUnauthorized, apidto.CodeUnauthorized)

			if c := svc.lastCall(); c != "" {
				t.Fatalf("service %s called despite failed auth", c)
			}

			w = do(t, h, rt.method, path, "tok-guest", rt.body)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d %s", w.Code, w.Body.String())
			}
			v := decode[apidto.PlayerView](t, w)
			if v.You != "p_guest" || v.GameID != "ABCDE" || v.Revision != 7 {
				t.Fatalf("view = %+v", v)
			}
			if svc.lastCall() != rt.call || svc.lastID != "ABCDE" {
				t.Fatalf("last call = %s on %s", svc.lastCall(), svc.lastID)
			}
		})
	}
}

func TestAuthedRouteArgs(t *testing.T) {
	svc := &fakeService{}
	h := newTestRouter(t, svc)

	do(t, h, "POST", "/api/games/ABCDE/start", "tok-host", `{"firstPlayerId":"p_guest"}`)
	if svc.lastFirst != "p_guest" {
		t.Fatalf("first = %q", svc.lastFirst)
	}
	do(t, h, "POST", "/api/games/ABCDE/next-player", "tok-host", `{"playerId":"p_guest"}`)
	if svc.lastNext != "p_guest" {
		t.Fatalf("next = %q", svc.lastNext)
	}
	// Lower-case scheme and extra whitespace are accepted; the Jester maps to the domain Jester.
	r := httptest.NewRequest("POST", "/api/games/ABCDE/play", strings.NewReader(`{"cards":[{"suit":"","rank":0}]}`))
	r.Header.Set("Authorization", "bearer   tok-host ")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK || len(svc.lastCards) != 1 || !svc.lastCards[0].IsJester() {
		t.Fatalf("play: %d cards %v", w.Code, svc.lastCards)
	}
	// An empty list is passed on (the domain answers illegal_discard).
	do(t, h, "POST", "/api/games/ABCDE/discard", "tok-host", `{"cards":[]}`)
	if svc.lastCards == nil || len(svc.lastCards) != 0 {
		t.Fatalf("empty discard cards = %#v", svc.lastCards)
	}
	do(t, h, "PUT", "/api/games/ABCDE/signal", "tok-host", `{"signal":"none"}`)
	if svc.lastSignal != domain.SignalNone || svc.lastCall() != "SetJesterSignal" {
		t.Fatalf("signal = %q", svc.lastSignal)
	}
	do(t, h, "PUT", "/api/games/ABCDE/signal", "tok-host", `{"signal":"ratherNot"}`)
	if svc.lastSignal != domain.SignalRatherNot {
		t.Fatalf("signal = %q", svc.lastSignal)
	}

	w = do(t, h, "POST", "/api/games/ABCDE/rematch", "tok-host", "")
	if v := decode[apidto.PlayerView](t, w); v.NextGameID != "NEXT2" {
		t.Fatalf("rematch view nextGameId = %q", v.NextGameID)
	}
}

func TestLeave204(t *testing.T) {
	svc := &fakeService{}
	h := newTestRouter(t, svc)
	wantError(t, do(t, h, "POST", "/api/games/ABCDE/leave", "", ""), http.StatusUnauthorized, apidto.CodeUnauthorized)
	w := do(t, h, "POST", "/api/games/abcde/leave", "tok-guest", "")
	if w.Code != http.StatusNoContent || w.Body.Len() != 0 {
		t.Fatalf("leave = %d %q", w.Code, w.Body.String())
	}
	if svc.lastCall() != "LeaveGame" || svc.lastID != "ABCDE" {
		t.Fatalf("last call = %s on %s", svc.lastCall(), svc.lastID)
	}
	svc.err = domain.ErrCannotLeave
	wantError(t, do(t, h, "POST", "/api/games/ABCDE/leave", "tok-host", ""), http.StatusConflict, apidto.CodeCannotLeave)
}

func TestStartEmptyBodyIsRandom(t *testing.T) {
	svc := &fakeService{lastFirst: "stale"}
	h := newTestRouter(t, svc)
	w := do(t, h, "POST", "/api/games/ABCDE/start", "tok-host", "")
	if w.Code != http.StatusOK || svc.lastFirst != "" || svc.lastCall() != "StartGame" {
		t.Fatalf("start = %d first %q", w.Code, svc.lastFirst)
	}
	svc.lastFirst = "stale"
	w = do(t, h, "POST", "/api/games/ABCDE/start", "tok-host", `{}`)
	if w.Code != http.StatusOK || svc.lastFirst != "" {
		t.Fatalf("start {} = %d first %q", w.Code, svc.lastFirst)
	}
}

func TestErrorTable(t *testing.T) {
	tests := []struct {
		err    error
		status int
		code   string
	}{
		{ports.ErrUnauthorized, 401, apidto.CodeUnauthorized},
		{domain.ErrNotHost, 403, apidto.CodeNotHost},
		{ports.ErrGameNotFound, 404, apidto.CodeGameNotFound},
		{domain.ErrWrongPhase, 409, apidto.CodeWrongPhase},
		{domain.ErrWrongStep, 409, apidto.CodeWrongStep},
		{domain.ErrGameFull, 409, apidto.CodeGameFull},
		{domain.ErrNameTaken, 409, apidto.CodeNameTaken},
		{domain.ErrNotYourTurn, 409, apidto.CodeNotYourTurn},
		{domain.ErrRematchExists, 409, apidto.CodeRematchExists},
		{domain.ErrCannotLeave, 409, apidto.CodeCannotLeave},
		{ports.ErrVersionConflict, 409, apidto.CodeConflict},
		{domain.ErrInvalidName, 422, apidto.CodeInvalidName},
		{domain.ErrUnknownPlayer, 422, apidto.CodeUnknownPlayer},
		{domain.ErrCardNotInHand, 422, apidto.CodeCardNotInHand},
		{fmt.Errorf("%w: mixedRanks", domain.ErrIllegalPlay), 422, apidto.CodeIllegalPlay},
		{domain.ErrIllegalDiscard, 422, apidto.CodeIllegalDiscard},
		{domain.ErrYieldNotAllowed, 422, apidto.CodeYieldNotAllowed},
		{domain.ErrNotSolo, 422, apidto.CodeNotSolo},
		{domain.ErrNoSoloJester, 422, apidto.CodeNoSoloJester},
		{domain.ErrIllegalSignal, 422, apidto.CodeIllegalSignal},
		{domain.ErrDuplicatePlayer, 500, apidto.CodeInternal},
		{errors.New("disk on fire"), 500, apidto.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.code+"/"+tt.err.Error(), func(t *testing.T) {
			var logBuf bytes.Buffer
			svc := &fakeService{err: tt.err}
			h := NewRouter(Deps{
				Logger: slog.New(slog.NewTextHandler(&logBuf, nil)),
				Health: func(context.Context) error { return nil },
				Games:  svc,
			})
			w := do(t, h, "POST", "/api/games/ABCDE/play", "tok-host", `{"cards":[{"suit":"hearts","rank":5}]}`)
			wantError(t, w, tt.status, tt.code)
			e := decode[apidto.ErrorResponse](t, w)
			if tt.status == 500 {
				if e.Error.Message != apidto.InternalMessage {
					t.Fatalf("500 leaks message %q", e.Error.Message)
				}
				if !strings.Contains(logBuf.String(), tt.err.Error()) {
					t.Fatalf("500 not logged: %q", logBuf.String())
				}
			}
		})
	}
}

func TestAPINotFoundBeatsSPA(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>spa</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := NewRouter(Deps{
		Health:    func(context.Context) error { return nil },
		StaticDir: dir,
		Games:     &fakeService{},
	})
	for _, p := range []string{"/api/nope", "/api/games/ABCDE/unknown", "/api/"} {
		wantError(t, do(t, h, "GET", p, "", ""), http.StatusNotFound, apidto.CodeNotFound)
	}
	w := do(t, h, "GET", "/g/ABCDE", "", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "spa") {
		t.Fatalf("SPA fallback = %d %s", w.Code, w.Body.String())
	}
	// Without the Games service the routes are absent but still JSON 404.
	h = NewRouter(Deps{Health: func(context.Context) error { return nil }})
	wantError(t, do(t, h, "POST", "/api/games", "", `{}`), http.StatusNotFound, apidto.CodeNotFound)
}

func TestWSMountAndLogging(t *testing.T) {
	var logBuf bytes.Buffer
	var gotCode string
	wsH := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCode = r.PathValue("code")
		w.WriteHeader(http.StatusTeapot)
	})
	h := NewRouter(Deps{
		Logger: slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug})),
		Health: func(context.Context) error { return nil },
		Games:  &fakeService{},
		WS:     wsH,
	})
	w := do(t, h, "GET", "/ws/games/abcde?token=supersecret", "", "")
	if w.Code != http.StatusTeapot || gotCode != "abcde" {
		t.Fatalf("ws mount: %d code %q", w.Code, gotCode)
	}
	do(t, h, "GET", "/api/games/ABCDE/view?token=supersecret", "tok-host", "")
	if strings.Contains(logBuf.String(), "supersecret") {
		t.Fatalf("query logged: %s", logBuf.String())
	}
	if !strings.Contains(logBuf.String(), "/ws/games/abcde") {
		t.Fatalf("path not logged: %s", logBuf.String())
	}
}
