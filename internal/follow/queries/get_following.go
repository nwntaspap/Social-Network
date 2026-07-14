package queries

import (
	"context"

	"social-network/internal/follow"
)

type GetFollowingQuery struct {
	UserID string
}

type GetFollowingResolver struct {
	repo follow.Repository
}

func NewGetFollowingResolver(repo follow.Repository) *GetFollowingResolver {
	return &GetFollowingResolver{repo: repo}
}

func (r *GetFollowingResolver) Resolve(ctx context.Context, q GetFollowingQuery) ([]follow.Follow, error) {
	return r.repo.GetFollowing(ctx, q.UserID)
}
