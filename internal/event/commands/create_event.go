package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	"social-network/internal/event"
	"social-network/internal/pkg/uuid"
)

var (
	ErrUserIDRequired   = errors.New("user ID is required")
	ErrGroupIDRequired  = errors.New("group ID is required")
	ErrEventIDRequired  = errors.New("event ID is required")
	ErrTitleRequired    = errors.New("title is required")
	ErrTitleTooLong     = errors.New("title must be 100 characters or fewer")
	ErrDescRequired     = errors.New("description is required")
	ErrDescTooLong      = errors.New("description must be 500 characters or fewer")
	ErrTimeRequired     = errors.New("event time is required")
	ErrTimeInPast       = errors.New("event time must be in the future")
	ErrMinOptions       = errors.New("at least 2 options are required")
	ErrOptionLabelEmpty = errors.New("option label must not be empty")
	ErrNotGroupMember   = errors.New("user is not a member of this group")
	ErrInvalidOption    = errors.New("option does not belong to this event")
)

type GroupMemberChecker interface {
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
}

type CreateEventCommand struct {
	UserID        string
	GroupID       string
	Title         string
	Description   string
	ScheduledTime time.Time
	Options       []string
}

type CreateEventHandler struct {
	repo   event.Repository
	member GroupMemberChecker
	bus    event.Bus
}

func NewCreateEventHandler(repo event.Repository, member GroupMemberChecker, bus event.Bus) *CreateEventHandler {
	return &CreateEventHandler{repo: repo, member: member, bus: bus}
}

func (h *CreateEventHandler) Execute(ctx context.Context, cmd CreateEventCommand) (*event.Event, []event.Option, error) {
	if cmd.UserID == "" {
		return nil, nil, ErrUserIDRequired
	}
	if cmd.GroupID == "" {
		return nil, nil, ErrGroupIDRequired
	}

	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return nil, nil, ErrTitleRequired
	}
	if len(title) > 100 {
		return nil, nil, ErrTitleTooLong
	}

	desc := strings.TrimSpace(cmd.Description)
	if desc == "" {
		return nil, nil, ErrDescRequired
	}
	if len(desc) > 500 {
		return nil, nil, ErrDescTooLong
	}

	if cmd.ScheduledTime.IsZero() {
		return nil, nil, ErrTimeRequired
	}
	if cmd.ScheduledTime.Before(time.Now()) {
		return nil, nil, ErrTimeInPast
	}

	if len(cmd.Options) < 2 {
		return nil, nil, ErrMinOptions
	}
	for _, opt := range cmd.Options {
		if strings.TrimSpace(opt) == "" {
			return nil, nil, ErrOptionLabelEmpty
		}
	}

	isMember, err := h.member.IsMember(ctx, cmd.GroupID, cmd.UserID)
	if err != nil {
		return nil, nil, err
	}
	if !isMember {
		return nil, nil, ErrNotGroupMember
	}

	e := &event.Event{
		ID:            uuid.NewProvider().NewUUID(),
		GroupID:       cmd.GroupID,
		CreatorID:     cmd.UserID,
		Title:         title,
		Description:   desc,
		ScheduledTime: cmd.ScheduledTime,
	}

	if err := h.repo.CreateEvent(ctx, e); err != nil {
		return nil, nil, err
	}

	opts := make([]event.Option, len(cmd.Options))
	for i, label := range cmd.Options {
		opts[i] = event.Option{
			ID:      uuid.NewProvider().NewUUID(),
			EventID: e.ID,
			Label:   strings.TrimSpace(label),
		}
	}
	if err := h.repo.CreateOptions(ctx, opts); err != nil {
		return nil, nil, err
	}

	_ = h.bus.Publish(ctx, "event.created", e)

	return e, opts, nil
}
