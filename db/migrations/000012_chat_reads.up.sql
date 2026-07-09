CREATE TABLE chat_reads (
    chat_id TEXT NOT NULL REFERENCES direct_chats(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_message_id INTEGER REFERENCES chat_messages(id) ON DELETE SET NULL,
    last_read_at DATETIME,
    unread_count INTEGER DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX idx_chat_reads_user_id ON chat_reads(user_id);
CREATE INDEX idx_chat_reads_unread ON chat_reads(user_id, unread_count) WHERE unread_count > 0;
CREATE INDEX idx_chat_reads_chat_user ON chat_reads(chat_id, user_id);
