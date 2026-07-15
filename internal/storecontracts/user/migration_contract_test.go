package user_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	domainuser "social-network/internal/domain/user"
	legacystore "social-network/internal/infra/storage/sqlite/users"
	"social-network/internal/platform/database"
	"social-network/internal/user"
	newstore "social-network/internal/user/store"
)

// Legacy schema matches production: has age/gender, no date_of_birth/about_me/is_private.
const legacySchema = `
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
    avatar_url TEXT
);`

// New schema matches internal/user/store: has date_of_birth/about_me/is_private, keeps age/gender for compat.
const newSchema = `
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

func setupLegacyDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.ExecContext(context.Background(), legacySchema); err != nil {
		t.Fatalf("create legacy schema: %v", err)
	}
	return db
}

func setupNewDB(t *testing.T) database.DB {
	t.Helper()
	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("open new db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.ExecContext(context.Background(), newSchema); err != nil {
		t.Fatalf("create new schema: %v", err)
	}
	return db
}

// legacyAdapter wraps the legacy domain/user.Repository to accept new user.User as input.
type legacyAdapter struct {
	repo *legacystore.Repo
}

func (a *legacyAdapter) Create(ctx context.Context, u *user.User) error {
	return a.repo.UserRegister(ctx, &domainuser.User{
		ID:        u.ID,
		Nickname:  u.Nickname,
		Email:     u.Email,
		Password:  u.PasswordHash,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	})
}

func (a *legacyAdapter) GetByID(_ context.Context, _ string) (*user.User, error) {
	return nil, errors.New("legacy store has no GetByID")
}

func (a *legacyAdapter) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	du, err := a.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &user.User{
		ID:           du.ID,
		Nickname:     du.Nickname,
		PasswordHash: du.Password,
	}, nil
}

func (a *legacyAdapter) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	du, err := a.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &user.User{
		ID:           du.ID,
		Nickname:     du.Nickname,
		Email:        du.Email,
		PasswordHash: du.Password,
		CreatedAt:    du.CreatedAt,
	}, nil
}

func (a *legacyAdapter) Update(_ context.Context, _ *user.User) error {
	return errors.New("legacy store has no Update")
}

func (a *legacyAdapter) TogglePrivacy(_ context.Context, _ string, _ bool) error {
	return errors.New("legacy store has no TogglePrivacy")
}

func (a *legacyAdapter) ListAll(ctx context.Context) ([]user.User, error) {
	domainUsers, err := a.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]user.User, 0, len(domainUsers))
	for _, du := range domainUsers {
		users = append(users, user.User{
			ID:        du.ID,
			Nickname:  du.Nickname,
			Email:     du.Email,
			FirstName: du.FirstName,
			LastName:  du.LastName,
			CreatedAt: du.CreatedAt,
		})
	}
	return users, nil
}

// --- shared test data ---

var baseData = struct {
	id, nickname, email, password, firstName, lastName string
	createdAt                                          time.Time
}{
	id:        "user-1",
	nickname:  "alice",
	email:     "alice@example.com",
	password:  "hashed_password",
	firstName: "Alice",
	lastName:  "Smith",
	createdAt: time.Now(),
}

// --- contract tests ---

func TestMigrationContract_CreateAndGetByEmail(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		db := setupLegacyDB(t)
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		u := &user.User{
			ID:           baseData.id,
			Email:        baseData.email,
			PasswordHash: baseData.password,
			Nickname:     baseData.nickname,
			FirstName:    baseData.firstName,
			LastName:     baseData.lastName,
			CreatedAt:    baseData.createdAt,
		}
		if err := adapter.Create(context.Background(), u); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := adapter.GetByEmail(context.Background(), baseData.email)
		if err != nil {
			t.Fatalf("GetByEmail() error = %v", err)
		}
		if got.ID != baseData.id {
			t.Errorf("ID = %q, want %q", got.ID, baseData.id)
		}
		if got.Nickname != baseData.nickname {
			t.Errorf("Nickname = %q, want %q", got.Nickname, baseData.nickname)
		}
		if got.PasswordHash != baseData.password {
			t.Errorf("PasswordHash = %q, want %q", got.PasswordHash, baseData.password)
		}
	})

	t.Run("new", func(t *testing.T) {
		db := setupNewDB(t)
		s := newstore.NewSQLiteStore(db)

		u := &user.User{
			ID:           baseData.id,
			Email:        baseData.email,
			PasswordHash: baseData.password,
			Nickname:     baseData.nickname,
			FirstName:    baseData.firstName,
			LastName:     baseData.lastName,
			CreatedAt:    baseData.createdAt,
		}
		if err := s.Create(context.Background(), u); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := s.GetByEmail(context.Background(), baseData.email)
		if err != nil {
			t.Fatalf("GetByEmail() error = %v", err)
		}
		if got.ID != baseData.id {
			t.Errorf("ID = %q, want %q", got.ID, baseData.id)
		}
		if got.Nickname != baseData.nickname {
			t.Errorf("Nickname = %q, want %q", got.Nickname, baseData.nickname)
		}
		if got.PasswordHash != baseData.password {
			t.Errorf("PasswordHash = %q, want %q", got.PasswordHash, baseData.password)
		}
	})
}

func TestMigrationContract_CreateAndGetByUsername(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		db := setupLegacyDB(t)
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		u := &user.User{
			ID:           baseData.id,
			Email:        baseData.email,
			PasswordHash: baseData.password,
			Nickname:     baseData.nickname,
			FirstName:    baseData.firstName,
			LastName:     baseData.lastName,
			CreatedAt:    baseData.createdAt,
		}
		if err := adapter.Create(context.Background(), u); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := adapter.GetByUsername(context.Background(), baseData.nickname)
		if err != nil {
			t.Fatalf("GetByUsername() error = %v", err)
		}
		if got.ID != baseData.id {
			t.Errorf("ID = %q, want %q", got.ID, baseData.id)
		}
		if got.Email != baseData.email {
			t.Errorf("Email = %q, want %q", got.Email, baseData.email)
		}
	})

	t.Run("new", func(t *testing.T) {
		db := setupNewDB(t)
		s := newstore.NewSQLiteStore(db)

		u := &user.User{
			ID:           baseData.id,
			Email:        baseData.email,
			PasswordHash: baseData.password,
			Nickname:     baseData.nickname,
			FirstName:    baseData.firstName,
			LastName:     baseData.lastName,
			CreatedAt:    baseData.createdAt,
		}
		if err := s.Create(context.Background(), u); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := s.GetByUsername(context.Background(), baseData.nickname)
		if err != nil {
			t.Fatalf("GetByUsername() error = %v", err)
		}
		if got.ID != baseData.id {
			t.Errorf("ID = %q, want %q", got.ID, baseData.id)
		}
		if got.Email != baseData.email {
			t.Errorf("Email = %q, want %q", got.Email, baseData.email)
		}
	})
}

func TestMigrationContract_DuplicateEmail(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		db := setupLegacyDB(t)
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		u1 := &user.User{ID: "u1", Email: "dup@example.com", Nickname: "alice", PasswordHash: "h"}
		u2 := &user.User{ID: "u2", Email: "dup@example.com", Nickname: "bob", PasswordHash: "h"}

		if err := adapter.Create(context.Background(), u1); err != nil {
			t.Fatalf("first Create() error = %v", err)
		}
		err := adapter.Create(context.Background(), u2)
		if err == nil {
			t.Fatal("second Create() expected error, got nil")
		}
	})

	t.Run("new", func(t *testing.T) {
		db := setupNewDB(t)
		s := newstore.NewSQLiteStore(db)

		u1 := &user.User{ID: "u1", Email: "dup@example.com", Nickname: "alice", PasswordHash: "h"}
		u2 := &user.User{ID: "u2", Email: "dup@example.com", Nickname: "bob", PasswordHash: "h"}

		if err := s.Create(context.Background(), u1); err != nil {
			t.Fatalf("first Create() error = %v", err)
		}
		err := s.Create(context.Background(), u2)
		if err == nil {
			t.Fatal("second Create() expected error, got nil")
		}
	})
}

func TestMigrationContract_DuplicateUsername(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		db := setupLegacyDB(t)
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		u1 := &user.User{ID: "u1", Email: "a@example.com", Nickname: "same", PasswordHash: "h"}
		u2 := &user.User{ID: "u2", Email: "b@example.com", Nickname: "same", PasswordHash: "h"}

		if err := adapter.Create(context.Background(), u1); err != nil {
			t.Fatalf("first Create() error = %v", err)
		}
		err := adapter.Create(context.Background(), u2)
		if err == nil {
			t.Fatal("second Create() expected error, got nil")
		}
	})

	t.Run("new", func(t *testing.T) {
		db := setupNewDB(t)
		s := newstore.NewSQLiteStore(db)

		u1 := &user.User{ID: "u1", Email: "a@example.com", Nickname: "same", PasswordHash: "h"}
		u2 := &user.User{ID: "u2", Email: "b@example.com", Nickname: "same", PasswordHash: "h"}

		if err := s.Create(context.Background(), u1); err != nil {
			t.Fatalf("first Create() error = %v", err)
		}
		err := s.Create(context.Background(), u2)
		if err == nil {
			t.Fatal("second Create() expected error, got nil")
		}
	})
}

func TestMigrationContract_GetNotFound(t *testing.T) {
	t.Run("legacy_email", func(t *testing.T) {
		db := setupLegacyDB(t)
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		_, err := adapter.GetByEmail(context.Background(), "nobody@example.com")
		if err == nil {
			t.Fatal("GetByEmail() expected error, got nil")
		}
	})

	t.Run("new_email", func(t *testing.T) {
		db := setupNewDB(t)
		s := newstore.NewSQLiteStore(db)

		_, err := s.GetByEmail(context.Background(), "nobody@example.com")
		if err == nil {
			t.Fatal("GetByEmail() expected error, got nil")
		}
	})

	t.Run("legacy_username", func(t *testing.T) {
		db := setupLegacyDB(t)
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		_, err := adapter.GetByUsername(context.Background(), "nobody")
		if err == nil {
			t.Fatal("GetByUsername() expected error, got nil")
		}
	})

	t.Run("new_username", func(t *testing.T) {
		db := setupNewDB(t)
		s := newstore.NewSQLiteStore(db)

		_, err := s.GetByUsername(context.Background(), "nobody")
		if err == nil {
			t.Fatal("GetByUsername() expected error, got nil")
		}
	})
}

func TestMigrationContract_ListAllSorted(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		db := setupLegacyDB(t)
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		if err := adapter.Create(context.Background(), &user.User{ID: "u1", Email: "b@example.com", Nickname: "bob", PasswordHash: "h"}); err != nil {
			t.Fatalf("Create(u1) error = %v", err)
		}
		if err := adapter.Create(context.Background(), &user.User{ID: "u2", Email: "a@example.com", Nickname: "alice", PasswordHash: "h"}); err != nil {
			t.Fatalf("Create(u2) error = %v", err)
		}

		users, err := adapter.ListAll(context.Background())
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}
		if len(users) != 2 {
			t.Fatalf("len(users) = %d, want 2", len(users))
		}
		if users[0].Nickname != "alice" {
			t.Errorf("users[0].Nickname = %q, want %q (sorted by username)", users[0].Nickname, "alice")
		}
		if users[1].Nickname != "bob" {
			t.Errorf("users[1].Nickname = %q, want %q", users[1].Nickname, "bob")
		}
	})

	t.Run("new", func(t *testing.T) {
		db := setupNewDB(t)
		s := newstore.NewSQLiteStore(db)

		if err := s.Create(context.Background(), &user.User{ID: "u1", Email: "b@example.com", Nickname: "bob", PasswordHash: "h"}); err != nil {
			t.Fatalf("Create(u1) error = %v", err)
		}
		if err := s.Create(context.Background(), &user.User{ID: "u2", Email: "a@example.com", Nickname: "alice", PasswordHash: "h"}); err != nil {
			t.Fatalf("Create(u2) error = %v", err)
		}

		users, err := s.ListAll(context.Background())
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}
		if len(users) != 2 {
			t.Fatalf("len(users) = %d, want 2", len(users))
		}
		if users[0].Nickname != "alice" {
			t.Errorf("users[0].Nickname = %q, want %q (sorted by username)", users[0].Nickname, "alice")
		}
		if users[1].Nickname != "bob" {
			t.Errorf("users[1].Nickname = %q, want %q", users[1].Nickname, "bob")
		}
	})
}
