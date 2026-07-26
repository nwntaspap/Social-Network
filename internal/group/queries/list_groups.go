package queries

import (
	"context"

	"social-network/internal/group"
)

type ListGroupsQuery struct {
	Page int
	Size int
}

type ListGroupsResult struct {
	Groups []group.Group
	Total  int
}

type ListGroupsResolver struct {
	repo group.Repository
}

func NewListGroupsResolver(repo group.Repository) *ListGroupsResolver {
	return &ListGroupsResolver{repo: repo}
}

func (r *ListGroupsResolver) Resolve(ctx context.Context, q ListGroupsQuery) (*ListGroupsResult, error) {
	groups, total, err := r.repo.ListGroups(ctx, q.Page, q.Size)
	if err != nil {
		return nil, err
	}
	return &ListGroupsResult{Groups: groups, Total: total}, nil
}
