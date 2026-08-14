package store

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/platform/database"
	"social-network/internal/topic"
)

const topicSchema = `
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE
);
CREATE TABLE topics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    image_path TEXT DEFAULT '',
    visibility INTEGER NOT NULL DEFAULT 0,
    group_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE topic_allowed_users (
    topic_id INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (topic_id, user_id)
);
CREATE TABLE comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    topic_id INTEGER REFERENCES topics(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL
);
CREATE TABLE votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id INTEGER REFERENCES topics(id) ON DELETE CASCADE,
    comment_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
    reaction_type INTEGER NOT NULL CHECK(reaction_type IN (-1, 1)),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, topic_id),
    UNIQUE (user_id, comment_id)
);
CREATE UNIQUE INDEX idx_topic_votes ON votes(user_id, topic_id) WHERE comment_id IS NULL;
CREATE INDEX idx_votes_topic_reaction ON votes(topic_id, reaction_type) WHERE comment_id IS NULL;
CREATE INDEX idx_votes_user ON votes(user_id);
CREATE INDEX idx_topics_user ON topics(user_id);
CREATE INDEX idx_topics_created ON topics(created_at DESC);
CREATE INDEX idx_topics_group ON topics(group_id);
CREATE INDEX idx_topic_allowed_users_topic ON topic_allowed_users(topic_id);
CREATE TABLE follows (
    follower_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, followee_id)
);
CREATE INDEX idx_follows_followee ON follows(followee_id);
INSERT INTO users (id, email, username) VALUES ('u1', 'a@b.com', 'alice');
INSERT INTO users (id, email, username) VALUES ('u2', 'c@d.com', 'bob');
INSERT INTO users (id, email, username) VALUES ('u3', 'e@f.com', 'charlie');
INSERT INTO follows (follower_id, followee_id) VALUES ('u2', 'u1');
INSERT INTO follows (follower_id, followee_id) VALUES ('u3', 'u1');
INSERT INTO follows (follower_id, followee_id) VALUES ('u1', 'u3');`

func strPtr(s string) *string {
	return &s
}

func setupTopicStore(t *testing.T) *SQLiteStore {
	t.Helper()
	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.ExecContext(context.Background(), topicSchema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return NewSQLiteStore(db)
}

func TestCreateTopic(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{
		UserID:     "u1",
		Title:      "Hello",
		Content:    "World",
		ImagePath:  "/img.png",
		Visibility: topic.VisibilityPublic,
	}
	if err := s.CreateTopic(context.Background(), top, nil); err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	if top.ID == 0 {
		t.Fatal("ID not set after create")
	}

	got, err := s.GetTopicByID(context.Background(), top.ID, nil)
	if err != nil {
		t.Fatalf("GetTopicByID: %v", err)
	}
	if got.Title != "Hello" {
		t.Errorf("Title = %q, want %q", got.Title, "Hello")
	}
	if got.ImagePath != "/img.png" {
		t.Errorf("ImagePath = %q, want %q", got.ImagePath, "/img.png")
	}
	if got.Visibility != topic.VisibilityPublic {
		t.Errorf("Visibility = %d, want %d", got.Visibility, topic.VisibilityPublic)
	}
}

func TestCreateTopic_WithGroup(t *testing.T) {
	s := setupTopicStore(t)

	gid := "group-1"
	top := &topic.Topic{
		UserID:  "u1",
		Title:   "Group Post",
		Content: "In a group",
		GroupID: &gid,
	}
	if err := s.CreateTopic(context.Background(), top, nil); err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}

	got, err := s.GetTopicByID(context.Background(), top.ID, nil)
	if err != nil {
		t.Fatalf("GetTopicByID: %v", err)
	}
	if got.GroupID == nil || *got.GroupID != "group-1" {
		t.Errorf("GroupID = %v, want group-1", got.GroupID)
	}
}

func TestCreateTopic_WithAllowedUsers(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "Private", Content: "Secret", Visibility: topic.VisibilityPrivate}
	if err := s.CreateTopic(context.Background(), top, []string{"u2"}); err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}

	got, err := s.GetTopicByID(context.Background(), top.ID, strPtr("u1"))
	if err != nil {
		t.Fatalf("GetTopicByID: %v", err)
	}
	if len(got.AllowedUsers) != 1 || got.AllowedUsers[0] != "u2" {
		t.Errorf("AllowedUsers = %v, want [u2]", got.AllowedUsers)
	}
}

