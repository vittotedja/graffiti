package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomLike(t *testing.T) Like {
	user := createRandomUser(t)
	post := createRandomPost(t)

	arg := CreateLikeParams{
		PostID: post.ID,
		UserID: user.ID,
	}

	like, err := testHub.CreateLike(context.Background(), arg)

	require.NoError(t, err, "Error occurred while creating the like")
	require.NotEmpty(t, like.ID, "Like ID should not be empty")
	require.Equal(t, post.ID, like.PostID, "The PostID of the like should match the created post")
	require.Equal(t, user.ID, like.UserID, "The UserID of the like should match the created user")
	require.NotEmpty(t, like.LikedAt, "The LikedAt timestamp should not be empty")

	return like
}

func TestCreateGetDeleteLike(t *testing.T) {
    like := createRandomLike(t)

    fetchedLike, err := testHub.GetLike(context.Background(), GetLikeParams{
		PostID: like.PostID,
		UserID: like.UserID,
	})
    require.NoError(t, err, "Error occurred while fetching the like")
    require.Equal(t, like.ID, fetchedLike.ID, "Fetched like ID should match the created like ID")
    require.Equal(t, like.PostID, fetchedLike.PostID, "Fetched like PostID should match the created like PostID")
    require.Equal(t, like.UserID, fetchedLike.UserID, "Fetched like UserID should match the created like UserID")

    err = testHub.DeleteLike(context.Background(), DeleteLikeParams{
        PostID: like.PostID,
        UserID: like.UserID,
    })
    require.NoError(t, err, "Error occurred while deleting the like")

    deletedLike, err := testHub.GetLike(context.Background(), GetLikeParams{
        PostID: like.PostID,
        UserID: like.UserID,
    })
    require.Error(t, err, "Like should have been deleted and should not be found")
    require.Equal(t, pgtype.UUID{}, deletedLike.ID, "Deleted like ID should be empty")
}

func TestListLikes(t *testing.T) {
	initialLikes, err := testHub.ListLikes(context.Background())
	require.NoError(t, err)
	require.Len(t, initialLikes, 0, "Initial likes count should be 0")

	var createdLikes []Like
	for i := 0; i < 5; i++ {
		like := createRandomLike(t)
		createdLikes = append(createdLikes, like)
	}

	allLikes, err := testHub.ListLikes(context.Background())
	require.NoError(t, err)
	require.Len(t, allLikes, 5, "The number of likes after creating 5 likes should be 5")

	for _, like := range createdLikes {
		err := testHub.DeleteLike(context.Background(), DeleteLikeParams{
			PostID: like.PostID,
			UserID: like.UserID,
		})
		require.NoError(t, err, "Error occurred while deleting like")
	}

	finalLikes, err := testHub.ListLikes(context.Background())
	require.NoError(t, err)
	require.Len(t, finalLikes, 0, "After deletion, likes count should be 0")
}

func TestListLikesByPost(t *testing.T) {
    post := createRandomPost(t)

    var createdLikesForPost []Like
    for i := 0; i < 3; i++ {
        arg := CreateLikeParams{
            PostID: post.ID,
            UserID: createRandomUser(t).ID,
        }
        like, err := testHub.CreateLike(context.Background(), arg)
        require.NoError(t, err, "Error occurred while creating the like")
        createdLikesForPost = append(createdLikesForPost, like)
    }

    otherPost := createRandomPost(t)
    otherLikeArg := CreateLikeParams{
        PostID: otherPost.ID,
        UserID: createRandomUser(t).ID,
    }
    _, err := testHub.CreateLike(context.Background(), otherLikeArg)
    require.NoError(t, err, "Error occurred while creating the random like for the other post")

    likesForPost, err := testHub.ListLikesByPost(context.Background(), post.ID)
    require.NoError(t, err)
    require.Len(t, likesForPost, 3, "The number of likes for the post should be 3")

    // CLean up
    for _, like := range createdLikesForPost {
        err := testHub.DeleteLike(context.Background(), DeleteLikeParams{
            PostID: like.PostID,
            UserID: like.UserID,
        })
        require.NoError(t, err, "Error occurred while deleting like")
    }

    err = testHub.DeleteLike(context.Background(), DeleteLikeParams{
        PostID: otherPost.ID,
        UserID: otherLikeArg.UserID,
    })
    require.NoError(t, err, "Error occurred while deleting like for the other post")

    err = testHub.DeletePost(context.Background(), post.ID)
    require.NoError(t, err, "Error occurred while deleting the post")

    err = testHub.DeletePost(context.Background(), otherPost.ID)
    require.NoError(t, err, "Error occurred while deleting the other post")
}

