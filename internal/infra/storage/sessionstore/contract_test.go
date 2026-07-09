package sessionstore

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"social-network/internal/config"
	"social-network/internal/core/session/store/sessioncontract"

	_ "github.com/mattn/go-sqlite3"
)

func setupOldStore(t *testing.T, expiry time.Duration) *Manager {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.ExecContext(context.Background(),
		`CREATE TABLE sessions (
			token TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			refresh_token TEXT,
			refresh_token_expires_at DATETIME NOT NULL
		)`)
	if err != nil {
		t.Fatalf("create sessions table: %v", err)
	}

	cfg := config.SessionManagerConfig{
		DefaultExpiry:      expiry,
		RefreshTokenExpiry: expiry,
		SessionIDLength:    36,
		MaxSessionsPerUser: 1,
	}

	mgr, ok := NewSessionManager(db, cfg).(*Manager)
	if !ok {
		t.Fatal("NewSessionManager did not return *Manager")
	}
	return mgr
}

func TestOldStore_CreateAndGet(t *testing.T) {
	s := setupOldStore(t, 24*time.Hour)
	sessioncontract.TestCreateAndGet(t, s)
}

func TestOldStore_GetNotFound(t *testing.T) {
	s := setupOldStore(t, 24*time.Hour)
	sessioncontract.TestGetNotFound(t, s)
}

func TestOldStore_GetExpired(t *testing.T) {
	s := setupOldStore(t, 50*time.Millisecond)
	sessioncontract.TestGetExpired(t, s, 50*time.Millisecond)
}

func TestOldStore_Revoke(t *testing.T) {
	s := setupOldStore(t, 24*time.Hour)
	sessioncontract.TestRevoke(t, s)
}
