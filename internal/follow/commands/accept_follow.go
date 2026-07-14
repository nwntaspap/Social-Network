package commands

import (
	"context"
	"time"

	"social-network/internal/follow"
)

type AcceptRequestCommand struct {
	FollowerID string
	FolloweeID string // user accepting the request.
}

type AcceptRequestHandler struct {
	repo follow.Repository
	bus  EventBus
}

func NewAcceptRequestHandler(repo follow.Repository, bus EventBus) *AcceptRequestHandler {
	return &AcceptRequestHandler{
		repo: repo,
		bus:  bus,
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

	return h.bus.Publish(ctx, "follow.accepted", &follow.Request{
		FollowerID: cmd.FollowerID,
		FolloweeID: cmd.FolloweeID,
	})
}
