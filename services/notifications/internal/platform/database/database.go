package database

import (
	"context"
	"database/sql"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	PingContext(ctx context.Context) error
}

type Config struct {
	Driver string
	Path   string
	Pragma string
}

func NewDB(cfg Config) (DB, error) {
	switch cfg.Driver {
	case "sqlite3":
		return newSQLite(cfg)
	default:
		return newSQLite(cfg)
	}
}
