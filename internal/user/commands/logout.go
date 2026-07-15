package commands

import (
	"context"

	"social-network/internal/core/session"
)

type LogoutCommand struct {
	Token string
}

type LogoutHandler struct {
	sessions session.Manager
}

func NewLogoutHandler(sessions session.Manager) *LogoutHandler {
	return &LogoutHandler{sessions: sessions}
}

func (h *LogoutHandler) Execute(ctx context.Context, cmd LogoutCommand) error {
	return h.sessions.Revoke(ctx, cmd.Token)
}
