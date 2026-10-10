// Command server runs the Regicide backend: REST + WebSocket API backed by SQLite.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/martin-pouls/regicide/backend/internal/adapters/httpapi"
	"github.com/martin-pouls/regicide/backend/internal/adapters/sqlite"
	"github.com/martin-pouls/regicide/backend/internal/adapters/ws"
	"github.com/martin-pouls/regicide/backend/internal/core/service"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sqlite.Open(ctx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	repo := sqlite.NewGameRepository(db)
	hub := ws.NewHub(logger)
	svc := service.New(repo, hub, service.Config{Logger: logger})
	go janitor(ctx, logger, svc, cfg.GameRetention, cfg.CleanupInterval)

	srv := &http.Server{
		Addr: cfg.Addr,
		Handler: httpapi.NewRouter(httpapi.Deps{
			Logger:    logger,
			Health:    db.Ping,
			StaticDir: cfg.StaticDir,
			Games:     svc,
			WS:        ws.NewHandler(hub, svc, ws.Options{OriginPatterns: cfg.WSOriginPatterns}),
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", cfg.Addr, "db", cfg.DBPath)
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

// janitor purges abandoned games once at startup and then on every tick until ctx is cancelled.
func janitor(ctx context.Context, logger *slog.Logger, svc *service.Service, retention, interval time.Duration) {
	purge := func() {
		if _, err := svc.PurgeInactive(ctx, retention); err != nil && ctx.Err() == nil {
			logger.Error("purge inactive games", "err", err)
		}
	}
	purge()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			purge()
		}
	}
}

type config struct {
	Addr             string
	DBPath           string
	StaticDir        string
	WSOriginPatterns []string
	GameRetention    time.Duration
	CleanupInterval  time.Duration
}

func loadConfig() (config, error) {
	cfg := config{
		Addr:      env("ADDR", ":8081"), // :8080 is The Game's; the image sets ADDR=:8080
		DBPath:    env("DB_PATH", "data/regicide.db"),
		StaticDir: env("STATIC_DIR", ""),
	}
	for _, p := range strings.Split(env("WS_ORIGIN_PATTERNS", ""), ",") {
		if p = strings.TrimSpace(p); p != "" {
			cfg.WSOriginPatterns = append(cfg.WSOriginPatterns, p)
		}
	}
	var err error
	if cfg.GameRetention, err = envDuration("GAME_RETENTION", 168*time.Hour); err != nil {
		return config{}, err
	}
	if cfg.CleanupInterval, err = envDuration("CLEANUP_INTERVAL", time.Hour); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
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

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
