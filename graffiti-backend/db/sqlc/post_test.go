package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

func createRandomPost(t *testing.T) Post {
	wall := createRandomWall(t)
	user := createRandomUser(t)

	arg := CreatePostParams{
		WallID:   wall.ID,
		Author:   user.ID,
		MediaUrl: pgtype.Text{String: "https://example.com/media/" + util.RandomString(10) + ".jpg", Valid: true},
		PostType: NullPostType{
			PostType: "media",
			Valid:    true,
		},
	}

	post, err := testHub.CreatePost(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, post)

	require.Equal(t, arg.WallID, post.WallID)
	require.Equal(t, arg.Author, post.Author)
	require.Equal(t, arg.MediaUrl.String, post.MediaUrl.String)
	require.Equal(t, arg.PostType.PostType, post.PostType.PostType)
	require.False(t, post.IsHighlighted.Bool)
	require.Equal(t, pgtype.Int4(pgtype.Int4{Int32:0, Valid:true}), post.LikesCount)
	require.False(t, post.IsDeleted.Bool)
	require.NotZero(t, post.CreatedAt)

	return post
}

func TestCreatePost(t *testing.T) {
	createRandomPost(t)
}

func TestGetPost(t *testing.T) {
	post1 := createRandomPost(t)

	post2, err := testHub.GetPost(context.Background(), post1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, post2)

	require.Equal(t, post1.ID, post2.ID)
	require.Equal(t, post1.WallID, post2.WallID)
	require.Equal(t, post1.Author, post2.Author)
	require.Equal(t, post1.MediaUrl.String, post2.MediaUrl.String)
	require.Equal(t, post1.PostType.PostType, post2.PostType.PostType)
	require.Equal(t, post1.IsHighlighted.Bool, post2.IsHighlighted.Bool)
	require.Equal(t, post1.LikesCount, post2.LikesCount)
	require.Equal(t, post1.IsDeleted.Bool, post2.IsDeleted.Bool)
	require.Equal(t, post1.CreatedAt.Time.Unix(), post2.CreatedAt.Time.Unix())
}

func TestGetNonExistentPost(t *testing.T) {
	nonExistentID := pgtype.UUID{
		Bytes: [16]byte{},
		Valid: true,
	}

	_, err := testHub.GetPost(context.Background(), nonExistentID)
	require.Error(t, err, "Should return error for non-existent post")
}

func TestUpdatePost(t *testing.T) {
	post := createRandomPost(t)

	testCases := []struct {
		name           string
		updateField    string
		updateFunction func(oldPost Post) UpdatePostParams
		verifyUpdate   func(t *testing.T, oldPost, updatedPost Post)
	}{
		{
			name:        "Update MediaUrl",
			updateField: "media_url",
			updateFunction: func(oldPost Post) UpdatePostParams {
				return UpdatePostParams{
					ID:       oldPost.ID,
					MediaUrl: pgtype.Text{String: "https://example.com/media/" + util.RandomString(10) + ".jpg", Valid: true},
					PostType: NullPostType{Valid: false},
				}
			},
			verifyUpdate: func(t *testing.T, oldPost, updatedPost Post) {
				require.NotEqual(t, oldPost.MediaUrl.String, updatedPost.MediaUrl.String)
				require.Equal(t, oldPost.PostType.PostType, updatedPost.PostType.PostType)
			},
		},
		{
			name:        "Update PostType",
			updateField: "post_type",
			updateFunction: func(oldPost Post) UpdatePostParams {
				// Assuming PostType enum has more than one value, like VIDEO
				return UpdatePostParams{
					ID:       oldPost.ID,
					MediaUrl: pgtype.Text{Valid: false},
					PostType: NullPostType{PostType: "embed_link", Valid: true},
				}
			},
			verifyUpdate: func(t *testing.T, oldPost, updatedPost Post) {
				require.Equal(t, oldPost.MediaUrl.String, updatedPost.MediaUrl.String)
				require.NotEqual(t, oldPost.PostType.PostType, updatedPost.PostType.PostType)
			},
		},
		{
			name:        "Update Both Fields",
			updateField: "both_fields",
			updateFunction: func(oldPost Post) UpdatePostParams {
				return UpdatePostParams{
					ID:       oldPost.ID,
					MediaUrl: pgtype.Text{String: "https://example.com/media/" + util.RandomString(10) + ".jpg", Valid: true},
					PostType: NullPostType{PostType: "media", Valid: true},
				}
			},
			verifyUpdate: func(t *testing.T, oldPost, updatedPost Post) {
				require.NotEqual(t, oldPost.MediaUrl.String, updatedPost.MediaUrl.String)
				require.NotEqual(t, oldPost.PostType.PostType, updatedPost.PostType.PostType)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			updateParams := tc.updateFunction(post)
			updatedPost, err := testHub.UpdatePost(context.Background(), updateParams)
						
			require.NoError(t, err, "Should update %s successfully", tc.updateField)
			tc.verifyUpdate(t, post, updatedPost)

			// Update our reference post for next test
			post = updatedPost
		})
	}
}

