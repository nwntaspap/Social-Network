package commands

import (
	"context"
	"encoding/json"
	"errors"

	"social-network/internal/platform/eventbus"
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
	bus  eventbus.EventBus
}

func NewUpdateProfileHandler(repo user.Repository, bus eventbus.EventBus) *UpdateProfileHandler {
	return &UpdateProfileHandler{
		repo: repo,
		bus:  bus,
	}
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

	if err = h.repo.Update(ctx, u); err != nil {
		return err
	}

	body, _ := json.Marshal(eventbus.Notification{
		Type:        eventbus.ResourceUser,
		ActorID:     cmd.UserID,
		ActorName:   u.Nickname,
		ActorAvatar: u.AvatarPath,
	})
	return h.bus.Publish("notifications.exchange", eventbus.RoutingUpdated, body)
}
