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

-- name: ListReceivedPendingFriendRequests :many
SELECT f.*, users.fullname, users.username, users.profile_picture FROM friendships f
left join users on users.id = f.from_user
WHERE to_user = $1 AND status = 'pending';

-- name: ListSentPendingFriendRequests :many
SELECT f.*, users.fullname, users.username, users.profile_picture FROM friendships f
left join users on users.id = f.to_user
WHERE from_user = $1 AND status = 'pending';

-- name: ListFriendsDetailsByStatus :many
SELECT u.id as user_id, u.fullname, u.username, u.profile_picture, f.status, f.id
FROM friendships f
JOIN users u ON u.id = 
  CASE 
    WHEN f.from_user = $1 AND $2 = 'friends' THEN f.to_user
    WHEN f.from_user = $1 AND $2 = 'sent' THEN f.to_user
    WHEN f.to_user = $1 AND $2 = 'requested' THEN f.from_user
    WHEN f.to_user = $1 AND $2 = 'friends' THEN f.from_user
    ELSE NULL
  END
WHERE 
  (
    ($2 = 'friends' AND f.status = 'friends' AND (f.from_user = $1)) OR
    ($2 = 'sent' AND f.status = 'pending' AND f.from_user = $1) OR
    ($2 = 'requested' AND f.status = 'pending' AND f.to_user = $1)
  );


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



