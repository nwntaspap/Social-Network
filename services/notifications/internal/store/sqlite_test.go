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
    group_id TEXT NOT NULL DEFAULT '',
    actor_id TEXT NOT NULL,
    actor_name TEXT NOT NULL DEFAULT '',
    actor_avatar TEXT NOT NULL DEFAULT '',
    content_text TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    join_request_id TEXT NOT NULL DEFAULT '',
    event_id TEXT NOT NULL DEFAULT '',
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

func TestCreate_PreservesGroupID(t *testing.T) {
	s := setupStore(t)

	n := &Notification{
		RecipientID:  "u1",
		Type:         "comment",
		ResourceType: "post",
		ResourceID:   "post-uuid",
		GroupID:      "g1",
		ActorID:      "u2",
	}

	if err := s.Create(context.Background(), n); err != nil {
		t.Fatalf("Create: %v", err)
	}

	ns, total, err := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if err != nil {
		t.Fatalf("GetByRecipient: %v", err)
	}
	if total != 1 || len(ns) != 1 {
		t.Fatalf("got total=%d len=%d, want 1", total, len(ns))
	}
	if ns[0].GroupID != "g1" {
		t.Errorf("GroupID = %q, want %q", ns[0].GroupID, "g1")
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
	_, _ = s.DeleteByResource(context.Background(), "like", "u2", "post", "1")

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

	deleted, err := s.DeleteByResource(context.Background(), "like", "u2", "post", "1")
	if err != nil {
		t.Fatalf("DeleteByResource: %v", err)
	}
	if len(deleted) != 1 {
		t.Fatalf("deleted %d notifications, want 1", len(deleted))
	}
	if deleted[0].RecipientID != "u1" || deleted[0].Type != "like" {
		t.Errorf("deleted notification = %+v, want recipient u1 type like", deleted[0])
	}

	ns, total, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(ns) != 0 {
		t.Errorf("got %d, want 0", len(ns))
	}
}

func TestDeleteByResource_OnlyMatchingType(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "comment", ResourceType: "post", ResourceID: "1", ActorID: "u2"})

	deleted, err := s.DeleteByResource(context.Background(), "like", "u2", "post", "1")
	if err != nil {
		t.Fatalf("DeleteByResource: %v", err)
	}
	if len(deleted) != 1 || deleted[0].Type != "like" {
		t.Errorf("deleted = %+v, want only the like", deleted)
	}

	ns, total, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 1 {
		t.Errorf("total = %d, want 1 (only the like deleted)", total)
	}
	if len(ns) != 1 || ns[0].Type != "comment" {
		t.Errorf("remaining notification = %+v, want comment type", ns)
	}
}

func TestDeleteByResource_NotFound(t *testing.T) {
	s := setupStore(t)
	_, err := s.DeleteByResource(context.Background(), "like", "u1", "post", "999")
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestDeleteAllByResource(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u3"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u2", Type: "comment", ResourceType: "post", ResourceID: "1", ActorID: "u1"})

	deleted, err := s.DeleteAllByResource(context.Background(), "1")
	if err != nil {
		t.Fatalf("DeleteAllByResource: %v", err)
	}
	if len(deleted) != 3 {
		t.Errorf("deleted %d notifications, want 3", len(deleted))
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
	_, err := s.DeleteAllByResource(context.Background(), "999")
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestDeleteByJoinRequestID(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "a1", Type: "group_join_request", ResourceID: "g1", ActorID: "u1", JoinRequestID: "jr-1"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "a2", Type: "group_join_request", ResourceID: "g1", ActorID: "u1", JoinRequestID: "jr-1"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "b1", Type: "group_join_request", ResourceID: "g2", ActorID: "u1", JoinRequestID: "jr-2"})

	deleted, err := s.DeleteByJoinRequestID(context.Background(), "jr-1")
	if err != nil {
		t.Fatalf("DeleteByJoinRequestID: %v", err)
	}
	if len(deleted) != 2 {
		t.Errorf("deleted %d notifications, want 2", len(deleted))
	}
	for _, n := range deleted {
		if n.JoinRequestID != "jr-1" || n.RecipientID == "" {
			t.Errorf("deleted notification = %+v, want join request jr-1 with recipient", n)
		}
	}

	_, total1, _ := s.GetByRecipient(context.Background(), "a1", 10, 0)
	if total1 != 0 {
		t.Errorf("a1 total = %d, want 0", total1)
	}
	_, total2, _ := s.GetByRecipient(context.Background(), "a2", 10, 0)
	if total2 != 0 {
		t.Errorf("a2 total = %d, want 0", total2)
	}

	ns3, total3, _ := s.GetByRecipient(context.Background(), "b1", 10, 0)
	if total3 != 1 {
		t.Errorf("b1 total = %d, want 1 (other join request survives)", total3)
	}
	if len(ns3) == 1 && ns3[0].JoinRequestID != "jr-2" {
		t.Errorf("b1 JoinRequestID = %q, want jr-2", ns3[0].JoinRequestID)
	}
}

