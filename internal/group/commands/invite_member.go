package commands

import (
	"context"
	"errors"

	"social-network/internal/group"
	"social-network/internal/pkg/uuid"
)

var ErrNotConnected = errors.New("users are not connected: inviter must follow invitee")

type InviteMemberCommand struct {
	GroupID   string
	InviterID string
	InviteeID string
}

type InviteMemberHandler struct {
	repo   group.Repository
	follow group.FollowChecker
	bus    group.EventBus
}

func NewInviteMemberHandler(repo group.Repository, follow group.FollowChecker, bus group.EventBus) *InviteMemberHandler {
	return &InviteMemberHandler{repo: repo, follow: follow, bus: bus}
}

func (h *InviteMemberHandler) Execute(ctx context.Context, cmd InviteMemberCommand) (*group.Invitation, error) {
	if cmd.GroupID == "" || cmd.InviterID == "" || cmd.InviteeID == "" {
		return nil, errors.New("group_id, inviter_id, and invitee_id are required")
	}

	if cmd.InviterID == cmd.InviteeID {
		return nil, group.ErrSelfInvite
	}

	g, err := h.repo.GetGroupByID(ctx, cmd.GroupID)
	if err != nil {
		return nil, err
	}
	_ = g

	_, err = h.repo.GetMemberRole(ctx, cmd.GroupID, cmd.InviterID)
	if err != nil {
		return nil, group.ErrNotMember
	}

	isMember, err := h.repo.IsMember(ctx, cmd.GroupID, cmd.InviteeID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, group.ErrAlreadyMember
	}

	alreadyInvited, err := h.repo.IsInvited(ctx, cmd.GroupID, cmd.InviteeID)
	if err != nil {
		return nil, err
	}
	if alreadyInvited {
		return nil, group.ErrAlreadyInvited
	}

	connected, err := h.follow.AreConnected(ctx, cmd.InviterID, cmd.InviteeID)
	if err != nil {
		return nil, err
	}
	if !connected {
		return nil, ErrNotConnected
	}

	inv := &group.Invitation{
		ID:        uuid.NewProvider().NewUUID(),
		GroupID:   cmd.GroupID,
		InviterID: cmd.InviterID,
		InviteeID: cmd.InviteeID,
	}

	if err := h.repo.CreateInvitation(ctx, inv); err != nil {
		return nil, err
	}

	_ = h.bus.Publish(ctx, "group.invited", inv)

	return inv, nil
}
