package queries

import (
	"context"

	"social-network/internal/group"
)

type GetGroupQuery struct {
	GroupID string
	UserID  string
}

type GetGroupResult struct {
	Group            group.Group
	MembersCount     int
	MembershipStatus string
}

type GetGroupResolver struct {
	repo group.Repository
}

func NewGetGroupResolver(repo group.Repository) *GetGroupResolver {
	return &GetGroupResolver{repo: repo}
}

func (r *GetGroupResolver) Resolve(ctx context.Context, q GetGroupQuery) (*GetGroupResult, error) {
	g, err := r.repo.GetGroupByID(ctx, q.GroupID)
	if err != nil {
		return nil, err
	}

	status := computeMembershipStatus(ctx, r.repo, q.GroupID, q.UserID)

	count, err := r.repo.CountMembers(ctx, q.GroupID)
	if err != nil {
		return nil, err
	}

	return &GetGroupResult{
		Group:            *g,
		MembersCount:     count,
		MembershipStatus: status,
	}, nil
}

type membershipChecker interface {
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
	IsInvited(ctx context.Context, groupID, userID string) (bool, error)
	HasPendingRequest(ctx context.Context, groupID, userID string) (bool, error)
}

func computeMembershipStatus(ctx context.Context, repo membershipChecker, groupID, userID string) string {
	if userID == "" {
		return "none"
	}

	isMember, err := repo.IsMember(ctx, groupID, userID)
	if err != nil || isMember {
		if isMember {
			return "member"
		}
		return "none"
	}

	isInvited, err := repo.IsInvited(ctx, groupID, userID)
	if err != nil || isInvited {
		if isInvited {
			return "pending"
		}
		return "none"
	}

	hasPending, _ := repo.HasPendingRequest(ctx, groupID, userID)
	if hasPending {
		return "pending"
	}

	return "none"
}
