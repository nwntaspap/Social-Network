package queries

import (
	"context"
	"testing"
	"time"

	"social-network/internal/event"
)

type fakeEventRepo struct {
	events  map[string]*event.Event
	options map[string][]event.Option
	rsvps   map[string][]event.RSVP
}

func newFakeEventRepo() *fakeEventRepo {
	return &fakeEventRepo{
		events:  make(map[string]*event.Event),
		options: make(map[string][]event.Option),
		rsvps:   make(map[string][]event.RSVP),
	}
}

func (r *fakeEventRepo) CreateEvent(_ context.Context, e *event.Event) error {
	r.events[e.ID] = e
	return nil
}

func (r *fakeEventRepo) GetEvent(_ context.Context, id string) (*event.Event, error) {
	e, ok := r.events[id]
	if !ok {
		return nil, event.ErrEventNotFound
	}
	return e, nil
}

func (r *fakeEventRepo) UpdateEvent(_ context.Context, e *event.Event) error {
	if _, ok := r.events[e.ID]; !ok {
		return event.ErrEventNotFound
	}
	r.events[e.ID] = e
	return nil
}

func (r *fakeEventRepo) DeleteEvent(_ context.Context, id string) error {
	delete(r.events, id)
	return nil
}

func (r *fakeEventRepo) ListGroupEvents(_ context.Context, groupID, _ string, _ int) ([]event.Event, string, error) {
	var result []event.Event
	for _, e := range r.events {
		if e.GroupID == groupID {
			result = append(result, *e)
		}
	}
	return result, "", nil
}

func (r *fakeEventRepo) CreateOptions(_ context.Context, opts []event.Option) error {
	for _, o := range opts {
		r.options[o.EventID] = append(r.options[o.EventID], o)
	}
	return nil
}

func (r *fakeEventRepo) GetOptionsByEvent(_ context.Context, eventID string) ([]event.Option, error) {
	return r.options[eventID], nil
}

func (r *fakeEventRepo) GetOption(_ context.Context, id string) (*event.Option, error) {
	return nil, event.ErrOptionNotFound
}

func (r *fakeEventRepo) UpsertRSVP(_ context.Context, rsvp *event.RSVP) error {
	r.rsvps[rsvp.EventID] = append(r.rsvps[rsvp.EventID], *rsvp)
	return nil
}

func (r *fakeEventRepo) GetRSVPsByEvent(_ context.Context, eventID string) ([]event.RSVP, error) {
	return r.rsvps[eventID], nil
}

func (r *fakeEventRepo) GetUserRSVP(_ context.Context, _, _ string) (*event.RSVP, error) {
	return nil, event.ErrRSVPNotFound
}

func TestListGroupEventsResolver(t *testing.T) {
	ctx := context.Background()
	repo := newFakeEventRepo()

	repo.events["evt-1"] = &event.Event{ID: "evt-1", GroupID: "g1", Title: "Event 1", ScheduledTime: time.Now().Add(time.Hour)}
	repo.events["evt-2"] = &event.Event{ID: "evt-2", GroupID: "g1", Title: "Event 2", ScheduledTime: time.Now().Add(2 * time.Hour)}
	repo.options["evt-1"] = []event.Option{
		{ID: "opt-1", EventID: "evt-1", Label: "going"},
		{ID: "opt-2", EventID: "evt-1", Label: "not going"},
	}
	repo.rsvps["evt-1"] = []event.RSVP{
		{EventID: "evt-1", UserID: "u1", OptionID: "opt-1"},
		{EventID: "evt-1", UserID: "u2", OptionID: "opt-1"},
		{EventID: "evt-1", UserID: "u3", OptionID: "opt-2"},
	}

	resolver := NewListGroupEventsResolver(repo)
	result, _, err := resolver.Resolve(ctx, ListGroupEventsQuery{GroupID: "g1", Size: 10})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result))
	}

	for _, ew := range result {
		if ew.ID == "evt-1" {
			if len(ew.Options) != 2 {
				t.Errorf("expected 2 options for evt-1, got %d", len(ew.Options))
			}
			for _, o := range ew.Options {
				if o.Label == "going" && o.Tally != 2 {
					t.Errorf("going tally = %d, want 2", o.Tally)
				}
				if o.Label == "not going" && o.Tally != 1 {
					t.Errorf("not going tally = %d, want 1", o.Tally)
				}
			}
		}
	}
}
