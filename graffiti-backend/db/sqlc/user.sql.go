// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: user.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users(
 username,
 fullname,
 email,
 hashed_password 
) VALUES (
  $1, $2, $3, $4
) RETURNING id, username, fullname, email, hashed_password, profile_picture, bio, has_onboarded, background_image, onboarding_at, created_at, updated_at
`

type CreateUserParams struct {
	Username       string
	Fullname       pgtype.Text
	Email          string
	HashedPassword string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Username,
		arg.Fullname,
		arg.Email,
		arg.HashedPassword,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Fullname,
		&i.Email,
		&i.HashedPassword,
		&i.ProfilePicture,
		&i.Bio,
		&i.HasOnboarded,
		&i.BackgroundImage,
		&i.OnboardingAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1
`

func (q *Queries) DeleteUser(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteUser, id)
	return err
}

const finishOnboarding = `-- name: FinishOnboarding :exec
UPDATE users
SET 
    has_onboarded = true,
    onboarding_at = now()
WHERE id = $1
`

func (q *Queries) FinishOnboarding(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, finishOnboarding, id)
	return err
}

const getUser = `-- name: GetUser :one
SELECT id, username, fullname, email, hashed_password, profile_picture, bio, has_onboarded, background_image, onboarding_at, created_at, updated_at FROM users
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, id pgtype.UUID) (User, error) {
	row := q.db.QueryRow(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Fullname,
		&i.Email,
		&i.HashedPassword,
		&i.ProfilePicture,
		&i.Bio,
		&i.HasOnboarded,
		&i.BackgroundImage,
		&i.OnboardingAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT id, username, fullname, email, hashed_password, profile_picture, bio, has_onboarded, background_image, onboarding_at, created_at, updated_at FROM users
ORDER BY id
`

func (q *Queries) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.Query(ctx, listUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Fullname,
			&i.Email,
			&i.HashedPassword,
			&i.ProfilePicture,
			&i.Bio,
			&i.HasOnboarded,
			&i.BackgroundImage,
			&i.OnboardingAt,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateProfile = `-- name: UpdateProfile :one
UPDATE users
SET 
    profile_picture = COALESCE($2, profile_picture),
    bio = COALESCE($3, bio),
    background_image = COALESCE($4, background_image)
WHERE id = $1
RETURNING id, username, fullname, email, hashed_password, profile_picture, bio, has_onboarded, background_image, onboarding_at, created_at, updated_at
`

type UpdateProfileParams struct {
	ID              pgtype.UUID
	ProfilePicture  pgtype.Text
	Bio             pgtype.Text
	BackgroundImage pgtype.Text
}

func (q *Queries) UpdateProfile(ctx context.Context, arg UpdateProfileParams) (User, error) {
	row := q.db.QueryRow(ctx, updateProfile,
		arg.ID,
		arg.ProfilePicture,
		arg.Bio,
		arg.BackgroundImage,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Fullname,
		&i.Email,
		&i.HashedPassword,
		&i.ProfilePicture,
		&i.Bio,
		&i.HasOnboarded,
		&i.BackgroundImage,
		&i.OnboardingAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUser = `-- name: UpdateUser :one
UPDATE users
SET 
    username = COALESCE($2, username),
    fullname = COALESCE($3, fullname),
    email = COALESCE($4, email),
    hashed_password = COALESCE($5, hashed_password)
WHERE id = $1
RETURNING id, username, fullname, email, hashed_password, profile_picture, bio, has_onboarded, background_image, onboarding_at, created_at, updated_at
`

type UpdateUserParams struct {
	ID             pgtype.UUID
	Username       string
	Fullname       pgtype.Text
	Email          string
	HashedPassword string
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUser,
		arg.ID,
		arg.Username,
		arg.Fullname,
		arg.Email,
		arg.HashedPassword,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Fullname,
		&i.Email,
		&i.HashedPassword,
		&i.ProfilePicture,
		&i.Bio,
		&i.HasOnboarded,
		&i.BackgroundImage,
		&i.OnboardingAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
