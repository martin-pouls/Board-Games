package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/martin-pouls/the-game/backend/internal/adapters/apidto"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func TestBucketsBurstThenRefill(t *testing.T) {
	clock := &fakeClock{t: time.Unix(1_000_000, 0)}
	b := newBuckets(Rate{Burst: 3, Every: 10 * time.Second}, clock.now)

	for i := range 3 {
		if wait := b.take("a"); wait != 0 {
			t.Fatalf("take %d: wait %v, want a token", i+1, wait)
		}
	}
	if wait := b.take("a"); wait != 10*time.Second {
		t.Fatalf("empty bucket: wait %v, want 10s", wait)
	}
	if wait := b.take("b"); wait != 0 {
		t.Fatalf("other key: wait %v, want its own full bucket", wait)
	}
	clock.advance(4 * time.Second)
	if wait := b.wait("a"); wait != 6*time.Second {
		t.Fatalf("after 4s: wait %v, want 6s", wait)
	}
	clock.advance(6 * time.Second)
	if wait := b.take("a"); wait != 0 {
		t.Fatalf("after 10s: wait %v, want a token", wait)
	}
	if wait := b.wait("a"); wait == 0 {
		t.Fatal("wait must not use a token, and only one came back")
	}
}

func TestBucketsSweepDropsFullBuckets(t *testing.T) {
	clock := &fakeClock{t: time.Unix(1_000_000, 0)}
	b := newBuckets(Rate{Burst: 2, Every: time.Minute}, clock.now)
	b.take("idle")
	b.take("busy")
	b.take("busy")
	clock.advance(90 * time.Second) // idle is full again, busy has 1.5 tokens
	b.take("other")
	if _, ok := b.m["idle"]; ok {
		t.Error("full bucket was not swept")
	}
	if _, ok := b.m["busy"]; !ok {
		t.Error("partly used bucket was swept")
	}
}

func TestZeroRateDisablesLimit(t *testing.T) {
	var b *buckets = newBuckets(Rate{}, time.Now)
	if b != nil {
		t.Fatal("zero rate must give no buckets")
	}
	if b.take("a") != 0 || b.wait("a") != 0 {
		t.Fatal("nil buckets must never limit")
	}
}

func TestClientKey(t *testing.T) {
	tests := []struct {
		name    string
		remote  string
		xff     []string
		key     string
		limited bool
	}{
		{"peer without header", "203.0.113.7:4000", nil, "203.0.113.7", true},
		{"forwarded by proxy", "172.18.0.2:4000", []string{"203.0.113.7"}, "203.0.113.7", true},
		{"last entry of a list", "172.18.0.2:4000", []string{"198.51.100.1, 203.0.113.7"}, "203.0.113.7", true},
		{"last of several headers", "172.18.0.2:4000", []string{"198.51.100.1", "203.0.113.7"}, "203.0.113.7", true},
		{"ipv6 grouped by /64", "[2001:db8:1:2:aaaa::1]:4000", nil, "2001:db8:1:2::/64", true},
		{"ipv4-mapped ipv6", "[::ffff:203.0.113.7]:4000", nil, "203.0.113.7", true},
		{"loopback v4 is local dev", "127.0.0.1:4000", nil, "", false},
		{"loopback v6 is local dev", "[::1]:4000", nil, "", false},
		{"unparsable counts as is", "172.18.0.2:4000", []string{"garbage"}, "garbage", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remote
			for _, v := range tt.xff {
				r.Header.Add("X-Forwarded-For", v)
			}
			key, limited := clientKey(r)
			if key != tt.key || limited != tt.limited {
				t.Fatalf("clientKey = %q, %v; want %q, %v", key, limited, tt.key, tt.limited)
			}
		})
	}
}

// limitedRouter is newTestRouter with the given limits.
func limitedRouter(svc *fakeService, ws http.Handler, l RateLimits) http.Handler {
	return NewRouter(Deps{
		Health: func(context.Context) error { return nil },
		Games:  svc,
		WS:     ws,
		Limits: &l,
	})
}

func doFrom(h http.Handler, ip, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.RemoteAddr = ip + ":4000"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func assertRateLimited(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status %d, want 429; body %s", w.Code, w.Body)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("missing Retry-After")
	}
	var resp apidto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.Error.Code != apidto.CodeRateLimited {
		t.Fatalf("body %s, want code %s", w.Body, apidto.CodeRateLimited)
	}
}

func TestCreateLimit(t *testing.T) {
	svc := &fakeService{}
	h := limitedRouter(svc, nil, RateLimits{Create: Rate{Burst: 2, Every: time.Hour}})
	create := func(ip string) *httptest.ResponseRecorder {
		return doFrom(h, ip, http.MethodPost, "/api/games", `{"name":"Ann"}`)
	}

	for i := range 2 {
		if w := create("203.0.113.7"); w.Code != http.StatusCreated {
			t.Fatalf("create %d: status %d", i+1, w.Code)
		}
	}
	assertRateLimited(t, create("203.0.113.7"))
	if w := create("198.51.100.1"); w.Code != http.StatusCreated {
		t.Fatalf("another client: status %d", w.Code)
	}
	for range 5 {
		if w := create("127.0.0.1"); w.Code != http.StatusCreated {
			t.Fatalf("loopback: status %d", w.Code)
		}
	}
	if w := doFrom(h, "203.0.113.7", http.MethodGet, "/api/games/"+string(testGame), ""); w.Code != http.StatusOK {
		t.Fatalf("lookups are not limited by creates: status %d", w.Code)
	}
}

func TestMissLimit(t *testing.T) {
	svc := &fakeService{}
	h := limitedRouter(svc, nil, RateLimits{Misses: Rate{Burst: 2, Every: time.Hour}})
	get := func(code string) *httptest.ResponseRecorder {
		return doFrom(h, "203.0.113.7", http.MethodGet, "/api/games/"+code, "")
	}

	for range 5 {
		if w := get(string(testGame)); w.Code != http.StatusOK {
			t.Fatalf("hits must not count: status %d", w.Code)
		}
	}
	for i := range 2 {
		if w := get("ZZZZZ"); w.Code != http.StatusNotFound {
			t.Fatalf("miss %d: status %d", i+1, w.Code)
		}
	}
	assertRateLimited(t, get("ZZZZZ"))
	assertRateLimited(t, get(string(testGame)))
	before := len(svc.calls)
	assertRateLimited(t, doFrom(h, "203.0.113.7", http.MethodPost, "/api/games/"+string(testGame)+"/players", `{"name":"Ben"}`))
	if len(svc.calls) != before {
		t.Fatal("a refused request reached the service")
	}
}

func TestSocketLimits(t *testing.T) {
	release := make(chan struct{})
	var wg sync.WaitGroup
	entered := make(chan struct{}, 10)
	ws := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		<-release // the socket stays open until released
	})
	h := limitedRouter(&fakeService{}, ws, RateLimits{
		Sockets:    Rate{Burst: 3, Every: time.Hour},
		MaxSockets: 2,
	})
	connect := func() *httptest.ResponseRecorder {
		return doFrom(h, "203.0.113.7", http.MethodGet, "/ws/games/"+string(testGame), "")
	}

	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			connect()
		}()
		<-entered
	}
	assertRateLimited(t, connect()) // third open socket
	close(release)
	wg.Wait()
	assertRateLimited(t, connect()) // sockets closed, but 3 upgrades used the burst
}
