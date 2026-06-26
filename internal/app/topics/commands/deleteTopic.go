package topiccommands

import (
	"context"
	"strings"

	"social-network/internal/app/topics"
	"social-network/internal/domain/topic"
	"social-network/internal/domain/user"
)

type DeleteTopicRequest struct {
	User    *user.User
	TopicID int `json:"topicId"`
}

type DeleteTopicRequestHandler interface {
	Handle(ctx context.Context, req DeleteTopicRequest) error
}

type deleteTopicRequestHandler struct {
	repo        topic.Repository
	fileStorage topics.FileStorageManager
}

func NewDeleteTopicHandler(repo topic.Repository, fileStorage topics.FileStorageManager) DeleteTopicRequestHandler {
	return &deleteTopicRequestHandler{
		repo:        repo,
		fileStorage: fileStorage,
	}
}

func (h *deleteTopicRequestHandler) Handle(ctx context.Context, req DeleteTopicRequest) error {
	imagePath, err := h.repo.GetImagePathFromTopicID(ctx, req.TopicID, req.User.ID)
	if err != nil {
		return err
	}
	if imagePath != "" {
		h.fileStorage.Delete(ctx, strings.TrimPrefix(imagePath, uploadDir))
	}
	err = h.repo.DeleteTopic(ctx, req.User.ID, req.TopicID)
	if err != nil {
		return err
	}
	return nil
}
