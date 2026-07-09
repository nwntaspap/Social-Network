CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    image_path TEXT DEFAULT 'static/images/categories/default_category.png',
    color TEXT DEFAULT '#CCCCCC',
    slug TEXT DEFAULT 'default-slug',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_categories_created_by ON categories(created_by);
