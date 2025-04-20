package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Hub interface {
	Querier
	CreateFriendRequestTx(ctx context.Context, fromUser, toUser pgtype.UUID) (Friendship, error)
	CreateLikeTx(ctx context.Context, postID, userID pgtype.UUID) error
	CreateOrDeleteLikeTx(ctx context.Context, postID, userID pgtype.UUID) (liked bool, err error)
	AcceptFriendRequestTx(ctx context.Context, friendshipID pgtype.UUID) error
	BlockUserTx(ctx context.Context, fromUser, toUser pgtype.UUID) error
	UnblockUserTx(ctx context.Context, fromUser, toUser pgtype.UUID) error
	GetFriendsTx(ctx context.Context, userID pgtype.UUID) ([]Friendship, error)
	IsFriendTx(ctx context.Context, userID, otherUserID pgtype.UUID) (bool, error)
	GetPendingFriendRequestsTx(ctx context.Context, userID pgtype.UUID) ([]Friendship, error)
	GetSentFriendRequestsTx(ctx context.Context, userID pgtype.UUID) ([]Friendship, error)
	IsUserBlockedTx(ctx context.Context, fromUser, toUser pgtype.UUID) (bool, error)
	RefreshMaterializedViews(ctx context.Context) error
}

// SQLHub provides all functions to execute db SQL queries and transactions
type SQLHub struct {
	// a composition, a preferred way to extend struct functionality in Golang instead of inheritance
	*Queries
	pool *pgxpool.Pool
}

func NewHub(pool *pgxpool.Pool) Hub {
	return &SQLHub{
		pool:    pool,
		Queries: New(pool),
	}
}

// execTx executes a function within a database transaction
// It rolls back the transaction if the function returns an error
func (hub *SQLHub) execTx(ctx context.Context, fn func(*Queries) error) error {
	txOptions := pgx.TxOptions{}

	tx, err := hub.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit(ctx)
}

// CreateFriendRequestTx creates a new friendship request (one-way, pending status)
func (hub *SQLHub) CreateFriendRequestTx(ctx context.Context, fromUser, toUser pgtype.UUID) (Friendship, error) {
	var friendship Friendship

	err := hub.execTx(ctx, func(q *Queries) error {
		isBlocked, err := hub.IsUserBlockedTx(ctx, toUser, fromUser)
		if err != nil {
			return err
		}
		if isBlocked {
			return fmt.Errorf("cannot send friend request, you are blocked by the user")
		}

		existingFriendships, err := q.ListFriendshipsByUserId(ctx, fromUser)
		if err != nil {
			return err
		}

		for _, f := range existingFriendships {
			if (f.FromUser == fromUser && f.ToUser == toUser) ||
				(f.FromUser == toUser && f.ToUser == fromUser) {
				return fmt.Errorf("a relationship already exists between these users")
			}
		}

		pendingStatus := NullStatus{Status: "pending", Valid: true}
		arg := CreateFriendshipParams{
			FromUser: fromUser,
			ToUser:   toUser,
			Status:   pendingStatus,
		}

		friendship, err = q.CreateFriendship(ctx, arg)
		return err
	})

	return friendship, err
}

func (hub *SQLHub) CreateLikeTx(ctx context.Context, postID, userID pgtype.UUID) error {
	err := hub.execTx(ctx, func(q *Queries) error {
		arg := CreateLikeParams{
			PostID: postID,
			UserID: userID,
		}
		_, err := q.CreateLike(ctx, arg)
		if err != nil {
			return err
		}

		_, err = q.AddLikesCount(ctx, postID)

		return err
	})

	return err
}

func (hub *SQLHub) CreateOrDeleteLikeTx(ctx context.Context, postID, userID pgtype.UUID) (liked bool, err error) {
	err = hub.execTx(ctx, func(q *Queries) error {
		_, err := q.GetLike(ctx, GetLikeParams{
			PostID: postID,
			UserID: userID,
		})

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				if _, err := q.CreateLike(ctx, CreateLikeParams{
					PostID: postID,
					UserID: userID,
				}); err != nil {
					return err
				}
				if _, err := q.AddLikesCount(ctx, postID); err != nil {
					return err
				}
				liked = true
				return nil
			}
			return err
		}

		if err := q.DeleteLike(ctx, DeleteLikeParams{
			PostID: postID,
			UserID: userID,
		}); err != nil {
			return err
		}
		if _, err := q.RemoveLikesCount(ctx, postID); err != nil {
			return err
		}
		liked = false

		return nil
	})

	return liked, err
}

func (hub *SQLHub) AcceptFriendRequestTx(ctx context.Context, friendshipID pgtype.UUID) error {
	return hub.execTx(ctx, func(q *Queries) error {
		friendship, err := q.GetFriendship(ctx, friendshipID)
		if err != nil {
			return err
		}

		if friendship.Status.Status != "pending" {
			return fmt.Errorf("friendship is not in pending state")
		}

		friendsStatus := NullStatus{Status: "friends", Valid: true}
		reciprocalArg := CreateFriendshipParams{
			FromUser: friendship.ToUser,
			ToUser:   friendship.FromUser,
			Status:   friendsStatus,
		}
		_, err = q.CreateFriendship(ctx, reciprocalArg)
		if err != nil {
			return err
		}

		_, err = q.AcceptFriendship(ctx, friendshipID)

		return err
	})
}

