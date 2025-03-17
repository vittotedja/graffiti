-- name: CreatePost :one
INSERT INTO posts(
 wall_id,
 author,
 media_url,
 post_type
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetPost :one
SELECT * FROM posts
WHERE id = $1 LIMIT 1;

-- name: ListPosts :many
SELECT * FROM posts
ORDER BY id;