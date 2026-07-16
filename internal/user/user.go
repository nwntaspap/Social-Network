package user

import (
	"context"
	"errors"
	"time"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	DateOfBirth  time.Time
	Nickname     string
	AboutMe      string
	AvatarPath   string
	IsPrivate    bool
	CreatedAt    time.Time
}

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, u *User) error
	TogglePrivacy(ctx context.Context, id string, isPrivate bool) error
	ListAll(ctx context.Context) ([]User, error)
}
