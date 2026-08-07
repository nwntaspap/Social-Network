package store

import (
	"context"
	"testing"

	"social-network/internal/group"
	"social-network/internal/platform/database"
)

const groupChatSchema = `
CREATE TABLE groups (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    creator_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE group_members (
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('creator', 'admin', 'member')),
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(group_id, user_id)
);
CREATE TABLE group_chat_messages (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

func setupGroupChatStore(t *testing.T) *SQLiteStore {
	t.Helper()

	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(), groupChatSchema); err != nil {
		t.Fatalf("create tables: %v", err)
	}

	return NewSQLiteStore(db)
}

func TestSendGroupChatMessage_StoresAndOrdersByCreation(t *testing.T) {
	s := setupGroupChatStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u1"})
	if err := s.AddMember(ctx, "g1", "u1", group.RoleCreator); err != nil {
		t.Fatalf("add member: %v", err)
	}

	if err := s.SendGroupChatMessage(ctx, &group.ChatMessage{ID: "m1", GroupID: "g1", SenderID: "u1", Content: "first"}); err != nil {
		t.Fatalf("send message: %v", err)
	}
	if err := s.SendGroupChatMessage(ctx, &group.ChatMessage{ID: "m2", GroupID: "g1", SenderID: "u1", Content: "second"}); err != nil {
		t.Fatalf("send message: %v", err)
	}

	messages, err := s.GetGroupChatMessages(ctx, "g1", 10)
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("got %d messages, want 2", len(messages))
	}
	if messages[0].Content != "first" || messages[1].Content != "second" {
		t.Fatalf("messages out of order: %#v", messages)
	}
}

func TestGetGroupChatMessages_LimitsResults(t *testing.T) {
	s := setupGroupChatStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u1"})
	for i := range 5 {
		msg := &group.ChatMessage{ID: string(rune('a' + i)), GroupID: "g1", SenderID: "u1", Content: "msg"}
		if err := s.SendGroupChatMessage(ctx, msg); err != nil {
			t.Fatalf("send message: %v", err)
		}
	}

	messages, err := s.GetGroupChatMessages(ctx, "g1", 2)
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("got %d messages, want 2", len(messages))
	}
}

func TestListGroupMemberIDs_ReturnsAllMembers(t *testing.T) {
	s := setupGroupChatStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u1"})
	for _, m := range []struct{ uid, role string }{
		{"u1", "creator"},
		{"u2", "member"},
		{"u3", "member"},
	} {
		if err := s.AddMember(ctx, "g1", m.uid, group.Role(m.role)); err != nil {
			t.Fatalf("add member %s: %v", m.uid, err)
		}
	}

	ids, err := s.ListGroupMemberIDs(ctx, "g1")
	if err != nil {
		t.Fatalf("list member ids: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("got %d ids, want 3: %v", len(ids), ids)
	}
}
