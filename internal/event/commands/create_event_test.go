package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/event"
)

type fakeMemberChecker struct {
	isMember bool
	err      error
}

func (f *fakeMemberChecker) IsMember(_ context.Context, _, _ string) (bool, error) {
	return f.isMember, f.err
}

type fakeEventBus struct {
	published bool
}

func (f *fakeEventBus) Publish(_ context.Context, _ string, _ any) error {
	f.published = true
	return nil
}

type fakeRepo struct {
	events  map[string]*event.Event
	options map[string][]event.Option
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		events:  make(map[string]*event.Event),
		options: make(map[string][]event.Option),
	}
}

func (r *fakeRepo) CreateEvent(_ context.Context, e *event.Event) error {
	r.events[e.ID] = e
	return nil
}

func (r *fakeRepo) GetEvent(_ context.Context, id string) (*event.Event, error) {
	e, ok := r.events[id]
	if !ok {
		return nil, event.ErrEventNotFound
	}
	return e, nil
}

func (r *fakeRepo) DeleteEvent(_ context.Context, id string) error {
	delete(r.events, id)
	return nil
}

func (r *fakeRepo) ListGroupEvents(_ context.Context, _, _ string, _ int) ([]event.Event, string, error) {
	return nil, "", nil
}

func (r *fakeRepo) CreateOptions(_ context.Context, opts []event.Option) error {
	for _, o := range opts {
		r.options[o.EventID] = append(r.options[o.EventID], o)
	}
	return nil
}

func (r *fakeRepo) GetOptionsByEvent(_ context.Context, eventID string) ([]event.Option, error) {
	return r.options[eventID], nil
}

func (r *fakeRepo) GetOption(_ context.Context, id string) (*event.Option, error) {
	for _, opts := range r.options {
		for _, o := range opts {
			if o.ID == id {
				return &o, nil
			}
		}
	}
	return nil, event.ErrOptionNotFound
}

func (r *fakeRepo) UpsertRSVP(_ context.Context, _ *event.RSVP) error {
	return nil
}

func (r *fakeRepo) GetRSVPsByEvent(_ context.Context, _ string) ([]event.RSVP, error) {
	return nil, nil
}

func (r *fakeRepo) GetUserRSVP(_ context.Context, _, _ string) (*event.RSVP, error) {
	return nil, event.ErrRSVPNotFound
}

func TestCreateEventHandler_Validation(t *testing.T) {
	ctx := context.Background()
	bus := &fakeEventBus{}
	member := &fakeMemberChecker{isMember: true}
	repo := newFakeRepo()
	handler := NewCreateEventHandler(repo, member, bus)

	t.Run("empty user ID", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, CreateEventCommand{GroupID: "g1", Title: "T", Description: "D", ScheduledTime: time.Now().Add(time.Hour), Options: []string{"a", "b"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("empty group ID", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, CreateEventCommand{UserID: "u1", Title: "T", Description: "D", ScheduledTime: time.Now().Add(time.Hour), Options: []string{"a", "b"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("empty title", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, CreateEventCommand{UserID: "u1", GroupID: "g1", Description: "D", ScheduledTime: time.Now().Add(time.Hour), Options: []string{"a", "b"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("title too long", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, CreateEventCommand{UserID: "u1", GroupID: "g1", Title: string(make([]byte, 101)), Description: "D", ScheduledTime: time.Now().Add(time.Hour), Options: []string{"a", "b"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("empty description", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, CreateEventCommand{UserID: "u1", GroupID: "g1", Title: "T", ScheduledTime: time.Now().Add(time.Hour), Options: []string{"a", "b"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("time in past", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, CreateEventCommand{UserID: "u1", GroupID: "g1", Title: "T", Description: "D", ScheduledTime: time.Now().Add(-time.Hour), Options: []string{"a", "b"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("less than 2 options", func(t *testing.T) {
		_, _, err := handler.Execute(ctx, CreateEventCommand{UserID: "u1", GroupID: "g1", Title: "T", Description: "D", ScheduledTime: time.Now().Add(time.Hour), Options: []string{"a"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("not group member", func(t *testing.T) {
		m := &fakeMemberChecker{isMember: false}
		h := NewCreateEventHandler(repo, m, bus)
		_, _, err := h.Execute(ctx, CreateEventCommand{UserID: "u1", GroupID: "g1", Title: "T", Description: "D", ScheduledTime: time.Now().Add(time.Hour), Options: []string{"a", "b"}})
		if !errors.Is(err, ErrNotGroupMember) {
			t.Errorf("expected ErrNotGroupMember, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		e, opts, err := handler.Execute(ctx, CreateEventCommand{UserID: "u1", GroupID: "g1", Title: "T", Description: "D", ScheduledTime: time.Now().Add(time.Hour), Options: []string{"going", "not going"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e == nil {
			t.Fatal("expected event")
		}
		if len(opts) != 2 {
			t.Errorf("expected 2 options, got %d", len(opts))
		}
		if !bus.published {
			t.Error("expected event to be published")
		}
	})
}

func TestRSVPHandler_Validation(t *testing.T) {
	ctx := context.Background()

	t.Run("empty event ID", func(t *testing.T) {
		repo := newFakeRepo()
		handler := NewRSVPHandler(repo)
		err := handler.Execute(ctx, RSVPCommand{UserID: "u1", OptionID: "o1"})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("event not found", func(t *testing.T) {
		repo := newFakeRepo()
		handler := NewRSVPHandler(repo)
		err := handler.Execute(ctx, RSVPCommand{EventID: "nonexistent", UserID: "u1", OptionID: "o1"})
		if !errors.Is(err, event.ErrEventNotFound) {
			t.Errorf("expected ErrEventNotFound, got %v", err)
		}
	})

	t.Run("option not found", func(t *testing.T) {
		repo := newFakeRepo()
		repo.events["evt-1"] = &event.Event{ID: "evt-1"}
		handler := NewRSVPHandler(repo)
		err := handler.Execute(ctx, RSVPCommand{EventID: "evt-1", UserID: "u1", OptionID: "nonexistent"})
		if !errors.Is(err, event.ErrOptionNotFound) {
			t.Errorf("expected ErrOptionNotFound, got %v", err)
		}
	})

	t.Run("option belongs to different event", func(t *testing.T) {
		repo := newFakeRepo()
		repo.events["evt-1"] = &event.Event{ID: "evt-1"}
		repo.events["evt-2"] = &event.Event{ID: "evt-2"}
		repo.options["evt-2"] = []event.Option{{ID: "opt-2", EventID: "evt-2", Label: "going"}}
		handler := NewRSVPHandler(repo)
		err := handler.Execute(ctx, RSVPCommand{EventID: "evt-1", UserID: "u1", OptionID: "opt-2"})
		if !errors.Is(err, ErrInvalidOption) {
			t.Errorf("expected ErrInvalidOption, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		repo := newFakeRepo()
		repo.events["evt-1"] = &event.Event{ID: "evt-1"}
		repo.options["evt-1"] = []event.Option{{ID: "opt-1", EventID: "evt-1", Label: "going"}}
		handler := NewRSVPHandler(repo)
		err := handler.Execute(ctx, RSVPCommand{EventID: "evt-1", UserID: "u1", OptionID: "opt-1"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
