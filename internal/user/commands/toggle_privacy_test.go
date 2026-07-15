package commands

import (
	"context"
	"errors"
	"testing"
)

func TestTogglePrivacyHandler_Success(t *testing.T) {
	repo := &mockUserRepo{}
	h := NewTogglePrivacyHandler(repo)

	err := h.Execute(context.Background(), TogglePrivacyCommand{
		UserID:    "u1",
		IsPrivate: true,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestTogglePrivacyHandler_RepoFailure(t *testing.T) {
	repo := &mockUserRepo{togglePrivacyErr: errors.New("db failed")}
	h := NewTogglePrivacyHandler(repo)

	err := h.Execute(context.Background(), TogglePrivacyCommand{
		UserID:    "u1",
		IsPrivate: true,
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestNewTogglePrivacyHandler(t *testing.T) {
	h := NewTogglePrivacyHandler(&mockUserRepo{})
	if h == nil {
		t.Fatal("NewTogglePrivacyHandler() returned nil")
	}
}
