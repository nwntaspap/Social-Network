package store

import (
	"context"
	"errors"
	"sort"
	"testing"

	"social-network/internal/topic"
)

func seedVisibilityTopics(t *testing.T, s *SQLiteStore) int {
	t.Helper()
	pub := &topic.Topic{UserID: "u1", Title: "Public", Content: "x", Visibility: topic.VisibilityPublic}
	if err := s.CreateTopic(context.Background(), pub, nil); err != nil {
		t.Fatalf("create public: %v", err)
	}
	followers := &topic.Topic{UserID: "u1", Title: "Followers", Content: "x", Visibility: topic.VisibilityFollowers}
	if err := s.CreateTopic(context.Background(), followers, nil); err != nil {
		t.Fatalf("create followers: %v", err)
	}
	priv := &topic.Topic{UserID: "u1", Title: "Private", Content: "x", Visibility: topic.VisibilityPrivate}
	if err := s.CreateTopic(context.Background(), priv, []string{"u2"}); err != nil {
		t.Fatalf("create private: %v", err)
	}
	return priv.ID
}

func topicTitles(topics []topic.Topic) []string {
	titles := make([]string, 0, len(topics))
	for _, t := range topics {
		titles = append(titles, t.Title)
	}
	return titles
}

func assertTitles(t *testing.T, got []string, want ...string) {
	t.Helper()
	sort.Strings(got)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("titles = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("titles = %v, want %v", got, want)
		}
	}
}

func TestGetFeed_Visibility(t *testing.T) {
	s := setupTopicStore(t)
	seedVisibilityTopics(t, s)

	cases := []struct {
		name       string
		requester  string
		wantTitles []string
		wantCount  int
	}{
		{name: "anonymous only public", requester: "", wantTitles: []string{"Public"}, wantCount: 1},
		{name: "owner sees all", requester: "u1", wantTitles: []string{"Private", "Followers", "Public"}, wantCount: 3},
		{name: "follower sees public+followers", requester: "u3", wantTitles: []string{"Followers", "Public"}, wantCount: 2},
		{name: "follower allowed sees all", requester: "u2", wantTitles: []string{"Private", "Followers", "Public"}, wantCount: 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			topics, count, err := s.GetFeed(context.Background(), tc.requester, 1, 10, "created_at", "DESC", "")
			if err != nil {
				t.Fatalf("GetFeed: %v", err)
			}
			assertTitles(t, topicTitles(topics), tc.wantTitles...)
			if count != tc.wantCount {
				t.Errorf("count = %d, want %d", count, tc.wantCount)
			}
		})
	}
}

func TestGetFeed_VisibilityFilterCount(t *testing.T) {
	s := setupTopicStore(t)
	seedVisibilityTopics(t, s)

	// filter must not leak non-visible topics through the filter path
	topics, count, err := s.GetFeed(context.Background(), "u3", 1, 10, "created_at", "DESC", "Private")
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if count != 0 || len(topics) != 0 {
		t.Errorf("filtered = %d/%d, want 0/0", len(topics), count)
	}
}

func TestGetTopicsByUserID_Visibility(t *testing.T) {
	s := setupTopicStore(t)
	seedVisibilityTopics(t, s)

	cases := []struct {
		name       string
		owner      string
		requester  string
		wantTitles []string
		wantCount  int
	}{
		{name: "anonymous", owner: "u1", requester: "", wantTitles: []string{"Public"}, wantCount: 1},
		{name: "owner", owner: "u1", requester: "u1", wantTitles: []string{"Private", "Followers", "Public"}, wantCount: 3},
		{name: "follower", owner: "u1", requester: "u3", wantTitles: []string{"Followers", "Public"}, wantCount: 2},
		{name: "allowed follower", owner: "u1", requester: "u2", wantTitles: []string{"Private", "Followers", "Public"}, wantCount: 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			topics, count, err := s.GetTopicsByUserID(context.Background(), tc.owner, tc.requester, 1, 10)
			if err != nil {
				t.Fatalf("GetTopicsByUserID: %v", err)
			}
			assertTitles(t, topicTitles(topics), tc.wantTitles...)
			if count != tc.wantCount {
				t.Errorf("count = %d, want %d", count, tc.wantCount)
			}
		})
	}
}

func TestGetTopicByID_Visibility(t *testing.T) {
	s := setupTopicStore(t)
	privateID := seedVisibilityTopics(t, s)

	if _, err := s.GetTopicByID(context.Background(), privateID, nil); !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("anonymous private topic: err = %v, want ErrTopicNotFound", err)
	}
	if _, err := s.GetTopicByID(context.Background(), privateID, strPtr("u3")); !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("non-allowed follower private topic: err = %v, want ErrTopicNotFound", err)
	}
	allowed, err := s.GetTopicByID(context.Background(), privateID, strPtr("u2"))
	if err != nil {
		t.Fatalf("allowed follower private topic: %v", err)
	}
	if allowed.Title != "Private" {
		t.Errorf("Title = %q, want %q", allowed.Title, "Private")
	}
	if len(allowed.AllowedUsers) != 0 {
		t.Errorf("AllowedUsers for non-owner = %v, want none", allowed.AllowedUsers)
	}
	owner, err := s.GetTopicByID(context.Background(), privateID, strPtr("u1"))
	if err != nil {
		t.Fatalf("owner private topic: %v", err)
	}
	if len(owner.AllowedUsers) != 1 || owner.AllowedUsers[0] != "u2" {
		t.Errorf("owner AllowedUsers = %v, want [u2]", owner.AllowedUsers)
	}
}

