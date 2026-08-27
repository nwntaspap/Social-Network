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
);
CREATE TABLE IF NOT EXISTS follows (
    follower_id TEXT NOT NULL,
    followee_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(follower_id, followee_id),
    FOREIGN KEY(follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(followee_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS follow_requests (
    follower_id TEXT NOT NULL,
    followee_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(follower_id, followee_id),
    FOREIGN KEY(follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(followee_id) REFERENCES users(id) ON DELETE CASCADE
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

func TestCreate_GenderRoundTrip(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	if err := s.Create(ctx, &user.User{
		ID: "u1", Email: "g@example.com", Nickname: "gendered",
		PasswordHash: "h", Gender: "female", CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := s.GetByID(ctx, "u1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Gender != "female" {
		t.Errorf("Gender = %q, want %q", got.Gender, "female")
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

// Regression: OAuth signups insert users without dob/gender/about (NULLs).
// GetByID must scan them without error and yield zero-value fields.
func TestGetByID_OAuthShapedUser(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, username, email, password_hash, first_name, last_name, avatar_url)
		 VALUES ('oauth1', 'ghuser', 'gh@example.com', '', 'Git', 'Hub', '/img.png')`)
	if err != nil {
		t.Fatalf("seed oauth-shaped user: %v", err)
	}

	got, err := s.GetByID(ctx, "oauth1")
	if err != nil {
		t.Fatalf("GetByID() error = %v (NULL profile columns must not break scanning)", err)
	}
	if got.Email != "gh@example.com" {
		t.Errorf("Email = %q, want %q", got.Email, "gh@example.com")
	}
	if !got.DateOfBirth.IsZero() {
		t.Errorf("DateOfBirth = %v, want zero", got.DateOfBirth)
	}
	if got.Gender != "" || got.AboutMe != "" {
		t.Errorf("Gender/AboutMe = %q/%q, want empty", got.Gender, got.AboutMe)
	}
	if got.IsPrivate {
		t.Error("IsPrivate = true, want false (OAuth users are public by default)")
	}
}

func TestUpdate_PersistsDOBAndGender(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()
	dob := time.Date(2000, 5, 17, 0, 0, 0, 0, time.UTC)

	seedUser(t, s, &user.User{ID: "u1", Email: "u@example.com", Nickname: "u", PasswordHash: "h", CreatedAt: time.Now()})

	err := s.Update(ctx, &user.User{
		ID: "u1", Email: "u@example.com", Nickname: "u",
		DateOfBirth: dob, Gender: "other",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := s.GetByID(ctx, "u1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if !got.DateOfBirth.Equal(dob) {
		t.Errorf("DateOfBirth = %v, want %v", got.DateOfBirth, dob)
	}
	if got.Gender != "other" {
		t.Errorf("Gender = %q, want %q", got.Gender, "other")
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

func TestSearchUsers_MatchesColumns(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	seedUser(t, s, &user.User{ID: "u1", Email: "alice@example.com", Nickname: "alice", FirstName: "Alice", LastName: "Smith", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u2", Email: "bob@example.com", Nickname: "bobby", FirstName: "Bob", LastName: "Alice", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u3", Email: "carol@example.com", Nickname: "carol", FirstName: "Carol", LastName: "Jones", PasswordHash: "h", CreatedAt: time.Now()})

	users, err := s.SearchUsers(ctx, "alice", 10, 0)
	if err != nil {
		t.Fatalf("SearchUsers(alice) error = %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("SearchUsers(alice) returned %d users, want 2", len(users))
	}

	users, err = s.SearchUsers(ctx, "carol", 10, 0)
	if err != nil {
		t.Fatalf("SearchUsers(carol) error = %v", err)
	}
	if len(users) != 1 || users[0].ID != "u3" {
		t.Fatalf("SearchUsers(carol) = %+v, want [u3]", users)
	}

	users, err = s.SearchUsers(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("SearchUsers(empty) error = %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("SearchUsers(empty) returned %d users, want 3", len(users))
	}
	if users[0].Nickname != "alice" {
		t.Errorf("users[0].Nickname = %q, want %q (sorted)", users[0].Nickname, "alice")
	}
}

func TestSearchUsers_Paginates(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	seedUser(t, s, &user.User{ID: "u1", Email: "a@example.com", Nickname: "alice", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u2", Email: "b@example.com", Nickname: "bob", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u3", Email: "c@example.com", Nickname: "carol", PasswordHash: "h", CreatedAt: time.Now()})

	users, err := s.SearchUsers(ctx, "", 2, 2)
	if err != nil {
		t.Fatalf("SearchUsers(page2) error = %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("SearchUsers(page2) returned %d users, want 1", len(users))
	}
	if users[0].Nickname != "carol" {
		t.Errorf("users[0].Nickname = %q, want %q", users[0].Nickname, "carol")
	}
}

func TestCountUsers_MatchesFilter(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	seedUser(t, s, &user.User{ID: "u1", Email: "alice@example.com", Nickname: "alice", FirstName: "Alice", LastName: "Smith", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u2", Email: "bob@example.com", Nickname: "bobby", FirstName: "Bob", LastName: "Alice", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u3", Email: "carol@example.com", Nickname: "carol", FirstName: "Carol", LastName: "Jones", PasswordHash: "h", CreatedAt: time.Now()})

	count, err := s.CountUsers(ctx, "alice")
	if err != nil {
		t.Fatalf("CountUsers(alice) error = %v", err)
	}
	if count != 2 {
		t.Errorf("CountUsers(alice) = %d, want 2", count)
	}

	count, err = s.CountUsers(ctx, "nomatch")
	if err != nil {
		t.Fatalf("CountUsers(nomatch) error = %v", err)
	}
	if count != 0 {
		t.Errorf("CountUsers(nomatch) = %d, want 0", count)
	}

	count, err = s.CountUsers(ctx, "")
	if err != nil {
		t.Fatalf("CountUsers(empty) error = %v", err)
	}
	if count != 3 {
		t.Errorf("CountUsers(empty) = %d, want 3", count)
	}
}

func TestSearchUsersExcluding_ExcludesSelfAndFollowed(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	seedUser(t, s, &user.User{ID: "u1", Email: "alice@example.com", Nickname: "alice", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u2", Email: "bob@example.com", Nickname: "bob", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u3", Email: "carol@example.com", Nickname: "carol", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u4", Email: "dave@example.com", Nickname: "dave", PasswordHash: "h", CreatedAt: time.Now()})

	if _, err := s.db.ExecContext(ctx, `INSERT INTO follows (follower_id, followee_id) VALUES (?, ?)`, "u1", "u2"); err != nil {
		t.Fatalf("seed follow: %v", err)
	}

	users, err := s.SearchUsersExcluding(ctx, "", "u1", 10, 0)
	if err != nil {
		t.Fatalf("SearchUsersExcluding() error = %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("SearchUsersExcluding() returned %d users, want 2", len(users))
	}
	if users[0].Nickname != "carol" || users[1].Nickname != "dave" {
		t.Errorf("SearchUsersExcluding() = %+v, want [carol dave]", users)
	}

	count, err := s.CountUsersExcluding(ctx, "", "u1")
	if err != nil {
		t.Fatalf("CountUsersExcluding() error = %v", err)
	}
	if count != 2 {
		t.Errorf("CountUsersExcluding() = %d, want 2", count)
	}

	searched, err := s.SearchUsersExcluding(ctx, "bob", "u1", 10, 0)
	if err != nil {
		t.Fatalf("SearchUsersExcluding(bob) error = %v", err)
	}
	if len(searched) != 0 {
		t.Errorf("SearchUsersExcluding(bob) = %+v, want empty (followed user)", searched)
	}
}

func TestSearchUsersExcluding_ExcludesPendingFollowRequests(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()
	seedUser(t, s, &user.User{ID: "u1", Email: "alice@test.com", Nickname: "a", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u2", Email: "bob@test.com", Nickname: "b", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u3", Email: "carol@test.com", Nickname: "c", PasswordHash: "h", CreatedAt: time.Now()})
	if _, err := s.db.ExecContext(ctx, `INSERT INTO follow_requests (follower_id, followee_id) VALUES (?, ?)`, "u1", "u2"); err != nil {
		t.Fatalf("seed follow request: %v", err)
	}
	users, err := s.SearchUsersExcluding(ctx, "", "u1", 10, 0)
	if err != nil {
		t.Fatalf("SearchUsersExcluding() error = %v", err)
	}
	if len(users) != 1 || users[0].Nickname != "c" {
		t.Errorf("SearchUsersExcluding() = %+v, want [c]", users)
	}
	count, err := s.CountUsersExcluding(ctx, "", "u1")
	if err != nil {
		t.Fatalf("CountUsersExcluding() error = %v", err)
	}
	if count != 1 {
		t.Errorf("CountUsersExcluding() = %d, want 1", count)
	}
}

func TestSearchUsersExcluding_EmptyViewerBehavesLikeSearchUsers(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	seedUser(t, s, &user.User{ID: "u1", Email: "alice@example.com", Nickname: "alice", PasswordHash: "h", CreatedAt: time.Now()})
	seedUser(t, s, &user.User{ID: "u2", Email: "bob@example.com", Nickname: "bob", PasswordHash: "h", CreatedAt: time.Now()})

	users, err := s.SearchUsersExcluding(ctx, "", "", 10, 0)
	if err != nil {
		t.Fatalf("SearchUsersExcluding(empty viewer) error = %v", err)
	}
	if len(users) != 2 {
		t.Errorf("SearchUsersExcluding(empty viewer) = %+v, want 2 users", users)
	}

	count, err := s.CountUsersExcluding(ctx, "", "")
	if err != nil {
		t.Fatalf("CountUsersExcluding(empty viewer) error = %v", err)
	}
	if count != 2 {
		t.Errorf("CountUsersExcluding(empty viewer) = %d, want 2", count)
	}
}
