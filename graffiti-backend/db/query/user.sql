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
  set username = $2,
  fullname = $3,
  email = $4,
  hashed_password = $5
WHERE id = $1
RETURNING *;

-- name: UpdateProfile :one
UPDATE users
  set profile_picture = $2,
  bio = $3,
  background_image = $4
WHERE id = $1
RETURNING *;

-- name: FinishOnboarding :exec
UPDATE users
  set has_onboarded = true,
  onboarding_at = now()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
