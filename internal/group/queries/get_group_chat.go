package queries

import (
	"context"

	"social-network/internal/group"
)

type GetGroupChatQuery struct {
	GroupID string
	UserID  string
}

type GetGroupChatResult struct {
	Messages []group.PostComment
}

type GetGroupChatResolver struct {
	repo group.Repository
}

func NewGetGroupChatResolver(repo group.Repository) *GetGroupChatResolver {
	return &GetGroupChatResolver{repo: repo}
}

func (r *GetGroupChatResolver) Resolve(ctx context.Context, q GetGroupChatQuery) (*GetGroupChatResult, error) {
	if q.UserID != "" {
		isMember, err := r.repo.IsMember(ctx, q.GroupID, q.UserID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, group.ErrNotMember
		}
	}

	return &GetGroupChatResult{}, nil
}
