package commands

import (
	"context"
	"testing"
	"time"

	"social-network/internal/oauth"
)

type capturingRepo struct {
	createUser *oauth.User
}

func (r *capturingRepo) GetUserByProviderID(_ context.Context, _ oauth.Provider, _ string) (string, error) {
	return "", nil
}

func (r *capturingRepo) GetUserByEmail(_ context.Context, _ string) (string, error) { return "", nil }

func (r *capturingRepo) CreateOAuthUser(_ context.Context, user *oauth.User) (string, error) {
	r.createUser = user
	return "new-user-42", nil
}

func (r *capturingRepo) LinkOAuthProvider(_ context.Context, _ string, _ *oauth.User) error {
	return nil
}

func (r *capturingRepo) GetOAuthProvider(_ context.Context, _ string, _ oauth.Provider) (*oauth.User, error) {
	return nil, oauth.ErrUserNotFound
}

type fixedStateVerifier struct{ flow string }

func (v *fixedStateVerifier) Verify(_ string) (StateData, error) {
	return StateData{Flow: v.flow}, nil
}

type fixedProviderClient struct {
	user *oauth.User
}

func (p *fixedProviderClient) Name() string { return "google" }

func (p *fixedProviderClient) ExchangeCode(_ context.Context, _ string) (string, error) {
	return "access-token", nil
}

func (p *fixedProviderClient) GetUserInfo(_ context.Context, _ string) (*oauth.User, error) {
	return p.user, nil
}

type capturingSessionCreator struct {
	userID string
}

func (s *capturingSessionCreator) CreateSession(_ context.Context, userID string) (*oauth.Session, error) {
	s.userID = userID
	return &oauth.Session{
		AccessToken:  "tok",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil
}

func TestCallbackHandler_Login_GeneratesUserID(t *testing.T) {
	repo := &capturingRepo{}
	sessionCreator := &capturingSessionCreator{}
	handler := NewCallbackHandler(
		repo,
		&fixedStateVerifier{flow: "login"},
		&fixedProviderClient{user: &oauth.User{
			ProviderID: "gg_1",
			Email:      "new@example.com",
			Username:   "New User",
			Name:       "New User",
		}},
		sessionCreator,
		"google",
	)

	result, err := handler.Execute(context.Background(), CallbackCommand{
		Code:        "code",
		State:       "state",
		FrontendURL: "http://localhost:3001/auth/callback",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if repo.createUser == nil {
		t.Fatal("CreateOAuthUser was not called")
	}
	if repo.createUser.UserID == "" {
		t.Error("CreateOAuthUser called with empty UserID")
	}
	if sessionCreator.userID != "new-user-42" {
		t.Errorf("session created for %q, want %q", sessionCreator.userID, "new-user-42")
	}
	if result.Session == nil || result.Session.AccessToken != "tok" {
		t.Error("expected session with access token in result")
	}
	if result.Session.ExpiresAt.IsZero() {
		t.Error("expected session expiry to be propagated to result")
	}
}
