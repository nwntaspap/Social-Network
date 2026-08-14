package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"social-network/services/notifications/internal/handler"
	"social-network/services/notifications/internal/platform/eventbus"
	"social-network/services/notifications/internal/store"
)

const consumerSchema = `
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
    join_request_id TEXT NOT NULL DEFAULT '',
    event_id TEXT NOT NULL DEFAULT '',
    is_read BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_notifications_recipient ON notifications(recipient_id, created_at DESC);
`

type mockMessage struct {
	body       []byte
	routingKey string
}

func (m *mockMessage) Body() []byte            { return m.body }
func (m *mockMessage) RoutingKey() string      { return m.routingKey }
func (m *mockMessage) Ack() error              { return nil }
func (m *mockMessage) Nack(requeue bool) error { return nil }

type mockEventBus struct {
	msgs chan eventbus.Message
	once sync.Once
}

func (m *mockEventBus) Subscribe(ctx context.Context, queue string) (<-chan eventbus.Message, error) {
	return m.msgs, nil
}

func (m *mockEventBus) Publish(exchange, routingkey string, body []byte) error {
	m.once.Do(func() {
		m.msgs = make(chan eventbus.Message, 10)
	})
	m.msgs <- &mockMessage{body: body, routingKey: routingkey}
	return nil
}

func (m *mockEventBus) InitTopology(ctx context.Context) error { return nil }

func setupConsumerTest(t *testing.T) (*Consumer, *store.SQLiteStore, *handler.StreamHub) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.ExecContext(context.Background(), consumerSchema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	repo := store.NewSQLiteStore(db)
	hub := handler.NewStreamHub()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return New(nil, repo, hub, log), repo, hub
}

func TestConsumer_ProcessesValidEvent(t *testing.T) {
	consumer, repo, hub := setupConsumerTest(t)

	ch, unsubscribe := hub.Subscribe("u1")
	defer unsubscribe()

	env := EventEnvelope{
		Type:         "post.liked",
		RecipientID:  "u1",
		ActorID:      "u2",
		ActorName:    "Bob",
		ResourceType: "post",
		ResourceID:   "42",
		ContentText:  "Bob liked your post",
	}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "created"}

	consumer.handle(msg)

	time.Sleep(50 * time.Millisecond)

	ns, total, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 1 {
		t.Fatalf("total = %d, want 1", total)
	}
	if ns[0].Type != "like" {
		t.Errorf("Type = %q, want %q", ns[0].Type, "like")
	}

	select {
	case n := <-ch:
		if n.ResourceID != "42" {
			t.Errorf("hub ResourceID = %s, want 42", n.ResourceID)
		}
		if n.Deleted {
			t.Errorf("hub Deleted = true, want false for created event")
		}
	case <-time.After(time.Second):
		t.Fatal("hub did not receive notification")
	}
}

func TestConsumer_ProcessesDeletedEvent(t *testing.T) {
	consumer, repo, hub := setupConsumerTest(t)

	// First create a notification
	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID:  "u1",
		Type:         "like",
		ResourceType: "post",
		ResourceID:   "42",
		ActorID:      "u2",
		ActorName:    "Bob",
	})

	ch, unsubscribe := hub.Subscribe("u1")
	defer unsubscribe()

	env := EventEnvelope{
		Type:         "post.liked",
		RecipientID:  "u1",
		ActorID:      "u2",
		ActorName:    "Bob",
		ResourceType: "post",
		ResourceID:   "42",
		ContentText:  "Bob liked your post",
	}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "deleted"}

	consumer.handle(msg)

	time.Sleep(50 * time.Millisecond)

	// Notification should be hard-deleted from DB
	ns, total, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 0 {
		t.Errorf("total = %d, want 0 (notification should be deleted)", total)
	}
	_ = ns

	// SSE should carry Deleted=true
	select {
	case n := <-ch:
		if !n.Deleted {
			t.Errorf("hub Deleted = false, want true for deleted event")
		}
		if n.ResourceID != "42" {
			t.Errorf("hub ResourceID = %s, want 42", n.ResourceID)
		}
	case <-time.After(time.Second):
		t.Fatal("hub did not receive notification")
	}
}

