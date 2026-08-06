package transport

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"social-network/internal/event/commands"
	"social-network/internal/event/queries"
	"social-network/internal/pkg/helpers"
)

type createEventBody struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	EventDate   string   `json:"eventDate"`
	Options     []string `json:"options"`
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	groupID, ok := requirePathParam(w, r, "groupId", "Group ID")
	if !ok {
		return
	}

	var body createEventBody
	if _, err := helpers.ParseBodyRequest(r, &body); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	eventTime, err := time.Parse(time.RFC3339, body.EventDate)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid event date format, use RFC3339")
		return
	}

	cmd := commands.CreateEventCommand{
		UserID:        userID,
		GroupID:       groupID,
		Title:         body.Title,
		Description:   body.Description,
		ScheduledTime: eventTime,
		Options:       body.Options,
	}

	e, opts, err := h.createEvent.Execute(r.Context(), cmd)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	creator := h.lookupUser(r.Context(), e.CreatorID)
	optResp := make([]OptionResponse, len(opts))
	for i, o := range opts {
		optResp[i] = OptionResponse{ID: o.ID, Label: o.Label}
	}
	resp := EventResponse{
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
	helpers.RespondWithJSON(w, http.StatusCreated, nil, resp)
}

type updateEventBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"eventDate"`
}

func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	groupID, ok := requirePathParam(w, r, "groupId", "Group ID")
	if !ok {
		return
	}
	eventID, ok := requirePathParam(w, r, "eventId", "Event ID")
	if !ok {
		return
	}

	var body updateEventBody
	if _, err := helpers.ParseBodyRequest(r, &body); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	eventTime, err := time.Parse(time.RFC3339, body.EventDate)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid event date format, use RFC3339")
		return
	}

	cmd := commands.UpdateEventCommand{
		UserID:        userID,
		GroupID:       groupID,
		EventID:       eventID,
		Title:         body.Title,
		Description:   body.Description,
		ScheduledTime: eventTime,
	}

	e, opts, err := h.updateEvent.Execute(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, commands.ErrNotGroupCreator) {
			helpers.RespondWithError(w, http.StatusForbidden, err.Error())
			return
		}
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	creator := h.lookupUser(r.Context(), e.CreatorID)
	helpers.RespondWithJSON(w, http.StatusOK, nil, toEventResponse(e, creator, toOptionsWithTally(opts)))
}

func (h *Handler) ListGroupEvents(w http.ResponseWriter, r *http.Request) {
	groupID, ok := requirePathParam(w, r, "groupId", "Group ID")
	if !ok {
		return
	}

	cursor := r.URL.Query().Get("cursor")
	sizeStr := r.URL.Query().Get("size")
	size := 10
	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			size = s
		}
	}

	q := queries.ListGroupEventsQuery{
		GroupID: groupID,
		Cursor:  cursor,
		Size:    size,
	}

	events, nextCursor, err := h.listGroupEvents.Resolve(r.Context(), q)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]EventResponse, len(events))
	for i, ew := range events {
		creator := h.lookupUser(r.Context(), ew.CreatorID)
		resp[i] = toEventResponse(&ew.Event, creator, ew.Options)
	}

	payload := map[string]any{
		"events": resp,
	}
	if nextCursor != "" {
		payload["nextCursor"] = nextCursor
	}
	helpers.RespondWithJSON(w, http.StatusOK, nil, payload)
}

type respondEventBody struct {
	Response string `json:"response"`
}

func (h *Handler) RespondToEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	eventID, ok := requirePathParam(w, r, "eventId", "Event ID")
	if !ok {
		return
	}

	var body respondEventBody
	if _, err := helpers.ParseBodyRequest(r, &body); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	optionID := strings.TrimSpace(body.Response)
	if optionID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "response is required")
		return
	}

	cmd := commands.RSVPCommand{
		EventID:  eventID,
		UserID:   userID,
		OptionID: optionID,
	}
	if err := h.rsvp.Execute(r.Context(), cmd); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"status": "ok"})
}

func (h *Handler) ListEventResponders(w http.ResponseWriter, r *http.Request) {
	eventID, ok := requirePathParam(w, r, "eventId", "Event ID")
	if !ok {
		return
	}

	options, err := h.listEventRSVPs.Resolve(r.Context(), queries.ListEventRSVPsQuery{EventID: eventID})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := make([]EventRSVPOptionResponse, len(options))
	for i, o := range options {
		users := make([]*UserResult, 0, len(o.UserIDs))
		for _, uid := range o.UserIDs {
			if u := h.lookupUser(r.Context(), uid); u != nil {
				users = append(users, u)
			}
		}
		resp[i] = EventRSVPOptionResponse{OptionID: o.OptionID, OptionLabel: o.Label, Users: users}
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]any{"options": resp})
}

func (h *Handler) lookupUser(ctx context.Context, userID string) *UserResult {
	if h.userLookup == nil || userID == "" {
		return nil
	}
	u, err := h.userLookup.GetUserByID(ctx, userID)
	if err != nil {
		return nil
	}
	return u
}

func requirePathParam(w http.ResponseWriter, r *http.Request, name, label string) (string, bool) {
	val := r.PathValue(name)
	if val == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, label+" is required")
		return "", false
	}
	return val, true
}
