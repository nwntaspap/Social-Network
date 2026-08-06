package queries

import (
	"context"
	"testing"
	"time"

	"social-network/internal/event"
)

func TestListEventRSVPsResolver(t *testing.T) {
	ctx := context.Background()
	repo := newFakeEventRepo()

	repo.events["evt-1"] = &event.Event{ID: "evt-1", GroupID: "g1", Title: "Event 1", ScheduledTime: time.Now().Add(time.Hour)}
	repo.options["evt-1"] = []event.Option{
		{ID: "opt-1", EventID: "evt-1", Label: "going"},
		{ID: "opt-2", EventID: "evt-1", Label: "not going"},
	}
	repo.rsvps["evt-1"] = []event.RSVP{
		{EventID: "evt-1", UserID: "u1", OptionID: "opt-1"},
		{EventID: "evt-1", UserID: "u2", OptionID: "opt-1"},
		{EventID: "evt-1", UserID: "u3", OptionID: "opt-2"},
	}

	resolver := NewListEventRSVPsResolver(repo)
	result, err := resolver.Resolve(ctx, ListEventRSVPsQuery{EventID: "evt-1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 options, got %d", len(result))
	}

	byLabel := map[string]OptionRSVPs{}
	for _, o := range result {
		byLabel[o.Label] = o
	}

	if got := byLabel["going"].UserIDs; len(got) != 2 || got[0] != "u1" || got[1] != "u2" {
		t.Errorf("going users = %v, want [u1 u2]", got)
	}
	if got := byLabel["not going"].UserIDs; len(got) != 1 || got[0] != "u3" {
		t.Errorf("not going users = %v, want [u3]", got)
	}
}

func TestListEventRSVPsResolver_Empty(t *testing.T) {
	ctx := context.Background()
	repo := newFakeEventRepo()
	repo.events["evt-1"] = &event.Event{ID: "evt-1", GroupID: "g1", Title: "Event 1", ScheduledTime: time.Now().Add(time.Hour)}

	resolver := NewListEventRSVPsResolver(repo)
	result, err := resolver.Resolve(ctx, ListEventRSVPsQuery{EventID: "evt-1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 options, got %d", len(result))
	}
}

func TestListEventRSVPsResolver_EventNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeEventRepo()

	resolver := NewListEventRSVPsResolver(repo)
	_, err := resolver.Resolve(ctx, ListEventRSVPsQuery{EventID: "missing"})
	if err == nil {
		t.Fatal("expected error for missing event")
	}
}
