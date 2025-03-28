// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: friendship.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const acceptFriendship = `-- name: AcceptFriendship :one
UPDATE friendships
  SET status = 'friends'
WHERE id = $1
RETURNING id, from_user, to_user, status, created_at, updated_at
`

func (q *Queries) AcceptFriendship(ctx context.Context, id pgtype.UUID) (Friendship, error) {
	row := q.db.QueryRow(ctx, acceptFriendship, id)
	var i Friendship
	err := row.Scan(
		&i.ID,
		&i.FromUser,
		&i.ToUser,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const blockFriendship = `-- name: BlockFriendship :one
UPDATE friendships
  SET status = 'blocked'
WHERE id = $1
RETURNING id, from_user, to_user, status, created_at, updated_at
`

func (q *Queries) BlockFriendship(ctx context.Context, id pgtype.UUID) (Friendship, error) {
	row := q.db.QueryRow(ctx, blockFriendship, id)
	var i Friendship
	err := row.Scan(
		&i.ID,
		&i.FromUser,
		&i.ToUser,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createFriendship = `-- name: CreateFriendship :one
INSERT INTO friendships(
 from_user,
 to_user,
 status
) VALUES (
  $1, $2, $3
) RETURNING id, from_user, to_user, status, created_at, updated_at
`

type CreateFriendshipParams struct {
	FromUser pgtype.UUID
	ToUser   pgtype.UUID
	Status   NullStatus
}

func (q *Queries) CreateFriendship(ctx context.Context, arg CreateFriendshipParams) (Friendship, error) {
	row := q.db.QueryRow(ctx, createFriendship, arg.FromUser, arg.ToUser, arg.Status)
	var i Friendship
	err := row.Scan(
		&i.ID,
		&i.FromUser,
		&i.ToUser,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteFriendship = `-- name: DeleteFriendship :exec
DELETE FROM friendships
WHERE id = $1
`

func (q *Queries) DeleteFriendship(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteFriendship, id)
	return err
}

const getFriendship = `-- name: GetFriendship :one
SELECT id, from_user, to_user, status, created_at, updated_at FROM friendships
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetFriendship(ctx context.Context, id pgtype.UUID) (Friendship, error) {
	row := q.db.QueryRow(ctx, getFriendship, id)
	var i Friendship
	err := row.Scan(
		&i.ID,
		&i.FromUser,
		&i.ToUser,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getNumberOfFriends = `-- name: GetNumberOfFriends :one
SELECT COUNT(*) FROM friendships
WHERE ((from_user = $1) OR (to_user = $1)) AND status = 'friends'
`

func (q *Queries) GetNumberOfFriends(ctx context.Context, fromUser pgtype.UUID) (int64, error) {
	row := q.db.QueryRow(ctx, getNumberOfFriends, fromUser)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getNumberOfPendingFriendRequests = `-- name: GetNumberOfPendingFriendRequests :one
SELECT COUNT(*) FROM friendships
WHERE to_user = $1 AND status = 'pending'
`

func (q *Queries) GetNumberOfPendingFriendRequests(ctx context.Context, toUser pgtype.UUID) (int64, error) {
	row := q.db.QueryRow(ctx, getNumberOfPendingFriendRequests, toUser)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const listFriendshipByUserPairs = `-- name: ListFriendshipByUserPairs :one
SELECT id, from_user, to_user, status, created_at, updated_at FROM friendships
WHERE (from_user = $1 AND to_user = $2) OR (from_user = $2 AND to_user = $1)
`

type ListFriendshipByUserPairsParams struct {
	FromUser pgtype.UUID
	ToUser   pgtype.UUID
}

func (q *Queries) ListFriendshipByUserPairs(ctx context.Context, arg ListFriendshipByUserPairsParams) (Friendship, error) {
	row := q.db.QueryRow(ctx, listFriendshipByUserPairs, arg.FromUser, arg.ToUser)
	var i Friendship
	err := row.Scan(
		&i.ID,
		&i.FromUser,
		&i.ToUser,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listFriendships = `-- name: ListFriendships :many
SELECT id, from_user, to_user, status, created_at, updated_at FROM friendships
ORDER BY id
`

func (q *Queries) ListFriendships(ctx context.Context) ([]Friendship, error) {
	rows, err := q.db.Query(ctx, listFriendships)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Friendship
	for rows.Next() {
		var i Friendship
		if err := rows.Scan(
			&i.ID,
			&i.FromUser,
			&i.ToUser,
			&i.Status,
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

const listFriendshipsByUserId = `-- name: ListFriendshipsByUserId :many
SELECT id, from_user, to_user, status, created_at, updated_at FROM friendships
WHERE (from_user = $1 OR to_user = $1)
ORDER BY id
`

func (q *Queries) ListFriendshipsByUserId(ctx context.Context, fromUser pgtype.UUID) ([]Friendship, error) {
	rows, err := q.db.Query(ctx, listFriendshipsByUserId, fromUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Friendship
	for rows.Next() {
		var i Friendship
		if err := rows.Scan(
			&i.ID,
			&i.FromUser,
			&i.ToUser,
			&i.Status,
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

const listFriendshipsByUserIdAndStatus = `-- name: ListFriendshipsByUserIdAndStatus :many
SELECT id, from_user, to_user, status, created_at, updated_at FROM friendships
WHERE (from_user = $1 OR to_user = $1) AND status = $2
ORDER BY id
`

type ListFriendshipsByUserIdAndStatusParams struct {
	FromUser pgtype.UUID
	Status   NullStatus
}

func (q *Queries) ListFriendshipsByUserIdAndStatus(ctx context.Context, arg ListFriendshipsByUserIdAndStatusParams) ([]Friendship, error) {
	rows, err := q.db.Query(ctx, listFriendshipsByUserIdAndStatus, arg.FromUser, arg.Status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Friendship
	for rows.Next() {
		var i Friendship
		if err := rows.Scan(
			&i.ID,
			&i.FromUser,
			&i.ToUser,
			&i.Status,
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

const rejectFriendship = `-- name: RejectFriendship :exec
DELETE FROM friendships
WHERE id = $1
`

func (q *Queries) RejectFriendship(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, rejectFriendship, id)
	return err
}

const updateFriendship = `-- name: UpdateFriendship :one
UPDATE friendships
  SET status = $2
WHERE id = $1
RETURNING id, from_user, to_user, status, created_at, updated_at
`

type UpdateFriendshipParams struct {
	ID     pgtype.UUID
	Status NullStatus
}

func (q *Queries) UpdateFriendship(ctx context.Context, arg UpdateFriendshipParams) (Friendship, error) {
	row := q.db.QueryRow(ctx, updateFriendship, arg.ID, arg.Status)
	var i Friendship
	err := row.Scan(
		&i.ID,
		&i.FromUser,
		&i.ToUser,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
