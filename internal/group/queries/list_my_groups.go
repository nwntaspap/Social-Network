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
	CountGroupUnread(ctx context.Context, groupIDs []string, userID string) (map[string]int, error)
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

	ids := make([]string, len(groups))
	for i := range groups {
		ids[i] = groups[i].ID
	}
	unread, err := r.repo.CountGroupUnread(ctx, ids, q.UserID)
	if err != nil {
		return nil, err
	}

	result := &ListMyGroupsResult{Groups: groups, Total: total}
	for i := range result.Groups {
		result.Groups[i].MembershipStatus = computeMembershipStatus(ctx, r.repo, result.Groups[i].ID, q.UserID)
		result.Groups[i].UnreadCount = unread[result.Groups[i].ID]
	}

	return result, nil
}
