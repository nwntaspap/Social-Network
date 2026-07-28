package commands

import (
	"context"
	"errors"
	"strings"

	"social-network/internal/group"
	"social-network/internal/pkg/uuid"
)

var (
	ErrTitleRequired   = errors.New("title is required")
	ErrTitleTooLong    = errors.New("title must be 100 characters or fewer")
	ErrDescriptionLong = errors.New("description must be 500 characters or fewer")
	ErrUserIDRequired  = errors.New("user ID is required")
	ErrGroupIDRequired = errors.New("group ID is required")
)

type CreateGroupCommand struct {
	UserID      string
	Title       string
	Description string
}

type CreateGroupHandler struct {
	repo group.Repository
}

func NewCreateGroupHandler(repo group.Repository) *CreateGroupHandler {
	return &CreateGroupHandler{repo: repo}
}

func (h *CreateGroupHandler) Execute(ctx context.Context, cmd CreateGroupCommand) (*group.Group, error) {
	if cmd.UserID == "" {
		return nil, ErrUserIDRequired
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

	g := &group.Group{
		ID:          uuid.NewProvider().NewUUID(),
		Title:       title,
		Description: cmd.Description,
		CreatorID:   cmd.UserID,
	}

	if err := h.repo.CreateGroup(ctx, g); err != nil {
		return nil, err
	}

	if err := h.repo.AddMember(ctx, g.ID, cmd.UserID, group.RoleCreator); err != nil {
		return nil, err
	}

	return g, nil
}
