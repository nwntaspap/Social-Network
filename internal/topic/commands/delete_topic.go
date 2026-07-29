package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"social-network/internal/platform/eventbus"
	"social-network/internal/topic"
)

type DeleteTopicCommand struct {
	TopicID int
	UserID  string
}

type DeleteTopicHandler struct {
	repo topic.Repository
	bus  eventbus.EventBus
	img  topic.ImageStorage
}

func NewDeleteTopicHandler(repo topic.Repository, bus eventbus.EventBus, img topic.ImageStorage) *DeleteTopicHandler {
	return &DeleteTopicHandler{repo: repo, bus: bus, img: img}
}

func (h *DeleteTopicHandler) Execute(ctx context.Context, cmd DeleteTopicCommand) error {
	if cmd.UserID == "" {
		return topic.ErrUnauthorized
	}
	if cmd.TopicID == 0 {
		return topic.ErrTopicNotFound
	}

	imagePath, _ := h.repo.GetImagePathFromTopicID(ctx, cmd.TopicID, cmd.UserID)
	if imagePath != "" {
		_ = h.img.Delete(ctx, imagePath)
	}

	if err := h.repo.DeleteTopic(ctx, cmd.UserID, cmd.TopicID); err != nil {
		return fmt.Errorf("delete topic: %w", err)
	}

	body, _ := json.Marshal(eventbus.Envelope{
		Type:         "post.deleted",
		ResourceType: "post",
		ResourceID:   cmd.TopicID,
	})
	_ = h.bus.Publish("notifications.exchange", "post.deleted", body)

	return nil
}
