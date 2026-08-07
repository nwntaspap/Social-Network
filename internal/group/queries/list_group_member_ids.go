package queries

import (
	"context"

	"social-network/internal/group"
)

type ListGroupMemberIDsQuery struct {
	GroupID string
}

type ListGroupMemberIDsResolver struct {
	repo group.Repository
}

func NewListGroupMemberIDsResolver(repo group.Repository) *ListGroupMemberIDsResolver {
	return &ListGroupMemberIDsResolver{repo: repo}
}

func (r *ListGroupMemberIDsResolver) Resolve(ctx context.Context, q ListGroupMemberIDsQuery) ([]string, error) {
	return r.repo.ListGroupMemberIDs(ctx, q.GroupID)
}
