package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"social-network/internal/comment"
	"social-network/internal/pkg/imgutil"
	"social-network/internal/platform/eventbus"
	"social-network/internal/topic"
	"social-network/internal/user"
)

var (
	ErrEmptyUserID  = errors.New("user ID is required")
	ErrEmptyTopicID = errors.New("topic ID is required")
	ErrEmptyContent = errors.New("content is required")
)

type CreateCommentCommand struct {
	UserID        string
	TopicID       int
	Content       string
	ImageData     []byte
	ImageFileName string
}

type CreateCommentHandler struct {
	repo   comment.Repository
	bus    eventbus.EventBus
	users  user.Repository
	topics topic.Repository
	img    comment.ImageStorage
}

func NewCreateCommentHandler(repo comment.Repository, bus eventbus.EventBus, users user.Repository, topics topic.Repository, img comment.ImageStorage) *CreateCommentHandler {
	return &CreateCommentHandler{repo: repo, bus: bus, users: users, topics: topics, img: img}
}

func (h *CreateCommentHandler) Execute(ctx context.Context, cmd CreateCommentCommand) (*comment.Comment, error) {
	if cmd.UserID == "" {
		return nil, ErrEmptyUserID
	}
	if cmd.TopicID == 0 {
		return nil, ErrEmptyTopicID
	}
	if cmd.Content == "" {
		return nil, ErrEmptyContent
	}

	c := &comment.Comment{
		UserID:  cmd.UserID,
		TopicID: cmd.TopicID,
		Content: cmd.Content,
	}

	if len(cmd.ImageData) > 0 && cmd.ImageFileName != "" {
		if err := imgutil.ValidateImageHeader(cmd.ImageData); err != nil {
			return nil, fmt.Errorf("image validation: %w", err)
		}
		c.ImagePath = filepath.Join("/uploads", cmd.ImageFileName)
		if err := h.img.Upload(ctx, cmd.ImageData, cmd.ImageFileName); err != nil {
			return nil, fmt.Errorf("upload image: %w", err)
		}
	}

	if err := h.repo.CreateComment(ctx, c); err != nil {
		return nil, err
	}

	h.notifyCommentCreated(ctx, cmd)

	return c, nil
}

func (h *CreateCommentHandler) notifyCommentCreated(ctx context.Context, cmd CreateCommentCommand) {
	actor, err := h.users.GetByID(ctx, cmd.UserID)
	if err != nil {
		return
	}
	topic, err := h.topics.GetTopicByID(ctx, cmd.TopicID, nil)
	if err != nil || topic == nil {
		return
	}
	body, _ := json.Marshal(eventbus.Notification{
		Type:         eventbus.EventComment,
		RecipientID:  topic.UserID,
		ActorID:      cmd.UserID,
		ActorName:    actor.Nickname,
		ActorAvatar:  actor.AvatarPath,
		ResourceType: eventbus.ResourcePost,
		ResourceID:   strconv.Itoa(cmd.TopicID),
		ContentText:  topic.Content,
		ImageURL:     topic.ImagePath,
	})
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingCreated, body)
}
