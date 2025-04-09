package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

func createRandomWall(t *testing.T) Wall {
	user := createRandomUser(t)

	arg := CreateWallParams{
		UserID:          user.ID,
		Title:           "Wall Title" + util.RandomString(10),
		Description:     pgtype.Text{String: util.RandomString(20), Valid: true},
		BackgroundImage: pgtype.Text{String: "https://example.com/" + util.RandomString(10) + ".jpg", Valid: true},
	}

	wall, err := testHub.CreateWall(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, wall)

	require.Equal(t, arg.UserID, wall.UserID)
	require.Equal(t, arg.Title, wall.Title)
	require.Equal(t, arg.Description.String, wall.Description.String)
	require.Equal(t, arg.BackgroundImage.String, wall.BackgroundImage.String)
	require.False(t, wall.IsPublic.Bool)
	require.False(t, wall.IsArchived.Bool)
	require.False(t, wall.IsDeleted.Bool)
	require.Equal(t, pgtype.Float8(pgtype.Float8{Float64:0, Valid:true}), wall.PopularityScore)
	require.NotZero(t, wall.CreatedAt)
	require.NotZero(t, wall.UpdatedAt)

	return wall
}

func TestCreateWall(t *testing.T) {
	createRandomWall(t)
}

func TestGetWall(t *testing.T) {
	wall1 := createRandomWall(t)

	wall2, err := testHub.GetWall(context.Background(), wall1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, wall2)

	require.Equal(t, wall1.ID, wall2.ID)
	require.Equal(t, wall1.UserID, wall2.UserID)
	require.Equal(t, wall1.Description.String, wall2.Description.String)
	require.Equal(t, wall1.BackgroundImage.String, wall2.BackgroundImage.String)
	require.Equal(t, wall1.IsPublic.Bool, wall2.IsPublic.Bool)
	require.Equal(t, wall1.IsArchived.Bool, wall2.IsArchived.Bool)
	require.Equal(t, wall1.IsDeleted.Bool, wall2.IsDeleted.Bool)
	require.Equal(t, wall1.PopularityScore, wall2.PopularityScore)
	require.WithinDuration(t, wall1.CreatedAt.Time, wall2.CreatedAt.Time, time.Second)
	require.WithinDuration(t, wall1.UpdatedAt.Time, wall2.UpdatedAt.Time, time.Second)
}

func TestGetNonExistentWall(t *testing.T) {
	nonExistentID := pgtype.UUID{
		Bytes: [16]byte{},
		Valid: true,
	}

	_, err := testHub.GetWall(context.Background(), nonExistentID)
	require.Error(t, err, "Should return error for non-existent wall")
}

func TestUpdateWall(t *testing.T) {
	wall := createRandomWall(t)

	testCases := []struct {
		name           string
		updateField    string
		updateFunction func(oldWall Wall) UpdateWallParams
		verifyUpdate   func(t *testing.T, oldWall, updatedWall Wall)
	}{
		{
			name:        "Update Description",
			updateField: "description",
			updateFunction: func(oldWall Wall) UpdateWallParams {
				return UpdateWallParams{
					ID: oldWall.ID,
					Description: pgtype.Text{
						String: util.RandomString(20),
						Valid:  true,
					},
					BackgroundImage: pgtype.Text{Valid: false},
				}
			},
			verifyUpdate: func(t *testing.T, oldWall, updatedWall Wall) {
				require.NotEqual(t, oldWall.Description.String, updatedWall.Description.String)
				require.Equal(t, oldWall.BackgroundImage.String, updatedWall.BackgroundImage.String)
			},
		},
		{
			name:        "Update Background Image",
			updateField: "background_image",
			updateFunction: func(oldWall Wall) UpdateWallParams {
				return UpdateWallParams{
					ID:              oldWall.ID,
					Description:     pgtype.Text{Valid: false},
					BackgroundImage: pgtype.Text{String: "https://example.com/" + util.RandomString(10) + ".jpg", Valid: true},
				}
			},
			verifyUpdate: func(t *testing.T, oldWall, updatedWall Wall) {
				require.Equal(t, oldWall.Description.String, updatedWall.Description.String)
				require.NotEqual(t, oldWall.BackgroundImage.String, updatedWall.BackgroundImage.String)
			},
		},
		{
			name:        "Update Both Fields",
			updateField: "both_fields",
			updateFunction: func(oldWall Wall) UpdateWallParams {
				return UpdateWallParams{
					ID:              oldWall.ID,
					Description:     pgtype.Text{String: util.RandomString(20), Valid: true},
					BackgroundImage: pgtype.Text{String: "https://example.com/" + util.RandomString(10) + ".jpg", Valid: true},
				}
			},
			verifyUpdate: func(t *testing.T, oldWall, updatedWall Wall) {
				require.NotEqual(t, oldWall.Description.String, updatedWall.Description.String)
				require.NotEqual(t, oldWall.BackgroundImage.String, updatedWall.BackgroundImage.String)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			updateParams := tc.updateFunction(wall)
			updatedWall, err := testHub.UpdateWall(context.Background(), updateParams)
			
			require.NoError(t, err, "Should update %s successfully", tc.updateField)
			tc.verifyUpdate(t, wall, updatedWall)

			// Update our reference wall for next test
			wall = updatedWall
		})
	}
}

