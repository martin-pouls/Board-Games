// Command server runs the access gate: Caddy's forward_auth target and the login page.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/martin-pouls/shared/backend/serverkit"

	"github.com/martin-pouls/gate/backend/internal/gate"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func run(logger *slog.Logger) error {
	// :8080 is The Game's, :8081 Regicide's, :8082 Fox Duet's; the image sets ADDR=:8080.
	addr := env("ADDR", ":8083")
	cfg, err := gate.LoadConfig(
		env("GATE_SECRET_FILE", "data/secret"),
		env("GATE_PASSWORDS_FILE", "data/passwords"),
		env("GATE_BASE_PATH", "/gate"),
		env("GATE_DEFAULT_NEXT", "/board-games/"),
	)
	if err != nil {
		return err
	}
	logger.Info("config loaded", "passwords", len(cfg.Passwords))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return serverkit.Serve(ctx, logger, addr, gate.NewHandler(cfg, logger, nil))
}
