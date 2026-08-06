package commands

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"social-network/internal/topic"
)

type CreateTopicCommand struct {
	UserID         string
	Title          string
	Content        string
	ImageData      []byte
	ImageFileName  string
	Visibility     topic.Visibility
	GroupID        *string
	AllowedUserIDs []string
}

type CreateTopicHandler struct {
	repo topic.Repository
	bus  topic.EventBus
	img  topic.ImageStorage
}

func NewCreateTopicHandler(repo topic.Repository, bus topic.EventBus, img topic.ImageStorage) *CreateTopicHandler {
	return &CreateTopicHandler{repo: repo, bus: bus, img: img}
}

type TopicCreatedEvent struct {
	TopicID    int
	UserID     string
	GroupID    *string
	Visibility topic.Visibility
}

func (h *CreateTopicHandler) Execute(ctx context.Context, cmd CreateTopicCommand) (*topic.Topic, error) {
	if cmd.UserID == "" {
		return nil, topic.ErrUnauthorized
	}
	if cmd.Title == "" {
		return nil, errors.New("title is required")
	}
	if cmd.Content == "" {
		return nil, errors.New("content is required")
	}

	t := &topic.Topic{
		UserID:     cmd.UserID,
		Title:      cmd.Title,
		Content:    cmd.Content,
		Visibility: cmd.Visibility,
		GroupID:    cmd.GroupID,
	}

	if len(cmd.ImageData) > 0 && cmd.ImageFileName != "" {
		t.ImagePath = filepath.Join("/static/images/uploads", cmd.ImageFileName)
		if err := h.img.Upload(ctx, cmd.ImageData, cmd.ImageFileName); err != nil {
			return nil, fmt.Errorf("upload image: %w", err)
		}
	}

	if err := h.repo.CreateTopic(ctx, t, cmd.AllowedUserIDs); err != nil {
		return nil, fmt.Errorf("create topic: %w", err)
	}

	_ = h.bus.Publish(ctx, "post.created", TopicCreatedEvent{
		TopicID:    t.ID,
		UserID:     t.UserID,
		GroupID:    t.GroupID,
		Visibility: t.Visibility,
	})

	return t, nil
}