func TestConsumer_ProcessesCascadeDeletedEvent(t *testing.T) {
	consumer, repo, hub := setupConsumerTest(t)

	// Create multiple notifications for the same post from different actors
	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID:  "u1",
		Type:         "like",
		ResourceType: "post",
		ResourceID:   "42",
		ActorID:      "u2",
		ActorName:    "Bob",
	})
	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID:  "u1",
		Type:         "comment",
		ResourceType: "post",
		ResourceID:   "42",
		ActorID:      "u3",
		ActorName:    "Carol",
	})
	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID:  "u3",
		Type:         "like",
		ResourceType: "post",
		ResourceID:   "42",
		ActorID:      "u1",
		ActorName:    "Alice",
	})

	chU1, unsubscribeU1 := hub.Subscribe("u1")
	defer unsubscribeU1()
	chU3, unsubscribeU3 := hub.Subscribe("u3")
	defer unsubscribeU3()

	// Batch delete trigger: no recipient_id, no actor — it's just a cascade trigger.
	env := EventEnvelope{
		Type:         eventbus.EventPost,
		ResourceType: "post",
		ResourceID:   "42",
		ContentText:  "Your post was deleted",
	}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "deleted"}

	consumer.handle(msg)

	time.Sleep(50 * time.Millisecond)

	// All notifications for post 42 should be deleted (all recipients, all actors)
	ns1, total1, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total1 != 0 {
		t.Errorf("u1 total = %d, want 0 (all notifications for post 42 should be deleted)", total1)
	}
	_ = ns1

	ns3, total3, _ := repo.GetByRecipient(context.Background(), "u3", 10, 0)
	if total3 != 0 {
		t.Errorf("u3 total = %d, want 0 (notifications for post 42 from other recipients should also be deleted)", total3)
	}
	_ = ns3

	// u1 should receive exactly its two deleted notifications, each marked Deleted=true.
	u1Deleted := make([]store.Notification, 0, 2)
	for len(u1Deleted) < 2 {
		select {
		case n := <-chU1:
			if !n.Deleted {
				t.Errorf("hub Deleted = false, want true for deleted event")
			}
			if n.ResourceID != "42" {
				t.Errorf("hub ResourceID = %s, want 42", n.ResourceID)
			}
			u1Deleted = append(u1Deleted, n)
		case <-time.After(time.Second):
			t.Fatal("u1 did not receive all deleted notifications")
		}
	}
	gotU1Types := map[string]bool{}
	for _, n := range u1Deleted {
		gotU1Types[n.Type] = true
	}
	if !gotU1Types["like"] || !gotU1Types["comment"] {
		t.Errorf("u1 deleted types = %v, want like and comment", gotU1Types)
	}

	// u3 should receive its one deleted notification.
	select {
	case n := <-chU3:
		if !n.Deleted {
			t.Errorf("hub Deleted = false, want true for deleted event")
		}
		if n.ResourceID != "42" {
			t.Errorf("hub ResourceID = %s, want 42", n.ResourceID)
		}
	case <-time.After(time.Second):
		t.Fatal("u3 did not receive deleted notification")
	}
}

func TestConsumer_DeletesFollowNotifications(t *testing.T) {
	consumer, repo, hub := setupConsumerTest(t)

	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID: "u1",
		Type:        "follow_request",
		ActorID:     "u2",
	})
	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID:  "u1",
		Type:         "like",
		ResourceType: "post",
		ResourceID:   "42",
		ActorID:      "u2",
	})

	ch, unsubscribe := hub.Subscribe("u1")
	defer unsubscribe()

	env := EventEnvelope{
		Type:        eventbus.EventFollow,
		RecipientID: "u1",
		ActorID:     "u2",
	}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "deleted"}

	consumer.handle(msg)

	time.Sleep(50 * time.Millisecond)

	ns, total, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 1 {
		t.Fatalf("total = %d, want 1 (only the follow_request deleted)", total)
	}
	if ns[0].Type != "like" {
		t.Errorf("remaining type = %q, want %q", ns[0].Type, "like")
	}

	select {
	case n := <-ch:
		if !n.Deleted {
			t.Errorf("hub Deleted = false, want true for deleted event")
		}
	case <-time.After(time.Second):
		t.Fatal("hub did not receive notification")
	}
}

