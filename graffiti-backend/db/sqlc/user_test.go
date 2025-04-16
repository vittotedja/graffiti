package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

func createRandomUser(t *testing.T) User {
	hardCodedHashedPassword := "$2a$10$EIXk5q9vz8Z3W9vZ5uJ6Ku3v7X9vZ8Z3W9vZ5uJ6Ku3v7X9vZ8Z3W"

	arg := CreateUserParams{
		Username:       util.RandomUsername(),
		Fullname:       pgtype.Text{String: util.RandomFullname(), Valid: true},
		Email:          util.RandomEmail(),
		HashedPassword: hardCodedHashedPassword,
	}

	user, err := testHub.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Fullname.String, user.Fullname.String)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.False(t, user.HasOnboarded.Bool)
	require.NotZero(t, user.CreatedAt)
	require.NotZero(t, user.UpdatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestCreateUserDuplicateUsername(t *testing.T) {
	existingUser := createRandomUser(t)

	duplicateArg := CreateUserParams{
		Username:       existingUser.Username,
		Fullname:       pgtype.Text{String: util.RandomFullname(), Valid: true},
		Email:          util.RandomEmail(),
		HashedPassword: "$2a$10$EIXk5q9vz8Z3W9vZ5uJ6Ku3v7X9vZ5uJ6Ku3v7X9vZ8Z3W",
	}

	_, err := testHub.CreateUser(context.Background(), duplicateArg)
	require.Error(t, err, "Should not allow creating user with duplicate username")
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	existingUser := createRandomUser(t)

	duplicateArg := CreateUserParams{
		Username:       util.RandomUsername(),
		Fullname:       pgtype.Text{String: util.RandomFullname(), Valid: true},
		Email:          existingUser.Email,
		HashedPassword: "$2a$10$EIXk5q9vz8Z3W9vZ5uJ6Ku3v7X9vZ5uJ6Ku3v7X9vZ8Z3W",
	}

	_, err := testHub.CreateUser(context.Background(), duplicateArg)
	require.Error(t, err, "Should not allow creating user with duplicate email")
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)

	user2, err := testHub.GetUser(context.Background(), user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.Fullname.String, user2.Fullname.String)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.HasOnboarded.Bool, user2.HasOnboarded.Bool)
	require.WithinDuration(t, user1.CreatedAt.Time, user2.CreatedAt.Time, time.Second)
	require.WithinDuration(t, user1.UpdatedAt.Time, user2.UpdatedAt.Time, time.Second)
}

func TestGetNonExistentUser(t *testing.T) {
	nonExistentID := pgtype.UUID{
		Bytes: [16]byte{},
		Valid: true,
	}

	_, err := testHub.GetUser(context.Background(), nonExistentID)
	require.Error(t, err, "Should return error for non-existent user")
}

