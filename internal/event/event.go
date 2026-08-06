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

type Bus interface {
	Publish(ctx context.Context, eventType string, payload any) error
}

type Event struct {
	ID            string
	GroupID       string
	CreatorID     string
	Title         string
	Description   string
	ScheduledTime time.Time
	CreatedAt     time.Time
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
