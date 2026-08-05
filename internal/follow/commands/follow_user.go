package commands

import (
	"context"
	"encoding/json"
	"errors"

	"social-network/internal/follow"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
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
	bus     eventbus.EventBus
	users   user.Repository
}

func NewFollowUserHandler(repo follow.Repository, privacy follow.UserPrivacyChecker, bus eventbus.EventBus, users user.Repository) *FollowUserHandler {
	return &FollowUserHandler{
		repo:    repo,
		privacy: privacy,
		bus:     bus,
		users:   users,
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

	actor, err := h.users.GetByID(ctx, cmd.FollowerID)
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
			return "", err
		}

		body, _ := json.Marshal(eventbus.Notification{
			Type:         eventbus.EventFollowRequested,
			RecipientID:  cmd.TargetID,
			ActorID:      cmd.FollowerID,
			ActorName:    actor.Nickname,
			ActorAvatar:  actor.AvatarPath,
			ResourceType: eventbus.ResourceUser,
			ResourceID:   cmd.TargetID,
		})
		err = h.bus.Publish("notifications.exchange", eventbus.RoutingCreated, body)
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
	body, _ := json.Marshal(eventbus.Notification{
		Type:         eventbus.EventFollow,
		RecipientID:  cmd.TargetID,
		ActorID:      cmd.FollowerID,
		ActorName:    actor.Nickname,
		ActorAvatar:  actor.AvatarPath,
		ResourceType: eventbus.ResourceUser,
		ResourceID:   cmd.TargetID,
	})
	err = h.bus.Publish("notifications.exchange", eventbus.RoutingCreated, body)
	if err != nil {
		return "", err
	}
	return FollowedDirect, nil
}