func TestListLikesByUser(t *testing.T) {
    user1 := createRandomUser(t)
    user2 := createRandomUser(t)

    var posts []Post
    for i := 0; i < 3; i++ {
        post := createRandomPost(t)
        posts = append(posts, post)
    }

    for _, post := range posts {
        likeArg := CreateLikeParams{
            PostID: post.ID,
            UserID: user1.ID,
        }
        _, err := testHub.CreateLike(context.Background(), likeArg)
        require.NoError(t, err, "Error occurred while creating like for user1")
    }

    for _, post := range posts {
        likeArg := CreateLikeParams{
            PostID: post.ID,
            UserID: user2.ID, 
        }
        _, err := testHub.CreateLike(context.Background(), likeArg)
        require.NoError(t, err, "Error occurred while creating like for user2")
    }

    likesByUser1, err := testHub.ListLikesByUser(context.Background(), user1.ID)
    require.NoError(t, err)
    require.Len(t, likesByUser1, 3, "The number of likes by user1 should be 3")

    for _, like := range likesByUser1 {
        require.Equal(t, user1.ID, like.UserID, "Like should be associated with user1")
    }

    // Cleanup
    for _, post := range posts {
        err := testHub.DeleteLike(context.Background(), DeleteLikeParams{
            PostID: post.ID,
            UserID: user1.ID,
        })
        require.NoError(t, err, "Error occurred while deleting like for user1")

        err = testHub.DeleteLike(context.Background(), DeleteLikeParams{
            PostID: post.ID,
            UserID: user2.ID,
        })
        require.NoError(t, err, "Error occurred while deleting like for user2")
    }

	for _, post := range posts {
        err := testHub.DeletePost(context.Background(), post.ID)
        require.NoError(t, err, "Error occurred while deleting the post")
    }
}

func TestGetNumberOfLikesByPost(t *testing.T) {
	postA := createRandomPost(t)
	postB := createRandomPost(t)

	var usersA []User
	for i := 0; i < 5; i++ {
		user := createRandomUser(t)
		likeArg := CreateLikeParams{
			PostID: postA.ID,
			UserID: user.ID, 
		}
		_, err := testHub.CreateLike(context.Background(), likeArg)
		require.NoError(t, err, "Error occurred while creating like for postA")
		usersA = append(usersA, user)
	}

	var usersB []User
	for i := 0; i < 3; i++ {
		user := createRandomUser(t)
		likeArg := CreateLikeParams{
			PostID: postB.ID,
			UserID: user.ID, 
		}
		_, err := testHub.CreateLike(context.Background(), likeArg)
		require.NoError(t, err, "Error occurred while creating like for postB")
		usersB = append(usersB, user)
	}

	likesForPostA, err := testHub.ListLikesByPost(context.Background(), postA.ID)
	require.NoError(t, err)
	require.Len(t, likesForPostA, 5, "The number of likes for postA should be 5")

	likesForPostB, err := testHub.ListLikesByPost(context.Background(), postB.ID)
	require.NoError(t, err)
	require.Len(t, likesForPostB, 3, "The number of likes for postB should be 3")

	// Cleanup:
	for _, user := range usersA {
		err := testHub.DeleteLike(context.Background(), DeleteLikeParams{
			PostID: postA.ID,
			UserID: user.ID,
		})
		require.NoError(t, err, "Error occurred while deleting like for postA")
	}

	for _, user := range usersB {
		err := testHub.DeleteLike(context.Background(), DeleteLikeParams{
			PostID: postB.ID,
			UserID: user.ID,
		})
		require.NoError(t, err, "Error occurred while deleting like for postB")
	}

	err = testHub.DeletePost(context.Background(), postA.ID)
	require.NoError(t, err, "Error occurred while deleting postA")
	err = testHub.DeletePost(context.Background(), postB.ID)
	require.NoError(t, err, "Error occurred while deleting postB")
}