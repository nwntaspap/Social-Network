package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"social-network/services/notifications/internal/handler"
	"social-network/services/notifications/internal/platform/eventbus"
	"social-network/services/notifications/internal/store"
)

func TestConsumer_Integration_CreatedRoutingKey(t *testing.T) {
	broker, err := eventbus.NewGoBroker()
	if err != nil {
		t.Skipf("broker not available: %v", err)
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(context.Background(), consumerSchema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	repo := store.NewSQLiteStore(db)
	hub := handler.NewStreamHub()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	c := New(broker, repo, hub, log)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		if err := c.Start(ctx); err != nil && err != context.Canceled {
			t.Errorf("consumer stopped: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	tests := []struct {
		name       string
		routingKey string
		envelope   EventEnvelope
		wantType   string
	}{
		{
			name:       "post_liked",
			routingKey: "created",
			envelope: EventEnvelope{
				Type:         "post.liked",
				RecipientID:  "int-u1",
				ActorID:      "int-u2",
				ActorName:    "Alice",
				ResourceType: "post",
				ResourceID:   "1",
				ContentText:  "Alice liked your post",
			},
			wantType: "like",
		},
		{
			name:       "comment_liked",
			routingKey: "created",
			envelope: EventEnvelope{
				Type:         "comment.liked",
				RecipientID:  "int-u3",
				ActorID:      "int-u4",
				ActorName:    "Carol",
				ResourceType: "comment",
				ResourceID:   "10",
				ContentText:  "Carol liked your comment",
			},
			wantType: "like",
		},
		{
			name:       "follow_requested",
			routingKey: "created",
			envelope: EventEnvelope{
				Type:         "follow.requested",
				RecipientID:  "int-u5",
				ActorID:      "int-u6",
				ActorName:    "Eve",
				ResourceType: "user",
				ResourceID:   "6",
				ContentText:  "Eve wants to follow you",
			},
			wantType: "follow_request",
		},
		{
			name:       "follow_accepted",
			routingKey: "created",
			envelope: EventEnvelope{
				Type:         "follow.accepted",
				RecipientID:  "int-u7",
				ActorID:      "int-u8",
				ActorName:    "Grace",
				ResourceType: "user",
				ResourceID:   "8",
				ContentText:  "Grace accepted your follow request",
			},
			wantType: "follow_accept",
		},
		{
			name:       "group_invitation",
			routingKey: "created",
			envelope: EventEnvelope{
				Type:         "group.invitation",
				RecipientID:  "int-u9",
				ActorID:      "int-u10",
				ActorName:    "Ivan",
				ResourceType: "group",
				ResourceID:   "20",
				ContentText:  "Ivan invited you to Group",
			},
			wantType: "group_invite",
		},
		{
			name:       "event_created",
			routingKey: "created",
			envelope: EventEnvelope{
				Type:         "event.created",
				RecipientID:  "int-u11",
				ActorID:      "int-u12",
				ActorName:    "Karl",
				ResourceType: "event",
				ResourceID:   "30",
				ContentText:  "Karl created a new event",
			},
			wantType: "event_created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, unsubscribe := hub.Subscribe(tt.envelope.RecipientID)
			defer unsubscribe()

			body, err := json.Marshal(tt.envelope)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			if err := broker.Publish("notifications.exchange", tt.routingKey, body); err != nil {
				t.Fatalf("publish: %v", err)
			}

			select {
			case n := <-ch:
				if n.Type != tt.wantType {
					t.Errorf("hub notification type = %q, want %q", n.Type, tt.wantType)
				}
				if n.RecipientID != tt.envelope.RecipientID {
					t.Errorf("hub recipient_id = %q, want %q", n.RecipientID, tt.envelope.RecipientID)
				}
				if n.ActorID != tt.envelope.ActorID {
					t.Errorf("hub actor_id = %q, want %q", n.ActorID, tt.envelope.ActorID)
				}
				if n.ResourceID != tt.envelope.ResourceID {
					t.Errorf("hub resource_id = %s, want %s", n.ResourceID, tt.envelope.ResourceID)
				}
				if n.Deleted {
					t.Errorf("hub Deleted = true for created event, want false")
				}
			case <-time.After(5 * time.Second):
				t.Fatal("timed out waiting for notification on hub")
			}

			ns, total, err := repo.GetByRecipient(context.Background(), tt.envelope.RecipientID, 10, 0)
			if err != nil {
				t.Fatalf("get by recipient: %v", err)
			}
			if total != 1 {
				t.Fatalf("store total = %d, want 1 for recipient %s", total, tt.envelope.RecipientID)
			}
			if ns[0].Type != tt.wantType {
				t.Errorf("store notification type = %q, want %q", ns[0].Type, tt.wantType)
			}
			if ns[0].ActorID != tt.envelope.ActorID {
				t.Errorf("store actor_id = %q, want %q", ns[0].ActorID, tt.envelope.ActorID)
			}
		})
	}
}

func TestConsumer_Integration_RejectsInvalidJSON(t *testing.T) {
	broker, err := eventbus.NewGoBroker()
	if err != nil {
		t.Skipf("broker not available: %v", err)
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(context.Background(), consumerSchema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	repo := store.NewSQLiteStore(db)
	hub := handler.NewStreamHub()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	c := New(broker, repo, hub, log)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		if err := c.Start(ctx); err != nil && err != context.Canceled {
			t.Errorf("consumer stopped: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	invalidJSON := []byte(`not-json`)

	if err := broker.Publish("notifications.exchange", "created", invalidJSON); err != nil {
		t.Fatalf("publish: %v", err)
	}

	time.Sleep(2 * time.Second)

	_, total, err := repo.GetByRecipient(context.Background(), "anyone", 10, 0)
	if err != nil {
		t.Fatalf("get by recipient: %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0 notifications after invalid JSON, got %d", total)
	}
}

func TestConsumer_Integration_RejectsIncompleteEvent(t *testing.T) {
	broker, err := eventbus.NewGoBroker()
	if err != nil {
		t.Skipf("broker not available: %v", err)
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(context.Background(), consumerSchema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	repo := store.NewSQLiteStore(db)
	hub := handler.NewStreamHub()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	c := New(broker, repo, hub, log)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		if err := c.Start(ctx); err != nil && err != context.Canceled {
			t.Errorf("consumer stopped: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	env := EventEnvelope{Type: "post.liked", RecipientID: "int-missing"}
	body, _ := json.Marshal(env)

	if err := broker.Publish("notifications.exchange", "created", body); err != nil {
		t.Fatalf("publish: %v", err)
	}

	time.Sleep(2 * time.Second)

	_, total, err := repo.GetByRecipient(context.Background(), "int-missing", 10, 0)
	if err != nil {
		t.Fatalf("get by recipient: %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0 notifications after incomplete event, got %d", total)
	}
}
