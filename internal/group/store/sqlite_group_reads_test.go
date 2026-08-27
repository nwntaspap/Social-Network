package store

import (
	"context"
	"testing"

	"social-network/internal/group"
)

// seedGroupChatMessageAt inserts a group chat message with a fixed created_at
// so unread boundaries (which rely on created_at) are deterministic.
func seedGroupChatMessageAt(t *testing.T, s *SQLiteStore, msg *group.ChatMessage, createdAt string) {
	t.Helper()

	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO group_chat_messages (id, group_id, sender_id, content, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		msg.ID, msg.GroupID, msg.SenderID, msg.Content, createdAt)
	if err != nil {
		t.Fatalf("insert group chat message %s: %v", msg.ID, err)
	}
}

func TestMarkGroupRead_UpsertsReadMarker(t *testing.T) {
	s := setupGroupChatStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u1"})

	if err := s.MarkGroupRead(ctx, "g1", "u1"); err != nil {
		t.Fatalf("MarkGroupRead error = %v", err)
	}

	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM group_chat_reads WHERE group_id = ? AND user_id = ?`, "g1", "u1").Scan(&count)
	if err != nil {
		t.Fatalf("query read marker: %v", err)
	}
	if count != 1 {
		t.Fatalf("read markers = %d, want 1", count)
	}

	// Calling it again must not duplicate the row (upsert).
	if err = s.MarkGroupRead(ctx, "g1", "u1"); err != nil {
		t.Fatalf("second MarkGroupRead error = %v", err)
	}
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM group_chat_reads WHERE group_id = ? AND user_id = ?`, "g1", "u1").Scan(&count)
	if err != nil {
		t.Fatalf("query read marker after upsert: %v", err)
	}
	if count != 1 {
		t.Fatalf("read markers after upsert = %d, want 1", count)
	}
}

func TestCountGroupUnread_ExcludesOwnAndMarkedMessages(t *testing.T) {
	s := setupGroupChatStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u1"})

	// Own message + another member's message before the read marker.
	seedGroupChatMessageAt(t, s, &group.ChatMessage{ID: "m1", GroupID: "g1", SenderID: "u1", Content: "own"}, "2025-01-01 10:00:00")
	seedGroupChatMessageAt(t, s, &group.ChatMessage{ID: "m2", GroupID: "g1", SenderID: "u2", Content: "seen"}, "2025-01-02 10:00:00")
	// Another member's messages after the read marker.
	seedGroupChatMessageAt(t, s, &group.ChatMessage{ID: "m3", GroupID: "g1", SenderID: "u2", Content: "unread"}, "2025-01-03 10:00:00")
	seedGroupChatMessageAt(t, s, &group.ChatMessage{ID: "m4", GroupID: "g1", SenderID: "u3", Content: "unread2"}, "2025-01-04 10:00:00")

	// No marker yet: all messages from others are unread.
	unread, err := s.CountGroupUnread(ctx, []string{"g1"}, "u1")
	if err != nil {
		t.Fatalf("CountGroupUnread(no marker) error = %v", err)
	}
	if unread["g1"] != 3 {
		t.Fatalf("unread (no marker) = %d, want 3", unread["g1"])
	}

	// Mark read at 10:00:00 -> m1/m2 (<= 10:00:00) read, m3/m4 unread.
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO group_chat_reads (group_id, user_id, last_read_at) VALUES (?, ?, ?)`,
		"g1", "u1", "2025-01-02 10:00:00")
	if err != nil {
		t.Fatalf("seed read marker: %v", err)
	}

	unread, err = s.CountGroupUnread(ctx, []string{"g1"}, "u1")
	if err != nil {
		t.Fatalf("CountGroupUnread(with marker) error = %v", err)
	}
	if unread["g1"] != 2 {
		t.Fatalf("unread (with marker) = %d, want 2", unread["g1"])
	}
}

func TestCountGroupUnread_HandlesMultipleGroupsAndEmptyInput(t *testing.T) {
	s := setupGroupChatStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u1"})
	seedGroup(t, s, &group.Group{ID: "g2", Title: "Rust", CreatorID: "u2"})

	seedGroupChatMessageAt(t, s, &group.ChatMessage{ID: "m1", GroupID: "g1", SenderID: "u2", Content: "hi"}, "2025-01-01 10:00:00")
	seedGroupChatMessageAt(t, s, &group.ChatMessage{ID: "m2", GroupID: "g2", SenderID: "u3", Content: "yo"}, "2025-01-01 10:00:00")
	seedGroupChatMessageAt(t, s, &group.ChatMessage{ID: "m3", GroupID: "g2", SenderID: "u1", Content: "mine"}, "2025-01-01 10:00:00")

	unread, err := s.CountGroupUnread(ctx, []string{"g1", "g2"}, "u1")
	if err != nil {
		t.Fatalf("CountGroupUnread error = %v", err)
	}
	if unread["g1"] != 1 {
		t.Fatalf("unread g1 = %d, want 1", unread["g1"])
	}
	if unread["g2"] != 1 { // m2 from u3; own m3 excluded.
		t.Fatalf("unread g2 = %d, want 1", unread["g2"])
	}

	empty, err := s.CountGroupUnread(ctx, nil, "u1")
	if err != nil {
		t.Fatalf("CountGroupUnread(empty) error = %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("CountGroupUnread(empty) = %v, want empty", empty)
	}
}
