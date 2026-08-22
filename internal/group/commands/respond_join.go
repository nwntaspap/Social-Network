package commands

import (
	"context"
	"encoding/json"
	"errors"
	"log"

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

func NewRespondJoinHandler(repo group.Repository, bus eventbus.EventBus, user user.Repository) *RespondJoinHandler {
	return &RespondJoinHandler{
		repo: repo,
		bus:  bus,
		user: user,
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

	err = h.repo.DeleteJoinRequest(ctx, groupID, requesterID)
	if err != nil {
		return err
	}

	gr, err := h.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		log.Println("this is the error", err)
		return err
	}
	eventType := eventbus.EventGroupJoinDeclined
	if cmd.Accept {
		eventType = eventbus.EventGroupJoinAccepted
		err = h.repo.AddMember(ctx, groupID, requesterID, group.RoleMember)
		if err != nil {
			return err
		}
	}

	actor, err := h.user.GetByID(ctx, cmd.AdminID)
	if err != nil {
		log.Println("hello this is the eroor", err)
		return err
	}
	body, _ := json.Marshal(eventbus.Notification{
		Type:          eventType,
		RecipientID:   requesterID,
		ActorID:       cmd.AdminID,
		ActorName:     actor.Nickname,
		ActorAvatar:   actor.AvatarPath,
		ResourceType:  eventbus.ResourceGroup,
		ResourceID:    groupID,
		JoinRequestID: jr.ID,
		ContentText:   gr.Title,
	})
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingCreated, body)

	return nil
}
