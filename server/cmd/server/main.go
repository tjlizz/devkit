package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"devkit/server/internal/config"
	"devkit/server/internal/database"
	"devkit/server/internal/router"
	httpserver "devkit/server/internal/server"
	"devkit/server/migrations"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load(os.Getenv("DEVKIT_CONFIG"))
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(cfg.Database.Path, migrations.Files)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	app := httpserver.New(cfg, router.New(logger, router.WithAuth(db, cfg.Auth)), logger)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
