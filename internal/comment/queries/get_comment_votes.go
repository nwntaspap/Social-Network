package queries

import (
	"context"

	"social-network/internal/comment"
)

type GetVoteCountsQuery struct {
	CommentID int
}

type GetVoteCountsResolver struct {
	repo comment.Repository
}

func NewGetVoteCountsResolver(repo comment.Repository) *GetVoteCountsResolver {
	return &GetVoteCountsResolver{repo: repo}
}

func (r *GetVoteCountsResolver) Resolve(ctx context.Context, q GetVoteCountsQuery) (*comment.VoteCounts, error) {
	return r.repo.GetVoteCounts(ctx, q.CommentID)
}
