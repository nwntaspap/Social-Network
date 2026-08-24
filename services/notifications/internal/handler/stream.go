package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"social-network/services/notifications/internal/middleware"
	"social-network/services/notifications/internal/store"
)

const tickerTime = 10

type StreamHub struct {
	mu     sync.RWMutex
	subs   map[string]map[int]chan store.Notification
	nextID int
}

func NewStreamHub() *StreamHub {
	return &StreamHub{
		subs: make(map[string]map[int]chan store.Notification),
	}
}

func (h *StreamHub) Subscribe(userID string) (<-chan store.Notification, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.subs[userID] == nil {
		h.subs[userID] = make(map[int]chan store.Notification)
	}

	h.nextID++
	id := h.nextID
	ch := make(chan store.Notification, 10)
	h.subs[userID][id] = ch

	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.subs[userID], id)
		close(ch)
	}
}

func (h *StreamHub) Publish(userID string, n store.Notification) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	log.Printf("this is the recipient:%s and this is the notificaiton:%v", userID, n)
	for _, ch := range h.subs[userID] {
		select {
		case ch <- n:
		default:
		}
	}
}

func (h *Notifications) StreamNotifications(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		middleware.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		middleware.RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported")
		return
	}

	ch, unsubscribe := h.hub.Subscribe(userID)
	defer unsubscribe()

	log.Printf("stream: client connected user=%s", userID)
	defer log.Printf("stream: client disconnected user=%s", userID)

	unreadCount, err := h.repo.GetUnreadCount(r.Context(), userID)
	if err == nil {
		fmt.Fprintf(w, "event: connected\ndata: {\"type\":\"connected\"}\n\n")
		flusher.Flush()

		fmt.Fprintf(w, "data: {\"type\":\"unread_count\",\"count\":%d}\n\n", unreadCount)
		flusher.Flush()
	}

	ticker := time.NewTicker(tickerTime * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case notification := <-ch:
			data, marshalErr := json.Marshal(notification)
			if marshalErr != nil {
				continue
			}
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
