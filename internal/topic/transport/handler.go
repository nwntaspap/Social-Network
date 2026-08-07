package transport

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"social-network/internal/topic"
	"social-network/internal/topic/commands"
	"social-network/internal/topic/queries"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

type UserLookup interface {
	GetUserByID(ctx context.Context, id string) (*UserResult, error)
}

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

type DeleteVoteExecutor interface {
	Execute(ctx context.Context, cmd commands.DeleteVoteCommand) error
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

type UserResult struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Nickname    string `json:"nickname,omitempty"`
	AboutMe     string `json:"aboutMe,omitempty"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	DateOfBirth string `json:"dateOfBirth"`
	IsPublic    bool   `json:"isPublic"`
	CreatedAt   string `json:"createdAt"`
}

func (h *Handler) lookupUser(ctx context.Context, userID string) *UserResult {
	if h.userLookup == nil || userID == "" {
		return nil
	}
	u, err := h.userLookup.GetUserByID(ctx, userID)
	if err != nil {
		return nil
	}
	return u
}

type TopicResponse struct {
	ID            string      `json:"id"`
	UserID        string      `json:"userId"`
	User          *UserResult `json:"user"`
	GroupID       *string     `json:"groupId"`
	Title         string      `json:"title"`
	Content       string      `json:"content"`
	ImageURL      string      `json:"imageUrl,omitempty"`
	Visibility    string      `json:"privacy"`
	CreatedAt     string      `json:"createdAt"`
	UpdatedAt     string      `json:"updatedAt"`
	UpvoteCount   int         `json:"likesCount"`
	DownvoteCount int         `json:"downvotesCount"`
	CommentsCount int         `json:"commentsCount"`
	UserVote      *int        `json:"isLiked"`
	AllowedUsers  []string    `json:"allowedUsers,omitempty"`
}

type VoteCountsResponse struct {
	Upvotes   int `json:"upvotes"`
	Downvotes int `json:"downvotes"`
	Score     int `json:"score"`
}

func paginatedPayload(data any, total, page, limit int) map[string]any {
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}
	return map[string]any{
		"data":       data,
		"page":       page,
		"pageSize":   limit,
		"totalCount": total,
		"totalPages": totalPages,
	}
}

func toTopicResponse(t *topic.Topic, user *UserResult) TopicResponse {
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
		ID:            strconv.Itoa(t.ID),
		UserID:        t.UserID,
		User:          user,
		GroupID:       t.GroupID,
		Title:         t.Title,
		Content:       t.Content,
		ImageURL:      t.ImagePath,
		Visibility:    vis,
		CreatedAt:     t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     t.UpdatedAt.Format(time.RFC3339),
		UpvoteCount:   t.UpvoteCount,
		DownvoteCount: t.DownvoteCount,
		CommentsCount: t.CommentsCount,
		UserVote:      t.UserVote,
		AllowedUsers:  t.AllowedUsers,
	}
}

type Handler struct {
	createTopic CreateTopicExecutor
	updateTopic UpdateTopicExecutor
	deleteTopic DeleteTopicExecutor
	castVote    CastVoteExecutor
	deleteVote  DeleteVoteExecutor
	getFeed     GetFeedResolver
	getTopic    GetTopicResolver
	getByUser   GetTopicsByUserResolver
	getByGroup  GetTopicsByGroupResolver
	getVotes    GetVoteCountsResolver
	userLookup  UserLookup
	extractUser UserExtractor
}

func NewHandler(
	extractUser UserExtractor,
	userLookup UserLookup,
	createTopic CreateTopicExecutor,
	updateTopic UpdateTopicExecutor,
	deleteTopic DeleteTopicExecutor,
	castVote CastVoteExecutor,
	deleteVote DeleteVoteExecutor,
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
		deleteVote:  deleteVote,
		getFeed:     getFeed,
		getTopic:    getTopic,
		getByUser:   getByUser,
		getByGroup:  getByGroup,
		getVotes:    getVotes,
		userLookup:  userLookup,
		extractUser: extractUser,
	}
}
