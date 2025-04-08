-- Materialized view for accepted (mutual) friendships
CREATE MATERIALIZED VIEW accepted_friendships_mv AS
SELECT DISTINCT
    from_user AS user_id,
    to_user AS friend_id
FROM friendships
WHERE status = 'friends'
UNION
SELECT DISTINCT
    to_user AS user_id,
    from_user AS friend_id
FROM friendships
WHERE status = 'friends';

-- Indexes for materialized view
CREATE INDEX idx_accepted_friendships_user_id ON accepted_friendships_mv (user_id);
CREATE INDEX idx_accepted_friendships_friend_id ON accepted_friendships_mv (friend_id);
CREATE UNIQUE INDEX idx_accepted_friendships_mv_unique ON accepted_friendships_mv (user_id, friend_id);

