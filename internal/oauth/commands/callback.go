package commands

import (
	"context"
	"errors"
	"fmt"

	"social-network/internal/oauth"
	"social-network/internal/pkg/uuid"
)

var ErrCodeMissing = errors.New("oauth: code missing in callback")

// StateVerifier validates an OAuth state token and returns the embedded data.
type StateVerifier interface {
	Verify(state string) (StateData, error)
}

// ProviderClient exchanges codes and fetches user info from OAuth providers.
type ProviderClient interface {
	Name() string
	ExchangeCode(ctx context.Context, code string) (string, error)
	GetUserInfo(ctx context.Context, accessToken string) (*oauth.User, error)
}

// CallbackCommand handles the OAuth callback after the provider redirects back.
type CallbackCommand struct {
	Code        string
	State       string
	FrontendURL string // base URL for redirect with success/error params
}

// CallbackResult contains the session and redirect URL after processing.
type CallbackResult struct {
	Session     *oauth.Session
	FrontendURL string
}

// CallbackHandler processes OAuth callback responses.
type CallbackHandler struct {
	repo           oauth.Repository
	stateVerifier  StateVerifier
	provider       ProviderClient
	sessionCreator oauth.SessionCreator
	providerName   string
}

func NewCallbackHandler(
	repo oauth.Repository,
	sv StateVerifier,
	provider ProviderClient,
	sc oauth.SessionCreator,
	providerName string,
) *CallbackHandler {
	return &CallbackHandler{
		repo:           repo,
		stateVerifier:  sv,
		provider:       provider,
		sessionCreator: sc,
		providerName:   providerName,
	}
}

func (h *CallbackHandler) Execute(ctx context.Context, cmd CallbackCommand) (*CallbackResult, error) {
	if cmd.Code == "" {
		return nil, ErrCodeMissing
	}

	stateData, err := h.stateVerifier.Verify(cmd.State)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", oauth.ErrStateInvalid, err)
	}

	accessToken, err := h.provider.ExchangeCode(ctx, cmd.Code)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	providerUser, err := h.provider.GetUserInfo(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}

	// Ensure the provider user has the correct provider set
	providerUser.Provider = oauth.Provider(h.providerName)

	switch stateData.Flow {
	case "login":
		return h.handleLogin(ctx, stateData, providerUser, cmd.FrontendURL)
	case "link":
		return h.handleLink(ctx, stateData, providerUser, cmd.FrontendURL)
	default:
		return &CallbackResult{
			FrontendURL: appendErrorParam(cmd.FrontendURL, "flowNotRecognized"),
		}, nil
	}
}

func (h *CallbackHandler) handleLogin(ctx context.Context, _ StateData, providerUser *oauth.User, frontendURL string) (*CallbackResult, error) {
	// 1. Check if user exists by provider ID → login
	existingUserID, _ := h.repo.GetUserByProviderID(ctx, providerUser.Provider, providerUser.ProviderID)
	if existingUserID != "" {
		session, err := h.sessionCreator.CreateSession(ctx, existingUserID)
		if err != nil {
			return &CallbackResult{ //nolint:nilerr // Intentional: redirect to frontend with error param
				FrontendURL: appendErrorParam(frontendURL, "errorCreatingSession"),
			}, nil
		}
		return &CallbackResult{
			Session:     session,
			FrontendURL: appendSuccessParam(frontendURL, "login", h.providerName),
		}, nil
	}

	// 2. Check if user exists by email → return error
	existingByEmail, _ := h.repo.GetUserByEmail(ctx, providerUser.Email)
	if existingByEmail != "" {
		return &CallbackResult{
			FrontendURL: appendErrorParam(frontendURL, "email_exists"),
		}, nil
	}

	// 3. Create new user
	if providerUser.UserID == "" {
		providerUser.UserID = uuid.NewProvider().NewUUID()
	}
	newUserID, err := h.repo.CreateOAuthUser(ctx, providerUser)
	if err != nil {
		return &CallbackResult{
			FrontendURL: appendErrorParam(frontendURL, "errorAtOauthLogin"),
		}, fmt.Errorf("create oauth user: %w", err)
	}

	session, err := h.sessionCreator.CreateSession(ctx, newUserID)
	if err != nil {
		return &CallbackResult{ //nolint:nilerr // Intentional: redirect to frontend with error param
			FrontendURL: appendErrorParam(frontendURL, "errorCreatingSession"),
		}, nil
	}
	return &CallbackResult{
		Session:     session,
		FrontendURL: appendSuccessParam(frontendURL, "login", h.providerName),
	}, nil
}

func (h *CallbackHandler) handleLink(ctx context.Context, stateData StateData, providerUser *oauth.User, frontendURL string) (*CallbackResult, error) {
	// Check if provider account already linked to different user
	existing, _ := h.repo.GetUserByProviderID(ctx, providerUser.Provider, providerUser.ProviderID)
	if existing != "" && existing != stateData.UserID {
		return &CallbackResult{
			FrontendURL: appendErrorParam(frontendURL, "providerAccountBelongsToAnotherUser"),
		}, nil
	}
	if existing == stateData.UserID {
		return &CallbackResult{
			FrontendURL: appendErrorParam(frontendURL, "alreadyLinkedToProvider"),
		}, nil
	}

	err := h.repo.LinkOAuthProvider(ctx, stateData.UserID, providerUser)
	if err != nil {
		return &CallbackResult{
			FrontendURL: appendErrorParam(frontendURL, "errorAtOauthLogin"),
		}, fmt.Errorf("link oauth provider: %w", err)
	}

	return &CallbackResult{
		FrontendURL: appendSuccessParam(frontendURL, "link", h.providerName),
	}, nil
}

func appendErrorParam(baseURL, errCode string) string {
	if baseURL == "" {
		return ""
	}
	separator := "?"
	if contains(baseURL, "?") {
		separator = "&"
	}
	return baseURL + separator + "flow=error&error=" + errCode
}

func appendSuccessParam(baseURL, flow, provider string) string {
	if baseURL == "" {
		return ""
	}
	separator := "?"
	if contains(baseURL, "?") {
		separator = "&"
	}
	return baseURL + separator + "flow=" + flow + "&success=ok&provider=" + provider
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
