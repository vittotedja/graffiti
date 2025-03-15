-- name: CreatePost :one
INSERT INTO posts(
 wall_id,
 author,
 media_url,
 post_type,
 is_highlighted (default false),
 likes_count (default 0),
 is_deleted (default false),
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetPost :one
SELECT * FROM posts
WHERE id = $1 LIMIT 1;

-- name: ListPosts :many
SELECT * FROM posts
ORDER BY id
LIMIT $1 
OFFSET $2;