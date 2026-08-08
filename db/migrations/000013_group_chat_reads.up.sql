-- 000013_group_chat_reads.up.sql
--
-- Per-user read marker for group chat. Group chat messages are identified by
-- UUID (not monotonic), so unread is tracked with a timestamp: messages are
-- unread for a user when sender_id != user AND created_at > last_read_at.
CREATE TABLE IF NOT EXISTS group_chat_reads (
    group_id TEXT NOT NULL,
    user_id  TEXT NOT NULL,
    last_read_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id)  REFERENCES users(id)  ON DELETE CASCADE
);
