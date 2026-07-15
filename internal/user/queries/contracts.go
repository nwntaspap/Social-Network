package queries

import "context"

// FollowChecker determines whether followerID follows targetID.
// Contract: implementors must return false, nil for nonexistent users.
// Sprint 3: wire to follow.Repository.
type FollowChecker interface {
	IsFollowing(ctx context.Context, followerID, targetID string) (bool, error)
}

// FollowCounter returns follower/following counts for a user.
// Contract: implementors must return 0 for users with no followers/following.
// Sprint 3: wire to follow.Repository.
type FollowCounter interface {
	GetFollowerCount(ctx context.Context, userID string) (int, error)
	GetFollowingCount(ctx context.Context, userID string) (int, error)
}

// PostCounter returns the number of posts created by a user.
// Contract: implementors must return 0 for users with no posts, not an error.
// Sprint 4: wire to topic.Repository.
type PostCounter interface {
	GetPostCount(ctx context.Context, userID string) (int, error)
}

// CommentCounter returns the number of comments created by a user.
// Contract: implementors must return 0 for users with no comments, not an error.
// Sprint 5: wire to comment.Repository.
type CommentCounter interface {
	GetCommentCount(ctx context.Context, userID string) (int, error)
}

// VoteCounter returns the number of votes cast by a user.
// Contract: implementors must return 0 for users with no votes, not an error.
// Sprint 4: wire to topic.Repository.
type VoteCounter interface {
	GetVoteCount(ctx context.Context, userID string) (int, error)
}
