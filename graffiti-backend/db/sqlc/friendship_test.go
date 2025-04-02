package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomFriendship(t *testing.T) Friendship {
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)

	status := NullStatus{
		Status: "pending",
		Valid:  true,
	}

	friendshipParams := CreateFriendshipParams{
		FromUser: user1.ID,
		ToUser:   user2.ID,
		Status:   status,
	}

	friendship, err := testHub.CreateFriendship(context.Background(), friendshipParams)
	require.NoError(t, err, "Error occurred while creating friendship")
	require.NotEmpty(t, friendship.ID, "Friendship ID should not be empty")
	require.Equal(t, user1.ID, friendship.FromUser, "FromUser should be the same as the created user1")
	require.Equal(t, user2.ID, friendship.ToUser, "ToUser should be the same as the created user2")
	require.Equal(t, status.Status, friendship.Status.Status, "Friendship status should match the one set")

	return friendship
}

func cleanupFriendships(t *testing.T) {
	friendships, err := testHub.ListFriendships(context.Background())
	require.NoError(t, err, "Error occurred while listing friendships before cleanup")

	for _, friendship := range friendships {
		err := testHub.DeleteFriendship(context.Background(), friendship.ID)
		require.NoError(t, err, "Error occurred while deleting friendship")
	}

	remainingFriendships, err := testHub.ListFriendships(context.Background())
	require.NoError(t, err, "Error occurred while listing friendships after cleanup")
	require.Len(t, remainingFriendships, 0, "There are still friendships in the database after cleanup")
}

func TestCreateGetDeleteFriendship(t *testing.T) {
	friendship := createRandomFriendship(t)

	fetchedFriendship, err := testHub.GetFriendship(context.Background(), friendship.ID)
	require.NoError(t, err, "Error occurred while getting the friendship")
	require.Equal(t, friendship.ID, fetchedFriendship.ID, "Fetched friendship ID should match the created friendship ID")
	require.Equal(t, friendship.FromUser, fetchedFriendship.FromUser, "FromUser should match")
	require.Equal(t, friendship.ToUser, fetchedFriendship.ToUser, "ToUser should match")
	require.Equal(t, friendship.Status.Status, fetchedFriendship.Status.Status, "Status should match")

	err = testHub.DeleteFriendship(context.Background(), friendship.ID)
	require.NoError(t, err, "Error occurred while deleting the friendship")

	deletedFriendship, err := testHub.GetFriendship(context.Background(), friendship.ID)
	require.Error(t, err, "Getting a deleted friendship should return an error")
	require.Equal(t, pgtype.UUID{}, deletedFriendship.ID, "Deleted friendship ID should be empty")
}

func TestListFriendshipsAndListFriendshipsByUserId(t *testing.T) {
	defer cleanupFriendships(t)

	friendships, err := testHub.ListFriendships(context.Background())
	require.NoError(t, err)
	require.Len(t, friendships, 0, "The number of friendships should be 0 initially")

	for i := 0; i < 5; i++ {
		createRandomFriendship(t)
	}

	randomUser := createRandomUser(t)

	for i := 0; i < 3; i++ {
		anotherUser := createRandomUser(t)

		status := NullStatus{
			Status: "pending",
			Valid:  true,
		}

		friendshipParams := CreateFriendshipParams{
			FromUser: randomUser.ID,
			ToUser:   anotherUser.ID,
			Status:   status,
		}

		_, err := testHub.CreateFriendship(context.Background(), friendshipParams)
		require.NoError(t, err, "Error occurred while creating friendship")
	}

	friendships, err = testHub.ListFriendships(context.Background())
	require.NoError(t, err)
	require.Len(t, friendships, 8, "The number of friendships should be 8")

	friendshipsByUser, err := testHub.ListFriendshipsByUserId(context.Background(), randomUser.ID)
	require.NoError(t, err)
	require.Len(t, friendshipsByUser, 3, "The number of friendships for the random user should be 3")
}

