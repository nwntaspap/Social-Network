package commands

import (
	"context"
	"errors"
	"strings"

	"social-network/internal/group"
	"social-network/internal/pkg/uuid"
)

type CreateGroupPostCommentCommand struct {
	PostID    string
	AuthorID  string
	Content   string
	ImagePath string
}

type CreateGroupPostCommentHandler struct {
	repo group.Repository
}

func NewCreateGroupPostCommentHandler(repo group.Repository) *CreateGroupPostCommentHandler {
	return &CreateGroupPostCommentHandler{repo: repo}
}

func (h *CreateGroupPostCommentHandler) Execute(ctx context.Context, cmd CreateGroupPostCommentCommand) (*group.PostComment, error) {
	if cmd.PostID == "" || cmd.AuthorID == "" {
		return nil, errors.New("post_id and author_id are required")
	}
	if strings.TrimSpace(cmd.Content) == "" {
		return nil, errors.New("content is required")
	}

	c := &group.PostComment{
		ID:        uuid.NewProvider().NewUUID(),
		PostID:    cmd.PostID,
		AuthorID:  cmd.AuthorID,
		Content:   cmd.Content,
		ImagePath: cmd.ImagePath,
	}

	if err := h.repo.CreatePostComment(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}
