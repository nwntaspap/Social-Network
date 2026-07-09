package sessioncontract

import (
	"context"
	"testing"
	"time"

	domainSession "social-network/internal/domain/session"
)

type Store interface {
	CreateSession(ctx context.Context, userID string) (*domainSession.Session, error)
	GetSession(sessionID string) (*domainSession.Session, error)
	DeleteSession(sessionID string) error
}

func TestCreateAndGet(t *testing.T, s Store) {
	t.Helper()
	ctx := t.Context()

	sess, err := s.CreateSession(ctx, "user-1")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if sess.AccessToken == "" {
		t.Error("AccessToken is empty")
	}
	if sess.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", sess.UserID, "user-1")
	}
	if sess.Expiry.IsZero() {
		t.Error("Expiry is zero")
	}

	got, err := s.GetSession(sess.AccessToken)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if got.AccessToken != sess.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, sess.AccessToken)
	}
	if got.UserID != sess.UserID {
		t.Errorf("UserID = %q, want %q", got.UserID, sess.UserID)
	}
}

func TestGetNotFound(t *testing.T, s Store) {
	t.Helper()

	_, err := s.GetSession("non-existent-token")
	if err == nil {
		t.Fatal("expected error for non-existent session")
	}
}

func TestGetExpired(t *testing.T, s Store, expiry time.Duration) {
	t.Helper()
	ctx := t.Context()

	sess, err := s.CreateSession(ctx, "user-expired")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	time.Sleep(expiry + 50*time.Millisecond)

	_, err = s.GetSession(sess.AccessToken)
	if err == nil {
		t.Fatal("expected error for expired session")
	}
}

func TestRevoke(t *testing.T, s Store) {
	t.Helper()
	ctx := t.Context()

	sess, err := s.CreateSession(ctx, "user-revoke")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if derr := s.DeleteSession(sess.AccessToken); derr != nil {
		t.Fatalf("DeleteSession() error = %v", derr)
	}

	_, err = s.GetSession(sess.AccessToken)
	if err == nil {
		t.Fatal("expected error after revoke")
	}
}
