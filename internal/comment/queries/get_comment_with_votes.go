package queries

import (
	"context"

	"social-network/internal/comment"
)

type GetCommentByIDWithVotesQuery struct {
	CommentID int
	UserID    string
}

type GetCommentByIDWithVotesResolver struct {
	repo comment.Repository
}

func NewGetCommentByIDWithVotesResolver(repo comment.Repository) *GetCommentByIDWithVotesResolver {
	return &GetCommentByIDWithVotesResolver{repo: repo}
}

func (r *GetCommentByIDWithVotesResolver) Resolve(ctx context.Context, q GetCommentByIDWithVotesQuery) (*comment.Comment, error) {
	return r.repo.GetCommentByIDWithVotes(ctx, q.CommentID, &q.UserID)
}
