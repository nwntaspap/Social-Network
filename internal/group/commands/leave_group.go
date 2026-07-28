package commands

import (
	"context"
	"errors"

	"social-network/internal/group"
)

type LeaveGroupCommand struct {
	GroupID string
	UserID  string
}

type LeaveGroupHandler struct {
	memberRepo group.MemberRepository
}

func NewLeaveGroupHandler(memberRepo group.MemberRepository) *LeaveGroupHandler {
	return &LeaveGroupHandler{memberRepo: memberRepo}
}

func (h *LeaveGroupHandler) Execute(ctx context.Context, cmd LeaveGroupCommand) error {
	if cmd.UserID == "" {
		return ErrUserIDRequired
	}
	if cmd.GroupID == "" {
		return ErrGroupIDRequired
	}

	role, err := h.memberRepo.GetMemberRole(ctx, cmd.GroupID, cmd.UserID)
	if err != nil {
		if errors.Is(err, group.ErrNotMember) {
			return err
		}
		return err
	}

	if role == group.RoleCreator {
		return group.ErrNotCreator
	}

	return h.memberRepo.RemoveMember(ctx, cmd.GroupID, cmd.UserID)
}
