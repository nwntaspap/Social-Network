package commands

import (
	"context"
	"errors"

	"social-network/internal/group"
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
}

func NewRespondJoinHandler(repo group.Repository) *RespondJoinHandler {
	return &RespondJoinHandler{repo: repo}
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
	if role != group.RoleCreator && role != group.RoleAdmin {
		return group.ErrNotAdmin
	}

	_, err = h.repo.GetJoinRequest(ctx, groupID, requesterID)
	if err != nil {
		return err
	}

	if err := h.repo.DeleteJoinRequest(ctx, groupID, requesterID); err != nil {
		return err
	}

	if cmd.Accept {
		if err := h.repo.AddMember(ctx, groupID, requesterID, group.RoleMember); err != nil {
			return err
		}
	}

	return nil
}
