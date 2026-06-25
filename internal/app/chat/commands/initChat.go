package chatcommands

import (
	"context"

	"social-network/internal/domain/chat"
)

type InitChatRequest struct {
	MeID   string
	UserID string
}

type InitChatHandler interface {
	Handle(ctx context.Context, req InitChatRequest) (*chat.Chat, error)
}

type initChatHandler struct {
	chatRepo chat.Repository
}

func NewInitChatHandler(chatRepo chat.Repository) InitChatHandler {
	return &initChatHandler{
		chatRepo: chatRepo,
	}
}

func (h *initChatHandler) Handle(ctx context.Context, req InitChatRequest) (*chat.Chat, error) {
	c, err := h.chatRepo.GetOrCreateChat(ctx, req.MeID, req.UserID)
	if err != nil {
		return c, err
	}
	return c, nil
}
