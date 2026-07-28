package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"social-network/internal/platform/database"
	"social-network/internal/topic"
)

func (s *SQLiteStore) GetFeed(ctx context.Context, userID string, page, size int, orderBy, order, filter string) ([]topic.Topic, int, error) {
	whereClause := `WHERE 1=1`
	args := make([]any, 0)

	if filter != "" {
		whereClause += " AND (t.title LIKE ? OR t.content LIKE ?)"
		fp := "%" + filter + "%"
		args = append(args, fp, fp)
	}

	countQuery := "SELECT COUNT(DISTINCT t.id) FROM topics t " + whereClause
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count feed: %w", err)
	}

	orderByCol := sanitizeOrderBy(orderBy)
	orderDir := sanitizeOrder(order)

	query := "SELECT t.id, t.user_id, t.title, t.content, t.image_path, " +
		"COALESCE(t.visibility, 0), t.group_id, " +
		"t.created_at, t.updated_at, " +
		"COALESCE(u.username, ''), " +
		"COALESCE(vc.upvotes, 0), COALESCE(vc.downvotes, 0), COALESCE(vc.score, 0), " +
		"uv.reaction_type " +
		"FROM topics t " +
		"LEFT JOIN users u ON t.user_id = u.id " +
		"LEFT JOIN (SELECT topic_id, " +
		"COUNT(CASE WHEN reaction_type = 1 THEN 1 END) as upvotes, " +
		"COUNT(CASE WHEN reaction_type = -1 THEN 1 END) as downvotes, " +
		"COUNT(CASE WHEN reaction_type = 1 THEN 1 END) - COUNT(CASE WHEN reaction_type = -1 THEN 1 END) as score " +
		"FROM votes WHERE comment_id IS NULL GROUP BY topic_id) vc ON t.id = vc.topic_id " +
		"LEFT JOIN votes uv ON t.id = uv.topic_id AND uv.user_id = ? AND uv.comment_id IS NULL " +
		whereClause + " ORDER BY " + orderByCol + " " + orderDir + " LIMIT ? OFFSET ?"

	allArgs := make([]any, 0, len(args)+3)
	allArgs = append(allArgs, userID)
	allArgs = append(allArgs, args...)
	offset := (page - 1) * size
	allArgs = append(allArgs, size, offset)

	rows, err := s.db.QueryContext(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query feed: %w", err)
	}
	defer rows.Close()

	topics, err := collectTopics(rows, true)
	if err != nil {
		return nil, 0, err
	}
	return topics, total, nil
}

func (s *SQLiteStore) GetTopicsByUserID(ctx context.Context, ownerID, requesterID string, page, size int) ([]topic.Topic, int, error) {
	whereClause := `WHERE t.user_id = ?`
	args := []any{ownerID}

	countQuery := "SELECT COUNT(DISTINCT t.id) FROM topics t " + whereClause
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count user topics: %w", err)
	}

	query := "SELECT t.id, t.user_id, t.title, t.content, t.image_path, " +
		"COALESCE(t.visibility, 0), t.group_id, " +
		"t.created_at, t.updated_at, " +
		"COALESCE(u.username, ''), " +
		"COALESCE(vc.upvotes, 0), COALESCE(vc.downvotes, 0), COALESCE(vc.score, 0), " +
		"uv.reaction_type " +
		"FROM topics t " +
		"LEFT JOIN users u ON t.user_id = u.id " +
		"LEFT JOIN (SELECT topic_id, " +
		"COUNT(CASE WHEN reaction_type = 1 THEN 1 END) as upvotes, " +
		"COUNT(CASE WHEN reaction_type = -1 THEN 1 END) as downvotes, " +
		"COUNT(CASE WHEN reaction_type = 1 THEN 1 END) - COUNT(CASE WHEN reaction_type = -1 THEN 1 END) as score " +
		"FROM votes WHERE comment_id IS NULL GROUP BY topic_id) vc ON t.id = vc.topic_id " +
		"LEFT JOIN votes uv ON t.id = uv.topic_id AND uv.user_id = ? AND uv.comment_id IS NULL " +
		whereClause + " ORDER BY t.created_at DESC LIMIT ? OFFSET ?"

	offset := (page - 1) * size
	allArgs := []any{requesterID}
	allArgs = append(allArgs, args...)
	allArgs = append(allArgs, size, offset)

	rows, err := s.db.QueryContext(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query user topics: %w", err)
	}
	defer rows.Close()

	topics, err := collectTopics(rows, true)
	if err != nil {
		return nil, 0, err
	}
	return topics, total, nil
}

