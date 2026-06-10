/* Users table */
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE CHECK(email LIKE '%_@__%.__%'),
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT,
    age INTEGER,
    gender TEXT,
    first_name TEXT,
    last_name TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    avatar_url TEXT
);

/* OAuth */
CREATE TABLE IF NOT EXISTS oauth_providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    provider_user_id TEXT NOT NULL,
    email TEXT,
    username TEXT,
    avatar_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(provider, provider_user_id),
    UNIQUE(user_id, provider)
);

/* Sessions */
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    refresh_token TEXT,
    refresh_token_expires_at DATETIME NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

/* Categories */
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    image_path TEXT DEFAULT 'static/images/categories/default_category.png',
    color TEXT DEFAULT '#CCCCCC',
    slug TEXT DEFAULT 'default-slug',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL REFERENCES users(id)
);

/* Topics */
CREATE TABLE IF NOT EXISTS topics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    image_path TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

/* Topic/Category junction */
CREATE TABLE IF NOT EXISTS topic_categories (
    topic_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (topic_id, category_id),
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

/* Comments */
CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

/* Votes */
CREATE TABLE IF NOT EXISTS votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id INTEGER REFERENCES topics(id) ON DELETE CASCADE,
    comment_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
    reaction_type INTEGER NOT NULL CHECK(reaction_type IN (-1, 1)),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, topic_id),
    UNIQUE (user_id, comment_id)
);

/* Notifications */
CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    related_type TEXT,
    related_id INTEGER,
    is_read BOOLEAN DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

/* Direct_chats */
CREATE TABLE IF NOT EXISTS direct_chats (
    id TEXT PRIMARY KEY,
    user_low_id TEXT NOT NULL,
    user_high_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_message_id INTEGER,
    last_message_at DATETIME,
    UNIQUE(user_low_id, user_high_id)
);

/* Chat messages */
CREATE TABLE IF NOT EXISTS chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id TEXT NOT NULL REFERENCES direct_chats(id) ON DELETE CASCADE,
    sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    client_message_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(sender_id, client_message_id)
);

/* Chat reads */
CREATE TABLE IF NOT EXISTS chat_reads (
    chat_id TEXT NOT NULL REFERENCES direct_chats(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_message_id INTEGER REFERENCES chat_messages(id) ON DELETE SET NULL,
    last_read_at DATETIME,
    unread_count INTEGER DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, user_id)
);

/* Topic/category junction table indexes */
CREATE INDEX IF NOT EXISTS idx_topic_categories_topic_id ON topic_categories(topic_id);
CREATE INDEX IF NOT EXISTS idx_topic_categories_category_id ON topic_categories(category_id);

/* Users table indexes */
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

/* Sessions table indexes */
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

/* Categories table indexes */
CREATE INDEX IF NOT EXISTS idx_categories_created_by ON categories(created_by);

/* Topics table indexes */
CREATE INDEX IF NOT EXISTS idx_topics_user ON topics(user_id);
CREATE INDEX IF NOT EXISTS idx_topics_created ON topics(created_at DESC);

/* Comments table indexes */
CREATE INDEX IF NOT EXISTS idx_comments_topic ON comments(topic_id);
CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id);

/* Votes table indexes – prevent duplicate votes */
CREATE UNIQUE INDEX IF NOT EXISTS idx_topic_votes ON votes(user_id, topic_id) WHERE comment_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_comment_votes ON votes(user_id, comment_id) WHERE comment_id IS NOT NULL;

/* Fast vote counting and lookups */
CREATE INDEX IF NOT EXISTS idx_votes_topic_reaction ON votes(topic_id, reaction_type) WHERE comment_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_votes_comment_reaction ON votes(comment_id, reaction_type) WHERE comment_id IS NOT NULL;

/* User activity lookup */
CREATE INDEX IF NOT EXISTS idx_votes_user ON votes(user_id);

/* Notifications table indexes */
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);

/* Chat table indexes – get all messages in a chat, ordered for pagination */
CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_id ON chat_messages(chat_id, created_at DESC);

-- Find all chats for a user (since they can be user_low or user_high)
-- Helps build chat list for current user
CREATE INDEX IF NOT EXISTS idx_direct_chats_user_low_id ON direct_chats(user_low_id);
CREATE INDEX IF NOT EXISTS idx_direct_chats_user_high_id ON direct_chats(user_high_id);

/* Order chat list by most recent message (Discord-like behavior) */
CREATE INDEX IF NOT EXISTS idx_direct_chats_last_message_at ON direct_chats(last_message_at DESC);

/* Fast sender lookup for activity tracking */
CREATE INDEX IF NOT EXISTS idx_chat_messages_sender_id ON chat_messages(sender_id, created_at DESC);

/* Chat reads table indexes – fast lookup of read state for a user in a specific chat */
CREATE INDEX IF NOT EXISTS idx_chat_reads_user_id ON chat_reads(user_id);

/* Find all unread chats for a user (for navbar badge and chat list) */
CREATE INDEX IF NOT EXISTS idx_chat_reads_unread ON chat_reads(user_id, unread_count) WHERE unread_count > 0;

/* Fast check of what user has read in specific chat */
CREATE INDEX IF NOT EXISTS idx_chat_reads_chat_user ON chat_reads(chat_id, user_id);

CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
