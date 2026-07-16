package commands

import (
	"context"
	"errors"

	"social-network/internal/comment"
)

var ErrEmptyCommentID = errors.New("comment ID is required")

type UpdateCommentCommand struct {
	UserID    string
	CommentID int
	Content   string
}

type UpdateCommentHandler struct {
	repo comment.Repository
}

func NewUpdateCommentHandler(repo comment.Repository) *UpdateCommentHandler {
	return &UpdateCommentHandler{repo: repo}
}

func (h *UpdateCommentHandler) Execute(ctx context.Context, cmd UpdateCommentCommand) error {
	if cmd.UserID == "" {
		return ErrEmptyUserID
	}
	if cmd.CommentID == 0 {
		return ErrEmptyCommentID
	}
	if cmd.Content == "" {
		return ErrEmptyContent
	}

	return h.repo.UpdateComment(ctx, &comment.Comment{
		ID:      cmd.CommentID,
		UserID:  cmd.UserID,
		Content: cmd.Content,
	})
}
