package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/user"
)

type mockUserRepo struct {
	createErr         error
	getByIDUser       *user.User
	getByIDErr        error
	getByEmailUser    *user.User
	getByEmailErr     error
	getByUsernameUser *user.User
	getByUsernameErr  error
	updateErr         error
	togglePrivacyErr  error
}

func (m *mockUserRepo) Create(_ context.Context, _ *user.User) error {
	return m.createErr
}

func (m *mockUserRepo) GetByID(_ context.Context, _ string) (*user.User, error) {
	return m.getByIDUser, m.getByIDErr
}

func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return m.getByEmailUser, m.getByEmailErr
}

func (m *mockUserRepo) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return m.getByUsernameUser, m.getByUsernameErr
}

func (m *mockUserRepo) Update(_ context.Context, _ *user.User) error {
	return m.updateErr
}

func (m *mockUserRepo) TogglePrivacy(_ context.Context, _ string, _ bool) error {
	return m.togglePrivacyErr
}

func (m *mockUserRepo) ListAll(_ context.Context) ([]user.User, error) {
	return nil, nil
}

type mockUUID struct {
	id string
}

func (m *mockUUID) NewUUID() string {
	return m.id
}

type mockCrypto struct {
	hash string
	err  error
}

func (m *mockCrypto) Generate(_ string) (string, error) {
	return m.hash, m.err
}

func (m *mockCrypto) Matches(_, _ string) error {
	return nil
}

func TestRegisterHandler_Success(t *testing.T) {
	repo := &mockUserRepo{}
	uuid := &mockUUID{id: "test-uuid"}
	crypto := &mockCrypto{hash: "hashed_pass"}
	h := NewRegisterHandler(repo, uuid, crypto)

	cmd := RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	}

	got, err := h.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got.ID != "test-uuid" {
		t.Errorf("ID = %q, want %q", got.ID, "test-uuid")
	}
	if got.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", got.Email, "test@example.com")
	}
	if got.PasswordHash != "hashed_pass" {
		t.Errorf("PasswordHash = %q, want %q", got.PasswordHash, "hashed_pass")
	}
	if got.Nickname != "johndoe" {
		t.Errorf("Nickname = %q, want %q", got.Nickname, "johndoe")
	}
	if got.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", got.FirstName, "John")
	}
	if got.LastName != "Doe" {
		t.Errorf("LastName = %q, want %q", got.LastName, "Doe")
	}
}

func TestRegisterHandler_DuplicateEmail(t *testing.T) {
	existing := &user.User{ID: "existing", Email: "test@example.com"}
	repo := &mockUserRepo{getByEmailUser: existing}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"})

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("Execute() error = %v, want %v", err, ErrEmailTaken)
	}
}

func TestRegisterHandler_Underage(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"})

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-10 * 365.25 * 24 * time.Hour),
	})
	if !errors.Is(err, ErrUnderage) {
		t.Errorf("Execute() error = %v, want %v", err, ErrUnderage)
	}
}

func TestRegisterHandler_WeakPassword(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"})

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "short",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if !errors.Is(err, ErrWeakPassword) {
		t.Errorf("Execute() error = %v, want %v", err, ErrWeakPassword)
	}
}

func TestRegisterHandler_EmptyNickname(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"})

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		Nickname:    "",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if !errors.Is(err, ErrNicknameEmpty) {
		t.Errorf("Execute() error = %v, want %v", err, ErrNicknameEmpty)
	}
}

func TestRegisterHandler_InvalidEmail(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"})

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "not-an-email",
		Password:    "password123",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestRegisterHandler_EncryptionFailure(t *testing.T) {
	repo := &mockUserRepo{}
	crypto := &mockCrypto{err: errors.New("crypto failed")}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, crypto)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestRegisterHandler_RepoCreateFailure(t *testing.T) {
	repo := &mockUserRepo{createErr: errors.New("db failed")}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"})

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestRegisterHandler_GetByEmailError(t *testing.T) {
	repo := &mockUserRepo{getByEmailErr: errors.New("db down")}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"})

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestNewRegisterHandler(t *testing.T) {
	h := NewRegisterHandler(&mockUserRepo{}, &mockUUID{id: "id"}, &mockCrypto{hash: "h"})
	if h == nil {
		t.Fatal("NewRegisterHandler() returned nil")
	}
}
