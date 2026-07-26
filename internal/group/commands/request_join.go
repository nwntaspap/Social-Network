package commands

import (
	"context"
	"errors"

	"social-network/internal/group"
	"social-network/internal/pkg/uuid"
)

type RequestJoinCommand struct {
	GroupID     string
	RequesterID string
}

type RequestJoinHandler struct {
	repo group.Repository
	bus  group.EventBus
}

func NewRequestJoinHandler(repo group.Repository, bus group.EventBus) *RequestJoinHandler {
	return &RequestJoinHandler{repo: repo, bus: bus}
}

func (h *RequestJoinHandler) Execute(ctx context.Context, cmd RequestJoinCommand) (*group.JoinRequest, error) {
	if cmd.GroupID == "" || cmd.RequesterID == "" {
		return nil, errors.New("group_id and requester_id are required")
	}

	g, err := h.repo.GetGroupByID(ctx, cmd.GroupID)
	if err != nil {
		return nil, err
	}
	_ = g

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

	if err := h.repo.CreateJoinRequest(ctx, jr); err != nil {
		return nil, err
	}

	_ = h.bus.Publish(ctx, "group.join_requested", jr)

	return jr, nil
}
