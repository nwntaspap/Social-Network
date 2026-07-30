package store

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recipient_id TEXT NOT NULL,
    type TEXT NOT NULL,
    resource_type TEXT NOT NULL DEFAULT '',
    resource_id INTEGER NOT NULL DEFAULT 0,
    actor_id TEXT NOT NULL,
    actor_name TEXT NOT NULL DEFAULT '',
    actor_avatar TEXT NOT NULL DEFAULT '',
    content_text TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    is_read BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_notifications_recipient ON notifications(recipient_id, created_at DESC);
`

func setupStore(t *testing.T) *SQLiteStore {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.ExecContext(context.Background(), schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return NewSQLiteStore(db)
}

func TestCreate(t *testing.T) {
	s := setupStore(t)

	n := &Notification{
		RecipientID:  "u1",
		Type:         "like",
		ResourceType: "post",
		ResourceID:   "42",
		ActorID:      "u2",
		ActorName:    "Bob",
		ActorAvatar:  "/avatars/bob.png",
		ContentText:  "Bob liked your post",
	}

	if err := s.Create(context.Background(), n); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if n.ID == 0 {
		t.Fatal("ID not set after create")
	}
}

func TestGetByRecipient(t *testing.T) {
	s := setupStore(t)

	for i := range 3 {
		_ = s.Create(context.Background(), &Notification{
			RecipientID:  "u1",
			Type:         "like",
			ResourceType: "post",
			ResourceID:   strconv.Itoa(100 + i),
			ActorID:      "u2",
			ContentText:  "notification " + itoa(i),
		})
	}

	ns, total, err := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if err != nil {
		t.Fatalf("GetByRecipient: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(ns) != 3 {
		t.Fatalf("got %d notifications, want 3", len(ns))
	}
}

func TestGetByRecipient_Pagination(t *testing.T) {
	s := setupStore(t)

	for i := range 5 {
		_ = s.Create(context.Background(), &Notification{
			RecipientID:  "u1",
			Type:         "like",
			ResourceType: "post",
			ResourceID:   strconv.Itoa(i),
			ActorID:      "u2",
		})
	}

	ns, total, err := s.GetByRecipient(context.Background(), "u1", 2, 0)
	if err != nil {
		t.Fatalf("GetByRecipient: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(ns) != 2 {
		t.Errorf("got %d, want 2", len(ns))
	}
}

func TestGetByRecipient_HardDeleted(t *testing.T) {
	s := setupStore(t)

	n := &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"}
	_ = s.Create(context.Background(), n)
	_ = s.DeleteByResource(context.Background(), "u2", "post", "1")

	ns, total, err := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if err != nil {
		t.Fatalf("GetByRecipient: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(ns) != 0 {
		t.Errorf("got %d notifications, want 0", len(ns))
	}
}

func TestGetUnreadCount(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "2", ActorID: "u2"})

	count, err := s.GetUnreadCount(context.Background(), "u1")
	if err != nil {
		t.Fatalf("GetUnreadCount: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestMarkRead(t *testing.T) {
	s := setupStore(t)

	n := &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"}
	_ = s.Create(context.Background(), n)

	if err := s.MarkRead(context.Background(), n.ID, "u1"); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}

	count, _ := s.GetUnreadCount(context.Background(), "u1")
	if count != 0 {
		t.Errorf("unread count = %d, want 0", count)
	}
}

func TestMarkRead_WrongUser(t *testing.T) {
	s := setupStore(t)

	n := &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"}
	_ = s.Create(context.Background(), n)

	err := s.MarkRead(context.Background(), n.ID, "u2")
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestMarkAllRead(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "2", ActorID: "u3"})

	if err := s.MarkAllRead(context.Background(), "u1"); err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}

	count, _ := s.GetUnreadCount(context.Background(), "u1")
	if count != 0 {
		t.Errorf("unread count = %d, want 0", count)
	}
}

func TestDeleteByResource(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"})

	if err := s.DeleteByResource(context.Background(), "u2", "post", "1"); err != nil {
		t.Fatalf("DeleteByResource: %v", err)
	}

	ns, total, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(ns) != 0 {
		t.Errorf("got %d, want 0", len(ns))
	}
}

func TestDeleteByResource_NotFound(t *testing.T) {
	s := setupStore(t)
	err := s.DeleteByResource(context.Background(), "u1", "post", "999")
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestDeleteAllByResource(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u3"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u2", Type: "comment", ResourceType: "post", ResourceID: "1", ActorID: "u1"})

	if err := s.DeleteAllByResource(context.Background(), "1"); err != nil {
		t.Fatalf("DeleteAllByResource: %v", err)
	}

	ns1, total1, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if total1 != 0 {
		t.Errorf("u1 total = %d, want 0", total1)
	}
	_ = ns1

	ns2, total2, _ := s.GetByRecipient(context.Background(), "u2", 10, 0)
	if total2 != 0 {
		t.Errorf("u2 total = %d, want 0", total2)
	}
	_ = ns2
}

func TestDeleteAllByResource_NotFound(t *testing.T) {
	s := setupStore(t)
	err := s.DeleteAllByResource(context.Background(), "999")
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUpdateActorInfo(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{
		RecipientID:  "u1",
		Type:         "like",
		ResourceType: "post",
		ResourceID:   "1",
		ActorID:      "u2",
		ActorName:    "OldName",
		ActorAvatar:  "/old.png",
	})

	if err := s.UpdateActorInfo(context.Background(), "u2", "NewName", "/new.png"); err != nil {
		t.Fatalf("UpdateActorInfo: %v", err)
	}

	ns, _, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if len(ns) != 1 {
		t.Fatalf("got %d notifications, want 1", len(ns))
	}
	if ns[0].ActorName != "NewName" {
		t.Errorf("ActorName = %q, want %q", ns[0].ActorName, "NewName")
	}
	if ns[0].ActorAvatar != "/new.png" {
		t.Errorf("ActorAvatar = %q, want %q", ns[0].ActorAvatar, "/new.png")
	}
}

func itoa(i int) string {
	return string(rune('0' + i))
}
