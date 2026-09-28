ALTER TABLE notifications ADD COLUMN group_id TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_notifications_group ON notifications(group_id) WHERE group_id <> '';
