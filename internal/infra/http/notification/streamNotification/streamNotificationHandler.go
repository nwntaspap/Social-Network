package streamnotification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"social-network/internal/infra/middleware"

	notificationcommands "social-network/internal/app/notifications/commands"
)

const tickerTime = 10

type Handler struct {
	openStream notificationcommands.OpenStreamHandler
}

func NewHandler(openStream notificationcommands.OpenStreamHandler) *Handler {
	return &Handler{openStream: openStream}
}

func (h *Handler) StreamNotifications(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	userID := user.ID

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(
			w,
			"Streaming unsupported",
			http.StatusInternalServerError,
		)
		return
	}

	streamReq := notificationcommands.OpenStreamRequest{UserID: userID}
	streamResp, err := h.openStream.Handle(r.Context(), streamReq)
	if err != nil {
		http.Error(
			w,
			"failed to open notification stream",
			http.StatusInternalServerError,
		)
		return
	}
	defer h.openStream.Close(streamReq, streamResp.NotificationChan)

	// connected
	fmt.Fprintf(w, "event: connected\ndata: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	fmt.Fprintf(
		w,
		"data: {\"type\":\"unread_count\",\"count\":%d}\n\n", streamResp.UnreadCount,
	)
	flusher.Flush()

	ticker := time.NewTicker(tickerTime * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			// client disconected
			return
		case notification := <-streamResp.NotificationChan:
			data, err := json.Marshal(notification)
			if err != nil {
				continue
			}
			// notification
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			// heartbeat (comments are valid SSE — no change needed)
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
