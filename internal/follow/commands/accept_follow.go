package commands

import (
	"context"
	"encoding/json"
	"time"

	"social-network/internal/follow"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

type AcceptRequestCommand struct {
	FollowerID string
	FolloweeID string // user accepting the request.
}

type AcceptRequestHandler struct {
	repo  follow.Repository
	bus   eventbus.EventBus
	users user.Repository
}

func NewAcceptRequestHandler(repo follow.Repository, bus eventbus.EventBus, users user.Repository) *AcceptRequestHandler {
	return &AcceptRequestHandler{
		repo:  repo,
		bus:   bus,
		users: users,
	}
}

func (h *AcceptRequestHandler) Execute(ctx context.Context, cmd AcceptRequestCommand) error {
	err := h.repo.DeleteFollowRequest(ctx, cmd.FollowerID, cmd.FolloweeID)
	if err != nil {
		return err
	}

	f := &follow.Follow{
		FollowerID: cmd.FollowerID,
		FolloweeID: cmd.FolloweeID,
		CreatedAt:  time.Now(),
	}
	err = h.repo.CreateFollow(ctx, f)
	if err != nil {
		return err
	}

	actor, err := h.users.GetByID(ctx, cmd.FolloweeID)
	if err != nil {
		return err
	}
	body, err := json.Marshal(eventbus.Envelope{
		Type:        "follow.accepted",
		RecipientID: cmd.FollowerID,
		ActorID:     cmd.FolloweeID,
		ActorName:   actor.Nickname,
		ActorAvatar: actor.AvatarPath,
	})
	if err != nil {
		return err
	}
	return h.bus.Publish("notifications.exchange", "follow.accepted", body)
}
