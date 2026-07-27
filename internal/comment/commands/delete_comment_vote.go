package commands

import (
	"context"

	"social-network/internal/comment"
)

type DeleteCommentVoteCommand struct {
	UserID    string
	CommentID int
}

type CommentVoteDeletedEvent struct {
	CommentID int
	UserID    string
}

type DeleteCommentVoteHandler struct {
	repo comment.Repository
	bus  comment.EventBus
}

func NewDeleteCommentVoteHandler(repo comment.Repository, bus comment.EventBus) *DeleteCommentVoteHandler {
	return &DeleteCommentVoteHandler{repo: repo, bus: bus}
}

func (h *DeleteCommentVoteHandler) Execute(ctx context.Context, cmd DeleteCommentVoteCommand) error {
	if cmd.UserID == "" {
		return ErrEmptyUserID
	}
	if cmd.CommentID == 0 {
		return ErrEmptyCommentID
	}

	if err := h.repo.DeleteCommentVote(ctx, cmd.UserID, cmd.CommentID); err != nil {
		return err
	}

	_ = h.bus.Publish(ctx, "comment.vote.deleted", CommentVoteDeletedEvent{
		CommentID: cmd.CommentID,
		UserID:    cmd.UserID,
	})

	return nil
}
