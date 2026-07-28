package commands

import (
	"context"
	"errors"
	"strings"

	"social-network/internal/group"
	"social-network/internal/pkg/uuid"
)

var (
	ErrPostTitleRequired   = errors.New("post title is required")
	ErrPostContentRequired = errors.New("post content is required")
)

type CreateGroupPostCommand struct {
	GroupID   string
	AuthorID  string
	Title     string
	Content   string
	ImagePath string
}

type CreateGroupPostHandler struct {
	repo group.Repository
}

func NewCreateGroupPostHandler(repo group.Repository) *CreateGroupPostHandler {
	return &CreateGroupPostHandler{repo: repo}
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

	if err := h.repo.CreatePost(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}
