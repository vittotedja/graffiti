-- name: CreateUser :one
INSERT INTO users(
 username,
 email,
 hashed_password 
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY id
LIMIT $1 
OFFSET $2;