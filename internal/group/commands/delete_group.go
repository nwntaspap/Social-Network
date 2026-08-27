package commands

import (
	"context"
	"encoding/json"
	"errors"

	"social-network/internal/group"
	"social-network/internal/platform/eventbus"
)

type DeleteGroupCommand struct {
	GroupID string
	UserID  string
}

type DeleteGroupHandler struct {
	repo group.Repository
	bus  eventbus.EventBus
}

func NewDeleteGroupHandler(repo group.Repository, bus eventbus.EventBus) *DeleteGroupHandler {
	return &DeleteGroupHandler{
		repo: repo,
		bus:  bus,
	}
}

func (h *DeleteGroupHandler) Execute(ctx context.Context, cmd DeleteGroupCommand) error {
	if cmd.UserID == "" {
		return ErrUserIDRequired
	}
	if cmd.GroupID == "" {
		return ErrGroupIDRequired
	}

	role, err := h.repo.GetMemberRole(ctx, cmd.GroupID, cmd.UserID)
	if err != nil {
		if errors.Is(err, group.ErrNotMember) {
			return group.ErrNotCreator
		}
		return err
	}

	if role != group.RoleCreator {
		return group.ErrNotCreator
	}

	err = h.repo.DeleteGroup(ctx, cmd.GroupID)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(eventbus.Notification{
		Type:       eventbus.EventGroup,
		ResourceID: cmd.GroupID,
	})
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, body)
	return nil
}
