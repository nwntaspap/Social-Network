package commands

import (
	"context"
	"encoding/json"
	"errors"

	"social-network/internal/group"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

type RespondJoinCommand struct {
	RequestID   string
	GroupID     string
	RequesterID string
	AdminID     string
	Accept      bool
}

type RespondJoinHandler struct {
	repo group.Repository
	user user.Repository
	bus  eventbus.EventBus
}

func NewRespondJoinHandler(repo group.Repository, bus eventbus.EventBus) *RespondJoinHandler {
	return &RespondJoinHandler{
		repo: repo,
		bus:  bus,
	}
}

func (h *RespondJoinHandler) Execute(ctx context.Context, cmd RespondJoinCommand) error {
	if cmd.AdminID == "" {
		return errors.New("admin_id is required")
	}

	groupID := cmd.GroupID
	requesterID := cmd.RequesterID

	if cmd.RequestID != "" {
		jr, err := h.repo.GetJoinRequestByID(ctx, cmd.RequestID)
		if err != nil {
			return err
		}
		groupID = jr.GroupID
		requesterID = jr.RequesterID
	}

	if groupID == "" || requesterID == "" {
		return errors.New("group_id and requester_id are required")
	}

	role, err := h.repo.GetMemberRole(ctx, groupID, cmd.AdminID)
	if err != nil {
		return err
	}
	if role != group.RoleCreator {
		return group.ErrNotCreator
	}

	jr, err := h.repo.GetJoinRequest(ctx, groupID, requesterID)
	if err != nil {
		return err
	}

	if err := h.repo.DeleteJoinRequest(ctx, groupID, requesterID); err != nil {
		return err
	}

	gr, _ := h.repo.GetGroupByID(ctx, cmd.GroupID)
	eventType := eventbus.EventGroupJoinDeclined
	if cmd.Accept {
		eventType = eventbus.EventGroupJoinAccepted
		if err := h.repo.AddMember(ctx, groupID, requesterID, group.RoleMember); err != nil {
			return err
		}
	}

	actor, _ := h.user.GetByID(ctx, cmd.AdminID)
	body, _ := json.Marshal(eventbus.Notification{
		Type:          eventType,
		RecipientID:   cmd.RequesterID,
		ActorID:       cmd.AdminID,
		ActorName:     actor.Nickname,
		ActorAvatar:   actor.AvatarPath,
		ResourceType:  eventbus.ResourceGroup,
		ResourceID:    cmd.GroupID,
		JoinRequestID: jr.ID,
		ContentText:   gr.Title,
	})
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingCreated, body)

	return nil
}
