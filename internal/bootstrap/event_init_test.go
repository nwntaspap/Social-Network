package bootstrap

import (
	"context"
	"testing"

	"social-network/internal/platform/database"
)

func setupMemberChecker(t *testing.T) *groupMemberChecker {
	t.Helper()

	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	schema := `CREATE TABLE group_members (group_id TEXT NOT NULL, user_id TEXT NOT NULL, role TEXT NOT NULL, joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`
	if _, err := db.ExecContext(context.Background(), schema); err != nil {
		t.Fatalf("create table: %v", err)
	}

	return &groupMemberChecker{db: db}
}

func TestGroupMemberChecker_GetGroupMembers(t *testing.T) {
	c := setupMemberChecker(t)
	ctx := context.Background()

	seed := `INSERT INTO group_members (group_id, user_id, role) VALUES ('g1','u1','creator'),('g1','u2','member'),('g2','u3','member')`
	if _, err := c.db.ExecContext(ctx, seed); err != nil {
		t.Fatalf("seed: %v", err)
	}

	members, err := c.GetGroupMembers(ctx, "g1")
	if err != nil {
		t.Fatalf("GetGroupMembers() error = %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("len(members) = %d, want 2", len(members))
	}
	got := map[string]bool{}
	for _, m := range members {
		got[m] = true
	}
	if !got["u1"] || !got["u2"] {
		t.Errorf("members = %v, want u1 and u2", members)
	}
}

func TestGroupMemberChecker_IsMember(t *testing.T) {
	c := setupMemberChecker(t)
	ctx := context.Background()

	seed := `INSERT INTO group_members (group_id, user_id, role) VALUES ('g1','u1','creator')`
	if _, err := c.db.ExecContext(ctx, seed); err != nil {
		t.Fatalf("seed: %v", err)
	}

	isMember, err := c.IsMember(ctx, "g1", "u1")
	if err != nil {
		t.Fatalf("IsMember() error = %v", err)
	}
	if !isMember {
		t.Error("IsMember(g1,u1) = false, want true")
	}

	isMember, err = c.IsMember(ctx, "g1", "u9")
	if err != nil {
		t.Fatalf("IsMember() error = %v", err)
	}
	if isMember {
		t.Error("IsMember(g1,u9) = true, want false")
	}
}
