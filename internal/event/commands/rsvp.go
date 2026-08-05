package commands

import (
	"context"
	"encoding/json"

	"social-network/internal/event"
	"social-network/internal/platform/eventbus"
)

type RSVPCommand struct {
	EventID  string
	UserID   string
	OptionID string
}

type RSVPHandler struct {
	repo event.Repository
	bus  eventbus.EventBus
}

func NewRSVPHandler(repo event.Repository, bus eventbus.EventBus) *RSVPHandler {
	return &RSVPHandler{
		repo: repo,
		bus:  bus,
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

	payload, _ := json.Marshal(eventbus.Notification{
		Type:        eventbus.EventEvent,
		RecipientID: cmd.UserID,
		EventID:     cmd.EventID,
	})
	_ = h.bus.Publish("notifications.exchange", eventbus.RoutingDeleted, payload)
	return h.repo.UpsertRSVP(ctx, rsvp)
}
