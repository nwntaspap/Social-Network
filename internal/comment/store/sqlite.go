package store

import (
	"context"
	"database/sql"

	"social-network/internal/comment"
	"social-network/internal/platform/database"
)

type SQLiteStore struct {
	db database.DB
}

func NewSQLiteStore(db database.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) CreateComment(ctx context.Context, c *comment.Comment) error {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO comments (user_id, topic_id, content, image_path)
		VALUES (?, ?, ?, ?)`,
		c.UserID, c.TopicID, c.Content, nullStr(c.ImagePath))
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = int(id)
	return nil
}

func (s *SQLiteStore) UpdateComment(ctx context.Context, c *comment.Comment) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE comments SET content = ?, image_path = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?`,
		c.Content, nullStr(c.ImagePath), c.ID, c.UserID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return comment.ErrCommentNotFound
	}
	return nil
}

func (s *SQLiteStore) DeleteComment(ctx context.Context, userID string, commentID int) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM comments WHERE id = ? AND user_id = ?`,
		commentID, userID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return comment.ErrCommentNotFound
	}
	return nil
}

func (s *SQLiteStore) GetCommentByID(ctx context.Context, commentID int) (*comment.Comment, error) {
	c := &comment.Comment{}
	var imgPath sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, topic_id, content, image_path, created_at, updated_at
		FROM comments WHERE id = ?`, commentID).Scan(
		&c.ID, &c.UserID, &c.TopicID, &c.Content,
		&imgPath, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, comment.ErrCommentNotFound
	}
	if err != nil {
		return nil, err
	}
	c.ImagePath = imgPath.String
	return c, nil
}

func (s *SQLiteStore) GetCommentByIDWithVotes(ctx context.Context, commentID int, userID *string) (*comment.Comment, error) {
	query := `
		SELECT
			c.id, c.user_id, c.topic_id, c.content, c.image_path,
			c.created_at, c.updated_at,
			COALESCE(vc.upvotes, 0), COALESCE(vc.downvotes, 0), COALESCE(vc.score, 0)`
	args := []any{}
	if userID != nil {
		query += `, uv.reaction_type`
	}
	query += `
		FROM comments c
		LEFT JOIN (
			SELECT comment_id,
				COUNT(CASE WHEN reaction_type = 1 THEN 1 END) AS upvotes,
				COUNT(CASE WHEN reaction_type = -1 THEN 1 END) AS downvotes,
				(COUNT(CASE WHEN reaction_type = 1 THEN 1 END) - COUNT(CASE WHEN reaction_type = -1 THEN 1 END)) AS score
			FROM votes WHERE comment_id IS NOT NULL GROUP BY comment_id
		) vc ON c.id = vc.comment_id`
	if userID != nil {
		query += ` LEFT JOIN votes uv ON c.id = uv.comment_id AND uv.user_id = ?`
		args = append(args, *userID)
	}
	query += ` WHERE c.id = ?`
	args = append(args, commentID)

	c := &comment.Comment{}
	var imgPath sql.NullString
	var userVote sql.NullInt32
	scanFields := []any{
		&c.ID, &c.UserID, &c.TopicID, &c.Content, &imgPath,
		&c.CreatedAt, &c.UpdatedAt,
		&c.UpvoteCount, &c.DownvoteCount, &c.VoteScore,
	}
	if userID != nil {
		scanFields = append(scanFields, &userVote)
	}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(scanFields...)
	if err == sql.ErrNoRows {
		return nil, comment.ErrCommentNotFound
	}
	if err != nil {
		return nil, err
	}
	c.ImagePath = imgPath.String
	if userID != nil && userVote.Valid {
		v := int(userVote.Int32)
		c.UserVote = &v
	}
	return c, nil
}

func (s *SQLiteStore) GetCommentsByTopicID(ctx context.Context, topicID int) ([]comment.Comment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, topic_id, content, image_path, created_at, updated_at
		FROM comments WHERE topic_id = ? ORDER BY created_at ASC`, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []comment.Comment
	for rows.Next() {
		var c comment.Comment
		var imgPath sql.NullString
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.TopicID, &c.Content,
			&imgPath, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		c.ImagePath = imgPath.String
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (s *SQLiteStore) GetCommentsByTopicIDWithVotes(ctx context.Context, topicID int, userID *string) ([]comment.Comment, error) {
	query := `
		SELECT
			c.id, c.user_id, c.topic_id, c.content, c.image_path,
			c.created_at, c.updated_at,
			COALESCE(vc.upvotes, 0), COALESCE(vc.downvotes, 0), COALESCE(vc.score, 0)`
	args := []any{}
	if userID != nil {
		query += `, uv.reaction_type`
	}
	query += `
		FROM comments c
		LEFT JOIN (
			SELECT comment_id,
				COUNT(CASE WHEN reaction_type = 1 THEN 1 END) AS upvotes,
				COUNT(CASE WHEN reaction_type = -1 THEN 1 END) AS downvotes,
				(COUNT(CASE WHEN reaction_type = 1 THEN 1 END) - COUNT(CASE WHEN reaction_type = -1 THEN 1 END)) AS score
			FROM votes WHERE comment_id IS NOT NULL GROUP BY comment_id
		) vc ON c.id = vc.comment_id`
	if userID != nil {
		query += ` LEFT JOIN votes uv ON c.id = uv.comment_id AND uv.user_id = ?`
		args = append(args, *userID)
	}
	query += ` WHERE c.topic_id = ? ORDER BY c.created_at ASC`
	args = append(args, topicID)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []comment.Comment
	for rows.Next() {
		var c comment.Comment
		var imgPath sql.NullString
		var userVote sql.NullInt32
		scanFields := []any{
			&c.ID, &c.UserID, &c.TopicID, &c.Content, &imgPath,
			&c.CreatedAt, &c.UpdatedAt,
			&c.UpvoteCount, &c.DownvoteCount, &c.VoteScore,
		}
		if userID != nil {
			scanFields = append(scanFields, &userVote)
		}
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}
		c.ImagePath = imgPath.String
		if userID != nil && userVote.Valid {
			v := int(userVote.Int32)
			c.UserVote = &v
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
