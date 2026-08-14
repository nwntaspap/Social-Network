package commands

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"social-network/internal/comment"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

var ErrInvalidReactionType = errors.New("reaction_type must be 1 (upvote) or -1 (downvote)")

type CastCommentVoteCommand struct {
	UserID       string
	CommentID    int
	ReactionType int
}

type CastCommentVoteHandler struct {
	repo  comment.Repository
	bus   eventbus.EventBus
	users user.Repository
}

func NewCastCommentVoteHandler(repo comment.Repository, bus eventbus.EventBus, users user.Repository) *CastCommentVoteHandler {
	return &CastCommentVoteHandler{repo: repo, bus: bus, users: users}
}

func (h *CastCommentVoteHandler) Execute(ctx context.Context, cmd CastCommentVoteCommand) error {
	if cmd.UserID == "" {
		return ErrEmptyUserID
	}
	if cmd.CommentID == 0 {
		return ErrEmptyCommentID
	}
	if cmd.ReactionType != 1 && cmd.ReactionType != -1 {
		return ErrInvalidReactionType
	}
	change, err := h.repo.CastCommentVote(ctx, cmd.UserID, cmd.CommentID, cmd.ReactionType)
	if err != nil {
		return err
	}

	if change == comment.VoteChangeRemoved {
		body, _ := json.Marshal(eventbus.Notification{
			Type:         eventbus.EventCommentVoteDeleted,
			ActorID:      cmd.UserID,
			ResourceType: eventbus.ResourceComment,
			ResourceID:   strconv.Itoa(cmd.CommentID),
		})
		_ = h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, body)
		return nil
	}

	actor, err := h.users.GetByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	c, _ := h.repo.GetCommentByID(ctx, cmd.CommentID)

	eventType := eventbus.EventCommentLiked
	if cmd.ReactionType != 1 {
		eventType = eventbus.EventCommentDisliked
	}

	body, _ := json.Marshal(eventbus.Notification{
		Type:         eventType,
		RecipientID:  c.UserID,
		ActorID:      cmd.UserID,
		ActorName:    actor.Nickname,
		ActorAvatar:  actor.AvatarPath,
		ResourceType: eventbus.ResourceComment,
		ResourceID:   strconv.Itoa(cmd.CommentID),
		ContentText:  c.Content,
	})
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingCreated, body)

	return nil
}