func TestDeleteByJoinRequestID_NotFound(t *testing.T) {
	s := setupStore(t)
	_, err := s.DeleteByJoinRequestID(context.Background(), "999")
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestDeleteEventByRecipient(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "event", EventID: "evt-1"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "event", EventID: "evt-2"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u2", Type: "event", EventID: "evt-1"})

	deleted, err := s.DeleteEventByRecipient(context.Background(), "event", "u1", "evt-1")
	if err != nil {
		t.Fatalf("DeleteEventByRecipient: %v", err)
	}
	if len(deleted) != 1 || deleted[0].EventID != "evt-1" || deleted[0].RecipientID != "u1" {
		t.Errorf("deleted = %+v, want the u1 evt-1 notification", deleted)
	}

	ns, total, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 1 {
		t.Errorf("u1 total = %d, want 1", total)
	}
	if len(ns) == 1 && ns[0].EventID != "evt-2" {
		t.Errorf("u1 remaining EventID = %q, want evt-2", ns[0].EventID)
	}

	_, total2, _ := s.GetByRecipient(context.Background(), "u2", 10, 0)
	if total2 != 1 {
		t.Errorf("u2 total = %d, want 1 (different recipient survives)", total2)
	}
}

func TestDeleteEventByRecipient_NotFound(t *testing.T) {
	s := setupStore(t)
	_, err := s.DeleteEventByRecipient(context.Background(), "event", "u1", "999")
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestDeleteVoteNotifications_DeletesLikeAndDislike(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "dislike", ResourceType: "post", ResourceID: "1", ActorID: "u2"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "comment", ResourceType: "post", ResourceID: "1", ActorID: "u3"})

	deleted, err := s.DeleteVoteNotifications(context.Background(), "u2", "post", "1")
	if err != nil {
		t.Fatalf("DeleteVoteNotifications: %v", err)
	}
	if len(deleted) != 2 {
		t.Errorf("deleted %d notifications, want 2 (like and dislike)", len(deleted))
	}

	ns, total, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 1 {
		t.Errorf("total = %d, want 1 (like and dislike deleted)", total)
	}
	if len(ns) == 1 && ns[0].Type != "comment" {
		t.Errorf("remaining type = %q, want %q", ns[0].Type, "comment")
	}
}

func TestDeleteVoteNotifications_KeepsOtherResource(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "dislike", ResourceType: "post", ResourceID: "2", ActorID: "u2"})

	deleted, err := s.DeleteVoteNotifications(context.Background(), "u2", "post", "1")
	if err != nil {
		t.Fatalf("DeleteVoteNotifications: %v", err)
	}
	if len(deleted) != 1 || deleted[0].ResourceID != "1" {
		t.Errorf("deleted = %+v, want only the post 1 vote", deleted)
	}

	ns, total, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 1 {
		t.Errorf("total = %d, want 1 (only post 1 vote deleted)", total)
	}
	if len(ns) == 1 && ns[0].ResourceID != "2" {
		t.Errorf("remaining resource_id = %q, want %q", ns[0].ResourceID, "2")
	}
}

func TestDeleteVoteNotifications_NotFound(t *testing.T) {
	s := setupStore(t)
	_, err := s.DeleteVoteNotifications(context.Background(), "u1", "post", "999")
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestDeleteFollowNotifications(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u2", Type: "follow_request", ActorID: "u1"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "follow_accept", ActorID: "u2"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u2", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u1"})

	deleted, err := s.DeleteFollowNotifications(context.Background(), "u1", "u2")
	if err != nil {
		t.Fatalf("DeleteFollowNotifications: %v", err)
	}
	if len(deleted) != 1 || deleted[0].Type != "follow_accept" || deleted[0].RecipientID != "u1" {
		t.Errorf("deleted = %+v, want the follow_accept for u1 (follow_request was deduped on create)", deleted)
	}

	ns, total, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 0 {
		t.Errorf("u1 total = %d, want 0 (all follow notifications deleted)", total)
	}
	_ = ns

	ns2, total2, _ := s.GetByRecipient(context.Background(), "u2", 10, 0)
	if total2 != 1 {
		t.Errorf("u2 total = %d, want 1 (like survives)", total2)
	}
	if len(ns2) == 1 && ns2[0].Type != "like" {
		t.Errorf("remaining type = %q, want %q", ns2[0].Type, "like")
	}
}

func TestDeleteFollowNotifications_Symmetric(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u2", Type: "follow_request", ActorID: "u1"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "follow", ActorID: "u2"})

	deleted, err := s.DeleteFollowNotifications(context.Background(), "u2", "u1")
	if err != nil {
		t.Fatalf("DeleteFollowNotifications: %v", err)
	}
	if len(deleted) != 1 || deleted[0].Type != "follow" || deleted[0].RecipientID != "u1" {
		t.Errorf("deleted = %+v, want the follow for u1 (follow_request was deduped on create)", deleted)
	}

	_, total1, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if total1 != 0 {
		t.Errorf("u1 total = %d, want 0", total1)
	}
	_, total2, _ := s.GetByRecipient(context.Background(), "u2", 10, 0)
	if total2 != 0 {
		t.Errorf("u2 total = %d, want 0", total2)
	}
}

func TestDeleteFollowNotifications_NotFound(t *testing.T) {
	s := setupStore(t)
	_, err := s.DeleteFollowNotifications(context.Background(), "u1", "u2")
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestCreate_DedupesFollowNotifications(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u2", Type: "follow_request", ActorID: "u1"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "follow_accept", ActorID: "u2"})

	ns, total, _ := s.GetByRecipient(context.Background(), "u2", 10, 0)
	if total != 0 {
		t.Errorf("u2 total = %d, want 0 (stale follow_request replaced)", total)
	}
	_ = ns

	ns1, total1, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if total1 != 1 {
		t.Errorf("u1 total = %d, want 1", total1)
	}
	if len(ns1) == 1 && ns1[0].Type != "follow_accept" {
		t.Errorf("remaining type = %q, want %q", ns1[0].Type, "follow_accept")
	}
}

func TestCreate_NonFollowDoesNotDeleteFollow(t *testing.T) {
	s := setupStore(t)

	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "follow_request", ActorID: "u2"})
	_ = s.Create(context.Background(), &Notification{RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"})

	ns, total, _ := s.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 2 {
		t.Errorf("total = %d, want 2 (like creation must not touch follow)", total)
	}
	_ = ns
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
