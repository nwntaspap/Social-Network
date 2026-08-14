package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
	"social-network/internal/platform/eventbus"
	"social-network/internal/topic"
	"social-network/internal/user"
)

type mockUserRepo struct{}

func (m *mockUserRepo) Create(_ context.Context, _ *user.User) error { return nil }
func (m *mockUserRepo) GetByID(_ context.Context, id string) (*user.User, error) {
	return &user.User{Nickname: id + "-name", AvatarPath: ""}, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, errUserNotFound
}

func (m *mockUserRepo) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return nil, errUserNotFound
}
func (m *mockUserRepo) Update(_ context.Context, _ *user.User) error            { return nil }
func (m *mockUserRepo) TogglePrivacy(_ context.Context, _ string, _ bool) error { return nil }
func (m *mockUserRepo) ListAll(_ context.Context) ([]user.User, error)          { return nil, nil }

type mockTopicRepo struct{}

func (m *mockTopicRepo) CreateTopic(_ context.Context, _ *topic.Topic, _ []string) error { return nil }

func (m *mockTopicRepo) UpdateTopic(_ context.Context, _ *topic.Topic, _ []string) error { return nil }

func (m *mockTopicRepo) DeleteTopic(_ context.Context, _ string, _ int) error { return nil }

func (m *mockTopicRepo) GetTopicByID(_ context.Context, id int, _ *string) (*topic.Topic, error) {
	return &topic.Topic{ID: id, UserID: "author-1"}, nil
}

func (m *mockTopicRepo) GetImagePathFromTopicID(_ context.Context, _ int, _ string) (string, error) {
	return "", nil
}

func (m *mockTopicRepo) GetFeed(_ context.Context, _ string, _, _ int, _, _, _ string) ([]topic.Topic, int, error) {
	return nil, 0, nil
}

func (m *mockTopicRepo) GetTopicsByUserID(_ context.Context, _, _ string, _, _ int) ([]topic.Topic, int, error) {
	return nil, 0, nil
}

func (m *mockTopicRepo) GetTopicsByGroupID(_ context.Context, _ string, _, _ int) ([]topic.Topic, int, error) {
	return nil, 0, nil
}

func (m *mockTopicRepo) CastVote(_ context.Context, _ string, _ int, _ int) (topic.VoteChange, error) {
	return topic.VoteChangeAdded, nil
}
func (m *mockTopicRepo) DeleteVote(_ context.Context, _ string, _ int) error { return nil }
func (m *mockTopicRepo) GetVoteCounts(_ context.Context, _ int) (*topic.VoteCounts, error) {
	return &topic.VoteCounts{}, nil
}
func (m *mockTopicRepo) GetPostCount(_ context.Context, _ string) (int, error) { return 0, nil }
func (m *mockTopicRepo) GetVoteCount(_ context.Context, _ string) (int, error) { return 0, nil }

type mockRepo struct {
	createErr error
	updateErr error
	deleteErr error
	getResult *comment.Comment
	getErr    error
}

func (m *mockRepo) CreateComment(_ context.Context, c *comment.Comment) error {
	return m.createErr
}

func (m *mockRepo) UpdateComment(_ context.Context, _ *comment.Comment) error {
	return m.updateErr
}

func (m *mockRepo) DeleteComment(_ context.Context, _ string, _ int) error {
	return m.deleteErr
}

func (m *mockRepo) GetCommentByID(_ context.Context, _ int) (*comment.Comment, error) {
	return m.getResult, m.getErr
}

func (m *mockRepo) GetCommentByIDWithVotes(_ context.Context, _ int, _ *string) (*comment.Comment, error) {
	return m.getResult, m.getErr
}

func (m *mockRepo) GetCommentsByTopicID(_ context.Context, _ int) ([]comment.Comment, error) {
	return nil, nil
}

func (m *mockRepo) GetCommentsByTopicIDWithVotes(_ context.Context, _ int, _ *string) ([]comment.Comment, error) {
	return nil, nil
}

func (m *mockRepo) CastCommentVote(_ context.Context, _ string, _ int, _ int) (comment.VoteChange, error) {
	return comment.VoteChangeAdded, nil
}

func (m *mockRepo) DeleteCommentVote(_ context.Context, _ string, _ int) error {
	return nil
}

func (m *mockRepo) GetVoteCounts(_ context.Context, _ int) (*comment.VoteCounts, error) {
	return &comment.VoteCounts{}, nil
}

func (m *mockRepo) GetCommentCount(_ context.Context, _ string) (int, error) { return 0, nil }

var (
	validPNG  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
	badHeader = []byte{0x00, 0x01, 0x02, 0x03}
)

type mockBus struct {
	routingKey string
}

func (m *mockBus) Publish(_ string, routingKey string, _ []byte) error {
	m.routingKey = routingKey
	return nil
}

