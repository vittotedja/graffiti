-- Drop indexes
DROP INDEX IF EXISTS idx_users_fullname_trgm;

DROP INDEX IF EXISTS idx_users_username_trgm;

DROP INDEX IF EXISTS friendships_from_user_to_user_idx;

DROP INDEX IF EXISTS friendships_to_user_idx;

DROP INDEX IF EXISTS friendships_from_user_idx;

DROP INDEX IF EXISTS likes_post_id_user_id_idx;

DROP INDEX IF EXISTS likes_user_id_idx;

DROP INDEX IF EXISTS likes_post_id_idx;

DROP INDEX IF EXISTS walls_user_id_idx;

DROP INDEX IF EXISTS users_username_idx;

-- Drop tables
DROP TABLE IF EXISTS friendships;

DROP TABLE IF EXISTS likes;

DROP TABLE IF EXISTS posts;

DROP TABLE IF EXISTS walls;

DROP TABLE IF EXISTS users;

-- Drop types
DROP TYPE IF EXISTS post_type;

DROP TYPE IF EXISTS status;

-- Optionally drop extension (if you want)
-- DROP EXTENSION IF EXISTS pg_trgm;