package queries

import (
	"context"

	"social-network/internal/comment"
)

type GetCommentsByTopicQuery struct {
	TopicID int
}

type GetCommentsByTopicResolver struct {
	repo comment.Repository
}

func NewGetCommentsByTopicResolver(repo comment.Repository) *GetCommentsByTopicResolver {
	return &GetCommentsByTopicResolver{repo: repo}
}

func (r *GetCommentsByTopicResolver) Resolve(ctx context.Context, q GetCommentsByTopicQuery) ([]comment.Comment, error) {
	return r.repo.GetCommentsByTopicID(ctx, q.TopicID)
}
