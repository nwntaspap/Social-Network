package commands

import (
	"context"
	"errors"

	"social-network/internal/follow"
)

var ErrSelfFollow = errors.New("cannot follow yourself")

type FollowUserCommand struct {
	FollowerID string
	TargetID   string
}

type FollowUserHandler struct {
	repo    follow.Repository
	privacy follow.UserPrivacyChecker
	bus     follow.EventBus
}

func NewFollowUserHandler(repo follow.Repository, privacy follow.UserPrivacyChecker, bus follow.EventBus) *FollowUserHandler {
	return &FollowUserHandler{
		repo:    repo,
		privacy: privacy,
		bus:     bus,
	}
}

func (h *FollowUserHandler) Execute(ctx context.Context, cmd FollowUserCommand) error {
	if cmd.FollowerID == cmd.TargetID {
		return ErrSelfFollow
	}

	isPrivate, err := h.privacy.IsPrivate(ctx, cmd.TargetID)
	if err != nil {
		return err
	}

	if isPrivate {
		req := &follow.Request{
			FollowerID: cmd.FollowerID,
			FolloweeID: cmd.TargetID,
		}
		err = h.repo.CreateFollowRequest(ctx, req)
		if err != nil {
			return err
		}

		return h.bus.Publish(ctx, "follow.requested", req)
	}
	f := &follow.Follow{
		FollowerID: cmd.FollowerID,
		FolloweeID: cmd.TargetID,
	}
	err = h.repo.CreateFollow(ctx, f)
	if err != nil {
		return err
	}
	return h.bus.Publish(ctx, "follow.accepted", f)
}
