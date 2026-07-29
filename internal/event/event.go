package event

import (
	"context"
	"errors"
	"time"
)

var (
	ErrEventNotFound  = errors.New("event not found")
	ErrOptionNotFound = errors.New("event option not found")
	ErrRSVPNotFound   = errors.New("event RSVP not found")
)

type Event struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"group_id"`
	CreatorID     string    `json:"creator_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ScheduledTime time.Time `json:"scheduled_time"`
	CreatedAt     time.Time `json:"created_at"`
}

type Option struct {
	ID      string
	EventID string
	Label   string
}

type RSVP struct {
	EventID   string
	UserID    string
	OptionID  string
	UpdatedAt time.Time
}

type Repository interface {
	Repo
	OptionRepo
	RSVPRepo
}

type Repo interface {
	CreateEvent(ctx context.Context, e *Event) error
	GetEvent(ctx context.Context, eventID string) (*Event, error)
	UpdateEvent(ctx context.Context, e *Event) error
	DeleteEvent(ctx context.Context, eventID string) error
	ListGroupEvents(ctx context.Context, groupID, cursor string, size int) ([]Event, string, error)
}

type OptionRepo interface {
	CreateOptions(ctx context.Context, opts []Option) error
	GetOptionsByEvent(ctx context.Context, eventID string) ([]Option, error)
	GetOption(ctx context.Context, optionID string) (*Option, error)
}

type RSVPRepo interface {
	UpsertRSVP(ctx context.Context, rsvp *RSVP) error
	GetRSVPsByEvent(ctx context.Context, eventID string) ([]RSVP, error)
	GetUserRSVP(ctx context.Context, eventID, userID string) (*RSVP, error)
}
