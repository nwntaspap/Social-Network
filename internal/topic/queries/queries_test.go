package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/topic"
)

type mockTopicRepo struct {
	getByIDFn    func(ctx context.Context, id int, userID *string) (*topic.Topic, error)
	getFeedFn    func(ctx context.Context, userID string, page, size int, orderBy, order, filter string) ([]topic.Topic, int, error)
	getByUserFn  func(ctx context.Context, ownerID, requesterID string, page, size int) ([]topic.Topic, int, error)
	getByGroupFn func(ctx context.Context, groupID string, page, size int) ([]topic.Topic, int, error)
	getCountsFn  func(ctx context.Context, topicID int) (*topic.VoteCounts, error)
}

func (m *mockTopicRepo) CreateTopic(_ context.Context, _ *topic.Topic, _ []string) error {
	return nil
}

func (m *mockTopicRepo) UpdateTopic(_ context.Context, _ *topic.Topic, _ []string) error { return nil }

func (m *mockTopicRepo) DeleteTopic(_ context.Context, _ string, _ int) error { return nil }

func (m *mockTopicRepo) GetTopicByID(ctx context.Context, id int, userID *string) (*topic.Topic, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id, userID)
	}
	return nil, topic.ErrTopicNotFound
}

func (m *mockTopicRepo) GetImagePathFromTopicID(_ context.Context, _ int, _ string) (string, error) {
	return "", nil
}

func (m *mockTopicRepo) GetFeed(ctx context.Context, userID string, page, size int, orderBy, order, filter string) ([]topic.Topic, int, error) {
	if m.getFeedFn != nil {
		return m.getFeedFn(ctx, userID, page, size, orderBy, order, filter)
	}
	return nil, 0, nil
}

func (m *mockTopicRepo) GetTopicsByUserID(ctx context.Context, ownerID, requesterID string, page, size int) ([]topic.Topic, int, error) {
	if m.getByUserFn != nil {
		return m.getByUserFn(ctx, ownerID, requesterID, page, size)
	}
	return nil, 0, nil
}

func (m *mockTopicRepo) GetTopicsByGroupID(ctx context.Context, groupID string, page, size int) ([]topic.Topic, int, error) {
	if m.getByGroupFn != nil {
		return m.getByGroupFn(ctx, groupID, page, size)
	}
	return nil, 0, nil
}

func (m *mockTopicRepo) CastVote(_ context.Context, _ string, _ int, _ int) (topic.VoteChange, error) {
	return topic.VoteChangeAdded, nil
}
func (m *mockTopicRepo) DeleteVote(_ context.Context, _ string, _ int) error { return nil }
func (m *mockTopicRepo) GetVoteCounts(ctx context.Context, topicID int) (*topic.VoteCounts, error) {
	if m.getCountsFn != nil {
		return m.getCountsFn(ctx, topicID)
	}
	return &topic.VoteCounts{}, nil
}
func (m *mockTopicRepo) GetPostCount(_ context.Context, _ string) (int, error) { return 0, nil }
func (m *mockTopicRepo) GetVoteCount(_ context.Context, _ string) (int, error) { return 0, nil }

func TestGetFeed_Resolve(t *testing.T) {
	repo := &mockTopicRepo{
		getFeedFn: func(_ context.Context, _ string, _, _ int, _, _, _ string) ([]topic.Topic, int, error) {
			return []topic.Topic{{ID: 1, Title: "Feed"}}, 1, nil
		},
	}
	r := NewGetFeedResolver(repo)

	res, err := r.Resolve(context.Background(), GetFeedQuery{UserID: "u1", Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.Total != 1 {
		t.Errorf("Total = %d, want 1", res.Total)
	}
	if res.Topics[0].Title != "Feed" {
		t.Errorf("Title = %q, want %q", res.Topics[0].Title, "Feed")
	}
}

func TestGetFeed_Error(t *testing.T) {
	repo := &mockTopicRepo{
		getFeedFn: func(_ context.Context, _ string, _, _ int, _, _, _ string) ([]topic.Topic, int, error) {
			return nil, 0, errors.New("db error")
		},
	}
	r := NewGetFeedResolver(repo)

	_, err := r.Resolve(context.Background(), GetFeedQuery{UserID: "u1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetTopic_Resolve(t *testing.T) {
	repo := &mockTopicRepo{
		getByIDFn: func(_ context.Context, id int, _ *string) (*topic.Topic, error) {
			return &topic.Topic{ID: id, Title: "Hello"}, nil
		},
	}
	r := NewGetTopicResolver(repo)

	top, err := r.Resolve(context.Background(), GetTopicQuery{TopicID: 1})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if top.Title != "Hello" {
		t.Errorf("Title = %q, want %q", top.Title, "Hello")
	}
}

func TestGetTopic_NotFound(t *testing.T) {
	r := NewGetTopicResolver(&mockTopicRepo{})

	_, err := r.Resolve(context.Background(), GetTopicQuery{TopicID: 999})
	if err == nil {
		t.Fatal("expected error for missing topic")
	}
}

func TestGetTopicsByUser_Resolve(t *testing.T) {
	repo := &mockTopicRepo{
		getByUserFn: func(_ context.Context, _, _ string, _, _ int) ([]topic.Topic, int, error) {
			return []topic.Topic{{ID: 1}}, 1, nil
		},
	}
	r := NewGetTopicsByUserResolver(repo)

	res, err := r.Resolve(context.Background(), GetTopicsByUserQuery{OwnerID: "u1", RequesterID: "u1"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.Total != 1 {
		t.Errorf("Total = %d, want 1", res.Total)
	}
}

func TestGetTopicsByGroup_Resolve(t *testing.T) {
	repo := &mockTopicRepo{
		getByGroupFn: func(_ context.Context, _ string, _, _ int) ([]topic.Topic, int, error) {
			return []topic.Topic{{ID: 2}}, 1, nil
		},
	}
	r := NewGetTopicsByGroupResolver(repo)

	res, err := r.Resolve(context.Background(), GetTopicsByGroupQuery{GroupID: "g1"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.Total != 1 {
		t.Errorf("Total = %d, want 1", res.Total)
	}
}

func TestGetVoteCounts_Resolve(t *testing.T) {
	repo := &mockTopicRepo{
		getByIDFn: func(_ context.Context, _ int, _ *string) (*topic.Topic, error) {
			return &topic.Topic{ID: 1}, nil
		},
		getCountsFn: func(_ context.Context, _ int) (*topic.VoteCounts, error) {
			return &topic.VoteCounts{Upvotes: 3, Downvotes: 1, Score: 2}, nil
		},
	}
	r := NewGetVoteCountsResolver(repo)

	vc, err := r.Resolve(context.Background(), GetVoteCountsQuery{TopicID: 1})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if vc.Score != 2 {
		t.Errorf("Score = %d, want 2", vc.Score)
	}
}

func TestGetVoteCounts_NotVisible(t *testing.T) {
	r := NewGetVoteCountsResolver(&mockTopicRepo{})

	_, err := r.Resolve(context.Background(), GetVoteCountsQuery{TopicID: 999})
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("err = %v, want ErrTopicNotFound", err)
	}
}
