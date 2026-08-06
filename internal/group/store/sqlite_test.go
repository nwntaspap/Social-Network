package store

import (
	"context"
	"testing"

	"social-network/internal/group"
	"social-network/internal/platform/database"
)

const groupsSchema = `
CREATE TABLE groups (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    creator_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

func setupGroupStore(t *testing.T) *SQLiteStore {
	t.Helper()

	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(), groupsSchema); err != nil {
		t.Fatalf("create table: %v", err)
	}

	return NewSQLiteStore(db)
}

func seedGroup(t *testing.T, s *SQLiteStore, g *group.Group) {
	t.Helper()
	if err := s.CreateGroup(context.Background(), g); err != nil {
		t.Fatalf("seed group %s: %v", g.ID, err)
	}
}

func seedGroupAt(t *testing.T, s *SQLiteStore, g *group.Group, createdAt string) {
	t.Helper()
	if _, err := s.db.ExecContext(context.Background(),
		`INSERT INTO groups (id, title, description, creator_id, created_at) VALUES (?, ?, ?, ?, ?)`,
		g.ID, g.Title, g.Description, g.CreatorID, createdAt); err != nil {
		t.Fatalf("seed group %s: %v", g.ID, err)
	}
}

func TestSearchGroups_FiltersByTitleAndDescription(t *testing.T) {
	s := setupGroupStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go Meetup", Description: "Gophers hang out", CreatorID: "u1"})
	seedGroup(t, s, &group.Group{ID: "g2", Title: "Rust Users", Description: "Ferris fans", CreatorID: "u1"})
	seedGroup(t, s, &group.Group{ID: "g3", Title: "Weekly Reads", Description: "Read about Go and Rust", CreatorID: "u1"})

	groups, total, err := s.SearchGroups(ctx, "go", 1, 10)
	if err != nil {
		t.Fatalf("SearchGroups(go) error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(groups) != 2 {
		t.Fatalf("SearchGroups(go) returned %d groups, want 2", len(groups))
	}

	ids := map[string]bool{}
	for _, g := range groups {
		ids[g.ID] = true
	}
	if !ids["g1"] || !ids["g3"] {
		t.Errorf("SearchGroups(go) ids = %v, want g1 and g3", ids)
	}

	groups, total, err = s.SearchGroups(ctx, "ferris", 1, 10)
	if err != nil {
		t.Fatalf("SearchGroups(ferris) error = %v", err)
	}
	if total != 1 || len(groups) != 1 || groups[0].ID != "g2" {
		t.Errorf("SearchGroups(ferris) = %+v (total %d), want [g2] (1)", groups, total)
	}
}

func TestSearchGroups_EmptyQueryReturnsAll(t *testing.T) {
	s := setupGroupStore(t)
	ctx := context.Background()

	seedGroup(t, s, &group.Group{ID: "g1", Title: "Go", CreatorID: "u1"})
	seedGroup(t, s, &group.Group{ID: "g2", Title: "Rust", CreatorID: "u1"})
	seedGroup(t, s, &group.Group{ID: "g3", Title: "Zig", CreatorID: "u1"})

	groups, total, err := s.SearchGroups(ctx, "", 1, 10)
	if err != nil {
		t.Fatalf("SearchGroups(empty) error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(groups) != 3 {
		t.Errorf("SearchGroups(empty) returned %d groups, want 3", len(groups))
	}
}

func TestSearchGroups_Paginates(t *testing.T) {
	s := setupGroupStore(t)
	ctx := context.Background()

	seedGroupAt(t, s, &group.Group{ID: "g1", Title: "Go", CreatorID: "u1"}, "2024-01-01 10:00:00")
	seedGroupAt(t, s, &group.Group{ID: "g2", Title: "Rust", CreatorID: "u1"}, "2024-01-01 11:00:00")
	seedGroupAt(t, s, &group.Group{ID: "g3", Title: "Zig", CreatorID: "u1"}, "2024-01-01 12:00:00")

	groups, total, err := s.SearchGroups(ctx, "", 2, 2)
	if err != nil {
		t.Fatalf("SearchGroups(page2) error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(groups) != 1 {
		t.Fatalf("SearchGroups(page2) returned %d groups, want 1", len(groups))
	}
	if groups[0].ID != "g1" {
		t.Errorf("groups[0].ID = %q, want g1", groups[0].ID)
	}
}

const groupPostsSchema = `
CREATE TABLE group_posts (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    author_id TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    image_path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);
CREATE TABLE group_post_comments (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    author_id TEXT NOT NULL,
    content TEXT NOT NULL,
    image_path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE group_post_votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    post_id TEXT NOT NULL,
    reaction_type INTEGER NOT NULL CHECK(reaction_type IN (-1, 1)),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, post_id)
);`

func setupGroupPostStore(t *testing.T) *SQLiteStore {
	t.Helper()
	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.ExecContext(context.Background(), groupPostsSchema); err != nil {
		t.Fatalf("create table: %v", err)
	}
	return NewSQLiteStore(db)
}

func TestGetPostsByGroupID_NullImagePath(t *testing.T) {
	s := setupGroupPostStore(t)
	ctx := context.Background()

	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO group_posts (id, group_id, author_id, title, content, image_path)
		 VALUES ('p1', 'g1', 'u1', 'Post', 'content', NULL)`); err != nil {
		t.Fatalf("seed group post: %v", err)
	}

	posts, _, err := s.GetPostsByGroupID(ctx, "g1", "u1", 1, 10)
	if err != nil {
		t.Fatalf("GetPostsByGroupID: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("len = %d, want 1", len(posts))
	}
	if posts[0].ImagePath != "" {
		t.Errorf("ImagePath = %q, want empty string", posts[0].ImagePath)
	}
}

