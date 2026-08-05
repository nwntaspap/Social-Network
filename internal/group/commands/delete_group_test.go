package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
	"social-network/internal/platform/eventbus"
)

type deleteGroupStub struct {
	getGroupByIDFn  func(ctx context.Context, groupID string) (*group.Group, error)
	getMemberRoleFn func(ctx context.Context, groupID, userID string) (group.Role, error)
	deleteGroupFn   func(ctx context.Context, id string) error
}

func (s *deleteGroupStub) CreateGroup(_ context.Context, _ *group.Group) error { return nil }
func (s *deleteGroupStub) GetGroupByID(ctx context.Context, groupID string) (*group.Group, error) {
	if s.getGroupByIDFn != nil {
		return s.getGroupByIDFn(ctx, groupID)
	}
	return &group.Group{ID: groupID, CreatorID: "creator-1"}, nil
}

func (s *deleteGroupStub) ListGroups(_ context.Context, _, _ int) ([]group.Group, int, error) {
	return nil, 0, nil
}
func (s *deleteGroupStub) UpdateGroup(_ context.Context, _, _, _ string) error { return nil }
func (s *deleteGroupStub) DeleteGroup(ctx context.Context, id string) error {
	if s.deleteGroupFn != nil {
		return s.deleteGroupFn(ctx, id)
	}
	return nil
}

func (s *deleteGroupStub) AddMember(_ context.Context, _, _ string, _ group.Role) error { return nil }

func (s *deleteGroupStub) RemoveMember(_ context.Context, _, _ string) error { return nil }

func (s *deleteGroupStub) IsMember(_ context.Context, _, _ string) (bool, error) { return true, nil }

func (s *deleteGroupStub) GetMemberRole(ctx context.Context, groupID, userID string) (group.Role, error) {
	if s.getMemberRoleFn != nil {
		return s.getMemberRoleFn(ctx, groupID, userID)
	}
	return group.RoleCreator, nil
}
func (s *deleteGroupStub) CountMembers(_ context.Context, _ string) (int, error) { return 0, nil }
func (s *deleteGroupStub) GetGroupMembers(_ context.Context, _ string, _, _ int) ([]group.Member, int, error) {
	return nil, 0, group.ErrGroupNotFound
}

func (s *deleteGroupStub) GetGroupAdmins(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

type mockBus struct{}

func (m *mockBus) Publish(_ string, _ string, _ []byte) error { return nil }

func (m *mockBus) Subscribe(_ context.Context, _ string) (<-chan eventbus.Message, error) {
	ch := make(chan eventbus.Message)
	close(ch)
	return ch, nil
}

func (m *mockBus) InitTopology(_ context.Context) error { return nil }

func (s *deleteGroupStub) CreateInvitation(_ context.Context, _ *group.Invitation) error { return nil }

func (s *deleteGroupStub) DeleteInvitation(_ context.Context, _, _ string) error { return nil }

func (s *deleteGroupStub) GetInvitation(_ context.Context, _, _ string) (*group.Invitation, error) {
	return nil, group.ErrGroupNotFound
}

func (s *deleteGroupStub) IsInvited(_ context.Context, _, _ string) (bool, error) { return false, nil }

func (s *deleteGroupStub) GetPendingInvitations(_ context.Context, _ string) ([]group.Invitation, error) {
	return nil, nil
}

func (s *deleteGroupStub) CreateJoinRequest(_ context.Context, _ *group.JoinRequest) error {
	return nil
}
func (s *deleteGroupStub) DeleteJoinRequest(_ context.Context, _, _ string) error { return nil }
func (s *deleteGroupStub) GetJoinRequest(_ context.Context, _, _ string) (*group.JoinRequest, error) {
	return nil, group.ErrGroupNotFound
}

func (s *deleteGroupStub) GetJoinRequestByID(_ context.Context, _ string) (*group.JoinRequest, error) {
	return nil, group.ErrGroupNotFound
}

func (s *deleteGroupStub) HasPendingRequest(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (s *deleteGroupStub) GetPendingJoinRequests(_ context.Context, _ string) ([]group.JoinRequest, error) {
	return nil, nil
}
func (s *deleteGroupStub) CreatePost(_ context.Context, _ *group.Post) error { return nil }
func (s *deleteGroupStub) GetPostsByGroupID(_ context.Context, _ string, _ string, _, _ int) ([]group.Post, int, error) {
	return nil, 0, nil
}
func (s *deleteGroupStub) CastPostVote(_ context.Context, _, _ string, _ int) error { return nil }
func (s *deleteGroupStub) GetPostVoteCounts(_ context.Context, _ string) (*group.VoteCounts, error) {
	return &group.VoteCounts{}, nil
}

func (s *deleteGroupStub) CreatePostComment(_ context.Context, _ *group.PostComment) error {
	return nil
}

func (s *deleteGroupStub) GetPostComments(_ context.Context, _ string, _, _ int) ([]group.PostComment, int, error) {
	return nil, 0, nil
}

func (s *deleteGroupStub) CountPostComments(_ context.Context, _ string) (int, error) { return 0, nil }

func (s *deleteGroupStub) SendGroupChatMessage(_ context.Context, _ *group.ChatMessage) error {
	return nil
}

func (s *deleteGroupStub) GetGroupChatMessages(_ context.Context, _ string, _ int) ([]group.ChatMessage, error) {
	return nil, nil
}

func (s *deleteGroupStub) ListGroupMemberIDs(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func TestDeleteGroupHandler_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when user ID is empty", func(t *testing.T) {
		handler := NewDeleteGroupHandler(&deleteGroupStub{}, &mockBus{})
		err := handler.Execute(ctx, DeleteGroupCommand{
			UserID:  "",
			GroupID: "group-1",
		})
		if err == nil {
			t.Error("expected error for empty user ID")
		}
	})

	t.Run("returns error when group ID is empty", func(t *testing.T) {
		handler := NewDeleteGroupHandler(&deleteGroupStub{}, &mockBus{})
		err := handler.Execute(ctx, DeleteGroupCommand{
			UserID:  "user-1",
			GroupID: "",
		})
		if err == nil {
			t.Error("expected error for empty group ID")
		}
	})

	t.Run("returns error when user is not creator", func(t *testing.T) {
		stub := &deleteGroupStub{
			getMemberRoleFn: func(_ context.Context, _, _ string) (group.Role, error) {
				return group.RoleAdmin, nil
			},
		}
		handler := NewDeleteGroupHandler(stub, &mockBus{})
		err := handler.Execute(ctx, DeleteGroupCommand{
			UserID:  "user-1",
			GroupID: "group-1",
		})
		if !errors.Is(err, group.ErrNotCreator) {
			t.Errorf("expected ErrNotCreator, got %v", err)
		}
	})

	t.Run("creator can delete group", func(t *testing.T) {
		deleted := false
		stub := &deleteGroupStub{
			getMemberRoleFn: func(_ context.Context, _, _ string) (group.Role, error) {
				return group.RoleCreator, nil
			},
			deleteGroupFn: func(_ context.Context, _ string) error {
				deleted = true
				return nil
			},
		}
		handler := NewDeleteGroupHandler(stub, &mockBus{})
		err := handler.Execute(ctx, DeleteGroupCommand{
			UserID:  "creator-1",
			GroupID: "group-1",
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !deleted {
			t.Error("expected DeleteGroup to be called")
		}
	})
}
