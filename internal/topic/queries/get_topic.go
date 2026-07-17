package queries

import (
	"context"
	"fmt"

	"social-network/internal/topic"
)

type GetTopicQuery struct {
	TopicID int
	UserID  *string
}

type GetTopicResolver struct {
	repo topic.Repository
}

func NewGetTopicResolver(repo topic.Repository) *GetTopicResolver {
	return &GetTopicResolver{repo: repo}
}

func (r *GetTopicResolver) Resolve(ctx context.Context, q GetTopicQuery) (*topic.Topic, error) {
	t, err := r.repo.GetTopicByID(ctx, q.TopicID, q.UserID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("%w: %d", topic.ErrTopicNotFound, q.TopicID)
	}
	return t, nil
}
