package queries

import (
	"context"

	"social-network/internal/topic"
)

// TopicAccessChecker verifies that the requester may view a topic.
// Implementations must enforce visibility rules and return
// topic.ErrTopicNotFound when the topic is hidden or absent.
type TopicAccessChecker interface {
	GetTopicByID(ctx context.Context, topicID int, userID *string) (*topic.Topic, error)
}
