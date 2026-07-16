package queries

import (
	"context"

	"social-network/internal/user"
)

type ActivityResult struct {
	User           user.User
	PostCount      int
	CommentCount   int
	VoteCount      int
	FollowerCount  int
	FollowingCount int
}

type GetActivityQuery struct {
	UserID string
}

type GetActivityResolver struct {
	repo           user.Repository
	postCounter    PostCounter
	commentCounter CommentCounter
	voteCounter    VoteCounter
	followCounter  FollowCounter
}

func NewGetActivityResolver(
	repo user.Repository,
	pc PostCounter,
	cc CommentCounter,
	vc VoteCounter,
	fc FollowCounter,
) *GetActivityResolver {
	return &GetActivityResolver{
		repo:           repo,
		postCounter:    pc,
		commentCounter: cc,
		voteCounter:    vc,
		followCounter:  fc,
	}
}

func (r *GetActivityResolver) Resolve(ctx context.Context, q GetActivityQuery) (*ActivityResult, error) {
	u, err := r.repo.GetByID(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	postCount, err := r.postCounter.GetPostCount(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	commentCount, err := r.commentCounter.GetCommentCount(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	voteCount, err := r.voteCounter.GetVoteCount(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	followerCount, err := r.followCounter.GetFollowerCount(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	followingCount, err := r.followCounter.GetFollowingCount(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	return &ActivityResult{
		User:           *u,
		PostCount:      postCount,
		CommentCount:   commentCount,
		VoteCount:      voteCount,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
	}, nil
}
