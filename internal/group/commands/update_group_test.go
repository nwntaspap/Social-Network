package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
)

type updateGroupStub struct {
	getGroupByIDFn  func(ctx context.Context, groupID string) (*group.Group, error)
	getMemberRoleFn func(ctx context.Context, groupID, userID string) (group.Role, error)
	updateGroupFn   func(ctx context.Context, id, title, description string) error
}

func (s *updateGroupStub) CreateGroup(_ context.Context, _ *group.Group) error { return nil }
func (s *updateGroupStub) GetGroupByID(ctx context.Context, groupID string) (*group.Group, error) {
	if s.getGroupByIDFn != nil {
		return s.getGroupByIDFn(ctx, groupID)
	}
	return &group.Group{ID: groupID, CreatorID: "creator-1"}, nil
}

func (s *updateGroupStub) ListGroups(_ context.Context, _, _ int) ([]group.Group, int, error) {
	return nil, 0, nil
}

func (s *updateGroupStub) UpdateGroup(ctx context.Context, id, title, description string) error {
	if s.updateGroupFn != nil {
		return s.updateGroupFn(ctx, id, title, description)
	}
	return nil
}
func (s *updateGroupStub) DeleteGroup(_ context.Context, _ string) error { return nil }

func (s *updateGroupStub) AddMember(_ context.Context, _, _ string, _ group.Role) error { return nil }

func (s *updateGroupStub) RemoveMember(_ context.Context, _, _ string) error { return nil }

func (s *updateGroupStub) IsMember(_ context.Context, _, _ string) (bool, error) { return true, nil }

func (s *updateGroupStub) GetMemberRole(ctx context.Context, groupID, userID string) (group.Role, error) {
	if s.getMemberRoleFn != nil {
		return s.getMemberRoleFn(ctx, groupID, userID)
	}
	return group.RoleMember, nil
}
func (s *updateGroupStub) CountMembers(_ context.Context, _ string) (int, error) { return 0, nil }
func (s *updateGroupStub) GetGroupMembers(_ context.Context, _ string, _, _ int) ([]group.Member, int, error) {
	return nil, 0, group.ErrGroupNotFound
}

