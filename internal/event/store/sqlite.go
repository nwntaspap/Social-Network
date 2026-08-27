package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"social-network/internal/event"
	"social-network/internal/platform/database"
)

type SQLiteStore struct {
	db database.DB
}

func NewSQLiteStore(db database.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) CreateEvent(ctx context.Context, e *event.Event) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO events (id, group_id, creator_id, title, description, event_time)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		e.ID, e.GroupID, e.CreatorID, e.Title, e.Description, e.ScheduledTime,
	)
	return err
}

func (s *SQLiteStore) GetEvent(ctx context.Context, eventID string) (*event.Event, error) {
	var e event.Event
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, group_id, creator_id, title, description, event_time, created_at
		 FROM events WHERE id = ?`, eventID,
	).Scan(&e.ID, &e.GroupID, &e.CreatorID, &e.Title, &e.Description, &e.ScheduledTime, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, event.ErrEventNotFound
		}
		return nil, fmt.Errorf("get event: %w", err)
	}
	return &e, nil
}

func (s *SQLiteStore) UpdateEvent(ctx context.Context, e *event.Event) error {
	res, err := s.db.ExecContext(
		ctx,
		`UPDATE events SET title = ?, description = ?, event_time = ? WHERE id = ?`,
		e.Title, e.Description, e.ScheduledTime, e.ID,
	)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update event rows affected: %w", err)
	}
	if affected == 0 {
		return event.ErrEventNotFound
	}
	return nil
}

func (s *SQLiteStore) DeleteEvent(ctx context.Context, eventID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE id = ?`, eventID)
	return err
}

func (s *SQLiteStore) ListGroupEvents(ctx context.Context, groupID, cursor string, size int) ([]event.Event, string, error) {
	var rows *sql.Rows
	var err error

	if cursor == "" {
		rows, err = s.db.QueryContext(
			ctx,
			`SELECT id, group_id, creator_id, title, description, event_time, created_at
			 FROM events WHERE group_id = ?
			 ORDER BY created_at DESC
			 LIMIT ?`,
			groupID, size+1,
		)
	} else {
		rows, err = s.db.QueryContext(
			ctx,
			`SELECT id, group_id, creator_id, title, description, event_time, created_at
			 FROM events WHERE group_id = ? AND created_at < ?
			 ORDER BY created_at DESC
			 LIMIT ?`,
			groupID, cursor, size+1,
		)
	}
	if err != nil {
		return nil, "", fmt.Errorf("list group events: %w", err)
	}
	defer rows.Close()

	var events []event.Event
	for rows.Next() {
		var e event.Event
		if err := rows.Scan(&e.ID, &e.GroupID, &e.CreatorID, &e.Title, &e.Description, &e.ScheduledTime, &e.CreatedAt); err != nil {
			return nil, "", fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("list group events rows: %w", err)
	}

	var nextCursor string
	if len(events) > size {
		events = events[:size]
		nextCursor = events[len(events)-1].CreatedAt.Format("2006-01-02 15:04:05")
	}

	return events, nextCursor, nil
}

func (s *SQLiteStore) CreateOptions(ctx context.Context, opts []event.Option) error {
	if len(opts) == 0 {
		return nil
	}
	query := `INSERT INTO event_options (id, event_id, name) VALUES `
	var args []any
	var placeholders []string
	for _, o := range opts {
		placeholders = append(placeholders, "(?, ?, ?)")
		args = append(args, o.ID, o.EventID, o.Label)
	}
	query += strings.Join(placeholders, ", ")
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *SQLiteStore) GetOptionsByEvent(ctx context.Context, eventID string) ([]event.Option, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, event_id, name FROM event_options WHERE event_id = ?`, eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("get options by event: %w", err)
	}
	defer rows.Close()

	var opts []event.Option
	for rows.Next() {
		var o event.Option
		if err := rows.Scan(&o.ID, &o.EventID, &o.Label); err != nil {
			return nil, fmt.Errorf("scan option: %w", err)
		}
		opts = append(opts, o)
	}
	return opts, rows.Err()
}

func (s *SQLiteStore) GetOption(ctx context.Context, optionID string) (*event.Option, error) {
	var o event.Option
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, event_id, name FROM event_options WHERE id = ?`, optionID,
	).Scan(&o.ID, &o.EventID, &o.Label)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, event.ErrOptionNotFound
		}
		return nil, fmt.Errorf("get option: %w", err)
	}
	return &o, nil
}

func (s *SQLiteStore) UpsertRSVP(ctx context.Context, rsvp *event.RSVP) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO event_rsvps (event_id, user_id, option_id)
		 VALUES (?, ?, ?)
		 ON CONFLICT(event_id, user_id) DO UPDATE SET option_id = excluded.option_id, updated_at = CURRENT_TIMESTAMP`,
		rsvp.EventID, rsvp.UserID, rsvp.OptionID,
	)
	return err
}

func (s *SQLiteStore) GetRSVPsByEvent(ctx context.Context, eventID string) ([]event.RSVP, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT event_id, user_id, option_id, updated_at FROM event_rsvps WHERE event_id = ?`, eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("get rsvps by event: %w", err)
	}
	defer rows.Close()

	var rsvps []event.RSVP
	for rows.Next() {
		var r event.RSVP
		if err := rows.Scan(&r.EventID, &r.UserID, &r.OptionID, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan rsvp: %w", err)
		}
		rsvps = append(rsvps, r)
	}
	return rsvps, rows.Err()
}

func (s *SQLiteStore) GetUserRSVP(ctx context.Context, eventID, userID string) (*event.RSVP, error) {
	var r event.RSVP
	err := s.db.QueryRowContext(
		ctx,
		`SELECT event_id, user_id, option_id, updated_at
		 FROM event_rsvps WHERE event_id = ? AND user_id = ?`,
		eventID, userID,
	).Scan(&r.EventID, &r.UserID, &r.OptionID, &r.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, event.ErrRSVPNotFound
		}
		return nil, fmt.Errorf("get user rsvp: %w", err)
	}
	return &r, nil
}
