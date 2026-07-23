package transport

import (
	"context"
	"net/http"
	"time"

	"social-network/internal/topic"
	"social-network/internal/topic/commands"
	"social-network/internal/topic/queries"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

type CreateTopicExecutor interface {
	Execute(ctx context.Context, cmd commands.CreateTopicCommand) (*topic.Topic, error)
}

type UpdateTopicExecutor interface {
	Execute(ctx context.Context, cmd commands.UpdateTopicCommand) (*topic.Topic, error)
}

type DeleteTopicExecutor interface {
	Execute(ctx context.Context, cmd commands.DeleteTopicCommand) error
}

type CastVoteExecutor interface {
	Execute(ctx context.Context, cmd commands.CastVoteCommand) error
}

type GetFeedResolver interface {
	Resolve(ctx context.Context, q queries.GetFeedQuery) (*queries.GetFeedResult, error)
}

type GetTopicResolver interface {
	Resolve(ctx context.Context, q queries.GetTopicQuery) (*topic.Topic, error)
}

type GetTopicsByUserResolver interface {
	Resolve(ctx context.Context, q queries.GetTopicsByUserQuery) (*queries.GetTopicsByUserResult, error)
}

type GetTopicsByGroupResolver interface {
	Resolve(ctx context.Context, q queries.GetTopicsByGroupQuery) (*queries.GetTopicsByGroupResult, error)
}

type GetVoteCountsResolver interface {
	Resolve(ctx context.Context, q queries.GetVoteCountsQuery) (*topic.VoteCounts, error)
}

type TopicResponse struct {
	ID            int     `json:"id"`
	UserID        string  `json:"userId"`
	GroupID       *string `json:"groupId"`
	Title         string  `json:"title"`
	Content       string  `json:"content"`
	ImagePath     string  `json:"imagePath,omitempty"`
	Visibility    string  `json:"privacy"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
	OwnerUsername string  `json:"ownerUsername,omitempty"`
	UpvoteCount   int     `json:"likesCount"`
	UserVote      *int    `json:"isLiked"`
}

type VoteCountsResponse struct {
	Upvotes   int `json:"upvotes"`
	Downvotes int `json:"downvotes"`
	Score     int `json:"score"`
}

func toTopicResponse(t *topic.Topic) TopicResponse {
	var vis string
	switch t.Visibility {
	case topic.VisibilityPublic:
		vis = "public"
	case topic.VisibilityFollowers:
		vis = "followers"
	case topic.VisibilityPrivate:
		vis = "private"
	default:
		vis = "public"
	}
	return TopicResponse{
		ID:            t.ID,
		UserID:        t.UserID,
		GroupID:       t.GroupID,
		Title:         t.Title,
		Content:       t.Content,
		ImagePath:     t.ImagePath,
		Visibility:    vis,
		CreatedAt:     t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     t.UpdatedAt.Format(time.RFC3339),
		OwnerUsername: t.OwnerUsername,
		UpvoteCount:   t.UpvoteCount,
		UserVote:      t.UserVote,
	}
}

type Handler struct {
	createTopic CreateTopicExecutor
	updateTopic UpdateTopicExecutor
	deleteTopic DeleteTopicExecutor
	castVote    CastVoteExecutor
	getFeed     GetFeedResolver
	getTopic    GetTopicResolver
	getByUser   GetTopicsByUserResolver
	getByGroup  GetTopicsByGroupResolver
	getVotes    GetVoteCountsResolver
	extractUser UserExtractor
}

func NewHandler(
	extractUser UserExtractor,
	createTopic CreateTopicExecutor,
	updateTopic UpdateTopicExecutor,
	deleteTopic DeleteTopicExecutor,
	castVote CastVoteExecutor,
	getFeed GetFeedResolver,
	getTopic GetTopicResolver,
	getByUser GetTopicsByUserResolver,
	getByGroup GetTopicsByGroupResolver,
	getVotes GetVoteCountsResolver,
) *Handler {
	return &Handler{
		createTopic: createTopic,
		updateTopic: updateTopic,
		deleteTopic: deleteTopic,
		castVote:    castVote,
		getFeed:     getFeed,
		getTopic:    getTopic,
		getByUser:   getByUser,
		getByGroup:  getByGroup,
		getVotes:    getVotes,
		extractUser: extractUser,
	}
}
