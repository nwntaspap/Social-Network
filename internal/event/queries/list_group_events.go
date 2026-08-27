package queries

import (
	"context"
	"errors"

	"social-network/internal/event"
)

var ErrNotGroupMember = errors.New("user is not a member of this group")

// MemberChecker reports whether a user belongs to a group.
type MemberChecker interface {
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
}

type EventWithOptions struct {
	event.Event

	Options []OptionWithTally
}

type OptionWithTally struct {
	event.Option

	Tally int
}

type ListGroupEventsQuery struct {
	GroupID     string
	RequesterID string
	Cursor      string
	Size        int
}

type ListGroupEventsResolver struct {
	repo   event.Repository
	member MemberChecker
}

func NewListGroupEventsResolver(repo event.Repository, member MemberChecker) *ListGroupEventsResolver {
	return &ListGroupEventsResolver{repo: repo, member: member}
}

func (r *ListGroupEventsResolver) Resolve(ctx context.Context, q ListGroupEventsQuery) ([]EventWithOptions, string, error) {
	isMember, err := r.member.IsMember(ctx, q.GroupID, q.RequesterID)
	if err != nil {
		return nil, "", err
	}
	if !isMember {
		return nil, "", ErrNotGroupMember
	}

	size := q.Size
	if size <= 0 {
		size = 10
	}

	events, nextCursor, err := r.repo.ListGroupEvents(ctx, q.GroupID, q.Cursor, size)
	if err != nil {
		return nil, "", err
	}

	result := make([]EventWithOptions, len(events))
	for i, e := range events {
		opts, err := r.repo.GetOptionsByEvent(ctx, e.ID)
		if err != nil {
			return nil, "", err
		}

		rsvps, err := r.repo.GetRSVPsByEvent(ctx, e.ID)
		if err != nil {
			return nil, "", err
		}

		tallyMap := make(map[string]int)
		for _, rsvp := range rsvps {
			tallyMap[rsvp.OptionID]++
		}

		optsWithTally := make([]OptionWithTally, len(opts))
		for j, o := range opts {
			optsWithTally[j] = OptionWithTally{
				Option: o,
				Tally:  tallyMap[o.ID],
			}
		}

		result[i] = EventWithOptions{
			Event:   e,
			Options: optsWithTally,
		}
	}

	return result, nextCursor, nil
}
