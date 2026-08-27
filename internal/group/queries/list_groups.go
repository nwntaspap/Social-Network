package queries

import (
	"context"

	"social-network/internal/group"
)

type ListGroupsQuery struct {
	Query  string
	Page   int
	Size   int
	UserID string
}

type ListGroupsResult struct {
	Groups []group.Group
	Total  int
}

type ListGroupsRepository interface {
	SearchGroups(ctx context.Context, query string, page, size int) ([]group.Group, int, error)
	membershipChecker
}

type ListGroupsResolver struct {
	repo ListGroupsRepository
}

func NewListGroupsResolver(repo ListGroupsRepository) *ListGroupsResolver {
	return &ListGroupsResolver{repo: repo}
}

func (r *ListGroupsResolver) Resolve(ctx context.Context, q ListGroupsQuery) (*ListGroupsResult, error) {
	groups, total, err := r.repo.SearchGroups(ctx, q.Query, q.Page, q.Size)
	if err != nil {
		return nil, err
	}

	for i := range groups {
		groups[i].MembershipStatus = computeMembershipStatus(ctx, r.repo, groups[i].ID, q.UserID)
	}

	return &ListGroupsResult{Groups: groups, Total: total}, nil
}
