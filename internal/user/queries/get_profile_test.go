package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/user"
)

type mockUserRepoQ struct {
	getByIDUser *user.User
	getByIDErr  error
}

func (m *mockUserRepoQ) Create(_ context.Context, _ *user.User) error { return nil }
func (m *mockUserRepoQ) GetByID(_ context.Context, _ string) (*user.User, error) {
	return m.getByIDUser, m.getByIDErr
}

func (m *mockUserRepoQ) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}

func (m *mockUserRepoQ) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}
func (m *mockUserRepoQ) Update(_ context.Context, _ *user.User) error            { return nil }
func (m *mockUserRepoQ) TogglePrivacy(_ context.Context, _ string, _ bool) error { return nil }
func (m *mockUserRepoQ) ListAll(_ context.Context) ([]user.User, error)          { return nil, nil }

type mockFollowChecker struct {
	isFollowing bool
	err         error
}

func (m *mockFollowChecker) IsFollowing(_ context.Context, _, _ string) (bool, error) {
	return m.isFollowing, m.err
}

type mockFollowCounter struct {
	followerCount  int
	followingCount int
	followerErr    error
	followingErr   error
}

func (m *mockFollowCounter) GetFollowerCount(_ context.Context, _ string) (int, error) {
	return m.followerCount, m.followerErr
}

func (m *mockFollowCounter) GetFollowingCount(_ context.Context, _ string) (int, error) {
	return m.followingCount, m.followingErr
}

func TestGetProfileResolver_PublicProfile(t *testing.T) {
	u := &user.User{ID: "u1", Nickname: "nick", AboutMe: "hello"}
	repo := &mockUserRepoQ{getByIDUser: u}
	fc := &mockFollowChecker{isFollowing: true}
	counter := &mockFollowCounter{followerCount: 5, followingCount: 3}
	r := NewGetProfileResolver(repo, fc, counter)

	result, err := r.Resolve(context.Background(), GetProfileQuery{
		TargetID:    "u1",
		RequesterID: "u2",
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.User.ID != "u1" {
		t.Errorf("User.ID = %q, want %q", result.User.ID, "u1")
	}
	if result.FollowerCount != 5 {
		t.Errorf("FollowerCount = %d, want 5", result.FollowerCount)
	}
	if result.FollowingCount != 3 {
		t.Errorf("FollowingCount = %d, want 3", result.FollowingCount)
	}
	if !result.IsFollowing {
		t.Error("IsFollowing = false, want true")
	}
}

func TestGetProfileResolver_PrivateProfile_Following(t *testing.T) {
	u := &user.User{ID: "u1", Nickname: "nick", IsPrivate: true}
	repo := &mockUserRepoQ{getByIDUser: u}
	fc := &mockFollowChecker{isFollowing: true}
	counter := &mockFollowCounter{followerCount: 2, followingCount: 1}
	r := NewGetProfileResolver(repo, fc, counter)

	result, err := r.Resolve(context.Background(), GetProfileQuery{
		TargetID:    "u1",
		RequesterID: "u2",
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.User.Nickname != "nick" {
		t.Errorf("User.Nickname = %q, want %q", result.User.Nickname, "nick")
	}
	if result.FollowerCount != 2 {
		t.Errorf("FollowerCount = %d, want 2", result.FollowerCount)
	}
	if !result.IsFollowing {
		t.Error("IsFollowing = false, want true (private profile, following)")
	}
}

func TestGetProfileResolver_PrivateProfile_NotFollowing(t *testing.T) {
	u := &user.User{ID: "u1", Nickname: "nick", IsPrivate: true, AboutMe: "secret"}
	repo := &mockUserRepoQ{getByIDUser: u}
	fc := &mockFollowChecker{isFollowing: false}
	r := NewGetProfileResolver(repo, fc, &mockFollowCounter{})

	result, err := r.Resolve(context.Background(), GetProfileQuery{
		TargetID:    "u1",
		RequesterID: "u2",
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.User.Nickname != "nick" {
		t.Errorf("User.Nickname = %q, want %q", result.User.Nickname, "nick")
	}
	if result.User.AboutMe != "" {
		t.Errorf("User.AboutMe = %q, want empty (private, not following)", result.User.AboutMe)
	}
	if result.FollowerCount != 0 {
		t.Errorf("FollowerCount = %d, want 0 (limited profile)", result.FollowerCount)
	}
}

func TestGetProfileResolver_OwnProfile(t *testing.T) {
	u := &user.User{ID: "u1", Nickname: "nick", IsPrivate: true}
	repo := &mockUserRepoQ{getByIDUser: u}
	r := NewGetProfileResolver(repo, &mockFollowChecker{}, &mockFollowCounter{followerCount: 10, followingCount: 5})

	result, err := r.Resolve(context.Background(), GetProfileQuery{
		TargetID:    "u1",
		RequesterID: "u1",
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.FollowerCount != 10 {
		t.Errorf("FollowerCount = %d, want 10 (own profile, full access)", result.FollowerCount)
	}
	if result.IsFollowing {
		t.Error("IsFollowing = true, want false (own profile)")
	}
}

func TestGetProfileResolver_UserNotFound(t *testing.T) {
	repo := &mockUserRepoQ{getByIDErr: user.ErrUserNotFound}
	r := NewGetProfileResolver(repo, &mockFollowChecker{}, &mockFollowCounter{})

	_, err := r.Resolve(context.Background(), GetProfileQuery{
		TargetID:    "unknown",
		RequesterID: "u2",
	})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestGetProfileResolver_FollowCheckError(t *testing.T) {
	u := &user.User{ID: "u1", IsPrivate: true}
	repo := &mockUserRepoQ{getByIDUser: u}
	fc := &mockFollowChecker{err: errors.New("db down")}
	r := NewGetProfileResolver(repo, fc, &mockFollowCounter{})

	_, err := r.Resolve(context.Background(), GetProfileQuery{
		TargetID:    "u1",
		RequesterID: "u2",
	})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestGetProfileResolver_FollowerCountError(t *testing.T) {
	u := &user.User{ID: "u1"}
	repo := &mockUserRepoQ{getByIDUser: u}
	counter := &mockFollowCounter{followerErr: errors.New("db down")}
	r := NewGetProfileResolver(repo, &mockFollowChecker{}, counter)

	_, err := r.Resolve(context.Background(), GetProfileQuery{
		TargetID:    "u1",
		RequesterID: "u2",
	})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestNewGetProfileResolver(t *testing.T) {
	r := NewGetProfileResolver(&mockUserRepoQ{}, &mockFollowChecker{}, &mockFollowCounter{})
	if r == nil {
		t.Fatal("NewGetProfileResolver() returned nil")
	}
}
