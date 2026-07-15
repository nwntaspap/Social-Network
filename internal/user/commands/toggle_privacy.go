package commands

import (
	"context"

	"social-network/internal/user"
)

type TogglePrivacyCommand struct {
	UserID    string
	IsPrivate bool
}

type TogglePrivacyHandler struct {
	repo user.Repository
}

func NewTogglePrivacyHandler(repo user.Repository) *TogglePrivacyHandler {
	return &TogglePrivacyHandler{repo: repo}
}

func (h *TogglePrivacyHandler) Execute(ctx context.Context, cmd TogglePrivacyCommand) error {
	return h.repo.TogglePrivacy(ctx, cmd.UserID, cmd.IsPrivate)
}
