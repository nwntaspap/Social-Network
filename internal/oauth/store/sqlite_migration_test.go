package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"social-network/internal/oauth"
	"social-network/internal/platform/database"

	_ "github.com/mattn/go-sqlite3"
)

func setupNewStoreDB(t *testing.T) database.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL,
			username TEXT NOT NULL,
			password_hash TEXT,
			avatar_url TEXT,
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
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(provider, provider_user_id),
			UNIQUE(user_id, provider)
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func insertNewStoreUser(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO users (id, username, email, password_hash) VALUES (?, ?, ?, '')`,
		"user-1", "alice", "alice@example.com",
	)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
}

func insertNewStoreOAuthProvider(t *testing.T, db *sql.DB, userID, provider, providerUserID, email, username string) {
	t.Helper()
	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO oauth_providers (user_id, provider, provider_user_id, email, username, avatar_url) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, provider, providerUserID, email, username, "https://avatar.example.com/"+providerUserID,
	)
	if err != nil {
		t.Fatalf("failed to insert oauth provider: %v", err)
	}
}

// --- Contract tests: same behavior as old store ---

func TestNewStore_GetUserByProviderID_Found(t *testing.T) {
	db := setupNewStoreDB(t)
	defer db.Close()

	rawDB, _ := db.(*sql.DB)
	insertNewStoreUser(t, rawDB)
	insertNewStoreOAuthProvider(t, rawDB, "user-1", "github", "gh_123", "alice@example.com", "alice")

	store := NewSQLiteStore(db)

	got, err := store.GetUserByProviderID(context.Background(), oauth.ProviderGitHub, "gh_123")
	if err != nil {
		t.Fatalf("GetUserByProviderID() error = %v", err)
	}
	if got != "user-1" {
		t.Errorf("userID = %q, want %q", got, "user-1")
	}
}

func TestNewStore_GetUserByProviderID_NotFound(t *testing.T) {
	db := setupNewStoreDB(t)
	defer db.Close()

	store := NewSQLiteStore(db)

	_, err := store.GetUserByProviderID(context.Background(), oauth.ProviderGitHub, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent provider user ID")
	}
	if !errors.Is(err, oauth.ErrUserNotFound) {
		t.Errorf("error = %v, want %v", err, oauth.ErrUserNotFound)
	}
}

func TestNewStore_GetUserByEmail_Found(t *testing.T) {
	db := setupNewStoreDB(t)
	defer db.Close()

	rawDB, _ := db.(*sql.DB)
	insertNewStoreUser(t, rawDB)

	store := NewSQLiteStore(db)

	got, err := store.GetUserByEmail(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail() error = %v", err)
	}
	if got != "user-1" {
		t.Errorf("userID = %q, want %q", got, "user-1")
	}
}

func TestNewStore_GetUserByEmail_NotFound(t *testing.T) {
	db := setupNewStoreDB(t)
	defer db.Close()

	store := NewSQLiteStore(db)

	_, err := store.GetUserByEmail(context.Background(), "nobody@example.com")
	if err == nil {
		t.Fatal("expected error for nonexistent email")
	}
	if !errors.Is(err, oauth.ErrUserNotFound) {
		t.Errorf("error = %v, want %v", err, oauth.ErrUserNotFound)
	}
}

func TestNewStore_CreateOAuthUser_Success(t *testing.T) {
	db := setupNewStoreDB(t)
	defer db.Close()

	store := NewSQLiteStore(db)

	oauthUser := &oauth.User{
		UserID:     "new-user-1",
		ProviderID: "gh_456",
		Provider:   oauth.ProviderGitHub,
		Email:      "bob@example.com",
		Username:   "bob",
		AvatarURL:  "https://avatar.example.com/bob",
		Name:       "Bob",
	}

	got, err := store.CreateOAuthUser(context.Background(), oauthUser)
	if err != nil {
		t.Fatalf("CreateOAuthUser() error = %v", err)
	}
	if got != "new-user-1" {
		t.Errorf("userID = %q, want %q", got, "new-user-1")
	}

	// Verify the oauth_providers record was created
	oauthRecord, err := store.GetOAuthProvider(context.Background(), "new-user-1", oauth.ProviderGitHub)
	if err != nil {
		t.Fatalf("GetOAuthProvider() after create error = %v", err)
	}
	if oauthRecord.ProviderID != "gh_456" {
		t.Errorf("oauth ProviderID = %q, want %q", oauthRecord.ProviderID, "gh_456")
	}
	if oauthRecord.Email != "bob@example.com" {
		t.Errorf("oauth Email = %q, want %q", oauthRecord.Email, "bob@example.com")
	}
}

func TestNewStore_LinkOAuthProvider_Success(t *testing.T) {
	db := setupNewStoreDB(t)
	defer db.Close()

	rawDB, _ := db.(*sql.DB)
	insertNewStoreUser(t, rawDB)

	store := NewSQLiteStore(db)

	oauthUser := &oauth.User{
		ProviderID: "gh_789",
		Provider:   oauth.ProviderGitHub,
		Email:      "alice@example.com",
		Username:   "alice",
		AvatarURL:  "https://avatar.example.com/gh_789",
		Name:       "Alice",
	}

	err := store.LinkOAuthProvider(context.Background(), "user-1", oauthUser)
	if err != nil {
		t.Fatalf("LinkOAuthProvider() error = %v", err)
	}

	// Verify the link was created
	got, err := store.GetOAuthProvider(context.Background(), "user-1", oauth.ProviderGitHub)
	if err != nil {
		t.Fatalf("GetOAuthProvider() after link error = %v", err)
	}
	if got.ProviderID != "gh_789" {
		t.Errorf("ProviderID = %q, want %q", got.ProviderID, "gh_789")
	}
}

func TestNewStore_GetOAuthProvider_Found(t *testing.T) {
	db := setupNewStoreDB(t)
	defer db.Close()

	rawDB, _ := db.(*sql.DB)
	insertNewStoreUser(t, rawDB)
	insertNewStoreOAuthProvider(t, rawDB, "user-1", "github", "gh_123", "alice@example.com", "alice")

	store := NewSQLiteStore(db)

	got, err := store.GetOAuthProvider(context.Background(), "user-1", oauth.ProviderGitHub)
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

func TestNewStore_GetOAuthProvider_NotFound(t *testing.T) {
	db := setupNewStoreDB(t)
	defer db.Close()

	store := NewSQLiteStore(db)

	_, err := store.GetOAuthProvider(context.Background(), "nonexistent", oauth.ProviderGitHub)
	if err == nil {
		t.Fatal("expected error for nonexistent user/provider")
	}
}