func (s *updateGroupStub) GetGroupAdmins(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func (s *updateGroupStub) CreateInvitation(_ context.Context, _ *group.Invitation) error { return nil }

func (s *updateGroupStub) DeleteInvitation(_ context.Context, _, _ string) error { return nil }

func (s *updateGroupStub) GetInvitation(_ context.Context, _, _ string) (*group.Invitation, error) {
	return nil, group.ErrGroupNotFound
}

func (s *updateGroupStub) IsInvited(_ context.Context, _, _ string) (bool, error) { return false, nil }

func (s *updateGroupStub) GetPendingInvitations(_ context.Context, _ string) ([]group.Invitation, error) {
	return nil, group.ErrGroupNotFound
}

func (s *updateGroupStub) CreateJoinRequest(_ context.Context, _ *group.JoinRequest) error {
	return nil
}
func (s *updateGroupStub) DeleteJoinRequest(_ context.Context, _, _ string) error { return nil }
func (s *updateGroupStub) GetJoinRequest(_ context.Context, _, _ string) (*group.JoinRequest, error) {
	return nil, group.ErrGroupNotFound
}

func (s *updateGroupStub) GetJoinRequestByID(_ context.Context, _ string) (*group.JoinRequest, error) {
	return nil, group.ErrGroupNotFound
}

func (s *updateGroupStub) HasPendingRequest(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (s *updateGroupStub) GetPendingJoinRequests(_ context.Context, _ string) ([]group.JoinRequest, error) {
	return nil, group.ErrGroupNotFound
}
func (s *updateGroupStub) CreatePost(_ context.Context, _ *group.Post) error { return nil }
func (s *updateGroupStub) GetPostsByGroupID(_ context.Context, _ string, _ string, _, _ int) ([]group.Post, int, error) {
	return nil, 0, nil
}

func (s *updateGroupStub) CastPostVote(_ context.Context, _, _ string, _ int) (group.VoteChange, error) {
	return group.VoteChangeAdded, nil
}

func (s *updateGroupStub) GetPostVoteCounts(_ context.Context, _ string) (*group.VoteCounts, error) {
	return &group.VoteCounts{}, nil
}

func (s *updateGroupStub) CreatePostComment(_ context.Context, _ *group.PostComment) error {
	return nil
}

func (s *updateGroupStub) GetPostComments(_ context.Context, _ string, _, _ int) ([]group.PostComment, int, error) {
	return nil, 0, nil
}

func (s *updateGroupStub) CountPostComments(_ context.Context, _ string) (int, error) { return 0, nil }

func (s *updateGroupStub) SendGroupChatMessage(_ context.Context, _ *group.ChatMessage) error {
	return nil
}

func (s *updateGroupStub) GetGroupChatMessages(_ context.Context, _ string, _ int) ([]group.ChatMessage, error) {
	return nil, nil
}

func (s *updateGroupStub) ListGroupMemberIDs(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func TestUpdateGroupHandler_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when user ID is empty", func(t *testing.T) {
		handler := NewUpdateGroupHandler(&updateGroupStub{})
		_, err := handler.Execute(ctx, UpdateGroupCommand{
			UserID:  "",
			GroupID: "group-1",
			Title:   "New Title",
		})
		if err == nil {
			t.Error("expected error for empty user ID")
		}
	})

	t.Run("returns error when group ID is empty", func(t *testing.T) {
		handler := NewUpdateGroupHandler(&updateGroupStub{})
		_, err := handler.Execute(ctx, UpdateGroupCommand{
			UserID:  "user-1",
			GroupID: "",
			Title:   "New Title",
		})
		if err == nil {
			t.Error("expected error for empty group ID")
		}
	})

	t.Run("returns error when title is empty", func(t *testing.T) {
		handler := NewUpdateGroupHandler(&updateGroupStub{})
		_, err := handler.Execute(ctx, UpdateGroupCommand{
			UserID:  "user-1",
			GroupID: "group-1",
			Title:   "",
		})
		if err == nil {
			t.Error("expected error for empty title")
		}
	})

	t.Run("returns error when title is too long", func(t *testing.T) {
		handler := NewUpdateGroupHandler(&updateGroupStub{})
		_, err := handler.Execute(ctx, UpdateGroupCommand{
			UserID:  "user-1",
			GroupID: "group-1",
			Title:   string(make([]byte, 101)),
		})
		if err == nil {
			t.Error("expected error for title exceeding 100 characters")
		}
	})

	t.Run("returns error when description is too long", func(t *testing.T) {
		handler := NewUpdateGroupHandler(&updateGroupStub{})
		_, err := handler.Execute(ctx, UpdateGroupCommand{
			UserID:      "user-1",
			GroupID:     "group-1",
			Title:       "New Title",
			Description: string(make([]byte, 501)),
		})
		if err == nil {
			t.Error("expected error for description exceeding 500 characters")
		}
	})

	t.Run("returns error when user is not admin or creator", func(t *testing.T) {
		stub := &updateGroupStub{
			getMemberRoleFn: func(_ context.Context, _, _ string) (group.Role, error) {
				return group.RoleMember, nil
			},
		}
		handler := NewUpdateGroupHandler(stub)
		_, err := handler.Execute(ctx, UpdateGroupCommand{
			UserID:  "user-1",
			GroupID: "group-1",
			Title:   "New Title",
		})
		if !errors.Is(err, group.ErrNotAdmin) {
			t.Errorf("expected ErrNotAdmin, got %v", err)
		}
	})

	t.Run("admin can update group", func(t *testing.T) {
		updated := false
		stub := &updateGroupStub{
			getMemberRoleFn: func(_ context.Context, _, _ string) (group.Role, error) {
				return group.RoleAdmin, nil
			},
			updateGroupFn: func(_ context.Context, _, _, _ string) error {
				updated = true
				return nil
			},
		}
		handler := NewUpdateGroupHandler(stub)
		g, err := handler.Execute(ctx, UpdateGroupCommand{
			UserID:  "user-1",
			GroupID: "group-1",
			Title:   "New Title",
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !updated {
			t.Error("expected UpdateGroup to be called")
		}
		if g.Title != "New Title" {
			t.Errorf("expected title 'New Title', got %q", g.Title)
		}
	})
}

func (s *updateGroupStub) GetPostByID(_ context.Context, _ string) (*group.Post, error) {
	return nil, group.ErrPostNotFound
}
