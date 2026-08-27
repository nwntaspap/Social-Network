package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"social-network/services/notifications/internal/middleware"
	"social-network/services/notifications/internal/store"
)

const testSchema = `
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
    join_request_id TEXT NOT NULL DEFAULT '',
    event_id TEXT NOT NULL DEFAULT '',
    is_read BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_notifications_recipient ON notifications(recipient_id, created_at DESC);
CREATE UNIQUE INDEX idx_notifications_active ON notifications(recipient_id, type, resource_type, resource_id, actor_id) WHERE deleted_at IS NULL;
`

func setupHandlerTest(t *testing.T) (*Notifications, *store.SQLiteStore) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.ExecContext(context.Background(), testSchema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	repo := store.NewSQLiteStore(db)
	hub := NewStreamHub()
	h := New(repo, hub)
	return h, repo
}

func authenticatedRequest(t *testing.T, method, target string, userID string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, target, nil)
	ctx := context.WithValue(r.Context(), middleware.UserIDKey(), userID)
	return r.WithContext(ctx)
}

func TestGetNotifications_Success(t *testing.T) {
	h, repo := setupHandlerTest(t)

	ctx := context.Background()
	for i := range 3 {
		_ = repo.Create(ctx, &store.Notification{
			RecipientID:  "u1",
			Type:         "like",
			ResourceType: "post",
			ResourceID:   strconv.Itoa(100 + i),
			ActorID:      "u2",
			ContentText:  "notification",
		})
	}

	w := httptest.NewRecorder()
	r := authenticatedRequest(t, http.MethodGet, "/api/v1/notifications", "u1")
	h.GetNotifications(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	notifs, ok := resp["notifications"].([]any)
	if !ok {
		t.Fatal("notifications not an array")
	}
	if len(notifs) != 3 {
		t.Errorf("got %d notifications, want 3", len(notifs))
	}

	total, ok := resp["total"].(float64)
	if !ok || int(total) != 3 {
		t.Errorf("total = %v, want 3", total)
	}
}

func TestGetNotifications_Empty(t *testing.T) {
	h, _ := setupHandlerTest(t)

	w := httptest.NewRecorder()
	r := authenticatedRequest(t, http.MethodGet, "/api/v1/notifications", "u1")
	h.GetNotifications(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	notifs := resp["notifications"].([]any)
	if len(notifs) != 0 {
		t.Errorf("got %d notifications, want 0", len(notifs))
	}
}

func TestGetNotifications_Unauthorized(t *testing.T) {
	h, _ := setupHandlerTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	h.GetNotifications(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestGetUnreadCount_Success(t *testing.T) {
	h, repo := setupHandlerTest(t)

	ctx := context.Background()
	for i := range 2 {
		_ = repo.Create(ctx, &store.Notification{
			RecipientID:  "u1",
			Type:         "like",
			ResourceType: "post",
			ResourceID:   strconv.Itoa(i),
			ActorID:      "u2",
		})
	}

	w := httptest.NewRecorder()
	r := authenticatedRequest(t, http.MethodGet, "/api/v1/notifications/unread-count", "u1")
	h.GetUnreadCount(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]int
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["count"] != 2 {
		t.Errorf("count = %d, want 2", resp["count"])
	}
}

func TestGetUnreadCount_Unauthorized(t *testing.T) {
	h, _ := setupHandlerTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/unread-count", nil)
	h.GetUnreadCount(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMarkAsRead_Success(t *testing.T) {
	h, repo := setupHandlerTest(t)

	ctx := context.Background()
	n := &store.Notification{
		RecipientID:  "u1",
		Type:         "like",
		ResourceType: "post",
		ResourceID:   "1",
		ActorID:      "u2",
	}
	_ = repo.Create(ctx, n)

	w := httptest.NewRecorder()
	r := authenticatedRequest(t, http.MethodPatch, "/api/v1/notifications/read?id="+itoa(n.ID), "u1")
	h.MarkAsRead(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestMarkAsRead_InvalidID(t *testing.T) {
	h, _ := setupHandlerTest(t)

	w := httptest.NewRecorder()
	r := authenticatedRequest(t, http.MethodPatch, "/api/v1/notifications/read?id=abc", "u1")
	h.MarkAsRead(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMarkAsRead_Unauthorized(t *testing.T) {
	h, _ := setupHandlerTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/read?id=1", nil)
	h.MarkAsRead(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMarkAllAsRead_Success(t *testing.T) {
	h, repo := setupHandlerTest(t)

	ctx := context.Background()
	for i := range 3 {
		_ = repo.Create(ctx, &store.Notification{
			RecipientID:  "u1",
			Type:         "like",
			ResourceType: "post",
			ResourceID:   strconv.Itoa(i),
			ActorID:      "u2",
		})
	}

	w := httptest.NewRecorder()
	r := authenticatedRequest(t, http.MethodPatch, "/api/v1/notifications/read-all", "u1")
	h.MarkAllAsRead(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	count, _ := repo.GetUnreadCount(ctx, "u1")
	if count != 0 {
		t.Errorf("unread count = %d, want 0", count)
	}
}

func TestMarkAllAsRead_Unauthorized(t *testing.T) {
	h, _ := setupHandlerTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/read-all", nil)
	h.MarkAllAsRead(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func itoa(i int) string {
	return string(rune('0' + i))
}
