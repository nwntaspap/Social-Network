package queries

import (
	"context"

	"social-network/internal/comment"
)

type GetCommentsByTopicQuery struct {
	TopicID     int
	RequesterID string
}

type GetCommentsByTopicResolver struct {
	repo    comment.Repository
	topicOK TopicAccessChecker
}

func NewGetCommentsByTopicResolver(repo comment.Repository, topicOK TopicAccessChecker) *GetCommentsByTopicResolver {
	return &GetCommentsByTopicResolver{repo: repo, topicOK: topicOK}
}

func (r *GetCommentsByTopicResolver) Resolve(ctx context.Context, q GetCommentsByTopicQuery) ([]comment.Comment, error) {
	if err := r.checkAccess(ctx, q.TopicID, q.RequesterID); err != nil {
		return nil, err
	}
	return r.repo.GetCommentsByTopicID(ctx, q.TopicID)
}

func (r *GetCommentsByTopicResolver) checkAccess(ctx context.Context, topicID int, requesterID string) error {
	if r.topicOK == nil {
		return nil
	}
	var userID *string
	if requesterID != "" {
		userID = &requesterID
	}
	_, err := r.topicOK.GetTopicByID(ctx, topicID, userID)
	return err
}
