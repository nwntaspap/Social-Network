package notifications

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arnald/forum/internal/domain/notification"
)

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{DB: db}
}

func (r *Repo) Create(ctx context.Context, notification *notification.Notification) error {
	query := `
	INSERT INTO notifications (user_id, type, title, message, related_type, related_id, is_read)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id`

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(
		ctx,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.RelatedType,
		notification.RelatedID,
		notification.IsRead,
	).Scan(&notification.ID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

func (r *Repo) GetByUserID(ctx context.Context, userID string, limit int) ([]*notification.Notification, error) {
	query := `
	SELECT id, user_id, type, title, message, related_type, related_id, is_read, created_at
	FROM notifications
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT $2
	`

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(
		ctx,
		userID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	notifications := make([]*notification.Notification, 0)
	for rows.Next() {
		n := &notification.Notification{}
		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Message,
			&n.RelatedType,
			&n.RelatedID,
			&n.IsRead,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}

		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

func (r *Repo) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	query := `
	SELECT COUNT(*) FROM notifications
	WHERE user_id = $1 AND is_read = false`

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var count int
	err = stmt.QueryRowContext(
		ctx,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to scan: %w", err)
	}

	return count, nil
}

func (r *Repo) MarkAsRead(ctx context.Context, notificationID int, userID string) error {
	query := `
	UPDATE notifications
	SET is_read = true
	WHERE id = $1 AND user_id = $2`

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(
		ctx,
		notificationID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return err
}

func (r *Repo) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `
	UPDATE notifications
	SET is_read = true
	WHERE user_id = $1 AND is_read = false`

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(
		ctx,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return err
}
