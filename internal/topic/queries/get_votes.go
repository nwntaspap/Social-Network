package queries

import (
	"context"

	"social-network/internal/topic"
)

type GetVoteCountsQuery struct {
	TopicID     int
	RequesterID *string
}

type GetVoteCountsResolver struct {
	repo topic.Repository
}

func NewGetVoteCountsResolver(repo topic.Repository) *GetVoteCountsResolver {
	return &GetVoteCountsResolver{repo: repo}
}

func (r *GetVoteCountsResolver) Resolve(ctx context.Context, q GetVoteCountsQuery) (*topic.VoteCounts, error) {
	if _, err := r.repo.GetTopicByID(ctx, q.TopicID, q.RequesterID); err != nil {
		return nil, err
	}
	return r.repo.GetVoteCounts(ctx, q.TopicID)
}
