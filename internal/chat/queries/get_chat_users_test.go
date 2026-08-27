package queries

import (
	"context"
	"testing"
	"time"

	"social-network/internal/chat"
)

type chatRepoStub struct {
	chatsFn   func(ctx context.Context, userID string) ([]*chat.Chat, error)
	getChatFn func(ctx context.Context, chatID string) (*chat.Chat, error)
}

func (s *chatRepoStub) GetOrCreateChat(_ context.Context, _, _ string) (*chat.Chat, error) {
	return &chat.Chat{}, nil
}

func (s *chatRepoStub) GetChat(ctx context.Context, chatID string) (*chat.Chat, error) {
	if s.getChatFn != nil {
		return s.getChatFn(ctx, chatID)
	}
	return &chat.Chat{}, nil
}

func (s *chatRepoStub) GetChatsForUser(ctx context.Context, userID string) ([]*chat.Chat, error) {
	if s.chatsFn != nil {
		return s.chatsFn(ctx, userID)
	}
	return nil, nil
}

func (s *chatRepoStub) SendMessage(_ context.Context, _, _, _, _ string) (*chat.Message, error) {
	return &chat.Message{}, nil
}

func (s *chatRepoStub) GetMessagesForChat(_ context.Context, _ string, _ int) ([]*chat.Message, error) {
	return []*chat.Message{}, nil
}

func (s *chatRepoStub) GetMessagesForChatBefore(_ context.Context, _ string, _, _ int) ([]*chat.Message, error) {
	return []*chat.Message{}, nil
}

func (s *chatRepoStub) MarkAsRead(_ context.Context, _, _ string, _ int) error {
	return nil
}

func (s *chatRepoStub) GetUnreadCount(_ context.Context, _, _ string) (int, error) {
	return 0, nil
}

func (s *chatRepoStub) GetAllUnreadCounts(_ context.Context, _ string) (map[string]int, error) {
	return map[string]int{}, nil
}

func TestGetChatUsersResolver_ReturnsConversations(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	later := now.Add(time.Hour)

	users := []*chat.UserRef{
		{ID: "me", Nickname: "Me"},
		{ID: "u1", Nickname: "Alice"},
		{ID: "u2", Nickname: "Bob"},
	}
	userRepo := &chat.UserAdapter{GetAllFn: func(_ context.Context) ([]*chat.UserRef, error) {
		return users, nil
	}}
	broadcaster := &chat.BroadcasterAdapter{IsOnlineFn: func(id string) bool { return id == "u1" }}

	chats := []*chat.Chat{
		{ID: "c1", UserOneID: "me", UserTwoID: "u1", UnreadCount: 2, LastMessageAt: &later, CreatedAt: now},
		{ID: "c2", UserOneID: "u2", UserTwoID: "me", UnreadCount: 0, LastMessageAt: &now, CreatedAt: now},
	}
	repo := &chatRepoStub{chatsFn: func(_ context.Context, _ string) ([]*chat.Chat, error) {
		return chats, nil
	}}

	r := NewGetChatUsersResolver(repo, userRepo, broadcaster)
	result, err := r.Resolve(ctx, GetChatUsersRequest{MeID: "me"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 conversations, got %d", len(result))
	}

	c0 := result[0]
	if c0.ID != "c1" {
		t.Errorf("conversation 0 id = %q, want %q", c0.ID, "c1")
	}
	if c0.OtherUser.ID != "u1" || c0.OtherUser.Nickname != "Alice" {
		t.Errorf("other user = %+v, want u1/Alice", c0.OtherUser)
	}
	if c0.UnreadCount != 2 {
		t.Errorf("unread count = %d, want 2", c0.UnreadCount)
	}
	if !c0.IsOnline {
		t.Error("expected u1 to be online")
	}
	if c0.LastMessageAt == nil || !c0.LastMessageAt.Equal(later) {
		t.Errorf("last message at = %v, want %v", c0.LastMessageAt, later)
	}

	c1 := result[1]
	if c1.OtherUser.ID != "u2" {
		t.Errorf("other user 1 = %+v, want u2", c1.OtherUser)
	}
	if c1.IsOnline {
		t.Error("expected u2 to be offline")
	}
}

func TestGetChatUsersResolver_DropsUsersWithoutChats(t *testing.T) {
	ctx := context.Background()

	users := []*chat.UserRef{
		{ID: "me", Nickname: "Me"},
		{ID: "u1", Nickname: "Alice"},
		{ID: "u2", Nickname: "Bob"},
	}
	userRepo := &chat.UserAdapter{GetAllFn: func(_ context.Context) ([]*chat.UserRef, error) {
		return users, nil
	}}
	broadcaster := &chat.BroadcasterAdapter{IsOnlineFn: func(_ string) bool { return false }}

	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	chats := []*chat.Chat{
		{ID: "c1", UserOneID: "me", UserTwoID: "u1", UnreadCount: 1, LastMessageAt: &now, CreatedAt: now},
	}
	repo := &chatRepoStub{chatsFn: func(_ context.Context, _ string) ([]*chat.Chat, error) {
		return chats, nil
	}}

	r := NewGetChatUsersResolver(repo, userRepo, broadcaster)
	result, err := r.Resolve(ctx, GetChatUsersRequest{MeID: "me"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 conversation (u2 has no chat), got %d", len(result))
	}
	if result[0].OtherUser.ID != "u1" {
		t.Errorf("other user = %+v, want u1", result[0].OtherUser)
	}
}
