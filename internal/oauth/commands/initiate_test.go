package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/oauth"
)

type stubStateGenerator struct{}

func (s *stubStateGenerator) Generate(_ StateData) (string, error) {
	return "test-state", nil
}

type stubProviderRegistry struct{}

func (r *stubProviderRegistry) Get(_ string) (oauth.ProviderClient, bool) {
	return nil, false
}

type stubRepo struct{}

func (r *stubRepo) GetUserByProviderID(_ context.Context, _ oauth.Provider, _ string) (string, error) {
	return "", nil
}
func (r *stubRepo) GetUserByEmail(_ context.Context, _ string) (string, error) { return "", nil }
func (r *stubRepo) CreateOAuthUser(_ context.Context, _ *oauth.User) (string, error) {
	return "new-user", nil
}

func (r *stubRepo) LinkOAuthProvider(_ context.Context, _ string, _ *oauth.User) error { return nil }

func (r *stubRepo) GetOAuthProvider(_ context.Context, _ string, _ oauth.Provider) (*oauth.User, error) {
	return nil, errors.New("not implemented")
}

func TestInitiateHandler_ProviderNotFound(t *testing.T) {
	handler := NewInitiateHandler(&stubStateGenerator{}, &stubProviderRegistry{})
	_, err := handler.Execute(context.Background(), InitiateCommand{Provider: "unknown", Flow: "login"})
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestCallbackHandler_CodeMissing(t *testing.T) {
	handler := NewCallbackHandler(&stubRepo{}, nil, nil, nil, "github")
	_, err := handler.Execute(context.Background(), CallbackCommand{})
	if !errors.Is(err, ErrCodeMissing) {
		t.Errorf("expected ErrCodeMissing, got %v", err)
	}
}
