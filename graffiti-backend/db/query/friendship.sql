-- name: CreateFriendship :one
INSERT INTO friendships(
 user_id,
 friend_id,
 status
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetFriendship :one
SELECT * FROM friendships
WHERE id = $1 LIMIT 1;

-- name: ListFriendships :many
SELECT * FROM friendships
ORDER BY id;

-- name: GetNumberOfFriends :one
SELECT COUNT(*) FROM friendships
WHERE user_id = $1 AND status = 'accepted';

-- name: GetNumberOfPendingFriendRequests :one
SELECT COUNT(*) FROM friendships
WHERE friend_id = $1 AND status = 'pending';

-- name: UpdateFriendship :one
UPDATE friendships
  set status = $2
WHERE id = $1
RETURNING *;

-- name: DeleteFriendship :exec
DELETE FROM friendships
WHERE id = $1;



