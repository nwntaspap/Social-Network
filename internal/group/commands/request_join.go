package commands

import (
	"context"
	"encoding/json"
	"errors"

	"social-network/internal/group"
	"social-network/internal/pkg/uuid"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

type RequestJoinCommand struct {
	GroupID     string
	RequesterID string
}

type RequestJoinHandler struct {
	repo  group.Repository
	bus   eventbus.EventBus
	users user.Repository
}

func NewRequestJoinHandler(repo group.Repository, bus eventbus.EventBus, users user.Repository) *RequestJoinHandler {
	return &RequestJoinHandler{repo: repo, bus: bus, users: users}
}

func (h *RequestJoinHandler) Execute(ctx context.Context, cmd RequestJoinCommand) (*group.JoinRequest, error) {
	if cmd.GroupID == "" || cmd.RequesterID == "" {
		return nil, errors.New("group_id and requester_id are required")
	}

	g, err := h.repo.GetGroupByID(ctx, cmd.GroupID)
	if err != nil {
		return nil, err
	}

	if cmd.RequesterID == g.CreatorID {
		return nil, group.ErrSelfJoin
	}

	isMember, err := h.repo.IsMember(ctx, cmd.GroupID, cmd.RequesterID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, group.ErrAlreadyMember
	}

	hasPending, err := h.repo.HasPendingRequest(ctx, cmd.GroupID, cmd.RequesterID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, group.ErrAlreadyRequested
	}

	jr := &group.JoinRequest{
		ID:          uuid.NewProvider().NewUUID(),
		GroupID:     cmd.GroupID,
		RequesterID: cmd.RequesterID,
	}

	if err = h.repo.CreateJoinRequest(ctx, jr); err != nil {
		return nil, err
	}

	actor, err := h.users.GetByID(ctx, cmd.RequesterID)
	if err != nil {
		return nil, err
	}

	body, _ := json.Marshal(eventbus.Envelope{
		Type:         "group.join_requested",
		RecipientID:  g.CreatorID,
		ActorID:      cmd.RequesterID,
		ActorName:    actor.Nickname,
		ActorAvatar:  actor.AvatarPath,
		ResourceType: "group",
	})
	_ = h.bus.Publish("notifications.exchange", "group.join_requested", body)

	return jr, nil
}
