package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"social-network/internal/config"
)

// InitDB opens the database, runs migrations, optionally seeds, and returns the DB handle.
func InitDB(cfg config.ServerConfig) (DB, error) {
	db, err := NewDB(Config{
		Driver: cfg.Database.Driver,
		Path:   cfg.Database.Path,
		Pragma: cfg.Database.Pragma,
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	migrationsDir := "db/migrations"

	if cfg.Database.MigrateOnStart {
		migrator := NewMigrator(db, migrationsDir)
		if err := migrator.Up(context.Background()); err != nil {
			return nil, fmt.Errorf("run migrations: %w", err)
		}
	}

	if cfg.Database.SeedOnStart {
		if err := Seed(context.Background(), db, filepath.Dir(migrationsDir)); err != nil {
			return nil, fmt.Errorf("seed database: %w", err)
		}
	}

	return db, nil
}

// Seed reads and executes the seed SQL file if it exists.
// The seed file is expected at <seedsDir>/seeds/dev_data.sql relative to the working directory.
func Seed(ctx context.Context, db DB, seedsDir string) error {
	seedPath := filepath.Clean(filepath.Join(seedsDir, "seeds", "dev_data.sql"))

	data, err := os.ReadFile(seedPath) // #nosec G304 -- path is constructed from internal config, not user input
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("[seed] no seed file found, skipping")
			return nil
		}
		return fmt.Errorf("read seed file: %w", err)
	}

	if len(data) == 0 {
		log.Println("[seed] seed file is empty, skipping")
		return nil
	}

	if _, err := db.ExecContext(ctx, string(data)); err != nil {
		return fmt.Errorf("execute seed: %w", err)
	}

	log.Println("[seed] seed data applied successfully")
	return nil
}
