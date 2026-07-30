package commands

import (
	"context"
	"errors"
	"testing"

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

func TestNewUpdateProfileHandler(t *testing.T) {
	h := NewUpdateProfileHandler(&mockUserRepo{}, &mockBus{})
	if h == nil {
		t.Fatal("NewUpdateProfileHandler() returned nil")
	}
}
