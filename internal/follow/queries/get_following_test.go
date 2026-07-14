package queries

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/follow"
)

func TestGetFollowingResolver_Success(t *testing.T) {
	now := time.Now()
	expected := []follow.Follow{
		{FollowerID: "user-1", FolloweeID: "user-2", CreatedAt: now},
		{FollowerID: "user-1", FolloweeID: "user-3", CreatedAt: now},
	}
	repo := &mockRepo{getFollowingResult: expected}
	r := NewGetFollowingResolver(repo)

	result, err := r.Resolve(context.Background(), GetFollowingQuery{UserID: "user-1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].FolloweeID != "user-2" || result[1].FolloweeID != "user-3" {
		t.Errorf("result = %+v, want FolloweeID=user-2, user-3", result)
	}
}

func TestGetFollowingResolver_Empty(t *testing.T) {
	repo := &mockRepo{getFollowingResult: []follow.Follow{}}
	r := NewGetFollowingResolver(repo)

	result, err := r.Resolve(context.Background(), GetFollowingQuery{UserID: "user-1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("len(result) = %d, want 0", len(result))
	}
}

func TestGetFollowingResolver_Error(t *testing.T) {
	repo := &mockRepo{getFollowingErr: errors.New("db down")}
	r := NewGetFollowingResolver(repo)

	_, err := r.Resolve(context.Background(), GetFollowingQuery{UserID: "user-1"})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}
