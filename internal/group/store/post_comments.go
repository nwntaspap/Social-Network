package store

import (
	"context"
	"database/sql"
	"fmt"

	"social-network/internal/group"
)

func (s *SQLiteStore) CreatePostComment(ctx context.Context, c *group.PostComment) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO group_post_comments (id, post_id, author_id, content, image_path) VALUES (?, ?, ?, ?, ?)`,
		c.ID, c.PostID, c.AuthorID, c.Content, c.ImagePath)
	return err
}

func (s *SQLiteStore) GetPostComments(ctx context.Context, postID string, page, size int) ([]group.PostComment, int, error) {
	var total int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM group_post_comments WHERE post_id = ?`, postID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count post comments: %w", err)
	}

	offset := (page - 1) * size
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, post_id, author_id, content, image_path, created_at
		 FROM group_post_comments WHERE post_id = ?
		 ORDER BY created_at ASC LIMIT ? OFFSET ?`,
		postID, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list post comments: %w", err)
	}
	defer rows.Close()

	var comments []group.PostComment
	for rows.Next() {
		var c group.PostComment
		var imagePath sql.NullString
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Content, &imagePath, &c.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan post comment: %w", err)
		}
		c.ImagePath = imagePath.String
		comments = append(comments, c)
	}
	return comments, total, rows.Err()
}

func (s *SQLiteStore) CountPostComments(ctx context.Context, postID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM group_post_comments WHERE post_id = ?`, postID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count post comments: %w", err)
	}
	return count, nil
}
