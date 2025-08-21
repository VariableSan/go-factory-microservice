-- Drop indexes
DROP INDEX IF EXISTS idx_feeds_created_at;
DROP INDEX IF EXISTS idx_feeds_published;
DROP INDEX IF EXISTS idx_feeds_user_id;

-- Drop feeds table
DROP TABLE IF EXISTS feeds;
