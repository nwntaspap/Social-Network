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
	"users", "oauth_providers", "sessions",
	"topics", "topic_allowed_users", "comments", "votes",
	"notifications", "follows", "follow_requests",
	"chats", "messages", "chat_reads",
	"groups", "group_members", "group_invitations", "group_join_requests",
	"group_chat_messages", "group_posts", "group_post_comments",
	"events", "event_options", "event_rsvps",
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
		if !tables[name] {
			t.Errorf("expected table %q to still exist after partial Down()", name)
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
	row := db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&version); err != nil {
		t.Fatalf("scan schema_migrations version: %v", err)
	}

	totalMigrations, err := m.discover()
	if err != nil {
		t.Fatalf("discover() error = %v", err)
	}
	expectedVersion := len(totalMigrations)
	if version != expectedVersion {
		t.Errorf("version = %d, want %d", version, expectedVersion)
	}

	if err := m.Down(ctx); err != nil {
		t.Fatalf("Down() error = %v", err)
	}

	row = db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&version); err != nil {
		t.Fatalf("scan after down: %v", err)
	}
	if version != expectedVersion-1 {
		t.Errorf("version after Down() = %d, want %d", version, expectedVersion-1)
	}
}

func TestSeed_ApplySeedData(t *testing.T) {
	migrator, db := setupMigrator(t)
	ctx := context.Background()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	if err := Seed(ctx, db, "../../../db"); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	var userCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if userCount != 5 {
		t.Errorf("user count = %d, want 5", userCount)
	}

	var topicCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM topics").Scan(&topicCount); err != nil {
		t.Fatalf("count topics: %v", err)
	}
	if topicCount != 8 {
		t.Errorf("topic count = %d, want 8", topicCount)
	}

	var commentCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments").Scan(&commentCount); err != nil {
		t.Fatalf("count comments: %v", err)
	}
	if commentCount != 10 {
		t.Errorf("comment count = %d, want 10", commentCount)
	}
}

func TestSeed_Idempotent(t *testing.T) {
	migrator, db := setupMigrator(t)
	ctx := context.Background()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	if err := Seed(ctx, db, "../../../db"); err != nil {
		t.Fatalf("Seed() first call error = %v", err)
	}
	if err := Seed(ctx, db, "../../../db"); err != nil {
		t.Fatalf("Seed() second call error = %v", err)
	}

	var userCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if userCount != 5 {
		t.Errorf("user count after double seed = %d, want 5", userCount)
	}
}

func TestSeed_MissingFile(t *testing.T) {
	db, err := NewDB(Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := Seed(context.Background(), db, "/nonexistent/path"); err != nil {
		t.Errorf("Seed() with missing file should return nil, got = %v", err)
	}
}
