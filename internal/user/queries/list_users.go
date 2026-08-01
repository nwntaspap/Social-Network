package queries

import (
	"context"

	"social-network/internal/user"
)

type ListUsersQuery struct {
	Query string
	Page  int
	Limit int
}

type ListUsersResult struct {
	Users []user.User
	Total int
}

// SearchRepository is a local interface for paginated user search,
// implemented by store/sqlite.go.
type SearchRepository interface {
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]user.User, error)
	CountUsers(ctx context.Context, query string) (int, error)
}

type ListUsersResolver struct {
	repo SearchRepository
}

func NewListUsersResolver(repo SearchRepository) *ListUsersResolver {
	return &ListUsersResolver{repo: repo}
}

func (r *ListUsersResolver) Resolve(ctx context.Context, q ListUsersQuery) (*ListUsersResult, error) {
	page := max(q.Page, 1)
	limit := max(q.Limit, 1)
	offset := (page - 1) * limit

	users, err := r.repo.SearchUsers(ctx, q.Query, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := r.repo.CountUsers(ctx, q.Query)
	if err != nil {
		return nil, err
	}

	safe := make([]user.User, 0, len(users))
	for _, u := range users {
		safe = append(safe, user.User{
			ID:          u.ID,
			Email:       u.Email,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			Nickname:    u.Nickname,
			AboutMe:     u.AboutMe,
			AvatarPath:  u.AvatarPath,
			DateOfBirth: u.DateOfBirth,
			IsPrivate:   u.IsPrivate,
			CreatedAt:   u.CreatedAt,
		})
	}

	return &ListUsersResult{Users: safe, Total: total}, nil
}
