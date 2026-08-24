package commands

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"social-network/internal/pkg/imgutil"
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
	img  topic.ImageStorage
}

func NewCreateTopicHandler(repo topic.Repository, img topic.ImageStorage) *CreateTopicHandler {
	return &CreateTopicHandler{repo: repo, img: img}
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
		if err := imgutil.ValidateImageHeader(cmd.ImageData); err != nil {
			return nil, err
		}
		t.ImagePath = filepath.Join("/static/images/uploads", cmd.ImageFileName)
		if err := h.img.Upload(ctx, cmd.ImageData, cmd.ImageFileName); err != nil {
			return nil, fmt.Errorf("upload image: %w", err)
		}
	}

	if err := h.repo.CreateTopic(ctx, t, cmd.AllowedUserIDs); err != nil {
		return nil, fmt.Errorf("create topic: %w", err)
	}

	return t, nil
}