func TestHighlightAndUnhighlightPost(t *testing.T) {
	post := createRandomPost(t)
	require.False(t, post.IsHighlighted.Bool)
	
	// Test highlighting the post
	highlightedPost, err := testHub.HighlightPost(context.Background(), post.ID)
	require.NoError(t, err)
	require.True(t, highlightedPost.IsHighlighted.Bool)

	// Test unhighlighting the post
	unhighlightedPost, err := testHub.UnhighlightPost(context.Background(), post.ID)
	require.NoError(t, err)
	require.False(t, unhighlightedPost.IsHighlighted.Bool)
}

func TestAddLikesCount(t *testing.T) {
	post := createRandomPost(t)
	initialLikes := post.LikesCount.Int32
	
	// Add a like
	updatedPost, err := testHub.AddLikesCount(context.Background(), post.ID)
	require.NoError(t, err)
	require.Equal(t, initialLikes+1, updatedPost.LikesCount.Int32)

	// Add another like
	updatedPost, err = testHub.AddLikesCount(context.Background(), post.ID)
	require.NoError(t, err)
	require.Equal(t, initialLikes+2, updatedPost.LikesCount.Int32)
}

func TestDeletePost(t *testing.T) {
	post := createRandomPost(t)
	require.False(t, post.IsDeleted.Bool)
	
	// Delete the post (soft delete)
	err := testHub.DeletePost(context.Background(), post.ID)
	require.NoError(t, err)

	// Verify post is marked as deleted
	deletedPost, err := testHub.GetPost(context.Background(), post.ID)
	require.NoError(t, err)
	require.True(t, deletedPost.IsDeleted.Bool)
}

func TestListPosts(t *testing.T) {
	// Create multiple posts
	for i := 0; i < 5; i++ {
		createRandomPost(t)
	}

	// Fetch all posts
	allPosts, err := testHub.ListPosts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, allPosts, "Posts list should not be empty")

	// Ensure the number of fetched posts is greater than the number of created posts
	require.Greater(t, len(allPosts), 5, "The number of fetched posts should be greater than 5")
}

func TestListPostsByWall(t *testing.T) {
	// Create a wall
	wall := createRandomWall(t)

	// Create multiple posts for the wall
	postCount := 3
	for i := 0; i < postCount; i++ {
		user := createRandomUser(t)
		arg := CreatePostParams{
			WallID:   wall.ID,
			Author:   user.ID,
			MediaUrl: pgtype.Text{String: "https://example.com/media/" + util.RandomString(10) + ".jpg", Valid: true},
			PostType: NullPostType{PostType: "media", Valid: true},
		}
		_, err := testHub.CreatePost(context.Background(), arg)
		require.NoError(t, err)
	}

	// Create post for other wall
	createRandomPost(t)

	// Fetch posts for the specific wall
	wallPosts, err := testHub.ListPostsByWall(context.Background(), wall.ID)
	require.NoError(t, err)
	require.Equal(t, postCount, len(wallPosts))

	for _, post := range wallPosts {
		require.Equal(t, wall.ID, post.WallID)
	}
}

func TestGetHighlightedPosts(t *testing.T) {
	// Create multiple posts 
	posts := make([]Post, 3)
	for i := 0; i < 3; i++ {
		posts[i] = createRandomPost(t)
	}

	// Highlight two posts
	_, err := testHub.HighlightPost(context.Background(), posts[0].ID)
	require.NoError(t, err)
	_, err = testHub.HighlightPost(context.Background(), posts[2].ID)
	require.NoError(t, err)

	// Fetch highlighted posts
	highlightedPosts, err := testHub.GetHighlightedPosts(context.Background())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(highlightedPosts), 2)

	// Verify all returned posts are highlighted
	for _, post := range highlightedPosts {
		require.True(t, post.IsHighlighted.Bool)
	}
}

func TestGetHighlightedPostsByWall(t *testing.T) {
	// Create a wall
	wall := createRandomWall(t)

	// Create multiple posts for the wall
	posts := make([]Post, 3)
	for i := 0; i < 3; i++ {
		user := createRandomUser(t)
		arg := CreatePostParams{
			WallID:   wall.ID,
			Author:   user.ID,
			MediaUrl: pgtype.Text{String: "https://example.com/embed_link/" + util.RandomString(10) + ".jpg", Valid: true},
			PostType: NullPostType{PostType: "embed_link", Valid: true},
		}
		post, err := testHub.CreatePost(context.Background(), arg)
		require.NoError(t, err)
		posts[i] = post
	}

	// Create and highlight post for other wall
	otherPost := createRandomPost(t)
	_, err := testHub.HighlightPost(context.Background(), otherPost.ID)
	require.NoError(t, err)

	// Highlight two posts on our test wall
	_, err = testHub.HighlightPost(context.Background(), posts[0].ID)
	require.NoError(t, err)
	_, err = testHub.HighlightPost(context.Background(), posts[2].ID)
	require.NoError(t, err)

	// Fetch highlighted posts for the specific wall
	wallHighlightedPosts, err := testHub.GetHighlightedPostsByWall(context.Background(), wall.ID)
	require.NoError(t, err)
	require.Equal(t, 2, len(wallHighlightedPosts))

	for _, post := range wallHighlightedPosts {
		require.Equal(t, wall.ID, post.WallID)
		require.True(t, post.IsHighlighted.Bool)
	}
}