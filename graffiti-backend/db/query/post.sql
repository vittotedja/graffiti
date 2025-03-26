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

-- name: ListPostsByWall :many
SELECT * FROM posts
WHERE wall_id = $1
ORDER BY id DESC;

-- name: GetHighlightedPosts :many
SELECT * FROM posts
WHERE is_highlighted = true
ORDER BY id;

-- name: GetHighlightedPostsByWall :many
SELECT * FROM posts
WHERE wall_id = $1 AND is_highlighted = true
ORDER BY id;

-- name: UpdatePost :one
UPDATE posts
  set
    media_url = COALESCE($2, media_url),
    post_type = COALESCE($3, post_type)
WHERE id = $1
RETURNING *;

-- name: HighlightPost :one
UPDATE posts
  set is_highlighted = true
WHERE id = $1
RETURNING *;

-- name: UnhighlightPost :one
UPDATE posts
  set is_highlighted = false
WHERE id = $1
RETURNING *;

-- name: AddLikesCount :one
UPDATE posts
  set likes_count = likes_count + 1
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
UPDATE posts
  set is_deleted = true
WHERE id = $1;
