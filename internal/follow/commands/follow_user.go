package commands

import (
	"context"
	"errors"

	"social-network/internal/follow"
)

var ErrSelfFollow = errors.New("cannot follow yourself")

// FollowUserResult reports what happened on a follow action:
// "following" when the follower joined immediately, "pending" when the
// followee is private and the action became a follow request awaiting approval.
type FollowUserResult string

const (
	FollowedDirect FollowUserResult = "following"
	FollowPending  FollowUserResult = "pending"
)

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

func (h *FollowUserHandler) Execute(ctx context.Context, cmd FollowUserCommand) (FollowUserResult, error) {
	if cmd.FollowerID == cmd.TargetID {
		return "", ErrSelfFollow
	}

	isPrivate, err := h.privacy.IsPrivate(ctx, cmd.TargetID)
	if err != nil {
		return "", err
	}

	if isPrivate {
		req := &follow.Request{
			FollowerID: cmd.FollowerID,
			FolloweeID: cmd.TargetID,
		}
		err = h.repo.CreateFollowRequest(ctx, req)
		if err != nil {
			return "", err
		}

		err = h.bus.Publish(ctx, "follow.requested", req)
		if err != nil {
			return "", err
		}
		return FollowPending, nil
	}
	f := &follow.Follow{
		FollowerID: cmd.FollowerID,
		FolloweeID: cmd.TargetID,
	}
	err = h.repo.CreateFollow(ctx, f)
	if err != nil {
		return "", err
	}
	err = h.bus.Publish(ctx, "follow.accepted", f)
	if err != nil {
		return "", err
	}
	return FollowedDirect, nil
}