func TestConsumer_DeletesVoteNotifications(t *testing.T) {
	consumer, repo, hub := setupConsumerTest(t)

	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID:  "u1",
		Type:         "like",
		ResourceType: "post",
		ResourceID:   "42",
		ActorID:      "u2",
	})
	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID:  "u1",
		Type:         "dislike",
		ResourceType: "post",
		ResourceID:   "42",
		ActorID:      "u2",
	})
	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID:  "u1",
		Type:         "comment",
		ResourceType: "post",
		ResourceID:   "42",
		ActorID:      "u3",
	})

	ch, unsubscribe := hub.Subscribe("u1")
	defer unsubscribe()

	env := EventEnvelope{
		Type:         eventbus.EventPostVoteDeleted,
		RecipientID:  "u1",
		ActorID:      "u2",
		ResourceType: "post",
		ResourceID:   "42",
	}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "deleted"}

	consumer.handle(msg)

	time.Sleep(50 * time.Millisecond)

	ns, total, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 1 {
		t.Fatalf("total = %d, want 1 (like and dislike deleted)", total)
	}
	if ns[0].Type != "comment" {
		t.Errorf("remaining type = %q, want %q", ns[0].Type, "comment")
	}

	select {
	case n := <-ch:
		if !n.Deleted {
			t.Errorf("hub Deleted = false, want true for deleted event")
		}
	case <-time.After(time.Second):
		t.Fatal("hub did not receive notification")
	}
}

func TestConsumer_DeletesCommentVoteNotifications(t *testing.T) {
	consumer, repo, _ := setupConsumerTest(t)

	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID:  "u1",
		Type:         "like",
		ResourceType: "comment",
		ResourceID:   "7",
		ActorID:      "u2",
	})
	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID:  "u1",
		Type:         "like",
		ResourceType: "post",
		ResourceID:   "42",
		ActorID:      "u2",
	})

	env := EventEnvelope{
		Type:         eventbus.EventCommentVoteDeleted,
		RecipientID:  "u1",
		ActorID:      "u2",
		ResourceType: "comment",
		ResourceID:   "7",
	}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "deleted"}

	consumer.handle(msg)

	time.Sleep(50 * time.Millisecond)

	ns, total, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 1 {
		t.Fatalf("total = %d, want 1 (only the comment vote deleted)", total)
	}
	if ns[0].ResourceType != "post" {
		t.Errorf("remaining resource_type = %q, want %q", ns[0].ResourceType, "post")
	}
}

func TestConsumer_JoinRequestFanOut(t *testing.T) {
	consumer, repo, hub := setupConsumerTest(t)

	env := EventEnvelope{
		Type:               eventbus.EventGroupJoinRequested,
		ActorID:            "u1",
		ActorName:          "Alice",
		ActorAvatar:        "/alice.png",
		ResourceType:       "group",
		ResourceID:         "g1",
		ContentText:        "The Go Gophers",
		JoinRequestID:      "jr-1",
		MultipleRecipients: []string{"a1", "a2"},
	}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "created"}

	consumer.handle(msg)

	time.Sleep(50 * time.Millisecond)

	for _, admin := range []string{"a1", "a2"} {
		ns, total, _ := repo.GetByRecipient(context.Background(), admin, 10, 0)
		if total != 1 {
			t.Fatalf("admin %s total = %d, want 1", admin, total)
		}
		if ns[0].JoinRequestID != "jr-1" {
			t.Errorf("admin %s JoinRequestID = %q, want jr-1", admin, ns[0].JoinRequestID)
		}
		if ns[0].Type != "group_join_request" {
			t.Errorf("admin %s Type = %q, want group_join_request", admin, ns[0].Type)
		}
		if ns[0].RecipientID != admin {
			t.Errorf("admin %s RecipientID = %q, want %s", admin, ns[0].RecipientID, admin)
		}
	}

	_ = hub
}

