package oauth

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Provider represents an OAuth provider name.
type Provider string

const (
	ProviderGitHub Provider = "github"
	ProviderGoogle Provider = "google"
)

// User represents a third-party OAuth profile linked to a local user.
type User struct {
	UserID     string
	ProviderID string
	Provider   Provider
	Email      string
	Username   string
	AvatarURL  string
	Name       string
}

// Repository is implemented by store/sqlite.go.
// Each command/query in commands/ accepts this interface.
type Repository interface {
	GetUserByProviderID(ctx context.Context, provider Provider, providerUserID string) (string, error)
	GetUserByEmail(ctx context.Context, email string) (string, error)
	CreateOAuthUser(ctx context.Context, user *User) (string, error)
	LinkOAuthProvider(ctx context.Context, userID string, user *User) error
	GetOAuthProvider(ctx context.Context, userID string, provider Provider) (*User, error)
}

// ProviderClient is the interface that github/google clients satisfy.
type ProviderClient interface {
	Name() string
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (string, error)
	GetUserInfo(ctx context.Context, accessToken string) (*ProviderUserInfo, error)
}

// ProviderUserInfo holds the user profile data returned by an OAuth provider.
type ProviderUserInfo struct {
	ProviderID string
	Email      string
	Username   string
	Name       string
	AvatarURL  string
}

// Cross-slice interfaces (defined locally, satisfied by other slices via bootstrap wiring)

// SessionCreator creates a session for an authenticated user.
type SessionCreator interface {
	CreateSession(ctx context.Context, userID string) (*Session, error)
}

// Session holds the tokens returned after authentication.
type Session struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// CookieSetter writes session cookies to the HTTP response.
type CookieSetter interface {
	SetCookies(w http.ResponseWriter, session *Session)
}

// UserFinder looks up a local user by ID.
type UserFinder interface {
	GetByID(ctx context.Context, id string) (*LocalUser, error)
}

// LocalUser is a minimal projection of a local user — only fields needed by the OAuth slice.
type LocalUser struct {
	ID       string
	Email    string
	Nickname string
}

// Errors

var (
	ErrUserNotFound                        = errors.New("oauth: user not found by provider ID")
	ErrUserWithEmailExists                 = errors.New("oauth: user with email already exists")
	ErrProviderAccountBelongsToAnotherUser = errors.New("oauth: provider account belongs to another user")
	ErrAlreadyLinkedToProvider             = errors.New("oauth: user already linked to this provider")
	ErrProviderNotFound                    = errors.New("oauth: provider not found")
	ErrFlowNotRecognized                   = errors.New("oauth: flow not recognized")
	ErrStateInvalid                        = errors.New("oauth: state invalid or expired")
)
