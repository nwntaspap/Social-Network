package commands

import (
	"context"
	"errors"
	"strings"

	"social-network/internal/group"
)

type UpdateGroupCommand struct {
	GroupID     string
	UserID      string
	Title       string
	Description string
}

type UpdateGroupHandler struct {
	repo group.Repository
}

func NewUpdateGroupHandler(repo group.Repository) *UpdateGroupHandler {
	return &UpdateGroupHandler{repo: repo}
}

func (h *UpdateGroupHandler) Execute(ctx context.Context, cmd UpdateGroupCommand) (*group.Group, error) {
	if cmd.UserID == "" {
		return nil, ErrUserIDRequired
	}
	if cmd.GroupID == "" {
		return nil, ErrGroupIDRequired
	}

	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return nil, ErrTitleRequired
	}
	if len(title) > 100 {
		return nil, ErrTitleTooLong
	}
	if len(cmd.Description) > 500 {
		return nil, ErrDescriptionLong
	}

	g, err := h.repo.GetGroupByID(ctx, cmd.GroupID)
	if err != nil {
		return nil, err
	}

	role, err := h.repo.GetMemberRole(ctx, cmd.GroupID, cmd.UserID)
	if err != nil {
		if errors.Is(err, group.ErrNotMember) {
			return nil, group.ErrNotAdmin
		}
		return nil, err
	}

	if role != group.RoleCreator && role != group.RoleAdmin {
		return nil, group.ErrNotAdmin
	}

	if err := h.repo.UpdateGroup(ctx, cmd.GroupID, title, cmd.Description); err != nil {
		return nil, err
	}

	g.Title = title
	g.Description = cmd.Description
	return g, nil
}
