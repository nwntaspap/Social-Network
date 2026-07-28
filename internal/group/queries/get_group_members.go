package queries

import (
	"context"

	"social-network/internal/group"
)

type GetGroupMembersQuery struct {
	GroupID string
	Page    int
	Size    int
}

type GetGroupMembersResult struct {
	Members []group.Member
	Total   int
}

type GetGroupMembersResolver struct {
	repo group.MemberRepository
}

func NewGetGroupMembersResolver(repo group.MemberRepository) *GetGroupMembersResolver {
	return &GetGroupMembersResolver{repo: repo}
}

func (r *GetGroupMembersResolver) Resolve(ctx context.Context, q GetGroupMembersQuery) (*GetGroupMembersResult, error) {
	members, total, err := r.repo.GetGroupMembers(ctx, q.GroupID, q.Page, q.Size)
	if err != nil {
		return nil, err
	}
	return &GetGroupMembersResult{Members: members, Total: total}, nil
}
