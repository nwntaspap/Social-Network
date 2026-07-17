ALTER TABLE topics ADD COLUMN visibility INTEGER NOT NULL DEFAULT 0;

CREATE TABLE topic_allowed_users (
    topic_id INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (topic_id, user_id)
);

CREATE INDEX idx_topic_allowed_users_topic ON topic_allowed_users(topic_id);
