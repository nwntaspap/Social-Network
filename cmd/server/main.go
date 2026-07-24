package main

import (
	"context"
	"log"

	"social-network/internal/bootstrap"
	"social-network/internal/config"
	coreserver "social-network/internal/core/server"
	"social-network/internal/infra/storage/sqlite"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	db, err := sqlite.InitializeDB(*cfg)
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer db.Close()

	app := bootstrap.Bootstrap(db, cfg)

	srv := coreserver.New(
		cfg,
		coreserver.WithCORS(cfg.AllowedOrigins),
		coreserver.WithAuth(app.SessionStore, cfg.SessionManager.AccessCookieName),
		coreserver.WithHandlers(&coreserver.AllHandlers{
			Follow: app.Follow,
			Chat:   app.Chat,
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.ListenAndServe(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
