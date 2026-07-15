package queries_test

import (
	"context"
	"testing"
)

// Contract tests verify the behavioral requirements that any implementation
// of the cross-slice interfaces must satisfy. When a new slice implements
// one of these interfaces, add a test case here that calls the factory
// and runs the contract against the real implementation.

// --- FollowChecker contract ---

type followCheckerContract struct {
	name  string
	build func(t *testing.T) FollowChecker
}

type FollowChecker interface {
	IsFollowing(ctx context.Context, followerID, targetID string) (bool, error)
}

func TestFollowCheckerContract(t *testing.T) {
	contracts := []followCheckerContract{
		// Sprint 3: add real implementation
		// {name: "sqlite", build: func(t *testing.T) FollowChecker { ... }},
	}

	for _, c := range contracts {
		t.Run(c.name, func(t *testing.T) {
			fc := c.build(t)

			following, err := fc.IsFollowing(context.Background(), "nonexistent-follower", "nonexistent-target")
			if err != nil {
				t.Fatalf("IsFollowing() error = %v", err)
			}
			if following {
				t.Error("IsFollowing() = true for nonexistent users, want false")
			}
		})
	}
}

// --- FollowCounter contract ---

type followCounterContract struct {
	name  string
	build func(t *testing.T) FollowCounter
}

type FollowCounter interface {
	GetFollowerCount(ctx context.Context, userID string) (int, error)
	GetFollowingCount(ctx context.Context, userID string) (int, error)
}

func TestFollowCounterContract(t *testing.T) {
	contracts := []followCounterContract{
		// Sprint 3: add real implementation
	}

	for _, c := range contracts {
		t.Run(c.name, func(t *testing.T) {
			fc := c.build(t)

			count, err := fc.GetFollowerCount(context.Background(), "nonexistent-user")
			if err != nil {
				t.Fatalf("GetFollowerCount() error = %v", err)
			}
			if count != 0 {
				t.Errorf("GetFollowerCount() = %d, want 0 for nonexistent user", count)
			}

			count, err = fc.GetFollowingCount(context.Background(), "nonexistent-user")
			if err != nil {
				t.Fatalf("GetFollowingCount() error = %v", err)
			}
			if count != 0 {
				t.Errorf("GetFollowingCount() = %d, want 0 for nonexistent user", count)
			}
		})
	}
}

// --- PostCounter contract ---

type postCounterContract struct {
	name  string
	build func(t *testing.T) PostCounter
}

type PostCounter interface {
	GetPostCount(ctx context.Context, userID string) (int, error)
}

func TestPostCounterContract(t *testing.T) {
	contracts := []postCounterContract{
		// Sprint 4: add real implementation
	}

	for _, c := range contracts {
		t.Run(c.name, func(t *testing.T) {
			pc := c.build(t)

			count, err := pc.GetPostCount(context.Background(), "nonexistent-user")
			if err != nil {
				t.Fatalf("GetPostCount() error = %v", err)
			}
			if count != 0 {
				t.Errorf("GetPostCount() = %d, want 0 for nonexistent user", count)
			}
		})
	}
}

// --- CommentCounter contract ---

type commentCounterContract struct {
	name  string
	build func(t *testing.T) CommentCounter
}

type CommentCounter interface {
	GetCommentCount(ctx context.Context, userID string) (int, error)
}

func TestCommentCounterContract(t *testing.T) {
	contracts := []commentCounterContract{
		// Sprint 5: add real implementation
	}

	for _, c := range contracts {
		t.Run(c.name, func(t *testing.T) {
			cc := c.build(t)

			count, err := cc.GetCommentCount(context.Background(), "nonexistent-user")
			if err != nil {
				t.Fatalf("GetCommentCount() error = %v", err)
			}
			if count != 0 {
				t.Errorf("GetCommentCount() = %d, want 0 for nonexistent user", count)
			}
		})
	}
}

// --- VoteCounter contract ---

type voteCounterContract struct {
	name  string
	build func(t *testing.T) VoteCounter
}

type VoteCounter interface {
	GetVoteCount(ctx context.Context, userID string) (int, error)
}

func TestVoteCounterContract(t *testing.T) {
	contracts := []voteCounterContract{
		// Sprint 4: add real implementation
	}

	for _, c := range contracts {
		t.Run(c.name, func(t *testing.T) {
			vc := c.build(t)

			count, err := vc.GetVoteCount(context.Background(), "nonexistent-user")
			if err != nil {
				t.Fatalf("GetVoteCount() error = %v", err)
			}
			if count != 0 {
				t.Errorf("GetVoteCount() = %d, want 0 for nonexistent user", count)
			}
		})
	}
}
