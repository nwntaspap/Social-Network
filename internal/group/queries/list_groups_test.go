package queries

import (
	"context"
	"testing"
	"time"

	"social-network/internal/group"
)

type listGroupsStub struct {
	listFn func(ctx context.Context, page, size int) ([]group.Group, int, error)
	states map[string]string
}

func (s *listGroupsStub) CreateGroup(_ context.Context, _ *group.Group) error {
	return nil
}

func (s *listGroupsStub) GetGroupByID(_ context.Context, _ string) (*group.Group, error) {
	return &group.Group{}, nil
}

func (s *listGroupsStub) ListGroups(ctx context.Context, page, size int) ([]group.Group, int, error) {
	if s.listFn != nil {
		return s.listFn(ctx, page, size)
	}
	return nil, 0, nil
}

func (s *listGroupsStub) UpdateGroup(_ context.Context, _, _, _ string) error {
	return nil
}

func (s *listGroupsStub) DeleteGroup(_ context.Context, _ string) error {
	return nil
}

func (s *listGroupsStub) AddMember(_ context.Context, _, _ string, _ group.Role) error {
	return nil
}

func (s *listGroupsStub) RemoveMember(_ context.Context, _, _ string) error {
	return nil
}

func (s *listGroupsStub) IsMember(_ context.Context, groupID, _ string) (bool, error) {
	return s.states[groupID] == "member", nil
}

func (s *listGroupsStub) GetMemberRole(_ context.Context, _, _ string) (group.Role, error) {
	return group.RoleMember, nil
}

func (s *listGroupsStub) CountMembers(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (s *listGroupsStub) GetGroupMembers(_ context.Context, _ string, _, _ int) ([]group.Member, int, error) {
	return nil, 0, nil
}

func (s *listGroupsStub) CreateInvitation(_ context.Context, _ *group.Invitation) error {
	return nil
}

func (s *listGroupsStub) DeleteInvitation(_ context.Context, _, _ string) error {
	return nil
}

func (s *listGroupsStub) GetInvitation(_ context.Context, _, _ string) (*group.Invitation, error) {
	return &group.Invitation{}, nil
}

func (s *listGroupsStub) IsInvited(_ context.Context, groupID, _ string) (bool, error) {
	return s.states[groupID] == "pending", nil
}

func (s *listGroupsStub) GetPendingInvitations(_ context.Context, _ string) ([]group.Invitation, error) {
	return nil, nil
}

func (s *listGroupsStub) CreateJoinRequest(_ context.Context, _ *group.JoinRequest) error {
	return nil
}

func (s *listGroupsStub) DeleteJoinRequest(_ context.Context, _, _ string) error {
	return nil
}

func (s *listGroupsStub) GetJoinRequest(_ context.Context, _, _ string) (*group.JoinRequest, error) {
	return &group.JoinRequest{}, nil
}

func (s *listGroupsStub) GetJoinRequestByID(_ context.Context, _ string) (*group.JoinRequest, error) {
	return &group.JoinRequest{}, nil
}

func (s *listGroupsStub) HasPendingRequest(_ context.Context, groupID, _ string) (bool, error) {
	return s.states[groupID] == "pending", nil
}

func (s *listGroupsStub) GetPendingJoinRequests(_ context.Context, _ string) ([]group.JoinRequest, error) {
	return nil, nil
}

func (s *listGroupsStub) CreatePost(_ context.Context, _ *group.Post) error {
	return nil
}

func (s *listGroupsStub) GetPostsByGroupID(_ context.Context, _ string, _, _ int) ([]group.Post, int, error) {
	return nil, 0, nil
}

func (s *listGroupsStub) CreatePostComment(_ context.Context, _ *group.PostComment) error {
	return nil
}

func (s *listGroupsStub) GetPostComments(_ context.Context, _ string, _, _ int) ([]group.PostComment, int, error) {
	return nil, 0, nil
}

func (s *listGroupsStub) CountPostComments(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func TestListGroupsResolver_IncludesMembershipStatus(t *testing.T) {
	ctx := context.Background()

	groups := []group.Group{
		{ID: "g1", Title: "Go", CreatorID: "u2", CreatedAt: time.Now()},
		{ID: "g2", Title: "Rust", CreatorID: "u2", CreatedAt: time.Now()},
		{ID: "g3", Title: "Zig", CreatorID: "u2", CreatedAt: time.Now()},
	}

	stub := &listGroupsStub{
		listFn: func(_ context.Context, _, _ int) ([]group.Group, int, error) {
			return groups, len(groups), nil
		},
		states: map[string]string{
			"g1": "member",
			"g2": "pending",
			"g3": "none",
		},
	}

	resolver := NewListGroupsResolver(stub)
	result, err := resolver.Resolve(ctx, ListGroupsQuery{UserID: "u1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}

	statuses := map[string]string{}
	for _, g := range result.Groups {
		statuses[g.ID] = g.MembershipStatus
	}

	want := map[string]string{"g1": "member", "g2": "pending", "g3": "none"}
	for id, wantStatus := range want {
		if statuses[id] != wantStatus {
			t.Errorf("group %s membership status = %q, want %q", id, statuses[id], wantStatus)
		}
	}
}

func TestListGroupsResolver_NoUserIsNone(t *testing.T) {
	ctx := context.Background()

	groups := []group.Group{
		{ID: "g1", Title: "Go", CreatorID: "u2", CreatedAt: time.Now()},
	}

	stub := &listGroupsStub{
		listFn: func(_ context.Context, _, _ int) ([]group.Group, int, error) {
			return groups, len(groups), nil
		},
		states: map[string]string{"g1": "member"},
	}

	resolver := NewListGroupsResolver(stub)
	result, err := resolver.Resolve(ctx, ListGroupsQuery{UserID: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(result.Groups))
	}
	if result.Groups[0].MembershipStatus != "none" {
		t.Errorf("membership status without user = %q, want %q", result.Groups[0].MembershipStatus, "none")
	}
}
