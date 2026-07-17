package queries

import (
	"context"

	"social-network/internal/topic"
)

type GetTopicsByUserQuery struct {
	OwnerID     string
	RequesterID string
	Page        int
	Size        int
}

type GetTopicsByUserResult struct {
	Topics []topic.Topic
	Total  int
}

type GetTopicsByUserResolver struct {
	repo topic.Repository
}

func NewGetTopicsByUserResolver(repo topic.Repository) *GetTopicsByUserResolver {
	return &GetTopicsByUserResolver{repo: repo}
}

func (r *GetTopicsByUserResolver) Resolve(ctx context.Context, q GetTopicsByUserQuery) (*GetTopicsByUserResult, error) {
	topics, total, err := r.repo.GetTopicsByUserID(ctx, q.OwnerID, q.RequesterID, q.Page, q.Size)
	if err != nil {
		return nil, err
	}
	return &GetTopicsByUserResult{Topics: topics, Total: total}, nil
}
