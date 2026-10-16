package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/martin-pouls/shared/backend/idgen"

	"github.com/martin-pouls/regicide/backend/internal/adapters/memory"
	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/ports"
)

type noShuffle struct{}

func (noShuffle) Shuffle(int, func(i, j int)) {}

type published struct {
	GameID   domain.GameID
	Revision int64
	Events   []domain.Event
}

type recorder struct {
	mu   sync.Mutex
	list []published
}

func (r *recorder) Publish(_ context.Context, u ports.GameUpdate) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.list = append(r.list, published{u.Game.ID, u.Game.Revision, u.Events})
}

func (r *recorder) all() []published {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]published(nil), r.list...)
}

type fixture struct {
	svc  *Service
	repo *memory.GameRepository
	rec  *recorder
}

func newFixture(t *testing.T, mod func(*Config)) *fixture {
	t.Helper()
	var ids, toks, gids atomic.Int64
	cfg := Config{
		NewGameID:   func() domain.GameID { return domain.GameID(fmt.Sprintf("GAME%d", gids.Add(1))) },
		NewPlayerID: func() domain.PlayerID { return domain.PlayerID(fmt.Sprintf("p%d", ids.Add(1))) },
		NewToken:    func() string { return fmt.Sprintf("tok%d", toks.Add(1)) },
		NewRNG:      func() domain.RNG { return noShuffle{} },
	}
	if mod != nil {
		mod(&cfg)
	}
	repo := memory.NewGameRepository()
	rec := &recorder{}
	return &fixture{svc: New(repo, rec, cfg), repo: repo, rec: rec}
}

// started creates a game with n players (host first) and starts it with the host to begin.
func (f *fixture) started(t *testing.T, n int) (domain.GameID, []ports.Session) {
	t.Helper()
	ctx := context.Background()
	host, err := f.svc.CreateGame(ctx, "P1")
	if err != nil {
		t.Fatal(err)
	}
	sessions := []ports.Session{host}
	for i := 2; i <= n; i++ {
		s, err := f.svc.JoinGame(ctx, host.GameID, fmt.Sprintf("P%d", i))
		if err != nil {
			t.Fatal(err)
		}
		sessions = append(sessions, s)
	}
	if err := f.svc.StartGame(ctx, host.GameID, host.PlayerID, ""); err != nil {
		t.Fatal(err)
	}
	return host.GameID, sessions
}

func eventTypes(evs []domain.Event) []string {
	out := make([]string, len(evs))
	for i, e := range evs {
		out[i] = reflect.TypeOf(e).Name()
	}
	return out
}

func TestLifecycle(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	host, err := f.svc.CreateGame(ctx, "Ann")
	if err != nil {
		t.Fatal(err)
	}
	if host.Token == "" || host.View.Phase != domain.PhaseLobby || host.PlayerID == "" {
		t.Fatalf("bad host session: %+v", host)
	}
	bob, err := f.svc.JoinGame(ctx, " game1 ", "Bob") // code is normalised
	if err != nil {
		t.Fatal(err)
	}
	if bob.GameID != host.GameID || bob.PlayerID == host.PlayerID {
		t.Fatalf("bad join: %+v", bob)
	}
	sum, err := f.svc.Summary(ctx, host.GameID)
	if err != nil || sum.PlayerCount != 2 {
		t.Fatalf("summary: %+v %v", sum, err)
	}

	if err := f.svc.StartGame(ctx, host.GameID, host.PlayerID, bob.PlayerID); err != nil {
		t.Fatal(err)
	}
	v, err := f.svc.View(ctx, host.GameID, host.PlayerID)
	if err != nil || v.Phase != domain.PhasePlaying || v.Turn == nil || v.Turn.ActivePlayerID != bob.PlayerID {
		t.Fatalf("start: %+v %v", v, err)
	}
	if len(v.Hand) != 7 || v.Enemy == nil || v.Enemy.Card != domain.NewCard(domain.SuitSpades, domain.RankJack) {
		t.Fatalf("unexpected start view: hand %v enemy %+v", v.Hand, v.Enemy)
	}

	// Bob is active: he plays a single card (2 of spades, immune), then must discard.
	bv, _ := f.svc.View(ctx, host.GameID, bob.PlayerID)
	if len(bv.Hand) != 7 {
		t.Fatalf("bob hand = %v", bv.Hand)
	}
	if err := f.svc.Play(ctx, host.GameID, bob.PlayerID, bv.Hand[:1]); err != nil {
		t.Fatal(err)
	}
	bv, _ = f.svc.View(ctx, host.GameID, bob.PlayerID)
	if bv.Turn.Step != domain.StepDiscard {
		t.Fatalf("step = %s, want discard", bv.Turn.Step)
	}
	hv, _ := f.svc.View(ctx, host.GameID, host.PlayerID)
	if hv.Enemy == nil || len(hv.Enemy.Plays) != 1 || len(hv.Hand) != 7 {
		t.Fatalf("host view after bob's play: %+v", hv)
	}
}

