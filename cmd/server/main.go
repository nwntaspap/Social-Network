package main

import (
	"context"
	"log"

	"social-network/internal/bootstrap"
	"social-network/internal/config"
	coreserver "social-network/internal/core/server"
	"social-network/internal/platform/database"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	db, err := database.InitDB(*cfg)
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
			User:    app.User,
			Follow:  app.Follow,
			Chat:    app.Chat,
			Comment: app.Comment,
			Topic:   app.Topic,
			Group:   app.Group,
			Event:   app.Event,
			OAuth:   app.OAuth,
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.ListenAndServe(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
