package transport

import (
	"context"
	"net/http"

	"social-network/internal/event"
	"social-network/internal/event/commands"
	"social-network/internal/event/queries"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

type UserLookup interface {
	GetUserByID(ctx context.Context, id string) (*UserResult, error)
}

type CreateEventExecutor interface {
	Execute(ctx context.Context, cmd commands.CreateEventCommand) (*event.Event, []event.Option, error)
}

type RSVPExecutor interface {
	Execute(ctx context.Context, cmd commands.RSVPCommand) error
}

type ListGroupEventsResolver interface {
	Resolve(ctx context.Context, q queries.ListGroupEventsQuery) ([]queries.EventWithOptions, string, error)
}

type Handler struct {
	createEvent     CreateEventExecutor
	rsvp            RSVPExecutor
	listGroupEvents ListGroupEventsResolver
	userLookup      UserLookup
	extractUser     UserExtractor
}

func NewHandler(
	extractUser UserExtractor,
	userLookup UserLookup,
	createEvent CreateEventExecutor,
	rsvp RSVPExecutor,
	listGroupEvents ListGroupEventsResolver,
) *Handler {
	return &Handler{
		createEvent:     createEvent,
		rsvp:            rsvp,
		listGroupEvents: listGroupEvents,
		userLookup:      userLookup,
		extractUser:     extractUser,
	}
}
