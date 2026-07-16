package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"social-network/internal/platform/database"
	"social-network/internal/user"
)

type SQLiteStore struct {
	db database.DB
}

func NewSQLiteStore(db database.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func scanUser(row interface{ Scan(dest ...any) error }) (*user.User, error) {
	var u user.User
	var firstName, lastName sql.NullString
	var avatarURL sql.NullString
	var dateOfBirth sql.NullTime
	var aboutMe sql.NullString
	var isPrivate sql.NullBool

	err := row.Scan(
		&u.ID,
		&u.Nickname,
		&u.Email,
		&u.PasswordHash,
		&firstName,
		&lastName,
		&u.CreatedAt,
		&avatarURL,
		&dateOfBirth,
		&aboutMe,
		&isPrivate,
	)
	if err != nil {
		return nil, err
	}

	u.FirstName = firstName.String
	u.LastName = lastName.String
	u.AvatarPath = avatarURL.String
	if dateOfBirth.Valid {
		u.DateOfBirth = dateOfBirth.Time
	}
	u.AboutMe = aboutMe.String
	u.IsPrivate = isPrivate.Bool

	return &u, nil
}

const userColumns = `id, username, email, password_hash, first_name, last_name,
	created_at, avatar_url, date_of_birth, about_me, is_private`

func (s *SQLiteStore) Create(ctx context.Context, u *user.User) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO users (id, username, email, password_hash, first_name, last_name, created_at, avatar_url, date_of_birth, about_me, is_private)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Nickname, u.Email, u.PasswordHash, u.FirstName, u.LastName,
		u.CreatedAt, nullString(u.AvatarPath), nullTime(u.DateOfBirth), nullString(u.AboutMe), u.IsPrivate,
	)
	return err
}

func (s *SQLiteStore) GetByID(ctx context.Context, id string) (*user.User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = ?`, id)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user %s: %w", id, user.ErrUserNotFound)
	}
	return u, err
}

func (s *SQLiteStore) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = ?`, email)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user %s: %w", email, user.ErrUserNotFound)
	}
	return u, err
}

func (s *SQLiteStore) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE username = ?`, username)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user %s: %w", username, user.ErrUserNotFound)
	}
	return u, err
}

func (s *SQLiteStore) Update(ctx context.Context, u *user.User) error {
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE users SET email = ?, password_hash = ?, first_name = ?, last_name = ?,
		avatar_url = ?, date_of_birth = ?, about_me = ?, is_private = ?, updated_at = ?
		WHERE id = ?`,
		u.Email, u.PasswordHash, u.FirstName, u.LastName,
		nullString(u.AvatarPath), nullTime(u.DateOfBirth), nullString(u.AboutMe), u.IsPrivate,
		time.Now(), u.ID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user %s: %w", u.ID, user.ErrUserNotFound)
	}
	return nil
}

func (s *SQLiteStore) TogglePrivacy(ctx context.Context, id string, isPrivate bool) error {
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE users SET is_private = ?, updated_at = ? WHERE id = ?`,
		isPrivate, time.Now(), id,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user %s: %w", id, user.ErrUserNotFound)
	}
	return nil
}

func (s *SQLiteStore) ListAll(ctx context.Context) ([]user.User, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+userColumns+` FROM users ORDER BY username ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []user.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}
