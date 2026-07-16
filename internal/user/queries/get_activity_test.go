package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/user"
)

type mockPostCounter struct {
	count int
	err   error
}

func (m *mockPostCounter) GetPostCount(_ context.Context, _ string) (int, error) {
	return m.count, m.err
}

type mockCommentCounter struct {
	count int
	err   error
}

func (m *mockCommentCounter) GetCommentCount(_ context.Context, _ string) (int, error) {
	return m.count, m.err
}

type mockVoteCounter struct {
	count int
	err   error
}

func (m *mockVoteCounter) GetVoteCount(_ context.Context, _ string) (int, error) {
	return m.count, m.err
}

func TestGetActivityResolver_Success(t *testing.T) {
	u := &user.User{ID: "u1", Nickname: "nick"}
	repo := &mockUserRepoQ{getByIDUser: u}
	pc := &mockPostCounter{count: 10}
	cc := &mockCommentCounter{count: 5}
	vc := &mockVoteCounter{count: 20}
	fc := &mockFollowCounter{followerCount: 3, followingCount: 7}
	r := NewGetActivityResolver(repo, pc, cc, vc, fc)

	result, err := r.Resolve(context.Background(), GetActivityQuery{UserID: "u1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.User.ID != "u1" {
		t.Errorf("User.ID = %q, want %q", result.User.ID, "u1")
	}
	if result.PostCount != 10 {
		t.Errorf("PostCount = %d, want 10", result.PostCount)
	}
	if result.CommentCount != 5 {
		t.Errorf("CommentCount = %d, want 5", result.CommentCount)
	}
	if result.VoteCount != 20 {
		t.Errorf("VoteCount = %d, want 20", result.VoteCount)
	}
	if result.FollowerCount != 3 {
		t.Errorf("FollowerCount = %d, want 3", result.FollowerCount)
	}
	if result.FollowingCount != 7 {
		t.Errorf("FollowingCount = %d, want 7", result.FollowingCount)
	}
}

func TestGetActivityResolver_UserNotFound(t *testing.T) {
	repo := &mockUserRepoQ{getByIDErr: user.ErrUserNotFound}
	r := NewGetActivityResolver(repo, &mockPostCounter{}, &mockCommentCounter{}, &mockVoteCounter{}, &mockFollowCounter{})

	_, err := r.Resolve(context.Background(), GetActivityQuery{UserID: "unknown"})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestGetActivityResolver_PostCountError(t *testing.T) {
	u := &user.User{ID: "u1"}
	repo := &mockUserRepoQ{getByIDUser: u}
	pc := &mockPostCounter{err: errors.New("db down")}
	r := NewGetActivityResolver(repo, pc, &mockCommentCounter{}, &mockVoteCounter{}, &mockFollowCounter{})

	_, err := r.Resolve(context.Background(), GetActivityQuery{UserID: "u1"})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestGetActivityResolver_CommentCountError(t *testing.T) {
	u := &user.User{ID: "u1"}
	repo := &mockUserRepoQ{getByIDUser: u}
	cc := &mockCommentCounter{err: errors.New("db down")}
	r := NewGetActivityResolver(repo, &mockPostCounter{}, cc, &mockVoteCounter{}, &mockFollowCounter{})

	_, err := r.Resolve(context.Background(), GetActivityQuery{UserID: "u1"})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestGetActivityResolver_VoteCountError(t *testing.T) {
	u := &user.User{ID: "u1"}
	repo := &mockUserRepoQ{getByIDUser: u}
	vc := &mockVoteCounter{err: errors.New("db down")}
	r := NewGetActivityResolver(repo, &mockPostCounter{}, &mockCommentCounter{}, vc, &mockFollowCounter{})

	_, err := r.Resolve(context.Background(), GetActivityQuery{UserID: "u1"})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestGetActivityResolver_FollowerCountError(t *testing.T) {
	u := &user.User{ID: "u1"}
	repo := &mockUserRepoQ{getByIDUser: u}
	fc := &mockFollowCounter{followerErr: errors.New("db down")}
	r := NewGetActivityResolver(repo, &mockPostCounter{}, &mockCommentCounter{}, &mockVoteCounter{}, fc)

	_, err := r.Resolve(context.Background(), GetActivityQuery{UserID: "u1"})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestGetActivityResolver_FollowingCountError(t *testing.T) {
	u := &user.User{ID: "u1"}
	repo := &mockUserRepoQ{getByIDUser: u}
	fc := &mockFollowCounter{followingErr: errors.New("db down")}
	r := NewGetActivityResolver(repo, &mockPostCounter{}, &mockCommentCounter{}, &mockVoteCounter{}, fc)

	_, err := r.Resolve(context.Background(), GetActivityQuery{UserID: "u1"})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestNewGetActivityResolver(t *testing.T) {
	r := NewGetActivityResolver(&mockUserRepoQ{}, &mockPostCounter{}, &mockCommentCounter{}, &mockVoteCounter{}, &mockFollowCounter{})
	if r == nil {
		t.Fatal("NewGetActivityResolver() returned nil")
	}
}
