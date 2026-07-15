package commands

import (
	"context"
	"errors"

	"social-network/internal/user"
)

var ErrUnauthorized = errors.New("unauthorized to update this profile")

type UpdateProfileCommand struct {
	UserID    string
	FirstName string
	LastName  string
	Nickname  string
	AboutMe   string
}

type UpdateProfileHandler struct {
	repo user.Repository
}

func NewUpdateProfileHandler(repo user.Repository) *UpdateProfileHandler {
	return &UpdateProfileHandler{repo: repo}
}

func (h *UpdateProfileHandler) Execute(ctx context.Context, cmd UpdateProfileCommand) error {
	u, err := h.repo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	u.FirstName = cmd.FirstName
	u.LastName = cmd.LastName
	u.Nickname = cmd.Nickname
	u.AboutMe = cmd.AboutMe

	return h.repo.Update(ctx, u)
}
