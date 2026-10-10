package httpapi

import (
	"log/slog"
	"math"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/martin-pouls/regicide/backend/internal/adapters/apidto"
)

// RateLimits are per-client abuse limits. They are generous on purpose: they stop scripts that
// create games in a loop, guess game codes or hold many sockets open, never people playing or
// testing. A zero Rate (or MaxSockets) disables that limit.
type RateLimits struct {
	Create     Rate // POST /api/games
	Misses     Rate // /api/games/{code}... answered with 404 (guessed codes); used up, all of them are refused
	Sockets    Rate // WebSocket upgrades
	MaxSockets int  // open WebSockets at once
}

// Rate is a token bucket: Burst requests at once, then one more per Every.
type Rate struct {
	Burst int
	Every time.Duration
}

// DefaultRateLimits are the production limits.
func DefaultRateLimits() RateLimits {
	return RateLimits{
		Create:     Rate{Burst: 30, Every: 10 * time.Second}, // then 6 a minute
		Misses:     Rate{Burst: 60, Every: 5 * time.Second},  // then 12 a minute
		Sockets:    Rate{Burst: 60, Every: 2 * time.Second},  // then 30 a minute
		MaxSockets: 20,
	}
}

type rateLimiter struct {
	logger                  *slog.Logger
	create, misses, sockets *buckets
	maxSockets              int

	mu   sync.Mutex
	open map[string]int // open WebSockets per client
}

func newRateLimiter(l RateLimits, logger *slog.Logger, now func() time.Time) *rateLimiter {
	return &rateLimiter{
		logger:     logger,
		create:     newBuckets(l.Create, now),
		misses:     newBuckets(l.Misses, now),
		sockets:    newBuckets(l.Sockets, now),
		maxSockets: l.MaxSockets,
		open:       map[string]int{},
	}
}

func (rl *rateLimiter) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, limited := clientKey(r)
		path := r.URL.Path
		switch {
		case !limited:
			next.ServeHTTP(w, r)
		case r.Method == http.MethodPost && path == "/api/games":
			if wait := rl.create.take(key); wait > 0 {
				rl.refuse(w, r, "create", wait)
				return
			}
			next.ServeHTTP(w, r)
		case strings.HasPrefix(path, "/api/games/"):
			if wait := rl.misses.wait(key); wait > 0 {
				rl.refuse(w, r, "misses", wait)
				return
			}
			sw := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(sw, r)
			if sw.status == http.StatusNotFound {
				rl.misses.take(key)
			}
		case strings.HasPrefix(path, "/ws/"):
			if wait := rl.sockets.take(key); wait > 0 {
				rl.refuse(w, r, "sockets", wait)
				return
			}
			if !rl.acquire(key) {
				rl.refuse(w, r, "open sockets", 10*time.Second)
				return
			}
			defer rl.release(key)
			next.ServeHTTP(w, r) // returns when the socket closes
		default:
			next.ServeHTTP(w, r)
		}
	})
}

// refuse answers 429. The client address is not logged.
func (rl *rateLimiter) refuse(w http.ResponseWriter, r *http.Request, limit string, wait time.Duration) {
	rl.logger.Warn("rate limited", "limit", limit, "method", r.Method, "path", r.URL.Path)
	w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
	status, body := apidto.ErrorFor(apidto.ErrRateLimited)
	writeJSON(w, status, apidto.ErrorResponse{Error: body})
}

func (rl *rateLimiter) acquire(key string) bool {
	if rl.maxSockets <= 0 {
		return true
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.open[key] >= rl.maxSockets {
		return false
	}
	rl.open[key]++
	return true
}

func (rl *rateLimiter) release(key string) {
	if rl.maxSockets <= 0 {
		return
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.open[key]--; rl.open[key] <= 0 {
		delete(rl.open, key)
	}
}

// clientKey is what the limits count a request under: the client IP, or its /64 for IPv6 (a host
// usually has a whole /64). In production Caddy is the only way in: it sets X-Forwarded-For to
// the client address and drops values sent by the client, so the last entry is the client.
// Without the header the TCP peer counts. Loopback clients are not limited (limited is false):
// that is local dev, where the Vite proxy connects from localhost.
func clientKey(r *http.Request) (key string, limited bool) {
	addr := r.RemoteAddr
	if xff := r.Header.Values("X-Forwarded-For"); len(xff) > 0 {
		last := xff[len(xff)-1]
		addr = strings.TrimSpace(last[strings.LastIndexByte(last, ',')+1:])
	} else if host, _, err := net.SplitHostPort(addr); err == nil {
		addr = host
	}
	ip, err := netip.ParseAddr(addr)
	if err != nil {
		return addr, true
	}
	ip = ip.Unmap().WithZone("")
	switch {
	case ip.IsLoopback():
		return "", false
	case ip.Is6():
		p, _ := ip.Prefix(64)
		return p.String(), true
	default:
		return ip.String(), true
	}
}

// statusWriter records the response status.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// buckets holds one token bucket per key. Full buckets are dropped, so memory only holds
// clients seen within the last Burst×Every.
type buckets struct {
	rate Rate
	now  func() time.Time

	mu    sync.Mutex
	m     map[string]*bucket
	swept time.Time
}

type bucket struct {
	tokens float64
	at     time.Time
}

// newBuckets returns nil (no limit) for a zero rate; the methods accept a nil receiver.
func newBuckets(rate Rate, now func() time.Time) *buckets {
	if rate.Burst <= 0 || rate.Every <= 0 {
		return nil
	}
	return &buckets{rate: rate, now: now, m: map[string]*bucket{}, swept: now()}
}

// take uses one token of key's bucket. When none is left it uses nothing and returns how long
// until one is back; 0 means the token was taken.
func (b *buckets) take(key string) time.Duration {
	if b == nil {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	k := b.fill(key)
	if wait := b.waitFor(k); wait > 0 {
		return wait
	}
	k.tokens--
	return 0
}

// wait returns how long until key's bucket has a token; 0 means it has one now.
func (b *buckets) wait(key string) time.Duration {
	if b == nil {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.waitFor(b.fill(key))
}

func (b *buckets) waitFor(k *bucket) time.Duration {
	if k.tokens >= 1 {
		return 0
	}
	return time.Duration((1 - k.tokens) * float64(b.rate.Every))
}

// fill returns key's bucket topped up to now, after sweeping full buckets once a minute.
// Callers hold mu.
func (b *buckets) fill(key string) *bucket {
	now := b.now()
	burst := float64(b.rate.Burst)
	refill := func(k *bucket) float64 {
		return min(burst, k.tokens+float64(now.Sub(k.at))/float64(b.rate.Every))
	}
	if now.Sub(b.swept) >= time.Minute {
		for key, k := range b.m {
			if refill(k) >= burst {
				delete(b.m, key)
			}
		}
		b.swept = now
	}
	k, ok := b.m[key]
	if !ok {
		k = &bucket{tokens: burst, at: now}
		b.m[key] = k
		return k
	}
	k.tokens, k.at = refill(k), now
	return k
}
