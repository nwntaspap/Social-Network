package queries

import (
	"context"

	"social-network/internal/group"
)

type GetGroupFeedQuery struct {
	GroupID string
	UserID  string
	Page    int
	Size    int
}

type GetGroupFeedResult struct {
	Posts []group.Post
	Total int
}

type GetGroupFeedResolver struct {
	repo group.Repository
}

func NewGetGroupFeedResolver(repo group.Repository) *GetGroupFeedResolver {
	return &GetGroupFeedResolver{repo: repo}
}

func (r *GetGroupFeedResolver) Resolve(ctx context.Context, q GetGroupFeedQuery) (*GetGroupFeedResult, error) {
	if q.UserID != "" {
		isMember, err := r.repo.IsMember(ctx, q.GroupID, q.UserID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, group.ErrNotMember
		}
	}

	posts, total, err := r.repo.GetPostsByGroupID(ctx, q.GroupID, q.UserID, q.Page, q.Size)
	if err != nil {
		return nil, err
	}
	return &GetGroupFeedResult{Posts: posts, Total: total}, nil
}
