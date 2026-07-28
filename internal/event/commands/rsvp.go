package commands

import (
	"context"

	"social-network/internal/event"
)

type RSVPCommand struct {
	EventID  string
	UserID   string
	OptionID string
}

type RSVPHandler struct {
	repo event.Repository
}

func NewRSVPHandler(repo event.Repository) *RSVPHandler {
	return &RSVPHandler{repo: repo}
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

	_, err := h.repo.GetEvent(ctx, cmd.EventID)
	if err != nil {
		return err
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
