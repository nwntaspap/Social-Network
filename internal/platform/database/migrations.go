package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Migration struct {
	Version  int
	UpFile   string
	DownFile string
}

type Migrator struct {
	db  DB
	dir string
}

func NewMigrator(db DB, migrationsDir string) *Migrator {
	return &Migrator{
		db:  db,
		dir: migrationsDir,
	}
}

func (m *Migrator) Up(ctx context.Context) error {
	if err := m.ensureMetaTable(ctx); err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}

	migrations, err := m.discover()
	if err != nil {
		return err
	}

	applied, err := m.appliedVersions(ctx)
	if err != nil {
		return err
	}

	for _, mig := range migrations {
		if applied[mig.Version] {
			continue
		}

		content, err := os.ReadFile(mig.UpFile)
		if err != nil {
			return fmt.Errorf("read migration %d: %w", mig.Version, err)
		}

		if err := m.execBatch(ctx, string(content)); err != nil {
			return fmt.Errorf("migration %d: %w", mig.Version, err)
		}

		if _, err := m.db.ExecContext(
			ctx,
			"INSERT INTO schema_migrations (version) VALUES (?)", mig.Version,
		); err != nil {
			return fmt.Errorf("record migration %d: %w", mig.Version, err)
		}
	}

	return nil
}

func (m *Migrator) Down(ctx context.Context) error {
	if err := m.ensureMetaTable(ctx); err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}

	row := m.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	var version int
	if err := row.Scan(&version); err != nil {
		return fmt.Errorf("query latest migration: %w", err)
	}
	if version == 0 {
		return nil
	}

	mig, err := m.find(version)
	if err != nil {
		return err
	}
	if mig == nil {
		return fmt.Errorf("migration file for version %d not found", version)
	}

	content, err := os.ReadFile(mig.DownFile)
	if err != nil {
		return fmt.Errorf("read down migration %d: %w", version, err)
	}

	if err := m.execBatch(ctx, string(content)); err != nil {
		return fmt.Errorf("down migration %d: %w", version, err)
	}

	if _, err := m.db.ExecContext(
		ctx,
		"DELETE FROM schema_migrations WHERE version = ?", version,
	); err != nil {
		return fmt.Errorf("unrecord migration %d: %w", version, err)
	}

	return nil
}

func (m *Migrator) ensureMetaTable(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations
		(
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	return err
}

func (m *Migrator) discover() ([]Migration, error) {
	files, err := filepath.Glob(filepath.Join(m.dir, "*.up.sql"))
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, f := range files {
		base := filepath.Base(f)
		parts := strings.SplitN(base, "_", 2)
		if len(parts) < 2 {
			continue
		}
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		migrations = append(migrations, Migration{
			Version:  version,
			UpFile:   f,
			DownFile: strings.TrimSuffix(f, ".up.sql") + ".down.sql",
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (m *Migrator) find(version int) (*Migration, error) {
	migrations, err := m.discover()
	if err != nil {
		return nil, err
	}

	for _, mig := range migrations {
		if mig.Version == version {
			return &mig, nil
		}
	}

	return nil, fmt.Errorf("migration file for version %d not found", version)
}

func (m *Migrator) appliedVersions(ctx context.Context) (map[int]bool, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func (m *Migrator) execBatch(ctx context.Context, content string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for stmt := range strings.SplitSeq(content, ";") {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, trimmed); err != nil {
			return fmt.Errorf("execute %q: %w", trimmed, err)
		}
	}

	return tx.Commit()
}
