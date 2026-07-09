CREATE TABLE chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id TEXT NOT NULL REFERENCES direct_chats(id) ON DELETE CASCADE,
    sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    client_message_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(sender_id, client_message_id)
);

CREATE INDEX idx_chat_messages_chat_id ON chat_messages(chat_id, created_at DESC);
CREATE INDEX idx_chat_messages_sender_id ON chat_messages(sender_id, created_at DESC);
