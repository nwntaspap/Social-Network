package transport

import (
	"context"
	"net/http"

	"social-network/internal/follow"
	"social-network/internal/follow/commands"
	"social-network/internal/follow/queries"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

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

// UserLookup is a local interface for nested user lookups (avoids importing domain/user).
type UserLookup interface {
	GetUserByID(ctx context.Context, id string) (*UserResult, error)
}

type FollowUserExecutor interface {
	Execute(ctx context.Context, cmd commands.FollowUserCommand) (commands.FollowUserResult, error)
}

type UnfollowUserExecutor interface {
	Execute(ctx context.Context, cmd commands.UnfollowUserCommand) error
}

type AcceptRequestExecutor interface {
	Execute(ctx context.Context, cmd commands.AcceptRequestCommand) error
}

type DeclineRequestExecutor interface {
	Execute(ctx context.Context, cmd commands.DeclineRequestCommand) error
}

type FollowersResolver interface {
	Resolve(ctx context.Context, q queries.GetFollowersQuery) ([]follow.Follow, error)
}

type FollowingResolver interface {
	Resolve(ctx context.Context, q queries.GetFollowingQuery) ([]follow.Follow, error)
}

type PendingRequestsResolver interface {
	Resolve(ctx context.Context, q queries.GetPendingRequestsQuery) ([]follow.Request, error)
}

type ConnectedResolver interface {
	Resolve(ctx context.Context, q queries.AreConnectedQuery) (bool, error)
}

type Handler struct {
	followUser     FollowUserExecutor
	unfollowUser   UnfollowUserExecutor
	acceptRequest  AcceptRequestExecutor
	declineRequest DeclineRequestExecutor
	getFollowers   FollowersResolver
	getFollowing   FollowingResolver
	getPendingReqs PendingRequestsResolver
	areConnected   ConnectedResolver
	extractUser    UserExtractor
	userLookup     UserLookup
}

func NewHandler(
	extractUser UserExtractor,
	userLookup UserLookup,
	followUser FollowUserExecutor,
	unfollowUser UnfollowUserExecutor,
	acceptRequest AcceptRequestExecutor,
	declineRequest DeclineRequestExecutor,
	getFollowers FollowersResolver,
	getFollowing FollowingResolver,
	getPendingReqs PendingRequestsResolver,
	areConnected ConnectedResolver,
) *Handler {
	return &Handler{
		extractUser:    extractUser,
		userLookup:     userLookup,
		followUser:     followUser,
		unfollowUser:   unfollowUser,
		acceptRequest:  acceptRequest,
		declineRequest: declineRequest,
		getFollowers:   getFollowers,
		getFollowing:   getFollowing,
		getPendingReqs: getPendingReqs,
		areConnected:   areConnected,
	}
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
