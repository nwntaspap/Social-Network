DROP INDEX IF EXISTS idx_notifications_group;

ALTER TABLE notifications DROP COLUMN group_id;
