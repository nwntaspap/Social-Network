package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"social-network/internal/platform/eventbus"
	"social-network/internal/topic"
	"social-network/internal/user"
)

type CastVoteCommand struct {
	UserID       string
	TopicID      int
	ReactionType int
}

type CastVoteHandler struct {
	repo  topic.Repository
	bus   eventbus.EventBus
	users user.Repository
}

func NewCastVoteHandler(repo topic.Repository, bus eventbus.EventBus, users user.Repository) *CastVoteHandler {
	return &CastVoteHandler{repo: repo, bus: bus, users: users}
}

func (h *CastVoteHandler) Execute(ctx context.Context, cmd CastVoteCommand) error {
	if cmd.UserID == "" {
		return topic.ErrUnauthorized
	}
	if cmd.TopicID == 0 {
		return topic.ErrTopicNotFound
	}
	if cmd.ReactionType != 1 && cmd.ReactionType != -1 {
		return topic.ErrInvalidVoteValue
	}

	_, err := h.repo.GetVoteCounts(ctx, cmd.TopicID)
	if err != nil {
		return fmt.Errorf("get vote counts: %w", err)
	}

	change, err := h.repo.CastVote(ctx, cmd.UserID, cmd.TopicID, cmd.ReactionType)
	if err != nil {
		return fmt.Errorf("cast vote: %w", err)
	}

	actor, err := h.users.GetByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	if change == topic.VoteChangeRemoved {
		body, _ := json.Marshal(eventbus.Notification{
			Type:         eventbus.EventPostVoteDeleted,
			ActorID:      actor.ID,
			ResourceType: eventbus.ResourcePost,
			ResourceID:   strconv.Itoa(cmd.TopicID),
		})
		_ = h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, body)
		return nil
	}

	t, err := h.repo.GetTopicByID(ctx, cmd.TopicID, &cmd.UserID)
	if err != nil {
		return err
	}

	// Remove any stale vote notification (like ↔ dislike) for this actor on
	// this post before creating the new one, so the recipient never sees both.
	delBody, _ := json.Marshal(eventbus.Notification{
		Type:         eventbus.EventPostVoteDeleted,
		ActorID:      actor.ID,
		ResourceType: eventbus.ResourcePost,
		ResourceID:   strconv.Itoa(cmd.TopicID),
	})
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, delBody)

	eventType := eventbus.EventPostLiked
	if cmd.ReactionType != 1 {
		eventType = eventbus.EventPostDisliked
	}
	body, _ := json.Marshal(eventbus.Notification{
		Type:         eventType,
		RecipientID:  t.UserID,
		ActorID:      actor.ID,
		ActorName:    actor.Nickname,
		ActorAvatar:  actor.AvatarPath,
		ResourceType: eventbus.ResourcePost,
		ResourceID:   strconv.Itoa(cmd.TopicID),
		ContentText:  t.Content,
		ImageURL:     t.ImagePath,
	})
	log.Printf("this is the event Type:%s", eventType)
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingCreated, body)

	return nil
}
