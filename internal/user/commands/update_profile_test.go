package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

type mockBus struct{}

func (m *mockBus) Publish(_, _ string, _ []byte) error { return nil }
func (m *mockBus) Subscribe(_ context.Context, _ string) (<-chan eventbus.Message, error) {
	ch := make(chan eventbus.Message)
	close(ch)
	return ch, nil
}
func (m *mockBus) InitTopology(_ context.Context) error { return nil }

func TestUpdateProfileHandler_Success(t *testing.T) {
	existing := &user.User{ID: "u1", FirstName: "Old", LastName: "Name"}
	repo := &mockUserRepo{getByIDUser: existing}
	h := NewUpdateProfileHandler(repo, &mockBus{})

	err := h.Execute(context.Background(), UpdateProfileCommand{
		UserID:    "u1",
		FirstName: "New",
		LastName:  "Name",
		Nickname:  "newnick",
		AboutMe:   "hello",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestUpdateProfileHandler_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{getByIDErr: user.ErrUserNotFound}
	h := NewUpdateProfileHandler(repo, &mockBus{})

	err := h.Execute(context.Background(), UpdateProfileCommand{
		UserID:    "unknown",
		FirstName: "New",
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestUpdateProfileHandler_UpdateFailure(t *testing.T) {
	existing := &user.User{ID: "u1"}
	repo := &mockUserRepo{getByIDUser: existing, updateErr: errors.New("db failed")}
	h := NewUpdateProfileHandler(repo, &mockBus{})

	err := h.Execute(context.Background(), UpdateProfileCommand{
		UserID:    "u1",
		FirstName: "New",
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestUpdateProfileHandler_PersistsDOBAndGender(t *testing.T) {
	existing := &user.User{ID: "u1"}
	repo := &mockUserRepo{getByIDUser: existing}
	h := NewUpdateProfileHandler(repo, &mockBus{})
	dob := time.Date(2000, 5, 17, 0, 0, 0, 0, time.UTC)

	err := h.Execute(context.Background(), UpdateProfileCommand{
		UserID:      "u1",
		FirstName:   "New",
		Nickname:    "nick",
		DateOfBirth: dob,
		Gender:      "female",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	updated := repo.updatedUser
	if updated == nil {
		t.Fatal("Update() was never called")
	}
	if !updated.DateOfBirth.Equal(dob) {
		t.Errorf("DateOfBirth = %v, want %v", updated.DateOfBirth, dob)
	}
	if updated.Gender != "female" {
		t.Errorf("Gender = %q, want %q", updated.Gender, "female")
	}
}

func TestUpdateProfileHandler_EmptyDOBAndGenderKeptEmpty(t *testing.T) {
	existing := &user.User{ID: "u1", DateOfBirth: time.Date(1999, 1, 2, 0, 0, 0, 0, time.UTC), Gender: "male"}
	repo := &mockUserRepo{getByIDUser: existing}
	h := NewUpdateProfileHandler(repo, &mockBus{})

	err := h.Execute(context.Background(), UpdateProfileCommand{
		UserID:   "u1",
		Nickname: "nick",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	updated := repo.updatedUser
	if !updated.DateOfBirth.IsZero() {
		t.Errorf("DateOfBirth = %v, want zero (empty input clears)", updated.DateOfBirth)
	}
	if updated.Gender != "" {
		t.Errorf("Gender = %q, want empty (empty input clears)", updated.Gender)
	}
}

func TestUpdateProfileHandler_InvalidGender(t *testing.T) {
	repo := &mockUserRepo{getByIDUser: &user.User{ID: "u1"}}
	h := NewUpdateProfileHandler(repo, &mockBus{})

	err := h.Execute(context.Background(), UpdateProfileCommand{
		UserID:   "u1",
		Gender:   "banana",
		Nickname: "nick",
	})
	if !errors.Is(err, ErrInvalidGender) {
		t.Fatalf("Execute() error = %v, want ErrInvalidGender", err)
	}
}

func TestNewUpdateProfileHandler(t *testing.T) {
	h := NewUpdateProfileHandler(&mockUserRepo{}, &mockBus{})
	if h == nil {
		t.Fatal("NewUpdateProfileHandler() returned nil")
	}
}
