package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"social-network/services/notifications/internal/middleware"
	"social-network/services/notifications/internal/store"
)

type Notifications struct {
	repo store.Repository
	hub  *StreamHub
}

func New(repo store.Repository, hub *StreamHub) *Notifications {
	if hub == nil {
		hub = NewStreamHub()
	}
	return &Notifications{repo: repo, hub: hub}
}

func (h *Notifications) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		middleware.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	notifications, total, err := h.repo.GetByRecipient(r.Context(), userID, limit, 0)
	if err != nil {
		middleware.RespondWithError(w, http.StatusInternalServerError, "failed to fetch notifications")
		return
	}

	if notifications == nil {
		notifications = []store.Notification{}
	}

	resp := map[string]any{
		"notifications": notifications,
		"total":         total,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		middleware.RespondWithError(w, http.StatusInternalServerError, "failed to encode response")
	}
}

func (h *Notifications) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		middleware.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	count, err := h.repo.GetUnreadCount(r.Context(), userID)
	if err != nil {
		middleware.RespondWithError(w, http.StatusInternalServerError, "failed to get unread count")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]int{"count": count}); err != nil {
		middleware.RespondWithError(w, http.StatusInternalServerError, "failed to encode response")
	}
}

func (h *Notifications) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		middleware.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	notificationID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "invalid notification ID")
		return
	}

	if err := h.repo.MarkRead(r.Context(), int(notificationID), userID); err != nil {
		middleware.RespondWithError(w, http.StatusInternalServerError, "failed to mark as read")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Notifications) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		middleware.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.repo.MarkAllRead(r.Context(), userID); err != nil {
		middleware.RespondWithError(w, http.StatusInternalServerError, "failed to mark all as read")
		return
	}

	w.WriteHeader(http.StatusOK)
}
