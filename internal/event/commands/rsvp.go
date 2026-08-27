package commands

import (
	"context"

	"social-network/internal/event"
)

// GroupMembershipChecker reports whether a user belongs to a group.
type GroupMembershipChecker interface {
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
}

type RSVPCommand struct {
	EventID  string
	UserID   string
	OptionID string
}

type RSVPHandler struct {
	repo   event.Repository
	member GroupMembershipChecker
}

func NewRSVPHandler(repo event.Repository, member GroupMembershipChecker) *RSVPHandler {
	return &RSVPHandler{
		repo:   repo,
		member: member,
	}
}

func (h *RSVPHandler) Execute(ctx context.Context, cmd RSVPCommand) error {
	if cmd.EventID == "" {
		return ErrEventIDRequired
	}
	if cmd.UserID == "" {
		return ErrUserIDRequired
	}
	if cmd.OptionID == "" {
		return ErrInvalidOption
	}

	e, err := h.repo.GetEvent(ctx, cmd.EventID)
	if err != nil {
		return err
	}

	isMember, err := h.member.IsMember(ctx, e.GroupID, cmd.UserID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotGroupMember
	}

	opt, err := h.repo.GetOption(ctx, cmd.OptionID)
	if err != nil {
		return err
	}
	if opt.EventID != cmd.EventID {
		return ErrInvalidOption
	}

	rsvp := &event.RSVP{
		EventID:  cmd.EventID,
		UserID:   cmd.UserID,
		OptionID: cmd.OptionID,
	}

	return h.repo.UpsertRSVP(ctx, rsvp)
}
