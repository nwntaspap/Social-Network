package commands

import (
	"context"
	"errors"
	"strings"
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
	updatedUser       *user.User
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
	if m.getByUsernameUser != nil {
		return m.getByUsernameUser, m.getByUsernameErr
	}
	if m.getByUsernameErr != nil {
		return nil, m.getByUsernameErr
	}
	// Default mirrors store behaviour: unknown nickname -> ErrUserNotFound.
	return nil, user.ErrUserNotFound
}

func (m *mockUserRepo) Update(_ context.Context, u *user.User) error {
	m.updatedUser = u
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
	h := NewRegisterHandler(repo, uuid, crypto, nil)

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
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("Execute() error = %v, want %v", err, ErrEmailTaken)
	}
}

func TestRegisterHandler_Underage(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-10 * 365.25 * 24 * time.Hour),
	})
	if !errors.Is(err, ErrUnderage) {
		t.Errorf("Execute() error = %v, want %v", err, ErrUnderage)
	}
}

func TestRegisterHandler_WeakPassword(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

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
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	got, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v, want auto-generated nickname", err)
	}
	if got.Nickname == "" {
		t.Fatal("Nickname is empty, want auto-generated value like john_1234")
	}
	if !strings.HasPrefix(got.Nickname, "john_") {
		t.Errorf("Nickname = %q, want prefix %q", got.Nickname, "john_")
	}
}

func TestRegisterHandler_MissingFirstName(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "  ",
		LastName:    "Doe",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if !errors.Is(err, ErrFirstNameMissing) {
		t.Errorf("Execute() error = %v, want %v", err, ErrFirstNameMissing)
	}
}

func TestRegisterHandler_MissingLastName(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if !errors.Is(err, ErrLastNameMissing) {
		t.Errorf("Execute() error = %v, want %v", err, ErrLastNameMissing)
	}
}

func TestRegisterHandler_DuplicateNickname(t *testing.T) {
	existing := &user.User{ID: "existing", Nickname: "johndoe"}
	repo := &mockUserRepo{getByUsernameUser: existing}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "new@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if !errors.Is(err, ErrNicknameTaken) {
		t.Errorf("Execute() error = %v, want %v", err, ErrNicknameTaken)
	}
}

func TestRegisterHandler_AutoGeneratedNicknameAllTaken(t *testing.T) {
	// Every generated candidate already exists -> handler must give up with
	// ErrNicknameTaken instead of silently reusing a taken nickname.
	repo := &mockUserRepo{getByUsernameUser: &user.User{ID: "existing", Nickname: "taken"}}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if !errors.Is(err, ErrNicknameTaken) {
		t.Errorf("Execute() error = %v, want %v", err, ErrNicknameTaken)
	}
}

func TestRegisterHandler_InvalidEmail(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "not-an-email",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
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
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, crypto, nil)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestRegisterHandler_RepoCreateFailure(t *testing.T) {
	repo := &mockUserRepo{createErr: errors.New("db failed")}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestRegisterHandler_GetByEmailError(t *testing.T) {
	repo := &mockUserRepo{getByEmailErr: errors.New("db down")}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestNewRegisterHandler(t *testing.T) {
	h := NewRegisterHandler(&mockUserRepo{}, &mockUUID{id: "id"}, &mockCrypto{hash: "h"}, nil)
	if h == nil {
		t.Fatal("NewRegisterHandler() returned nil")
	}
}

func TestRegisterHandler_InvalidGender(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	_, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
		Gender:      "attack helicopter",
	})
	if !errors.Is(err, ErrInvalidGender) {
		t.Errorf("Execute() error = %v, want %v", err, ErrInvalidGender)
	}
}

func TestRegisterHandler_GenderPersisted(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	got, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
		Gender:      "female",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got.Gender != "female" {
		t.Errorf("Gender = %q, want %q", got.Gender, "female")
	}
}

func TestRegisterHandler_EmptyGenderDefaults(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewRegisterHandler(repo, &mockUUID{id: "new"}, &mockCrypto{hash: "h"}, nil)

	got, err := h.Execute(context.Background(), RegisterCommand{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		Nickname:    "johndoe",
		DateOfBirth: time.Now().Add(-20 * 365.25 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got.Gender != "prefer_not_to_say" {
		t.Errorf("Gender = %q, want default %q", got.Gender, "prefer_not_to_say")
	}
}
