package markallasread

import (
	"net/http"

	"social-network/internal/infra/middleware"

	notificationcommands "social-network/internal/app/notifications/commands"
)

type Handler struct {
	markAllAsRead notificationcommands.MarkAllAsReadHandler
}

func NewHandler(service notificationcommands.MarkAllAsReadHandler) *Handler {
	return &Handler{markAllAsRead: service}
}

func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
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

	err := h.markAllAsRead.Handle(r.Context(), notificationcommands.MarkAllAsReadRequest{UserID: userID})
	if err != nil {
		http.Error(
			w,
			"failed to mark all as read",
			http.StatusInternalServerError,
		)
		return
	}
	w.WriteHeader(http.StatusOK)
}
