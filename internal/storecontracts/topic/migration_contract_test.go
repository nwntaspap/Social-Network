package topic_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	legacydomain "social-network/internal/domain/topic"
	legacystore "social-network/internal/infra/storage/sqlite/topics"
	"social-network/internal/platform/database"
	"social-network/internal/topic"
	newstore "social-network/internal/topic/store"
)

// Legacy schema matches production: topics without visibility/group_id, with categories.
const legacySchema = `
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT,
    first_name TEXT,
    last_name TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    color TEXT DEFAULT '#CCCCCC',
    created_by TEXT NOT NULL REFERENCES users(id)
);
CREATE TABLE topics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    image_path TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE topic_categories (
    topic_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (topic_id, category_id),
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
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
CREATE INDEX idx_topics_created ON topics(created_at DESC);`

// New schema matches internal/topic/store: has visibility, group_id, topic_allowed_users.
const newSchema = legacySchema + `
ALTER TABLE topics ADD COLUMN visibility INTEGER NOT NULL DEFAULT 0;
ALTER TABLE topics ADD COLUMN group_id TEXT;
CREATE TABLE topic_allowed_users (
    topic_id INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (topic_id, user_id)
);
CREATE INDEX idx_topic_allowed_users_topic ON topic_allowed_users(topic_id);
CREATE INDEX idx_topics_group ON topics(group_id);`

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

func seedUser(t *testing.T, db *sql.DB, id, username string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO users (id, email, username, password_hash) VALUES (?, ?, ?, ?)`,
		id, id+"@example.com", username, "hash")
	if err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
}

func seedUserNewDB(t *testing.T, db database.DB, id, username string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO users (id, email, username, password_hash) VALUES (?, ?, ?, ?)`,
		id, id+"@example.com", username, "hash")
	if err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
}

// legacyAdapter wraps the legacy topic repo to present the new topic.Repository interface.
type legacyAdapter struct {
	repo *legacystore.Repo
}

func (a *legacyAdapter) CreateTopic(ctx context.Context, t *topic.Topic, _ []string) error {
	lt := &legacydomain.Topic{
		UserID:    t.UserID,
		Title:     t.Title,
		Content:   t.Content,
		ImagePath: t.ImagePath,
	}
	if err := a.repo.CreateTopic(ctx, lt); err != nil {
		return err
	}
	t.ID = lt.ID
	return nil
}

func (a *legacyAdapter) UpdateTopic(ctx context.Context, t *topic.Topic, _ []string) error {
	lt := &legacydomain.Topic{
		ID:        t.ID,
		UserID:    t.UserID,
		Title:     t.Title,
		Content:   t.Content,
		ImagePath: t.ImagePath,
	}
	return a.repo.UpdateTopic(ctx, lt)
}

func (a *legacyAdapter) DeleteTopic(ctx context.Context, userID string, topicID int) error {
	return a.repo.DeleteTopic(ctx, userID, topicID)
}

func (a *legacyAdapter) GetTopicByID(ctx context.Context, topicID int, userID *string) (*topic.Topic, error) {
	lt, err := a.repo.GetTopicByID(ctx, topicID, userID)
	if err != nil {
		return nil, err
	}
	return legacyToNewTopic(lt), nil
}

func (a *legacyAdapter) GetImagePathFromTopicID(ctx context.Context, topicID int, userID string) (string, error) {
	return a.repo.GetImagePathFromTopicID(ctx, topicID, userID)
}

func (a *legacyAdapter) GetFeed(ctx context.Context, userID string, page, size int, orderBy, order, filter string) ([]topic.Topic, int, error) {
	topics, err := a.repo.GetAllTopics(ctx, page, size, 0, orderBy, order, filter, &userID)
	if err != nil {
		return nil, 0, err
	}
	count, err := a.repo.GetTotalTopicsCount(ctx, filter, 0)
	if err != nil {
		return nil, 0, err
	}
	result := make([]topic.Topic, 0, len(topics))
	for i := range topics {
		result = append(result, *legacyToNewTopic(&topics[i]))
	}
	return result, count, nil
}

