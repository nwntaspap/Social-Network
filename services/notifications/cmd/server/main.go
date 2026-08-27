package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"social-network/services/notifications/internal/config"
	"social-network/services/notifications/internal/consumer"
	"social-network/services/notifications/internal/handler"
	"social-network/services/notifications/internal/middleware"
	"social-network/services/notifications/internal/platform/database"
	"social-network/services/notifications/internal/platform/eventbus"
	"social-network/services/notifications/internal/server"
	"social-network/services/notifications/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := database.NewDB(database.Config{
		Driver: "sqlite3",
		Path:   cfg.DBPath,
		Pragma: "_journal_mode=WAL&_busy_timeout=5000",
	})
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer db.Close()

	if err := database.NewMigrator(db, "migrations").Up(context.Background()); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	repo := store.NewSQLiteStore(db)
	hub := handler.NewStreamHub()
	notif := handler.New(repo, hub)
	auth := middleware.NewAuth(cfg.Backend)

	broker, err := eventbus.NewGoBroker()
	if err != nil {
		log.Fatalf("broker error: %v", err)
	}

	svc := consumer.New(broker, repo, hub, logger)
	go func() {
		if err := svc.Start(context.Background()); err != nil {
			logger.Error("consumer stopped", "error", err)
		}
	}()

	srv := server.New(cfg, logger, auth, notif)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.ListenAndServe(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
