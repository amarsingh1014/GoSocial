-- Remove unique constraint on email
ALTER TABLE users DROP CONSTRAINT IF EXISTS unique_email;

-- Drop indexes
DROP INDEX IF EXISTS idx_comments_content;
DROP INDEX IF EXISTS idx_posts_title;
DROP INDEX IF EXISTS idx_posts_tags;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_posts_user_id;
DROP INDEX IF EXISTS idx_comments_post_id;

-- Drop pg_trgm extension (optional, if no other indexes depend on it)
DROP EXTENSION IF EXISTS pg_trgm;
