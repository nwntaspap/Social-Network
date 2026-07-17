package topiccommands

import (
	"context"
	"io"
	"mime/multipart"

	"social-network/internal/app/topics"
	"social-network/internal/domain/topic"
	"social-network/internal/domain/user"
)

const (
	imagepathprefix = "/static/images/uploads/"
)

type CreateTopicRequest struct {
	User        *user.User
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	ImagePath   string     `json:"imagePath"`
	ImageFile   TopicImage `json:"topicImage"`
	CategoryIDs []int      `json:"categoryIds"`
}
type TopicImage struct {
	File   *multipart.File
	Header *multipart.FileHeader
}

type CreateTopicRequestHandler interface {
	Handle(ctx context.Context, req CreateTopicRequest) (*topic.Topic, error)
}

type createTopicRequestHandler struct {
	repo        topic.Repository
	fileStorage topics.FileStorageManager
}

func NewCreateTopicHandler(repo topic.Repository, fileStorage topics.FileStorageManager) CreateTopicRequestHandler {
	return &createTopicRequestHandler{
		repo:        repo,
		fileStorage: fileStorage,
	}
}

func (h *createTopicRequestHandler) Handle(ctx context.Context, req CreateTopicRequest) (*topic.Topic, error) {
	var topic topic.Topic
	if req.ImagePath != "" {
		topic.UserID = req.User.ID
		topic.CategoryIDs = req.CategoryIDs
		topic.Title = req.Title
		topic.Content = req.Content
		topic.ImagePath = imagepathprefix + req.ImagePath
		filecontent, err := io.ReadAll(*req.ImageFile.File)
		if err != nil {
			return nil, err
		}
		_ = h.fileStorage.Upload(ctx, filecontent, req.ImagePath)
	} else {
		topic.UserID = req.User.ID
		topic.CategoryIDs = req.CategoryIDs
		topic.Title = req.Title
		topic.Content = req.Content
	}
	err := h.repo.CreateTopic(ctx, &topic)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}
