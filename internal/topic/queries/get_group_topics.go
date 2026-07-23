package queries

import (
	"context"

	"social-network/internal/topic"
)

type GetTopicsByGroupQuery struct {
	GroupID string
	Page    int
	Size    int
}

type GetTopicsByGroupResult struct {
	Topics []topic.Topic
	Total  int
}

type GetTopicsByGroupResolver struct {
	repo topic.Repository
}

func NewGetTopicsByGroupResolver(repo topic.Repository) *GetTopicsByGroupResolver {
	return &GetTopicsByGroupResolver{repo: repo}
}

func (r *GetTopicsByGroupResolver) Resolve(ctx context.Context, q GetTopicsByGroupQuery) (*GetTopicsByGroupResult, error) {
	topics, total, err := r.repo.GetTopicsByGroupID(ctx, q.GroupID, q.Page, q.Size)
	if err != nil {
		return nil, err
	}
	return &GetTopicsByGroupResult{Topics: topics, Total: total}, nil
}
