package commands

import (
	"context"
	"encoding/json"
	"strconv"

	"social-network/internal/comment"
	"social-network/internal/platform/eventbus"
)

type DeleteCommentVoteCommand struct {
	UserID    string
	CommentID int
}

type CommentVoteDeletedEvent struct {
	CommentID int    `json:"comment_id"`
	UserID    string `json:"user_id"`
}

type DeleteCommentVoteHandler struct {
	repo comment.Repository
	bus  eventbus.EventBus
}

func NewDeleteCommentVoteHandler(repo comment.Repository, bus eventbus.EventBus) *DeleteCommentVoteHandler {
	return &DeleteCommentVoteHandler{repo: repo, bus: bus}
}

func (h *DeleteCommentVoteHandler) Execute(ctx context.Context, cmd DeleteCommentVoteCommand) error {
	if cmd.UserID == "" {
		return ErrEmptyUserID
	}
	if cmd.CommentID == 0 {
		return ErrEmptyCommentID
	}

	if err := h.repo.DeleteCommentVote(ctx, cmd.UserID, cmd.CommentID); err != nil {
		return err
	}
	body, _ := json.Marshal(eventbus.Notification{
		Type:       eventbus.EventCommentVoteDeleted,
		ResourceID: strconv.Itoa(cmd.CommentID),
	})
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, body)

	return nil
}
