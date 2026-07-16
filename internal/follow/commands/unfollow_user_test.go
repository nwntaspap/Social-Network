package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/follow"
)

type unfollowMockRepo struct {
	deleteFollowErr error
	deletedFollower string
	deletedFollowee string
}

func (m *unfollowMockRepo) CreateFollow(_ context.Context, _ *follow.Follow) error { return nil }
func (m *unfollowMockRepo) DeleteFollow(_ context.Context, followerID, followeeID string) error {
	m.deletedFollower = followerID
	m.deletedFollowee = followeeID
	return m.deleteFollowErr
}

func (m *unfollowMockRepo) GetFollowers(_ context.Context, _ string) ([]follow.Follow, error) {
	return nil, nil
}

func (m *unfollowMockRepo) GetFollowing(_ context.Context, _ string) ([]follow.Follow, error) {
	return nil, nil
}

func (m *unfollowMockRepo) CreateFollowRequest(_ context.Context, _ *follow.Request) error {
	return nil
}

func (m *unfollowMockRepo) DeleteFollowRequest(_ context.Context, _, _ string) error {
	return nil
}

func (m *unfollowMockRepo) GetPendingRequests(_ context.Context, _ string) ([]follow.Request, error) {
	return nil, nil
}

func (m *unfollowMockRepo) AreConnected(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func TestUnfollowUserHandler_Success(t *testing.T) {
	repo := &unfollowMockRepo{}
	h := NewUnfollowUserHandler(repo)

	err := h.Execute(context.Background(), UnfollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-2",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if repo.deletedFollower != "user-1" {
		t.Errorf("deletedFollower = %q, want %q", repo.deletedFollower, "user-1")
	}
	if repo.deletedFollowee != "user-2" {
		t.Errorf("deletedFollowee = %q, want %q", repo.deletedFollowee, "user-2")
	}
}

func TestUnfollowUserHandler_SelfUnfollow(t *testing.T) {
	repo := &unfollowMockRepo{}
	h := NewUnfollowUserHandler(repo)

	err := h.Execute(context.Background(), UnfollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-1",
	})
	if !errors.Is(err, ErrCannotUnfollowSelf) {
		t.Errorf("Execute() error = %v, want %v", err, ErrCannotUnfollowSelf)
	}
	if repo.deletedFollower != "" {
		t.Error("DeleteFollow called for self-unfollow")
	}
}

func TestUnfollowUserHandler_RepoError(t *testing.T) {
	repo := &unfollowMockRepo{deleteFollowErr: errors.New("db error")}
	h := NewUnfollowUserHandler(repo)

	err := h.Execute(context.Background(), UnfollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-2",
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}
