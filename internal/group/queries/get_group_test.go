package queries

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/group"
)

type getGroupStub struct {
	group      *group.Group
	count      int
	countErr   error
	member     bool
	memberErr  error
	invited    bool
	hasPending bool
}

func (s *getGroupStub) CreateGroup(_ context.Context, _ *group.Group) error { return nil }

func (s *getGroupStub) GetGroupByID(_ context.Context, _ string) (*group.Group, error) {
	return s.group, nil
}

func (s *getGroupStub) ListGroups(_ context.Context, _, _ int) ([]group.Group, int, error) {
	return nil, 0, nil
}

func (s *getGroupStub) SearchGroups(_ context.Context, _ string, _, _ int) ([]group.Group, int, error) {
	return nil, 0, nil
}

func (s *getGroupStub) UpdateGroup(_ context.Context, _, _, _ string) error { return nil }

func (s *getGroupStub) DeleteGroup(_ context.Context, _ string) error { return nil }

func (s *getGroupStub) AddMember(_ context.Context, _, _ string, _ group.Role) error { return nil }

func (s *getGroupStub) RemoveMember(_ context.Context, _, _ string) error { return nil }

func (s *getGroupStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	return s.member, s.memberErr
}

func (s *getGroupStub) GetMemberRole(_ context.Context, _, _ string) (group.Role, error) {
	return group.RoleMember, nil
}

func (s *getGroupStub) GetGroupAdmins(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func (s *getGroupStub) CountMembers(_ context.Context, _ string) (int, error) {
	return s.count, s.countErr
}

func (s *getGroupStub) GetGroupMembers(_ context.Context, _ string, _, _ int) ([]group.Member, int, error) {
	return nil, 0, nil
}

func (s *getGroupStub) CreateInvitation(_ context.Context, _ *group.Invitation) error { return nil }

func (s *getGroupStub) DeleteInvitation(_ context.Context, _, _ string) error { return nil }

func (s *getGroupStub) GetInvitation(_ context.Context, _, _ string) (*group.Invitation, error) {
	return &group.Invitation{}, nil
}

func (s *getGroupStub) IsInvited(_ context.Context, _, _ string) (bool, error) {
	return s.invited, nil
}

func (s *getGroupStub) GetPendingInvitations(_ context.Context, _ string) ([]group.Invitation, error) {
	return nil, nil
}

func (s *getGroupStub) CreateJoinRequest(_ context.Context, _ *group.JoinRequest) error { return nil }

func (s *getGroupStub) DeleteJoinRequest(_ context.Context, _, _ string) error { return nil }

func (s *getGroupStub) GetJoinRequest(_ context.Context, _, _ string) (*group.JoinRequest, error) {
	return &group.JoinRequest{}, nil
}

func (s *getGroupStub) GetJoinRequestByID(_ context.Context, _ string) (*group.JoinRequest, error) {
	return &group.JoinRequest{}, nil
}

func (s *getGroupStub) HasPendingRequest(_ context.Context, _, _ string) (bool, error) {
	return s.hasPending, nil
}

func (s *getGroupStub) GetPendingJoinRequests(_ context.Context, _ string) ([]group.JoinRequest, error) {
	return nil, nil
}

func (s *getGroupStub) GetSentInvitationInviteeIDs(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}

func (s *getGroupStub) CreatePost(_ context.Context, _ *group.Post) error { return nil }

func (s *getGroupStub) GetPostsByGroupID(_ context.Context, _, _ string, _, _ int) ([]group.Post, int, error) {
	return nil, 0, nil
}

func (s *getGroupStub) CastPostVote(_ context.Context, _, _ string, _ int) (group.VoteChange, error) {
	return group.VoteChangeAdded, nil
}

func (s *getGroupStub) GetPostVoteCounts(_ context.Context, _ string) (*group.VoteCounts, error) {
	return &group.VoteCounts{}, nil
}

func (s *getGroupStub) CreatePostComment(_ context.Context, _ *group.PostComment) error { return nil }

func (s *getGroupStub) GetPostComments(_ context.Context, _ string, _, _ int) ([]group.PostComment, int, error) {
	return nil, 0, nil
}

func (s *getGroupStub) CountPostComments(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (s *getGroupStub) SendGroupChatMessage(_ context.Context, _ *group.ChatMessage) error {
	return nil
}

func (s *getGroupStub) GetGroupChatMessages(_ context.Context, _ string, _ int) ([]group.ChatMessage, error) {
	return nil, nil
}

func (s *getGroupStub) ListGroupMemberIDs(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func TestGetGroupResolver_IncludesMembersCount(t *testing.T) {
	ctx := context.Background()
	stub := &getGroupStub{
		group:  &group.Group{ID: "g1", Title: "Go", CreatorID: "u2", CreatedAt: time.Now()},
		count:  4,
		member: true,
	}

	resolver := NewGetGroupResolver(stub)
	res, err := resolver.Resolve(ctx, GetGroupQuery{GroupID: "g1", UserID: "u1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if res.Group.ID != "g1" {
		t.Errorf("group ID = %q, want g1", res.Group.ID)
	}
	if res.MembersCount != 4 {
		t.Errorf("MembersCount = %d, want 4", res.MembersCount)
	}
	if res.MembershipStatus != "member" {
		t.Errorf("MembershipStatus = %q, want member", res.MembershipStatus)
	}
}

func TestGetGroupResolver_ForwardsCountError(t *testing.T) {
	ctx := context.Background()
	stub := &getGroupStub{
		group:    &group.Group{ID: "g1", Title: "Go", CreatorID: "u2"},
		countErr: errors.New("count failed"),
	}

	resolver := NewGetGroupResolver(stub)
	_, err := resolver.Resolve(ctx, GetGroupQuery{GroupID: "g1", UserID: "u1"})
	if err == nil {
		t.Fatal("Resolve() expected error from CountMembers")
	}
}

func (s *getGroupStub) GetPostByID(_ context.Context, _ string) (*group.Post, error) {
	return nil, group.ErrPostNotFound
}
