package commands

import (
	"context"
	"encoding/json"
	"errors"

	"social-network/internal/follow"
	"social-network/internal/platform/eventbus"
)

var ErrCannotUnfollowSelf = errors.New("cannot unfollow yourself")

type UnfollowUserCommand struct {
	FollowerID string
	TargetID   string
}

type UnfollowUserHandler struct {
	repo follow.Repository
	bus  eventbus.EventBus
}

func NewUnfollowUserHandler(repo follow.Repository, bus eventbus.EventBus) *UnfollowUserHandler {
	return &UnfollowUserHandler{
		repo: repo,
		bus:  bus,
	}
}

func (h *UnfollowUserHandler) Execute(ctx context.Context, cmd UnfollowUserCommand) error {
	if cmd.FollowerID == cmd.TargetID {
		return ErrCannotUnfollowSelf
	}
	err := h.repo.DeleteFollow(ctx, cmd.FollowerID, cmd.TargetID)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(eventbus.Notification{
		Type:        eventbus.EventFollow,
		RecipientID: cmd.TargetID,
		ActorID:     cmd.FollowerID,
	})
	return h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, body)
}