func TestAcceptBlockRejectFriendship(t *testing.T) {
	friendship := createRandomFriendship(t)

	acceptedFriendship, err := testHub.AcceptFriendship(context.Background(), friendship.ID)
	require.NoError(t, err, "Error occurred while accepting the friendship")
	require.True(t, acceptedFriendship.Status.Valid, "Status should be valid after acceptance")
	require.Equal(t, "friends", string(acceptedFriendship.Status.Status), "Status should be 'friends' after acceptance")

	blockedFriendship, err := testHub.BlockFriendship(context.Background(), friendship.ID)
	require.NoError(t, err, "Error occurred while blocking the friendship")
	require.True(t, blockedFriendship.Status.Valid, "Status should be valid after blocking")
	require.Equal(t, "blocked", string(blockedFriendship.Status.Status), "Status should be 'blocked' after blocking")

	err = testHub.RejectFriendship(context.Background(), friendship.ID)
	require.NoError(t, err, "Error occurred while rejecting the friendship")

	_, err = testHub.GetFriendship(context.Background(), friendship.ID)
	require.Error(t, err, "Fetching a rejected friendship should return an error")
}

func TestListFriendshipsByUserIdAndStatus(t *testing.T) {
	defer cleanupFriendships(t)

	randomUser := createRandomUser(t)

	for i := 0; i < 3; i++ {
		anotherUser := createRandomUser(t)
		status := NullStatus{
			Status: "pending",
			Valid:  true,
		}

		friendshipParams := CreateFriendshipParams{
			FromUser: randomUser.ID,
			ToUser:   anotherUser.ID,
			Status:   status,
		}

		_, err := testHub.CreateFriendship(context.Background(), friendshipParams)
		require.NoError(t, err, "Error occurred while creating pending friendship")
	}

	for i := 0; i < 1; i++ {
		anotherUser := createRandomUser(t)
		status := NullStatus{
			Status: "friends",
			Valid:  true,
		}

		friendshipParams := CreateFriendshipParams{
			FromUser: randomUser.ID,
			ToUser:   anotherUser.ID,
			Status:   status,
		}

		_, err := testHub.CreateFriendship(context.Background(), friendshipParams)
		require.NoError(t, err, "Error occurred while creating accepted friendship")
	}

	for i := 0; i < 2; i++ {
		anotherUser := createRandomUser(t)
		status := NullStatus{
			Status: "blocked",
			Valid:  true,
		}

		friendshipParams := CreateFriendshipParams{
			FromUser: randomUser.ID,
			ToUser:   anotherUser.ID,
			Status:   status,
		}

		_, err := testHub.CreateFriendship(context.Background(), friendshipParams)
		require.NoError(t, err, "Error occurred while creating blocked friendship")
	}

	for i := 0; i < 2; i++ {
		createRandomFriendship(t)
	}

	pendingFriendships, err := testHub.ListFriendshipsByUserIdAndStatus(context.Background(), ListFriendshipsByUserIdAndStatusParams{
		FromUser: randomUser.ID,
		Status: NullStatus{
			Status: "pending",
			Valid:  true,
		},
	})
	require.NoError(t, err, "Error occurred while fetching pending friendships")
	require.Len(t, pendingFriendships, 3, "The number of pending friendships should be 3")

	acceptedFriendships, err := testHub.ListFriendshipsByUserIdAndStatus(context.Background(), ListFriendshipsByUserIdAndStatusParams{
		FromUser: randomUser.ID,
		Status: NullStatus{
			Status: "friends",
			Valid:  true,
		},
	})
	require.NoError(t, err, "Error occurred while fetching accepted friendships")
	require.Len(t, acceptedFriendships, 1, "The number of accepted friendships should be 1")

	blockedFriendships, err := testHub.ListFriendshipsByUserIdAndStatus(context.Background(), ListFriendshipsByUserIdAndStatusParams{
		FromUser: randomUser.ID,
		Status: NullStatus{
			Status: "blocked",
			Valid:  true,
		},
	})
	require.NoError(t, err, "Error occurred while fetching blocked friendships")
	require.Len(t, blockedFriendships, 2, "The number of blocked friendships should be 2")
}

func TestGetNumberOfFriendsAndGetNumberOfPendingFriendRequests(t *testing.T) {
	var createdFriendships []Friendship
	defer cleanupFriendships(t)

	user := createRandomUser(t)

	for i := 0; i < 5; i++ {
		friendshipParams := CreateFriendshipParams{
			FromUser: createRandomUser(t).ID,
			ToUser:   user.ID,
			Status:   NullStatus{Status: "pending", Valid: true},
		}
		friendship, err := testHub.CreateFriendship(context.Background(), friendshipParams)
		require.NoError(t, err, "Error occurred while creating friendship")
		createdFriendships = append(createdFriendships, friendship)
	}

	pendingRequests, err := testHub.GetNumberOfPendingFriendRequests(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(5), pendingRequests, "The number of pending friend requests should be 5")

	friends, err := testHub.GetNumberOfFriends(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), friends, "The number of friends should be 0 initially")

	for i := 0; i < 3; i++ {
		acceptedFriendship, err := testHub.AcceptFriendship(context.Background(), createdFriendships[i].ID)
		require.NoError(t, err, "Error occurred while accepting the friendship")
		require.Equal(t, "friends", string(acceptedFriendship.Status.Status), "The friendship status should be 'friends' after acceptance")
	}

	pendingRequests, err = testHub.GetNumberOfPendingFriendRequests(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), pendingRequests, "The number of pending friend requests should be 2 after accepting 3")

	friends, err = testHub.GetNumberOfFriends(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(3), friends, "The number of friends should be 3 after accepting 3 friendships")
}

