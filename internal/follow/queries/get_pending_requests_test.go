package queries

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/follow"
)

func TestGetPendingRequestsResolver_Success(t *testing.T) {
	now := time.Now()
	expected := []follow.Request{
		{FollowerID: "user-2", FolloweeID: "user-1", CreatedAt: now},
		{FollowerID: "user-3", FolloweeID: "user-1", CreatedAt: now},
	}
	repo := &mockRepo{getPendingResult: expected}
	r := NewGetPendingRequestsResolver(repo)

	result, err := r.Resolve(context.Background(), GetPendingRequestsQuery{UserID: "user-1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].FollowerID != "user-2" || result[1].FollowerID != "user-3" {
		t.Errorf("result = %+v, want FollowerID=user-2, user-3", result)
	}
}

func TestGetPendingRequestsResolver_Empty(t *testing.T) {
	repo := &mockRepo{getPendingResult: []follow.Request{}}
	r := NewGetPendingRequestsResolver(repo)

	result, err := r.Resolve(context.Background(), GetPendingRequestsQuery{UserID: "user-1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("len(result) = %d, want 0", len(result))
	}
}

func TestGetPendingRequestsResolver_Error(t *testing.T) {
	repo := &mockRepo{getPendingErr: errors.New("db down")}
	r := NewGetPendingRequestsResolver(repo)

	_, err := r.Resolve(context.Background(), GetPendingRequestsQuery{UserID: "user-1"})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}
