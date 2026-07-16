package store

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
	"social-network/internal/platform/database"
)

const (
	schema = `
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL UNIQUE
		);
		CREATE TABLE topics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL REFERENCES users(id),
			title TEXT NOT NULL,
			content TEXT NOT NULL
		);
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL REFERENCES users(id),
			topic_id INTEGER NOT NULL REFERENCES topics(id),
			content TEXT NOT NULL,
			image_path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE votes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL REFERENCES users(id),
			topic_id INTEGER REFERENCES topics(id),
			comment_id INTEGER REFERENCES comments(id),
			reaction_type INTEGER NOT NULL CHECK(reaction_type IN (-1, 1)),
			UNIQUE (user_id, topic_id),
			UNIQUE (user_id, comment_id)
		);`
	seedData = `
		INSERT INTO users (id, email, username) VALUES ('u1', 'a@b.com', 'alice');
		INSERT INTO users (id, email, username) VALUES ('u2', 'c@d.com', 'bob');
		INSERT INTO topics (id, user_id, title, content) VALUES (1, 'u1', 't1', 'c1');`
)

func setupStore(t *testing.T) *SQLiteStore {
	t.Helper()
	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(), schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), seedData); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return NewSQLiteStore(db)
}

func mustCreate(t *testing.T, s *SQLiteStore, c *comment.Comment) {
	t.Helper()
	if err := s.CreateComment(context.Background(), c); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
}

func mustVote(t *testing.T, s *SQLiteStore, userID string, commentID int, reaction int) {
	t.Helper()
	if _, err := s.db.ExecContext(context.Background(),
		`INSERT INTO votes (user_id, comment_id, reaction_type) VALUES (?, ?, ?)`,
		userID, commentID, reaction); err != nil {
		t.Fatalf("insert vote: %v", err)
	}
}

func TestCreateComment(t *testing.T) {
	s := setupStore(t)
	c := &comment.Comment{UserID: "u1", TopicID: 1, Content: "hello", ImagePath: "/img.png"}
	mustCreate(t, s, c)
	if c.ID == 0 {
		t.Fatal("CreateComment did not set ID")
	}
}

func TestGetCommentByID(t *testing.T) {
	s := setupStore(t)
	created := &comment.Comment{UserID: "u1", TopicID: 1, Content: "hello"}
	mustCreate(t, s, created)

	got, err := s.GetCommentByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetCommentByID: %v", err)
	}
	if got.Content != "hello" {
		t.Errorf("Content = %q, want %q", got.Content, "hello")
	}
}

func TestGetCommentByID_NotFound(t *testing.T) {
	s := setupStore(t)
	_, err := s.GetCommentByID(context.Background(), 999)
	if !errors.Is(err, comment.ErrCommentNotFound) {
		t.Errorf("expected ErrCommentNotFound, got %v", err)
	}
}

func TestUpdateComment(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	c := &comment.Comment{UserID: "u1", TopicID: 1, Content: "old"}
	mustCreate(t, s, c)

	c.Content = "new"
	if err := s.UpdateComment(ctx, c); err != nil {
		t.Fatalf("UpdateComment: %v", err)
	}

	got, _ := s.GetCommentByID(ctx, c.ID)
	if got.Content != "new" {
		t.Errorf("Content = %q, want %q", got.Content, "new")
	}
}

func TestUpdateComment_NotFound(t *testing.T) {
	s := setupStore(t)
	err := s.UpdateComment(context.Background(), &comment.Comment{ID: 999, UserID: "u1", Content: "x"})
	if !errors.Is(err, comment.ErrCommentNotFound) {
		t.Errorf("expected ErrCommentNotFound, got %v", err)
	}
}

func TestDeleteComment(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	c := &comment.Comment{UserID: "u1", TopicID: 1, Content: "del"}
	mustCreate(t, s, c)

	if err := s.DeleteComment(ctx, "u1", c.ID); err != nil {
		t.Fatalf("DeleteComment: %v", err)
	}

	_, err := s.GetCommentByID(ctx, c.ID)
	if !errors.Is(err, comment.ErrCommentNotFound) {
		t.Errorf("expected ErrCommentNotFound after delete, got %v", err)
	}
}

func TestDeleteComment_WrongUser(t *testing.T) {
	s := setupStore(t)
	c := &comment.Comment{UserID: "u1", TopicID: 1, Content: "x"}
	mustCreate(t, s, c)

	err := s.DeleteComment(context.Background(), "u2", c.ID)
	if !errors.Is(err, comment.ErrCommentNotFound) {
		t.Errorf("expected ErrCommentNotFound for wrong user, got %v", err)
	}
}

