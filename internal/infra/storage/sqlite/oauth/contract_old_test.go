package oauthrepo

import (
	"context"
	"database/sql"
	"testing"

	"social-network/internal/domain/oauth"

	_ "github.com/mattn/go-sqlite3"
)

func setupContractDB(t *testing.T) *sql.DB {
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

func insertTestUser(t *testing.T, db *sql.DB, id, username, email string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO users (id, username, email, password_hash) VALUES (?, ?, ?, '')`,
		id, username, email,
	)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
}

func insertTestOAuthProvider(t *testing.T, db *sql.DB, userID, provider, providerUserID, email, username string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO oauth_providers (user_id, provider, provider_user_id, email, username, avatar_url) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, provider, providerUserID, email, username, "https://avatar.example.com/"+providerUserID,
	)
	if err != nil {
		t.Fatalf("failed to insert oauth provider: %v", err)
	}
}

// TestOldStore_GetUserByProviderID_Found verifies that GetUserByProviderID
// returns the linked user when the provider+providerUserID combination exists.
func TestOldStore_GetUserByProviderID_Found(t *testing.T) {
	db := setupContractDB(t)
	defer db.Close()

	repo := NewOAuthRepository(db)

	insertTestUser(t, db, "user-1", "alice", "alice@example.com")
	insertTestOAuthProvider(t, db, "user-1", "github", "gh_123", "alice@example.com", "alice")

	got, err := repo.GetUserByProviderID(context.Background(), oauth.ProviderGitHub, "gh_123")
	if err != nil {
		t.Fatalf("GetUserByProviderID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetUserByProviderID() returned nil user")
	}
	if got.ID != "user-1" {
		t.Errorf("ID = %q, want %q", got.ID, "user-1")
	}
	if got.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", got.Email, "alice@example.com")
	}
	if got.Nickname != "alice" {
		t.Errorf("Nickname = %q, want %q", got.Nickname, "alice")
	}
}

// TestOldStore_GetUserByProviderID_NotFound verifies that GetUserByProviderID
// returns ErrUserNotFound when no matching record exists.
func TestOldStore_GetUserByProviderID_NotFound(t *testing.T) {
	db := setupContractDB(t)
	defer db.Close()

	repo := NewOAuthRepository(db)

	_, err := repo.GetUserByProviderID(context.Background(), oauth.ProviderGitHub, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent provider user ID")
	}
	if err != oauth.ErrUserNotFound {
		t.Errorf("error = %v, want %v", err, oauth.ErrUserNotFound)
	}
}

// TestOldStore_GetUserByEmail_Found verifies that GetUserByEmail
// returns the user when the email exists.
func TestOldStore_GetUserByEmail_Found(t *testing.T) {
	db := setupContractDB(t)
	defer db.Close()

	repo := NewOAuthRepository(db)

	insertTestUser(t, db, "user-1", "alice", "alice@example.com")

	got, err := repo.GetUserByEmail(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetUserByEmail() returned nil user")
	}
	if got.ID != "user-1" {
		t.Errorf("ID = %q, want %q", got.ID, "user-1")
	}
	if got.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", got.Email, "alice@example.com")
	}
}

// TestOldStore_GetUserByEmail_NotFound verifies that GetUserByEmail
// returns ErrUserNotFound when no user has that email.
func TestOldStore_GetUserByEmail_NotFound(t *testing.T) {
	db := setupContractDB(t)
	defer db.Close()

	repo := NewOAuthRepository(db)

	_, err := repo.GetUserByEmail(context.Background(), "nobody@example.com")
	if err == nil {
		t.Fatal("expected error for nonexistent email")
	}
	if err != oauth.ErrUserNotFound {
		t.Errorf("error = %v, want %v", err, oauth.ErrUserNotFound)
	}
}

// TestOldStore_CreateOAuthUser_Success verifies that CreateOAuthUser
// creates both a users record and an oauth_providers record, and returns
// the created user.
func TestOldStore_CreateOAuthUser_Success(t *testing.T) {
	db := setupContractDB(t)
	defer db.Close()

	repo := NewOAuthRepository(db)

	oauthUser := &oauth.User{
		UserID:     "new-user-1",
		ProviderID: "gh_456",
		Provider:   oauth.ProviderGitHub,
		Email:      "bob@example.com",
		Username:   "bob",
		AvatarURL:  "https://avatar.example.com/bob",
		Name:       "Bob",
	}

	got, err := repo.CreateOAuthUser(context.Background(), oauthUser)
	if err != nil {
		t.Fatalf("CreateOAuthUser() error = %v", err)
	}
	if got == nil {
		t.Fatal("CreateOAuthUser() returned nil user")
	}
	if got.ID != "new-user-1" {
		t.Errorf("ID = %q, want %q", got.ID, "new-user-1")
	}
	if got.Email != "bob@example.com" {
		t.Errorf("Email = %q, want %q", got.Email, "bob@example.com")
	}
	if got.Nickname != "bob" {
		t.Errorf("Nickname = %q, want %q", got.Nickname, "bob")
	}

	// Verify the oauth_providers record was created
	oauthRecord, err := repo.GetOAuthProvider(context.Background(), "new-user-1", oauth.ProviderGitHub)
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

// TestOldStore_LinkOAuthProvider_Success verifies that LinkOAuthProvider
// inserts an oauth_providers record for an existing user.
func TestOldStore_LinkOAuthProvider_Success(t *testing.T) {
	db := setupContractDB(t)
	defer db.Close()

	repo := NewOAuthRepository(db)

	insertTestUser(t, db, "user-1", "alice", "alice@example.com")

	oauthUser := &oauth.User{
		ProviderID: "gh_789",
		Provider:   oauth.ProviderGitHub,
		Email:      "alice@example.com",
		Username:   "alice",
		AvatarURL:  "https://avatar.example.com/gh_789",
		Name:       "Alice",
	}

	err := repo.LinkOAuthProvider(context.Background(), "user-1", oauthUser)
	if err != nil {
		t.Fatalf("LinkOAuthProvider() error = %v", err)
	}

	// Verify the link was created
	got, err := repo.GetOAuthProvider(context.Background(), "user-1", oauth.ProviderGitHub)
	if err != nil {
		t.Fatalf("GetOAuthProvider() after link error = %v", err)
	}
	if got.ProviderID != "gh_789" {
		t.Errorf("ProviderID = %q, want %q", got.ProviderID, "gh_789")
	}
}

// TestOldStore_GetOAuthProvider_Found verifies that GetOAuthProvider
// returns the linked OAuth profile when it exists.
func TestOldStore_GetOAuthProvider_Found(t *testing.T) {
	db := setupContractDB(t)
	defer db.Close()

	repo := NewOAuthRepository(db)

	insertTestUser(t, db, "user-1", "alice", "alice@example.com")
	insertTestOAuthProvider(t, db, "user-1", "github", "gh_123", "alice@example.com", "alice")

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

// TestOldStore_GetOAuthProvider_NotFound verifies that GetOAuthProvider
// returns an error when no linked provider exists for the user.
func TestOldStore_GetOAuthProvider_NotFound(t *testing.T) {
	db := setupContractDB(t)
	defer db.Close()

	repo := NewOAuthRepository(db)

	_, err := repo.GetOAuthProvider(context.Background(), "nonexistent", oauth.ProviderGitHub)
	if err == nil {
		t.Fatal("expected error for nonexistent user/provider")
	}
}