func TestStartFirstPlayer(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	host, _ := f.svc.CreateGame(ctx, "Ann")
	bob, _ := f.svc.JoinGame(ctx, host.GameID, "Bob")

	if err := f.svc.StartGame(ctx, host.GameID, bob.PlayerID, ""); !errors.Is(err, domain.ErrNotHost) {
		t.Fatalf("non-host start: %v", err)
	}
	if err := f.svc.StartGame(ctx, host.GameID, host.PlayerID, "ghost"); !errors.Is(err, domain.ErrUnknownPlayer) {
		t.Fatalf("unknown first player: %v", err)
	}
	// Empty first player: random, which noShuffle turns into seat 0.
	if err := f.svc.StartGame(ctx, host.GameID, host.PlayerID, ""); err != nil {
		t.Fatal(err)
	}
	v, err := f.svc.View(ctx, host.GameID, host.PlayerID)
	if err != nil || v.Turn.ActivePlayerID != host.PlayerID {
		t.Fatalf("random start: %+v %v", v.Turn, err)
	}
	f2 := newFixture(t, nil)
	h2, _ := f2.svc.CreateGame(ctx, "Ann")
	b2, _ := f2.svc.JoinGame(ctx, h2.GameID, "Bob")
	if err := f2.svc.StartGame(ctx, h2.GameID, h2.PlayerID, b2.PlayerID); err != nil {
		t.Fatal(err)
	}
	v, err = f2.svc.View(ctx, h2.GameID, h2.PlayerID)
	if err != nil || v.Turn.ActivePlayerID != b2.PlayerID {
		t.Fatalf("chosen start: %+v %v", v.Turn, err)
	}
}

func TestLeaveGame(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	host, _ := f.svc.CreateGame(ctx, "Ann")
	bob, _ := f.svc.JoinGame(ctx, host.GameID, "Bob")
	before := len(f.rec.all())

	if err := f.svc.LeaveGame(ctx, host.GameID, bob.PlayerID); err != nil {
		t.Fatal(err)
	}
	pubs := f.rec.all()[before:]
	if len(pubs) != 1 || !reflect.DeepEqual(eventTypes(pubs[0].Events), []string{"PlayerLeft"}) {
		t.Fatalf("unexpected publishes: %+v", pubs)
	}
	sum, _ := f.svc.Summary(ctx, host.GameID)
	if sum.PlayerCount != 1 {
		t.Fatalf("players = %d", sum.PlayerCount)
	}
	if _, err := f.svc.Authenticate(ctx, host.GameID, bob.Token); !errors.Is(err, ports.ErrUnauthorized) {
		t.Fatalf("left player still authenticates: %v", err)
	}
	if _, err := f.svc.Authenticate(ctx, host.GameID, host.Token); err != nil {
		t.Fatalf("host: %v", err)
	}
	if err := f.svc.LeaveGame(ctx, host.GameID, host.PlayerID); !errors.Is(err, domain.ErrCannotLeave) {
		t.Fatalf("sole player leave: %v", err)
	}
	if err := f.svc.LeaveGame(ctx, host.GameID, bob.PlayerID); !errors.Is(err, domain.ErrUnknownPlayer) {
		t.Fatalf("leave twice: %v", err)
	}
	// Rejoining works with a fresh seat.
	if _, err := f.svc.JoinGame(ctx, host.GameID, "Bob"); err != nil {
		t.Fatalf("rejoin: %v", err)
	}
	// After the start nobody can leave.
	if err := f.svc.StartGame(ctx, host.GameID, host.PlayerID, ""); err != nil {
		t.Fatal(err)
	}
	if err := f.svc.LeaveGame(ctx, host.GameID, host.PlayerID); !errors.Is(err, domain.ErrWrongPhase) {
		t.Fatalf("leave after start: %v", err)
	}
}

