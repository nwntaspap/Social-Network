package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
)

type leaveGroupStub struct {
	getMemberRoleFn func(ctx context.Context, groupID, userID string) (group.Role, error)
	removeMemberFn  func(ctx context.Context, groupID, userID string) error
}

func (s *leaveGroupStub) AddMember(_ context.Context, _, _ string, _ group.Role) error {
	return nil
}

func (s *leaveGroupStub) RemoveMember(ctx context.Context, groupID, userID string) error {
	if s.removeMemberFn != nil {
		return s.removeMemberFn(ctx, groupID, userID)
	}
	return nil
}

func (s *leaveGroupStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}

func (s *leaveGroupStub) GetMemberRole(ctx context.Context, groupID, userID string) (group.Role, error) {
	if s.getMemberRoleFn != nil {
		return s.getMemberRoleFn(ctx, groupID, userID)
	}
	return group.RoleMember, nil
}

func (s *leaveGroupStub) CountMembers(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (s *leaveGroupStub) GetGroupAdmins(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func (s *leaveGroupStub) GetGroupMembers(_ context.Context, _ string, _, _ int) ([]group.Member, int, error) {
	return nil, 0, nil
}

func TestLeaveGroupHandler_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when user ID is empty", func(t *testing.T) {
		handler := NewLeaveGroupHandler(&leaveGroupStub{})
		err := handler.Execute(ctx, LeaveGroupCommand{
			UserID:  "",
			GroupID: "group-1",
		})
		if err == nil {
			t.Error("expected error for empty user ID")
		}
	})

	t.Run("returns error when group ID is empty", func(t *testing.T) {
		handler := NewLeaveGroupHandler(&leaveGroupStub{})
		err := handler.Execute(ctx, LeaveGroupCommand{
			UserID:  "user-1",
			GroupID: "",
		})
		if err == nil {
			t.Error("expected error for empty group ID")
		}
	})

	t.Run("returns error when user is creator", func(t *testing.T) {
		stub := &leaveGroupStub{
			getMemberRoleFn: func(_ context.Context, _, _ string) (group.Role, error) {
				return group.RoleCreator, nil
			},
		}
		handler := NewLeaveGroupHandler(stub)
		err := handler.Execute(ctx, LeaveGroupCommand{
			UserID:  "user-1",
			GroupID: "group-1",
		})
		if !errors.Is(err, group.ErrNotCreator) {
			t.Errorf("expected ErrNotCreator, got %v", err)
		}
	})

	t.Run("returns error when user is not a member", func(t *testing.T) {
		stub := &leaveGroupStub{
			getMemberRoleFn: func(_ context.Context, _, _ string) (group.Role, error) {
				return "", group.ErrNotMember
			},
		}
		handler := NewLeaveGroupHandler(stub)
		err := handler.Execute(ctx, LeaveGroupCommand{
			UserID:  "user-1",
			GroupID: "group-1",
		})
		if !errors.Is(err, group.ErrNotMember) {
			t.Errorf("expected ErrNotMember, got %v", err)
		}
	})

	t.Run("successfully removes member", func(t *testing.T) {
		removed := false
		stub := &leaveGroupStub{
			getMemberRoleFn: func(_ context.Context, _, _ string) (group.Role, error) {
				return group.RoleMember, nil
			},
			removeMemberFn: func(_ context.Context, _, _ string) error {
				removed = true
				return nil
			},
		}
		handler := NewLeaveGroupHandler(stub)
		err := handler.Execute(ctx, LeaveGroupCommand{
			UserID:  "user-1",
			GroupID: "group-1",
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !removed {
			t.Error("expected RemoveMember to be called")
		}
	})

	t.Run("admin can leave", func(t *testing.T) {
		removed := false
		stub := &leaveGroupStub{
			getMemberRoleFn: func(_ context.Context, _, _ string) (group.Role, error) {
				return group.RoleAdmin, nil
			},
			removeMemberFn: func(_ context.Context, _, _ string) error {
				removed = true
				return nil
			},
		}
		handler := NewLeaveGroupHandler(stub)
		err := handler.Execute(ctx, LeaveGroupCommand{
			UserID:  "user-1",
			GroupID: "group-1",
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !removed {
			t.Error("expected RemoveMember to be called for admin")
		}
	})
}
