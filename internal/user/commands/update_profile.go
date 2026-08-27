package commands

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

var ErrUnauthorized = errors.New("unauthorized to update this profile")

type UpdateProfileCommand struct {
	UserID      string
	FirstName   string
	LastName    string
	Nickname    string
	AboutMe     string
	DateOfBirth time.Time
	Gender      string
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

	gender := strings.TrimSpace(cmd.Gender)
	if gender != "" && !allowedGenders[gender] {
		return ErrInvalidGender
	}

	u.FirstName = cmd.FirstName
	u.LastName = cmd.LastName
	u.Nickname = cmd.Nickname
	u.AboutMe = cmd.AboutMe
	u.DateOfBirth = cmd.DateOfBirth
	u.Gender = gender

	if err = h.repo.Update(ctx, u); err != nil {
		return err
	}

	body, _ := json.Marshal(eventbus.Notification{
		Type:         eventbus.EventProfileUpdate,
		RecipientID:  cmd.UserID,
		ActorID:      cmd.UserID,
		ActorName:    u.Nickname,
		ActorAvatar:  u.AvatarPath,
		ResourceType: eventbus.ResourceUser,
		ResourceID:   cmd.UserID,
	})
	return h.bus.Publish("notifications.exchange", eventbus.RoutingUpdated, body)
}
