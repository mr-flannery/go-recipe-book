-- Rollback: Remove performance indexes

DROP INDEX IF EXISTS idx_recipes_author_id;
DROP INDEX IF EXISTS idx_recipes_created_at;
DROP INDEX IF EXISTS idx_comments_recipe_id;
DROP INDEX IF EXISTS idx_comments_author_id;
DROP INDEX IF EXISTS idx_recipe_tags_tag_id;
