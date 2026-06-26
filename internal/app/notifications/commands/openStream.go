package notificationcommands

import (
	"context"

	"social-network/internal/app/notifications"
	"social-network/internal/domain/notification"
)

type OpenStreamRequest struct {
	UserID string
}

type OpenStreamResponse struct {
	NotificationChan chan *notification.Notification
	UnreadCount      int
}

type OpenStreamHandler interface {
	Handle(ctx context.Context, req OpenStreamRequest) (*OpenStreamResponse, error)
	Close(req OpenStreamRequest, notificationChan chan *notification.Notification)
}

type openStreamHandler struct {
	repo     notification.Repository
	notifier notifications.Notifier
}

func NewOpenStreamHandler(repo notification.Repository, notifier notifications.Notifier) OpenStreamHandler {
	return &openStreamHandler{
		repo:     repo,
		notifier: notifier,
	}
}

func (h *openStreamHandler) Handle(ctx context.Context, req OpenStreamRequest) (*OpenStreamResponse, error) {
	notificationChan := h.notifier.RegisterClient(req.UserID)

	unreadCount, err := h.repo.GetUnreadCount(ctx, req.UserID)
	if err != nil {
		h.notifier.UnregisterClient(req.UserID, notificationChan)
		return nil, err
	}

	return &OpenStreamResponse{
		NotificationChan: notificationChan,
		UnreadCount:      unreadCount,
	}, nil
}

func (h *openStreamHandler) Close(req OpenStreamRequest, notificationChan chan *notification.Notification) {
	h.notifier.UnregisterClient(req.UserID, notificationChan)
}
