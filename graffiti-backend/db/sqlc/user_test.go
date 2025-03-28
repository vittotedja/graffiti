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

func TestUpdateUserPartialFields(t *testing.T) {
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
				require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
				require.Equal(t, oldUser.Fullname.String, updatedUser.Fullname.String)
				require.Equal(t, oldUser.Email, updatedUser.Email)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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