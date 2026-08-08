package queries

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/user"
)

type mockListUsersRepo struct {
	users  []user.User
	count  int
	err    error
	offset int
	limit  int
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

func (m *mockListUsersRepo) SearchUsers(_ context.Context, _ string, limit int, offset int) ([]user.User, error) {
	m.limit = limit
	m.offset = offset
	return m.users, m.err
}

func (m *mockListUsersRepo) CountUsers(_ context.Context, _ string) (int, error) {
	return m.count, m.err
}

func TestListUsersResolver_Success(t *testing.T) {
	now := time.Now()
	users := []user.User{
		{ID: "u1", Nickname: "alice", Email: "a@b.com", PasswordHash: "secret", CreatedAt: now},
		{ID: "u2", Nickname: "bob", Email: "c@d.com", PasswordHash: "secret", CreatedAt: now},
	}
	repo := &mockListUsersRepo{users: users, count: 2}
	r := NewListUsersResolver(repo, nil)

	result, err := r.Resolve(context.Background(), ListUsersQuery{Page: 1, Limit: 10})
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
	if result.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Total)
	}
}

func TestListUsersResolver_PaginatesOffset(t *testing.T) {
	repo := &mockListUsersRepo{count: 25}
	r := NewListUsersResolver(repo, nil)

	_, err := r.Resolve(context.Background(), ListUsersQuery{Page: 3, Limit: 10})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if repo.offset != 20 {
		t.Errorf("offset = %d, want 20", repo.offset)
	}
	if repo.limit != 10 {
		t.Errorf("limit = %d, want 10", repo.limit)
	}
}

func TestListUsersResolver_PreservesUserFields(t *testing.T) {
	dob := time.Date(1995, 3, 2, 0, 0, 0, 0, time.UTC)
	users := []user.User{{
		ID: "u1", Nickname: "alice", Email: "a@b.com", PasswordHash: "secret",
		FirstName: "Alice", LastName: "Smith", AboutMe: "hi", IsPrivate: true,
		DateOfBirth: dob, AvatarPath: "/img.png", CreatedAt: dob,
	}}
	repo := &mockListUsersRepo{users: users, count: 1}
	r := NewListUsersResolver(repo, nil)

	result, err := r.Resolve(context.Background(), ListUsersQuery{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	u := result.Users[0]
	if u.DateOfBirth.IsZero() || u.DateOfBirth.Year() != 1995 {
		t.Errorf("DateOfBirth not preserved: %v", u.DateOfBirth)
	}
	if u.AboutMe != "hi" {
		t.Errorf("AboutMe = %q, want %q", u.AboutMe, "hi")
	}
	if !u.IsPrivate {
		t.Error("IsPrivate = false, want true")
	}
}

func TestListUsersResolver_IsOnlineFlags(t *testing.T) {
	users := []user.User{
		{ID: "u1", Nickname: "alice"},
		{ID: "u2", Nickname: "bob"},
		{ID: "u3", Nickname: "carol"},
	}
	repo := &mockListUsersRepo{users: users, count: 3}
	online := map[string]bool{"u1": true, "u3": true}
	r := NewListUsersResolver(repo, func(id string) bool { return online[id] })

	result, err := r.Resolve(context.Background(), ListUsersQuery{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	flags := map[string]bool{}
	for _, u := range result.Users {
		flags[u.ID] = u.IsOnline
	}
	if !flags["u1"] {
		t.Error("u1 should be marked online")
	}
	if flags["u2"] {
		t.Error("u2 should not be marked online")
	}
	if !flags["u3"] {
		t.Error("u3 should be marked online")
	}
}

func TestListUsersResolver_IsOnlineNilChecker(t *testing.T) {
	repo := &mockListUsersRepo{users: []user.User{{ID: "u1", Nickname: "alice"}}, count: 1}
	r := NewListUsersResolver(repo, nil)

	result, err := r.Resolve(context.Background(), ListUsersQuery{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.Users[0].IsOnline {
		t.Error("IsOnline should default to false without a checker")
	}
}

func TestListUsersResolver_Empty(t *testing.T) {
	repo := &mockListUsersRepo{users: []user.User{}}
	r := NewListUsersResolver(repo, nil)

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
	r := NewListUsersResolver(repo, nil)

	_, err := r.Resolve(context.Background(), ListUsersQuery{})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestNewListUsersResolver(t *testing.T) {
	r := NewListUsersResolver(&mockListUsersRepo{}, nil)
	if r == nil {
		t.Fatal("NewListUsersResolver() returned nil")
	}
}
