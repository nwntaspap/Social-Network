package main

import (
	"log"

	"social-network/internal/bootstrap"
	"social-network/internal/config"
	"social-network/internal/infra/http"
	"social-network/internal/infra/storage/sqlite"
)

func main() {
	// 1. Load configuration first
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// 2. Initialize DB connection
	db, err := sqlite.InitializeDB(*cfg)
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer db.Close()

	app := bootstrap.Bootstrap(db, cfg)
	HTTPServer := http.NewServer(cfg, app)
	HTTPServer.ListenAndServe()
}
