package commands

import (
	"context"
	"encoding/json"
	"log"

	"social-network/internal/group"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

type CastGroupPostVoteCommand struct {
	UserID       string
	PostID       string
	ReactionType int
}

type CastGroupPostVoteHandler struct {
	repo  group.PostRepository
	bus   eventbus.EventBus
	users user.Repository
}

func NewCastGroupPostVoteHandler(repo group.PostRepository, bus eventbus.EventBus, users user.Repository) *CastGroupPostVoteHandler {
	return &CastGroupPostVoteHandler{repo: repo, bus: bus, users: users}
}

func (h *CastGroupPostVoteHandler) Execute(ctx context.Context, cmd CastGroupPostVoteCommand) error {
	if cmd.UserID == "" {
		return ErrUserIDRequired
	}
	if cmd.PostID == "" {
		return group.ErrPostNotFound
	}
	if cmd.ReactionType != 1 && cmd.ReactionType != -1 {
		return group.ErrInvalidVoteValue
	}

	post, err := h.repo.GetPostByID(ctx, cmd.PostID)
	if err != nil {
		return err
	}

	change, err := h.repo.CastPostVote(ctx, cmd.UserID, cmd.PostID, cmd.ReactionType)
	if err != nil {
		return err
	}

	if post.AuthorID == cmd.UserID {
		return nil
	}

	actor, err := h.users.GetByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	if change == group.VoteChangeRemoved {
		body, _ := json.Marshal(eventbus.Notification{
			Type:         eventbus.EventPostVoteDeleted,
			ActorID:      actor.ID,
			ResourceType: eventbus.ResourcePost,
			ResourceID:   cmd.PostID,
		})
		_ = h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, body)
		return nil
	}

	delBody, _ := json.Marshal(eventbus.Notification{
		Type:         eventbus.EventPostVoteDeleted,
		ActorID:      actor.ID,
		ResourceType: eventbus.ResourcePost,
		ResourceID:   cmd.PostID,
	})
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, delBody)

	eventType := eventbus.EventPostLiked
	if cmd.ReactionType != 1 {
		eventType = eventbus.EventPostDisliked
	}
	body, _ := json.Marshal(eventbus.Notification{
		Type:         eventType,
		RecipientID:  post.AuthorID,
		ActorID:      actor.ID,
		ActorName:    actor.Nickname,
		ActorAvatar:  actor.AvatarPath,
		ResourceType: eventbus.ResourcePost,
		ResourceID:   cmd.PostID,
		ContentText:  post.Content,
	})
	log.Printf("group post vote notification: type=%s post=%s actor=%s", eventType, cmd.PostID, actor.ID)
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingCreated, body)

	return nil
}