func TestPublicizeAndPrivatizeWall(t *testing.T) {
	wall := createRandomWall(t)
	require.False(t, wall.IsPublic.Bool)

	// Test publicizing the wall
	publicWall, err := testHub.PublicizeWall(context.Background(), wall.ID)
	require.NoError(t, err)
	require.True(t, publicWall.IsPublic.Bool)

	// Test privatizing the wall
	privateWall, err := testHub.PrivatizeWall(context.Background(), wall.ID)
	require.NoError(t, err)
	require.False(t, privateWall.IsPublic.Bool)
}

func TestArchiveAndUnarchiveWall(t *testing.T) {
	wall := createRandomWall(t)
	require.False(t, wall.IsArchived.Bool)

	// Test archiving the wall
	err := testHub.ArchiveWall(context.Background(), wall.ID)
	require.NoError(t, err)

	// Verify wall is archived
	archivedWall, err := testHub.GetWall(context.Background(), wall.ID)
	require.NoError(t, err)
	require.True(t, archivedWall.IsArchived.Bool)

	// Test unarchiving the wall
	err = testHub.UnarchiveWall(context.Background(), wall.ID)
	require.NoError(t, err)

	// Verify wall is unarchived
	unarchivedWall, err := testHub.GetWall(context.Background(), wall.ID)
	require.NoError(t, err)
	require.False(t, unarchivedWall.IsArchived.Bool)
}

func TestDeleteWall(t *testing.T) {
	wall := createRandomWall(t)
	require.False(t, wall.IsDeleted.Bool)

	// Delete the wall (soft delete)
	err := testHub.DeleteWall(context.Background(), wall.ID)
	require.NoError(t, err)

	// Verify wall is marked as deleted
	deletedWall, err := testHub.GetWall(context.Background(), wall.ID)
	require.NoError(t, err)
	require.True(t, deletedWall.IsDeleted.Bool)
}

func TestListWalls(t *testing.T) {
	// Create multiple walls
	for i := 0; i < 5; i++ {
		createRandomWall(t)
	}

	// Fetch all walls
	allWalls, err := testHub.ListWalls(context.Background())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(allWalls), 5)
}

func TestListWallsByUser(t *testing.T) {
	// Create a user
	user := createRandomUser(t)

	// Create multiple walls for the user
	wallCount := 3
	for i := 0; i < wallCount; i++ {
		arg := CreateWallParams{
			UserID:          user.ID,
			Description:     pgtype.Text{String: util.RandomString(20), Valid: true},
			BackgroundImage: pgtype.Text{String: "https://example.com/" + util.RandomString(10) + ".jpg", Valid: true},
		}
		_, err := testHub.CreateWall(context.Background(), arg)
		require.NoError(t, err)
	}

	// Create wall for other user
	createRandomWall(t)

	// Fetch walls for the specific user
	userWalls, err := testHub.ListWallsByUser(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, wallCount, len(userWalls))

	for _, wall := range userWalls {
		require.Equal(t, user.ID, wall.UserID)
	}
}

func TestPinUnpinWall(t *testing.T) {
    wall := createRandomWall(t)
    require.False(t, wall.IsPinned)

    // Test pinning the wall
    pinnedWall, err := testHub.PinUnpinWall(context.Background(), wall.ID)
    require.NoError(t, err)
    require.True(t, pinnedWall.IsPinned)

    // Test unpinning the wall
    unpinnedWall, err := testHub.PinUnpinWall(context.Background(), wall.ID)
    require.NoError(t, err)
    require.False(t, unpinnedWall.IsPinned)
}
