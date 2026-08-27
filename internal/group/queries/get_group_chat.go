package queries

import (
	"context"

	"social-network/internal/group"
)

type GetGroupChatQuery struct {
	GroupID string
	UserID  string
	Limit   int
}

type GetGroupChatResult struct {
	Messages []group.ChatMessage
}

type GetGroupChatResolver struct {
	repo group.Repository
}

func NewGetGroupChatResolver(repo group.Repository) *GetGroupChatResolver {
	return &GetGroupChatResolver{repo: repo}
}

func (r *GetGroupChatResolver) Resolve(ctx context.Context, q GetGroupChatQuery) (*GetGroupChatResult, error) {
	isMember, err := r.repo.IsMember(ctx, q.GroupID, q.UserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, group.ErrNotMember
	}

	limit := q.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	messages, err := r.repo.GetGroupChatMessages(ctx, q.GroupID, limit)
	if err != nil {
		return nil, err
	}
	return &GetGroupChatResult{Messages: messages}, nil
}
