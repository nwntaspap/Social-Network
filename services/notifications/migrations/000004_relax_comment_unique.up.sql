DROP INDEX IF EXISTS idx_notifications_active;
CREATE UNIQUE INDEX idx_notifications_active
    ON notifications(recipient_id, type, resource_type, resource_id, actor_id)
    WHERE deleted_at IS NULL AND type <> 'comment';
