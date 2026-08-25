package handler

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"social-network/services/notifications/internal/store"
)

func TestStreamHub_SubscribePublish(t *testing.T) {
	hub := NewStreamHub()

	ch, unsubscribe := hub.Subscribe("u1")
	defer unsubscribe()

	n := store.Notification{
		ID: 1, RecipientID: "u1", Type: "like",
		ResourceType: "post", ResourceID: "1", ActorID: "u2",
	}
	hub.Publish("u1", n)

	select {
	case received := <-ch:
		if received.ID != 1 {
			t.Errorf("ID = %d, want 1", received.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for notification")
	}
}

func TestStreamHub_MultipleSubscribers(t *testing.T) {
	hub := NewStreamHub()

	ch1, unsub1 := hub.Subscribe("u1")
	defer unsub1()
	ch2, unsub2 := hub.Subscribe("u1")
	defer unsub2()

	n := store.Notification{ID: 1, RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"}
	hub.Publish("u1", n)

	select {
	case <-ch1:
	case <-time.After(time.Second):
		t.Error("subscriber 1 didn't receive")
	}
	select {
	case <-ch2:
	case <-time.After(time.Second):
		t.Error("subscriber 2 didn't receive")
	}
}

func TestStreamHub_UnsubscribeStopsDelivery(t *testing.T) {
	hub := NewStreamHub()

	ch, unsubscribe := hub.Subscribe("u1")
	unsubscribe()

	n := store.Notification{ID: 1, RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"}
	hub.Publish("u1", n)

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("received notification after unsubscribe")
		}
	case <-time.After(100 * time.Millisecond):
	}
}

func TestStreamNotifications_ConnectedEvent(t *testing.T) {
	h, _ := setupHandlerTest(t)

	w := httptest.NewRecorder()
	r := authenticatedRequest(t, http.MethodGet, "/api/v1/notifications/stream", "u1")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	r = r.WithContext(ctx)

	done := make(chan struct{})
	go func() {
		h.StreamNotifications(w, r)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	body := w.Body.String()
	if !strings.Contains(body, `event: connected`) {
		t.Error("missing connected event in SSE response")
	}
	if !strings.Contains(body, `"type":"connected"`) {
		t.Error("missing connected data in SSE response")
	}
}

func TestStreamNotifications_ReadsLines(t *testing.T) {
	h, _ := setupHandlerTest(t)

	w := httptest.NewRecorder()
	r := authenticatedRequest(t, http.MethodGet, "/api/v1/notifications/stream", "u1")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	r = r.WithContext(ctx)

	done := make(chan struct{})
	go func() {
		h.StreamNotifications(w, r)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	scanner := bufio.NewScanner(strings.NewReader(w.Body.String()))
	lines := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event:") || strings.HasPrefix(line, "data:") || strings.HasPrefix(line, ":") {
			lines++
		}
	}
	if lines == 0 {
		t.Error("no SSE events found in response")
	}
}

func TestStreamNotifications_EmitsUnreadCountAfterNotification(t *testing.T) {
	h, repo := setupHandlerTest(t)

	ctx := context.Background()
	err := repo.Create(ctx, &store.Notification{
		RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2",
	})
	if err != nil {
		t.Fatalf("seed notification: %v", err)
	}

	w := httptest.NewRecorder()
	r := authenticatedRequest(t, http.MethodGet, "/api/v1/notifications/stream", "u1")

	streamCtx, cancel := context.WithCancel(r.Context())
	defer cancel()
	r = r.WithContext(streamCtx)

	done := make(chan struct{})
	go func() {
		h.StreamNotifications(w, r)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	h.hub.Publish("u1", store.Notification{ID: 1, RecipientID: "u1", Type: "like", ResourceType: "post", ResourceID: "1", ActorID: "u2"})
	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	body := w.Body.String()
	notifIdx := strings.Index(body, "event: notification")
	if notifIdx == -1 {
		t.Fatalf("missing notification event in SSE response:\n%s", body)
	}
	// Every notification frame must be followed by an authoritative unread
	// count so the client badge can never drift from the server.
	lastCountIdx := strings.LastIndex(body, `"type":"unread_count"`)
	if lastCountIdx < notifIdx {
		t.Errorf("no unread_count frame after the notification event:\n%s", body)
	}
	if !strings.Contains(body, `"count":1`) {
		t.Errorf("unread count does not reflect the new notification:\n%s", body)
	}
}

func TestStreamNotifications_Unauthorized(t *testing.T) {
	h, _ := setupHandlerTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/stream", nil)
	h.StreamNotifications(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