func TestDeleteTopic(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "Del", Content: "x"}
	if err := s.CreateTopic(context.Background(), top, nil); err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}

	if err := s.DeleteTopic(context.Background(), "u1", top.ID); err != nil {
		t.Fatalf("DeleteTopic: %v", err)
	}

	_, err := s.GetTopicByID(context.Background(), top.ID, nil)
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("GetTopicByID after delete: err = %v, want ErrTopicNotFound", err)
	}
}

func TestDeleteTopic_WrongUser(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "Mine", Content: "x"}
	if err := s.CreateTopic(context.Background(), top, nil); err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}

	err := s.DeleteTopic(context.Background(), "u2", top.ID)
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("DeleteTopic wrong user: err = %v, want ErrTopicNotFound", err)
	}
}

func TestUpdateTopic(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "Old", Content: "Old"}
	if err := s.CreateTopic(context.Background(), top, nil); err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}

	top.Title = "New"
	top.Content = "New"
	top.Visibility = topic.VisibilityFollowers
	if err := s.UpdateTopic(context.Background(), top, nil); err != nil {
		t.Fatalf("UpdateTopic: %v", err)
	}

	got, err := s.GetTopicByID(context.Background(), top.ID, strPtr("u1"))
	if err != nil {
		t.Fatalf("GetTopicByID: %v", err)
	}
	if got.Title != "New" {
		t.Errorf("Title = %q, want %q", got.Title, "New")
	}
	if got.Visibility != topic.VisibilityFollowers {
		t.Errorf("Visibility = %d, want %d", got.Visibility, topic.VisibilityFollowers)
	}
}

func TestUpdateTopic_NotFound(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{ID: 9999, UserID: "u1", Title: "X", Content: "X"}
	err := s.UpdateTopic(context.Background(), top, nil)
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("UpdateTopic not found: err = %v, want ErrTopicNotFound", err)
	}
}

func TestGetTopicByID_NotFound(t *testing.T) {
	s := setupTopicStore(t)

	_, err := s.GetTopicByID(context.Background(), 9999, nil)
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("err = %v, want ErrTopicNotFound", err)
	}
}

func TestGetFeed(t *testing.T) {
	s := setupTopicStore(t)

	for i := range 5 {
		top := &topic.Topic{UserID: "u1", Title: "Post", Content: "x"}
		if err := s.CreateTopic(context.Background(), top, nil); err != nil {
			t.Fatalf("CreateTopic(%d): %v", i, err)
		}
	}

	topics, count, err := s.GetFeed(context.Background(), "u1", 1, 3, "created_at", "DESC", "")
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
	if len(topics) != 3 {
		t.Errorf("len = %d, want 3", len(topics))
	}
}

func TestGetFeed_WithUserVote(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "Vote", Content: "x"}
	if err := s.CreateTopic(context.Background(), top, nil); err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	if _, err := s.CastVote(context.Background(), "u1", top.ID, 1); err != nil {
		t.Fatalf("CastVote: %v", err)
	}

	topics, _, err := s.GetFeed(context.Background(), "u1", 1, 10, "created_at", "DESC", "")
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if len(topics) != 1 {
		t.Fatalf("len = %d, want 1", len(topics))
	}
	if topics[0].UserVote == nil || *topics[0].UserVote != 1 {
		t.Errorf("UserVote = %v, want 1", topics[0].UserVote)
	}
	if topics[0].UpvoteCount != 1 {
		t.Errorf("UpvoteCount = %d, want 1", topics[0].UpvoteCount)
	}
}

func TestGetFeed_Filter(t *testing.T) {
	s := setupTopicStore(t)

	_ = s.CreateTopic(context.Background(), &topic.Topic{UserID: "u1", Title: "Alpha", Content: "x"}, nil)
	_ = s.CreateTopic(context.Background(), &topic.Topic{UserID: "u1", Title: "Beta", Content: "y"}, nil)

	topics, count, err := s.GetFeed(context.Background(), "u1", 1, 10, "created_at", "DESC", "Alpha")
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if len(topics) != 1 || topics[0].Title != "Alpha" {
		t.Errorf("topics = %v, want [Alpha]", topics)
	}
}

