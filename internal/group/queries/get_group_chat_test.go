package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
)

type groupChatQueryStub struct {
	group.Repository

	isMember  bool
	memberErr error
	messages  []group.ChatMessage
}

func (s *groupChatQueryStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	return s.isMember, s.memberErr
}

func (s *groupChatQueryStub) GetGroupChatMessages(_ context.Context, _ string, _ int) ([]group.ChatMessage, error) {
	return s.messages, nil
}

func TestGetGroupChatResolver_ReturnsMessages(t *testing.T) {
	ctx := context.Background()
	repo := &groupChatQueryStub{
		isMember: true,
		messages: []group.ChatMessage{{ID: "m1", GroupID: "g1", SenderID: "u1", Content: "hi"}},
	}
	r := NewGetGroupChatResolver(repo)

	result, err := r.Resolve(ctx, GetGroupChatQuery{GroupID: "g1", UserID: "u2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) != 1 || result.Messages[0].Content != "hi" {
		t.Fatalf("unexpected messages: %#v", result.Messages)
	}
}

func TestGetGroupChatResolver_RejectsNonMember(t *testing.T) {
	ctx := context.Background()
	repo := &groupChatQueryStub{isMember: false}
	r := NewGetGroupChatResolver(repo)

	_, err := r.Resolve(ctx, GetGroupChatQuery{GroupID: "g1", UserID: "u2"})
	if !errors.Is(err, group.ErrNotMember) {
		t.Fatalf("expected ErrNotMember, got %v", err)
	}
}

func TestGetGroupChatResolver_DefaultLimit(t *testing.T) {
	ctx := context.Background()
	var gotLimit int
	r := NewGetGroupChatResolver(&limitCapturingStub{capture: func(l int) { gotLimit = l }})

	_, err := r.Resolve(ctx, GetGroupChatQuery{GroupID: "g1", UserID: "u1", Limit: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotLimit != 20 {
		t.Fatalf("limit = %d, want default 20", gotLimit)
	}
}

type limitCapturingStub struct {
	group.Repository

	capture func(limit int)
}

func (s *limitCapturingStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}

func (s *limitCapturingStub) GetGroupChatMessages(_ context.Context, _ string, limit int) ([]group.ChatMessage, error) {
	s.capture(limit)
	return nil, nil
}
