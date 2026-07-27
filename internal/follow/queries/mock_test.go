package queries

import (
	"context"

	"social-network/internal/follow"
)

type mockRepo struct {
	getFollowersResult []follow.Follow
	getFollowersErr    error
	getFollowingResult []follow.Follow
	getFollowingErr    error
	getPendingResult   []follow.Request
	getPendingErr      error
	areConnectedResult bool
	areConnectedErr    error
}

func (m *mockRepo) CreateFollow(_ context.Context, _ *follow.Follow) error         { return nil }
func (m *mockRepo) DeleteFollow(_ context.Context, _, _ string) error              { return nil }
func (m *mockRepo) CreateFollowRequest(_ context.Context, _ *follow.Request) error { return nil }
func (m *mockRepo) DeleteFollowRequest(_ context.Context, _, _ string) error       { return nil }
func (m *mockRepo) WithTx(_ context.Context, fn func(follow.Repository) error) error {
	return fn(m)
}

func (m *mockRepo) GetFollowers(_ context.Context, _ string) ([]follow.Follow, error) {
	return m.getFollowersResult, m.getFollowersErr
}

func (m *mockRepo) GetFollowing(_ context.Context, _ string) ([]follow.Follow, error) {
	return m.getFollowingResult, m.getFollowingErr
}

func (m *mockRepo) GetPendingRequests(_ context.Context, _ string) ([]follow.Request, error) {
	return m.getPendingResult, m.getPendingErr
}

func (m *mockRepo) AreConnected(_ context.Context, _, _ string) (bool, error) {
	return m.areConnectedResult, m.areConnectedErr
}

func (m *mockRepo) GetFollowerCount(_ context.Context, _ string) (int, error)  { return 0, nil }
func (m *mockRepo) GetFollowingCount(_ context.Context, _ string) (int, error) { return 0, nil }
