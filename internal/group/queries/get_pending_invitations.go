package queries

import (
	"context"

	"social-network/internal/group"
)

type GetPendingInvitationsQuery struct {
	UserID string
}

type InvitationWithGroup struct {
	Invitation group.Invitation
	Group      group.Group
}

type GetPendingInvitationsResult struct {
	Invitations []InvitationWithGroup
}

type GetPendingInvitationsResolver struct {
	repo pendingInvitationsRepo
}

type pendingInvitationsRepo interface {
	GetPendingInvitations(ctx context.Context, userID string) ([]group.Invitation, error)
	GetGroupByID(ctx context.Context, groupID string) (*group.Group, error)
}

func NewGetPendingInvitationsResolver(repo pendingInvitationsRepo) *GetPendingInvitationsResolver {
	return &GetPendingInvitationsResolver{repo: repo}
}

func (r *GetPendingInvitationsResolver) Resolve(ctx context.Context, q GetPendingInvitationsQuery) (*GetPendingInvitationsResult, error) {
	invs, err := r.repo.GetPendingInvitations(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	result := make([]InvitationWithGroup, 0, len(invs))
	for i := range invs {
		g, err := r.repo.GetGroupByID(ctx, invs[i].GroupID)
		if err != nil {
			return nil, err
		}
		result = append(result, InvitationWithGroup{Invitation: invs[i], Group: *g})
	}
	return &GetPendingInvitationsResult{Invitations: result}, nil
}
