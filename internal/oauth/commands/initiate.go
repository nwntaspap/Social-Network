package commands

import (
	"context"
	"errors"

	"social-network/internal/oauth"
)

var ErrProviderNotFound = errors.New("oauth: provider not found")

// StateData holds the data embedded in the OAuth state token.
type StateData struct {
	Flow     string
	Provider string
	UserID   string
}

// StateGenerator creates and verifies OAuth state tokens.
// Satisfied by internal/pkg/oAuth.StateManager.
type StateGenerator interface {
	Generate(data StateData) (string, error)
}

// ProviderRegistry returns a ProviderClient by name.
type ProviderRegistry interface {
	Get(name string) (oauth.ProviderClient, bool)
}

// InitiateCommand starts an OAuth login or link flow.
type InitiateCommand struct {
	Provider string // "github" or "google"
	Flow     string // "login" or "link"
	UserID   string // only set for "link" flow
}

// InitiateResult contains the redirect URL for the OAuth provider.
type InitiateResult struct {
	RedirectURL string
}

// InitiateHandler generates an OAuth state and returns the provider's authorization URL.
type InitiateHandler struct {
	stateManager StateGenerator
	providers    ProviderRegistry
}

func NewInitiateHandler(sm StateGenerator, pr ProviderRegistry) *InitiateHandler {
	return &InitiateHandler{
		stateManager: sm,
		providers:    pr,
	}
}

func (h *InitiateHandler) Execute(_ context.Context, cmd InitiateCommand) (*InitiateResult, error) {
	provider, ok := h.providers.Get(cmd.Provider)
	if !ok {
		return nil, ErrProviderNotFound
	}

	stateData := StateData{
		Flow:     cmd.Flow,
		Provider: cmd.Provider,
		UserID:   cmd.UserID,
	}

	state, err := h.stateManager.Generate(stateData)
	if err != nil {
		return nil, err
	}

	authURL := provider.GetAuthURL(state)
	return &InitiateResult{RedirectURL: authURL}, nil
}
