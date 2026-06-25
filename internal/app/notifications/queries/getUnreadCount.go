package notificationqueries

import (
	"context"

	"social-network/internal/domain/notification"
)

type GetUnreadCountRequest struct {
	UserID string
}

type GetUnreadCountHandler interface {
	Handle(ctx context.Context, req GetUnreadCountRequest) (int, error)
}

type getUnreadCountHandler struct {
	repo notification.Repository
}

func NewGetUnreadCountHandler(repo notification.Repository) GetUnreadCountHandler {
	return &getUnreadCountHandler{
		repo: repo,
	}
}

func (h *getUnreadCountHandler) Handle(ctx context.Context, req GetUnreadCountRequest) (int, error) {
	return h.repo.GetUnreadCount(ctx, req.UserID)
}
