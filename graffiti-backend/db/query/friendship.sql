-- name: CreateFriendship :one
INSERT INTO friendships(
 from_user,
 to_user,
 status
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetFriendship :one
SELECT * FROM friendships
WHERE id = $1 LIMIT 1;

-- name: ListFriendshipByUserPairs :one
SELECT * FROM friendships
WHERE (from_user = $1 AND to_user = $2) OR (from_user = $2 AND to_user = $1);

-- name: ListFriendships :many
SELECT * FROM friendships
ORDER BY id;

-- name: ListFriendshipsByUserId :many
SELECT * FROM friendships
WHERE (from_user = $1 OR to_user = $1)
ORDER BY id;

-- name: ListFriendshipsByUserIdAndStatus :many
SELECT * FROM friendships
WHERE (from_user = $1 OR to_user = $1) AND status = $2
ORDER BY id;

-- name: AcceptFriendship :one
UPDATE friendships
  SET status = 'friends'
WHERE id = $1
RETURNING *;

-- name: RejectFriendship :exec
DELETE FROM friendships
WHERE id = $1;

-- name: BlockFriendship :one
UPDATE friendships
  SET status = 'blocked'
WHERE id = $1
RETURNING *;

-- name: UpdateFriendship :one
UPDATE friendships
  SET status = $2
WHERE id = $1
RETURNING *;

-- name: GetNumberOfFriends :one
SELECT COUNT(*) FROM friendships
WHERE ((from_user = $1) OR (to_user = $1)) AND status = 'friends';

-- name: GetNumberOfPendingFriendRequests :one
SELECT COUNT(*) FROM friendships
WHERE to_user = $1 AND status = 'pending';

-- name: DeleteFriendship :exec
DELETE FROM friendships
WHERE id = $1;



