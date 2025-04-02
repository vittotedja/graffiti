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
					ID: oldUser.ID,
					Username: oldUser.Username,
					Fullname: pgtype.Text{
						String: util.RandomFullname(),
						Valid:  true,
					},
					Email: oldUser.Email,
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
					ID:    oldUser.ID,
					Username: oldUser.Username,
					Fullname: oldUser.Fullname,
					Email: util.RandomEmail(),
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
					Username: oldUser.Username,
					Fullname: oldUser.Fullname,
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
			
			// Call the update function with the generated old user
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

