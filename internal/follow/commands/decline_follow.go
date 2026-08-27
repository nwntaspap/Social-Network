package commands

import (
	"context"
	"encoding/json"

	"social-network/internal/follow"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

type DeclineRequestCommand struct {
	FollowerID string
	FolloweeID string
}

type DeclineRequestHandler struct {
	repo  follow.Repository
	bus   eventbus.EventBus
	users user.Repository
}

func NewDeclineRequestHandler(repo follow.Repository, bus eventbus.EventBus, users user.Repository) *DeclineRequestHandler {
	return &DeclineRequestHandler{
		repo:  repo,
		bus:   bus,
		users: users,
	}
}

func (h *DeclineRequestHandler) Execute(ctx context.Context, cmd DeclineRequestCommand) error {
	err := h.repo.DeleteFollowRequest(ctx, cmd.FollowerID, cmd.FolloweeID)
	if err != nil {
		return err
	}

	actor, err := h.users.GetByID(ctx, cmd.FolloweeID)
	if err != nil {
		return err
	}
	body, err := json.Marshal(eventbus.Notification{
		Type:         eventbus.EventFollowDeclined,
		RecipientID:  cmd.FollowerID,
		ActorID:      cmd.FolloweeID,
		ActorName:    actor.Nickname,
		ActorAvatar:  actor.AvatarPath,
		ResourceType: eventbus.ResourceUser,
		ResourceID:   cmd.FollowerID,
	})
	if err != nil {
		return err
	}
	return h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, body)
}
