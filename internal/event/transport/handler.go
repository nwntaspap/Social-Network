package transport

import (
	"context"
	"net/http"

	"social-network/internal/event"
	"social-network/internal/event/commands"
	"social-network/internal/event/queries"
	"social-network/internal/platform/logger"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

type UserLookup interface {
	GetUserByID(ctx context.Context, id string) (*UserResult, error)
}

type CreateEventExecutor interface {
	Execute(ctx context.Context, cmd commands.CreateEventCommand) (*event.Event, []event.Option, error)
}

type UpdateEventExecutor interface {
	Execute(ctx context.Context, cmd commands.UpdateEventCommand) (*event.Event, []event.Option, error)
}

type RSVPExecutor interface {
	Execute(ctx context.Context, cmd commands.RSVPCommand) error
}

type ListGroupEventsResolver interface {
	Resolve(ctx context.Context, q queries.ListGroupEventsQuery) ([]queries.EventWithOptions, string, error)
}

type ListEventRSVPsResolver interface {
	Resolve(ctx context.Context, q queries.ListEventRSVPsQuery) ([]queries.OptionRSVPs, error)
}

type Handler struct {
	createEvent     CreateEventExecutor
	updateEvent     UpdateEventExecutor
	rsvp            RSVPExecutor
	listGroupEvents ListGroupEventsResolver
	listEventRSVPs  ListEventRSVPsResolver
	userLookup      UserLookup
	extractUser     UserExtractor
	logger          logger.Logger
}

func NewHandler(
	extractUser UserExtractor,
	userLookup UserLookup,
	createEvent CreateEventExecutor,
	updateEvent UpdateEventExecutor,
	rsvp RSVPExecutor,
	listGroupEvents ListGroupEventsResolver,
	listEventRSVPs ListEventRSVPsResolver,
	logger logger.Logger,
) *Handler {
	return &Handler{
		createEvent:     createEvent,
		updateEvent:     updateEvent,
		rsvp:            rsvp,
		listGroupEvents: listGroupEvents,
		listEventRSVPs:  listEventRSVPs,
		userLookup:      userLookup,
		extractUser:     extractUser,
		logger:          logger,
	}
}
