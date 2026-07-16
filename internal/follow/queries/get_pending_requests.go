package queries

import (
	"context"

	"social-network/internal/follow"
)

type GetPendingRequestsQuery struct {
	UserID string
}

type GetPendingRequestsResolver struct {
	repo follow.Repository
}

func NewGetPendingRequestsResolver(repo follow.Repository) *GetPendingRequestsResolver {
	return &GetPendingRequestsResolver{repo: repo}
}

func (r *GetPendingRequestsResolver) Resolve(ctx context.Context, q GetPendingRequestsQuery) ([]follow.Request, error) {
	return r.repo.GetPendingRequests(ctx, q.UserID)
}
