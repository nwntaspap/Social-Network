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
    resource_type TEXT NOT NULL,
    resource_id INTEGER NOT NULL,
    actor_id TEXT NOT NULL,
    actor_name TEXT NOT NULL DEFAULT '',
    actor_avatar TEXT NOT NULL DEFAULT '',
    content_text TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    is_read BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_notifications_recipient ON notifications(recipient_id, created_at DESC);
CREATE UNIQUE INDEX idx_notifications_active ON notifications(recipient_id, type, resource_type, resource_id, actor_id) WHERE deleted_at IS NULL;
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
		ResourceID:   42,
		ContentText:  "Bob liked your post",
	}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "post.liked"}

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
		if n.ResourceID != 42 {
			t.Errorf("hub ResourceID = %d, want 42", n.ResourceID)
		}
	case <-time.After(time.Second):
		t.Fatal("hub did not receive notification")
	}
}

func TestConsumer_RejectsIncompleteEvent(t *testing.T) {
	consumer, repo, _ := setupConsumerTest(t)

	env := EventEnvelope{Type: "post.liked", RecipientID: "u1"}
	body, _ := json.Marshal(env)
	msg := &mockMessage{body: body, routingKey: "post.liked"}

	consumer.handle(msg)

	_, total, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
}

func TestConsumer_RejectsInvalidJSON(t *testing.T) {
	consumer, repo, _ := setupConsumerTest(t)

	msg := &mockMessage{body: []byte("not-json"), routingKey: "post.liked"}
	consumer.handle(msg)

	_, total, _ := repo.GetByRecipient(context.Background(), "u1", 10, 0)
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
}
