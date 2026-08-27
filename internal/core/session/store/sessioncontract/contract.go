package sessioncontract

import (
	"context"
	"testing"
	"time"

	coresession "social-network/internal/core/session"
)

type Store interface {
	Create(ctx context.Context, userID string) (*coresession.Session, error)
	Get(ctx context.Context, token string) (*coresession.Session, error)
	Revoke(ctx context.Context, token string) error
}

func TestCreateAndGet(t *testing.T, s Store) {
	t.Helper()
	ctx := t.Context()

	sess, err := s.Create(ctx, "user-1")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if sess.Token == "" {
		t.Error("Token is empty")
	}
	if sess.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", sess.UserID, "user-1")
	}
	if sess.ExpiresAt.IsZero() {
		t.Error("ExpiresAt is zero")
	}

	got, err := s.Get(ctx, sess.Token)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Token != sess.Token {
		t.Errorf("Token = %q, want %q", got.Token, sess.Token)
	}
	if got.UserID != sess.UserID {
		t.Errorf("UserID = %q, want %q", got.UserID, sess.UserID)
	}
}

func TestGetNotFound(t *testing.T, s Store) {
	t.Helper()
	ctx := t.Context()

	_, err := s.Get(ctx, "non-existent-token")
	if err == nil {
		t.Fatal("expected error for non-existent session")
	}
}

func TestGetExpired(t *testing.T, s Store, expiry time.Duration) {
	t.Helper()
	ctx := t.Context()

	sess, err := s.Create(ctx, "user-expired")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	time.Sleep(expiry + 50*time.Millisecond)

	_, err = s.Get(ctx, sess.Token)
	if err == nil {
		t.Fatal("expected error for expired session")
	}
}

func TestRevoke(t *testing.T, s Store) {
	t.Helper()
	ctx := t.Context()

	sess, err := s.Create(ctx, "user-revoke")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if derr := s.Revoke(ctx, sess.Token); derr != nil {
		t.Fatalf("Revoke() error = %v", derr)
	}

	_, err = s.Get(ctx, sess.Token)
	if err == nil {
		t.Fatal("expected error after revoke")
	}
}
