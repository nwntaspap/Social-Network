package oauth

import "context"

// Provider defines the interface that OAuth provider clients (GitHub, Google) must satisfy.
type Provider interface {
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