// countingRNG counts how often the service asks for a fresh RNG.
func TestPlayUsesFreshRNG(t *testing.T) {
	ctx := context.Background()
	var n atomic.Int64
	f := newFixture(t, func(c *Config) {
		c.NewRNG = func() domain.RNG { n.Add(1); return noShuffle{} }
	})
	id, sessions := f.started(t, 2)
	if n.Load() != 1 {
		t.Fatalf("RNG created %d times after start, want 1", n.Load())
	}
	v, _ := f.svc.View(ctx, id, sessions[0].PlayerID)
	if err := f.svc.Play(ctx, id, sessions[0].PlayerID, v.Hand[:1]); err != nil {
		t.Fatal(err)
	}
	if n.Load() != 2 {
		t.Fatalf("RNG created %d times after play, want 2", n.Load())
	}
	// Non-RNG commands and failed plays do not matter for correctness, but yield must not need one.
	before := n.Load()
	if err := f.svc.GiveUp(ctx, id, sessions[0].PlayerID); err != nil {
		t.Fatal(err)
	}
	if n.Load() != before {
		t.Fatalf("GiveUp created an RNG")
	}
}

func TestAuthenticate(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	host, _ := f.svc.CreateGame(ctx, "Ann")
	bob, _ := f.svc.JoinGame(ctx, host.GameID, "Bob")

	for _, s := range []ports.Session{host, bob} {
		got, err := f.svc.Authenticate(ctx, host.GameID, s.Token)
		if err != nil || got != s.PlayerID {
			t.Errorf("auth %s: got %q err %v", s.PlayerID, got, err)
		}
	}
	if _, err := f.svc.Authenticate(ctx, host.GameID, "nope"); !errors.Is(err, ports.ErrUnauthorized) {
		t.Errorf("bad token: %v", err)
	}
	if _, err := f.svc.Authenticate(ctx, host.GameID, ""); !errors.Is(err, ports.ErrUnauthorized) {
		t.Errorf("empty token: %v", err)
	}
	if _, err := f.svc.Authenticate(ctx, "ZZZZZ", host.Token); !errors.Is(err, ports.ErrGameNotFound) {
		t.Errorf("unknown game: %v", err)
	}
	// Only hashes are stored.
	rec, _ := f.repo.Load(ctx, host.GameID)
	if rec.TokenHashes[host.PlayerID] != idgen.HashToken(host.Token) || rec.TokenHashes[host.PlayerID] == host.Token {
		t.Errorf("token not stored hashed")
	}
}

