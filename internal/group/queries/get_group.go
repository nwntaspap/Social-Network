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

	status := r.computeMembershipStatus(ctx, q.GroupID, q.UserID)

	return &GetGroupResult{
		Group:            *g,
		MembershipStatus: status,
	}, nil
}

func (r *GetGroupResolver) computeMembershipStatus(ctx context.Context, groupID, userID string) string {
	if userID == "" {
		return "none"
	}

	isMember, err := r.repo.IsMember(ctx, groupID, userID)
	if err != nil || isMember {
		if isMember {
			return "member"
		}
		return "none"
	}

	isInvited, err := r.repo.IsInvited(ctx, groupID, userID)
	if err != nil || isInvited {
		if isInvited {
			return "pending"
		}
		return "none"
	}

	hasPending, _ := r.repo.HasPendingRequest(ctx, groupID, userID)
	if hasPending {
		return "pending"
	}

	return "none"
}
