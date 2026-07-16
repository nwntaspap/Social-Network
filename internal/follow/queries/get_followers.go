package queries

import (
	"context"

	"social-network/internal/follow"
)

type GetFollowersQuery struct {
	UserID string
}

type GetFollowersResolver struct {
	repo follow.Repository
}

func NewGetFollowersResolver(repo follow.Repository) *GetFollowersResolver {
	return &GetFollowersResolver{repo: repo}
}

func (r *GetFollowersResolver) Resolve(ctx context.Context, q GetFollowersQuery) ([]follow.Follow, error) {
	return r.repo.GetFollowers(ctx, q.UserID)
}
