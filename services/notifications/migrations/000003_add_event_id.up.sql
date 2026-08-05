ALTER TABLE notifications ADD COLUMN event_id TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_notifications_event ON notifications(event_id) WHERE event_id <> '';