func TestCreatePost_ThenGetPostsByGroupID_ReturnsIt(t *testing.T) {
	s := setupGroupPostStore(t)
	ctx := context.Background()

	p := &group.Post{
		ID:        "p-roundtrip",
		GroupID:   "g1",
		AuthorID:  "u1",
		Title:     "Hello Group",
		Content:   "A new group post",
		ImagePath: "/static/images/uploads/pic.png",
	}
	if err := s.CreatePost(ctx, p); err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	posts, total, err := s.GetPostsByGroupID(ctx, "g1", "u1", 1, 10)
	if err != nil {
		t.Fatalf("GetPostsByGroupID: %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(posts) != 1 {
		t.Fatalf("len = %d, want 1", len(posts))
	}
	got := posts[0]
	if got.ID != p.ID || got.GroupID != p.GroupID || got.AuthorID != p.AuthorID ||
		got.Title != p.Title || got.Content != p.Content || got.ImagePath != p.ImagePath {
		t.Errorf("post = %+v, want %+v", got, p)
	}
}

func TestGetPostComments_NullImagePath(t *testing.T) {
	s := setupGroupPostStore(t)
	ctx := context.Background()

	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO group_posts (id, group_id, author_id, title, content, image_path)
		 VALUES ('p1', 'g1', 'u1', 'Post', 'content', NULL)`); err != nil {
		t.Fatalf("seed group post: %v", err)
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO group_post_comments (id, post_id, author_id, content, image_path)
		 VALUES ('c1', 'p1', 'u1', 'comment', NULL)`); err != nil {
		t.Fatalf("seed post comment: %v", err)
	}

	comments, _, err := s.GetPostComments(ctx, "p1", 1, 10)
	if err != nil {
		t.Fatalf("GetPostComments: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("len = %d, want 1", len(comments))
	}
	if comments[0].ImagePath != "" {
		t.Errorf("ImagePath = %q, want empty string", comments[0].ImagePath)
	}
}

func TestCastPostVote_Toggle(t *testing.T) {
	s := setupGroupPostStore(t)
	ctx := context.Background()

	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO group_posts (id, group_id, author_id, title, content)
		 VALUES ('p1', 'g1', 'u1', 'Post', 'content')`); err != nil {
		t.Fatalf("seed group post: %v", err)
	}

	if err := s.CastPostVote(ctx, "u1", "p1", 1); err != nil {
		t.Fatalf("CastPostVote(like): %v", err)
	}
	counts, err := s.GetPostVoteCounts(ctx, "p1")
	if err != nil {
		t.Fatalf("GetPostVoteCounts: %v", err)
	}
	if counts.Upvotes != 1 || counts.Downvotes != 0 || counts.Score != 1 {
		t.Errorf("after like: counts = %+v, want up=1 down=0 score=1", counts)
	}

	if err := s.CastPostVote(ctx, "u1", "p1", 1); err != nil {
		t.Fatalf("CastPostVote(same like): %v", err)
	}
	counts, _ = s.GetPostVoteCounts(ctx, "p1")
	if counts.Upvotes != 0 {
		t.Errorf("after toggle-off: upvotes = %d, want 0", counts.Upvotes)
	}

	if err := s.CastPostVote(ctx, "u1", "p1", -1); err != nil {
		t.Fatalf("CastPostVote(dislike): %v", err)
	}
	counts, _ = s.GetPostVoteCounts(ctx, "p1")
	if counts.Downvotes != 1 || counts.Score != -1 {
		t.Errorf("after dislike: counts = %+v, want down=1 score=-1", counts)
	}

	if err := s.CastPostVote(ctx, "u1", "p1", 1); err != nil {
		t.Fatalf("CastPostVote(switch to like): %v", err)
	}
	counts, _ = s.GetPostVoteCounts(ctx, "p1")
	if counts.Upvotes != 1 || counts.Downvotes != 0 || counts.Score != 1 {
		t.Errorf("after switch: counts = %+v, want up=1 down=0 score=1", counts)
	}
}

func TestGetPostsByGroupID_IncludesVoteState(t *testing.T) {
	s := setupGroupPostStore(t)
	ctx := context.Background()

	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO group_posts (id, group_id, author_id, title, content)
		 VALUES ('p1', 'g1', 'u1', 'Post', 'content')`); err != nil {
		t.Fatalf("seed group post: %v", err)
	}

	if err := s.CastPostVote(ctx, "u1", "p1", 1); err != nil {
		t.Fatalf("CastPostVote: %v", err)
	}
	if err := s.CastPostVote(ctx, "u2", "p1", -1); err != nil {
		t.Fatalf("CastPostVote u2: %v", err)
	}

	posts, _, err := s.GetPostsByGroupID(ctx, "g1", "u1", 1, 10)
	if err != nil {
		t.Fatalf("GetPostsByGroupID: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("len = %d, want 1", len(posts))
	}
	got := posts[0]
	if got.UpvoteCount != 1 || got.DownvoteCount != 1 || got.VoteScore != 0 {
		t.Errorf("counts = %d/%d/%d, want 1/1/0", got.UpvoteCount, got.DownvoteCount, got.VoteScore)
	}
	if got.UserVote == nil || *got.UserVote != 1 {
		t.Errorf("UserVote = %v, want 1", got.UserVote)
	}
}
