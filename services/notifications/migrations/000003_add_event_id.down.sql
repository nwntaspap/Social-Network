DROP INDEX IF EXISTS idx_notifications_event;

ALTER TABLE notifications DROP COLUMN event_id;
