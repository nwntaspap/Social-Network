package database

import (
	"context"
	"testing"
)

func setupMigrator(t *testing.T) (*Migrator, DB) {
	t.Helper()
	db, err := NewDB(Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewMigrator(db, "../../../db/migrations"), db
}

func tableNames(ctx context.Context, db DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make(map[string]bool)
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names[n] = true
	}
	return names, rows.Err()
}

var domainTables = []string{
	"users", "oauth_providers", "sessions", "categories",
	"topics", "topic_categories", "comments", "votes",
	"notifications", "direct_chats", "chat_messages", "chat_reads",
}

func TestMigrator_Up_CreatesTables(t *testing.T) {
	m, db := setupMigrator(t)
	ctx := t.Context()

	if err := m.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	tables, err := tableNames(ctx, db)
	if err != nil {
		t.Fatalf("tableNames() error = %v", err)
	}

	allTables := append([]string{"schema_migrations"}, domainTables...)
	for _, name := range allTables {
		if !tables[name] {
			t.Errorf("expected table %q to exist", name)
		}
	}
}

func TestMigrator_Up_Idempotent(t *testing.T) {
	m, _ := setupMigrator(t)
	ctx := t.Context()

	if err := m.Up(ctx); err != nil {
		t.Fatalf("first Up() error = %v", err)
	}
	if err := m.Up(ctx); err != nil {
		t.Fatalf("second Up() error = %v", err)
	}
}

func TestMigrator_Down_RollsBack(t *testing.T) {
	m, db := setupMigrator(t)
	ctx := t.Context()

	if err := m.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}
	if err := m.Down(ctx); err != nil {
		t.Fatalf("Down() error = %v", err)
	}

	tables, err := tableNames(ctx, db)
	if err != nil {
		t.Fatalf("tableNames() error = %v", err)
	}

	for _, name := range domainTables {
		if tables[name] {
			t.Errorf("expected table %q to be dropped after Down()", name)
		}
	}
	if !tables["schema_migrations"] {
		t.Error("expected schema_migrations table to persist after Down()")
	}
}

func TestMigrator_Down_OnCleanDB(t *testing.T) {
	m, _ := setupMigrator(t)
	ctx := t.Context()

	if err := m.Down(ctx); err != nil {
		t.Fatalf("Down() on clean DB error = %v", err)
	}
}

func TestMigrator_SchemaMigrations_Tracking(t *testing.T) {
	m, db := setupMigrator(t)
	ctx := t.Context()

	if err := m.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	var version int
	row := db.QueryRowContext(ctx, "SELECT version FROM schema_migrations")
	if err := row.Scan(&version); err != nil {
		t.Fatalf("scan schema_migrations version: %v", err)
	}
	if version != 1 {
		t.Errorf("version = %d, want 1", version)
	}

	if err := m.Down(ctx); err != nil {
		t.Fatalf("Down() error = %v", err)
	}

	row = db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&version); err != nil {
		t.Fatalf("scan after down: %v", err)
	}
	if version != 0 {
		t.Errorf("version after Down() = %d, want 0", version)
	}
}
