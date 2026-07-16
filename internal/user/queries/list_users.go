package queries

import (
	"context"

	"social-network/internal/user"
)

type ListUsersQuery struct{}

type ListUsersResult struct {
	Users []user.User
}

type ListUsersResolver struct {
	repo user.Repository
}

func NewListUsersResolver(repo user.Repository) *ListUsersResolver {
	return &ListUsersResolver{repo: repo}
}

func (r *ListUsersResolver) Resolve(ctx context.Context, _ ListUsersQuery) (*ListUsersResult, error) {
	users, err := r.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	safe := make([]user.User, 0, len(users))
	for _, u := range users {
		safe = append(safe, user.User{
			ID:         u.ID,
			Email:      u.Email,
			FirstName:  u.FirstName,
			LastName:   u.LastName,
			Nickname:   u.Nickname,
			AvatarPath: u.AvatarPath,
			CreatedAt:  u.CreatedAt,
		})
	}

	return &ListUsersResult{Users: safe}, nil
}
