package commands

import (
	"context"
	"errors"

	"social-network/internal/group"
)

type MarkGroupReadCommand struct {
	GroupID string
	UserID  string
}

// markGroupReadRepository is the slice of group.Repository the handler needs.
type markGroupReadRepository interface {
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
	MarkGroupRead(ctx context.Context, groupID, userID string) error
}

type MarkGroupReadHandler struct {
	repo markGroupReadRepository
}

func NewMarkGroupReadHandler(repo markGroupReadRepository) *MarkGroupReadHandler {
	return &MarkGroupReadHandler{repo: repo}
}

// Execute marks every group chat message up to now as read for the user.
// Non-members cannot mark a group read.
func (h *MarkGroupReadHandler) Execute(ctx context.Context, cmd MarkGroupReadCommand) error {
	if cmd.GroupID == "" || cmd.UserID == "" {
		return errors.New("group_id and user_id are required")
	}

	isMember, err := h.repo.IsMember(ctx, cmd.GroupID, cmd.UserID)
	if err != nil {
		return err
	}
	if !isMember {
		return group.ErrNotMember
	}

	return h.repo.MarkGroupRead(ctx, cmd.GroupID, cmd.UserID)
}
