package queries

import (
	"context"
)

type GetSentInvitationIDsQuery struct {
	GroupID   string
	InviterID string
}

type GetSentInvitationIDsResult struct {
	InviteeIDs []string
}

type GetSentInvitationIDsResolver struct {
	repo sentInvitationIDsRepo
}

type sentInvitationIDsRepo interface {
	GetSentInvitationInviteeIDs(ctx context.Context, groupID, inviterID string) ([]string, error)
}

func NewGetSentInvitationIDsResolver(repo sentInvitationIDsRepo) *GetSentInvitationIDsResolver {
	return &GetSentInvitationIDsResolver{repo: repo}
}

func (r *GetSentInvitationIDsResolver) Resolve(ctx context.Context, q GetSentInvitationIDsQuery) (*GetSentInvitationIDsResult, error) {
	ids, err := r.repo.GetSentInvitationInviteeIDs(ctx, q.GroupID, q.InviterID)
	if err != nil {
		return nil, err
	}
	return &GetSentInvitationIDsResult{InviteeIDs: ids}, nil
}
