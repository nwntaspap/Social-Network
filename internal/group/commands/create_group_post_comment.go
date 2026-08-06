package commands

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"social-network/internal/group"
	"social-network/internal/pkg/imgutil"
	"social-network/internal/pkg/uuid"
)

type CreateGroupPostCommentCommand struct {
	PostID        string
	AuthorID      string
	Content       string
	ImageData     []byte
	ImageFileName string
}

type CreateGroupPostCommentHandler struct {
	repo group.Repository
	img  group.ImageStorage
}

func NewCreateGroupPostCommentHandler(repo group.Repository, img group.ImageStorage) *CreateGroupPostCommentHandler {
	return &CreateGroupPostCommentHandler{repo: repo, img: img}
}

func (h *CreateGroupPostCommentHandler) Execute(ctx context.Context, cmd CreateGroupPostCommentCommand) (*group.PostComment, error) {
	if cmd.PostID == "" || cmd.AuthorID == "" {
		return nil, errors.New("post_id and author_id are required")
	}
	if strings.TrimSpace(cmd.Content) == "" {
		return nil, errors.New("content is required")
	}

	c := &group.PostComment{
		ID:       uuid.NewProvider().NewUUID(),
		PostID:   cmd.PostID,
		AuthorID: cmd.AuthorID,
		Content:  cmd.Content,
	}

	if len(cmd.ImageData) > 0 && cmd.ImageFileName != "" {
		if err := imgutil.ValidateImageHeader(cmd.ImageData); err != nil {
			return nil, err
		}
		c.ImagePath = filepath.Join("/static/images/uploads", cmd.ImageFileName)
		if err := h.img.Upload(ctx, cmd.ImageData, cmd.ImageFileName); err != nil {
			return nil, err
		}
	}

	if err := h.repo.CreatePostComment(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}
