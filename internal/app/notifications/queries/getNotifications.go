package notificationqueries

import (
	"context"

	"social-network/internal/domain/notification"
)

type GetNotificationsRequest struct {
	UserID string
	Limit  int
}

type GetNotificationsHandler interface {
	Handle(ctx context.Context, req GetNotificationsRequest) ([]*notification.Notification, error)
}

type getNotificationsHandler struct {
	repo notification.Repository
}

func NewGetNotificationsHandler(repo notification.Repository) GetNotificationsHandler {
	return &getNotificationsHandler{
		repo: repo,
	}
}

func (h *getNotificationsHandler) Handle(ctx context.Context, req GetNotificationsRequest) ([]*notification.Notification, error) {
	return h.repo.GetByUserID(ctx, req.UserID, req.Limit)
}
