package store

import (
	"context"
	"fmt"
	"time"

	"social-network/services/notifications/internal/platform/database"
)

type SQLiteStore struct {
	db database.DB
}

func NewSQLiteStore(db database.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) Create(ctx context.Context, n *Notification) error {
	query := `
		INSERT INTO notifications
			(recipient_id, type, resource_type, resource_id, actor_id,
			 actor_name, actor_avatar, content_text, image_url,
			 is_read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?)`

	now := time.Now()
	result, err := s.db.ExecContext(
		ctx, query,
		n.RecipientID, n.Type, n.ResourceType, n.ResourceID, n.ActorID,
		n.ActorName, n.ActorAvatar, n.ContentText, n.ImageURL,
		now,
	)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	n.ID = int(id)
	n.CreatedAt = now
	return nil
}

func (s *SQLiteStore) GetByRecipient(ctx context.Context, recipientID string, limit, offset int) ([]Notification, int, error) {
	countQuery := `SELECT COUNT(*) FROM notifications WHERE recipient_id = ?`
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, recipientID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	query := `
		SELECT id, recipient_id, type, resource_type, resource_id,
		       actor_id, actor_name, actor_avatar, content_text, image_url,
		       is_read, created_at
		FROM notifications
		WHERE recipient_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := s.db.QueryContext(ctx, query, recipientID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get notifications: %w", err)
	}
	defer rows.Close()

	var ns []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(
			&n.ID, &n.RecipientID, &n.Type, &n.ResourceType, &n.ResourceID,
			&n.ActorID, &n.ActorName, &n.ActorAvatar, &n.ContentText, &n.ImageURL,
			&n.IsRead, &n.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan notification: %w", err)
		}
		ns = append(ns, n)
	}
	return ns, total, rows.Err()
}

func (s *SQLiteStore) GetUnreadCount(ctx context.Context, recipientID string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE recipient_id = ? AND is_read = 0`
	var count int
	if err := s.db.QueryRowContext(ctx, query, recipientID).Scan(&count); err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}
	return count, nil
}

func (s *SQLiteStore) MarkRead(ctx context.Context, id int, recipientID string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE notifications SET is_read = 1 WHERE id = ? AND recipient_id = ?`,
		id, recipientID)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) MarkAllRead(ctx context.Context, recipientID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE notifications SET is_read = 1 WHERE recipient_id = ? AND is_read = 0`,
		recipientID)
	if err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}

func (s *SQLiteStore) DeleteByResource(ctx context.Context, actorID, resourceType string, resourceID string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM notifications
		 WHERE actor_id = ? AND resource_type = ? AND resource_id = ?`,
		actorID, resourceType, resourceID)
	if err != nil {
		return fmt.Errorf("delete notification by resource: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) DeleteAllByResource(ctx context.Context, resourceID string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM notifications
		 WHERE resource_id = ?`,
		resourceID)
	if err != nil {
		return fmt.Errorf("delete all notifications by resource: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) UpdateActorInfo(ctx context.Context, actorID, name, avatar string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE notifications SET actor_name = ?, actor_avatar = ? WHERE actor_id = ?`,
		name, avatar, actorID)
	if err != nil {
		return fmt.Errorf("update actor info: %w", err)
	}
	return nil
}
