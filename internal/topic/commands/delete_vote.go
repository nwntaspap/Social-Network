package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"social-network/internal/platform/eventbus"
	"social-network/internal/topic"
)

type DeleteVoteCommand struct {
	UserID  string
	TopicID int
}

type DeleteVoteHandler struct {
	repo topic.Repository
	bus  eventbus.EventBus
}

func NewDeleteVoteHandler(repo topic.Repository, bus eventbus.EventBus) *DeleteVoteHandler {
	return &DeleteVoteHandler{repo: repo, bus: bus}
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

	body, _ := json.Marshal(eventbus.Envelope{
		Type:         "post.liked.deleted",
		ResourceType: "post",
		ResourceID:   cmd.TopicID,
	})
	_ = h.bus.Publish("notifications.exchange", "post.liked.deleted", body)

	return nil
}
