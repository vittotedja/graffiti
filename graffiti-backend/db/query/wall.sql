-- name: CreateWall :one
INSERT INTO walls(
    user_id,
    description,
    background_image
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: CreateTestWall :one
INSERT INTO walls(
    user_id,
    title,
    description,
    is_public,
    background_image
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetWall :one
SELECT * FROM walls
WHERE id = $1 LIMIT 1;

-- name: ListWalls :many
SELECT * FROM walls
ORDER BY id DESC;

-- name: ListWallsByUser :many
SELECT * FROM walls
WHERE user_id = $1
AND is_deleted = false
AND is_archived = false
ORDER BY created_at DESC;

-- name: UpdateWall :one
UPDATE walls
SET 
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    background_image = COALESCE($4, background_image),
    is_public = COALESCE($5, is_public)
WHERE id = $1
RETURNING *;

-- name: DeleteWall :exec
UPDATE walls
    set is_deleted = true
WHERE id = $1;

-- name: ArchiveWall :exec
UPDATE walls
    set is_archived = true
WHERE id = $1
RETURNING *;

-- name: UnarchiveWall :exec
UPDATE walls
    set is_archived = false
WHERE id = $1
RETURNING *;

-- name: PublicizeWall :one
UPDATE walls
    set is_public = true
WHERE id = $1
RETURNING *;

-- name: PrivatizeWall :one
UPDATE walls
    set is_public = false
WHERE id = $1
RETURNING *;

-- name: PinUnpinWall :one
UPDATE walls
    set is_pinned = not is_pinned
WHERE id = $1
RETURNING *;
