package oauth

import (
	"context"

	"social-network/internal/domain/user"
)

type Repository interface {
	GetUserByProviderID(ctx context.Context, provider Provider, providerUserID string) (*user.User, error)
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	CreateOAuthUser(ctx context.Context, oauthUser *User) (*user.User, error)
	LinkOAuthProvider(ctx context.Context, userID string, oauthUser *User) error
	GetOAuthProvider(ctx context.Context, userID string, provider Provider) (*User, error)
}
