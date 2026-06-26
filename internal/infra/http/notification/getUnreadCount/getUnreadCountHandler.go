package getunreadcount

import (
	"encoding/json"
	"net/http"

	"social-network/internal/infra/middleware"

	notificationqueries "social-network/internal/app/notifications/queries"
)

type Handler struct {
	getUnreadCount notificationqueries.GetUnreadCountHandler
}

func NewHandler(service notificationqueries.GetUnreadCountHandler) *Handler {
	return &Handler{getUnreadCount: service}
}

func (h *Handler) GetUnread(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	if user.ID == "" {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	userID := user.ID

	count, err := h.getUnreadCount.Handle(r.Context(), notificationqueries.GetUnreadCountRequest{UserID: userID})
	if err != nil {
		http.Error(
			w,
			"failed to get count",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]int{"count": count})
	if err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
		return
	}
}
