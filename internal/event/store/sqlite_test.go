package store

import (
	"errors"
	"testing"
	"time"

	"social-network/internal/event"
	"social-network/internal/platform/database"
)

func setupTestDB(t *testing.T) database.DB {
	t.Helper()
	db, err := database.NewDB(database.Config{Driver: "sqlite3", Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := t.Context()
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, username TEXT, email TEXT, password_hash TEXT, nickname TEXT, first_name TEXT, last_name TEXT, about_me TEXT, avatar_path TEXT, is_private INTEGER DEFAULT 0, date_of_birth TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS groups (id TEXT PRIMARY KEY, title TEXT NOT NULL, description TEXT, creator_id TEXT NOT NULL REFERENCES users(id), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS group_members (group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE, user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE, role TEXT NOT NULL DEFAULT 'member', joined_at DATETIME DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (group_id, user_id))`,
		`CREATE TABLE IF NOT EXISTS events (id TEXT PRIMARY KEY, group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE, creator_id TEXT NOT NULL REFERENCES users(id), title TEXT NOT NULL, description TEXT NOT NULL, event_time TIMESTAMP NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS event_options (id TEXT PRIMARY KEY, event_id TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE, name TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS event_rsvps (event_id TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE, user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE, option_id TEXT NOT NULL REFERENCES event_options(id) ON DELETE CASCADE, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY(event_id, user_id))`,
	}
	for _, stmt := range tables {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("exec DDL: %v", err)
		}
	}
	return db
}

func seedTestData(t *testing.T, db database.DB) (userID, groupID string) {
	t.Helper()
	ctx := t.Context()
	userID = "user-1"
	groupID = "group-1"

	_, err := db.ExecContext(ctx, `INSERT INTO users (id, username, email, password_hash) VALUES (?, ?, ?, ?)`,
		userID, "testuser", "test@example.com", "hash")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	_, err = db.ExecContext(ctx, `INSERT INTO groups (id, title, description, creator_id) VALUES (?, ?, ?, ?)`,
		groupID, "Test Group", "A test group", userID)
	if err != nil {
		t.Fatalf("seed group: %v", err)
	}

	_, err = db.ExecContext(ctx, `INSERT INTO group_members (group_id, user_id, role) VALUES (?, ?, ?)`,
		groupID, userID, "creator")
	if err != nil {
		t.Fatalf("seed member: %v", err)
	}

	return userID, groupID
}

func TestCreateAndGetEvent(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)
	ctx := t.Context()
	userID, groupID := seedTestData(t, db)

	e := &event.Event{
		ID:            "evt-1",
		GroupID:       groupID,
		CreatorID:     userID,
		Title:         "Meetup",
		Description:   "Monthly meetup",
		ScheduledTime: time.Now().Add(24 * time.Hour),
	}

	if err := store.CreateEvent(ctx, e); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	got, err := store.GetEvent(ctx, "evt-1")
	if err != nil {
		t.Fatalf("GetEvent() error = %v", err)
	}
	if got.Title != "Meetup" {
		t.Errorf("Title = %q, want %q", got.Title, "Meetup")
	}
	if got.CreatorID != userID {
		t.Errorf("CreatorID = %q, want %q", got.CreatorID, userID)
	}
}

func TestGetEvent_NotFound(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	_, err := store.GetEvent(t.Context(), "nonexistent")
	if !errors.Is(err, event.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestUpdateEvent(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)
	ctx := t.Context()
	userID, groupID := seedTestData(t, db)

	e := &event.Event{
		ID:            "evt-1",
		GroupID:       groupID,
		CreatorID:     userID,
		Title:         "Meetup",
		Description:   "Monthly meetup",
		ScheduledTime: time.Now().Add(24 * time.Hour),
	}
	if err := store.CreateEvent(ctx, e); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	updated := &event.Event{
		ID:            "evt-1",
		GroupID:       groupID,
		CreatorID:     userID,
		Title:         "Big Meetup",
		Description:   "Annual meetup",
		ScheduledTime: time.Now().Add(48 * time.Hour),
	}
	if err := store.UpdateEvent(ctx, updated); err != nil {
		t.Fatalf("UpdateEvent() error = %v", err)
	}

	got, err := store.GetEvent(ctx, "evt-1")
	if err != nil {
		t.Fatalf("GetEvent() error = %v", err)
	}
	if got.Title != "Big Meetup" {
		t.Errorf("Title = %q, want %q", got.Title, "Big Meetup")
	}
	if got.Description != "Annual meetup" {
		t.Errorf("Description = %q, want %q", got.Description, "Annual meetup")
	}
}

func TestUpdateEvent_NotFound(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	err := store.UpdateEvent(t.Context(), &event.Event{ID: "nonexistent"})
	if !errors.Is(err, event.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestListGroupEvents_CursorPagination(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)
	ctx := t.Context()
	userID, groupID := seedTestData(t, db)

	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := range 5 {
		ts := baseTime.Add(time.Duration(i) * time.Hour)
		_, err := db.ExecContext(
			ctx,
			`INSERT INTO events (id, group_id, creator_id, title, description, event_time, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			"evt-"+string(rune('0'+i)), groupID, userID, "Event", "Desc", ts, ts,
		)
		if err != nil {
			t.Fatalf("insert event: %v", err)
		}
	}

	events, nextCursor, err := store.ListGroupEvents(ctx, groupID, "", 3)
	if err != nil {
		t.Fatalf("ListGroupEvents() error = %v", err)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
	if nextCursor == "" {
		t.Error("expected non-empty nextCursor")
	}

	events2, nextCursor2, err := store.ListGroupEvents(ctx, groupID, nextCursor, 3)
	if err != nil {
		t.Fatalf("ListGroupEvents() page 2 error = %v", err)
	}
	if len(events2) != 2 {
		t.Errorf("expected 2 events on page 2, got %d", len(events2))
	}
	if nextCursor2 != "" {
		t.Error("expected empty nextCursor on last page")
	}
}

func TestCreateAndGetOptions(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)
	ctx := t.Context()
	userID, groupID := seedTestData(t, db)

	e := &event.Event{
		ID:            "evt-1",
		GroupID:       groupID,
		CreatorID:     userID,
		Title:         "Event",
		Description:   "Desc",
		ScheduledTime: time.Now().Add(time.Hour),
	}
	if err := store.CreateEvent(ctx, e); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	opts := []event.Option{
		{ID: "opt-1", EventID: "evt-1", Label: "going"},
		{ID: "opt-2", EventID: "evt-1", Label: "not going"},
	}
	if err := store.CreateOptions(ctx, opts); err != nil {
		t.Fatalf("CreateOptions() error = %v", err)
	}

	got, err := store.GetOptionsByEvent(ctx, "evt-1")
	if err != nil {
		t.Fatalf("GetOptionsByEvent() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 options, got %d", len(got))
	}
}

func TestUpsertRSVP(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)
	ctx := t.Context()
	userID, groupID := seedTestData(t, db)

	e := &event.Event{
		ID:            "evt-1",
		GroupID:       groupID,
		CreatorID:     userID,
		Title:         "Event",
		Description:   "Desc",
		ScheduledTime: time.Now().Add(time.Hour),
	}
	if err := store.CreateEvent(ctx, e); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	opts := []event.Option{
		{ID: "opt-1", EventID: "evt-1", Label: "going"},
		{ID: "opt-2", EventID: "evt-1", Label: "not going"},
	}
	if err := store.CreateOptions(ctx, opts); err != nil {
		t.Fatalf("CreateOptions() error = %v", err)
	}

	rsvp := &event.RSVP{EventID: "evt-1", UserID: userID, OptionID: "opt-1"}
	if err := store.UpsertRSVP(ctx, rsvp); err != nil {
		t.Fatalf("UpsertRSVP() error = %v", err)
	}

	got, err := store.GetUserRSVP(ctx, "evt-1", userID)
	if err != nil {
		t.Fatalf("GetUserRSVP() error = %v", err)
	}
	if got.OptionID != "opt-1" {
		t.Errorf("OptionID = %q, want %q", got.OptionID, "opt-1")
	}

	rsvp.OptionID = "opt-2"
	if err = store.UpsertRSVP(ctx, rsvp); err != nil {
		t.Fatalf("UpsertRSVP() update error = %v", err)
	}
	got, err = store.GetUserRSVP(ctx, "evt-1", userID)
	if err != nil {
		t.Fatalf("GetUserRSVP() after update error = %v", err)
	}
	if got.OptionID != "opt-2" {
		t.Errorf("OptionID after update = %q, want %q", got.OptionID, "opt-2")
	}
}

func TestGetRSVPsByEvent(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)
	ctx := t.Context()
	userID, groupID := seedTestData(t, db)

	e := &event.Event{
		ID:            "evt-1",
		GroupID:       groupID,
		CreatorID:     userID,
		Title:         "Event",
		Description:   "Desc",
		ScheduledTime: time.Now().Add(time.Hour),
	}
	if err := store.CreateEvent(ctx, e); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	opts := []event.Option{
		{ID: "opt-1", EventID: "evt-1", Label: "going"},
		{ID: "opt-2", EventID: "evt-1", Label: "not going"},
	}
	if err := store.CreateOptions(ctx, opts); err != nil {
		t.Fatalf("CreateOptions() error = %v", err)
	}

	// Create second user
	_, _ = db.ExecContext(ctx, `INSERT INTO users (id, username, email, password_hash) VALUES (?, ?, ?, ?)`,
		"user-2", "user2", "u2@example.com", "hash")

	rsvps := []event.RSVP{
		{EventID: "evt-1", UserID: userID, OptionID: "opt-1"},
		{EventID: "evt-1", UserID: "user-2", OptionID: "opt-2"},
	}
	for _, r := range rsvps {
		if err := store.UpsertRSVP(ctx, &r); err != nil {
			t.Fatalf("UpsertRSVP() error = %v", err)
		}
	}

	got, err := store.GetRSVPsByEvent(ctx, "evt-1")
	if err != nil {
		t.Fatalf("GetRSVPsByEvent() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 RSVPs, got %d", len(got))
	}
}

func TestDeleteEvent(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)
	ctx := t.Context()
	userID, groupID := seedTestData(t, db)

	e := &event.Event{
		ID:            "evt-1",
		GroupID:       groupID,
		CreatorID:     userID,
		Title:         "Event",
		Description:   "Desc",
		ScheduledTime: time.Now().Add(time.Hour),
	}
	if err := store.CreateEvent(ctx, e); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	if err := store.DeleteEvent(ctx, "evt-1"); err != nil {
		t.Fatalf("DeleteEvent() error = %v", err)
	}

	_, err := store.GetEvent(ctx, "evt-1")
	if !errors.Is(err, event.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound after delete, got %v", err)
	}
}
