package queries

import (
	"context"

	"social-network/internal/comment"
)

type GetCommentsByTopicWithVotesQuery struct {
	TopicID int
	UserID  string
}

type GetCommentsByTopicWithVotesResolver struct {
	repo comment.Repository
}

func NewGetCommentsByTopicWithVotesResolver(repo comment.Repository) *GetCommentsByTopicWithVotesResolver {
	return &GetCommentsByTopicWithVotesResolver{repo: repo}
}

func (r *GetCommentsByTopicWithVotesResolver) Resolve(ctx context.Context, q GetCommentsByTopicWithVotesQuery) ([]comment.Comment, error) {
	return r.repo.GetCommentsByTopicIDWithVotes(ctx, q.TopicID, &q.UserID)
}
