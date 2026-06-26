package chatqueries

import (
	"context"

	"social-network/internal/domain/chat"
)

type GetChatHistoryRequest struct {
	ChatID          string
	BeforeMessageID int
	Limit           int
}

type GetChatHistoryHandler interface {
	Handle(ctx context.Context, req GetChatHistoryRequest) (messages []*chat.Message, err error)
}

type getChatHistoryHandler struct {
	chatRepo chat.Repository
}

func NewGetChatHistoryHandler(chatRepo chat.Repository) GetChatHistoryHandler {
	return &getChatHistoryHandler{
		chatRepo: chatRepo,
	}
}

func (h *getChatHistoryHandler) Handle(ctx context.Context, req GetChatHistoryRequest) (messages []*chat.Message, err error) {
	limit := req.Limit
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	if req.BeforeMessageID > 0 {
		messages, err = h.chatRepo.GetMessagesForChatBefore(ctx, req.ChatID, req.BeforeMessageID, limit)
		if err != nil {
			return nil, err
		}
	} else {
		messages, err = h.chatRepo.GetMessagesForChat(ctx, req.ChatID, limit)
		if err != nil {
			return nil, err
		}
	}
	return messages, nil
}
