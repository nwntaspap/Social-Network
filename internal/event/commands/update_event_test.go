package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/event"
)

type fakeRoleChecker struct {
	role string
	err  error
}

func (f *fakeRoleChecker) GetMemberRole(_ context.Context, _, _ string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.role, nil
}

func TestUpdateEventHandler_Validation(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(time.Hour)

	repo := newFakeRepo()
	repo.events["evt-1"] = &event.Event{ID: "evt-1", GroupID: "g1", CreatorID: "u1", Title: "Old", Description: "Old desc", ScheduledTime: future}
	repo.options["evt-1"] = []event.Option{{ID: "opt-1", EventID: "evt-1", Label: "going"}}

	handler := NewUpdateEventHandler(repo, &fakeRoleChecker{role: "creator"})

	t.Run("empty event ID", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, UpdateEventCommand{UserID: "u1", GroupID: "g1", Title: "T", Description: "D", ScheduledTime: future})
		if !errors.Is(err, ErrEventIDRequired) {
			t.Errorf("expected ErrEventIDRequired, got %v", err)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, UpdateEventCommand{UserID: "u1", GroupID: "g1", EventID: "evt-1", Description: "D", ScheduledTime: future})
		if !errors.Is(err, ErrTitleRequired) {
			t.Errorf("expected ErrTitleRequired, got %v", err)
		}
	})

	t.Run("time in past", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, UpdateEventCommand{UserID: "u1", GroupID: "g1", EventID: "evt-1", Title: "T", Description: "D", ScheduledTime: time.Now().Add(-time.Hour)})
		if !errors.Is(err, ErrTimeInPast) {
			t.Errorf("expected ErrTimeInPast, got %v", err)
		}
	})

	t.Run("event not found", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, UpdateEventCommand{UserID: "u1", GroupID: "g1", EventID: "missing", Title: "T", Description: "D", ScheduledTime: future})
		if !errors.Is(err, event.ErrEventNotFound) {
			t.Errorf("expected ErrEventNotFound, got %v", err)
		}
	})

	t.Run("event belongs to different group", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, UpdateEventCommand{UserID: "u1", GroupID: "g2", EventID: "evt-1", Title: "T", Description: "D", ScheduledTime: future})
		if !errors.Is(err, ErrEventGroupMismatch) {
			t.Errorf("expected ErrEventGroupMismatch, got %v", err)
		}
	})

	t.Run("non-creator rejected", func(t *testing.T) {
		h := NewUpdateEventHandler(repo, &fakeRoleChecker{role: "member"})
		_, _, err := h.Execute(ctx, UpdateEventCommand{UserID: "u9", GroupID: "g1", EventID: "evt-1", Title: "T", Description: "D", ScheduledTime: future})
		if !errors.Is(err, ErrNotGroupCreator) {
			t.Errorf("expected ErrNotGroupCreator, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		e, opts, err := handler.Execute(ctx, UpdateEventCommand{UserID: "u1", GroupID: "g1", EventID: "evt-1", Title: "New title", Description: "New desc", ScheduledTime: future})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Title != "New title" || e.Description != "New desc" {
			t.Errorf("event = %+v, want updated title/description", e)
		}
		if len(opts) != 1 {
			t.Errorf("expected 1 option, got %d", len(opts))
		}
	})
}
