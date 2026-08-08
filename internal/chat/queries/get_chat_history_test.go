package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/chat"
)

type followCheckerStub struct {
	connected bool
	err       error
}

func (f *followCheckerStub) AreConnected(_ context.Context, _, _ string) (bool, error) {
	return f.connected, f.err
}

func TestGetChatHistoryResolver_ParticipantCheck(t *testing.T) {
	ctx := context.Background()
	repo := &chatRepoStub{getChatFn: func(_ context.Context, _ string) (*chat.Chat, error) {
		return &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"}, nil
	}}
	follow := &followCheckerStub{connected: true}

	r := NewGetChatHistoryResolver(repo, follow)

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

func TestGetChatHistoryResolver_NotConnected(t *testing.T) {
	ctx := context.Background()
	repo := &chatRepoStub{getChatFn: func(_ context.Context, _ string) (*chat.Chat, error) {
		return &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"}, nil
	}}
	follow := &followCheckerStub{connected: false}

	r := NewGetChatHistoryResolver(repo, follow)

	_, err := r.Resolve(ctx, GetChatHistoryQuery{ChatID: "c1", RequesterID: "u1"})
	if !errors.Is(err, chat.ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected, got %v", err)
	}
}

func TestGetChatHistoryResolver_ConnectedEitherDirection(t *testing.T) {
	ctx := context.Background()
	repo := &chatRepoStub{getChatFn: func(_ context.Context, _ string) (*chat.Chat, error) {
		return &chat.Chat{ID: "c1", UserOneID: "u1", UserTwoID: "u2"}, nil
	}}
	// Either direction counts as connected: u2 follows u1 even though u1
	// does not follow u2.
	follow := &followCheckerStub{connected: true}

	r := NewGetChatHistoryResolver(repo, follow)

	msgs, err := r.Resolve(ctx, GetChatHistoryQuery{ChatID: "c1", RequesterID: "u1"})
	if err != nil {
		t.Fatalf("unexpected error for connected users: %v", err)
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
	follow := &followCheckerStub{connected: false}

	r := NewGetChatHistoryResolver(repo, follow)
	msgs, err := r.Resolve(ctx, GetChatHistoryQuery{ChatID: "c1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgs == nil {
		t.Fatal("expected messages slice")
	}
}
