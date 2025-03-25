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
	require.False(t, user.HasOnboarded.Bool) // Assuming default value is false
	require.NotZero(t, user.CreatedAt)
	require.NotZero(t, user.UpdatedAt)

	return user
}

// TestCreateUser tests the creation of a new user
func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

// TestGetUser tests retrieving a user by ID
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

// TestUpdateUserOnlyFullname tests updating only the fullname of a user
// TODO: still buggy and  apply update to all field even if i only want to edit one field
func TestUpdateUserOnlyFullname(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullname := util.RandomFullname()
	updatedUser, err := testHub.UpdateUser(context.Background(), UpdateUserParams{
		ID: oldUser.ID,
		Fullname: pgtype.Text{
			String: newFullname,
			Valid:  true,
		},
	})
	require.NoError(t, err)
	require.NotEqual(t, oldUser.Fullname.String, updatedUser.Fullname.String)
	require.Equal(t, newFullname, updatedUser.Fullname.String)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

// TestUpdateUserOnlyEmail tests updating only the email of a user
func TestUpdateUserOnlyEmail(t *testing.T) {
	oldUser := createRandomUser(t)

	newEmail := util.RandomEmail()
	updatedUser, err := testHub.UpdateUser(context.Background(), UpdateUserParams{
		ID:    oldUser.ID,
		Email: newEmail,
	})
	require.NoError(t, err)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, oldUser.Fullname.String, updatedUser.Fullname.String)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

// TestUpdateUserOnlyPassword tests updating only the password of a user
func TestUpdateUserOnlyPassword(t *testing.T) {
	oldUser := createRandomUser(t)

	newPassword := util.RandomString(6)
	newHashedPassword, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testHub.UpdateUser(context.Background(), UpdateUserParams{
		ID:             oldUser.ID,
		HashedPassword: newHashedPassword,
	})
	require.NoError(t, err)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.Fullname.String, updatedUser.Fullname.String)
	require.Equal(t, oldUser.Email, updatedUser.Email)
}

// TestUpdateUserAllFields tests updating all fields of a user
func TestUpdateUserAllFields(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullname := util.RandomFullname()
	newEmail := util.RandomEmail()
	newPassword := util.RandomString(6)
	newHashedPassword, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testHub.UpdateUser(context.Background(), UpdateUserParams{
		ID:             oldUser.ID,
		Fullname:       pgtype.Text{String: newFullname, Valid: true},
		Email:          newEmail,
		HashedPassword: newHashedPassword,
	})
	require.NoError(t, err)
	require.NotEqual(t, oldUser.Fullname.String, updatedUser.Fullname.String)
	require.Equal(t, newFullname, updatedUser.Fullname.String)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, newEmail, updatedUser.Email)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
}
