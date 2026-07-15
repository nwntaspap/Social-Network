package commands

import (
	"context"
	"errors"
	"time"

	"social-network/internal/pkg/bcrypt"
	"social-network/internal/pkg/helpers"
	"social-network/internal/pkg/uuid"
	"social-network/internal/user"
)

var (
	ErrEmailTaken    = errors.New("email already registered")
	ErrUnderage      = errors.New("must be at least 13 years old")
	ErrWeakPassword  = errors.New("password must be at least 8 characters")
	ErrNicknameEmpty = errors.New("nickname is required")
)

type RegisterCommand struct {
	Email       string
	Password    string
	FirstName   string
	LastName    string
	Nickname    string
	DateOfBirth time.Time
}

type RegisterHandler struct {
	repo   user.Repository
	uuid   uuid.Provider
	crypto bcrypt.Provider
}

func NewRegisterHandler(repo user.Repository, uuid uuid.Provider, crypto bcrypt.Provider) *RegisterHandler {
	return &RegisterHandler{
		repo:   repo,
		uuid:   uuid,
		crypto: crypto,
	}
}

func (h *RegisterHandler) Execute(ctx context.Context, cmd RegisterCommand) (*user.User, error) {
	if err := helpers.ValidateEmail(cmd.Email); err != nil {
		return nil, err
	}

	if len(cmd.Password) < 8 {
		return nil, ErrWeakPassword
	}

	if cmd.Nickname == "" {
		return nil, ErrNicknameEmpty
	}

	if time.Since(cmd.DateOfBirth) < 13*365.25*24*time.Hour {
		return nil, ErrUnderage
	}

	existing, err := h.repo.GetByEmail(ctx, cmd.Email)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hashed, err := h.crypto.Generate(cmd.Password)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		ID:           h.uuid.NewUUID(),
		Email:        cmd.Email,
		PasswordHash: hashed,
		FirstName:    cmd.FirstName,
		LastName:     cmd.LastName,
		Nickname:     cmd.Nickname,
		DateOfBirth:  cmd.DateOfBirth,
		CreatedAt:    time.Now(),
	}

	if err := h.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}
