package commands

import (
	"context"
	"encoding/json"
	"strconv"

	"social-network/internal/comment"
	"social-network/internal/platform/eventbus"
	"social-network/internal/topic"
	"social-network/internal/user"
)

type DeleteCommentCommand struct {
	UserID    string
	CommentID int
}

type DeleteCommentHandler struct {
	repo  comment.Repository
	bus   eventbus.EventBus
	topic topic.Repository
	users user.Repository
}

func NewDeleteCommentHandler(repo comment.Repository, bus eventbus.EventBus, topic topic.Repository, user user.Repository) *DeleteCommentHandler {
	return &DeleteCommentHandler{
		repo:  repo,
		bus:   bus,
		topic: topic,
		users: user,
	}
}

func (h *DeleteCommentHandler) Execute(ctx context.Context, cmd DeleteCommentCommand) error {
	if cmd.UserID == "" {
		return ErrEmptyUserID
	}
	if cmd.CommentID == 0 {
		return ErrEmptyCommentID
	}
	c, _ := h.repo.GetCommentByID(ctx, cmd.CommentID)
	actor, _ := h.users.GetByID(ctx, cmd.UserID)

	topic, _ := h.topic.GetTopicByID(ctx, c.TopicID, nil)
	body, _ := json.Marshal(eventbus.Notification{
		Type:         eventbus.EventComment,
		ActorID:      actor.ID,
		ResourceType: eventbus.ResourcePost,
		ResourceID:   strconv.Itoa(topic.ID),
	})
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, body)

	return h.repo.DeleteComment(ctx, cmd.UserID, cmd.CommentID)
}
