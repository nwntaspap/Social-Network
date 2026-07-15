package store

import (
	"context"
	"testing"
	"time"

	"social-network/internal/platform/database"
	"social-network/internal/user"
)

const testSchema = `
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE CHECK(email LIKE '%_@__%.__%'),
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT,
    age INTEGER,
    gender TEXT,
    first_name TEXT,
    last_name TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    avatar_url TEXT,
    date_of_birth DATETIME,
    about_me TEXT,
    is_private BOOLEAN DEFAULT FALSE
);`

func setupStore(t *testing.T) *SQLiteStore {
	t.Helper()

	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(), testSchema); err != nil {
		t.Fatalf("create table: %v", err)
	}

	return NewSQLiteStore(db)
}

func seedUser(t *testing.T, s *SQLiteStore, u *user.User) {
	t.Helper()
	if err := s.Create(context.Background(), u); err != nil {
		t.Fatalf("seed user %s: %v", u.ID, err)
	}
}

func TestCreate(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	u := &user.User{
		ID:           "u1",
		Email:        "alice@example.com",
		PasswordHash: "hash123",
		FirstName:    "Alice",
		LastName:     "Smith",
		Nickname:     "alice",
		CreatedAt:    time.Now(),
	}

	if err := s.Create(ctx, u); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := s.GetByID(ctx, "u1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", got.Email, "alice@example.com")
	}
	if got.Nickname != "alice" {
		t.Errorf("Nickname = %q, want %q", got.Nickname, "alice")
	}
}

func TestCreate_DuplicateEmail(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	u1 := &user.User{ID: "u1", Email: "dup@example.com", Nickname: "user1", PasswordHash: "h"}
	u2 := &user.User{ID: "u2", Email: "dup@example.com", Nickname: "user2", PasswordHash: "h"}

	if err := s.Create(ctx, u1); err != nil {
		t.Fatalf("Create(u1) error = %v", err)
	}
	if err := s.Create(ctx, u2); err == nil {
		t.Fatal("Create(u2) expected error for duplicate email")
	}
}

func TestCreate_DuplicateUsername(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	u1 := &user.User{ID: "u1", Email: "a@example.com", Nickname: "dup", PasswordHash: "h"}
	u2 := &user.User{ID: "u2", Email: "b@example.com", Nickname: "dup", PasswordHash: "h"}

	if err := s.Create(ctx, u1); err != nil {
		t.Fatalf("Create(u1) error = %v", err)
	}
	if err := s.Create(ctx, u2); err == nil {
		t.Fatal("Create(u2) expected error for duplicate username")
	}
}

func TestGetByID(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	seedUser(t, s, &user.User{
		ID: "u1", Email: "a@example.com", Nickname: "alice",
		PasswordHash: "hash", FirstName: "A", LastName: "B",
		CreatedAt: time.Now(),
	})

	got, err := s.GetByID(ctx, "u1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != "u1" {
		t.Errorf("ID = %q, want %q", got.ID, "u1")
	}
	if got.FirstName != "A" {
		t.Errorf("FirstName = %q, want %q", got.FirstName, "A")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	s := setupStore(t)
	_, err := s.GetByID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetByID() expected error for nonexistent user")
	}
}

