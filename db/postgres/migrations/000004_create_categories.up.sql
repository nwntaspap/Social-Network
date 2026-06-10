CREATE TABLE IF NOT EXISTS categories (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    image_path TEXT DEFAULT 'static/images/categories/default_category.png',
    color TEXT DEFAULT '#CCCCCC',
    slug TEXT DEFAULT 'default-slug',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_categories_created_by ON categories(created_by);
