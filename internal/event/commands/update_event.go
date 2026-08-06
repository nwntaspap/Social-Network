package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	"social-network/internal/event"
)

var (
	ErrNotGroupCreator    = errors.New("only the group creator can modify this event")
	ErrEventGroupMismatch = errors.New("event does not belong to this group")
)

type GroupRoleChecker interface {
	GetMemberRole(ctx context.Context, groupID, userID string) (string, error)
}

type UpdateEventCommand struct {
	UserID        string
	GroupID       string
	EventID       string
	Title         string
	Description   string
	ScheduledTime time.Time
}

type UpdateEventHandler struct {
	repo  event.Repository
	group GroupRoleChecker
}

func NewUpdateEventHandler(repo event.Repository, group GroupRoleChecker) *UpdateEventHandler {
	return &UpdateEventHandler{repo: repo, group: group}
}

func (h *UpdateEventHandler) Execute(ctx context.Context, cmd UpdateEventCommand) (*event.Event, []event.Option, error) {
	if cmd.UserID == "" {
		return nil, nil, ErrUserIDRequired
	}
	if cmd.GroupID == "" {
		return nil, nil, ErrGroupIDRequired
	}
	if cmd.EventID == "" {
		return nil, nil, ErrEventIDRequired
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

	e, err := h.repo.GetEvent(ctx, cmd.EventID)
	if err != nil {
		return nil, nil, err
	}
	if e.GroupID != cmd.GroupID {
		return nil, nil, ErrEventGroupMismatch
	}

	role, err := h.group.GetMemberRole(ctx, cmd.GroupID, cmd.UserID)
	if err != nil {
		return nil, nil, err
	}
	if role != "creator" {
		return nil, nil, ErrNotGroupCreator
	}

	e.Title = title
	e.Description = desc
	e.ScheduledTime = cmd.ScheduledTime

	err = h.repo.UpdateEvent(ctx, e)
	if err != nil {
		return nil, nil, err
	}

	opts, err := h.repo.GetOptionsByEvent(ctx, e.ID)
	if err != nil {
		return nil, nil, err
	}
	return e, opts, nil
}
