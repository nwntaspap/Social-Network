package commands

import (
	"context"
	"errors"
	"testing"
)

func TestLogoutHandler_Success(t *testing.T) {
	sessions := &mockSessionManager{}
	h := NewLogoutHandler(sessions)

	err := h.Execute(context.Background(), LogoutCommand{Token: "tok123"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestLogoutHandler_RevokeFailure(t *testing.T) {
	sessions := &mockSessionManager{revokeErr: errors.New("revoke failed")}
	h := NewLogoutHandler(sessions)

	err := h.Execute(context.Background(), LogoutCommand{Token: "tok123"})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestNewLogoutHandler(t *testing.T) {
	h := NewLogoutHandler(&mockSessionManager{})
	if h == nil {
		t.Fatal("NewLogoutHandler() returned nil")
	}
}