func TestConsumer_JoinRequestRejectsIncomplete(t *testing.T) {
	consumer, repo, _ := setupConsumerTest(t)

	env := EventEnvelope{
		Type:          eventbus.EventGroupJoinRequested,
		ActorID:       "u1",
		JoinRequestID: "jr-1",
	}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "created"}

	consumer.handle(msg)

	_, total, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
}

func TestConsumer_JoinRequestCleanupOnResponse(t *testing.T) {
	consumer, repo, _ := setupConsumerTest(t)

	// Fan out a join request to two admins.
	reqEnv := EventEnvelope{
		Type:               eventbus.EventGroupJoinRequested,
		ActorID:            "u1",
		ResourceType:       "group",
		ResourceID:         "g1",
		JoinRequestID:      "jr-1",
		MultipleRecipients: []string{"a1", "a2"},
	}
	body, _ := json.Marshal(reqEnv)
	consumer.handle(&mockMessage{body: body, routingKey: "created"})

	time.Sleep(50 * time.Millisecond)

	for _, admin := range []string{"a1", "a2"} {
		_, total, _ := repo.GetByRecipient(context.Background(), admin, 10, 0)
		if total != 1 {
			t.Fatalf("admin %s total = %d, want 1 before response", admin, total)
		}
	}

	// Admin a1 accepts: pending fan-out rows must be removed and the requester notified.
	acceptEnv := EventEnvelope{
		Type:          eventbus.EventGroupJoinAccepted,
		RecipientID:   "u1",
		ActorID:       "a1",
		ResourceType:  "group",
		ResourceID:    "g1",
		JoinRequestID: "jr-1",
	}
	body, _ = json.Marshal(acceptEnv)
	consumer.handle(&mockMessage{body: body, routingKey: "created"})

	time.Sleep(50 * time.Millisecond)

	for _, admin := range []string{"a1", "a2"} {
		_, total, _ := repo.GetByRecipient(context.Background(), admin, 10, 0)
		if total != 0 {
			t.Errorf("admin %s total = %d, want 0 after response", admin, total)
		}
	}

	ns, total, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 1 {
		t.Fatalf("requester total = %d, want 1", total)
	}
	if ns[0].Type != "group_join_accept" {
		t.Errorf("requester Type = %q, want group_join_accept", ns[0].Type)
	}
}

func TestConsumer_RejectsIncompleteEvent(t *testing.T) {
	consumer, repo, _ := setupConsumerTest(t)

	env := EventEnvelope{Type: "post.liked", RecipientID: "u1"}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "created"}

	consumer.handle(msg)

	_, total, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
}

func TestConsumer_RejectsInvalidJSON(t *testing.T) {
	consumer, repo, _ := setupConsumerTest(t)

	msg := &mockMessage{body: []byte("not-json"), routingKey: "created"}
	consumer.handle(msg)

	_, total, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
}

func TestConsumer_ProcessesUpdatedEvent(t *testing.T) {
	consumer, repo, _ := setupConsumerTest(t)

	_ = repo.Create(context.Background(), &store.Notification{
		RecipientID:  "u1",
		Type:         "like",
		ResourceType: "post",
		ResourceID:   "1",
		ActorID:      "u2",
		ActorName:    "OldName",
		ActorAvatar:  "/old.png",
	})

	env := EventEnvelope{
		ActorID:     "u2",
		ActorName:   "NewName",
		ActorAvatar: "/new.png",
	}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "updated"}
	consumer.handle(msg)

	ns, _, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
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

func TestConsumer_RejectsIncompleteUpdateEvent(t *testing.T) {
	consumer, repo, _ := setupConsumerTest(t)

	env := EventEnvelope{ActorName: "Name", ActorAvatar: "/a.png"}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "updated"}
	consumer.handle(msg)

	ns, _, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if len(ns) != 0 {
		t.Errorf("got %d notifications, want 0 (no update should occur)", len(ns))
	}
}
