package transport

import (
	"context"
	"net/http"

	"social-network/internal/follow"
	"social-network/internal/follow/commands"
	"social-network/internal/follow/queries"
)

type mockFollowUser struct {
	result commands.FollowUserResult
	err    error
}

func (m *mockFollowUser) Execute(_ context.Context, _ commands.FollowUserCommand) (commands.FollowUserResult, error) {
	return m.result, m.err
}

type mockUnfollowUser struct {
	err error
}

func (m *mockUnfollowUser) Execute(_ context.Context, _ commands.UnfollowUserCommand) error {
	return m.err
}

type mockAcceptRequest struct {
	err error
}

func (m *mockAcceptRequest) Execute(_ context.Context, _ commands.AcceptRequestCommand) error {
	return m.err
}

type mockDeclineRequest struct {
	err error
}

func (m *mockDeclineRequest) Execute(_ context.Context, _ commands.DeclineRequestCommand) error {
	return m.err
}

type mockGetFollowers struct {
	result []follow.Follow
	err    error
}

func (m *mockGetFollowers) Resolve(_ context.Context, _ queries.GetFollowersQuery) ([]follow.Follow, error) {
	return m.result, m.err
}

type mockGetFollowing struct {
	result []follow.Follow
	err    error
}

func (m *mockGetFollowing) Resolve(_ context.Context, _ queries.GetFollowingQuery) ([]follow.Follow, error) {
	return m.result, m.err
}

type mockGetPendingRequests struct {
	result []follow.Request
	err    error
}

func (m *mockGetPendingRequests) Resolve(_ context.Context, _ queries.GetPendingRequestsQuery) ([]follow.Request, error) {
	return m.result, m.err
}

type mockAreConnected struct {
	result bool
	err    error
}

func (m *mockAreConnected) Resolve(_ context.Context, _ queries.AreConnectedQuery) (bool, error) {
	return m.result, m.err
}

type mockUserLookup struct {
	byID map[string]*UserResult
	err  error
}

func (m *mockUserLookup) GetUserByID(_ context.Context, id string) (*UserResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if u, ok := m.byID[id]; ok {
		return u, nil
	}
	return &UserResult{ID: id, Username: "user-" + id}, nil
}

func testExtractor(userID string) UserExtractor {
	return func(_ *http.Request) (string, bool) {
		if userID == "" {
			return "", false
		}
		return userID, true
	}
}
