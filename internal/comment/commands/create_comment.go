package commands

import (
	"context"
	"errors"
	"fmt"

	"social-network/internal/comment"
	"social-network/internal/pkg/imgutil"
)

var (
	ErrEmptyUserID  = errors.New("user ID is required")
	ErrEmptyTopicID = errors.New("topic ID is required")
	ErrEmptyContent = errors.New("content is required")
)

type CreateCommentCommand struct {
	UserID    string
	TopicID   int
	Content   string
	ImageData []byte
}

type CreateCommentHandler struct {
	repo comment.Repository
	bus  comment.EventBus
}

type CommentCreatedEvent struct {
	CommentID int
	TopicID   int
	UserID    string
}

func NewCreateCommentHandler(repo comment.Repository, bus comment.EventBus) *CreateCommentHandler {
	return &CreateCommentHandler{repo: repo, bus: bus}
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

	if len(cmd.ImageData) > 0 {
		if err := imgutil.ValidateImageHeader(cmd.ImageData); err != nil {
			return nil, fmt.Errorf("image validation: %w", err)
		}
	}

	if err := h.repo.CreateComment(ctx, c); err != nil {
		return nil, err
	}

	h.bus.Publish(ctx, "comment.created", CommentCreatedEvent{
		CommentID: c.ID,
		TopicID:   cmd.TopicID,
		UserID:    cmd.UserID,
	})

	return c, nil
}
