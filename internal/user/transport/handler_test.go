package transport

import (
	"context"
	"net/http"
	"testing"
	"time"

	"social-network/internal/user"
	"social-network/internal/user/commands"
	"social-network/internal/user/queries"
)

type stubRegister struct {
	user *user.User
	err  error
	got  commands.RegisterCommand
}

func (s *stubRegister) Execute(_ context.Context, cmd commands.RegisterCommand) (*user.User, error) {
	s.got = cmd
	return s.user, s.err
}

type stubLogin struct {
	result *commands.LoginResult
	err    error
}

func (s *stubLogin) Execute(_ context.Context, _ commands.LoginCommand) (*commands.LoginResult, error) {
	return s.result, s.err
}

type stubLogout struct {
	err error
}

func (s *stubLogout) Execute(_ context.Context, _ commands.LogoutCommand) error {
	return s.err
}

type stubUpdateProfile struct {
	err error
}

func (s *stubUpdateProfile) Execute(_ context.Context, _ commands.UpdateProfileCommand) error {
	return s.err
}

type stubTogglePrivacy struct {
	err error
}

func (s *stubTogglePrivacy) Execute(_ context.Context, _ commands.TogglePrivacyCommand) error {
	return s.err
}

type stubGetProfile struct {
	result *queries.ProfileResult
	err    error
}

func (s *stubGetProfile) Resolve(_ context.Context, _ queries.GetProfileQuery) (*queries.ProfileResult, error) {
	return s.result, s.err
}

type stubGetActivity struct {
	result *queries.ActivityResult
	err    error
}

func (s *stubGetActivity) Resolve(_ context.Context, _ queries.GetActivityQuery) (*queries.ActivityResult, error) {
	return s.result, s.err
}

type stubListUsers struct {
	result *queries.ListUsersResult
	err    error
}

func (s *stubListUsers) Resolve(_ context.Context, _ queries.ListUsersQuery) (*queries.ListUsersResult, error) {
	return s.result, s.err
}

type stubAuth struct {
	userID string
	ok     bool
}

func (s *stubAuth) Extract(_ *http.Request) (string, bool) {
	return s.userID, s.ok
}

type stubCookieWriter struct {
	setToken     string
	setExpiry    time.Time
	setCalled    bool
	deleteCalled bool
}

func (s *stubCookieWriter) SetAccessCookie(_ http.ResponseWriter, token string, expiresAt time.Time) {
	s.setToken = token
	s.setExpiry = expiresAt
	s.setCalled = true
}

func (s *stubCookieWriter) DeleteAccessCookie(_ http.ResponseWriter) {
	s.deleteCalled = true
}

func newTestHandler(opts ...func(*Handler)) *Handler {
	h := &Handler{}
	for _, o := range opts {
		o(h)
	}
	return h
}

func withDefaults(h *Handler) {
	if h.auth == nil {
		h.auth = &stubAuth{ok: true}
	}
	if h.register == nil {
		h.register = &stubRegister{}
	}
	if h.login == nil {
		h.login = &stubLogin{}
	}
	if h.logout == nil {
		h.logout = &stubLogout{}
	}
	if h.updateProfile == nil {
		h.updateProfile = &stubUpdateProfile{}
	}
	if h.togglePrivacy == nil {
		h.togglePrivacy = &stubTogglePrivacy{}
	}
	if h.getProfile == nil {
		h.getProfile = &stubGetProfile{}
	}
	if h.getActivity == nil {
		h.getActivity = &stubGetActivity{}
	}
	if h.listUsers == nil {
		h.listUsers = &stubListUsers{}
	}
	if h.sessionCookies == nil {
		h.sessionCookies = &stubCookieWriter{}
	}
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(
		&stubAuth{},
		&stubRegister{}, &stubLogin{}, &stubLogout{},
		&stubUpdateProfile{}, &stubTogglePrivacy{},
		&stubGetProfile{}, &stubGetActivity{}, &stubListUsers{},
		&stubCookieWriter{},
	)
	if h == nil {
		t.Fatal("NewHandler() returned nil")
	}
}
