package oauthrepo

import (
	"context"
	"database/sql"
	"testing"

	"social-network/internal/domain/oauth"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetOAuthProvider_NoScanCtx(t *testing.T) {
	db := setupSQLiteDB(t)
	defer db.Close()

	repo := NewOAuthRepository(db)

	insertUser(t, db, "user-1", "alice", "alice@example.com")
	insertOAuthProvider(t, db, "user-1", "github", "gh_123", "alice@example.com", "alice")

	got, err := repo.GetOAuthProvider(context.Background(), "user-1", oauth.ProviderGitHub)
	if err != nil {
		t.Fatalf("GetOAuthProvider() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetOAuthProvider() returned nil")
	}
	if got.ProviderID != "gh_123" {
		t.Errorf("ProviderID = %q, want %q", got.ProviderID, "gh_123")
	}
	if got.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", got.Email, "alice@example.com")
	}
	if got.Username != "alice" {
		t.Errorf("Username = %q, want %q", got.Username, "alice")
	}
	if got.Provider != oauth.ProviderGitHub {
		t.Errorf("Provider = %q, want %q", got.Provider, oauth.ProviderGitHub)
	}
}

func TestGetOAuthProvider_NotFound(t *testing.T) {
	db := setupSQLiteDB(t)
	defer db.Close()

	repo := NewOAuthRepository(db)

	_, err := repo.GetOAuthProvider(context.Background(), "nonexistent", oauth.ProviderGitHub)
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}

func setupSQLiteDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL,
			username TEXT NOT NULL,
			password_hash TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS oauth_providers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			provider TEXT NOT NULL,
			provider_user_id TEXT NOT NULL,
			email TEXT,
			username TEXT,
			avatar_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func insertUser(t *testing.T, db *sql.DB, id, username, email string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO users (id, username, email, password_hash) VALUES (?, ?, ?, '')`,
		id, username, email,
	)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
}

func insertOAuthProvider(t *testing.T, db *sql.DB, userID, provider, providerUserID, email, username string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO oauth_providers (user_id, provider, provider_user_id, email, username, avatar_url) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, provider, providerUserID, email, username, "https://avatar.example.com/gh_123",
	)
	if err != nil {
		t.Fatalf("failed to insert oauth provider: %v", err)
	}
}
