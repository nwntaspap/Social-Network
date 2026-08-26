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

var (
	ErrPostTitleRequired   = errors.New("post title is required")
	ErrPostContentRequired = errors.New("post content is required")
)

type CreateGroupPostCommand struct {
	GroupID       string
	AuthorID      string
	Title         string
	Content       string
	ImageData     []byte
	ImageFileName string
	ImagePath     string
}

type CreateGroupPostHandler struct {
	repo group.Repository
	img  group.ImageStorage
}

func NewCreateGroupPostHandler(repo group.Repository, img group.ImageStorage) *CreateGroupPostHandler {
	return &CreateGroupPostHandler{repo: repo, img: img}
}

func (h *CreateGroupPostHandler) Execute(ctx context.Context, cmd CreateGroupPostCommand) (*group.Post, error) {
	if cmd.GroupID == "" || cmd.AuthorID == "" {
		return nil, errors.New("group_id and author_id are required")
	}

	isMember, err := h.repo.IsMember(ctx, cmd.GroupID, cmd.AuthorID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, group.ErrNotMember
	}

	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return nil, ErrPostTitleRequired
	}
	if strings.TrimSpace(cmd.Content) == "" {
		return nil, ErrPostContentRequired
	}

	p := &group.Post{
		ID:        uuid.NewProvider().NewUUID(),
		GroupID:   cmd.GroupID,
		AuthorID:  cmd.AuthorID,
		Title:     title,
		Content:   cmd.Content,
		ImagePath: cmd.ImagePath,
	}

	if len(cmd.ImageData) > 0 && cmd.ImageFileName != "" {
		if err := imgutil.ValidateImageHeader(cmd.ImageData); err != nil {
			return nil, err
		}
		p.ImagePath = filepath.Join("/uploads", cmd.ImageFileName)
		if err := h.img.Upload(ctx, cmd.ImageData, cmd.ImageFileName); err != nil {
			return nil, err
		}
	}

	if err := h.repo.CreatePost(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}
