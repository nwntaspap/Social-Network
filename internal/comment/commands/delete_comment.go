package commands

import (
	"context"

	"social-network/internal/comment"
)

type DeleteCommentCommand struct {
	UserID    string
	CommentID int
}

type DeleteCommentHandler struct {
	repo comment.Repository
}

func NewDeleteCommentHandler(repo comment.Repository) *DeleteCommentHandler {
	return &DeleteCommentHandler{repo: repo}
}

func (h *DeleteCommentHandler) Execute(ctx context.Context, cmd DeleteCommentCommand) error {
	if cmd.UserID == "" {
		return ErrEmptyUserID
	}
	if cmd.CommentID == 0 {
		return ErrEmptyCommentID
	}

	return h.repo.DeleteComment(ctx, cmd.UserID, cmd.CommentID)
}
