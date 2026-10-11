// Package serverkit is the bootstrap every game's cmd/server shares: configuration from the
// environment, serving HTTP with graceful shutdown, and the janitor that purges idle games.
package serverkit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// Config is a game server's configuration (see the env vars in the root CLAUDE.md).
type Config struct {
	Addr             string        // ADDR
	DBPath           string        // DB_PATH
	StaticDir        string        // STATIC_DIR (optional)
	WSOriginPatterns []string      // WS_ORIGIN_PATTERNS (comma-separated, optional)
	GameRetention    time.Duration // GAME_RETENTION, default 168h
	CleanupInterval  time.Duration // CLEANUP_INTERVAL, default 1h
}

// LoadConfig reads the environment. defaultAddr and defaultDBPath are the game's local
// defaults (each game has its own port and database file).
func LoadConfig(defaultAddr, defaultDBPath string) (Config, error) {
	return loadConfig(os.LookupEnv, defaultAddr, defaultDBPath)
}

func loadConfig(lookup func(string) (string, bool), defaultAddr, defaultDBPath string) (Config, error) {
	env := func(key, fallback string) string {
		if v, ok := lookup(key); ok && v != "" {
			return v
		}
		return fallback
	}
	duration := func(key string, fallback time.Duration) (time.Duration, error) {
		v := env(key, "")
		if v == "" {
			return fallback, nil
		}
		d, err := time.ParseDuration(v)
		if err != nil {
			return 0, fmt.Errorf("invalid %s %q: %w", key, v, err)
		}
		if d <= 0 {
			return 0, fmt.Errorf("invalid %s %q: must be positive", key, v)
		}
		return d, nil
	}

	cfg := Config{
		Addr:      env("ADDR", defaultAddr),
		DBPath:    env("DB_PATH", defaultDBPath),
		StaticDir: env("STATIC_DIR", ""),
	}
	for _, p := range strings.Split(env("WS_ORIGIN_PATTERNS", ""), ",") {
		if p = strings.TrimSpace(p); p != "" {
			cfg.WSOriginPatterns = append(cfg.WSOriginPatterns, p)
		}
	}
	var err error
	if cfg.GameRetention, err = duration("GAME_RETENTION", 168*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.CleanupInterval, err = duration("CLEANUP_INTERVAL", time.Hour); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Serve runs handler on addr until ctx is cancelled, then shuts down gracefully (10 s).
func Serve(ctx context.Context, logger *slog.Logger, addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
	return nil
}

// Janitor calls purge once at startup and then on every tick of interval until ctx is
// cancelled. purge deletes games idle longer than the retention and logs what it removed.
func Janitor(ctx context.Context, logger *slog.Logger, interval time.Duration, purge func(context.Context) error) {
	run := func() {
		if err := purge(ctx); err != nil && ctx.Err() == nil {
			logger.Error("purge inactive games", "err", err)
		}
	}
	run()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			run()
		}
	}
}