func TestDefaultGenerators(t *testing.T) {
	cfg := New(nil, nil, Config{}).cfg
	for range 200 {
		id := cfg.NewGameID()
		if len(id) != 5 {
			t.Fatalf("bad id %q", id)
		}
		for _, c := range string(id) {
			if !containsRune(idgen.CodeAlphabet, c) {
				t.Fatalf("bad char %q in %q", c, id)
			}
		}
	}
	pid := cfg.NewPlayerID()
	if len(pid) != 12 || pid[:2] != "p_" {
		t.Errorf("bad player id %q", pid)
	}
	if a, b := cfg.NewToken(), cfg.NewToken(); a == b || len(a) != 43 {
		t.Errorf("bad tokens %q %q", a, b)
	}
	if cfg.NewRNG() == nil {
		t.Error("nil rng")
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

func TestCreateGameIDCollisionRetry(t *testing.T) {
	ctx := context.Background()
	var n int
	f := newFixture(t, func(c *Config) {
		c.NewGameID = func() domain.GameID {
			n++
			if n <= 2 {
				return "TAKEN"
			}
			return domain.GameID(fmt.Sprintf("NEW%d", n))
		}
	})
	if _, err := f.svc.CreateGame(ctx, "Ann"); err != nil {
		t.Fatal(err)
	}
	s, err := f.svc.CreateGame(ctx, "Bob")
	if err != nil || s.GameID != "NEW3" {
		t.Fatalf("expected retry to NEW3, got %v %v", s.GameID, err)
	}

	// Exhausted attempts.
	f2 := newFixture(t, func(c *Config) { c.NewGameID = func() domain.GameID { return "SAME" } })
	if _, err := f2.svc.CreateGame(ctx, "Ann"); err != nil {
		t.Fatal(err)
	}
	if _, err := f2.svc.CreateGame(ctx, "Bob"); !errors.Is(err, ports.ErrDuplicateGameID) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestCreateGameValidationError(t *testing.T) {
	f := newFixture(t, nil)
	if _, err := f.svc.CreateGame(context.Background(), "  "); !errors.Is(err, domain.ErrInvalidName) {
		t.Fatalf("got %v", err)
	}
	if got := f.rec.all(); len(got) != 0 {
		t.Fatalf("create must not publish: %v", got)
	}
}

func TestEventsPublishedOncePerCommandInOrder(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	host, _ := f.svc.CreateGame(ctx, "Ann")
	f.svc.JoinGame(ctx, host.GameID, "Bob")
	f.svc.StartGame(ctx, host.GameID, host.PlayerID, "")

	got := f.rec.all()
	want := [][]string{
		{"PlayerJoined"},
		{"GameStarted", "EnemyRevealed", "TurnStarted"},
	}
	if len(got) != len(want) {
		t.Fatalf("published %d updates, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if !reflect.DeepEqual(eventTypes(got[i].Events), want[i]) {
			t.Errorf("update %d: events %v, want %v", i, eventTypes(got[i].Events), want[i])
		}
		if i > 0 && got[i].Revision != got[i-1].Revision+1 {
			t.Errorf("update %d: revision %d after %d", i, got[i].Revision, got[i-1].Revision)
		}
	}

	// A play publishes exactly one more update.
	v, _ := f.svc.View(ctx, host.GameID, host.PlayerID)
	if err := f.svc.Play(ctx, host.GameID, host.PlayerID, v.Hand[:1]); err != nil {
		t.Fatal(err)
	}
	got = f.rec.all()
	if len(got) != 3 || eventTypes(got[2].Events)[0] != "CardsPlayed" {
		t.Fatalf("unexpected publishes after play: %+v", got)
	}
}

func TestDomainErrorsNotSavedNorPublished(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	host, _ := f.svc.CreateGame(ctx, "Ann")
	bob, _ := f.svc.JoinGame(ctx, host.GameID, "Bob")
	before := f.rec.all()
	recBefore, _ := f.repo.Load(ctx, host.GameID)

	checks := []struct {
		name string
		err  error
		want error
	}{
		{"non-host start", f.svc.StartGame(ctx, host.GameID, bob.PlayerID, ""), domain.ErrNotHost},
		{"play in lobby", f.svc.Play(ctx, host.GameID, host.PlayerID, []domain.Card{domain.NewCard(domain.SuitHearts, 2)}), domain.ErrWrongPhase},
		{"yield in lobby", f.svc.Yield(ctx, host.GameID, host.PlayerID), domain.ErrWrongPhase},
		{"duplicate name", errOf(f.svc.JoinGame(ctx, host.GameID, "ann")), domain.ErrNameTaken},
		{"rematch in lobby", f.svc.Rematch(ctx, host.GameID, host.PlayerID), domain.ErrWrongPhase},
		{"unknown game", f.svc.StartGame(ctx, "NOPE1", host.PlayerID, ""), ports.ErrGameNotFound},
	}
	for _, c := range checks {
		if !errors.Is(c.err, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, c.err, c.want)
		}
	}
	if after := f.rec.all(); len(after) != len(before) {
		t.Errorf("published on error: %d -> %d", len(before), len(after))
	}
	recAfter, _ := f.repo.Load(ctx, host.GameID)
	if recAfter.Version != recBefore.Version || recAfter.Game.Revision != recBefore.Game.Revision {
		t.Errorf("state saved on error")
	}
	if f.svc.locks.Len() != 0 {
		t.Errorf("lock entries leaked: %d", f.svc.locks.Len())
	}
}

func errOf(_ ports.Session, err error) error { return err }

func TestConcurrentYieldSerialised(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	id, sessions := f.started(t, 2)
	host := sessions[0]
	base := len(f.rec.all())

	const workers = 8
	var wg sync.WaitGroup
	var ok atomic.Int64
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := f.svc.Yield(ctx, id, host.PlayerID); err == nil {
				ok.Add(1)
			}
		}()
	}
	wg.Wait()
	if ok.Load() != 1 {
		t.Fatalf("expected exactly one successful yield, got %d", ok.Load())
	}
	if got := len(f.rec.all()) - base; got != 1 {
		t.Fatalf("expected one publish, got %d", got)
	}
	if f.svc.locks.Len() != 0 {
		t.Errorf("lock entries leaked")
	}
}

