package commands

import (
	"context"
	"errors"

	"social-network/internal/comment"
)

var ErrInvalidReactionType = errors.New("reaction_type must be 1 (upvote) or -1 (downvote)")

type CastCommentVoteCommand struct {
	UserID       string
	CommentID    int
	ReactionType int
}

type CastCommentVoteHandler struct {
	repo comment.Repository
}

func NewCastCommentVoteHandler(repo comment.Repository) *CastCommentVoteHandler {
	return &CastCommentVoteHandler{repo: repo}
}

func (h *CastCommentVoteHandler) Execute(ctx context.Context, cmd CastCommentVoteCommand) error {
	if cmd.UserID == "" {
		return ErrEmptyUserID
	}
	if cmd.CommentID == 0 {
		return ErrEmptyCommentID
	}
	if cmd.ReactionType != 1 && cmd.ReactionType != -1 {
		return ErrInvalidReactionType
	}
	return h.repo.CastCommentVote(ctx, cmd.UserID, cmd.CommentID, cmd.ReactionType)
}