func TestGetByEmail(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	seedUser(t, s, &user.User{
		ID: "u1", Email: "find@example.com", Nickname: "alice",
		PasswordHash: "hash", CreatedAt: time.Now(),
	})

	got, err := s.GetByEmail(ctx, "find@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}
	if got.ID != "u1" {
		t.Errorf("ID = %q, want %q", got.ID, "u1")
	}
}

func TestGetByEmail_NotFound(t *testing.T) {
	s := setupStore(t)
	_, err := s.GetByEmail(context.Background(), "nope@example.com")
	if err == nil {
		t.Fatal("GetByEmail() expected error for nonexistent email")
	}
}

func TestGetByUsername(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	seedUser(t, s, &user.User{
		ID: "u1", Email: "a@example.com", Nickname: "findme",
		PasswordHash: "hash", CreatedAt: time.Now(),
	})

	got, err := s.GetByUsername(ctx, "findme")
	if err != nil {
		t.Fatalf("GetByUsername() error = %v", err)
	}
	if got.ID != "u1" {
		t.Errorf("ID = %q, want %q", got.ID, "u1")
	}
}

func TestGetByUsername_NotFound(t *testing.T) {
	s := setupStore(t)
	_, err := s.GetByUsername(context.Background(), "ghost")
	if err == nil {
		t.Fatal("GetByUsername() expected error for nonexistent username")
	}
}

func TestUpdate(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	seedUser(t, s, &user.User{
		ID: "u1", Email: "old@example.com", Nickname: "alice",
		PasswordHash: "hash", CreatedAt: time.Now(),
	})

	updated := &user.User{
		ID:           "u1",
		Email:        "new@example.com",
		PasswordHash: "newhash",
		FirstName:    "Updated",
		LastName:     "Name",
		Nickname:     "alice",
		AboutMe:      "new bio",
		CreatedAt:    time.Now(),
	}

	if err := s.Update(ctx, updated); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := s.GetByID(ctx, "u1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Email != "new@example.com" {
		t.Errorf("Email = %q, want %q", got.Email, "new@example.com")
	}
	if got.FirstName != "Updated" {
		t.Errorf("FirstName = %q, want %q", got.FirstName, "Updated")
	}
	if got.AboutMe != "new bio" {
		t.Errorf("AboutMe = %q, want %q", got.AboutMe, "new bio")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	s := setupStore(t)
	u := &user.User{ID: "nonexistent", Email: "x@x.com", Nickname: "x"}
	err := s.Update(context.Background(), u)
	if err == nil {
		t.Fatal("Update() expected error for nonexistent user")
	}
}

func TestTogglePrivacy(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	seedUser(t, s, &user.User{
		ID: "u1", Email: "a@example.com", Nickname: "alice",
		PasswordHash: "hash", CreatedAt: time.Now(),
	})

	if err := s.TogglePrivacy(ctx, "u1", true); err != nil {
		t.Fatalf("TogglePrivacy(true) error = %v", err)
	}
	got, err := s.GetByID(ctx, "u1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if !got.IsPrivate {
		t.Error("IsPrivate = false, want true after toggle")
	}

	err = s.TogglePrivacy(ctx, "u1", false)
	if err != nil {
		t.Fatalf("TogglePrivacy(false) error = %v", err)
	}
	got, err = s.GetByID(ctx, "u1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.IsPrivate {
		t.Error("IsPrivate = true, want false after second toggle")
	}
}

func TestTogglePrivacy_NotFound(t *testing.T) {
	s := setupStore(t)
	err := s.TogglePrivacy(context.Background(), "nonexistent", true)
	if err == nil {
		t.Fatal("TogglePrivacy() expected error for nonexistent user")
	}
}

func TestListAll(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	seedUser(t, s, &user.User{
		ID: "u1", Email: "a@example.com", Nickname: "charlie",
		PasswordHash: "h", CreatedAt: time.Now(),
	})
	seedUser(t, s, &user.User{
		ID: "u2", Email: "b@example.com", Nickname: "alice",
		PasswordHash: "h", CreatedAt: time.Now(),
	})
	seedUser(t, s, &user.User{
		ID: "u3", Email: "c@example.com", Nickname: "bob",
		PasswordHash: "h", CreatedAt: time.Now(),
	})

	users, err := s.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("ListAll() returned %d users, want 3", len(users))
	}
	if users[0].Nickname != "alice" {
		t.Errorf("users[0].Nickname = %q, want %q (sorted)", users[0].Nickname, "alice")
	}
	if users[1].Nickname != "bob" {
		t.Errorf("users[1].Nickname = %q, want %q (sorted)", users[1].Nickname, "bob")
	}
}

func TestListAll_Empty(t *testing.T) {
	s := setupStore(t)
	users, err := s.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("ListAll() returned %d users, want 0", len(users))
	}
}