// conflictRepo wraps a repository and fails the first n Updates with ErrVersionConflict.
type conflictRepo struct {
	ports.GameRepository
	fail int
}

func (c *conflictRepo) Update(ctx context.Context, rec *ports.GameRecord) error {
	if c.fail > 0 {
		c.fail--
		return ports.ErrVersionConflict
	}
	return c.GameRepository.Update(ctx, rec)
}

func TestOCCRetry(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		fail    int
		wantErr error
		pubs    int
	}{{2, nil, 1}, {3, ports.ErrVersionConflict, 0}} {
		repo := &conflictRepo{GameRepository: memory.NewGameRepository()}
		rec := &recorder{}
		svc := New(repo, rec, Config{})
		host, err := svc.CreateGame(ctx, "Ann")
		if err != nil {
			t.Fatal(err)
		}
		repo.fail = tc.fail
		_, err = svc.JoinGame(ctx, host.GameID, "Bob")
		if !errors.Is(err, tc.wantErr) {
			t.Errorf("fail=%d: got %v, want %v", tc.fail, err, tc.wantErr)
		}
		if got := len(rec.all()); got != tc.pubs {
			t.Errorf("fail=%d: %d publishes, want %d", tc.fail, got, tc.pubs)
		}
	}
}

// finish ends the game by giving up (any seated player may).
func finish(t *testing.T, f *fixture, id domain.GameID, by domain.PlayerID) {
	t.Helper()
	if err := f.svc.GiveUp(context.Background(), id, by); err != nil {
		t.Fatal(err)
	}
}

