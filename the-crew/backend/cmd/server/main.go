// Command server runs the The Crew backend: REST + WebSocket API backed by SQLite.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/martin-pouls/shared/backend/serverkit"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/httpapi"
	"github.com/martin-pouls/the-crew/backend/internal/adapters/sqlite"
	"github.com/martin-pouls/the-crew/backend/internal/adapters/ws"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
	"github.com/martin-pouls/the-crew/backend/internal/core/service"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	// One deployment serves one edition (EDITION). Local ports and databases are per edition:
	// Planet Nine :8084 (data/the-crew.db), Deep Sea :8085 (data/deep-sea.db); :8080 is The Game's,
	// :8081 Regicide's, :8082 Fox Duet's; the images set ADDR=:8080 and DB_PATH.
	edition := domain.Edition(os.Getenv("EDITION"))
	if edition == "" {
		edition = domain.EditionPlanetNine
	}
	if !edition.Valid() {
		return fmt.Errorf("invalid EDITION %q: want %q or %q", edition, domain.EditionPlanetNine, domain.EditionDeepSea)
	}
	addr, dbPath := ":8084", "data/the-crew.db"
	if edition == domain.EditionDeepSea {
		addr, dbPath = ":8085", "data/deep-sea.db"
	}
	cfg, err := serverkit.LoadConfig(addr, dbPath)
	if err != nil {
		return err
	}
	logger.Info("edition", "edition", edition)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sqlite.Open(ctx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()
	logger.Info("database open", "db", cfg.DBPath)

	repo := sqlite.NewGameRepository(db)
	hub := ws.NewHub(logger)
	svc := service.New(repo, hub, service.Config{Edition: edition, Logger: logger})
	go serverkit.Janitor(ctx, logger, cfg.CleanupInterval, func(ctx context.Context) error {
		_, err := svc.PurgeInactive(ctx, cfg.GameRetention)
		return err
	})

	limits := httpapi.DefaultRateLimits()
	return serverkit.Serve(ctx, logger, cfg.Addr, httpapi.NewRouter(httpapi.Deps{
		Logger:    logger,
		Edition:   edition,
		Health:    db.Ping,
		StaticDir: cfg.StaticDir,
		Games:     svc,
		WS:        ws.NewHandler(hub, svc, ws.Options{OriginPatterns: cfg.WSOriginPatterns}),
		Limits:    &limits,
	}))
}
