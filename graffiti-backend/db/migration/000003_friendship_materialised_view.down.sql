-- Drop indexes before dropping the materialized view
DROP INDEX IF EXISTS idx_accepted_friendships_mv_unique;
DROP INDEX IF EXISTS idx_accepted_friendships_user_id;
DROP INDEX IF EXISTS idx_accepted_friendships_friend_id;

-- Drop the materialized view
DROP MATERIALIZED VIEW IF EXISTS accepted_friendships_mv;