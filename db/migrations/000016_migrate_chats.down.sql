-- 000016 down: Revert new chats/messages schema.
-- Drops new tables and restores legacy chat_reads with FK to direct_chats.

DROP TRIGGER IF EXISTS sync_chats_insert;
DROP TRIGGER IF EXISTS sync_chats_update;
DROP TRIGGER IF EXISTS sync_chats_delete;

-- Restore old chat_reads with FK pointing to direct_chats
CREATE TEMPORARY TABLE _chat_reads_backup AS SELECT * FROM chat_reads;
DROP TABLE chat_reads;

CREATE TABLE chat_reads (
    chat_id TEXT NOT NULL REFERENCES direct_chats(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_message_id INTEGER REFERENCES chat_messages(id) ON DELETE SET NULL,
    last_read_at DATETIME,
    unread_count INTEGER DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX idx_chat_reads_user ON chat_reads(user_id);
CREATE INDEX idx_chat_reads_unread ON chat_reads(user_id, unread_count) WHERE unread_count > 0;

INSERT INTO chat_reads SELECT * FROM _chat_reads_backup;
DROP TABLE _chat_reads_backup;

-- Drop new tables
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chats;
