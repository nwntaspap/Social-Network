package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/chat"
)

type mockChatRepo struct {
	getOrCreateChatFn func(ctx context.Context, a, b string) (*chat.Chat, error)
	sendMessageFn     func(ctx context.Context, chatID, senderID, content, clientMsgID string) (*chat.Message, error)
}

func (m *mockChatRepo) GetOrCreateChat(ctx context.Context, a, b string) (*chat.Chat, error) {
	if m.getOrCreateChatFn != nil {
		return m.getOrCreateChatFn(ctx, a, b)
	}
	return &chat.Chat{ID: "chat-1", UserOneID: a, UserTwoID: b, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *mockChatRepo) GetChat(_ context.Context, _ string) (*chat.Chat, error) {
	return nil, errors.New("not implemented")
}

func (m *mockChatRepo) GetChatsForUser(_ context.Context, _ string) ([]*chat.Chat, error) {
	return nil, errors.New("not implemented")
}

func (m *mockChatRepo) SendMessage(ctx context.Context, chatID, senderID, content, clientMsgID string) (*chat.Message, error) {
	if m.sendMessageFn != nil {
		return m.sendMessageFn(ctx, chatID, senderID, content, clientMsgID)
	}
	return &chat.Message{ID: 1, ChatID: chatID, SenderID: senderID, Content: content, CreatedAt: time.Now()}, nil
}

func (m *mockChatRepo) GetMessagesForChat(_ context.Context, _ string, _ int) ([]*chat.Message, error) {
	return nil, errors.New("not implemented")
}

func (m *mockChatRepo) GetMessagesForChatBefore(_ context.Context, _ string, _ int, _ int) ([]*chat.Message, error) {
	return nil, errors.New("not implemented")
}
func (m *mockChatRepo) MarkAsRead(_ context.Context, _, _ string, _ int) error { return nil }
func (m *mockChatRepo) GetUnreadCount(_ context.Context, _, _ string) (int, error) {
	return 0, nil
}

func (m *mockChatRepo) GetAllUnreadCounts(_ context.Context, _ string) (map[string]int, error) {
	return nil, errors.New("not implemented")
}

type mockFollowChecker struct {
	connectedFn func(a, b string) bool
	err         error
}

func (m *mockFollowChecker) AreConnected(_ context.Context, a, b string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.connectedFn != nil {
		return m.connectedFn(a, b), nil
	}
	return true, nil
}

func newTestHandler(follow chat.FollowChecker) *SendPrivateMessageHandler {
	return NewSendPrivateMessageHandler(&mockChatRepo{}, NewMessageGate(follow))
}

func TestSendPrivateMessage_NotConnected(t *testing.T) {
	h := newTestHandler(&mockFollowChecker{connectedFn: func(_, _ string) bool { return false }})
	_, err := h.Execute(context.Background(), SendPrivateMessageCommand{
		SenderID: "u1", ReceiverID: "u2", Content: "hello",
	})
	if !errors.Is(err, ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected, got %v", err)
	}
}

func TestSendPrivateMessage_ApprovedFollowerCanMessagePrivateRecipient(t *testing.T) {
	// sender follows a private recipient; the established follow enables messaging
	h := newTestHandler(
		&mockFollowChecker{connectedFn: func(a, b string) bool { return a == "u1" && b == "u2" }},
	)
	res, err := h.Execute(context.Background(), SendPrivateMessageCommand{
		SenderID: "u1", ReceiverID: "u2", Content: "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Message == nil || res.Message.Content != "hello" {
		t.Fatalf("unexpected message: %#v", res.Message)
	}
}

func TestSendPrivateMessage_PrivateSenderMessagesPublicFollower(t *testing.T) {
	// private sender messages a public recipient who follows back
	h := newTestHandler(
		&mockFollowChecker{connectedFn: func(a, b string) bool { return a == "u2" && b == "u1" }},
	)
	res, err := h.Execute(context.Background(), SendPrivateMessageCommand{
		SenderID: "u1", ReceiverID: "u2", Content: "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Message == nil || res.Message.Content != "hello" {
		t.Fatalf("unexpected message: %#v", res.Message)
	}
}

func TestSendPrivateMessage_SenderFollows_PublicRecipient(t *testing.T) {
	// sender follows receiver, receiver does not follow back
	h := newTestHandler(
		&mockFollowChecker{connectedFn: func(a, b string) bool { return a == "u1" && b == "u2" }},
	)
	res, err := h.Execute(context.Background(), SendPrivateMessageCommand{
		SenderID: "u1", ReceiverID: "u2", Content: "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Message == nil || res.Message.Content != "hello" {
		t.Fatalf("unexpected message: %#v", res.Message)
	}
}

func TestSendPrivateMessage_RecipientFollowsBack(t *testing.T) {
	// receiver follows sender, so messaging is enabled
	h := newTestHandler(
		&mockFollowChecker{connectedFn: func(a, b string) bool { return a == "u2" && b == "u1" }},
	)
	res, err := h.Execute(context.Background(), SendPrivateMessageCommand{
		SenderID: "u1", ReceiverID: "u2", Content: "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Message == nil {
		t.Fatal("expected message")
	}
	if res.RecipientID != "u2" {
		t.Fatalf("recipient = %q, want u2", res.RecipientID)
	}
}

func TestSendPrivateMessage_FollowCheckError(t *testing.T) {
	sentinel := errors.New("db down")
	h := newTestHandler(&mockFollowChecker{err: sentinel})
	_, err := h.Execute(context.Background(), SendPrivateMessageCommand{
		SenderID: "u1", ReceiverID: "u2", Content: "hello",
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestOpenPrivateChat_Gate(t *testing.T) {
	h := NewOpenPrivateChatHandler(&mockChatRepo{}, NewMessageGate(
		&mockFollowChecker{connectedFn: func(a, b string) bool { return a == "u1" && b == "u2" }},
	))
	res, err := h.Execute(context.Background(), OpenPrivateChatCommand{SenderID: "u1", ReceiverID: "u2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Chat == nil || res.Chat.ID != "chat-1" {
		t.Fatalf("unexpected chat: %#v", res.Chat)
	}
}

func TestOpenPrivateChat_NotConnected(t *testing.T) {
	h := NewOpenPrivateChatHandler(&mockChatRepo{}, NewMessageGate(
		&mockFollowChecker{connectedFn: func(_, _ string) bool { return false }},
	))
	_, err := h.Execute(context.Background(), OpenPrivateChatCommand{SenderID: "u1", ReceiverID: "u2"})
	if !errors.Is(err, ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected, got %v", err)
	}
}
