package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/chat"
)

func TestGetChatHistoryResolver_ParticipantCheck(t *testing.T) {
	ctx := context.Background()
	repo := &chatRepoStub{getChatFn: func(_ context.Context, _ string) (*chat.Chat, error) {
		return &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"}, nil
	}}

	r := NewGetChatHistoryResolver(repo)

	_, err := r.Resolve(ctx, GetChatHistoryQuery{ChatID: "c1", RequesterID: "u3"})
	if !errors.Is(err, chat.ErrNotParticipant) {
		t.Fatalf("expected ErrNotParticipant, got %v", err)
	}

	msgs, err := r.Resolve(ctx, GetChatHistoryQuery{ChatID: "c1", RequesterID: "u2"})
	if err != nil {
		t.Fatalf("unexpected error for participant: %v", err)
	}
	if msgs == nil {
		t.Fatal("expected messages slice")
	}
}

func TestGetChatHistoryResolver_EmptyRequesterSkipsCheck(t *testing.T) {
	ctx := context.Background()
	repo := &chatRepoStub{getChatFn: func(_ context.Context, _ string) (*chat.Chat, error) {
		return &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"}, nil
	}}

	r := NewGetChatHistoryResolver(repo)
	msgs, err := r.Resolve(ctx, GetChatHistoryQuery{ChatID: "c1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgs == nil {
		t.Fatal("expected messages slice")
	}
}
