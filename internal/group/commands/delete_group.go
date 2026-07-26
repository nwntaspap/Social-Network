package commands

import (
	"context"
	"errors"

	"social-network/internal/group"
)

type DeleteGroupCommand struct {
	GroupID string
	UserID  string
}

type DeleteGroupHandler struct {
	repo group.Repository
}

func NewDeleteGroupHandler(repo group.Repository) *DeleteGroupHandler {
	return &DeleteGroupHandler{repo: repo}
}

func (h *DeleteGroupHandler) Execute(ctx context.Context, cmd DeleteGroupCommand) error {
	if cmd.UserID == "" {
		return ErrUserIDRequired
	}
	if cmd.GroupID == "" {
		return ErrGroupIDRequired
	}

	role, err := h.repo.GetMemberRole(ctx, cmd.GroupID, cmd.UserID)
	if err != nil {
		if errors.Is(err, group.ErrNotMember) {
			return group.ErrNotCreator
		}
		return err
	}

	if role != group.RoleCreator {
		return group.ErrNotCreator
	}

	return h.repo.DeleteGroup(ctx, cmd.GroupID)
}
