package queries

import (
	"context"

	"social-network/internal/group"
)

type ListMyGroupsQuery struct {
	UserID string
	Page   int
	Size   int
}

type ListMyGroupsResult struct {
	Groups []group.Group
	Total  int
}

type MyGroupsRepository interface {
	ListUserGroups(ctx context.Context, userID string, page, size int) ([]group.Group, int, error)
	membershipChecker
}

type ListMyGroupsResolver struct {
	repo MyGroupsRepository
}

func NewListMyGroupsResolver(repo MyGroupsRepository) *ListMyGroupsResolver {
	return &ListMyGroupsResolver{repo: repo}
}

func (r *ListMyGroupsResolver) Resolve(ctx context.Context, q ListMyGroupsQuery) (*ListMyGroupsResult, error) {
	groups, total, err := r.repo.ListUserGroups(ctx, q.UserID, q.Page, q.Size)
	if err != nil {
		return nil, err
	}

	for i := range groups {
		groups[i].MembershipStatus = computeMembershipStatus(ctx, r.repo, groups[i].ID, q.UserID)
	}

	return &ListMyGroupsResult{Groups: groups, Total: total}, nil
}
