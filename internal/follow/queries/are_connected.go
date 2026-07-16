package queries

import (
	"context"

	"social-network/internal/follow"
)

type AreConnectedQuery struct {
	UserID   string
	TargetID string
}

type AreConnectedResolver struct {
	repo follow.Repository
}

func NewAreConnectedResolver(repo follow.Repository) *AreConnectedResolver {
	return &AreConnectedResolver{repo: repo}
}

func (r *AreConnectedResolver) Resolve(ctx context.Context, q AreConnectedQuery) (bool, error) {
	return r.repo.AreConnected(ctx, q.UserID, q.TargetID)
}