func TestUpdateUser(t *testing.T) {
	testCases := []struct {
		name           string
		updateField    string
		updateFunction func(oldUser User) UpdateUserParams
		verifyUpdate   func(t *testing.T, oldUser, updatedUser User)
	}{
		{
			name:        "Update Fullname",
			updateField: "fullname",
			updateFunction: func(oldUser User) UpdateUserParams {
				return UpdateUserParams{
					ID:       oldUser.ID,
					Username: oldUser.Username,
					Fullname: pgtype.Text{
						String: util.RandomFullname(),
						Valid:  true,
					},
					Email:          oldUser.Email,
					HashedPassword: oldUser.HashedPassword,
				}
			},
			verifyUpdate: func(t *testing.T, oldUser, updatedUser User) {
				require.NotEqual(t, oldUser.Fullname.String, updatedUser.Fullname.String)
				require.Equal(t, oldUser.Username, updatedUser.Username)
				require.Equal(t, oldUser.Email, updatedUser.Email)
				require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
			},
		},
		{
			name:        "Update Username",
			updateField: "username",
			updateFunction: func(oldUser User) UpdateUserParams {
				return UpdateUserParams{
					ID: oldUser.ID,
					Username: util.RandomUsername(),
					Fullname: oldUser.Fullname,
					Email: oldUser.Email,
					HashedPassword: oldUser.HashedPassword,
				}
			},
			verifyUpdate: func(t *testing.T, oldUser, updatedUser User) {
				require.NotEqual(t, oldUser.Username, updatedUser.Username)
				require.Equal(t, oldUser.Fullname.String, updatedUser.Fullname.String)
				require.Equal(t, oldUser.Email, updatedUser.Email)
				require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
			},
		},
		{
			name:        "Update Email",
			updateField: "email",
			updateFunction: func(oldUser User) UpdateUserParams {
				return UpdateUserParams{
					ID:             oldUser.ID,
					Username:       oldUser.Username,
					Fullname:       oldUser.Fullname,
					Email:          util.RandomEmail(),
					HashedPassword: oldUser.HashedPassword,
				}
			},
			verifyUpdate: func(t *testing.T, oldUser, updatedUser User) {
				require.Equal(t, oldUser.Username, updatedUser.Username)
				require.NotEqual(t, oldUser.Email, updatedUser.Email)
				require.Equal(t, oldUser.Fullname.String, updatedUser.Fullname.String)
				require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
			},
		},
		{
			name:        "Update Password",
			updateField: "password",
			updateFunction: func(oldUser User) UpdateUserParams {
				newPassword := util.RandomString(6)
				hashedPassword, err := util.HashPassword(newPassword)
				require.NoError(t, err)
				return UpdateUserParams{
					ID:             oldUser.ID,
					Username:       oldUser.Username,
					Fullname:       oldUser.Fullname,
					Email:          oldUser.Email,
					HashedPassword: hashedPassword,
				}
			},
			verifyUpdate: func(t *testing.T, oldUser, updatedUser User) {
				require.Equal(t, oldUser.Username, updatedUser.Username)
				require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
				require.Equal(t, oldUser.Fullname.String, updatedUser.Fullname.String)
				require.Equal(t, oldUser.Email, updatedUser.Email)
			},
		},
		{
			name:        "Update All Fields",
			updateField: "all fields",
			updateFunction: func(oldUser User) UpdateUserParams {
				return UpdateUserParams{
					ID: oldUser.ID,
					Username: util.RandomUsername(),
					Fullname: pgtype.Text{
						String: util.RandomFullname(),
						Valid:  true,
					},
					Email:          util.RandomEmail(),
					HashedPassword: oldUser.HashedPassword, // You can keep the password hash the same or update it here
				}
			},
			verifyUpdate: func(t *testing.T, oldUser, updatedUser User) {
				require.NotEqual(t, oldUser.Username, updatedUser.Username)
				require.NotEqual(t, oldUser.Fullname.String, updatedUser.Fullname.String)
				require.NotEqual(t, oldUser.Email, updatedUser.Email)
				require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword) // Check if the password is not updated (if not changed)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a random user for testing
			oldUser := createRandomUser(t)
			updateParams := tc.updateFunction(oldUser)
			updatedUser, err := testHub.UpdateUser(context.Background(), updateParams)

			require.NoError(t, err, "Should update %s successfully", tc.updateField)
			tc.verifyUpdate(t, oldUser, updatedUser)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	user := createRandomUser(t)

	// Delete the user
	err := testHub.DeleteUser(context.Background(), user.ID)
	require.NoError(t, err, "Should delete user successfully")

	// Verify user is deleted by trying to fetch
	_, err = testHub.GetUser(context.Background(), user.ID)
	require.Error(t, err, "Should not be able to fetch deleted user")
}

func TestDeleteNonExistentUser(t *testing.T) {
	// Create multiple users
	users := make([]User, 5)
	for i := 0; i < 5; i++ {
		users[i] = createRandomUser(t)
	}

	// Get number of users
	initialUsers, err := testHub.ListUsers(context.Background())
	require.NoError(t, err, "Should fetch users successfully before deletion")

	// Delete Non Existent User
	nonExistentID := pgtype.UUID{
		Bytes: [16]byte{},
		Valid: true,
	}

	err = testHub.DeleteUser(context.Background(), nonExistentID)
	require.NoError(t, err, "No user should be deleted and no error should be thrown")

	// Get number of users and ensure no user was deleted
	finalUsers, err := testHub.ListUsers(context.Background())
	require.NoError(t, err, "Should fetch users successfully after deletion attempt")
	require.Equal(t, len(initialUsers), len(finalUsers), "The number of users should remain the same after deleting a non-existent user")
}

func TestListUsers(t *testing.T) {
	// Create multiple users
	users := make([]User, 5)
	for i := 0; i < 5; i++ {
		users[i] = createRandomUser(t)
	}

	// Fetch all users
	allUsers, err := testHub.ListUsers(context.Background())
	require.NoError(t, err, "Should list users successfully")
	require.GreaterOrEqual(t, len(allUsers), 5, "Should have at least the created users")
}

func TestFinishOnboarding(t *testing.T) {
	// Create a random user
	user := createRandomUser(t)

	// Call FinishOnboarding to set has_onboarded to true and update onboarding_at
	err := testHub.FinishOnboarding(context.Background(), user.ID)
	require.NoError(t, err, "Should finish onboarding without error")

	// Fetch the user again from the database
	updatedUser, err := testHub.GetUser(context.Background(), user.ID)
	require.NoError(t, err, "Should fetch the user successfully after onboarding")

	// Check that has_onboarded is true
	require.True(t, updatedUser.HasOnboarded.Bool, "User should have completed onboarding")

	// Check that onboarding_at is set (not zero)
	require.NotZero(t, updatedUser.OnboardingAt.Time, "Onboarding timestamp should be set")
}

func TestUpdateProfile(t *testing.T) {
	testCases := []struct {
		name           string
		updateFields   []string
		updateFunction func(oldUser User) UpdateProfileParams
		verifyUpdate   func(t *testing.T, oldUser, updatedUser User)
	}{
		// Update one field at a time
		{
			name:         "Update Profile Picture",
			updateFields: []string{"profile_picture"},
			updateFunction: func(oldUser User) UpdateProfileParams {
				return UpdateProfileParams{
					ID:             oldUser.ID,
					ProfilePicture: pgtype.Text{String: util.RandomString(10), Valid: true},
					Bio:            oldUser.Bio,
					BackgroundImage: oldUser.BackgroundImage,
				}
			},
			verifyUpdate: func(t *testing.T, oldUser, updatedUser User) {
				require.NotEqual(t, oldUser.ProfilePicture.String, updatedUser.ProfilePicture.String)
				require.Equal(t, oldUser.Bio.String, updatedUser.Bio.String)
				require.Equal(t, oldUser.BackgroundImage.String, updatedUser.BackgroundImage.String)
			},
		},
		{
			name:         "Update Bio",
			updateFields: []string{"bio"},
			updateFunction: func(oldUser User) UpdateProfileParams {
				return UpdateProfileParams{
					ID:             oldUser.ID,
					ProfilePicture: oldUser.ProfilePicture,
					Bio:            pgtype.Text{String: util.RandomString(10), Valid: true},
					BackgroundImage: oldUser.BackgroundImage,
				}
			},
			verifyUpdate: func(t *testing.T, oldUser, updatedUser User) {
				require.Equal(t, oldUser.ProfilePicture.String, updatedUser.ProfilePicture.String)
				require.NotEqual(t, oldUser.Bio.String, updatedUser.Bio.String)
				require.Equal(t, oldUser.BackgroundImage.String, updatedUser.BackgroundImage.String)
			},
		},
		{
			name:         "Update Background Image",
			updateFields: []string{"background_image"},
			updateFunction: func(oldUser User) UpdateProfileParams {
				return UpdateProfileParams{
					ID:             oldUser.ID,
					ProfilePicture: oldUser.ProfilePicture,
					Bio:            oldUser.Bio,
					BackgroundImage: pgtype.Text{String: util.RandomString(10), Valid: true},
				}
			},
			verifyUpdate: func(t *testing.T, oldUser, updatedUser User) {
				require.Equal(t, oldUser.ProfilePicture.String, updatedUser.ProfilePicture.String)
				require.Equal(t, oldUser.Bio.String, updatedUser.Bio.String)
				require.NotEqual(t, oldUser.BackgroundImage.String, updatedUser.BackgroundImage.String)
			},
		},

		// Update all fields at once
		{
			name:         "Update All Profile Fields",
			updateFields: []string{"profile_picture", "bio", "background_image"},
			updateFunction: func(oldUser User) UpdateProfileParams {
				return UpdateProfileParams{
					ID:             oldUser.ID,
					ProfilePicture: pgtype.Text{String: util.RandomString(10), Valid: true},
					Bio:            pgtype.Text{String: util.RandomString(10), Valid: true},
					BackgroundImage: pgtype.Text{String: util.RandomString(10), Valid: true},
				}
			},
			verifyUpdate: func(t *testing.T, oldUser, updatedUser User) {
				require.NotEqual(t, oldUser.ProfilePicture.String, updatedUser.ProfilePicture.String)
				require.NotEqual(t, oldUser.Bio.String, updatedUser.Bio.String)
				require.NotEqual(t, oldUser.BackgroundImage.String, updatedUser.BackgroundImage.String)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a random user for testing
			oldUser := createRandomUser(t)

			// Call the update function with the generated old user
			updateParams := tc.updateFunction(oldUser)
			updatedUser, err := testHub.UpdateProfile(context.Background(), updateParams)

			// Ensure no error occurred
			require.NoError(t, err, "Failed to update profile")

			// Verify the updated user based on the test case's rules
			tc.verifyUpdate(t, oldUser, updatedUser)
		})
	}
}

func TestSearchUsersILike(t *testing.T) {
	ctx := context.Background()

	// Arrange - create a user that should match with ILIKE
	targetUser := createRandomUser(t)
	uniqueUsername := "vid_" + util.RandomString(6)
	targetUser.Username = uniqueUsername
	targetUser.Fullname = pgtype.Text{String: "Vid Tonic " + util.RandomString(4), Valid: true}

	_, err := testHub.UpdateUser(ctx, UpdateUserParams{
		ID:             targetUser.ID,
		Username:       targetUser.Username,
		Fullname:       targetUser.Fullname,
		Email:          targetUser.Email,
		HashedPassword: targetUser.HashedPassword,
	})
	require.NoError(t, err)

	// Create a distractor user
	distractor := createRandomUser(t)
	distractor.Username = "random"
	distractor.Fullname = pgtype.Text{String: "No Match", Valid: true}

	_, err = testHub.UpdateUser(ctx, UpdateUserParams{
		ID:             distractor.ID,
		Username:       distractor.Username,
		Fullname:       distractor.Fullname,
		Email:          distractor.Email,
		HashedPassword: distractor.HashedPassword,
	})
	require.NoError(t, err)

	// Act - ILIKE match (short term, <3)
	searchTerm := pgtype.Text{String: targetUser.Username, Valid: true}
	results, err := testHub.SearchUsersILike(ctx, searchTerm)
	require.NoError(t, err)
	require.NotEmpty(t, results, "Expected matches with ILIKE search")

	// Assert
	var found bool
	for _, u := range results {
		if u.Username == targetUser.Username || (u.Fullname.Valid && u.Fullname.String == targetUser.Fullname.String) {
			found = true
			break
		}
	}
	require.True(t, found, "Target user not found in ILIKE search")
}

func TestSearchUsersTrigram(t *testing.T) {
	ctx := context.Background()

	// Arrange - create a user that should match trigram search
	targetUser := createRandomUser(t)
	uniqueUsername := "vittotedja_" + util.RandomString(6)
	targetUser.Username = uniqueUsername
	targetUser.Fullname = pgtype.Text{String: "Vitto Tedja " + util.RandomString(4), Valid: true}

	_, err := testHub.UpdateUser(ctx, UpdateUserParams{
		ID:             targetUser.ID,
		Username:       targetUser.Username,
		Fullname:       targetUser.Fullname,
		Email:          targetUser.Email,
		HashedPassword: targetUser.HashedPassword,
	})
	require.NoError(t, err)

	// Create a distractor user
	distractor := createRandomUser(t)
	distractor.Username = "irrelevant"
	distractor.Fullname = pgtype.Text{String: "Totally Off", Valid: true}

	_, err = testHub.UpdateUser(ctx, UpdateUserParams{
		ID:             distractor.ID,
		Username:       distractor.Username,
		Fullname:       distractor.Fullname,
		Email:          distractor.Email,
		HashedPassword: distractor.HashedPassword,
	})
	require.NoError(t, err)

	// Act - trigram search
	results, err := testHub.SearchUsersTrigram(ctx, targetUser.Username[:5])
	require.NoError(t, err)
	require.NotEmpty(t, results, "Expected results from trigram search")

	// Assert
	var found bool
	for _, u := range results {
		if u.Username == targetUser.Username || (u.Fullname.Valid && u.Fullname.String == targetUser.Fullname.String) {
			found = true
			break
		}
	}
	require.True(t, found, "Target user not found in trigram search results")
}

func TestGetNumberOfMutualFriends(t *testing.T) {
	ctx := context.Background()

	// Create 3 users
	userA := createRandomUser(t)
	userB := createRandomUser(t)
	mutual := createRandomUser(t)

	// Create friendships to form a mutual connection
	_, err := testHub.CreateFriendship(ctx, CreateFriendshipParams{
		FromUser: userA.ID,
		ToUser:   mutual.ID,
		Status: NullStatus{
			Status: "friends",
			Valid:  true,
		},
	})
	require.NoError(t, err)

	_, err = testHub.CreateFriendship(ctx, CreateFriendshipParams{
		FromUser: userB.ID,
		ToUser:   mutual.ID,
		Status: NullStatus{
			Status: "friends",
			Valid:  true,
		},
	})
	require.NoError(t, err)

	// Refresh materialized view
	err = testHub.RefreshMaterializedViews(ctx)
	require.NoError(t, err)

	// Call query
	count, err := testHub.GetNumberOfMutualFriends(ctx, GetNumberOfMutualFriendsParams{
		UserID:   userA.ID,
		UserID_2: userB.ID,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func TestDiscoverFriendsByMutuals(t *testing.T) {
	ctx := context.Background()

	// User A is the main user
	userA := createRandomUser(t)
	mutual := createRandomUser(t)
	suggested := createRandomUser(t)

	// A <-> Mutual
	_, err := testHub.CreateFriendship(ctx, CreateFriendshipParams{
		FromUser: userA.ID,
		ToUser:   mutual.ID,
		Status: NullStatus{
			Status: "friends",
			Valid:  true,
		},
	})
	require.NoError(t, err)

	// Suggested <-> Mutual
	_, err = testHub.CreateFriendship(ctx, CreateFriendshipParams{
		FromUser: suggested.ID,
		ToUser:   mutual.ID,
		Status: NullStatus{
			Status: "friends",
			Valid:  true,
		},
	})
	require.NoError(t, err)

	// Refresh MV
	err = testHub.RefreshMaterializedViews(ctx)
	require.NoError(t, err)

	// Discover friends
	discoveries, err := testHub.DiscoverFriendsByMutuals(ctx, userA.ID)
	require.NoError(t, err)

	// Check if suggested user appears
	var found bool
	for _, u := range discoveries {
		if u.ID == suggested.ID {
			found = true
			require.Equal(t, int64(1), u.MutualFriendCount)
			break
		}
	}
	require.True(t, found, "Suggested user not found in discover results")
}

func TestGetMutualFriends(t *testing.T) {
	ctx := context.Background()

	// Create 3 users
	userA := createRandomUser(t)
	userB := createRandomUser(t)
	mutual := createRandomUser(t)

	// Create friendships to form mutual connections
	_, err := testHub.CreateFriendship(ctx, CreateFriendshipParams{
		FromUser: userA.ID,
		ToUser:   mutual.ID,
		Status: NullStatus{
			Status: "friends",
			Valid:  true,
		},
	})
	require.NoError(t, err)

	_, err = testHub.CreateFriendship(ctx, CreateFriendshipParams{
		FromUser: userB.ID,
		ToUser:   mutual.ID,
		Status: NullStatus{
			Status: "friends",
			Valid:  true,
		},
	})
	require.NoError(t, err)

	// Refresh materialized view
	err = testHub.RefreshMaterializedViews(ctx)
	require.NoError(t, err)

	// Call query to get mutual friends
	results, err := testHub.ListMutualFriends(ctx, ListMutualFriendsParams{
		UserID:   userA.ID,
		UserID_2: userB.ID,
	})
	require.NoError(t, err)

	// Expect exactly 1 mutual friend with matching ID
	require.Len(t, results, 1)
	require.Equal(t, mutual.ID, results[0].ID)
}