func (a *legacyAdapter) GetTopicsByUserID(_ context.Context, _, _ string, _, _ int) ([]topic.Topic, int, error) {
	return nil, 0, nil
}

func (a *legacyAdapter) GetTopicsByGroupID(_ context.Context, _ string, _, _ int) ([]topic.Topic, int, error) {
	return nil, 0, nil
}

func (a *legacyAdapter) CastVote(_ context.Context, _ string, _ int, _ int) (topic.VoteChange, error) {
	return topic.VoteChangeAdded, nil
}

func (a *legacyAdapter) DeleteVote(_ context.Context, _ string, _ int) error {
	return nil
}

func (a *legacyAdapter) GetVoteCounts(_ context.Context, _ int) (*topic.VoteCounts, error) {
	return &topic.VoteCounts{}, nil
}

func legacyToNewTopic(lt *legacydomain.Topic) *topic.Topic {
	return &topic.Topic{
		ID:            lt.ID,
		UserID:        lt.UserID,
		Title:         lt.Title,
		Content:       lt.Content,
		ImagePath:     lt.ImagePath,
		OwnerUsername: lt.OwnerUsername,
		UpvoteCount:   lt.UpvoteCount,
		DownvoteCount: lt.DownvoteCount,
		VoteScore:     lt.VoteScore,
		UserVote:      lt.UserVote,
	}
}

// --- contract tests ---

func TestMigrationContract_CreateAndGetTopic(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		db := setupLegacyDB(t)
		seedUser(t, db, "user-1", "alice")
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		topic := &topic.Topic{
			UserID:    "user-1",
			Title:     "Test Title",
			Content:   "Test Content",
			ImagePath: "",
		}
		if err := adapter.CreateTopic(context.Background(), topic, nil); err != nil {
			t.Fatalf("CreateTopic() error = %v", err)
		}
		if topic.ID == 0 {
			t.Fatal("CreateTopic() did not set ID")
		}

		got, err := adapter.GetTopicByID(context.Background(), topic.ID, nil)
		if err != nil {
			t.Fatalf("GetTopicByID() error = %v", err)
		}
		if got.Title != "Test Title" {
			t.Errorf("Title = %q, want %q", got.Title, "Test Title")
		}
		if got.Content != "Test Content" {
			t.Errorf("Content = %q, want %q", got.Content, "Test Content")
		}
		if got.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", got.UserID, "user-1")
		}
	})

	t.Run("new", func(t *testing.T) {
		db := setupNewDB(t)
		seedUserNewDB(t, db, "user-1", "alice")
		s := newstore.NewSQLiteStore(db)

		topic := &topic.Topic{
			UserID:    "user-1",
			Title:     "Test Title",
			Content:   "Test Content",
			ImagePath: "",
		}
		if err := s.CreateTopic(context.Background(), topic, nil); err != nil {
			t.Fatalf("CreateTopic() error = %v", err)
		}
		if topic.ID == 0 {
			t.Fatal("CreateTopic() did not set ID")
		}

		got, err := s.GetTopicByID(context.Background(), topic.ID, nil)
		if err != nil {
			t.Fatalf("GetTopicByID() error = %v", err)
		}
		if got.Title != "Test Title" {
			t.Errorf("Title = %q, want %q", got.Title, "Test Title")
		}
		if got.Content != "Test Content" {
			t.Errorf("Content = %q, want %q", got.Content, "Test Content")
		}
		if got.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", got.UserID, "user-1")
		}
	})
}

