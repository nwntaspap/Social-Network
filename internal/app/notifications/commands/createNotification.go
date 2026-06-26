package notificationcommands

import (
	"context"

	"social-network/internal/app/notifications"
	"social-network/internal/domain/notification"
)

type CreateNotificationRequest struct {
	Notification *notification.Notification
}

type CreateNotificationHandler interface {
	Handle(ctx context.Context, req CreateNotificationRequest) error
}

type createNotificationHandler struct {
	repo     notification.Repository
	notifier notifications.Notifier
}

func NewCreateNotificationHandler(repo notification.Repository, notifier notifications.Notifier) CreateNotificationHandler {
	return &createNotificationHandler{
		repo:     repo,
		notifier: notifier,
	}
}

func (h *createNotificationHandler) Handle(ctx context.Context, req CreateNotificationRequest) error {
	err := h.repo.Create(ctx, req.Notification)
	if err != nil {
		return err
	}
	h.notifier.BroadcastToUser(req.Notification.UserID, req.Notification)
	return nil
}
