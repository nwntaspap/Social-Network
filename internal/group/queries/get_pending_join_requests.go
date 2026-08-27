package queries

import (
	"context"
	"errors"

	"social-network/internal/group"
)

type GetPendingJoinRequestsQuery struct {
	GroupID string
	UserID  string
}

type JoinRequestWithGroup struct {
	Request group.JoinRequest
	Group   group.Group
}

type GetPendingJoinRequestsResult struct {
	Requests []JoinRequestWithGroup
}

type GetPendingJoinRequestsResolver struct {
	repo pendingJoinRequestsRepo
}

type pendingJoinRequestsRepo interface {
	GetMemberRole(ctx context.Context, groupID, userID string) (group.Role, error)
	GetPendingJoinRequests(ctx context.Context, groupID string) ([]group.JoinRequest, error)
	GetGroupByID(ctx context.Context, groupID string) (*group.Group, error)
}

func NewGetPendingJoinRequestsResolver(repo pendingJoinRequestsRepo) *GetPendingJoinRequestsResolver {
	return &GetPendingJoinRequestsResolver{repo: repo}
}

func (r *GetPendingJoinRequestsResolver) Resolve(ctx context.Context, q GetPendingJoinRequestsQuery) (*GetPendingJoinRequestsResult, error) {
	if q.GroupID == "" || q.UserID == "" {
		return nil, errors.New("group_id and user_id are required")
	}

	role, err := r.repo.GetMemberRole(ctx, q.GroupID, q.UserID)
	if err != nil {
		if errors.Is(err, group.ErrNotMember) {
			return nil, group.ErrNotAdmin
		}
		return nil, err
	}
	if role != group.RoleCreator {
		return nil, group.ErrNotAdmin
	}

	requests, err := r.repo.GetPendingJoinRequests(ctx, q.GroupID)
	if err != nil {
		return nil, err
	}

	result := make([]JoinRequestWithGroup, 0, len(requests))
	for i := range requests {
		g, err := r.repo.GetGroupByID(ctx, requests[i].GroupID)
		if err != nil {
			return nil, err
		}
		result = append(result, JoinRequestWithGroup{Request: requests[i], Group: *g})
	}
	return &GetPendingJoinRequestsResult{Requests: result}, nil
}
