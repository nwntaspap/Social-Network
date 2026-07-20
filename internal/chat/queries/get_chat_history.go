package queries

import (
	"context"

	"social-network/internal/chat"
)

type GetChatHistoryQuery struct {
	ChatID          string
	BeforeMessageID int
	Limit           int
}

type GetChatHistoryResolver struct {
	repo chat.Repository
}

func NewGetChatHistoryResolver(repo chat.Repository) *GetChatHistoryResolver {
	return &GetChatHistoryResolver{repo: repo}
}

func (r *GetChatHistoryResolver) Resolve(ctx context.Context, q GetChatHistoryQuery) ([]*chat.Message, error) {
	limit := q.Limit
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	if q.BeforeMessageID > 0 {
		return r.repo.GetMessagesForChatBefore(ctx, q.ChatID, q.BeforeMessageID, limit)
	}
	return r.repo.GetMessagesForChat(ctx, q.ChatID, limit)
}
