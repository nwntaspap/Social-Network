CREATE TABLE IF NOT EXISTS chat_reads (
    chat_id TEXT NOT NULL REFERENCES direct_chats(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_message_id INTEGER REFERENCES chat_messages(id) ON DELETE SET NULL,
    last_read_at TIMESTAMPTZ,
    unread_count INTEGER DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_reads_user_id ON chat_reads(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_reads_unread ON chat_reads(user_id, unread_count) WHERE unread_count > 0;
