package bootstrap

import (
	"context"
	"database/sql"
	"net/http"

	"social-network/internal/core/middleware"
	"social-network/internal/event"
	eventcommands "social-network/internal/event/commands"
	eventqueries "social-network/internal/event/queries"
	eventstore "social-network/internal/event/store"
	eventtransport "social-network/internal/event/transport"
	"social-network/internal/user"
	userstore "social-network/internal/user/store"
)

func initEvent(db *sql.DB) *eventtransport.Handler {
	store := eventstore.NewSQLiteStore(db)
	bus := &eventEventBus{}
	groupMember := &groupMemberChecker{db: db}

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	userLookup := &eventUserLookupAdapter{repo: userstore.NewSQLiteStore(db)}

	return eventtransport.NewHandler(
		extractUser,
		userLookup,
		eventcommands.NewCreateEventHandler(store, groupMember, bus),
		eventcommands.NewRSVPHandler(store),
		eventqueries.NewListGroupEventsResolver(store),
	)
}

type eventEventBus struct{}

func (b *eventEventBus) Publish(_ context.Context, _ string, _ any) error {
	return nil
}

type groupMemberChecker struct {
	db *sql.DB
}

func (c *groupMemberChecker) IsMember(ctx context.Context, groupID, userID string) (bool, error) {
	var exists bool
	err := c.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)`,
		groupID, userID,
	).Scan(&exists)
	return exists, err
}

type eventUserLookupAdapter struct {
	repo user.Repository
}

func (a *eventUserLookupAdapter) GetUserByID(ctx context.Context, id string) (*eventtransport.UserResult, error) {
	u, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	result := &eventtransport.UserResult{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Nickname,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  u.Nickname,
		AboutMe:   u.AboutMe,
		IsPublic:  !u.IsPrivate,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if !u.DateOfBirth.IsZero() {
		result.DateOfBirth = u.DateOfBirth.Format("2006-01-02")
	}
	if u.AvatarPath != "" {
		result.AvatarURL = u.AvatarPath
	}
	return result, nil
}

var (
	_ event.Bus                        = (*eventEventBus)(nil)
	_ eventcommands.GroupMemberChecker = (*groupMemberChecker)(nil)
	_ eventtransport.UserLookup        = (*eventUserLookupAdapter)(nil)
)
