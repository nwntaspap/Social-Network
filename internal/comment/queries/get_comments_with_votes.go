package queries

import (
	"context"

	"social-network/internal/comment"
)

type GetCommentsByTopicWithVotesQuery struct {
	TopicID     int
	RequesterID string
}

type GetCommentsByTopicWithVotesResolver struct {
	repo    comment.Repository
	topicOK TopicAccessChecker
}

func NewGetCommentsByTopicWithVotesResolver(repo comment.Repository, topicOK TopicAccessChecker) *GetCommentsByTopicWithVotesResolver {
	return &GetCommentsByTopicWithVotesResolver{repo: repo, topicOK: topicOK}
}

func (r *GetCommentsByTopicWithVotesResolver) Resolve(ctx context.Context, q GetCommentsByTopicWithVotesQuery) ([]comment.Comment, error) {
	if r.topicOK != nil {
		var userID *string
		if q.RequesterID != "" {
			userID = &q.RequesterID
		}
		if _, err := r.topicOK.GetTopicByID(ctx, q.TopicID, userID); err != nil {
			return nil, err
		}
	}
	return r.repo.GetCommentsByTopicIDWithVotes(ctx, q.TopicID, &q.RequesterID)
}
