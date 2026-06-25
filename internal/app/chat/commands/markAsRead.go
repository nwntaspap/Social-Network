package chatcommands

import (
	"context"

	"social-network/internal/domain/chat"
)

type MarkAsReadRequest struct {
	ChatID        string
	UserID        string
	UpToMessageID int
}

type MarkAsReadHandler interface {
	Handle(ctx context.Context, req MarkAsReadRequest) error
}

type markAsReadHandler struct {
	chatRepo chat.Repository
}

func NewMarkAsReadHandler(chatRepo chat.Repository) MarkAsReadHandler {
	return &markAsReadHandler{
		chatRepo: chatRepo,
	}
}

func (h *markAsReadHandler) Handle(ctx context.Context, req MarkAsReadRequest) error {
	return h.chatRepo.MarkAsRead(ctx, req.ChatID, req.UserID, req.UpToMessageID)
}
