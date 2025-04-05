-- Enable trigram indexing
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Index for faster search on comment content
CREATE INDEX IF NOT EXISTS idx_comments_content 
ON comments USING GIN (content gin_trgm_ops);

-- Index for searching posts by title
CREATE INDEX IF NOT EXISTS idx_posts_title 
ON posts USING GIN (title gin_trgm_ops);

-- Index for searching posts by tags
CREATE INDEX IF NOT EXISTS idx_posts_tags 
ON posts USING GIN (tags);

-- Index for faster username lookups
CREATE INDEX IF NOT EXISTS idx_users_username 
ON users (username);

-- Foreign key-based indexes for joins
CREATE INDEX IF NOT EXISTS idx_posts_user_id 
ON posts (user_id);

CREATE INDEX IF NOT EXISTS idx_comments_post_id 
ON comments (post_id);

-- Enforce unique email constraint
ALTER TABLE users ADD CONSTRAINT IF NOT EXISTS unique_email UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT IF NOT EXISTS unique_username UNIQUE (username);