func TestPrivateAccessRevokedOnUnfollow(t *testing.T) {
	s := setupTopicStore(t)
	privateID := seedVisibilityTopics(t, s)

	if _, err := s.GetTopicByID(context.Background(), privateID, strPtr("u2")); err != nil {
		t.Fatalf("allowed follower should see private topic: %v", err)
	}

	if _, err := s.db.ExecContext(context.Background(),
		`DELETE FROM follows WHERE follower_id = 'u2' AND followee_id = 'u1'`); err != nil {
		t.Fatalf("delete follow: %v", err)
	}
	if _, err := s.GetTopicByID(context.Background(), privateID, strPtr("u2")); !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("private topic after unfollow: err = %v, want ErrTopicNotFound", err)
	}
	topics, count, err := s.GetFeed(context.Background(), "u2", 1, 10, "created_at", "DESC", "")
	if err != nil {
		t.Fatalf("GetFeed after unfollow: %v", err)
	}
	assertTitles(t, topicTitles(topics), "Public")
	if count != 1 {
		t.Errorf("feed count after unfollow = %d, want 1", count)
	}

	if _, err = s.db.ExecContext(context.Background(),
		`INSERT INTO follows (follower_id, followee_id) VALUES ('u2', 'u1')`); err != nil {
		t.Fatalf("insert follow: %v", err)
	}
	if _, err = s.GetTopicByID(context.Background(), privateID, strPtr("u2")); err != nil {
		t.Errorf("private topic after re-follow: %v, want nil", err)
	}
	topics, count, err = s.GetFeed(context.Background(), "u2", 1, 10, "created_at", "DESC", "")
	if err != nil {
		t.Fatalf("GetFeed after re-follow: %v", err)
	}
	assertTitles(t, topicTitles(topics), "Private", "Followers", "Public")
	if count != 3 {
		t.Errorf("feed count after re-follow = %d, want 3", count)
	}
}

// TestPrivateAuthorPublicPostsVisibleToNonFollowers verifies that public posts
// are always visible regardless of the author's profile privacy setting.
// Followers-only and private posts remain restricted to followers.
func TestPrivateAuthorPublicPostsVisibleToNonFollowers(t *testing.T) {
	s := setupTopicStore(t)
	seedVisibilityTopics(t, s)

	// Make author u1 private. u2 and u3 follow u1; a new user "stranger" does not.
	if _, err := s.db.ExecContext(context.Background(),
		`UPDATE users SET is_private = 1 WHERE id = 'u1'`); err != nil {
		t.Fatalf("set private: %v", err)
	}

	// Non-follower sees only the public post from the private-profile author.
	topics, count, err := s.GetTopicsByUserID(context.Background(), "u1", "stranger", 1, 10)
	if err != nil {
		t.Fatalf("GetTopicsByUserID stranger: %v", err)
	}
	assertTitles(t, topicTitles(topics), "Public")
	if count != 1 {
		t.Errorf("stranger profile view = %d, want 1", count)
	}

	// Non-follower sees the public post in the feed.
	feedTopics, _, err := s.GetFeed(context.Background(), "stranger", 1, 10, "created_at", "DESC", "")
	if err != nil {
		t.Fatalf("GetFeed stranger: %v", err)
	}
	var seen []string
	for _, tpc := range feedTopics {
		if tpc.UserID == "u1" {
			seen = append(seen, tpc.Title)
		}
	}
	assertTitles(t, seen, "Public")

	// Follower still sees public + followers-only posts (but not private unless allowed).
	topics, count, err = s.GetTopicsByUserID(context.Background(), "u1", "u3", 1, 10)
	if err != nil {
		t.Fatalf("GetTopicsByUserID follower: %v", err)
	}
	assertTitles(t, topicTitles(topics), "Followers", "Public")
	if count != 2 {
		t.Errorf("follower profile view count = %d, want 2", count)
	}

	// Owner always sees everything.
	topics, count, err = s.GetTopicsByUserID(context.Background(), "u1", "u1", 1, 10)
	if err != nil {
		t.Fatalf("GetTopicsByUserID owner: %v", err)
	}
	assertTitles(t, topicTitles(topics), "Private", "Followers", "Public")
	if count != 3 {
		t.Errorf("owner profile view count = %d, want 3", count)
	}
}

func TestDeleteVote_Nonexistent(t *testing.T) {
	s := setupTopicStore(t)

	top := &topic.Topic{UserID: "u1", Title: "Vote", Content: "x"}
	_ = s.CreateTopic(context.Background(), top, nil)

	err := s.DeleteVote(context.Background(), "u1", top.ID)
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("DeleteVote no vote: err = %v, want ErrTopicNotFound", err)
	}
}
