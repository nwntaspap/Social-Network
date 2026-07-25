package commands

import (
	"context"

	"social-network/internal/chat"
)

type MarkAsReadCommand struct {
	ChatID        string
	UserID        string
	UpToMessageID int
}

type MarkAsReadHandler struct {
	repo chat.Repository
}

func NewMarkAsReadHandler(repo chat.Repository) *MarkAsReadHandler {
	return &MarkAsReadHandler{repo: repo}
}

func (h *MarkAsReadHandler) Execute(ctx context.Context, cmd MarkAsReadCommand) error {
	return h.repo.MarkAsRead(ctx, cmd.ChatID, cmd.UserID, cmd.UpToMessageID)
}
