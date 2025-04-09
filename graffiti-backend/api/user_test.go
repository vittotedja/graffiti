package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
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

type eqCreateUserParamsMatcher struct {
    arg      db.CreateUserParams
    password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
    arg, ok := x.(db.CreateUserParams)
    if !ok {
        return false
    }
    if arg.HashedPassword != e.password {
        return false
    }

    return arg.Username == e.arg.Username &&
           arg.Fullname.String == e.arg.Fullname.String &&
           arg.Email == e.arg.Email
}

func (e eqCreateUserParamsMatcher) String() string {
    return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
    return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
                "username": user.Username,
                "password": password,
                "fullname": user.Fullname.String,
                "email":    user.Email,
            },
			setupMock: func(mockHub *mockdb.MockHub) {
				arg := db.CreateUserParams{
					Username: user.Username,
					Fullname: pgtype.Text{String: user.Fullname.String, Valid: true},
					Email:    user.Email,
				}
				
				// Use gomock.Any() for the hashed password since it might be processed
				mockHub.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.CreateUserParams) (db.User, error) {
						// Verify the parameters match what we expect
						require.Equal(t, arg.Username, params.Username)
						require.Equal(t, arg.Fullname.String, params.Fullname.String)
						require.Equal(t, arg.Email, params.Email)
						require.Equal(t, password, params.HashedPassword) // Since hashPassword just returns the password in tests
						
						return user, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUserResponse(t, recorder.Body, user)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"fullname": user.Fullname.String,
				"email":     user.Email,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"fullname": user.Fullname.String,
				"email":     user.Email,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, db.ErrUniqueViolation)
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
			require.True(t, ok, "server.hub is not a *mockdb.MockHub")
			
			tc.setupMock(mockHub)
			
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetUserAPI tests the getUser handler
func TestGetUserAPI(t *testing.T) {
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

			url := fmt.Sprintf("/api/v1/users/%s", tc.userID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListUsersAPI tests the listUsers handler
func TestListUsersAPI(t *testing.T) {
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

			url := "/api/v1/users"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestUpdateUserAPI tests the updateUser handler
func TestUpdateUserAPI(t *testing.T) {
	user, _ := randomUser(t)
	
	// Create updated user
	updatedUser := user
	newUsername := util.RandomUsername()
	updatedUser.Username = newUsername

	testCases := []struct {
		name          string
		userID        string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			userID: user.ID.String(),
			body: gin.H{
				"username": newUsername,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				var id pgtype.UUID
				id.Scan(user.ID.String())
				
				// First expect GetUser to be called
				mockHub.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, uuid pgtype.UUID) (db.User, error) {
						require.Equal(t, user.ID.String(), uuid.String())
						return user, nil
					})
					
				// Then expect UpdateUser to be called with the updated params
				mockHub.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.UpdateUserParams) (db.User, error) {
						require.Equal(t, id.String(), params.ID.String())
						require.Equal(t, newUsername, params.Username)
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
			name:   "UserNotFound",
			userID: uuid.New().String(), // Use a random valid UUID that won't be found
			body: gin.H{
				"username": newUsername,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Return(db.User{}, db.ErrRecordNotFound)
					
				// UpdateUser should not be called
				mockHub.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			userID: "invalid-id", // Not a valid UUID format
			body: gin.H{
				"username": newUsername,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				// No database calls should be made
				mockHub.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			userID: user.ID.String(),
			body: gin.H{
				"username": newUsername,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Return(user, nil)
					
				mockHub.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
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

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			// Make sure the URL matches exactly how it's registered in your router
			t.Logf("UserID type: %T, value: %v", tc.userID, tc.userID)
			url := fmt.Sprintf("/api/v1/users/%s", tc.userID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}


// TestUpdateProfileAPI tests the updateProfile handler
func TestUpdateProfileAPI(t *testing.T) {
	user, _ := randomUser(t)
	
	// Create updated profile
	updatedUser := user
	newBio := "New bio text"
	updatedUser.Bio = pgtype.Text{String: newBio, Valid: true}

	testCases := []struct {
		name          string
		userID        string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			userID: user.ID.String(),
			body: gin.H{
				"bio": newBio,
				"profile_picture": "",
				"background_image": "",
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.UpdateProfileParams) (db.User, error) {
						require.Equal(t, newBio, params.Bio.String)
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
				
				require.Equal(t, newBio, gotResponse.Bio)
			},
		},
		{
			name:   "InvalidID",
			userID: "invalid-id",
			body: gin.H{
				"bio": newBio,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			userID: user.ID.String(),
			body: gin.H{
				"bio": newBio,
				"profile_picture": "",
				"background_image": "",
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					UpdateProfile(gomock.Any(), gomock.Any()).
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

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/api/v1/users/%s/profile", tc.userID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestFinishOnboardingAPI tests the finishOnboarding handler
func TestFinishOnboardingAPI(t *testing.T) {
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

			url := fmt.Sprintf("/api/v1/users/%s/onboarding", tc.userID)
			request, err := http.NewRequest(http.MethodPut, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestDeleteUserAPI tests the deleteUser handler
func TestDeleteUserAPI(t *testing.T) {
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

			url := fmt.Sprintf("/api/v1/users/%s", tc.userID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestSearchUsersAPI tests the searchUsers handler
func TestSearchUsersAPI(t *testing.T) {
	searchUser, _ := randomUser(t)
	
	// Create test search results
	searchResultTrigram := db.SearchUsersTrigramRow{
		ID: searchUser.ID,
		Username: searchUser.Username,
		Fullname: searchUser.Fullname,
		ProfilePicture: searchUser.ProfilePicture,
	}
	
	searchResultILike := db.SearchUsersILikeRow{
		ID: searchUser.ID,
		Username: searchUser.Username,
		Fullname: searchUser.Fullname,
		ProfilePicture: searchUser.ProfilePicture,
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK_Trigram",
			body: gin.H{
				"search_term": "test_search_term_longer",
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					SearchUsersTrigram(gomock.Any(), gomock.Eq("test_search_term_longer")).
					Times(1).
					Return([]db.SearchUsersTrigramRow{searchResultTrigram}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				
				var response []UserSearchResponse
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				
				err = json.Unmarshal(data, &response)
				require.NoError(t, err)
				
				require.Len(t, response, 1)
				require.Equal(t, searchUser.ID.String(), response[0].ID)
				require.Equal(t, searchUser.Username, response[0].Username)
			},
		},
		{
			name: "OK_ILike",
			body: gin.H{
				"search_term": "ab",
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					SearchUsersILike(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, searchTerm pgtype.Text) ([]db.SearchUsersILikeRow, error) {
						require.Equal(t, "ab", searchTerm.String)
						return []db.SearchUsersILikeRow{searchResultILike}, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				
				var response []UserSearchResponse
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				
				err = json.Unmarshal(data, &response)
				require.NoError(t, err)
				
				require.Len(t, response, 1)
				require.Equal(t, searchUser.ID.String(), response[0].ID)
			},
		},
		{
			name: "MissingSearchTerm",
			body: gin.H{
				// Missing search_term
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					SearchUsersTrigram(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					SearchUsersILike(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"search_term": "test_search",
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					SearchUsersTrigram(gomock.Any(), gomock.Any()).
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
			testUser, _ := randomUser(t)
			server.router.POST("/test/search", func(ctx *gin.Context) {
				ctx.Set("currentUser", testUser)
				server.searchUsers(ctx)
			})

			recorder := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/test/search"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestUpdateUserNewAPI tests the updateUserNew handler
func TestUpdateUserNewAPI(t *testing.T) {
	currentUser, _ := randomUser(t)
	
	// Create updated user
	updatedUser := currentUser
	newUsername := util.RandomUsername()
	updatedUser.Username = newUsername

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
				"username": []string{"invalid"}, // Invalid type
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

		
			server.router.POST("/test/updateuser/v2", func(ctx *gin.Context) {
				ctx.Set("currentUser", currentUser)
				server.updateUserNew(ctx)
			})

			tc.setupMock(mockHub)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/test/updateuser/v2"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUUIDGeneration(t *testing.T) {
    // Generate a UUID directly
    rawUUID := uuid.New()
    rawUUIDString := rawUUID.String()
    t.Logf("Raw UUID string: %s", rawUUIDString)
    
    // Test the pgtype.UUID conversion
    var pgUUID pgtype.UUID
    err := pgUUID.Scan(rawUUIDString)
    require.NoError(t, err)
    t.Logf("pgtype.UUID after scanning: %s", pgUUID.String())
    
    // Test full user creation
	var testUser db.User
	testUser.ID = pgUUID
    t.Logf("User ID: %s", testUser.ID.String())
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
        OnboardingAt:    pgtype.Timestamp{}, // Empty timestamp
        CreatedAt:       createdAt,
        UpdatedAt:       updateAt,
    }

    return user, password
}

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

