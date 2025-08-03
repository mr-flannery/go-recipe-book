-- Migration: Initial schema setup

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Recipes table
CREATE TABLE IF NOT EXISTS recipes (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    ingredients_md TEXT NOT NULL,
    instructions_md TEXT NOT NULL,
    prep_time INT,
    cook_time INT,
    calories INT,
    author_id INT REFERENCES users(id),
    image BYTEA,
    parent_id INT REFERENCES recipes(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Labels table
CREATE TABLE IF NOT EXISTS labels (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- RecipeLabels table
CREATE TABLE IF NOT EXISTS recipe_labels (
    recipe_id INT REFERENCES recipes(id),
    label_id INT REFERENCES labels(id),
    PRIMARY KEY (recipe_id, label_id)
);

-- Comments table
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    recipe_id INT REFERENCES recipes(id),
    author_id INT REFERENCES users(id),
    content_md TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ProposedChanges table
CREATE TABLE IF NOT EXISTS proposed_changes (
    id SERIAL PRIMARY KEY,
    recipe_id INT REFERENCES recipes(id),
    proposer_id INT REFERENCES users(id),
    title TEXT,
    ingredients_md TEXT,
    instructions_md TEXT,
    prep_time INT,
    cook_time INT,
    calories INT,
    image BYTEA,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status TEXT CHECK (status IN ('pending', 'accepted', 'rejected'))
);
