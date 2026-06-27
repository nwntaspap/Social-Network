package store

import (
	"context"
	"testing"
	"time"

	"social-network/internal/core/session/store/sessioncontract"
	"social-network/internal/domain/session"
	"social-network/internal/platform/database"
)

func setupNewStore(t *testing.T, expiry time.Duration) *Store {
	t.Helper()

	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.ExecContext(context.Background(),
		`CREATE TABLE sessions (
			token TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			expires_at DATETIME NOT NULL
		)`)
	if err != nil {
		t.Fatalf("create sessions table: %v", err)
	}

	return NewSessionStore(db, WithExpiry(expiry))
}

type newStoreAdapter struct {
	*Store
}

func (a *newStoreAdapter) CreateSession(ctx context.Context, userID string) (*session.Session, error) {
	sess, err := a.Create(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &session.Session{
		AccessToken: sess.Token,
		UserID:      sess.UserID,
		Expiry:      sess.ExpiresAt,
	}, nil
}

func (a *newStoreAdapter) GetSession(sessionID string) (*session.Session, error) {
	sess, err := a.Get(context.Background(), sessionID)
	if err != nil {
		return nil, err
	}
	return &session.Session{
		AccessToken: sess.Token,
		UserID:      sess.UserID,
		Expiry:      sess.ExpiresAt,
	}, nil
}

func (a *newStoreAdapter) DeleteSession(sessionID string) error {
	return a.Revoke(context.Background(), sessionID)
}

func TestNewStore_CreateAndGet(t *testing.T) {
	s := setupNewStore(t, 24*time.Hour)
	sessioncontract.TestCreateAndGet(t, &newStoreAdapter{s})
}

func TestNewStore_GetNotFound(t *testing.T) {
	s := setupNewStore(t, 24*time.Hour)
	sessioncontract.TestGetNotFound(t, &newStoreAdapter{s})
}

func TestNewStore_GetExpired(t *testing.T) {
	s := setupNewStore(t, 50*time.Millisecond)
	sessioncontract.TestGetExpired(t, &newStoreAdapter{s}, 50*time.Millisecond)
}

func TestNewStore_Revoke(t *testing.T) {
	s := setupNewStore(t, 24*time.Hour)
	sessioncontract.TestRevoke(t, &newStoreAdapter{s})
}