func TestGetCommentsByTopicID(t *testing.T) {
	s := setupStore(t)
	mustCreate(t, s, &comment.Comment{UserID: "u1", TopicID: 1, Content: "a"})
	mustCreate(t, s, &comment.Comment{UserID: "u2", TopicID: 1, Content: "b"})
	mustCreate(t, s, &comment.Comment{UserID: "u1", TopicID: 1, Content: "c"})

	comments, err := s.GetCommentsByTopicID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetCommentsByTopicID: %v", err)
	}
	if len(comments) != 3 {
		t.Fatalf("got %d comments, want 3", len(comments))
	}
}

func TestGetCommentsByTopicID_Empty(t *testing.T) {
	s := setupStore(t)
	comments, err := s.GetCommentsByTopicID(context.Background(), 999)
	if err != nil {
		t.Fatalf("GetCommentsByTopicID: %v", err)
	}
	if len(comments) != 0 {
		t.Fatalf("got %d comments, want 0", len(comments))
	}
}

func TestCreateComment_WithImage(t *testing.T) {
	s := setupStore(t)
	c := &comment.Comment{UserID: "u1", TopicID: 1, Content: "pic", ImagePath: "/uploads/photo.jpg"}
	mustCreate(t, s, c)

	got, _ := s.GetCommentByID(context.Background(), c.ID)
	if got.ImagePath != "/uploads/photo.jpg" {
		t.Errorf("ImagePath = %q, want %q", got.ImagePath, "/uploads/photo.jpg")
	}
}

func TestCreateComment_NoImage(t *testing.T) {
	s := setupStore(t)
	c := &comment.Comment{UserID: "u1", TopicID: 1, Content: "no pic"}
	mustCreate(t, s, c)

	got, _ := s.GetCommentByID(context.Background(), c.ID)
	if got.ImagePath != "" {
		t.Errorf("ImagePath = %q, want empty", got.ImagePath)
	}
}

func TestGetCommentsByTopicIDWithVotes(t *testing.T) {
	s := setupStore(t)
	c1 := &comment.Comment{UserID: "u1", TopicID: 1, Content: "a"}
	mustCreate(t, s, c1)
	c2 := &comment.Comment{UserID: "u2", TopicID: 1, Content: "b"}
	mustCreate(t, s, c2)

	mustVote(t, s, "u2", c1.ID, 1)

	comments, err := s.GetCommentsByTopicIDWithVotes(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("GetCommentsByTopicIDWithVotes: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("got %d comments, want 2", len(comments))
	}
	if comments[0].UpvoteCount != 1 {
		t.Errorf("c1 UpvoteCount = %d, want 1", comments[0].UpvoteCount)
	}
}

func TestGetCommentsByTopicIDWithVotes_UserVote(t *testing.T) {
	s := setupStore(t)
	c1 := &comment.Comment{UserID: "u1", TopicID: 1, Content: "a"}
	mustCreate(t, s, c1)

	mustVote(t, s, "u2", c1.ID, 1)

	uid := "u2"
	comments, err := s.GetCommentsByTopicIDWithVotes(context.Background(), 1, &uid)
	if err != nil {
		t.Fatalf("GetCommentsByTopicIDWithVotes: %v", err)
	}
	if comments[0].UserVote == nil || *comments[0].UserVote != 1 {
		t.Errorf("UserVote = %v, want 1", comments[0].UserVote)
	}
}

func TestGetCommentByIDWithVotes(t *testing.T) {
	s := setupStore(t)
	c1 := &comment.Comment{UserID: "u1", TopicID: 1, Content: "vote me"}
	mustCreate(t, s, c1)

	mustVote(t, s, "u2", c1.ID, -1)

	got, err := s.GetCommentByIDWithVotes(context.Background(), c1.ID, nil)
	if err != nil {
		t.Fatalf("GetCommentByIDWithVotes: %v", err)
	}
	if got.DownvoteCount != 1 {
		t.Errorf("DownvoteCount = %d, want 1", got.DownvoteCount)
	}
	if got.VoteScore != -1 {
		t.Errorf("VoteScore = %d, want -1", got.VoteScore)
	}
}

func TestGetCommentByIDWithVotes_NotFound(t *testing.T) {
	s := setupStore(t)
	_, err := s.GetCommentByIDWithVotes(context.Background(), 999, nil)
	if !errors.Is(err, comment.ErrCommentNotFound) {
		t.Errorf("expected ErrCommentNotFound, got %v", err)
	}
}
