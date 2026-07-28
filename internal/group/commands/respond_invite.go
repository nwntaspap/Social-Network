package commands

import (
	"context"
	"errors"

	"social-network/internal/group"
)

type RespondInviteCommand struct {
	GroupID   string
	InviteeID string
	Accept    bool
}

type RespondInviteHandler struct {
	repo group.Repository
}

func NewRespondInviteHandler(repo group.Repository) *RespondInviteHandler {
	return &RespondInviteHandler{repo: repo}
}

func (h *RespondInviteHandler) Execute(ctx context.Context, cmd RespondInviteCommand) error {
	if cmd.GroupID == "" || cmd.InviteeID == "" {
		return errors.New("group_id and invitee_id are required")
	}

	_, err := h.repo.GetInvitation(ctx, cmd.GroupID, cmd.InviteeID)
	if err != nil {
		return err
	}

	if err := h.repo.DeleteInvitation(ctx, cmd.GroupID, cmd.InviteeID); err != nil {
		return err
	}

	if cmd.Accept {
		if err := h.repo.AddMember(ctx, cmd.GroupID, cmd.InviteeID, group.RoleMember); err != nil {
			return err
		}
	}

	return nil
}
