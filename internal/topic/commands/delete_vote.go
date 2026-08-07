package commands

import (
	"context"
	"fmt"

	"social-network/internal/topic"
)

type DeleteVoteCommand struct {
	UserID  string
	TopicID int
}

type DeleteVoteHandler struct {
	repo topic.Repository
}

func NewDeleteVoteHandler(repo topic.Repository) *DeleteVoteHandler {
	return &DeleteVoteHandler{repo: repo}
}

func (h *DeleteVoteHandler) Execute(ctx context.Context, cmd DeleteVoteCommand) error {
	if cmd.UserID == "" {
		return topic.ErrUnauthorized
	}
	if cmd.TopicID == 0 {
		return topic.ErrTopicNotFound
	}

	if err := h.repo.DeleteVote(ctx, cmd.UserID, cmd.TopicID); err != nil {
		return fmt.Errorf("delete vote: %w", err)
	}
	return nil
}
