package store

import (
	"context"
	"testing"

	"social-network/internal/chat"
	"social-network/internal/platform/database"
)

// testSchema mirrors the chats and messages tables from
// db/migrations/000001_initial_schema.up.sql plus the unique pair index added
// in db/migrations/000012_chat_pair_unique.up.sql.
const testSchema = `
CREATE TABLE chats (
    id TEXT PRIMARY KEY,
    user_one_id TEXT NOT NULL,
    user_two_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    last_message_id INTEGER,
    last_message_at TIMESTAMP
);

CREATE UNIQUE INDEX idx_chats_pair ON chats (user_one_id, user_two_id);

CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    client_message_id TEXT
);`

func setupStore(t *testing.T) *SQLiteStore {
	t.Helper()

	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(), testSchema); err != nil {
		t.Fatalf("create tables: %v", err)
	}

	return NewSQLiteStore(db)
}

func seedChat(t *testing.T, s *SQLiteStore, b string) *chat.Chat {
	t.Helper()
	c, err := s.GetOrCreateChat(context.Background(), "user-a", b)
	if err != nil {
		t.Fatalf("GetOrCreateChat(user-a, %s) error = %v", b, err)
	}
	return c
}

func TestGetOrCreateChat_ReusesExistingChat(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	first, err := s.GetOrCreateChat(ctx, "user-b", "user-a")
	if err != nil {
		t.Fatalf("GetOrCreateChat(user-b, user-a) error = %v", err)
	}

	second, err := s.GetOrCreateChat(ctx, "user-a", "user-b")
	if err != nil {
		t.Fatalf("GetOrCreateChat(user-a, user-b) error = %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("GetOrCreateChat returned different chats: %q vs %q, want the same chat for the same pair", first.ID, second.ID)
	}
	if first.UserOneID != "user-a" || first.UserTwoID != "user-b" {
		t.Fatalf("chat pair not normalized: got (%q, %q), want (user-a, user-b)", first.UserOneID, first.UserTwoID)
	}

	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM chats").Scan(&count); err != nil {
		t.Fatalf("count chats: %v", err)
	}
	if count != 1 {
		t.Fatalf("chats rows = %d, want 1 (INSERT OR IGNORE must reuse the existing chat)", count)
	}
}

func TestGetOrCreateChat_AllowsDistinctPairs(t *testing.T) {
	s := setupStore(t)

	a := seedChat(t, s, "user-b")
	b := seedChat(t, s, "user-c")

	if a.ID == b.ID {
		t.Fatalf("different pairs must not share a chat, got %q", a.ID)
	}
}

func TestGetMessagesForChat_ReturnsOldestFirst(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	c := seedChat(t, s, "user-b")
	insertMessage(t, s, 1, c.ID, "oldest", "2024-01-01 00:00:00")
	insertMessage(t, s, 3, c.ID, "middle", "2024-01-01 00:00:01")
	insertMessage(t, s, 2, c.ID, "newest", "2024-01-01 00:00:02")

	messages, err := s.GetMessagesForChat(ctx, c.ID, 10)
	if err != nil {
		t.Fatalf("GetMessagesForChat() error = %v", err)
	}

	want := []int{1, 3, 2}
	if len(messages) != len(want) {
		t.Fatalf("messages = %d, want %d", len(messages), len(want))
	}
	for i, id := range want {
		if messages[i].ID != id {
			t.Fatalf("messages[%d].ID = %d, want %d (oldest first)", i, messages[i].ID, id)
		}
	}
}

func TestGetMessagesForChat_RespectsLimit(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	c := seedChat(t, s, "user-b")
	insertMessage(t, s, 1, c.ID, "oldest", "2024-01-01 00:00:00")
	insertMessage(t, s, 3, c.ID, "middle", "2024-01-01 00:00:01")
	insertMessage(t, s, 2, c.ID, "newest", "2024-01-01 00:00:02")

	messages, err := s.GetMessagesForChat(ctx, c.ID, 2)
	if err != nil {
		t.Fatalf("GetMessagesForChat() error = %v", err)
	}

	want := []int{3, 2}
	if len(messages) != len(want) {
		t.Fatalf("messages = %d, want %d", len(messages), len(want))
	}
	for i, id := range want {
		if messages[i].ID != id {
			t.Fatalf("messages[%d].ID = %d, want %d (oldest first)", i, messages[i].ID, id)
		}
	}
}

func TestGetMessagesForChatBefore_ReturnsOldestFirst(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	c := seedChat(t, s, "user-b")
	insertMessage(t, s, 1, c.ID, "oldest", "2024-01-01 00:00:00")
	insertMessage(t, s, 2, c.ID, "middle", "2024-01-01 00:00:01")
	insertMessage(t, s, 3, c.ID, "newest", "2024-01-01 00:00:02")
	insertMessage(t, s, 4, c.ID, "latest", "2024-01-01 00:00:03")

	messages, err := s.GetMessagesForChatBefore(ctx, c.ID, 4, 10)
	if err != nil {
		t.Fatalf("GetMessagesForChatBefore() error = %v", err)
	}

	want := []int{1, 2, 3}
	if len(messages) != len(want) {
		t.Fatalf("messages = %d, want %d", len(messages), len(want))
	}
	for i, id := range want {
		if messages[i].ID != id {
			t.Fatalf("messages[%d].ID = %d, want %d (oldest first)", i, messages[i].ID, id)
		}
	}
}

func insertMessage(t *testing.T, s *SQLiteStore, id int, chatID, content, createdAt string) {
	t.Helper()
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO messages (id, chat_id, sender_id, content, created_at)
		VALUES (?, ?, 'user-a', ?, ?)
	`, id, chatID, content, createdAt)
	if err != nil {
		t.Fatalf("insert message %d: %v", id, err)
	}
}