// BlockUserTx blocks a user
func (hub *SQLHub) BlockUserTx(ctx context.Context, fromUser, toUser pgtype.UUID) error {
	return hub.execTx(ctx, func(q *Queries) error {

		existingFriendship, err := q.ListFriendshipByUserPairs(ctx, ListFriendshipByUserPairsParams{
			FromUser: fromUser,
			ToUser:   toUser,
		})

		if err == nil {
			_, err = q.BlockFriendship(ctx, existingFriendship.ID)
			if err != nil {
				return err
			}
			return nil
		}

		blockedStatus := NullStatus{Status: "blocked", Valid: true}
		_, err = q.CreateFriendship(ctx, CreateFriendshipParams{
			FromUser: fromUser,
			ToUser:   toUser,
			Status:   blockedStatus,
		})

		return err
	})
}

// UnblockUserTx removes a block between users
func (hub *SQLHub) UnblockUserTx(ctx context.Context, fromUser, toUser pgtype.UUID) error {
	return hub.execTx(ctx, func(q *Queries) error {
		
		existingFriendship, err := q.ListFriendshipByUserPairs(ctx, ListFriendshipByUserPairsParams{
			FromUser: fromUser,
			ToUser:   toUser,
		})

		if err != nil || existingFriendship.Status.Status != "blocked" {
			return fmt.Errorf("no blocked relationship to unblock")
		}

		err = q.DeleteFriendship(ctx, existingFriendship.ID)

		return err
	})
}

// GetFriendsTx returns all accepted friendships for a user
func (hub *SQLHub) GetFriendsTx(ctx context.Context, userID pgtype.UUID) ([]Friendship, error) {
	var friends []Friendship

	err := hub.execTx(ctx, func(q *Queries) error {
		relationships, err := q.ListFriendshipsByUserId(ctx, userID)
		if err != nil {
			return err
		}

		for _, f := range relationships {
			if f.Status.Status == "friends" &&
				(f.FromUser == userID || f.ToUser == userID) {
				friends = append(friends, f)
			}
		}

		return nil
	})

	return friends, err
}

// IsFriendTx checks if two users are friends
func (hub *SQLHub) IsFriendTx(ctx context.Context, userID, otherUserID pgtype.UUID) (bool, error) {
	isFriend := false

	err := hub.execTx(ctx, func(q *Queries) error {
		relationships, err := q.ListFriendshipsByUserId(ctx, userID)
		if err != nil {
			return err
		}

		for _, f := range relationships {
			if f.Status.Status == "friends" {
				if (f.FromUser == userID && f.ToUser == otherUserID) ||
					(f.FromUser == otherUserID && f.ToUser == userID) {
					isFriend = true
					break
				}
			}
		}

		return nil
	})

	return isFriend, err
}

// GetPendingFriendRequestsTx returns all pending friend requests received by a user
func (hub *SQLHub) GetPendingFriendRequestsTx(ctx context.Context, userID pgtype.UUID) ([]Friendship, error) {
	var pendingRequests []Friendship

	err := hub.execTx(ctx, func(q *Queries) error {
		relationships, err := q.ListFriendshipsByUserId(ctx, userID)
		if err != nil {
			return err
		}

		for _, f := range relationships {
			if f.Status.Status == "pending" && f.ToUser == userID {
				pendingRequests = append(pendingRequests, f)
			}
		}

		return nil
	})

	return pendingRequests, err
}

// GetSentFriendRequestsTx returns all pending friend requests sent by a user
func (hub *SQLHub) GetSentFriendRequestsTx(ctx context.Context, userID pgtype.UUID) ([]Friendship, error) {
	var sentRequests []Friendship

	err := hub.execTx(ctx, func(q *Queries) error {
		relationships, err := q.ListFriendshipsByUserId(ctx, userID)
		if err != nil {
			return err
		}

		for _, f := range relationships {
			if f.Status.Status == "pending" && f.FromUser == userID {
				sentRequests = append(sentRequests, f)
			}
		}

		return nil
	})

	return sentRequests, err
}

// IsUserBlockedTx checks if toUser is blocked by fromUser
func (hub *SQLHub) IsUserBlockedTx(ctx context.Context, fromUser, toUser pgtype.UUID) (bool, error) {
	var isBlocked bool

	err := hub.execTx(ctx, func(q *Queries) error {
		relationships, err := q.ListFriendshipsByUserId(ctx, fromUser)
		if err != nil {
			return err
		}

		for _, f := range relationships {
			if f.FromUser == fromUser && f.ToUser == toUser && f.Status.Status == "blocked" {
				isBlocked = true
				break
			}
		}

		return nil
	})

	return isBlocked, err
}

func (h *SQLHub) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return h.db.Exec(ctx, sql, args...)
}

func (h *SQLHub) RefreshMaterializedViews(ctx context.Context) error {
	_, err := h.db.Exec(ctx, "REFRESH MATERIALIZED VIEW accepted_friendships_mv")
	return err
}

// CountUnreadNotifications counts unread notifications for a user
func (h *SQLHub) CountUnreadNotifications(ctx context.Context, userID pgtype.UUID) (int64, error) {
	var count int64
	err := h.db.QueryRow(ctx, "SELECT COUNT(*) FROM notifications WHERE recipient_id = $1 AND is_read = false", userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}