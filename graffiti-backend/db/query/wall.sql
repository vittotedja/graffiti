-- name: CreateWall :one
INSERT INTO walls(
 user_id,
 description,
 background_image
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetWall :one
SELECT * FROM walls
WHERE id = $1 LIMIT 1;

-- name: ListWalls :many
SELECT * FROM walls
ORDER BY id;