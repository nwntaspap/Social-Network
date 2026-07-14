package commands

import (
	"context"

	"social-network/internal/follow"
)

type DeclineRequestCommand struct {
	FollowerID string
	FolloweeID string
}

type DeclineRequestHandler struct {
	repo follow.Repository
	bus  EventBus
}

func NewDeclineRequestHandler(repo follow.Repository, bus EventBus) *DeclineRequestHandler {
	return &DeclineRequestHandler{
		repo: repo,
		bus:  bus,
	}
}

func (h *DeclineRequestHandler) Execute(ctx context.Context, cmd DeclineRequestCommand) error {
	err := h.repo.DeleteFollowRequest(ctx, cmd.FollowerID, cmd.FolloweeID)
	if err != nil {
		return err
	}

	return h.bus.Publish(ctx, "follow.declined", &follow.Request{
		FollowerID: cmd.FollowerID,
		FolloweeID: cmd.FolloweeID,
	})
}
