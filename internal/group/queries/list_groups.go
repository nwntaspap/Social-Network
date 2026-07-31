package queries

import (
	"context"

	"social-network/internal/group"
)

type ListGroupsQuery struct {
	Page   int
	Size   int
	UserID string
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

	for i := range groups {
		groups[i].MembershipStatus = computeMembershipStatus(ctx, r.repo, groups[i].ID, q.UserID)
	}

	return &ListGroupsResult{Groups: groups, Total: total}, nil
}
