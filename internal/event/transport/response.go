package transport

import (
	"time"

	"social-network/internal/event"
	"social-network/internal/event/queries"
)

type UserResult struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Nickname    string `json:"nickname,omitempty"`
	AboutMe     string `json:"aboutMe,omitempty"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	DateOfBirth string `json:"dateOfBirth"`
	IsPublic    bool   `json:"isPublic"`
	CreatedAt   string `json:"createdAt"`
}

type OptionResponse struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Tally int    `json:"tally"`
}

type EventResponse struct {
	ID          string           `json:"id"`
	GroupID     string           `json:"groupId"`
	CreatorID   string           `json:"creatorId"`
	Creator     *UserResult      `json:"creator"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	EventDate   string           `json:"eventDate"`
	CreatedAt   string           `json:"createdAt"`
	Options     []OptionResponse `json:"options"`
}

func toEventResponse(e *event.Event, creator *UserResult, opts []queries.OptionWithTally) EventResponse {
	optResp := make([]OptionResponse, len(opts))
	for i, o := range opts {
		optResp[i] = OptionResponse{
			ID:    o.ID,
			Label: o.Label,
			Tally: o.Tally,
		}
	}
	return EventResponse{
		ID:          e.ID,
		GroupID:     e.GroupID,
		CreatorID:   e.CreatorID,
		Creator:     creator,
		Title:       e.Title,
		Description: e.Description,
		EventDate:   formatTime(e.ScheduledTime),
		CreatedAt:   formatTime(e.CreatedAt),
		Options:     optResp,
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
