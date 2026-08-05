DROP INDEX IF EXISTS idx_notifications_join_request;

ALTER TABLE notifications DROP COLUMN join_request_id;