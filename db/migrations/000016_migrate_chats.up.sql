-- 000016: Create new chats/messages schema and migrate data from legacy tables.
-- Legacy tables (direct_chats, chat_messages) are kept alongside for strangler fig coexistence.
-- chat_reads is recreated with FK pointing to new chats table.
-- Sync triggers keep direct_chats in sync so old code's FK stays valid.

-- 1. New chats table
CREATE TABLE chats (
    id TEXT PRIMARY KEY,
    user_one_id TEXT NOT NULL,
    user_two_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_message_id INTEGER,
    last_message_at DATETIME,
    UNIQUE(user_one_id, user_two_id),
    FOREIGN KEY(user_one_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(user_two_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_chats_user_one ON chats(user_one_id);
CREATE INDEX idx_chats_user_two ON chats(user_two_id);
CREATE INDEX idx_chats_last_message ON chats(last_message_at DESC);

-- 2. New messages table
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id TEXT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    client_message_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(sender_id, client_message_id)
);

CREATE INDEX idx_messages_chat ON messages(chat_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender_id, created_at DESC);

-- 3. Migrate data from legacy tables
INSERT INTO chats (id, user_one_id, user_two_id, created_at, updated_at, last_message_id, last_message_at)
SELECT id, user_low_id, user_high_id, created_at, updated_at, last_message_id, last_message_at
FROM direct_chats;

INSERT INTO messages (id, chat_id, sender_id, content, client_message_id, created_at)
SELECT id, chat_id, sender_id, content, client_message_id, created_at
FROM chat_messages;

-- 4. Recreate chat_reads with FK pointing to new chats table
CREATE TEMPORARY TABLE _chat_reads_backup AS SELECT * FROM chat_reads;
DROP TABLE chat_reads;

CREATE TABLE chat_reads (
    chat_id TEXT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_message_id INTEGER REFERENCES messages(id) ON DELETE SET NULL,
    last_read_at DATETIME,
    unread_count INTEGER DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX idx_chat_reads_user ON chat_reads(user_id);
CREATE INDEX idx_chat_reads_unread ON chat_reads(user_id, unread_count) WHERE unread_count > 0;

INSERT INTO chat_reads SELECT * FROM _chat_reads_backup;
DROP TABLE _chat_reads_backup;

-- 5. Sync triggers: new code writes to chats, triggers keep direct_chats in sync
--    so old code's FK (chat_reads -> direct_chats) stays valid during strangler fig.
CREATE TRIGGER sync_chats_insert AFTER INSERT ON chats BEGIN
    INSERT OR IGNORE INTO direct_chats (id, user_low_id, user_high_id, created_at, updated_at, last_message_id, last_message_at)
    VALUES (NEW.id, NEW.user_one_id, NEW.user_two_id, NEW.created_at, NEW.updated_at, NEW.last_message_id, NEW.last_message_at);
END;

CREATE TRIGGER sync_chats_update AFTER UPDATE ON chats BEGIN
    UPDATE direct_chats SET
        updated_at = NEW.updated_at,
        last_message_id = NEW.last_message_id,
        last_message_at = NEW.last_message_at
    WHERE id = NEW.id;
END;

CREATE TRIGGER sync_chats_delete AFTER DELETE ON chats BEGIN
    DELETE FROM direct_chats WHERE id = OLD.id;
END;
