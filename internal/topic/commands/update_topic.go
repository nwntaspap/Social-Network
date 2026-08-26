package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"social-network/internal/pkg/imgutil"
	"social-network/internal/topic"
)

type UpdateTopicCommand struct {
	TopicID        int
	UserID         string
	Title          string
	Content        string
	ImageData      []byte
	ImageFileName  string
	Visibility     topic.Visibility
	AllowedUserIDs []string
}

type UpdateTopicHandler struct {
	repo topic.Repository
	img  topic.ImageStorage
}

func NewUpdateTopicHandler(repo topic.Repository, img topic.ImageStorage) *UpdateTopicHandler {
	return &UpdateTopicHandler{repo: repo, img: img}
}

func (h *UpdateTopicHandler) Execute(ctx context.Context, cmd UpdateTopicCommand) (*topic.Topic, error) {
	if cmd.UserID == "" {
		return nil, topic.ErrUnauthorized
	}
	if cmd.TopicID == 0 {
		return nil, topic.ErrTopicNotFound
	}

	t := &topic.Topic{
		ID:         cmd.TopicID,
		UserID:     cmd.UserID,
		Title:      cmd.Title,
		Content:    cmd.Content,
		Visibility: cmd.Visibility,
	}

	if len(cmd.ImageData) > 0 && cmd.ImageFileName != "" {
		if err := imgutil.ValidateImageHeader(cmd.ImageData); err != nil {
			return nil, err
		}
		oldPath, _ := h.repo.GetImagePathFromTopicID(ctx, cmd.TopicID, cmd.UserID)
		if oldPath != "" {
			_ = h.img.Delete(ctx, oldPath)
		}
		t.ImagePath = filepath.Join("/uploads", cmd.ImageFileName)
		if err := h.img.Upload(ctx, cmd.ImageData, cmd.ImageFileName); err != nil {
			return nil, fmt.Errorf("upload image: %w", err)
		}
	}

	if err := h.repo.UpdateTopic(ctx, t, cmd.AllowedUserIDs); err != nil {
		return nil, fmt.Errorf("update topic: %w", err)
	}

	return t, nil
}
