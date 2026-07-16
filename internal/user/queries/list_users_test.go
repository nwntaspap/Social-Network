package queries

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/user"
)

type mockListUsersRepo struct {
	users []user.User
	err   error
}

func (m *mockListUsersRepo) Create(_ context.Context, _ *user.User) error { return nil }
func (m *mockListUsersRepo) GetByID(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}

func (m *mockListUsersRepo) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}

func (m *mockListUsersRepo) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}
func (m *mockListUsersRepo) Update(_ context.Context, _ *user.User) error { return nil }
func (m *mockListUsersRepo) TogglePrivacy(_ context.Context, _ string, _ bool) error {
	return nil
}

func (m *mockListUsersRepo) ListAll(_ context.Context) ([]user.User, error) {
	return m.users, m.err
}

func TestListUsersResolver_Success(t *testing.T) {
	now := time.Now()
	users := []user.User{
		{ID: "u1", Nickname: "alice", Email: "a@b.com", PasswordHash: "secret", CreatedAt: now},
		{ID: "u2", Nickname: "bob", Email: "c@d.com", PasswordHash: "secret", CreatedAt: now},
	}
	repo := &mockListUsersRepo{users: users}
	r := NewListUsersResolver(repo)

	result, err := r.Resolve(context.Background(), ListUsersQuery{})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result.Users) != 2 {
		t.Fatalf("len(Users) = %d, want 2", len(result.Users))
	}
	for _, u := range result.Users {
		if u.PasswordHash != "" {
			t.Errorf("PasswordHash should be stripped, got %q for user %s", u.PasswordHash, u.ID)
		}
	}
	if result.Users[0].Nickname != "alice" {
		t.Errorf("Users[0].Nickname = %q, want %q", result.Users[0].Nickname, "alice")
	}
}

func TestListUsersResolver_Empty(t *testing.T) {
	repo := &mockListUsersRepo{users: []user.User{}}
	r := NewListUsersResolver(repo)

	result, err := r.Resolve(context.Background(), ListUsersQuery{})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result.Users) != 0 {
		t.Errorf("len(Users) = %d, want 0", len(result.Users))
	}
}

func TestListUsersResolver_Error(t *testing.T) {
	repo := &mockListUsersRepo{err: errors.New("db down")}
	r := NewListUsersResolver(repo)

	_, err := r.Resolve(context.Background(), ListUsersQuery{})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestNewListUsersResolver(t *testing.T) {
	r := NewListUsersResolver(&mockListUsersRepo{})
	if r == nil {
		t.Fatal("NewListUsersResolver() returned nil")
	}
}
