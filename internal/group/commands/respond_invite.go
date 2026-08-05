package commands

import (
	"context"
	"encoding/json"
	"errors"

	"social-network/internal/group"
	"social-network/internal/pkg/uuid"
	"social-network/internal/platform/eventbus"
)

type RespondInviteCommand struct {
	GroupID   string
	InviteeID string
	Accept    bool
}

// RespondInviteResult reports what happened on an accepted invitation:
// "member" when the invitee joined immediately, "pending" when the invitee's
// acceptance became a join request awaiting the group creator's approval.
type RespondInviteResult string

const (
	RespondInviteMember  RespondInviteResult = "member"
	RespondInvitePending RespondInviteResult = "pending"
)

type RespondInviteHandler struct {
	repo group.Repository
	bus  eventbus.EventBus
}

func NewRespondInviteHandler(repo group.Repository, bus eventbus.EventBus) *RespondInviteHandler {
	return &RespondInviteHandler{
		repo: repo,
		bus:  bus,
	}
}

func (h *RespondInviteHandler) Execute(ctx context.Context, cmd RespondInviteCommand) (RespondInviteResult, error) {
	if cmd.GroupID == "" || cmd.InviteeID == "" {
		return "", errors.New("group_id and invitee_id are required")
	}

	inv, err := h.repo.GetInvitation(ctx, cmd.GroupID, cmd.InviteeID)
	if err != nil {
		return "", err
	}

	if deleteErr := h.repo.DeleteInvitation(ctx, cmd.GroupID, cmd.InviteeID); deleteErr != nil {
		return "", deleteErr
	}

	if !cmd.Accept {
		return "", nil
	}

	role, err := h.repo.GetMemberRole(ctx, cmd.GroupID, inv.InviterID)
	if err != nil {
		return "", err
	}

	// Creator invites join the group immediately. Anyone else's invite only
	// becomes a join request that the group creator approves or rejects.
	if role == group.RoleCreator {
		if addErr := h.repo.AddMember(ctx, cmd.GroupID, cmd.InviteeID, group.RoleMember); addErr != nil {
			return "", addErr
		}
		return RespondInviteMember, nil
	}

	isMember, err := h.repo.IsMember(ctx, cmd.GroupID, cmd.InviteeID)
	if err != nil {
		return "", err
	}
	if isMember {
		return RespondInviteMember, nil
	}

	hasPending, err := h.repo.HasPendingRequest(ctx, cmd.GroupID, cmd.InviteeID)
	if err != nil {
		return "", err
	}
	if !hasPending {
		jr := &group.JoinRequest{
			ID:          uuid.NewProvider().NewUUID(),
			GroupID:     cmd.GroupID,
			RequesterID: cmd.InviteeID,
		}
		if err := h.repo.CreateJoinRequest(ctx, jr); err != nil {
			return "", err
		}
	}

	// i didnt add a notification for the accept or decline
	body, _ := json.Marshal(eventbus.Notification{
		Type:         eventbus.EventGroupInvitation,
		RecipientID:  cmd.InviteeID,
		ActorID:      inv.InviterID,
		ResourceType: eventbus.ResourceGroup,
		ResourceID:   cmd.GroupID,
	})
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, body)

	return RespondInvitePending, nil
}