func TestMigrationContract_DeleteTopic(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		db := setupLegacyDB(t)
		seedUser(t, db, "user-1", "alice")
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		topic := &topic.Topic{UserID: "user-1", Title: "Del Me", Content: "x"}
		if err := adapter.CreateTopic(context.Background(), topic, nil); err != nil {
			t.Fatalf("CreateTopic() error = %v", err)
		}

		if err := adapter.DeleteTopic(context.Background(), "user-1", topic.ID); err != nil {
			t.Fatalf("DeleteTopic() error = %v", err)
		}

		_, err := adapter.GetTopicByID(context.Background(), topic.ID, nil)
		if err == nil {
			t.Fatal("GetTopicByID() after delete expected error, got nil")
		}
	})

	t.Run("new", func(t *testing.T) {
		db := setupNewDB(t)
		seedUserNewDB(t, db, "user-1", "alice")
		s := newstore.NewSQLiteStore(db)

		topic := &topic.Topic{UserID: "user-1", Title: "Del Me", Content: "x"}
		if err := s.CreateTopic(context.Background(), topic, nil); err != nil {
			t.Fatalf("CreateTopic() error = %v", err)
		}

		if err := s.DeleteTopic(context.Background(), "user-1", topic.ID); err != nil {
			t.Fatalf("DeleteTopic() error = %v", err)
		}

		_, err := s.GetTopicByID(context.Background(), topic.ID, nil)
		if err == nil {
			t.Fatal("GetTopicByID() after delete expected error, got nil")
		}
	})
}

func TestMigrationContract_DeleteTopic_WrongUser(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		db := setupLegacyDB(t)
		seedUser(t, db, "user-1", "alice")
		seedUser(t, db, "user-2", "bob")
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		topic := &topic.Topic{UserID: "user-1", Title: "Mine", Content: "x"}
		if err := adapter.CreateTopic(context.Background(), topic, nil); err != nil {
			t.Fatalf("CreateTopic() error = %v", err)
		}

		err := adapter.DeleteTopic(context.Background(), "user-2", topic.ID)
		if err == nil {
			t.Fatal("DeleteTopic() by wrong user expected error, got nil")
		}
	})

	t.Run("new", func(t *testing.T) {
		db := setupNewDB(t)
		seedUserNewDB(t, db, "user-1", "alice")
		seedUserNewDB(t, db, "user-2", "bob")
		s := newstore.NewSQLiteStore(db)

		topic := &topic.Topic{UserID: "user-1", Title: "Mine", Content: "x"}
		if err := s.CreateTopic(context.Background(), topic, nil); err != nil {
			t.Fatalf("CreateTopic() error = %v", err)
		}

		err := s.DeleteTopic(context.Background(), "user-2", topic.ID)
		if err == nil {
			t.Fatal("DeleteTopic() by wrong user expected error, got nil")
		}
	})
}

func TestMigrationContract_UpdateTopic(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		db := setupLegacyDB(t)
		seedUser(t, db, "user-1", "alice")
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		top := &topic.Topic{UserID: "user-1", Title: "Old", Content: "Old content"}
		if err := adapter.CreateTopic(context.Background(), top, nil); err != nil {
			t.Fatalf("CreateTopic() error = %v", err)
		}

		top.Title = "New"
		top.Content = "New content"
		if err := adapter.UpdateTopic(context.Background(), top, nil); err != nil {
			t.Fatalf("UpdateTopic() error = %v", err)
		}

		got, err := adapter.GetTopicByID(context.Background(), top.ID, nil)
		if err != nil {
			t.Fatalf("GetTopicByID() error = %v", err)
		}
		if got.Title != "New" {
			t.Errorf("Title = %q, want %q", got.Title, "New")
		}
		if got.Content != "New content" {
			t.Errorf("Content = %q, want %q", got.Content, "New content")
		}
	})

	t.Run("new", func(t *testing.T) {
		db := setupNewDB(t)
		seedUserNewDB(t, db, "user-1", "alice")
		s := newstore.NewSQLiteStore(db)

		top := &topic.Topic{UserID: "user-1", Title: "Old", Content: "Old content"}
		if err := s.CreateTopic(context.Background(), top, nil); err != nil {
			t.Fatalf("CreateTopic() error = %v", err)
		}

		top.Title = "New"
		top.Content = "New content"
		if err := s.UpdateTopic(context.Background(), top, nil); err != nil {
			t.Fatalf("UpdateTopic() error = %v", err)
		}

		got, err := s.GetTopicByID(context.Background(), top.ID, nil)
		if err != nil {
			t.Fatalf("GetTopicByID() error = %v", err)
		}
		if got.Title != "New" {
			t.Errorf("Title = %q, want %q", got.Title, "New")
		}
		if got.Content != "New content" {
			t.Errorf("Content = %q, want %q", got.Content, "New content")
		}
	})
}

