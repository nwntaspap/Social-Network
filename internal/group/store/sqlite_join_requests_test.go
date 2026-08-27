package store

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
	"social-network/internal/platform/database"
)

const joinRequestSchema = `
CREATE TABLE group_join_requests (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    requester_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

func setupJoinRequestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.ExecContext(context.Background(), groupsSchema); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), joinRequestSchema); err != nil {
		t.Fatalf("create table: %v", err)
	}
	return NewSQLiteStore(db)
}

func TestGetPendingJoinRequests_ReturnsGroupRows(t *testing.T) {
	s := setupJoinRequestStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u1"})
	if err := s.CreateJoinRequest(ctx, &group.JoinRequest{ID: "jr1", GroupID: "g1", RequesterID: "u3"}); err != nil {
		t.Fatalf("CreateJoinRequest: %v", err)
	}
	if err := s.CreateJoinRequest(ctx, &group.JoinRequest{ID: "jr2", GroupID: "g1", RequesterID: "u4"}); err != nil {
		t.Fatalf("CreateJoinRequest: %v", err)
	}

	requests, err := s.GetPendingJoinRequests(ctx, "g1")
	if err != nil {
		t.Fatalf("GetPendingJoinRequests: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("len = %d, want 2", len(requests))
	}
	if requests[0].ID != "jr1" || requests[1].ID != "jr2" {
		t.Errorf("requests = %+v, want jr1 then jr2", requests)
	}
}

func TestGetPendingJoinRequests_Empty(t *testing.T) {
	s := setupJoinRequestStore(t)
	ctx := context.Background()

	requests, err := s.GetPendingJoinRequests(ctx, "g1")
	if err != nil {
		t.Fatalf("GetPendingJoinRequests: %v", err)
	}
	if len(requests) != 0 {
		t.Errorf("len = %d, want 0", len(requests))
	}
}

func TestDeleteJoinRequest_RemovesRow(t *testing.T) {
	s := setupJoinRequestStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u1"})
	if err := s.CreateJoinRequest(ctx, &group.JoinRequest{ID: "jr1", GroupID: "g1", RequesterID: "u3"}); err != nil {
		t.Fatalf("CreateJoinRequest: %v", err)
	}

	if err := s.DeleteJoinRequest(ctx, "g1", "u3"); err != nil {
		t.Fatalf("DeleteJoinRequest: %v", err)
	}

	if _, err := s.GetJoinRequest(ctx, "g1", "u3"); !errors.Is(err, group.ErrJoinRequestNotFound) {
		t.Errorf("expected ErrJoinRequestNotFound after delete, got %v", err)
	}
}
