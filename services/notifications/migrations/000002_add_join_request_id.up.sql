ALTER TABLE notifications ADD COLUMN join_request_id TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_notifications_join_request ON notifications(join_request_id) WHERE join_request_id <> '';