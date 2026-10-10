// Command server runs the The Game backend: REST + WebSocket API backed by SQLite.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/martin-pouls/the-game/backend/internal/adapters/httpapi"
	"github.com/martin-pouls/the-game/backend/internal/adapters/sqlite"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sqlite.Open(ctx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	srv := &http.Server{
		Addr: cfg.Addr,
		Handler: httpapi.NewRouter(httpapi.Deps{
			Logger:    logger,
			Health:    db.Ping,
			StaticDir: cfg.StaticDir,
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

type config struct {
	Addr      string
	DBPath    string
	StaticDir string
}

func loadConfig() config {
	return config{
		Addr:      env("ADDR", ":8080"),
		DBPath:    env("DB_PATH", "data/the-game.db"),
		StaticDir: env("STATIC_DIR", ""),
	}
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
