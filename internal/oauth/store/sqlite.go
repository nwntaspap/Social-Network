package store

import (
	"context"
	"database/sql"
	"fmt"

	"social-network/internal/oauth"
	"social-network/internal/platform/database"
)

type SQLiteStore struct {
	db database.DB
}

func NewSQLiteStore(db database.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetUserByProviderID returns the local user ID that owns the given provider account.
func (s *SQLiteStore) GetUserByProviderID(ctx context.Context, provider oauth.Provider, providerUserID string) (string, error) {
	query := `
	SELECT u.id
	FROM users u
	INNER JOIN oauth_providers op ON u.id = op.user_id
	WHERE op.provider = ? AND op.provider_user_id = ?
	`

	var userID string
	err := s.db.QueryRowContext(ctx, query, string(provider), providerUserID).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", oauth.ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get user by provider ID: %w", err)
	}
	return userID, nil
}

// GetUserByEmail returns the local user ID for the given email address.
func (s *SQLiteStore) GetUserByEmail(ctx context.Context, email string) (string, error) {
	query := `SELECT id FROM users WHERE email = ?`

	var userID string
	err := s.db.QueryRowContext(ctx, query, email).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", oauth.ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get user by email: %w", err)
	}
	return userID, nil
}

// CreateOAuthUser inserts a new local user and their OAuth provider record in a transaction.
func (s *SQLiteStore) CreateOAuthUser(ctx context.Context, user *oauth.User) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO users (id, username, email, password_hash, avatar_url) VALUES (?, ?, ?, '', ?)`,
		user.UserID, user.Username, user.Email, user.AvatarURL)
	if err != nil {
		return "", fmt.Errorf("insert user: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO oauth_providers (user_id, provider, provider_user_id, email, username, avatar_url) VALUES (?, ?, ?, ?, ?, ?)`,
		user.UserID, string(user.Provider), user.ProviderID, user.Email, user.Username, user.AvatarURL)
	if err != nil {
		return "", fmt.Errorf("insert oauth provider: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("commit transaction: %w", err)
	}

	return user.UserID, nil
}

// LinkOAuthProvider adds an OAuth provider record to an existing local user.
func (s *SQLiteStore) LinkOAuthProvider(ctx context.Context, userID string, user *oauth.User) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO oauth_providers (user_id, provider, provider_user_id, email, username, avatar_url) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, string(user.Provider), user.ProviderID, user.Email, user.Username, user.AvatarURL)
	if err != nil {
		return fmt.Errorf("link oauth provider: %w", err)
	}
	return nil
}

// GetOAuthProvider returns the OAuth provider record for a given user and provider.
func (s *SQLiteStore) GetOAuthProvider(ctx context.Context, userID string, provider oauth.Provider) (*oauth.User, error) {
	query := `
	SELECT provider_user_id, email, username, avatar_url
	FROM oauth_providers
	WHERE user_id = ? AND provider = ?
	`

	var u oauth.User
	u.Provider = provider

	err := s.db.QueryRowContext(ctx, query, userID, string(provider)).Scan(
		&u.ProviderID,
		&u.Email,
		&u.Username,
		&u.AvatarURL,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("get oauth provider: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("get oauth provider: %w", err)
	}

	return &u, nil
}
