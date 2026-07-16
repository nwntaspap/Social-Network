package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/core/session"
	"social-network/internal/user"
)

type mockSessionManager struct {
	createSession *session.Session
	createErr     error
	revokeErr     error
}

func (m *mockSessionManager) Create(_ context.Context, _ string) (*session.Session, error) {
	return m.createSession, m.createErr
}

func (m *mockSessionManager) Get(_ context.Context, _ string) (*session.Session, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSessionManager) Revoke(_ context.Context, _ string) error {
	return m.revokeErr
}

type mockLoginCrypto struct {
	matchErr error
}

func (m *mockLoginCrypto) Generate(_ string) (string, error) {
	return "", nil
}

func (m *mockLoginCrypto) Matches(_, _ string) error {
	return m.matchErr
}

func TestLoginHandler_SuccessByEmail(t *testing.T) {
	u := &user.User{ID: "u1", Email: "a@b.com", Nickname: "nick", PasswordHash: "hash"}
	repo := &mockUserRepo{getByEmailUser: u}
	sessions := &mockSessionManager{createSession: &session.Session{Token: "tok123"}}
	crypto := &mockLoginCrypto{}
	h := NewLoginHandler(repo, crypto, sessions)

	result, err := h.Execute(context.Background(), LoginCommand{
		Identifier: "a@b.com",
		Password:   "pass",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Token != "tok123" {
		t.Errorf("Token = %q, want %q", result.Token, "tok123")
	}
	if result.User.ID != "u1" {
		t.Errorf("User.ID = %q, want %q", result.User.ID, "u1")
	}
}

func TestLoginHandler_SuccessByUsername(t *testing.T) {
	u := &user.User{ID: "u2", Nickname: "nick"}
	repo := &mockUserRepo{getByEmailErr: user.ErrUserNotFound}
	repo.getByUsernameUser = u
	sessions := &mockSessionManager{createSession: &session.Session{Token: "tok456"}}
	crypto := &mockLoginCrypto{}
	h := NewLoginHandler(repo, crypto, sessions)

	result, err := h.Execute(context.Background(), LoginCommand{
		Identifier: "nick",
		Password:   "pass",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Token != "tok456" {
		t.Errorf("Token = %q, want %q", result.Token, "tok456")
	}
}

func TestLoginHandler_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{getByEmailErr: user.ErrUserNotFound, getByUsernameErr: user.ErrUserNotFound}
	h := NewLoginHandler(repo, &mockLoginCrypto{}, &mockSessionManager{})

	_, err := h.Execute(context.Background(), LoginCommand{
		Identifier: "unknown",
		Password:   "pass",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Execute() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	u := &user.User{ID: "u1", PasswordHash: "hash"}
	repo := &mockUserRepo{getByEmailUser: u}
	crypto := &mockLoginCrypto{matchErr: errors.New("mismatch")}
	h := NewLoginHandler(repo, crypto, &mockSessionManager{})

	_, err := h.Execute(context.Background(), LoginCommand{
		Identifier: "a@b.com",
		Password:   "wrong",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Execute() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestLoginHandler_SessionCreateFailure(t *testing.T) {
	u := &user.User{ID: "u1", PasswordHash: "hash"}
	repo := &mockUserRepo{getByEmailUser: u}
	sessions := &mockSessionManager{createErr: errors.New("session failed")}
	h := NewLoginHandler(repo, &mockLoginCrypto{}, sessions)

	_, err := h.Execute(context.Background(), LoginCommand{
		Identifier: "a@b.com",
		Password:   "pass",
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestNewLoginHandler(t *testing.T) {
	h := NewLoginHandler(&mockUserRepo{}, &mockLoginCrypto{}, &mockSessionManager{})
	if h == nil {
		t.Fatal("NewLoginHandler() returned nil")
	}
}
