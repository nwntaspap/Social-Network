package commands

import (
	"context"
	"encoding/json"
	"errors"

	"social-network/internal/follow"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

var ErrCannotUnfollowSelf = errors.New("cannot unfollow yourself")

type UnfollowUserCommand struct {
	FollowerID string
	TargetID   string
}

type UnfollowUserHandler struct {
	repo  follow.Repository
	bus   eventbus.EventBus
	users user.Repository
}

func NewUnfollowUserHandler(repo follow.Repository, bus eventbus.EventBus, users user.Repository) *UnfollowUserHandler {
	return &UnfollowUserHandler{
		repo:  repo,
		bus:   bus,
		users: users,
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

	actor, err := h.users.GetByID(ctx, cmd.FollowerID)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(eventbus.Envelope{
		Type:        "follow.deleted",
		RecipientID: cmd.TargetID,
		ActorID:     cmd.FollowerID,
		ActorName:   actor.Nickname,
		ActorAvatar: actor.AvatarPath,
	})
	return h.bus.Publish("notifications.exchange", "follow.deleted", body)
}
