CREATE TABLE IF NOT EXISTS direct_chats (
    id TEXT PRIMARY KEY,
    user_low_id TEXT NOT NULL,
    user_high_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_message_id INTEGER,
    last_message_at TIMESTAMPTZ,
    UNIQUE(user_low_id, user_high_id)
);

CREATE INDEX IF NOT EXISTS idx_direct_chats_user_low_id ON direct_chats(user_low_id);
CREATE INDEX IF NOT EXISTS idx_direct_chats_user_high_id ON direct_chats(user_high_id);
CREATE INDEX IF NOT EXISTS idx_direct_chats_last_message_at ON direct_chats(last_message_at DESC);
