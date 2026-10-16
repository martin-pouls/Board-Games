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

	"github.com/martin-pouls/the-game/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-game/backend/internal/core/domain"
	"github.com/martin-pouls/the-game/backend/internal/core/ports"
)

// fakeService is a canned ports.GameService. Game "ABCDE" exists; token
// "tok-host" is p_host and "tok-guest" is p_guest. err, when set, is
// returned by every call except Authenticate.
type fakeService struct {
	mu    sync.Mutex
	err   error
	calls []string

	lastName string
	lastOpts domain.Options
	lastID   domain.GameID
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
		Options:  domain.DefaultOptions(),
		Players:  []domain.PlayerSummary{{ID: p, Name: "Ann", IsYou: true}},
	}
}

func (f *fakeService) session(id domain.GameID, name string) (ports.Session, error) {
	f.mu.Lock()
	f.lastName = name
	f.mu.Unlock()
	return ports.Session{GameID: id, PlayerID: "p_new", Token: "tok-new", View: fakeView(id, "p_new")}, nil
}

func (f *fakeService) CreateGame(_ context.Context, name string, opts domain.Options) (ports.Session, error) {
	if err := f.record("CreateGame", testGame); err != nil {
		return ports.Session{}, err
	}
	f.mu.Lock()
	f.lastOpts = opts
	f.mu.Unlock()
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
	return domain.GameSummary{GameID: id, Phase: domain.PhaseLobby, HostName: "Ann", PlayerCount: 1, MaxPlayers: 5, Joinable: true}, nil
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

// The gameplay calls are WebSocket-only; the REST tests never reach them.

func (f *fakeService) SetOptions(context.Context, domain.GameID, domain.PlayerID, domain.Options) error {
	return f.record("SetOptions", testGame)
}

func (f *fakeService) StartGame(context.Context, domain.GameID, domain.PlayerID) error {
	return f.record("StartGame", testGame)
}

func (f *fakeService) ChooseFirstPlayer(context.Context, domain.GameID, domain.PlayerID, domain.PlayerID) error {
	return f.record("ChooseFirstPlayer", testGame)
}

func (f *fakeService) PlayCard(context.Context, domain.GameID, domain.PlayerID, int, int) error {
	return f.record("PlayCard", testGame)
}

func (f *fakeService) UndoPlay(context.Context, domain.GameID, domain.PlayerID) error {
	return f.record("UndoPlay", testGame)
}

func (f *fakeService) EndTurn(context.Context, domain.GameID, domain.PlayerID) error {
	return f.record("EndTurn", testGame)
}

func (f *fakeService) GiveUp(context.Context, domain.GameID, domain.PlayerID) error {
	return f.record("GiveUp", testGame)
}

func (f *fakeService) EndGame(context.Context, domain.GameID, domain.PlayerID) error {
	return f.record("EndGame", testGame)
}

func (f *fakeService) Rematch(context.Context, domain.GameID, domain.PlayerID) error {
	return f.record("Rematch", testGame)
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
	if svc.lastName != "Ann" || svc.lastOpts != domain.DefaultOptions() {
		t.Fatalf("service got name %q opts %+v", svc.lastName, svc.lastOpts)
	}
	// Slices are never null on the wire.
	if !strings.Contains(w.Body.String(), `"hand":[]`) {
		t.Fatalf("hand should be [] in %s", w.Body.String())
	}

	w = do(t, h, "POST", "/api/games", "", `{"name":"Ann","options":{"variant":"","pro":true,"reducedHandSize":true}}`) // empty variant = default (extreme)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d", w.Code)
	}
	want := domain.Options{Variant: domain.VariantExtreme, Pro: true, ReducedHandSize: true}
	if svc.lastOpts != want {
		t.Fatalf("opts = %+v, want %+v", svc.lastOpts, want)
	}

	w = do(t, h, "POST", "/api/games", "", `{"name":"Ann","options":{"variant":"extreme","pro":true}}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d", w.Code)
	}
	want = domain.Options{Variant: domain.VariantExtreme, Pro: true}
	if svc.lastOpts != want {
		t.Fatalf("opts = %+v, want %+v", svc.lastOpts, want)
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
	if s := decode[apidto.GameSummary](t, w); s.GameID != "ABCDE" || !s.Joinable || s.MaxPlayers != 5 {
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

func TestView(t *testing.T) {
	svc := &fakeService{}
	h := newTestRouter(t, svc)
	path := "/api/games/abcde/view"

	wantError(t, do(t, h, "GET", path, "", ""), http.StatusUnauthorized, apidto.CodeUnauthorized)
	wantError(t, do(t, h, "GET", path, "wrong", ""), http.StatusUnauthorized, apidto.CodeUnauthorized)
	wantError(t, do(t, h, "GET", "/api/games/ZZZZZ/view", "tok-host", ""), http.StatusNotFound, apidto.CodeGameNotFound)

	// Non-bearer scheme is rejected.
	r := httptest.NewRequest("GET", path, nil)
	r.Header.Set("Authorization", "Basic tok-host")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	wantError(t, w, http.StatusUnauthorized, apidto.CodeUnauthorized)

	if c := svc.lastCall(); c != "" {
		t.Fatalf("service %s called despite failed auth", c)
	}

	w = do(t, h, "GET", path, "tok-guest", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d %s", w.Code, w.Body.String())
	}
	v := decode[apidto.PlayerView](t, w)
	if v.You != "p_guest" || v.GameID != "ABCDE" || v.Revision != 7 {
		t.Fatalf("view = %+v", v)
	}
	if svc.lastCall() != "View" || svc.lastID != "ABCDE" {
		t.Fatalf("last call = %s on %s", svc.lastCall(), svc.lastID)
	}

	// Lower-case scheme and extra whitespace are accepted.
	r = httptest.NewRequest("GET", path, nil)
	r.Header.Set("Authorization", "bearer   tok-host ")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("lower-case bearer: %d", w.Code)
	}
}

func TestLeave204(t *testing.T) {
	svc := &fakeService{}
	h := newTestRouter(t, svc)
	wantError(t, do(t, h, "POST", "/api/games/ABCDE/leave", "", ""), http.StatusUnauthorized, apidto.CodeUnauthorized)
	if c := svc.lastCall(); c != "" {
		t.Fatalf("service %s called despite failed auth", c)
	}
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

// Gameplay is WebSocket-only: its former REST routes do not exist.
func TestGameplayRoutesAreGone(t *testing.T) {
	svc := &fakeService{}
	h := newTestRouter(t, svc)
	for _, rt := range []struct{ method, path string }{
		{"PUT", "options"}, {"POST", "start"}, {"POST", "first-player"}, {"POST", "plays"}, {"POST", "undo"},
		{"POST", "end-turn"}, {"POST", "give-up"}, {"POST", "end-game"}, {"POST", "rematch"},
	} {
		t.Run(rt.path, func(t *testing.T) {
			wantError(t, do(t, h, rt.method, "/api/games/ABCDE/"+rt.path, "tok-host", `{}`), http.StatusNotFound, apidto.CodeNotFound)
		})
	}
	if c := svc.lastCall(); c != "" {
		t.Fatalf("service %s called", c)
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
		{domain.ErrGameFull, 409, apidto.CodeGameFull},
		{domain.ErrNameTaken, 409, apidto.CodeNameTaken},
		{domain.ErrNotYourTurn, 409, apidto.CodeNotYourTurn},
		{domain.ErrRematchExists, 409, apidto.CodeRematchExists},
		{domain.ErrCannotLeave, 409, apidto.CodeCannotLeave},
		{domain.ErrNothingToUndo, 409, apidto.CodeNothingToUndo},
		{ports.ErrVersionConflict, 409, apidto.CodeConflict},
		{domain.ErrInvalidName, 422, apidto.CodeInvalidName},
		{domain.ErrInvalidOptions, 422, apidto.CodeInvalidOptions},
		{domain.ErrUnknownPlayer, 422, apidto.CodeUnknownPlayer},
		{domain.ErrCardNotInHand, 422, apidto.CodeCardNotInHand},
		{domain.ErrInvalidPile, 422, apidto.CodeInvalidPile},
		{fmt.Errorf("%w: 42 on pile 1", domain.ErrIllegalPlacement), 422, apidto.CodeIllegalPlacement},
		{domain.ErrMinimumNotReached, 422, apidto.CodeMinimumNotReached},
		{domain.ErrBackwardsTrickBlocked, 422, apidto.CodeBackwardsTrickBlocked},
		{domain.ErrSinglePileViolation, 422, apidto.CodeSinglePileViolation},
		{domain.ErrExactCountExceeded, 422, apidto.CodeExactCountExceeded},
		{domain.ErrExactCountUnreachable, 422, apidto.CodeExactCountUnreachable},
		{domain.ErrExactCountRequired, 422, apidto.CodeExactCountRequired},
		{domain.ErrStopTooEarly, 422, apidto.CodeStopTooEarly},
		{fmt.Errorf("STOP would end the turn: %w", domain.ErrSkullUncovered), 422, apidto.CodeSkullUncovered},
		{domain.ErrSkullUncoverable, 422, apidto.CodeSkullUncoverable},
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
			w := do(t, h, "GET", "/api/games/ABCDE/view", "tok-host", "")
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