func TestGetFeed_ExcludesGroupPosts(t *testing.T) {
	s := setupTopicStore(t)

	_ = s.CreateTopic(context.Background(), &topic.Topic{UserID: "u1", Title: "Regular", Content: "x"}, nil)
	gid := "group-1"
	_ = s.CreateTopic(context.Background(), &topic.Topic{UserID: "u1", Title: "GroupOnly", Content: "y", GroupID: &gid}, nil)

	topics, count, err := s.GetFeed(context.Background(), "u1", 1, 10, "created_at", "DESC", "")
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if len(topics) != 1 || topics[0].Title != "Regular" {
		t.Errorf("topics = %v, want [Regular]", topics)
	}
}

func TestCommentsCount(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "Commented", Content: "x"}
	if err := s.CreateTopic(context.Background(), top, nil); err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	if _, err := s.db.ExecContext(context.Background(),
		`INSERT INTO comments (topic_id, user_id, content) VALUES (?, 'u2', 'c1'), (?, 'u2', 'c2')`,
		top.ID, top.ID); err != nil {
		t.Fatalf("insert comments: %v", err)
	}

	topics, _, err := s.GetFeed(context.Background(), "u1", 1, 10, "created_at", "DESC", "")
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if len(topics) != 1 || topics[0].CommentsCount != 2 {
		t.Errorf("GetFeed CommentsCount = %d, want 2", topics[0].CommentsCount)
	}

	got, err := s.GetTopicByID(context.Background(), top.ID, nil)
	if err != nil {
		t.Fatalf("GetTopicByID: %v", err)
	}
	if got.CommentsCount != 2 {
		t.Errorf("GetTopicByID CommentsCount = %d, want 2", got.CommentsCount)
	}
}

func TestGetImagePath(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "Img", Content: "x", ImagePath: "/images/photo.jpg"}
	_ = s.CreateTopic(context.Background(), top, nil)

	path, err := s.GetImagePathFromTopicID(context.Background(), top.ID, "u1")
	if err != nil {
		t.Fatalf("GetImagePath: %v", err)
	}
	if path != "/images/photo.jpg" {
		t.Errorf("path = %q, want %q", path, "/images/photo.jpg")
	}
}

func TestGetTopicsByUserID(t *testing.T) {
	s := setupTopicStore(t)

	_ = s.CreateTopic(context.Background(), &topic.Topic{UserID: "u1", Title: "A", Content: "x"}, nil)
	_ = s.CreateTopic(context.Background(), &topic.Topic{UserID: "u1", Title: "B", Content: "y"}, nil)
	_ = s.CreateTopic(context.Background(), &topic.Topic{UserID: "u2", Title: "C", Content: "z"}, nil)

	topics, count, err := s.GetTopicsByUserID(context.Background(), "u1", "u1", 1, 10)
	if err != nil {
		t.Fatalf("GetTopicsByUserID: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	if len(topics) != 2 {
		t.Errorf("len = %d, want 2", len(topics))
	}
}

func TestGetTopicByID_NullImagePath(t *testing.T) {
	s := setupTopicStore(t)

	result, err := s.db.ExecContext(context.Background(),
		`INSERT INTO topics (user_id, title, content, image_path) VALUES ('u1', 'No Image', 'x', NULL)`)
	if err != nil {
		t.Fatalf("seed topic: %v", err)
	}
	id, _ := result.LastInsertId()

	got, err := s.GetTopicByID(context.Background(), int(id), nil)
	if err != nil {
		t.Fatalf("GetTopicByID: %v", err)
	}
	if got.ImagePath != "" {
		t.Errorf("ImagePath = %q, want empty string", got.ImagePath)
	}
}

func TestGetFeed_NullImagePath(t *testing.T) {
	s := setupTopicStore(t)

	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO topics (user_id, title, content, image_path) VALUES ('u1', 'No Image', 'x', NULL)`)
	if err != nil {
		t.Fatalf("seed topic: %v", err)
	}

	topics, _, err := s.GetFeed(context.Background(), "u1", 1, 10, "created_at", "DESC", "")
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if len(topics) != 1 {
		t.Fatalf("len = %d, want 1", len(topics))
	}
	if topics[0].ImagePath != "" {
		t.Errorf("ImagePath = %q, want empty string", topics[0].ImagePath)
	}
}
