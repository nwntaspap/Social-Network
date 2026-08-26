package commands

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"social-network/internal/group"
	"social-network/internal/pkg/imgutil"
	"social-network/internal/pkg/uuid"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

type CreateGroupPostCommentCommand struct {
	PostID        string
	AuthorID      string
	Content       string
	ImageData     []byte
	ImageFileName string
}

type CreateGroupPostCommentHandler struct {
	repo  group.Repository
	img   group.ImageStorage
	bus   eventbus.EventBus
	users user.Repository
}

func NewCreateGroupPostCommentHandler(repo group.Repository, img group.ImageStorage, bus eventbus.EventBus, users user.Repository) *CreateGroupPostCommentHandler {
	return &CreateGroupPostCommentHandler{repo: repo, img: img, bus: bus, users: users}
}

func (h *CreateGroupPostCommentHandler) Execute(ctx context.Context, cmd CreateGroupPostCommentCommand) (*group.PostComment, error) {
	if cmd.PostID == "" || cmd.AuthorID == "" {
		return nil, errors.New("post_id and author_id are required")
	}
	if strings.TrimSpace(cmd.Content) == "" {
		return nil, errors.New("content is required")
	}

	post, err := h.repo.GetPostByID(ctx, cmd.PostID)
	if err != nil {
		return nil, err
	}
	isMember, err := h.repo.IsMember(ctx, post.GroupID, cmd.AuthorID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, group.ErrNotMember
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
		c.ImagePath = filepath.Join("/uploads", cmd.ImageFileName)
		if err := h.img.Upload(ctx, cmd.ImageData, cmd.ImageFileName); err != nil {
			return nil, err
		}
	}

	if err := h.repo.CreatePostComment(ctx, c); err != nil {
		return nil, err
	}

	if post.AuthorID != cmd.AuthorID {
		actor, _ := h.users.GetByID(ctx, cmd.AuthorID)
		body, _ := json.Marshal(eventbus.Notification{
			Type:         eventbus.EventComment,
			RecipientID:  post.AuthorID,
			ActorID:      actor.ID,
			ActorName:    actor.Nickname,
			ActorAvatar:  actor.AvatarPath,
			ResourceType: eventbus.ResourcePost,
			ResourceID:   cmd.PostID,
			ContentText:  post.Content,
		})
		_ = h.bus.Publish("notifications.exchange", eventbus.RoutingCreated, body)
	}

	return c, nil
}
