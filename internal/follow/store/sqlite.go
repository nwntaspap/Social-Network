package store

import (
	"context"

	"social-network/internal/follow"
	"social-network/internal/platform/database"
)

type SQLiteStore struct {
	db database.DB
}

func NewSQLiteStore(db database.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) CreateFollow(ctx context.Context, f *follow.Follow) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO follows (follower_id, followee_id) VALUES (?, ?)`,
		f.FollowerID, f.FolloweeID)
	return err
}

func (s *SQLiteStore) DeleteFollow(ctx context.Context, followerID, followeeID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM follows WHERE follower_id = ? AND followee_id = ?`,
		followerID, followeeID)
	return err
}

func (s *SQLiteStore) GetFollowers(ctx context.Context, userID string) ([]follow.Follow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT follower_id, followee_id FROM follows WHERE followee_id = ?`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []follow.Follow
	for rows.Next() {
		var f follow.Follow
		if err := rows.Scan(&f.FollowerID, &f.FolloweeID); err != nil {
			return nil, err
		}
		followers = append(followers, f)
	}
	return followers, rows.Err()
}

func (s *SQLiteStore) GetFollowing(ctx context.Context, userID string) ([]follow.Follow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT follower_id, followee_id FROM follows WHERE follower_id = ?`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var following []follow.Follow
	for rows.Next() {
		var f follow.Follow
		if err := rows.Scan(&f.FollowerID, &f.FolloweeID); err != nil {
			return nil, err
		}
		following = append(following, f)
	}
	return following, rows.Err()
}

func (s *SQLiteStore) CreateFollowRequest(ctx context.Context, req *follow.Request) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO follow_requests (follower_id, followee_id) VALUES (?, ?)`,
		req.FollowerID, req.FolloweeID)
	return err
}

func (s *SQLiteStore) DeleteFollowRequest(ctx context.Context, followerID, followeeID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM follow_requests WHERE follower_id = ? AND followee_id = ?`,
		followerID, followeeID)
	return err
}

func (s *SQLiteStore) GetPendingRequests(ctx context.Context, userID string) ([]follow.Request, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT follower_id, followee_id, created_at FROM follow_requests WHERE followee_id = ?`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []follow.Request
	for rows.Next() {
		var r follow.Request
		if err := rows.Scan(&r.FollowerID, &r.FolloweeID, &r.CreatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, rows.Err()
}

func (s *SQLiteStore) AreConnected(ctx context.Context, a, b string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = ? AND followee_id = ?)`
	var exists bool
	err := s.db.QueryRowContext(ctx, query, a, b).Scan(&exists)
	return exists, err
}

func (s *SQLiteStore) GetFollowerCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM follows WHERE followee_id = ?`, userID).Scan(&count)
	return count, err
}

func (s *SQLiteStore) GetFollowingCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM follows WHERE follower_id = ?`, userID).Scan(&count)
	return count, err
}
