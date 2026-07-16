package queries

import (
	"context"

	"social-network/internal/comment"
)

type GetCommentByIDQuery struct {
	CommentID int
	UserID    *string
}

type GetCommentByIDResolver struct {
	repo comment.Repository
}

func NewGetCommentByIDResolver(repo comment.Repository) *GetCommentByIDResolver {
	return &GetCommentByIDResolver{repo: repo}
}

func (r *GetCommentByIDResolver) Resolve(ctx context.Context, q GetCommentByIDQuery) (*comment.Comment, error) {
	if q.UserID != nil {
		return r.repo.GetCommentByIDWithVotes(ctx, q.CommentID, q.UserID)
	}
	return r.repo.GetCommentByID(ctx, q.CommentID)
}
