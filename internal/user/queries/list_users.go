package queries

import (
	"context"

	"social-network/internal/user"
)

type ListUsersQuery struct {
	Query         string
	Page          int
	Limit         int
	ExcludeUserID string
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
	SearchUsersExcluding(ctx context.Context, query, excludeUserID string, limit, offset int) ([]user.User, error)
	CountUsersExcluding(ctx context.Context, query, excludeUserID string) (int, error)
}

type ListUsersResolver struct {
	repo     SearchRepository
	isOnline func(string) bool
}

func NewListUsersResolver(repo SearchRepository, isOnline func(string) bool) *ListUsersResolver {
	return &ListUsersResolver{repo: repo, isOnline: isOnline}
}

func (r *ListUsersResolver) Resolve(ctx context.Context, q ListUsersQuery) (*ListUsersResult, error) {
	page := max(q.Page, 1)
	limit := max(q.Limit, 1)
	offset := (page - 1) * limit

	var users []user.User
	var total int
	var err error
	if q.ExcludeUserID != "" {
		users, err = r.repo.SearchUsersExcluding(ctx, q.Query, q.ExcludeUserID, limit, offset)
	} else {
		users, err = r.repo.SearchUsers(ctx, q.Query, limit, offset)
	}
	if err != nil {
		return nil, err
	}

	if q.ExcludeUserID != "" {
		total, err = r.repo.CountUsersExcluding(ctx, q.Query, q.ExcludeUserID)
	} else {
		total, err = r.repo.CountUsers(ctx, q.Query)
	}
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
			IsOnline:    r.isOnline != nil && r.isOnline(u.ID),
			CreatedAt:   u.CreatedAt,
		})
	}

	return &ListUsersResult{Users: safe, Total: total}, nil
}
