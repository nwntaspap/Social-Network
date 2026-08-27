package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
	"social-network/internal/topic"
)

func TestGetCommentsByTopicWithVotesResolver_Success(t *testing.T) {
	expected := []comment.Comment{
		{ID: 1, TopicID: 1, UserID: "u1", Content: "a", UpvoteCount: 3},
		{ID: 2, TopicID: 1, UserID: "u2", Content: "b", UpvoteCount: 1},
	}
	repo := &mockRepo{getResultWV: expected}
	r := NewGetCommentsByTopicWithVotesResolver(repo, &fakeTopicChecker{})

	result, err := r.Resolve(context.Background(), GetCommentsByTopicWithVotesQuery{TopicID: 1, RequesterID: "u2"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].UpvoteCount != 3 {
		t.Errorf("UpvoteCount = %d, want 3", result[0].UpvoteCount)
	}
}

func TestGetCommentsByTopicWithVotesResolver_Empty(t *testing.T) {
	repo := &mockRepo{getResultWV: []comment.Comment{}}
	r := NewGetCommentsByTopicWithVotesResolver(repo, &fakeTopicChecker{})

	result, err := r.Resolve(context.Background(), GetCommentsByTopicWithVotesQuery{TopicID: 1, RequesterID: "u2"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("len(result) = %d, want 0", len(result))
	}
}

func TestGetCommentsByTopicWithVotesResolver_Error(t *testing.T) {
	repo := &mockRepo{getErrWV: errors.New("db down")}
	r := NewGetCommentsByTopicWithVotesResolver(repo, &fakeTopicChecker{})

	_, err := r.Resolve(context.Background(), GetCommentsByTopicWithVotesQuery{TopicID: 1, RequesterID: "u2"})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestGetCommentsByTopicWithVotesResolver_HiddenTopicDenied(t *testing.T) {
	repo := &mockRepo{getResultWV: []comment.Comment{{ID: 1, TopicID: 1}}}
	r := NewGetCommentsByTopicWithVotesResolver(repo, &fakeTopicChecker{err: topic.ErrTopicNotFound})

	_, err := r.Resolve(context.Background(), GetCommentsByTopicWithVotesQuery{TopicID: 1, RequesterID: "u2"})
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("Resolve() error = %v, want ErrTopicNotFound", err)
	}
}