func TestListFriendshipByUserPairs(t *testing.T) {
	defer cleanupFriendships(t)

	user1 := createRandomUser(t)
	user2 := createRandomUser(t)

	status := NullStatus{
		Status: "pending",
		Valid:  true,
	}

	friendshipParams := CreateFriendshipParams{
		FromUser: user1.ID,
		ToUser:   user2.ID,
		Status:   status,
	}

	friendship, err := testHub.CreateFriendship(context.Background(), friendshipParams)
	require.NoError(t, err, "Error occurred while creating friendship")

	retrievedFriendship, err := testHub.ListFriendshipByUserPairs(context.Background(), ListFriendshipByUserPairsParams{
		FromUser: user1.ID,
		ToUser:   user2.ID,
	})

	require.NoError(t, err, "Error occurred while retrieving friendship")

	require.Equal(t, friendship.ID, retrievedFriendship.ID, "Friendship IDs should match")
	require.Equal(t, user1.ID, retrievedFriendship.FromUser, "FromUser should be the same as user1")
	require.Equal(t, user2.ID, retrievedFriendship.ToUser, "ToUser should be the same as user2")
	require.Equal(t, status.Status, retrievedFriendship.Status.Status, "Friendship status should match the one set")
}

func TestUpdateFriendship(t *testing.T) {
	defer cleanupFriendships(t)

	user1 := createRandomUser(t)
	user2 := createRandomUser(t)

	status := NullStatus{
		Status: "pending",
		Valid:  true,
	}

	friendshipParams := CreateFriendshipParams{
		FromUser: user1.ID,
		ToUser:   user2.ID,
		Status:   status,
	}

	friendship, err := testHub.CreateFriendship(context.Background(), friendshipParams)
	require.NoError(t, err, "Error occurred while creating friendship")
	require.NotEmpty(t, friendship.ID, "Friendship ID should not be empty")
	require.Equal(t, user1.ID, friendship.FromUser, "FromUser should be the same as user1")
	require.Equal(t, user2.ID, friendship.ToUser, "ToUser should be the same as user2")
	require.Equal(t, status.Status, friendship.Status.Status, "Friendship status should match the one set")

	updatedStatus := NullStatus{
		Status: "friends", 
		Valid:  true,
	}

	updatedFriendship, err := testHub.UpdateFriendship(context.Background(), UpdateFriendshipParams{
		ID:     friendship.ID,
		Status: updatedStatus,
	})

	require.NoError(t, err, "Error occurred while updating friendship status")
	require.Equal(t, "friends", string(updatedFriendship.Status.Status), "Friendship status should be 'friends' after update")
	require.Equal(t, friendship.ID, updatedFriendship.ID, "Friendship ID should remain the same")

	blockedStatus := NullStatus{
		Status: "blocked", 
		Valid:  true,
	}

	blockedFriendship, err := testHub.UpdateFriendship(context.Background(), UpdateFriendshipParams{
		ID:     friendship.ID,
		Status: blockedStatus,
	})

	require.NoError(t, err, "Error occurred while updating friendship status to 'blocked'")
	require.Equal(t, "blocked", string(blockedFriendship.Status.Status), "Friendship status should be 'blocked' after update")
	require.Equal(t, friendship.ID, blockedFriendship.ID, "Friendship ID should remain the same")

	pendingStatus := NullStatus{
		Status: "pending",
		Valid:  true,
	}

	resetFriendship, err := testHub.UpdateFriendship(context.Background(), UpdateFriendshipParams{
		ID:     friendship.ID,
		Status: pendingStatus,
	})

	require.NoError(t, err, "Error occurred while updating friendship status to 'pending'")
	require.Equal(t, "pending", string(resetFriendship.Status.Status), "Friendship status should be 'pending' after reset")
	require.Equal(t, friendship.ID, resetFriendship.ID, "Friendship ID should remain the same")
}