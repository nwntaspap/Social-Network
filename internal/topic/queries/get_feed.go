package queries

import (
	"context"

	"social-network/internal/topic"
)

type GetFeedQuery struct {
	UserID  string
	Page    int
	Size    int
	OrderBy string
	Order   string
	Filter  string
}

type GetFeedResult struct {
	Topics []topic.Topic
	Total  int
}

type GetFeedResolver struct {
	repo topic.Repository
}

func NewGetFeedResolver(repo topic.Repository) *GetFeedResolver {
	return &GetFeedResolver{repo: repo}
}

func (r *GetFeedResolver) Resolve(ctx context.Context, q GetFeedQuery) (*GetFeedResult, error) {
	topics, total, err := r.repo.GetFeed(ctx, q.UserID, q.Page, q.Size, q.OrderBy, q.Order, q.Filter)
	if err != nil {
		return nil, err
	}
	return &GetFeedResult{Topics: topics, Total: total}, nil
}
