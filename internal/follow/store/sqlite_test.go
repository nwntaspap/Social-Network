package store

import (
	"context"
	"testing"

	"social-network/internal/follow"
	"social-network/internal/platform/database"
)

func setupStore(t *testing.T) *SQLiteStore {
	t.Helper()

	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS follows (
			follower_id TEXT NOT NULL,
			followee_id TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(follower_id, followee_id)
		);
		CREATE TABLE IF NOT EXISTS follow_requests (
			follower_id TEXT NOT NULL,
			followee_id TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(follower_id, followee_id)
		);`)
	if err != nil {
		t.Fatalf("create tables: %v", err)
	}

	return NewSQLiteStore(db)
}

func TestSQLiteStore_CreateFollow(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	f := &follow.Follow{FollowerID: "user-1", FolloweeID: "user-2"}
	if err := s.CreateFollow(ctx, f); err != nil {
		t.Fatalf("CreateFollow() error = %v", err)
	}

	connected, err := s.AreConnected(ctx, "user-1", "user-2")
	if err != nil {
		t.Fatalf("AreConnected() error = %v", err)
	}
	if !connected {
		t.Error("AreConnected() = false, want true after CreateFollow")
	}
}

func TestSQLiteStore_DeleteFollow(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	f := &follow.Follow{FollowerID: "user-1", FolloweeID: "user-2"}
	if err := s.CreateFollow(ctx, f); err != nil {
		t.Fatalf("CreateFollow() error = %v", err)
	}

	if err := s.DeleteFollow(ctx, "user-1", "user-2"); err != nil {
		t.Fatalf("DeleteFollow() error = %v", err)
	}

	connected, err := s.AreConnected(ctx, "user-1", "user-2")
	if err != nil {
		t.Fatalf("AreConnected() error = %v", err)
	}
	if connected {
		t.Error("AreConnected() = true, want false after DeleteFollow")
	}
}

func TestSQLiteStore_GetFollowers(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	if err := s.CreateFollow(ctx, &follow.Follow{FollowerID: "user-1", FolloweeID: "user-2"}); err != nil {
		t.Fatalf("CreateFollow(user-1→user-2) error = %v", err)
	}
	if err := s.CreateFollow(ctx, &follow.Follow{FollowerID: "user-3", FolloweeID: "user-2"}); err != nil {
		t.Fatalf("CreateFollow(user-3→user-2) error = %v", err)
	}

	followers, err := s.GetFollowers(ctx, "user-2")
	if err != nil {
		t.Fatalf("GetFollowers() error = %v", err)
	}
	if len(followers) != 2 {
		t.Fatalf("GetFollowers() returned %d followers, want 2", len(followers))
	}
}

func TestSQLiteStore_GetFollowing(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	if err := s.CreateFollow(ctx, &follow.Follow{FollowerID: "user-1", FolloweeID: "user-2"}); err != nil {
		t.Fatalf("CreateFollow(user-1→user-2) error = %v", err)
	}
	if err := s.CreateFollow(ctx, &follow.Follow{FollowerID: "user-1", FolloweeID: "user-3"}); err != nil {
		t.Fatalf("CreateFollow(user-1→user-3) error = %v", err)
	}

	following, err := s.GetFollowing(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetFollowing() error = %v", err)
	}
	if len(following) != 2 {
		t.Fatalf("GetFollowing() returned %d following, want 2", len(following))
	}
}

func TestSQLiteStore_CreateFollowRequest(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	req := &follow.Request{FollowerID: "user-1", FolloweeID: "user-2"}
	if err := s.CreateFollowRequest(ctx, req); err != nil {
		t.Fatalf("CreateFollowRequest() error = %v", err)
	}

	requests, err := s.GetPendingRequests(ctx, "user-2")
	if err != nil {
		t.Fatalf("GetPendingRequests() error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("GetPendingRequests() returned %d, want 1", len(requests))
	}
}

func TestSQLiteStore_CreateFollowRequest_Idempotent(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	req := &follow.Request{FollowerID: "user-1", FolloweeID: "user-2"}
	if err := s.CreateFollowRequest(ctx, req); err != nil {
		t.Fatalf("CreateFollowRequest() first call error = %v", err)
	}
	if err := s.CreateFollowRequest(ctx, req); err != nil {
		t.Fatalf("CreateFollowRequest() duplicate call error = %v, want nil", err)
	}

	requests, err := s.GetPendingRequests(ctx, "user-2")
	if err != nil {
		t.Fatalf("GetPendingRequests() error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("GetPendingRequests() returned %d, want 1 (no duplicate rows)", len(requests))
	}
}

func TestSQLiteStore_CreateFollow_Idempotent(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	f := &follow.Follow{FollowerID: "user-1", FolloweeID: "user-2"}
	if err := s.CreateFollow(ctx, f); err != nil {
		t.Fatalf("CreateFollow() first call error = %v", err)
	}
	if err := s.CreateFollow(ctx, f); err != nil {
		t.Fatalf("CreateFollow() duplicate call error = %v, want nil", err)
	}

	following, err := s.GetFollowing(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetFollowing() error = %v", err)
	}
	if len(following) != 1 {
		t.Fatalf("GetFollowing() returned %d, want 1 (no duplicate rows)", len(following))
	}
}

func TestSQLiteStore_DeleteFollowRequest(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	req := &follow.Request{FollowerID: "user-1", FolloweeID: "user-2"}
	if err := s.CreateFollowRequest(ctx, req); err != nil {
		t.Fatalf("CreateFollowRequest() error = %v", err)
	}

	if err := s.DeleteFollowRequest(ctx, "user-1", "user-2"); err != nil {
		t.Fatalf("DeleteFollowRequest() error = %v", err)
	}

	requests, err := s.GetPendingRequests(ctx, "user-2")
	if err != nil {
		t.Fatalf("GetPendingRequests() error = %v", err)
	}
	if len(requests) != 0 {
		t.Fatalf("GetPendingRequests() returned %d, want 0 after delete", len(requests))
	}
}

func TestSQLiteStore_AreConnected(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	connected, err := s.AreConnected(ctx, "user-1", "user-2")
	if err != nil {
		t.Fatalf("AreConnected() error = %v", err)
	}
	if connected {
		t.Error("AreConnected() = true, want false for unconnected users")
	}

	if err = s.CreateFollow(ctx, &follow.Follow{FollowerID: "user-1", FolloweeID: "user-2"}); err != nil {
		t.Fatalf("CreateFollow() error = %v", err)
	}

	connected, err = s.AreConnected(ctx, "user-1", "user-2")
	if err != nil {
		t.Fatalf("AreConnected() error = %v", err)
	}
	if !connected {
		t.Error("AreConnected() = false, want true for connected users")
	}
}
