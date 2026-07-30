package store

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("notification not found")

type Notification struct {
	ID           int
	RecipientID  string
	Type         string
	ResourceType string
	ResourceID   string
	ActorID      string
	ActorName    string
	ActorAvatar  string
	ContentText  string
	ImageURL     string
	IsRead       bool
	CreatedAt    time.Time
	Deleted      bool
}

type Repository interface {
	Create(ctx context.Context, n *Notification) error
	GetByRecipient(ctx context.Context, recipientID string, limit, offset int) ([]Notification, int, error)
	GetUnreadCount(ctx context.Context, recipientID string) (int, error)
	MarkRead(ctx context.Context, id int, recipientID string) error
	MarkAllRead(ctx context.Context, recipientID string) error
	DeleteByResource(ctx context.Context, actorID, resourceType string, resourceID string) error
	DeleteAllByResource(ctx context.Context, resourceID string) error
	UpdateActorInfo(ctx context.Context, actorID, name, avatar string) error
}
