package users

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"social-network/internal/domain/user"

	_ "github.com/mattn/go-sqlite3"
)

func TestUserRegister_UsesPreparedStatement(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	repo := Repo{DB: db}

	u := &user.User{
		ID:        "user-1",
		Nickname:  "testuser",
		Email:     "test@example.com",
		Password:  "hashed_password",
		FirstName: "John",
		LastName:  "Doe",
		Age:       30,
		Gender:    "male",
	}

	err := repo.UserRegister(context.Background(), u)
	if err != nil {
		t.Fatalf("UserRegister() error = %v", err)
	}

	var (
		gotID       string
		gotUsername string
		gotEmail    string
		gotPassword string
	)
	err = db.QueryRow(
		`SELECT id, username, email, password_hash FROM users WHERE id = ?`, "user-1",
	).Scan(&gotID, &gotUsername, &gotEmail, &gotPassword)
	if err != nil {
		t.Fatalf("failed to query user: %v", err)
	}
	if gotID != "user-1" {
		t.Errorf("id = %q, want %q", gotID, "user-1")
	}
	if gotUsername != "testuser" {
		t.Errorf("username = %q, want %q", gotUsername, "testuser")
	}
	if gotEmail != "test@example.com" {
		t.Errorf("email = %q, want %q", gotEmail, "test@example.com")
	}
	if gotPassword != "hashed_password" {
		t.Errorf("password_hash = %q, want %q", gotPassword, "hashed_password")
	}
}

func TestUserRegister_DuplicateEmail(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	repo := Repo{DB: db}

	u1 := &user.User{
		ID: "user-1", Nickname: "alice", Email: "alice@example.com",
		Password: "hash", FirstName: "Alice", LastName: "A", Age: 25, Gender: "female",
	}
	u2 := &user.User{
		ID: "user-2", Nickname: "bob", Email: "alice@example.com",
		Password: "hash", FirstName: "Bob", LastName: "B", Age: 30, Gender: "male",
	}

	err := repo.UserRegister(context.Background(), u1)
	if err != nil {
		t.Fatalf("first UserRegister() error = %v", err)
	}

	err = repo.UserRegister(context.Background(), u2)
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Errorf("second UserRegister() error = %v, want %v", err, ErrDuplicateEmail)
	}
}

func setupDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE CHECK(email LIKE '%_@__%.__%'),
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT,
			age INTEGER,
			gender TEXT,
			first_name TEXT,
			last_name TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			avatar_url TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// productionSchema mirrors db/migrations/000001_initial_schema.up.sql exactly.
const productionSchema = `
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    avatar_url TEXT,
    username TEXT UNIQUE,
    date_of_birth DATE,
    about_me TEXT,
    is_private BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);`

// TestGetAll_AgainstProductionSchema reproduces a bug where Repo.GetAll
// selected legacy age/gender columns that do not exist in the production
// schema, breaking GET /api/v1/chat/users.
func TestGetAll_AgainstProductionSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(productionSchema); err != nil {
		t.Fatalf("failed to create production schema: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash, first_name, last_name)
		VALUES ('user-1', 'alice', 'alice@example.com', 'hash', 'Alice', 'A')`)
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	repo := Repo{DB: db}
	users, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() against production schema error = %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("len(users) = %d, want 1", len(users))
	}
	if users[0].ID != "user-1" || users[0].Nickname != "alice" {
		t.Errorf("got %+v, want user-1/alice", users[0])
	}
}
