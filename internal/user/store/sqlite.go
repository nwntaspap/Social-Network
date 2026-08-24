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
	var gender sql.NullString

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
		&gender,
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
	u.Gender = gender.String

	return &u, nil
}

const userColumns = `id, username, email, password_hash, first_name, last_name,
	created_at, avatar_url, date_of_birth, about_me, is_private, gender`

func (s *SQLiteStore) Create(ctx context.Context, u *user.User) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO users (id, username, email, password_hash, first_name, last_name, created_at, avatar_url, date_of_birth, about_me, is_private, gender)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Nickname, u.Email, u.PasswordHash, u.FirstName, u.LastName,
		u.CreatedAt, nullString(u.AvatarPath), nullTime(u.DateOfBirth), nullString(u.AboutMe), u.IsPrivate, u.Gender,
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

// IsPrivate reports whether a user's profile is private.
func (s *SQLiteStore) IsPrivate(ctx context.Context, userID string) (bool, error) {
	var isPrivate bool
	err := s.db.QueryRowContext(ctx,
		`SELECT is_private FROM users WHERE id = ?`, userID).Scan(&isPrivate)
	if errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("user %s: %w", userID, user.ErrUserNotFound)
	}
	return isPrivate, err
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
		avatar_url = ?, date_of_birth = ?, about_me = ?, is_private = ?, gender = ?, updated_at = ?
		WHERE id = ?`,
		u.Email, u.PasswordHash, u.FirstName, u.LastName,
		nullString(u.AvatarPath), nullTime(u.DateOfBirth), nullString(u.AboutMe), u.IsPrivate, u.Gender,
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

func userSearchFilter(query string) (string, []any) {
	if query == "" {
		return "", nil
	}
	pattern := "%" + query + "%"
	where := `WHERE (username LIKE ? OR first_name LIKE ? OR last_name LIKE ? OR email LIKE ?)`
	args := []any{pattern, pattern, pattern, pattern}
	return where, args
}

func excludeFollowed(where string, args []any, userID string) (string, []any) {
	if userID == "" {
		return where, args
	}
	clause := `id != ? AND id NOT IN (SELECT followee_id FROM follows WHERE follower_id = ?)`
	if where == "" {
		where = `WHERE ` + clause
	} else {
		where += ` AND ` + clause
	}
	args = append(args, userID, userID)
	return where, args
}

func (s *SQLiteStore) SearchUsers(ctx context.Context, query string, limit, offset int) ([]user.User, error) {
	where, args := userSearchFilter(query)
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+userColumns+` FROM users `+where+` ORDER BY username ASC LIMIT ? OFFSET ?`, args...)
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

func (s *SQLiteStore) CountUsers(ctx context.Context, query string) (int, error) {
	where, args := userSearchFilter(query)
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users `+where, args...).Scan(&count)
	return count, err
}

func (s *SQLiteStore) SearchUsersExcluding(ctx context.Context, query, excludeUserID string, limit, offset int) ([]user.User, error) {
	where, args := userSearchFilter(query)
	where, args = excludeFollowed(where, args, excludeUserID)
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+userColumns+` FROM users `+where+` ORDER BY username ASC LIMIT ? OFFSET ?`, args...)
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

func (s *SQLiteStore) CountUsersExcluding(ctx context.Context, query, excludeUserID string) (int, error) {
	where, args := userSearchFilter(query)
	where, args = excludeFollowed(where, args, excludeUserID)
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users `+where, args...).Scan(&count)
	return count, err
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