func TestMigrationContract_GetImagePath(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		db := setupLegacyDB(t)
		seedUser(t, db, "user-1", "alice")
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		top := &topic.Topic{UserID: "user-1", Title: "Img", Content: "x", ImagePath: "/images/test.jpg"}
		if err := adapter.CreateTopic(context.Background(), top, nil); err != nil {
			t.Fatalf("CreateTopic() error = %v", err)
		}

		path, err := adapter.GetImagePathFromTopicID(context.Background(), top.ID, "user-1")
		if err != nil {
			t.Fatalf("GetImagePathFromTopicID() error = %v", err)
		}
		if path != "/images/test.jpg" {
			t.Errorf("ImagePath = %q, want %q", path, "/images/test.jpg")
		}
	})

	t.Run("new", func(t *testing.T) {
		db := setupNewDB(t)
		seedUserNewDB(t, db, "user-1", "alice")
		s := newstore.NewSQLiteStore(db)

		top := &topic.Topic{UserID: "user-1", Title: "Img", Content: "x", ImagePath: "/images/test.jpg"}
		if err := s.CreateTopic(context.Background(), top, nil); err != nil {
			t.Fatalf("CreateTopic() error = %v", err)
		}

		path, err := s.GetImagePathFromTopicID(context.Background(), top.ID, "user-1")
		if err != nil {
			t.Fatalf("GetImagePathFromTopicID() error = %v", err)
		}
		if path != "/images/test.jpg" {
			t.Errorf("ImagePath = %q, want %q", path, "/images/test.jpg")
		}
	})
}

func TestMigrationContract_GetFeed(t *testing.T) {
	t.Run("legacy", func(t *testing.T) {
		db := setupLegacyDB(t)
		seedUser(t, db, "user-1", "alice")
		adapter := &legacyAdapter{repo: legacystore.NewRepo(db)}

		for i := range 5 {
			top := &topic.Topic{UserID: "user-1", Title: "Post", Content: "x"}
			if err := adapter.CreateTopic(context.Background(), top, nil); err != nil {
				t.Fatalf("CreateTopic(%d) error = %v", i, err)
			}
		}

		topics, count, err := adapter.GetFeed(context.Background(), "user-1", 1, 3, "created_at", "DESC", "")
		if err != nil {
			t.Fatalf("GetFeed() error = %v", err)
		}
		if count != 5 {
			t.Errorf("count = %d, want 5", count)
		}
		if len(topics) != 3 {
			t.Errorf("len(topics) = %d, want 3", len(topics))
		}
	})

	t.Run("new", func(t *testing.T) {
		db := setupNewDB(t)
		seedUserNewDB(t, db, "user-1", "alice")
		s := newstore.NewSQLiteStore(db)

		for i := range 5 {
			top := &topic.Topic{UserID: "user-1", Title: "Post", Content: "x"}
			if err := s.CreateTopic(context.Background(), top, nil); err != nil {
				t.Fatalf("CreateTopic(%d) error = %v", i, err)
			}
		}

		topics, count, err := s.GetFeed(context.Background(), "user-1", 1, 3, "created_at", "DESC", "")
		if err != nil {
			t.Fatalf("GetFeed() error = %v", err)
		}
		if count != 5 {
			t.Errorf("count = %d, want 5", count)
		}
		if len(topics) != 3 {
			t.Errorf("len(topics) = %d, want 3", len(topics))
		}
	})
}
