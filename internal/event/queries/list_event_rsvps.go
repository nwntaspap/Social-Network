package queries

import (
	"context"

	"social-network/internal/event"
)

type OptionRSVPs struct {
	OptionID string
	Label    string
	UserIDs  []string
}

type ListEventRSVPsQuery struct {
	EventID string
}

type ListEventRSVPsResolver struct {
	repo event.Repository
}

func NewListEventRSVPsResolver(repo event.Repository) *ListEventRSVPsResolver {
	return &ListEventRSVPsResolver{repo: repo}
}

func (r *ListEventRSVPsResolver) Resolve(ctx context.Context, q ListEventRSVPsQuery) ([]OptionRSVPs, error) {
	if _, err := r.repo.GetEvent(ctx, q.EventID); err != nil {
		return nil, err
	}

	opts, err := r.repo.GetOptionsByEvent(ctx, q.EventID)
	if err != nil {
		return nil, err
	}

	rsvps, err := r.repo.GetRSVPsByEvent(ctx, q.EventID)
	if err != nil {
		return nil, err
	}

	result := make([]OptionRSVPs, 0, len(opts))
	byOption := make(map[string]int, len(opts))
	for _, o := range opts {
		byOption[o.ID] = len(result)
		result = append(result, OptionRSVPs{OptionID: o.ID, Label: o.Label})
	}

	for _, rsvp := range rsvps {
		if idx, ok := byOption[rsvp.OptionID]; ok {
			result[idx].UserIDs = append(result[idx].UserIDs, rsvp.UserID)
		}
	}

	return result, nil
}
