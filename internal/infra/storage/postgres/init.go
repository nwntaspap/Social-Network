package postgres

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/arnald/forum/internal/config"
)

func InitializeDB(cfg config.ServerConfig) (*sql.DB, error) {
	db, result, err := OpenDB(cfg)
	if err != nil {
		return result, err
	}

	return db, nil
}

func OpenDB(cfg config.ServerConfig) (*sql.DB, *sql.DB, error) {
	db, err := sql.Open("postgres", cfg.Database.PostgresURL)
	if err != nil {
		return nil, nil, err
	}

	db.SetMaxOpenConns(cfg.Database.OpenConn)
	return db, nil, nil
}
