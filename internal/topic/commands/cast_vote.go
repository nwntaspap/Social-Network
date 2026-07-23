package commands

import (
	"context"
	"fmt"

	"social-network/internal/topic"
)

type CastVoteCommand struct {
	UserID       string
	TopicID      int
	ReactionType int
}

type CastVoteHandler struct {
	repo topic.Repository
	bus  topic.EventBus
}

func NewCastVoteHandler(repo topic.Repository, bus topic.EventBus) *CastVoteHandler {
	return &CastVoteHandler{repo: repo, bus: bus}
}

type PostLikedEvent struct {
	TopicID int
	UserID  string
}

type PostUnlikedEvent struct {
	TopicID int
	UserID  string
}

func (h *CastVoteHandler) Execute(ctx context.Context, cmd CastVoteCommand) error {
	if cmd.UserID == "" {
		return topic.ErrUnauthorized
	}
	if cmd.TopicID == 0 {
		return topic.ErrTopicNotFound
	}
	if cmd.ReactionType != 1 && cmd.ReactionType != -1 {
		return topic.ErrInvalidVoteValue
	}

	vc, err := h.repo.GetVoteCounts(ctx, cmd.TopicID)
	if err != nil {
		return fmt.Errorf("get vote counts: %w", err)
	}
	_ = vc

	if err := h.repo.CastVote(ctx, cmd.UserID, cmd.TopicID, cmd.ReactionType); err != nil {
		return fmt.Errorf("cast vote: %w", err)
	}

	if cmd.ReactionType == 1 {
		_ = h.bus.Publish(ctx, "post.liked", PostLikedEvent{
			TopicID: cmd.TopicID,
			UserID:  cmd.UserID,
		})
	} else {
		_ = h.bus.Publish(ctx, "post.unliked", PostUnlikedEvent{
			TopicID: cmd.TopicID,
			UserID:  cmd.UserID,
		})
	}

	return nil
}