func (s *SQLiteStore) GetTopicsByGroupID(ctx context.Context, groupID string, page, size int) ([]topic.Topic, int, error) {
	whereClause := `WHERE t.group_id = ?`
	args := []any{groupID}

	countQuery := "SELECT COUNT(DISTINCT t.id) FROM topics t " + whereClause
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count group topics: %w", err)
	}

	query := "SELECT t.id, t.user_id, t.title, t.content, t.image_path, " +
		"COALESCE(t.visibility, 0), t.group_id, " +
		"t.created_at, t.updated_at, " +
		"COALESCE(u.username, ''), " +
		"COALESCE(vc.upvotes, 0), COALESCE(vc.downvotes, 0), COALESCE(vc.score, 0) " +
		"FROM topics t " +
		"LEFT JOIN users u ON t.user_id = u.id " +
		"LEFT JOIN (SELECT topic_id, " +
		"COUNT(CASE WHEN reaction_type = 1 THEN 1 END) as upvotes, " +
		"COUNT(CASE WHEN reaction_type = -1 THEN 1 END) as downvotes, " +
		"COUNT(CASE WHEN reaction_type = 1 THEN 1 END) - COUNT(CASE WHEN reaction_type = -1 THEN 1 END) as score " +
		"FROM votes WHERE comment_id IS NULL GROUP BY topic_id) vc ON t.id = vc.topic_id " +
		whereClause + " ORDER BY t.created_at DESC LIMIT ? OFFSET ?"

	offset := (page - 1) * size
	allArgs := append(args, size, offset)

	rows, err := s.db.QueryContext(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query group topics: %w", err)
	}
	defer rows.Close()

	topics, err := collectTopics(rows, false)
	if err != nil {
		return nil, 0, err
	}
	return topics, total, nil
}

func collectTopics(rows *sql.Rows, includeUserVote bool) ([]topic.Topic, error) {
	var topics []topic.Topic
	for rows.Next() {
		var t topic.Topic
		var groupID sql.NullString
		var updatedAt sql.NullTime

		scanArgs := []any{
			&t.ID, &t.UserID, &t.Title, &t.Content, &t.ImagePath,
			&t.Visibility, &groupID,
			&t.CreatedAt, &updatedAt,
			&t.OwnerUsername,
			&t.UpvoteCount, &t.DownvoteCount, &t.VoteScore,
		}
		if includeUserVote {
			var userVote sql.NullInt32
			scanArgs = append(scanArgs, &userVote)
			if err := rows.Scan(scanArgs...); err != nil {
				return nil, fmt.Errorf("scan topic: %w", err)
			}
			if userVote.Valid {
				v := int(userVote.Int32)
				t.UserVote = &v
			}
		} else {
			if err := rows.Scan(scanArgs...); err != nil {
				return nil, fmt.Errorf("scan topic: %w", err)
			}
		}
		t.UpdatedAt = database.ResolveTime(updatedAt, t.CreatedAt)
		if groupID.Valid {
			t.GroupID = &groupID.String
		}
		topics = append(topics, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iter: %w", err)
	}
	return topics, nil
}

func (s *SQLiteStore) CastVote(ctx context.Context, userID string, topicID int, reactionType int) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO votes (user_id, topic_id, comment_id, reaction_type)
		 VALUES (?, ?, NULL, ?)
		 ON CONFLICT (user_id, topic_id) DO UPDATE SET reaction_type = EXCLUDED.reaction_type, created_at = CURRENT_TIMESTAMP`,
		userID, topicID, reactionType)
	if err != nil {
		return fmt.Errorf("cast vote: %w", err)
	}
	return nil
}

func (s *SQLiteStore) DeleteVote(ctx context.Context, userID string, topicID int) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM votes WHERE user_id = ? AND topic_id = ? AND comment_id IS NULL`,
		userID, topicID)
	if err != nil {
		return fmt.Errorf("delete vote: %w", err)
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

func (s *SQLiteStore) GetVoteCounts(ctx context.Context, topicID int) (*topic.VoteCounts, error) {
	var vc topic.VoteCounts
	err := s.db.QueryRowContext(ctx,
		`SELECT
			COUNT(CASE WHEN reaction_type = 1 THEN 1 END),
			COUNT(CASE WHEN reaction_type = -1 THEN 1 END),
			COUNT(CASE WHEN reaction_type = 1 THEN 1 END) - COUNT(CASE WHEN reaction_type = -1 THEN 1 END)
		FROM votes WHERE topic_id = ? AND comment_id IS NULL`, topicID).
		Scan(&vc.Upvotes, &vc.Downvotes, &vc.Score)
	if err != nil {
		return nil, fmt.Errorf("get vote counts: %w", err)
	}
	return &vc, nil
}

func (s *SQLiteStore) GetPostCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM topics WHERE user_id = ?`, userID).Scan(&count)
	return count, err
}

func (s *SQLiteStore) GetVoteCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM votes WHERE user_id = ?`, userID).Scan(&count)
	return count, err
}

func (s *SQLiteStore) getAllowedUsers(ctx context.Context, topicID int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT user_id FROM topic_allowed_users WHERE topic_id = ?`, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func sanitizeOrderBy(orderBy string) string {
	whitelist := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"title":      true,
		"vote_score": true,
	}
	if !whitelist[orderBy] {
		return "t.created_at"
	}
	if orderBy == "vote_score" {
		return "vc.score"
	}
	return "t." + orderBy
}

func sanitizeOrder(order string) string {
	if strings.ToUpper(order) == "ASC" {
		return "ASC"
	}
	return "DESC"
}
