package commands

import (
	"context"
	"errors"

	"social-network/internal/core/session"
	"social-network/internal/pkg/bcrypt"
	"social-network/internal/user"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type LoginCommand struct {
	Identifier string
	Password   string
}

type LoginResult struct {
	User  *user.User
	Token string
}

type LoginHandler struct {
	repo     user.Repository
	crypto   bcrypt.Provider
	sessions session.Manager
}

func NewLoginHandler(repo user.Repository, crypto bcrypt.Provider, sessions session.Manager) *LoginHandler {
	return &LoginHandler{
		repo:     repo,
		crypto:   crypto,
		sessions: sessions,
	}
}

func (h *LoginHandler) Execute(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	u, err := h.repo.GetByEmail(ctx, cmd.Identifier)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			u, err = h.repo.GetByUsername(ctx, cmd.Identifier)
		}
		if err != nil {
			return nil, ErrInvalidCredentials
		}
	}

	if matchErr := h.crypto.Matches(u.PasswordHash, cmd.Password); matchErr != nil {
		return nil, ErrInvalidCredentials
	}

	sess, err := h.sessions.Create(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:  u,
		Token: sess.Token,
	}, nil
}
