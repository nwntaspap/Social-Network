package store

import (
	"context"
	"testing"

	"social-network/internal/group"
	"social-network/internal/platform/database"
)

const groupsSchema = `
CREATE TABLE groups (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    creator_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

func setupGroupStore(t *testing.T) *SQLiteStore {
	t.Helper()

	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(), groupsSchema); err != nil {
		t.Fatalf("create table: %v", err)
	}

	return NewSQLiteStore(db)
}

func seedGroup(t *testing.T, s *SQLiteStore, g *group.Group) {
	t.Helper()
	if err := s.CreateGroup(context.Background(), g); err != nil {
		t.Fatalf("seed group %s: %v", g.ID, err)
	}
}

func seedGroupAt(t *testing.T, s *SQLiteStore, g *group.Group, createdAt string) {
	t.Helper()
	if _, err := s.db.ExecContext(context.Background(),
		`INSERT INTO groups (id, title, description, creator_id, created_at) VALUES (?, ?, ?, ?, ?)`,
		g.ID, g.Title, g.Description, g.CreatorID, createdAt); err != nil {
		t.Fatalf("seed group %s: %v", g.ID, err)
	}
}

func TestSearchGroups_FiltersByTitleAndDescription(t *testing.T) {
	s := setupGroupStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go Meetup", Description: "Gophers hang out", CreatorID: "u1"})
	seedGroup(t, s, &group.Group{ID: "g2", Title: "Rust Users", Description: "Ferris fans", CreatorID: "u1"})
	seedGroup(t, s, &group.Group{ID: "g3", Title: "Weekly Reads", Description: "Read about Go and Rust", CreatorID: "u1"})

	groups, total, err := s.SearchGroups(ctx, "go", 1, 10)
	if err != nil {
		t.Fatalf("SearchGroups(go) error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(groups) != 2 {
		t.Fatalf("SearchGroups(go) returned %d groups, want 2", len(groups))
	}

	ids := map[string]bool{}
	for _, g := range groups {
		ids[g.ID] = true
	}
	if !ids["g1"] || !ids["g3"] {
		t.Errorf("SearchGroups(go) ids = %v, want g1 and g3", ids)
	}

	groups, total, err = s.SearchGroups(ctx, "ferris", 1, 10)
	if err != nil {
		t.Fatalf("SearchGroups(ferris) error = %v", err)
	}
	if total != 1 || len(groups) != 1 || groups[0].ID != "g2" {
		t.Errorf("SearchGroups(ferris) = %+v (total %d), want [g2] (1)", groups, total)
	}
}

func TestSearchGroups_EmptyQueryReturnsAll(t *testing.T) {
	s := setupGroupStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go", CreatorID: "u1"})
	seedGroup(t, s, &group.Group{ID: "g2", Title: "Rust", CreatorID: "u1"})
	seedGroup(t, s, &group.Group{ID: "g3", Title: "Zig", CreatorID: "u1"})

	groups, total, err := s.SearchGroups(ctx, "", 1, 10)
	if err != nil {
		t.Fatalf("SearchGroups(empty) error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(groups) != 3 {
		t.Errorf("SearchGroups(empty) returned %d groups, want 3", len(groups))
	}
}

func TestSearchGroups_Paginates(t *testing.T) {
	s := setupGroupStore(t)
	ctx := context.Background()

	seedGroupAt(t, s, &group.Group{ID: "g1", Title: "Go", CreatorID: "u1"}, "2024-01-01 10:00:00")
	seedGroupAt(t, s, &group.Group{ID: "g2", Title: "Rust", CreatorID: "u1"}, "2024-01-01 11:00:00")
	seedGroupAt(t, s, &group.Group{ID: "g3", Title: "Zig", CreatorID: "u1"}, "2024-01-01 12:00:00")

	groups, total, err := s.SearchGroups(ctx, "", 2, 2)
	if err != nil {
		t.Fatalf("SearchGroups(page2) error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(groups) != 1 {
		t.Fatalf("SearchGroups(page2) returned %d groups, want 1", len(groups))
	}
	if groups[0].ID != "g1" {
		t.Errorf("groups[0].ID = %q, want g1", groups[0].ID)
	}
}
