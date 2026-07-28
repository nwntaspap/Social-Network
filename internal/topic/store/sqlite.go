package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"social-network/internal/platform/database"
	"social-network/internal/topic"
)

type SQLiteStore struct {
	db database.DB
}

func NewSQLiteStore(db database.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) CreateTopic(ctx context.Context, t *topic.Topic, allowedUserIDs []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO topics (user_id, title, content, image_path, visibility, group_id)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		t.UserID, t.Title, t.Content, t.ImagePath, t.Visibility, t.GroupID)
	if err != nil {
		return fmt.Errorf("insert topic: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	t.ID = int(id)

	err = tx.QueryRowContext(ctx,
		`SELECT created_at FROM topics WHERE id = ?`, t.ID).Scan(&t.CreatedAt)
	if err != nil {
		return fmt.Errorf("read created_at: %w", err)
	}

	if len(allowedUserIDs) > 0 {
		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO topic_allowed_users (topic_id, user_id) VALUES (?, ?)`)
		if err != nil {
			return fmt.Errorf("prepare allowed users: %w", err)
		}
		defer stmt.Close()
		for _, uid := range allowedUserIDs {
			if _, err := stmt.ExecContext(ctx, t.ID, uid); err != nil {
				return fmt.Errorf("insert allowed user %s: %w", uid, err)
			}
		}
	}

	return nil
}

func (s *SQLiteStore) UpdateTopic(ctx context.Context, t *topic.Topic, allowedUserIDs []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	result, err := tx.ExecContext(ctx,
		`UPDATE topics SET title = ?, content = ?, image_path = ?, visibility = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND user_id = ?`,
		t.Title, t.Content, t.ImagePath, t.Visibility, t.ID, t.UserID)
	if err != nil {
		return fmt.Errorf("update topic: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return topic.ErrTopicNotFound
	}

	if allowedUserIDs == nil {
		return nil
	}
	if _, delErr := tx.ExecContext(ctx, `DELETE FROM topic_allowed_users WHERE topic_id = ?`, t.ID); delErr != nil {
		return fmt.Errorf("delete allowed users: %w", delErr)
	}
	if len(allowedUserIDs) == 0 {
		return nil
	}
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO topic_allowed_users (topic_id, user_id) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare allowed users: %w", err)
	}
	defer stmt.Close()
	for _, uid := range allowedUserIDs {
		if _, err := stmt.ExecContext(ctx, t.ID, uid); err != nil {
			return fmt.Errorf("insert allowed user %s: %w", uid, err)
		}
	}

	return nil
}

func (s *SQLiteStore) DeleteTopic(ctx context.Context, userID string, topicID int) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM topics WHERE id = ? AND user_id = ?`, topicID, userID)
	if err != nil {
		return fmt.Errorf("delete topic: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return topic.ErrTopicNotFound
	}
	return nil
}

func (s *SQLiteStore) GetTopicByID(ctx context.Context, topicID int, userID *string) (*topic.Topic, error) {
	query := `
		SELECT
			t.id, t.user_id, t.title, t.content, t.image_path,
			COALESCE(t.visibility, 0), t.group_id,
			t.created_at, t.updated_at,
			COALESCE(u.username, ''),
			COALESCE(vc.upvotes, 0), COALESCE(vc.downvotes, 0), COALESCE(vc.score, 0)`

	if userID != nil {
		query += `, uv.reaction_type`
	}

	query += `
		FROM topics t
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN (
			SELECT topic_id,
				COUNT(CASE WHEN reaction_type = 1 THEN 1 END) as upvotes,
				COUNT(CASE WHEN reaction_type = -1 THEN 1 END) as downvotes,
				COUNT(CASE WHEN reaction_type = 1 THEN 1 END) - COUNT(CASE WHEN reaction_type = -1 THEN 1 END) as score
			FROM votes WHERE comment_id IS NULL GROUP BY topic_id
		) vc ON t.id = vc.topic_id`

	if userID != nil {
		query += `
		LEFT JOIN votes uv ON t.id = uv.topic_id AND uv.user_id = ? AND uv.comment_id IS NULL`
	}

	query += ` WHERE t.id = ?`

	args := make([]any, 0)
	if userID != nil {
		args = append(args, *userID)
	}
	args = append(args, topicID)

	var t topic.Topic
	var userVote sql.NullInt32
	var groupID sql.NullString
	var updatedAt sql.NullTime

	scanFields := []any{
		&t.ID, &t.UserID, &t.Title, &t.Content, &t.ImagePath,
		&t.Visibility, &groupID,
		&t.CreatedAt, &updatedAt,
		&t.OwnerUsername,
		&t.UpvoteCount, &t.DownvoteCount, &t.VoteScore,
	}
	if userID != nil {
		scanFields = append(scanFields, &userVote)
	}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(scanFields...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, topic.ErrTopicNotFound
		}
		return nil, fmt.Errorf("get topic: %w", err)
	}

	if groupID.Valid {
		t.GroupID = &groupID.String
	}
	t.UpdatedAt = database.ResolveTime(updatedAt, t.CreatedAt)
	if userID != nil && userVote.Valid {
		v := int(userVote.Int32)
		t.UserVote = &v
	}

	t.AllowedUsers, _ = s.getAllowedUsers(ctx, topicID)
	return &t, nil
}

func (s *SQLiteStore) GetImagePathFromTopicID(ctx context.Context, topicID int, userID string) (string, error) {
	var p sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT image_path FROM topics WHERE id = ? AND user_id = ?`, topicID, userID).Scan(&p)
	if err != nil {
		return "", err
	}
	if !p.Valid {
		return "", nil
	}
	return p.String, nil
}
