-- name: CreateUser :one
INSERT INTO users(
 username,
 fullname,
 email,
 hashed_password 
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY id;


-- name: UpdateUser :one
UPDATE users
SET 
    username = COALESCE($2, username),
    fullname = COALESCE($3, fullname),
    email = COALESCE($4, email),
    hashed_password = COALESCE($5, hashed_password)
WHERE id = $1
RETURNING *;

-- name: UpdateProfile :one
UPDATE users
SET 
    profile_picture = COALESCE($2, profile_picture),
    bio = COALESCE($3, bio),
    background_image = COALESCE($4, background_image)
WHERE id = $1
RETURNING *;

-- name: FinishOnboarding :exec
UPDATE users
SET 
    has_onboarded = true,
    onboarding_at = now()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: SearchUsersTrigram :many
SELECT
    id,
    username,
    fullname,
    profile_picture
FROM users
WHERE
    username % sqlc.arg(search_term)
    OR
    fullname % sqlc.arg(search_term)
ORDER BY GREATEST(
    similarity(username, sqlc.arg(search_term)),
    similarity(fullname, sqlc.arg(search_term))
) DESC
LIMIT 10;


-- name: SearchUsersILike :many
SELECT
    id,
    username,
    fullname,
    profile_picture
FROM users
WHERE
    username ILIKE '%' || sqlc.arg(search_term) || '%'
    OR
    fullname ILIKE '%' || sqlc.arg(search_term) || '%'
ORDER BY username ASC
LIMIT 10;




