-- name: CreateLike :one
INSERT INTO likes(
 post_id,
 user_id
) VALUES (
  $1, $2
) RETURNING *;

-- name: GetLike :one
SELECT * FROM likes
WHERE post_id = $1 AND user_id = $2 LIMIT 1;

-- name: ListLikes :many
SELECT * FROM likes
ORDER BY liked_at DESC;

-- name: ListLikesByPost :many
SELECT * FROM likes
WHERE post_id = $1
ORDER BY liked_at DESC;

-- name: ListLikesByUser :many
SELECT * FROM likes
WHERE user_id = $1
ORDER BY liked_at DESC;

-- name: getNumberOfLikesByPost :one
SELECT COUNT(*) FROM likes
WHERE post_id = $1;

-- name: DeleteLike :exec
DELETE FROM likes
WHERE post_id = $1 AND user_id = $2;


