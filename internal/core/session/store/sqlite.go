package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"social-network/internal/core/session"
	"social-network/internal/platform/database"

	"github.com/google/uuid"
)

type Store struct {
	db     database.DB
	now    func() time.Time
	uuid   func() string
	expiry time.Duration
}

type Option func(*Store)

func WithExpiry(d time.Duration) Option {
	return func(s *Store) {
		s.expiry = d
	}
}

func WithClock(now func() time.Time) Option {
	return func(s *Store) {
		s.now = now
	}
}

func NewSessionStore(db database.DB, opts ...Option) *Store {
	s := &Store{
		db:     db,
		now:    time.Now,
		uuid:   uuid.NewString,
		expiry: 24 * time.Hour,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Store) Create(ctx context.Context, userID string) (*session.Session, error) {
	token := s.uuid()
	expiresAt := s.now().UTC().Add(s.expiry)

	if _, err := s.db.ExecContext(
		ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expiresAt,
	); err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}

	if _, err := s.db.ExecContext(
		ctx,
		`DELETE FROM sessions WHERE user_id = ? AND token != ?`,
		userID, token,
	); err != nil {
		return nil, fmt.Errorf("cleanup old sessions: %w", err)
	}

	return &session.Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Store) Get(ctx context.Context, token string) (*session.Session, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT token, user_id, expires_at FROM sessions WHERE token = ?`,
		token,
	)

	var sess session.Session
	if err := row.Scan(&sess.Token, &sess.UserID, &sess.ExpiresAt); err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	if sess.ExpiresAt.Before(s.now().UTC()) {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
		return nil, errors.New("session expired")
	}

	return &sess, nil
}

func (s *Store) Revoke(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	return err
}
