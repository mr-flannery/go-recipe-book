-- Migration: Add indexes for common query patterns

-- recipes.author_id: filtering recipes by author
CREATE INDEX IF NOT EXISTS idx_recipes_author_id ON recipes(author_id);

-- recipes.created_at: pagination and sorting by date
CREATE INDEX IF NOT EXISTS idx_recipes_created_at ON recipes(created_at DESC);

-- comments.recipe_id: fetching comments for a recipe
CREATE INDEX IF NOT EXISTS idx_comments_recipe_id ON comments(recipe_id);

-- comments.author_id: fetching a user's comments
CREATE INDEX IF NOT EXISTS idx_comments_author_id ON comments(author_id);

-- recipe_tags.tag_id: JOINs on tag_id for tag filtering
CREATE INDEX IF NOT EXISTS idx_recipe_tags_tag_id ON recipe_tags(tag_id);
