package store

import (
	"context"
	"database/sql"

	"social-network/internal/group"
)

func (s *SQLiteStore) CastPostVote(ctx context.Context, userID, postID string, reactionType int) error {
	var existingReaction sql.NullInt32
	checkQuery := `SELECT reaction_type FROM group_post_votes WHERE user_id = ? AND post_id = ?`
	err := s.db.QueryRowContext(ctx, checkQuery, userID, postID).Scan(&existingReaction)

	if err == nil && existingReaction.Valid && int(existingReaction.Int32) == reactionType {
		deleteQuery := `DELETE FROM group_post_votes WHERE user_id = ? AND post_id = ?`
		_, delErr := s.db.ExecContext(ctx, deleteQuery, userID, postID)
		return delErr
	}

	query := `
		INSERT INTO group_post_votes (user_id, post_id, reaction_type)
		VALUES (?, ?, ?)
		ON CONFLICT (user_id, post_id) DO UPDATE SET
			reaction_type = EXCLUDED.reaction_type,
			created_at = CURRENT_TIMESTAMP`
	_, err = s.db.ExecContext(ctx, query, userID, postID, reactionType)
	return err
}

func (s *SQLiteStore) GetPostVoteCounts(ctx context.Context, postID string) (*group.VoteCounts, error) {
	counts := &group.VoteCounts{}
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(CASE WHEN reaction_type = 1 THEN 1 END),
		        COUNT(CASE WHEN reaction_type = -1 THEN 1 END),
		        COUNT(CASE WHEN reaction_type = 1 THEN 1 END) - COUNT(CASE WHEN reaction_type = -1 THEN 1 END)
		 FROM group_post_votes WHERE post_id = ?`,
		postID).Scan(&counts.Upvotes, &counts.Downvotes, &counts.Score)
	if err != nil {
		return nil, err
	}
	return counts, nil
}
