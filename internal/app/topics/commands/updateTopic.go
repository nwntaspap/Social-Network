package topiccommands

import (
	"context"
	"io"
	"strings"

	"social-network/internal/app/topics"
	"social-network/internal/domain/topic"
	"social-network/internal/domain/user"
)

const (
	uploadDir = "/static/images/uploads/"
)

type UpdateTopicRequest struct {
	User         *user.User
	Title        string `json:"title"`
	Content      string `json:"content"`
	ImagePath    string `json:"imagePath"`
	OldImagePath string
	ImageFile    TopicImage `json:"topicImage"`
	CategoryIDs  []int      `json:"categoryIds"`
	TopicID      int        `json:"topicId"`
}

type UpdateTopicRequestHandler interface {
	Handle(ctx context.Context, req UpdateTopicRequest) (*topic.Topic, error)
}

type updateTopicRequestHandler struct {
	repo        topic.Repository
	fileStorage topics.FileStorageManager
}

func NewUpdateTopicHandler(repo topic.Repository, fileStorage topics.FileStorageManager) UpdateTopicRequestHandler {
	return &updateTopicRequestHandler{
		repo:        repo,
		fileStorage: fileStorage,
	}
}

func (h *updateTopicRequestHandler) Handle(ctx context.Context, req UpdateTopicRequest) (*topic.Topic, error) {
	if req.ImagePath != "" {
		return h.UpdateWithImage(ctx, req)
	}

	return h.UpdateWithoutImage(ctx, req)
}

func (h *updateTopicRequestHandler) UpdateWithImage(ctx context.Context, req UpdateTopicRequest) (*topic.Topic, error) {
	var topic topic.Topic
	topic.UserID = req.User.ID
	topic.ID = req.TopicID
	topic.CategoryIDs = req.CategoryIDs
	topic.Title = req.Title
	topic.Content = req.Content
	topic.ImagePath = uploadDir + req.ImagePath

	filecontent, err := io.ReadAll(*req.ImageFile.File)
	if err != nil {
		return nil, err
	}
	if req.OldImagePath != "" {
		_ = h.fileStorage.Delete(ctx, strings.TrimPrefix(req.OldImagePath, uploadDir))
	}
	_ = h.fileStorage.Upload(ctx, filecontent, req.ImagePath)
	err = h.repo.UpdateTopic(ctx, &topic)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (h *updateTopicRequestHandler) UpdateWithoutImage(ctx context.Context, req UpdateTopicRequest) (*topic.Topic, error) {
	var topic topic.Topic
	topic.UserID = req.User.ID
	topic.ID = req.TopicID
	topic.CategoryIDs = req.CategoryIDs
	topic.Title = req.Title
	topic.Content = req.Content
	err := h.repo.UpdateTopic(ctx, &topic)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}
