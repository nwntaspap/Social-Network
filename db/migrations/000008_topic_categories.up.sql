CREATE TABLE topic_categories (
    topic_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (topic_id, category_id),
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE INDEX idx_topic_categories_topic_id ON topic_categories(topic_id);
CREATE INDEX idx_topic_categories_category_id ON topic_categories(category_id);
