package bootstrap

import (
	"context"
	"net/http"

	"social-network/internal/core/middleware"
	eventcommands "social-network/internal/event/commands"
	eventqueries "social-network/internal/event/queries"
	eventstore "social-network/internal/event/store"
	eventtransport "social-network/internal/event/transport"
	"social-network/internal/platform/database"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
	userstore "social-network/internal/user/store"
)

func initEvent(db database.DB, bus eventbus.EventBus) *eventtransport.Handler {
	store := eventstore.NewSQLiteStore(db)
	groupMember := &groupMemberChecker{db: db}
	users := userstore.NewSQLiteStore(db)

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}
	groupRole := &eventGroupRoleChecker{db: db}
	userLookup := &eventUserLookupAdapter{repo: userstore.NewSQLiteStore(db)}

	return eventtransport.NewHandler(
		extractUser,
		userLookup,
		eventcommands.NewCreateEventHandler(store, groupMember, bus, users),
		eventcommands.NewUpdateEventHandler(store, groupRole),
		eventcommands.NewRSVPHandler(store, bus),
		eventqueries.NewListGroupEventsResolver(store),
		eventqueries.NewListEventRSVPsResolver(store),
	)
}

type groupMemberChecker struct {
	db database.DB
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

type eventGroupRoleChecker struct {
	db database.DB
}

func (c *eventGroupRoleChecker) GetMemberRole(ctx context.Context, groupID, userID string) (string, error) {
	var role string
	err := c.db.QueryRowContext(
		ctx,
		`SELECT role FROM group_members WHERE group_id = ? AND user_id = ?`,
		groupID, userID,
	).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func (c *groupMemberChecker) GetGroupMembers(ctx context.Context, groupID string) ([]string, error) {
	var users []string
	rows, err := c.db.QueryContext(
		ctx,
		`SELECT user_id FROM group_members WHERE group_id = ?)`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		users = append(users, userID)
	}
	return users, rows.Err()
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
	_ eventcommands.GroupMemberChecker = (*groupMemberChecker)(nil)
	_ eventcommands.GroupRoleChecker   = (*eventGroupRoleChecker)(nil)
	_ eventtransport.UserLookup        = (*eventUserLookupAdapter)(nil)
)
