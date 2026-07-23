ALTER TABLE topics ADD COLUMN group_id TEXT REFERENCES groups(id) ON DELETE SET NULL;

CREATE INDEX idx_topics_group ON topics(group_id);
