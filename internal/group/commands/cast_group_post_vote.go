package commands

import (
	"context"

	"social-network/internal/group"
)

type CastGroupPostVoteCommand struct {
	UserID       string
	PostID       string
	ReactionType int
}

type CastGroupPostVoteHandler struct {
	repo group.PostRepository
}

func NewCastGroupPostVoteHandler(repo group.PostRepository) *CastGroupPostVoteHandler {
	return &CastGroupPostVoteHandler{repo: repo}
}

func (h *CastGroupPostVoteHandler) Execute(ctx context.Context, cmd CastGroupPostVoteCommand) error {
	if cmd.UserID == "" {
		return ErrUserIDRequired
	}
	if cmd.PostID == "" {
		return group.ErrPostNotFound
	}
	if cmd.ReactionType != 1 && cmd.ReactionType != -1 {
		return group.ErrInvalidVoteValue
	}

	return h.repo.CastPostVote(ctx, cmd.UserID, cmd.PostID, cmd.ReactionType)
}
