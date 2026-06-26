package health

import (
	"net/http"
	"time"

	"social-network/internal/app/health/queries"
	"social-network/internal/domain/notification"
	"social-network/internal/infra/logger"
	"social-network/internal/pkg/helpers"

	notificationcommands "social-network/internal/app/notifications/commands"
)

type Handler struct {
	Logger             logger.Logger
	createNotification notificationcommands.CreateNotificationHandler
}

func NewHandler(logger logger.Logger, createNotificationHandler notificationcommands.CreateNotificationHandler) *Handler {
	return &Handler{
		Logger:             logger,
		createNotification: createNotificationHandler,
	}
}

func (h Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.Logger.PrintError(logger.ErrInvalidRequestMethod, nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")

		return
	}

	response := queries.HealthResponse{
		Status:    queries.StatusUp,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	userID := "df16d238-e4dd-4645-9101-54aed9c0fbf4"
	relatedID := "testUserID"
	testNotification := &notification.Notification{
		UserID:      userID,
		Type:        notification.NotificationTypeLike,
		Title:       "Test Notification",
		Message:     "This is a test notification from health check",
		RelatedType: "post",
		RelatedID:   relatedID,
		IsRead:      false,
	}
	err := h.createNotification.Handle(r.Context(), notificationcommands.CreateNotificationRequest{Notification: testNotification})
	if err != nil {
		h.Logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to create test notification")

		return
	}
	helpers.RespondWithJSON(
		w,
		http.StatusOK,
		nil,
		response,
	)
}