func (m *mockBus) Subscribe(_ context.Context, _ string) (<-chan eventbus.Message, error) {
	ch := make(chan eventbus.Message)
	close(ch)
	return ch, nil
}

func (m *mockBus) InitTopology(_ context.Context) error {
	return nil
}

type mockStorage struct {
	uploaded bool
	gotPath  string
}

func (m *mockStorage) Upload(_ context.Context, _ []byte, path string) error {
	m.uploaded = true
	m.gotPath = path
	return nil
}

func TestCreateComment_Success(t *testing.T) {
	repo := &mockRepo{}
	h := NewCreateCommentHandler(repo, &mockBus{}, &mockUserRepo{}, &mockTopicRepo{}, &mockStorage{})

	c, err := h.Execute(context.Background(), CreateCommentCommand{
		UserID:  "u1",
		TopicID: 1,
		Content: "hello",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if c.UserID != "u1" {
		t.Errorf("UserID = %q, want %q", c.UserID, "u1")
	}
	if c.Content != "hello" {
		t.Errorf("Content = %q, want %q", c.Content, "hello")
	}
}

func TestCreateComment_WithValidImage(t *testing.T) {
	repo := &mockRepo{}
	store := &mockStorage{}
	h := NewCreateCommentHandler(repo, &mockBus{}, &mockUserRepo{}, &mockTopicRepo{}, store)

	c, err := h.Execute(context.Background(), CreateCommentCommand{
		UserID:        "u1",
		TopicID:       1,
		Content:       "pic",
		ImageData:     validPNG,
		ImageFileName: "pic.png",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if c == nil {
		t.Fatal("expected comment, got nil")
	}
	if !store.uploaded {
		t.Error("expected image to be uploaded")
	}
	if c.ImagePath == "" {
		t.Error("expected ImagePath to be set")
	}
	if store.gotPath != "pic.png" {
		t.Errorf("upload path = %q, want bare filename pic.png", store.gotPath)
	}
}

func TestCreateComment_InvalidImage(t *testing.T) {
	repo := &mockRepo{}
	h := NewCreateCommentHandler(repo, &mockBus{}, &mockUserRepo{}, &mockTopicRepo{}, &mockStorage{})

	_, err := h.Execute(context.Background(), CreateCommentCommand{
		UserID:        "u1",
		TopicID:       1,
		Content:       "bad pic",
		ImageData:     badHeader,
		ImageFileName: "bad.png",
	})
	if err == nil {
		t.Fatal("expected error for bad image, got nil")
	}
}

func TestCreateComment_EmptyUserID(t *testing.T) {
	h := NewCreateCommentHandler(&mockRepo{}, &mockBus{}, &mockUserRepo{}, &mockTopicRepo{}, &mockStorage{})
	_, err := h.Execute(context.Background(), CreateCommentCommand{TopicID: 1, Content: "x"})
	if !errors.Is(err, ErrEmptyUserID) {
		t.Errorf("error = %v, want ErrEmptyUserID", err)
	}
}

func TestCreateComment_ZeroTopicID(t *testing.T) {
	h := NewCreateCommentHandler(&mockRepo{}, &mockBus{}, &mockUserRepo{}, &mockTopicRepo{}, &mockStorage{})
	_, err := h.Execute(context.Background(), CreateCommentCommand{UserID: "u1", Content: "x"})
	if !errors.Is(err, ErrEmptyTopicID) {
		t.Errorf("error = %v, want ErrEmptyTopicID", err)
	}
}

func TestCreateComment_EmptyContent(t *testing.T) {
	h := NewCreateCommentHandler(&mockRepo{}, &mockBus{}, &mockUserRepo{}, &mockTopicRepo{}, &mockStorage{})
	_, err := h.Execute(context.Background(), CreateCommentCommand{UserID: "u1", TopicID: 1})
	if !errors.Is(err, ErrEmptyContent) {
		t.Errorf("error = %v, want ErrEmptyContent", err)
	}
}

func TestCreateComment_RepoError(t *testing.T) {
	repo := &mockRepo{createErr: errors.New("db fail")}
	h := NewCreateCommentHandler(repo, &mockBus{}, &mockUserRepo{}, &mockTopicRepo{}, &mockStorage{})

	_, err := h.Execute(context.Background(), CreateCommentCommand{
		UserID:  "u1",
		TopicID: 1,
		Content: "x",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateComment_NoImageAllowed(t *testing.T) {
	repo := &mockRepo{}
	h := NewCreateCommentHandler(repo, &mockBus{}, &mockUserRepo{}, &mockTopicRepo{}, &mockStorage{})

	c, err := h.Execute(context.Background(), CreateCommentCommand{
		UserID:  "u1",
		TopicID: 1,
		Content: "no image",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if c.ImagePath != "" {
		t.Errorf("ImagePath = %q, want empty", c.ImagePath)
	}
}
