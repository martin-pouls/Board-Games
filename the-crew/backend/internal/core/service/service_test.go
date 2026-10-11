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

	"github.com/martin-pouls/the-crew/backend/internal/adapters/memory"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
	"github.com/martin-pouls/the-crew/backend/internal/core/ports"
)

type noShuffle struct{}

func (noShuffle) Shuffle(int, func(i, j int)) {}
func (noShuffle) IntN(int) int                { return 0 }

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

// lobby creates a two-player lobby (host first).
func (f *fixture) lobby(t *testing.T) (domain.GameID, []ports.Session) {
	t.Helper()
	ctx := context.Background()
	host, err := f.svc.CreateGame(ctx, "P1")
	if err != nil {
		t.Fatal(err)
	}
	guest, err := f.svc.JoinGame(ctx, host.GameID, "P2")
	if err != nil {
		t.Fatal(err)
	}
	return host.GameID, []ports.Session{host, guest}
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
	if err != nil || sum.PlayerCount != 2 || !sum.Joinable || sum.MaxPlayers != 5 {
		t.Fatalf("summary: %+v %v", sum, err)
	}
	for _, name := range []string{"Cleo", "Dan", "Eve"} {
		if _, err := f.svc.JoinGame(ctx, host.GameID, name); err != nil {
			t.Fatalf("join %s: %v", name, err)
		}
	}
	if _, err := f.svc.JoinGame(ctx, host.GameID, "Finn"); !errors.Is(err, domain.ErrGameFull) {
		t.Fatalf("sixth player: %v", err)
	}
	sum, _ = f.svc.Summary(ctx, host.GameID)
	if sum.PlayerCount != 5 || sum.Joinable {
		t.Fatalf("full summary: %+v", sum)
	}

	hv, err := f.svc.View(ctx, host.GameID, host.PlayerID)
	if err != nil || len(hv.Players) != 5 || hv.HostID != host.PlayerID || hv.Phase != domain.PhaseLobby {
		t.Fatalf("host view: %+v %v", hv, err)
	}
	if _, err := f.svc.View(ctx, host.GameID, "ghost"); !errors.Is(err, domain.ErrUnknownPlayer) {
		t.Fatalf("unknown viewer: %v", err)
	}

	// The host ends the lobby.
	ev, err := f.svc.EndGame(ctx, host.GameID, host.PlayerID)
	if err != nil || ev.Phase != domain.PhaseEnded {
		t.Fatalf("end: %+v %v", ev, err)
	}
	if _, err := f.svc.JoinGame(ctx, host.GameID, "Late"); !errors.Is(err, domain.ErrWrongPhase) {
		t.Fatalf("join after end: %v", err)
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
	// Once the lobby is ended nobody can leave.
	if _, err := f.svc.EndGame(ctx, host.GameID, host.PlayerID); err != nil {
		t.Fatal(err)
	}
	if err := f.svc.LeaveGame(ctx, host.GameID, host.PlayerID); !errors.Is(err, domain.ErrWrongPhase) {
		t.Fatalf("leave after end: %v", err)
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
	if rec.TokenHashes[host.PlayerID] != hashToken(host.Token) || rec.TokenHashes[host.PlayerID] == host.Token {
		t.Errorf("token not stored hashed")
	}
}

func TestDefaultGenerators(t *testing.T) {
	for range 200 {
		id := defaultGameID()
		if len(id) != 5 {
			t.Fatalf("bad id %q", id)
		}
		for _, c := range string(id) {
			if !containsRune(idgen.CodeAlphabet, c) {
				t.Fatalf("bad char %q in %q", c, id)
			}
		}
	}
	pid := defaultPlayerID()
	if len(pid) != 12 || pid[:2] != "p_" {
		t.Errorf("bad player id %q", pid)
	}
	if a, b := defaultToken(), defaultToken(); a == b || len(a) != 43 {
		t.Errorf("bad tokens %q %q", a, b)
	}
	if defaultRNG() == nil {
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
	if _, err := f.svc.JoinGame(ctx, host.GameID, "Bob"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.EndGame(ctx, host.GameID, host.PlayerID); err != nil {
		t.Fatal(err)
	}

	got := f.rec.all()
	want := [][]string{
		{"PlayerJoined"},
		{"GameEnded"},
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
		{"non-host end", second(f.svc.EndGame(ctx, host.GameID, bob.PlayerID)), domain.ErrNotHost},
		{"unknown player end", second(f.svc.EndGame(ctx, host.GameID, "ghost")), domain.ErrUnknownPlayer},
		{"name taken", second2(f.svc.JoinGame(ctx, host.GameID, "ann")), domain.ErrNameTaken},
		{"invalid name", second2(f.svc.JoinGame(ctx, host.GameID, " ")), domain.ErrInvalidName},
		{"unknown game", second(f.svc.EndGame(ctx, "NOPE1", host.PlayerID)), ports.ErrGameNotFound},
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

func second(_ domain.PlayerView, err error) error { return err }
func second2(_ ports.Session, err error) error    { return err }

func TestConcurrentEndGameSerialised(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	id, sessions := f.lobby(t)
	host := sessions[0]
	base := len(f.rec.all())

	const workers = 8
	var wg sync.WaitGroup
	var ok atomic.Int64
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := f.svc.EndGame(ctx, id, host.PlayerID); err == nil {
				ok.Add(1)
			}
		}()
	}
	wg.Wait()
	if ok.Load() != 1 {
		t.Fatalf("expected exactly one successful end, got %d", ok.Load())
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
