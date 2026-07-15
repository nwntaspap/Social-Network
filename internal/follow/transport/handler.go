package transport

import (
	"context"
	"net/http"

	"social-network/internal/follow"
	"social-network/internal/follow/commands"
	"social-network/internal/follow/queries"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

type FollowUserExecutor interface {
	Execute(ctx context.Context, cmd commands.FollowUserCommand) error
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
	acceptRequest  AcceptRequestExecutor
	declineRequest DeclineRequestExecutor
	getFollowers   FollowersResolver
	getFollowing   FollowingResolver
	getPendingReqs PendingRequestsResolver
	areConnected   ConnectedResolver
	extractUser    UserExtractor
}

func NewHandler(
	extractUser UserExtractor,
	followUser FollowUserExecutor,
	acceptRequest AcceptRequestExecutor,
	declineRequest DeclineRequestExecutor,
	getFollowers FollowersResolver,
	getFollowing FollowingResolver,
	getPendingReqs PendingRequestsResolver,
	areConnected ConnectedResolver,
) *Handler {
	return &Handler{
		extractUser:    extractUser,
		followUser:     followUser,
		acceptRequest:  acceptRequest,
		declineRequest: declineRequest,
		getFollowers:   getFollowers,
		getFollowing:   getFollowing,
		getPendingReqs: getPendingReqs,
		areConnected:   areConnected,
	}
}
