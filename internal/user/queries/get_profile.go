package queries

import (
	"context"

	"social-network/internal/user"
)

type ProfileResult struct {
	User           user.User
	FollowerCount  int
	FollowingCount int
}

type GetProfileQuery struct {
	TargetID    string
	RequesterID string
}

type GetProfileResolver struct {
	repo          user.Repository
	followChecker FollowChecker
	followCounter FollowCounter
}

func NewGetProfileResolver(repo user.Repository, fc FollowChecker, counter FollowCounter) *GetProfileResolver {
	return &GetProfileResolver{
		repo:          repo,
		followChecker: fc,
		followCounter: counter,
	}
}

func (r *GetProfileResolver) Resolve(ctx context.Context, q GetProfileQuery) (*ProfileResult, error) {
	u, err := r.repo.GetByID(ctx, q.TargetID)
	if err != nil {
		return nil, err
	}

	if u.IsPrivate && q.TargetID != q.RequesterID {
		var isFollowing bool
		isFollowing, err = r.followChecker.IsFollowing(ctx, q.RequesterID, q.TargetID)
		if err != nil {
			return nil, err
		}
		if !isFollowing {
			return &ProfileResult{
				User: user.User{
					ID:         u.ID,
					Nickname:   u.Nickname,
					AvatarPath: u.AvatarPath,
				},
			}, nil
		}
	}

	followerCount, err := r.followCounter.GetFollowerCount(ctx, q.TargetID)
	if err != nil {
		return nil, err
	}

	followingCount, err := r.followCounter.GetFollowingCount(ctx, q.TargetID)
	if err != nil {
		return nil, err
	}

	return &ProfileResult{
		User:           *u,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
	}, nil
}
