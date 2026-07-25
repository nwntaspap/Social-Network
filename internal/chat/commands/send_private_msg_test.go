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
	connected bool
	err       error
}

func (m *mockFollowChecker) AreConnected(_ context.Context, _, _ string) (bool, error) {
	return m.connected, m.err
}

func TestSendPrivateMessage_NotConnected(t *testing.T) {
	h := NewSendPrivateMessageHandler(&mockChatRepo{}, &mockFollowChecker{connected: false})
	_, err := h.Execute(context.Background(), SendPrivateMessageCommand{
		SenderID:   "u1",
		ReceiverID: "u2",
		Content:    "hello",
	})
	if !errors.Is(err, ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected, got %v", err)
	}
}

func TestSendPrivateMessage_Success(t *testing.T) {
	h := NewSendPrivateMessageHandler(&mockChatRepo{}, &mockFollowChecker{connected: true})
	res, err := h.Execute(context.Background(), SendPrivateMessageCommand{
		SenderID:   "u1",
		ReceiverID: "u2",
		Content:    "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Message == nil {
		t.Fatal("expected message")
	}
	if res.Message.Content != "hello" {
		t.Fatalf("expected content 'hello', got %q", res.Message.Content)
	}
}

func TestSendPrivateMessage_FollowCheckError(t *testing.T) {
	sentinel := errors.New("db down")
	h := NewSendPrivateMessageHandler(&mockChatRepo{}, &mockFollowChecker{err: sentinel})
	_, err := h.Execute(context.Background(), SendPrivateMessageCommand{
		SenderID:   "u1",
		ReceiverID: "u2",
		Content:    "hello",
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
