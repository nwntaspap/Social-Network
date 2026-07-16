package queries

import (
	"context"

	"social-network/internal/comment"
)

type GetCommentsByTopicQuery struct {
	TopicID int
	UserID  *string
}

type GetCommentsByTopicResolver struct {
	repo comment.Repository
}

func NewGetCommentsByTopicResolver(repo comment.Repository) *GetCommentsByTopicResolver {
	return &GetCommentsByTopicResolver{repo: repo}
}

func (r *GetCommentsByTopicResolver) Resolve(ctx context.Context, q GetCommentsByTopicQuery) ([]comment.Comment, error) {
	if q.UserID != nil {
		return r.repo.GetCommentsByTopicIDWithVotes(ctx, q.TopicID, q.UserID)
	}
	return r.repo.GetCommentsByTopicID(ctx, q.TopicID)
}
