package markasread

import (
	"net/http"
	"strconv"

	"social-network/internal/infra/middleware"

	notificationcommands "social-network/internal/app/notifications/commands"
)

type Handler struct {
	markAsRead notificationcommands.MarkAsReadHandler
}

func NewHandler(service notificationcommands.MarkAsReadHandler) *Handler {
	return &Handler{markAsRead: service}
}

func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
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

	notificationID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(
			w,
			"invalida notification ID",
			http.StatusBadRequest,
		)
		return
	}

	err = h.markAsRead.Handle(r.Context(), notificationcommands.MarkAsReadRequest{NotificationID: int(notificationID), UserID: userID})
	if err != nil {
		http.Error(
			w,
			"failed to mark as read",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusOK)
}
