-- Migration: Add tags system (author tags and user tags)

-- Drop old labels tables
DROP TABLE IF EXISTS recipe_labels;
DROP TABLE IF EXISTS labels;

-- Tags table (global author tags, case-insensitive, normalized to lowercase)
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- Recipe-Tags junction table (author tags on recipes)
CREATE TABLE IF NOT EXISTS recipe_tags (
    recipe_id INT REFERENCES recipes(id) ON DELETE CASCADE,
    tag_id INT REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (recipe_id, tag_id)
);

-- User tags table (personal tags per user per recipe)
CREATE TABLE IF NOT EXISTS user_tags (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    recipe_id INT REFERENCES recipes(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    UNIQUE(user_id, recipe_id, name)
);

-- Index for faster tag lookups
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_user_tags_user_recipe ON user_tags(user_id, recipe_id);
CREATE INDEX IF NOT EXISTS idx_user_tags_name ON user_tags(name);
