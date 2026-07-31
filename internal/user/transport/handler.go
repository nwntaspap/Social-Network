package transport

import (
	"context"
	"net/http"
	"time"

	"social-network/internal/user"
	"social-network/internal/user/commands"
	"social-network/internal/user/queries"
)

type AuthUserExtractor interface {
	Extract(r *http.Request) (userID string, ok bool)
}

type RegisterExecutor interface {
	Execute(ctx context.Context, cmd commands.RegisterCommand) (*user.User, error)
}

type LoginExecutor interface {
	Execute(ctx context.Context, cmd commands.LoginCommand) (*commands.LoginResult, error)
}

type LogoutExecutor interface {
	Execute(ctx context.Context, cmd commands.LogoutCommand) error
}

type UpdateProfileExecutor interface {
	Execute(ctx context.Context, cmd commands.UpdateProfileCommand) error
}

type TogglePrivacyExecutor interface {
	Execute(ctx context.Context, cmd commands.TogglePrivacyCommand) error
}

type ProfileResolver interface {
	Resolve(ctx context.Context, q queries.GetProfileQuery) (*queries.ProfileResult, error)
}

type ActivityResolver interface {
	Resolve(ctx context.Context, q queries.GetActivityQuery) (*queries.ActivityResult, error)
}

type ListUsersResolver interface {
	Resolve(ctx context.Context, q queries.ListUsersQuery) (*queries.ListUsersResult, error)
}

type SessionCookieWriter interface {
	SetAccessCookie(w http.ResponseWriter, token string, expiresAt time.Time)
	DeleteAccessCookie(w http.ResponseWriter)
}

type Handler struct {
	auth           AuthUserExtractor
	register       RegisterExecutor
	login          LoginExecutor
	logout         LogoutExecutor
	updateProfile  UpdateProfileExecutor
	togglePrivacy  TogglePrivacyExecutor
	getProfile     ProfileResolver
	getActivity    ActivityResolver
	listUsers      ListUsersResolver
	sessionCookies SessionCookieWriter
}

func NewHandler(
	auth AuthUserExtractor,
	register RegisterExecutor,
	login LoginExecutor,
	logout LogoutExecutor,
	updateProfile UpdateProfileExecutor,
	togglePrivacy TogglePrivacyExecutor,
	getProfile ProfileResolver,
	getActivity ActivityResolver,
	listUsers ListUsersResolver,
	sessionCookies SessionCookieWriter,
) *Handler {
	return &Handler{
		auth:           auth,
		register:       register,
		login:          login,
		logout:         logout,
		updateProfile:  updateProfile,
		togglePrivacy:  togglePrivacy,
		getProfile:     getProfile,
		getActivity:    getActivity,
		listUsers:      listUsers,
		sessionCookies: sessionCookies,
	}
}
