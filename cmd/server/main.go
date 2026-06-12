package main

import (
	"log"

	"github.com/arnald/forum/internal/bootstrap"
	"github.com/arnald/forum/internal/config"
	"github.com/arnald/forum/internal/infra/http"
	"github.com/arnald/forum/internal/infra/storage/postgres"
)

func main() {
	// 1. Load configuration first
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// 2. Initialize DB connection
	db, err := postgres.InitializeDB(*cfg)
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer db.Close()

	app := bootstrap.Bootstrap(db, cfg)
	HTTPServer := http.NewServer(cfg, app)
	HTTPServer.ListenAndServe()
}
