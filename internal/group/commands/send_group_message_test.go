package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
)

type groupChatRepoStub struct {
	group.Repository

	isMember  bool
	memberErr error
	sendErr   error
	stored    *group.ChatMessage
}

func (s *groupChatRepoStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	return s.isMember, s.memberErr
}

func (s *groupChatRepoStub) SendGroupChatMessage(_ context.Context, msg *group.ChatMessage) error {
	if s.sendErr != nil {
		return s.sendErr
	}
	s.stored = msg
	return nil
}

func TestSendGroupMessageHandler_StoresMessage(t *testing.T) {
	ctx := context.Background()
	repo := &groupChatRepoStub{isMember: true}
	h := NewSendGroupMessageHandler(repo)

	result, err := h.Execute(ctx, SendGroupMessageCommand{GroupID: "g1", SenderID: "u1", Content: "  hello  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Message == nil || result.Message.Content != "hello" {
		t.Fatalf("unexpected message: %#v", result.Message)
	}
	if repo.stored == nil || repo.stored.GroupID != "g1" || repo.stored.SenderID != "u1" {
		t.Fatalf("message not stored correctly: %#v", repo.stored)
	}
	if result.Message.ID == "" {
		t.Error("expected generated message ID")
	}
}

func TestSendGroupMessageHandler_RejectsNonMember(t *testing.T) {
	ctx := context.Background()
	repo := &groupChatRepoStub{isMember: false}
	h := NewSendGroupMessageHandler(repo)

	_, err := h.Execute(ctx, SendGroupMessageCommand{GroupID: "g1", SenderID: "u1", Content: "hello"})
	if !errors.Is(err, group.ErrNotMember) {
		t.Fatalf("expected ErrNotMember, got %v", err)
	}
}

func TestSendGroupMessageHandler_Validation(t *testing.T) {
	ctx := context.Background()
	repo := &groupChatRepoStub{isMember: true}
	h := NewSendGroupMessageHandler(repo)

	tests := []SendGroupMessageCommand{
		{GroupID: "", SenderID: "u1", Content: "hello"},
		{GroupID: "g1", SenderID: "", Content: "hello"},
		{GroupID: "g1", SenderID: "u1", Content: "   "},
	}
	for i, cmd := range tests {
		if _, err := h.Execute(ctx, cmd); err == nil {
			t.Fatalf("case %d: expected error, got nil", i)
		}
	}
}
