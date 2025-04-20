package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	mockdb "github.com/vittotedja/graffiti/graffiti-backend/db/mock"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

// TestGetUserAPI tests the getUser handler
func TestGetUserAPIFixed(t *testing.T) {
	user, _ := randomUser(t)

	testCases := []struct {
		name          string
		userID        string
		setupAuth     func(t *testing.T, request *http.Request, currentUser db.User)
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			userID: user.ID.String(),
			setupAuth: func(t *testing.T, request *http.Request, currentUser db.User) {
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				var id pgtype.UUID
				id.Scan(user.ID.String())

				mockHub.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUserResponse(t, recorder.Body, user)
			},
		},
		{
			name:   "UserNotFound",
			userID: user.ID.String(),
			setupAuth: func(t *testing.T, request *http.Request, currentUser db.User) {
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, db.ErrRecordNotFound)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			userID: "invalid-id",
			setupAuth: func(t *testing.T, request *http.Request, currentUser db.User) {
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			userID: user.ID.String(),
			setupAuth: func(t *testing.T, request *http.Request, currentUser db.User) {
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t)
			mockHub, ok := server.hub.(*mockdb.MockHub)
			require.True(t, ok)

			tc.setupMock(mockHub)
			recorder := httptest.NewRecorder()

			server.router.GET("/test/users/:id", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.getUser(ctx)
			})

			url := "/test/users/" + tc.userID
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, user)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListUsersAPIFixed tests the listUsers handler
func TestListUsersAPIFixed(t *testing.T) {
	user1, _ := randomUser(t)
	user2, _ := randomUser(t)

	testCases := []struct {
		name          string
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListUsers(gomock.Any()).
					Times(1).
					Return([]db.User{user1, user2}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				
				var gotResponse []getUserResponse
				err = json.Unmarshal(data, &gotResponse)
				require.NoError(t, err)
				require.Len(t, gotResponse, 2)
				
				require.Equal(t, user1.ID.String(), gotResponse[0].ID)
				require.Equal(t, user2.ID.String(), gotResponse[1].ID)
			},
		},
		{
			name: "InternalError",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListUsers(gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t)
			mockHub, ok := server.hub.(*mockdb.MockHub)
			require.True(t, ok)

			tc.setupMock(mockHub)
			recorder := httptest.NewRecorder()

			url := "/test/users"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.GET("/test/users", server.listUsers)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestUpdateUserNewAPI tests the updateUserNew handler
func TestUpdateUserNewAPIFixed(t *testing.T) {
	currentUser, _ := randomUser(t)
	
	updatedUser := currentUser
	newUsername := "new_" + currentUser.Username

	testCases := []struct {
		name          string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": newUsername,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					UpdateUserNew(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.UpdateUserNewParams) (db.User, error) {
						require.Equal(t, currentUser.ID, params.ID)
						require.Equal(t, newUsername, params.Username)
						
						updatedUser.Username = newUsername
						return updatedUser, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				
				var gotResponse getUserResponse
				err = json.Unmarshal(data, &gotResponse)
				require.NoError(t, err)
				
				require.Equal(t, newUsername, gotResponse.Username)
			},
		},
		{
			name: "InvalidRequest",
			body: gin.H{
				"username": []string{"invalid"},
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					UpdateUserNew(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username": newUsername,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					UpdateUserNew(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t)
			mockHub, ok := server.hub.(*mockdb.MockHub)
			require.True(t, ok)

			server.router.POST("/test/updateuser", func(ctx *gin.Context) {
				ctx.Set("currentUser", currentUser)
				server.updateUserNew(ctx)
			})

			tc.setupMock(mockHub)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/test/updateuser"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestFinishOnboardingAPIFixed tests the finishOnboarding handler
func TestFinishOnboardingAPIFixed(t *testing.T) {
	user, _ := randomUser(t)

	testCases := []struct {
		name          string
		userID        string
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			userID: user.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				var id pgtype.UUID
				id.Scan(user.ID.String())
				
				mockHub.EXPECT().
					FinishOnboarding(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			userID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					FinishOnboarding(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			userID: user.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					FinishOnboarding(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t)
			mockHub, ok := server.hub.(*mockdb.MockHub)
			require.True(t, ok)

			tc.setupMock(mockHub)
			recorder := httptest.NewRecorder()

			server.router.PUT("/test/users/:id/onboarding", server.finishOnboarding)

			url := "/test/users/" + tc.userID + "/onboarding"
			request, err := http.NewRequest(http.MethodPut, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestDeleteUserAPIFixed tests the deleteUser handler
func TestDeleteUserAPIFixed(t *testing.T) {
	user, _ := randomUser(t)

	testCases := []struct {
		name          string
		userID        string
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			userID: user.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				var id pgtype.UUID
				id.Scan(user.ID.String())
				
				mockHub.EXPECT().
					DeleteUser(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				
				var response map[string]string
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				
				err = json.Unmarshal(data, &response)
				require.NoError(t, err)
				
				require.Equal(t, user.ID.String(), response["id"])
				require.Equal(t, "User deleted successfully!", response["message"])
			},
		},
		{
			name:   "InvalidID",
			userID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			userID: user.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t)
			mockHub, ok := server.hub.(*mockdb.MockHub)
			require.True(t, ok)

			tc.setupMock(mockHub)
			recorder := httptest.NewRecorder()

			server.router.DELETE("/test/users/:id", server.deleteUser)

			url := "/test/users/" + tc.userID
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}


// Helper function to match user response
func requireBodyMatchUserResponse(t *testing.T, body *bytes.Buffer, user db.User) {
    data, err := io.ReadAll(body)
    require.NoError(t, err)

    var gotResponse getUserResponse
    err = json.Unmarshal(data, &gotResponse)
    require.NoError(t, err)

    require.Equal(t, user.ID.String(), gotResponse.ID)
    require.Equal(t, user.Username, gotResponse.Username)
    require.Equal(t, user.Fullname.String, gotResponse.Fullname)
    require.Equal(t, user.Email, gotResponse.Email)
    require.Equal(t, user.HasOnboarded.Bool, gotResponse.HasOnboarded)
    
    if user.ProfilePicture.Valid {
        require.Equal(t, user.ProfilePicture.String, gotResponse.ProfilePicture)
    }
    
    if user.Bio.Valid {
        require.Equal(t, user.Bio.String, gotResponse.Bio)
    }
    
    if user.BackgroundImage.Valid {
        require.Equal(t, user.BackgroundImage.String, gotResponse.BackgroundImage)
    }
    
    if !user.CreatedAt.Time.IsZero() {
        createdAt, err := time.Parse(time.RFC3339, gotResponse.CreatedAt)
        require.NoError(t, err)
        require.WithinDuration(t, user.CreatedAt.Time, createdAt, time.Second)
    }
    
    if user.UpdatedAt.Valid && !user.UpdatedAt.Time.IsZero() {
        updatedAt, err := time.Parse(time.RFC3339, gotResponse.UpdatedAt)
        require.NoError(t, err)
        require.WithinDuration(t, user.UpdatedAt.Time, updatedAt, time.Second)
    }
}

func randomUser(t *testing.T) (db.User, string) {
    password := util.RandomString(6)
    hashedPassword, _ := util.HashPassword(password)
    
	rawUUID := uuid.New()
	uuidString := rawUUID.String()

    id := pgtype.UUID{}
	err := id.Scan(uuidString)
	require.NoError(t, err)
    
    fullname := pgtype.Text{}
    fullname.Scan(util.RandomFullname())
    
    profilePicture := pgtype.Text{}
    profilePicture.Scan(util.RandomProfilePictureURL())
    
    bio := pgtype.Text{}
    bio.Scan(util.RandomBio())
    
    hasOnboarded := pgtype.Bool{}
    hasOnboarded.Scan(false)
    
    backgroundImage := pgtype.Text{}
    backgroundImage.Scan(util.RandomProfilePictureURL())
    
    createdAt := pgtype.Timestamp{}
    createdAt.Scan(time.Now())
    
    updateAt := pgtype.Timestamp{}
    updateAt.Scan(time.Now())
    
    user := db.User{
        ID:              id,
        Username:        util.RandomUsername(),
        Fullname:        fullname,
        Email:           util.RandomEmail(),
        HashedPassword:  hashedPassword,
        ProfilePicture:  profilePicture,
        Bio:             bio,
        HasOnboarded:    hasOnboarded,
        BackgroundImage: backgroundImage,
        OnboardingAt:    pgtype.Timestamp{},
        CreatedAt:       createdAt,
        UpdatedAt:       updateAt,
    }

    return user, password
}


