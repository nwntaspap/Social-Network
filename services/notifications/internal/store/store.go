package store

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("notification not found")

type Notification struct {
	ID            int       `json:"id"`
	RecipientID   string    `json:"recipient_id"`
	Type          string    `json:"type"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    string    `json:"resource_id"`
	ActorID       string    `json:"actor_id"`
	ActorName     string    `json:"actor_name"`
	ActorAvatar   string    `json:"actor_avatar"`
	ContentText   string    `json:"content_text"`
	ImageURL      string    `json:"image_url"`
	JoinRequestID string    `json:"join_request_id"`
	EventID       string    `json:"event_id"`
	IsRead        bool      `json:"is_read"`
	CreatedAt     time.Time `json:"created_at"`
	Deleted       bool      `json:"deleted"`
}

type Repository interface {
	Create(ctx context.Context, n *Notification) error
	GetByRecipient(ctx context.Context, recipientID string, limit, offset int) ([]Notification, int, error)
	GetUnreadCount(ctx context.Context, recipientID string) (int, error)
	MarkRead(ctx context.Context, id int, recipientID string) error
	MarkAllRead(ctx context.Context, recipientID string) error
	DeleteByResource(ctx context.Context, typ, actorID, resourceType, resourceID string) error
	DeleteEventByRecipient(ctx context.Context, typ, recipientId, eventID string) error
	DeleteFollowNotifications(ctx context.Context, userID, otherUserID string) error
	DeleteVoteNotifications(ctx context.Context, actorID, resourceType, resourceID string) error
	DeleteAllByResource(ctx context.Context, resourceID string) error
	DeleteByJoinRequestID(ctx context.Context, joinRequestID string) error
	UpdateActorInfo(ctx context.Context, actorID, name, avatar string) error
}
