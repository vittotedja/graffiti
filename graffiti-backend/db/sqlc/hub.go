package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Hub provides all functions to execute db queries and transactions
type Hub struct {
	// a composition, a preferred way to extend struct functionality in Golang instead of inheritance
	// All individual query functions are defined in the Queries struct
	*Queries
	pool *pgxpool.Pool
}

func NewHub(pool *pgxpool.Pool) *Hub {
	return &Hub{
		pool:    pool,
		Queries: New(pool),
	}
}

// execTx executes a function within a database transaction
// It rolls back the transaction if the function returns an error
func (hub *Hub) execTx(ctx context.Context, fn func(*Queries) error) error {
	// Create empty TxOptions for default options
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
func (hub *Hub) CreateFriendRequestTx(ctx context.Context, fromUser, toUser pgtype.UUID) (Friendship, error) {
    var friendship Friendship

    err := hub.execTx(ctx, func(q *Queries) error {
        // Check if the toUser has blocked the fromUser
        isBlocked, err := hub.IsUserBlockedTx(ctx, toUser, fromUser)
        if err != nil {
            return err
        }
        if isBlocked {
            return fmt.Errorf("cannot send friend request, you are blocked by the user")
        }

        // Check if a relationship already exists in either direction
        existingFriendships, err := q.ListFriendshipsByUserId(ctx, fromUser)
        if err != nil {
            return err
        }

        // Check for existing relationships in either direction
        for _, f := range existingFriendships {
            if (f.FromUser == fromUser && f.ToUser == toUser) || 
               (f.FromUser == toUser && f.ToUser == fromUser) {
                return fmt.Errorf("a relationship already exists between these users")
            }
        }

        // Create the friend request (fromUser -> toUser with pending status)
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

func (hub *Hub) AcceptFriendRequestTx(ctx context.Context, friendshipID pgtype.UUID) error {
    return hub.execTx(ctx, func(q *Queries) error {
        // Get the friendship to accept
        friendship, err := q.GetFriendship(ctx, friendshipID)
        if err != nil {
            return err
        }

        // Verify it's a pending request
        if friendship.Status.Status != "pending" {
            return fmt.Errorf("friendship is not in pending state")
        }

        // Create a reciprocal friendship record
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

        // Update the original friendship to friends status
        _, err = q.AcceptFriendship(ctx, friendshipID)
        
        return err
    })
}

// BlockUserTx blocks a user
func (hub *Hub) BlockUserTx(ctx context.Context, fromUser, toUser pgtype.UUID) error {
    return hub.execTx(ctx, func(q *Queries) error {
        
       	// First, try to find an existing friendship between these users
	   	existingFriendship, err := q.ListFriendshipByUserPairs(ctx, ListFriendshipByUserPairsParams{
			FromUser: fromUser,
			ToUser:   toUser,
		})
	
		// If a friendship exists, block it
		if err == nil {
			_, err = q.BlockFriendship(ctx, existingFriendship.ID)
			if err != nil {
				return err
			}
			return nil
		}

		// If no existing friendship is found, create a new blocked friendship
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
func (hub *Hub) UnblockUserTx(ctx context.Context, fromUser, toUser pgtype.UUID) error {
    return hub.execTx(ctx, func(q *Queries) error {
        // Find existing friendship between users
        existingFriendship, err := q.ListFriendshipByUserPairs(ctx, ListFriendshipByUserPairsParams{
            FromUser: fromUser,
            ToUser:   toUser,
        })

		// If no existing friendship or the existing one isn't blocked, return
        if err != nil || existingFriendship.Status.Status != "blocked" {
            return fmt.Errorf("no blocked relationship to unblock")
        }

        // If there is a blocked relationship, delete it
		err = q.DeleteFriendship(ctx, existingFriendship.ID)

        return err
    })
}

// GetFriendsTx returns all accepted friendships for a user
func (hub *Hub) GetFriendsTx(ctx context.Context, userID pgtype.UUID) ([]Friendship, error) {
    var friends []Friendship
    
    err := hub.execTx(ctx, func(q *Queries) error {
        // Get all relationships for this user
        relationships, err := q.ListFriendshipsByUserId(ctx, userID)
        if err != nil {
            return err
        }
        
        // Find friendships (status = "friends") where the user is either initiator or recipient
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
func (hub *Hub) IsFriendTx(ctx context.Context, userID, otherUserID pgtype.UUID) (bool, error) {
    isFriend := false
    
    err := hub.execTx(ctx, func(q *Queries) error {
        // Get all relationships for this user
        relationships, err := q.ListFriendshipsByUserId(ctx, userID)
        if err != nil {
            return err
        }
        
        // Check if there's a friendship relationship
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
func (hub *Hub) GetPendingFriendRequestsTx(ctx context.Context, userID pgtype.UUID) ([]Friendship, error) {
    var pendingRequests []Friendship
    
    err := hub.execTx(ctx, func(q *Queries) error {
        // Get all relationships for this user
        relationships, err := q.ListFriendshipsByUserId(ctx, userID)
        if err != nil {
            return err
        }
        
        // Find pending requests where the user is the recipient
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
func (hub *Hub) GetSentFriendRequestsTx(ctx context.Context, userID pgtype.UUID) ([]Friendship, error) {
    var sentRequests []Friendship
    
    err := hub.execTx(ctx, func(q *Queries) error {
        // Get all relationships for this user
        relationships, err := q.ListFriendshipsByUserId(ctx, userID)
        if err != nil {
            return err
        }
        
        // Find pending requests where the user is the sender
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
func (hub *Hub) IsUserBlockedTx(ctx context.Context, fromUser, toUser pgtype.UUID) (bool, error) {
    var isBlocked bool
    
    err := hub.execTx(ctx, func(q *Queries) error {
        // Get all relationships for this user
        relationships, err := q.ListFriendshipsByUserId(ctx, fromUser)
        if err != nil {
            return err
        }
        
        // Check if there's a block relationship
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
