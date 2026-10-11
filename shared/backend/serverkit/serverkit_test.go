package serverkit

import (
	"context"
	"log/slog"
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func lookupFrom(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := loadConfig(lookupFrom(nil), ":8082", "data/x.db")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":8082" || cfg.DBPath != "data/x.db" || cfg.StaticDir != "" || cfg.WSOriginPatterns != nil ||
		cfg.GameRetention != 168*time.Hour || cfg.CleanupInterval != time.Hour {
		t.Fatalf("defaults = %+v", cfg)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	cfg, err := loadConfig(lookupFrom(map[string]string{
		"ADDR":               ":9000",
		"DB_PATH":            "/data/y.db",
		"STATIC_DIR":         "../frontend/build",
		"WS_ORIGIN_PATTERNS": " 192.168.*:5173 , ,localhost:* ",
		"GAME_RETENTION":     "2h",
		"CLEANUP_INTERVAL":   "5m",
	}), ":8082", "data/x.db")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":9000" || cfg.DBPath != "/data/y.db" || cfg.StaticDir != "../frontend/build" ||
		!slices.Equal(cfg.WSOriginPatterns, []string{"192.168.*:5173", "localhost:*"}) ||
		cfg.GameRetention != 2*time.Hour || cfg.CleanupInterval != 5*time.Minute {
		t.Fatalf("config = %+v", cfg)
	}
}

func TestLoadConfigRejectsBadDurations(t *testing.T) {
	for _, env := range []map[string]string{
		{"GAME_RETENTION": "soon"},
		{"GAME_RETENTION": "-1h"},
		{"CLEANUP_INTERVAL": "0s"},
	} {
		if _, err := loadConfig(lookupFrom(env), ":1", "x.db"); err == nil {
			t.Fatalf("%v: want an error", env)
		}
	}
}

func TestJanitorPurgesAtStartAndOnTicks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var runs atomic.Int32
	done := make(chan struct{})
	go func() {
		defer close(done)
		Janitor(ctx, slog.New(slog.DiscardHandler), 5*time.Millisecond, func(context.Context) error {
			runs.Add(1)
			return nil
		})
	}()
	deadline := time.Now().Add(5 * time.Second)
	for runs.Load() < 3 {
		if time.Now().After(deadline) {
			t.Fatalf("only %d purges", runs.Load())
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("janitor did not stop")
	}
}
