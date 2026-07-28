package queries

import (
	"context"

	"social-network/internal/group"
)

type GetGroupPostCommentsQuery struct {
	PostID string
	Page   int
	Size   int
}

type GetGroupPostCommentsResult struct {
	Comments []group.PostComment
	Total    int
}

type GetGroupPostCommentsResolver struct {
	repo group.Repository
}

func NewGetGroupPostCommentsResolver(repo group.Repository) *GetGroupPostCommentsResolver {
	return &GetGroupPostCommentsResolver{repo: repo}
}

func (r *GetGroupPostCommentsResolver) Resolve(ctx context.Context, q GetGroupPostCommentsQuery) (*GetGroupPostCommentsResult, error) {
	comments, total, err := r.repo.GetPostComments(ctx, q.PostID, q.Page, q.Size)
	if err != nil {
		return nil, err
	}
	return &GetGroupPostCommentsResult{Comments: comments, Total: total}, nil
}