func TestRematch(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	id, sessions := f.started(t, 2)
	host, bob := sessions[0], sessions[1]
	finish(t, f, id, host.PlayerID)
	before := len(f.rec.all())

	if err := f.svc.Rematch(ctx, id, bob.PlayerID); err != nil {
		t.Fatal(err)
	}
	v, err := f.svc.View(ctx, id, bob.PlayerID)
	if err != nil {
		t.Fatal(err)
	}
	if v.NextGameID == "" || v.NextGameID == id {
		t.Fatalf("NextGameID not set: %q", v.NextGameID)
	}
	pubs := f.rec.all()[before:]
	if len(pubs) != 1 || pubs[0].GameID != id || !reflect.DeepEqual(eventTypes(pubs[0].Events), []string{"RematchCreated"}) {
		t.Fatalf("unexpected publishes: %+v", pubs)
	}

	// Old tokens work in the new game; the caller is host; same players.
	for _, s := range sessions {
		got, err := f.svc.Authenticate(ctx, v.NextGameID, s.Token)
		if err != nil || got != s.PlayerID {
			t.Errorf("token for %s in new game: %q %v", s.PlayerID, got, err)
		}
	}
	nv, err := f.svc.View(ctx, v.NextGameID, host.PlayerID)
	if err != nil {
		t.Fatal(err)
	}
	if nv.Phase != domain.PhaseLobby || nv.HostID != bob.PlayerID || len(nv.Players) != 2 {
		t.Errorf("bad new game view: phase=%s host=%s players=%d", nv.Phase, nv.HostID, len(nv.Players))
	}

	// Idempotent: second call (by anyone) changes nothing and publishes nothing.
	recBefore, _ := f.repo.Load(ctx, id)
	if err := f.svc.Rematch(ctx, id, host.PlayerID); err != nil {
		t.Fatalf("idempotent rematch: %v", err)
	}
	recAfter, _ := f.repo.Load(ctx, id)
	if recAfter.Version != recBefore.Version || len(f.rec.all()) != before+1 {
		t.Errorf("idempotent rematch changed state or published")
	}
}

func TestRematchIDCollisionRetry(t *testing.T) {
	ctx := context.Background()
	var calls int
	var f *fixture
	f = newFixture(t, func(c *Config) {
		base := c.NewGameID
		c.NewGameID = func() domain.GameID {
			calls++
			if calls > 3 && calls <= 5 { // after game creation: first rematch candidates collide
				return "GAME1"
			}
			return base()
		}
	})
	id, sessions := f.started(t, 1)
	finish(t, f, id, sessions[0].PlayerID)
	calls = 3
	if err := f.svc.Rematch(ctx, id, sessions[0].PlayerID); err != nil {
		t.Fatal(err)
	}
	v, err := f.svc.View(ctx, id, sessions[0].PlayerID)
	if err != nil {
		t.Fatal(err)
	}
	if v.NextGameID == "" || v.NextGameID == id {
		t.Fatalf("NextGameID = %q", v.NextGameID)
	}
	if _, err := f.svc.View(ctx, v.NextGameID, sessions[0].PlayerID); err != nil {
		t.Fatal(err)
	}
}

func TestRematchErrors(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	id, sessions := f.started(t, 2)
	if err := f.svc.Rematch(ctx, id, sessions[0].PlayerID); !errors.Is(err, domain.ErrWrongPhase) {
		t.Errorf("unfinished: %v", err)
	}
	finish(t, f, id, sessions[0].PlayerID)
	if err := f.svc.Rematch(ctx, id, "ghost"); !errors.Is(err, domain.ErrUnknownPlayer) {
		t.Errorf("unknown player: %v", err)
	}
	if err := f.svc.Rematch(ctx, "NOPE1", sessions[0].PlayerID); !errors.Is(err, ports.ErrGameNotFound) {
		t.Errorf("unknown game: %v", err)
	}
}

func TestPurgeInactive(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	clock := now.Add(-48 * time.Hour)
	repo := memory.NewGameRepository(memory.WithNow(func() time.Time { return clock }))
	svc := New(repo, nil, Config{})
	old, _ := svc.CreateGame(ctx, "Ann")
	clock = now
	fresh, _ := svc.CreateGame(ctx, "Bob")

	n, err := svc.PurgeInactive(ctx, 24*time.Hour)
	if err != nil || n != 1 {
		t.Fatalf("purged %d, %v", n, err)
	}
	if _, err := svc.Summary(ctx, old.GameID); !errors.Is(err, ports.ErrGameNotFound) {
		t.Errorf("old game still there: %v", err)
	}
	if _, err := svc.Summary(ctx, fresh.GameID); err != nil {
		t.Errorf("fresh game gone: %v", err)
	}
	if n, _ := svc.PurgeInactive(ctx, 24*time.Hour); n != 0 {
		t.Errorf("second purge removed %d", n)
	}
}
