package commands

import (
	"context"
	"errors"

	"social-network/internal/follow"
)

var ErrCannotUnfollowSelf = errors.New("cannot unfollow yourself")

type UnfollowUserCommand struct {
	FollowerID string
	TargetID   string
}

type UnfollowUserHandler struct {
	repo follow.Repository
}

func NewUnfollowUserHandler(repo follow.Repository) *UnfollowUserHandler {
	return &UnfollowUserHandler{
		repo: repo,
	}
}

func (h *UnfollowUserHandler) Execute(ctx context.Context, cmd UnfollowUserCommand) error {
	if cmd.FollowerID == cmd.TargetID {
		return ErrCannotUnfollowSelf
	}
	return h.repo.DeleteFollow(ctx, cmd.FollowerID, cmd.TargetID)
}
