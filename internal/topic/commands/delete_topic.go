package commands

import (
	"context"
	"fmt"

	"social-network/internal/topic"
)

type DeleteTopicCommand struct {
	TopicID int
	UserID  string
}

type DeleteTopicHandler struct {
	repo topic.Repository
	bus  topic.EventBus
	img  topic.ImageStorage
}

func NewDeleteTopicHandler(repo topic.Repository, bus topic.EventBus, img topic.ImageStorage) *DeleteTopicHandler {
	return &DeleteTopicHandler{repo: repo, bus: bus, img: img}
}

type TopicDeletedEvent struct {
	TopicID int
	UserID  string
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

	_ = h.bus.Publish(ctx, "post.deleted", TopicDeletedEvent(cmd))

	return nil
}
