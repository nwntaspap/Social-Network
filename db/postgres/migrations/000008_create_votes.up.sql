CREATE TABLE IF NOT EXISTS votes (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id INTEGER REFERENCES topics(id) ON DELETE CASCADE,
    comment_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
    reaction_type INTEGER NOT NULL CHECK(reaction_type IN (-1, 1)),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, topic_id),
    UNIQUE (user_id, comment_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_topic_votes ON votes(user_id, topic_id) WHERE comment_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_comment_votes ON votes(user_id, comment_id) WHERE comment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_votes_topic_reaction ON votes(topic_id, reaction_type) WHERE comment_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_votes_comment_reaction ON votes(comment_id, reaction_type) WHERE comment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_votes_user ON votes(user_id);
