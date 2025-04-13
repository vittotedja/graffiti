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
)

// TestCreateFriendRequestAPI tests the createFriendRequest handler
func TestCreateFriendRequestAPI(t *testing.T) {
	user1, _ := randomUser(t)
	user2, _ := randomUser(t)

	friendship := randomFriendship(t, user1.ID, user2.ID)

	testCases := []struct {
		name          string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"to_user_id": user2.ID.String(),
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateFriendRequestTx(gomock.Any(), gomock.Eq(user1.ID), gomock.Any()).
					Times(1).
					Return(friendship, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchFriendshipResponse(t, recorder.Body, friendship)
			},
		},
		{
			name: "BadRequest_MissingToUserID",
			body: gin.H{},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateFriendRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequest_InvalidToUserID",
			body: gin.H{
				"to_user_id": "invalid-uuid",
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateFriendRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequest_SelfFriending",
			body: gin.H{
				"to_user_id": user1.ID.String(), // Same as current user
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateFriendRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"to_user_id": user2.ID.String(),
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateFriendRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Friendship{}, sql.ErrConnDone)
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

			server.router.POST("/test/friend-requests", func(ctx *gin.Context) {
				ctx.Set("currentUser", user1)
				server.createFriendRequest(ctx)
			})

			recorder := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/test/friend-requests", bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestAcceptFriendRequestAPI tests the acceptFriendRequest handler
func TestAcceptFriendRequestAPI(t *testing.T) {
	user1, _ := randomUser(t)
	user2, _ := randomUser(t)
	friendship := randomFriendship(t, user1.ID, user2.ID)
	
	// Third user for unauthorized test
	user3, _ := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		currentUser   db.User
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"friendship_id": friendship.ID.String(),
			},
			currentUser: user2, 
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any()).
					Times(1).
					Return(friendship, nil)
					
				mockHub.EXPECT().
					AcceptFriendRequestTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "BadRequest_MissingID",
			body: gin.H{},
			currentUser: user2,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					AcceptFriendRequestTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequest_InvalidID",
			body: gin.H{
				"friendship_id": "invalid-uuid",
			},
			currentUser: user2,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					AcceptFriendRequestTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "FriendshipNotFound",
			body: gin.H{
				"friendship_id": uuid.New().String(),
			},
			currentUser: user2,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Friendship{}, db.ErrRecordNotFound)
				mockHub.EXPECT().
					AcceptFriendRequestTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			body: gin.H{
				"friendship_id": friendship.ID.String(),
			},
			currentUser: user3, // Neither sender nor recipient
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any()).
					Times(1).
					Return(friendship, nil)
				mockHub.EXPECT().
					AcceptFriendRequestTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"friendship_id": friendship.ID.String(),
			},
			currentUser: user2,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetFriendship(gomock.Any(), gomock.Any()).
					Times(1).
					Return(friendship, nil)
					
				mockHub.EXPECT().
					AcceptFriendRequestTx(gomock.Any(), gomock.Any()).
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

			server.router.PUT("/test/friend-requests/accept", func(ctx *gin.Context) {
				ctx.Set("currentUser", tc.currentUser)
				server.acceptFriendRequest(ctx)
			})

			recorder := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPut, "/test/friend-requests/accept", bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListFriendshipsByUserIdAPI tests the listFriendshipsByUserId handler
func TestListFriendshipsByUserIdAPI(t *testing.T) {
	user, _ := randomUser(t)
	
	n := 5
	friendships := make([]db.Friendship, n)
	for i := 0; i < n; i++ {
		otherUser, _ := randomUser(t)
		friendships[i] = randomFriendship(t, user.ID, otherUser.ID)
	}

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
					ListFriendshipsByUserId(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(friendships, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchFriendshipsResponse(t, recorder.Body, friendships)
			},
		},
		{
			name:   "InvalidID",
			userID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendshipsByUserId(gomock.Any(), gomock.Any()).
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
					ListFriendshipsByUserId(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/users/%s/friendships", tc.userID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetNumberOfFriendsAPI tests the getNumberOfFriends handler
func TestGetNumberOfFriendsAPI(t *testing.T) {
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
					GetNumberOfFriends(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(int64(10), nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				
				var response struct {
					Count int64 `json:"count"`
				}
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				
				err = json.Unmarshal(data, &response)
				require.NoError(t, err)
				require.Equal(t, int64(10), response.Count)
			},
		},
		{
			name:   "InvalidID",
			userID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetNumberOfFriends(gomock.Any(), gomock.Any()).
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
					GetNumberOfFriends(gomock.Any(), gomock.Any()).
					Times(1).
					Return(int64(0), sql.ErrConnDone)
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

			url := fmt.Sprintf("/api/v1/users/%s/accepted-friends/count", tc.userID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetFriendsByStatusAPI tests the getFriendsByStatus handler
func TestGetFriendsByStatusAPI(t *testing.T) {
	user, _ := randomUser(t)

	n := 5
	friendDetails := make([]db.ListFriendsDetailsByStatusRow, n)
	for i := 0; i < n; i++ {
		otherUser, _ := randomUser(t)
		
		var status db.NullStatus
		status.Status = "friends"
		status.Valid = true
		
		friendDetails[i] = db.ListFriendsDetailsByStatusRow{
			UserID:         otherUser.ID,
			Fullname:       otherUser.Fullname,
			Username:       otherUser.Username,
			ProfilePicture: otherUser.ProfilePicture,
			Status:         status,
			ID:             pgtype.UUID{}, // Friendship ID
		}
	}

	testCases := []struct {
		name          string
		queryType     string
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK_Friends",
			queryType: "friends",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendsDetailsByStatus(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.ListFriendsDetailsByStatusParams) ([]db.ListFriendsDetailsByStatusRow, error) {
						require.Equal(t, user.ID, params.FromUser)
						require.Equal(t, "friends", params.Column2)
						return friendDetails, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				
				var response []db.ListFriendsDetailsByStatusRow
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				
				err = json.Unmarshal(data, &response)
				require.NoError(t, err)
				require.Len(t, response, n)
			},
		},
		{
			name:      "OK_Sent",
			queryType: "sent",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendsDetailsByStatus(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.ListFriendsDetailsByStatusParams) ([]db.ListFriendsDetailsByStatusRow, error) {
						require.Equal(t, user.ID, params.FromUser)
						require.Equal(t, "sent", params.Column2)
						return friendDetails, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "OK_Requested",
			queryType: "requested",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendsDetailsByStatus(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.ListFriendsDetailsByStatusParams) ([]db.ListFriendsDetailsByStatusRow, error) {
						require.Equal(t, user.ID, params.FromUser)
						require.Equal(t, "requested", params.Column2)
						return friendDetails, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "BadRequest_InvalidType",
			queryType: "invalid",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendsDetailsByStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			queryType: "friends",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendsDetailsByStatus(gomock.Any(), gomock.Any()).
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

			server.router.GET("/test/friends", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.getFriendsByStatus(ctx)
			})

			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/test/friends?type=%s", tc.queryType)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestDeleteFriendshipAPI tests the deleteFriendship handler
func TestDeleteFriendshipAPI(t *testing.T) {
	user, _ := randomUser(t)
	friendship := randomFriendship(t, user.ID, pgtype.UUID{})

	testCases := []struct {
		name          string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"friendship_id": friendship.ID.String(),
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				var id pgtype.UUID
				id.Scan(friendship.ID.String())

				mockHub.EXPECT().
					DeleteFriendship(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "BadRequest_MissingID",
			body: gin.H{},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					DeleteFriendship(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequest_InvalidID",
			body: gin.H{
				"friendship_id": "invalid-uuid",
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					DeleteFriendship(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"friendship_id": friendship.ID.String(),
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					DeleteFriendship(gomock.Any(), gomock.Any()).
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

			server.router.DELETE("/test/friendships", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.deleteFriendship(ctx)
			})

			recorder := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodDelete, "/test/friendships", bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListFriendshipByUserPairsAPI tests the listFriendshipByUserPairs handler
func TestListFriendshipByUserPairsAPI(t *testing.T) {
	user1, _ := randomUser(t)
	user2, _ := randomUser(t)
	friendship := randomFriendship(t, user1.ID, user2.ID)

	testCases := []struct {
		name          string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"to_user_id": user2.ID.String(),
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendshipByUserPairs(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.ListFriendshipByUserPairsParams) (db.Friendship, error) {
						require.Equal(t, user1.ID, params.FromUser)
						require.Equal(t, user2.ID.String(), params.ToUser.String())
						return friendship, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchFriendshipResponse(t, recorder.Body, friendship)
			},
		},
		{
			name: "BadRequest_MissingToUserID",
			body: gin.H{},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendshipByUserPairs(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequest_InvalidToUserID",
			body: gin.H{
				"to_user_id": "invalid-uuid",
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendshipByUserPairs(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequest_SelfFriendship",
			body: gin.H{
				"to_user_id": user1.ID.String(), // Same as currentUser
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendshipByUserPairs(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NoFriendship",
			body: gin.H{
				"to_user_id": user2.ID.String(),
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendshipByUserPairs(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Friendship{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				
				var response map[string]interface{}
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				
				err = json.Unmarshal(data, &response)
				require.NoError(t, err)
				require.Equal(t, "00000000", response["ID"])
				require.NotNil(t, response["Status"])
			},
		}, 
		{
			name: "InternalError",
			body: gin.H{
				"to_user_id": user2.ID.String(),
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListFriendshipByUserPairs(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Friendship{}, sql.ErrConnDone) // Different than ErrNoRows
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

			server.router.POST("/test/friendships", func(ctx *gin.Context) {
				ctx.Set("currentUser", user1)
				server.listFriendshipByUserPairs(ctx)
			})

			recorder := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/test/friendships", bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetPendingFriendRequestsAPI tests the getPendingFriendRequests handler
func TestGetPendingFriendRequestsAPI(t *testing.T) {
	user, _ := randomUser(t)
	
	n := 3
	requests := make([]db.Friendship, n)
	for i := 0; i < n; i++ {
		otherUser, _ := randomUser(t)
		requests[i] = randomFriendship(t, otherUser.ID, user.ID)
		requests[i].Status.Status = "pending"
	}

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
				mockHub.EXPECT().
					GetPendingFriendRequestsTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(requests, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchFriendshipsResponse(t, recorder.Body, requests)
			},
		},
		{
			name:   "InvalidID",
			userID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPendingFriendRequestsTx(gomock.Any(), gomock.Any()).
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
					GetPendingFriendRequestsTx(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/users/%s/friend-requests/pending", tc.userID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetSentFriendRequestsAPI tests the getSentFriendRequests handler
func TestGetSentFriendRequestsAPI(t *testing.T) {
	user, _ := randomUser(t)
	
	n := 3
	requests := make([]db.Friendship, n)
	for i := 0; i < n; i++ {
		otherUser, _ := randomUser(t)
		requests[i] = randomFriendship(t, user.ID, otherUser.ID)
		requests[i].Status.Status = "pending"
	}

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
				mockHub.EXPECT().
					GetSentFriendRequestsTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(requests, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchFriendshipsResponse(t, recorder.Body, requests)
			},
		},
		{
			name:   "InvalidID",
			userID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetSentFriendRequestsTx(gomock.Any(), gomock.Any()).
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
					GetSentFriendRequestsTx(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/users/%s/friend-requests/sent", tc.userID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetNumberOfPendingFriendRequestsAPI tests the getNumberOfPendingFriendRequests handler
func TestGetNumberOfPendingFriendRequestsAPI(t *testing.T) {
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
				mockHub.EXPECT().
					GetNumberOfPendingFriendRequests(gomock.Any(), gomock.Any()).
					Times(1).
					Return(int64(5), nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				
				var response struct {
					Count int64 `json:"count"`
				}
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				
				err = json.Unmarshal(data, &response)
				require.NoError(t, err)
				require.Equal(t, int64(5), response.Count)
			},
		},
		{
			name:   "InvalidID",
			userID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetNumberOfPendingFriendRequests(gomock.Any(), gomock.Any()).
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
					GetNumberOfPendingFriendRequests(gomock.Any(), gomock.Any()).
					Times(1).
					Return(int64(0), sql.ErrConnDone)
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

			url := fmt.Sprintf("/api/v1/users/%s/friend-requests/pending/count", tc.userID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// Helper function to create a random friendship
func randomFriendship(t *testing.T, fromUserID, toUserID pgtype.UUID) db.Friendship {
	id := pgtype.UUID{}
	id.Scan(uuid.New().String())
	
	// If toUserID is empty, generate a random one
	if !toUserID.Valid {
		toUserID = pgtype.UUID{}
		toUserID.Scan(uuid.New().String())
	}
	
	status := db.NullStatus{}
	status.Status = "pending"
	status.Valid = true
	
	createdAt := pgtype.Timestamp{}
	createdAt.Scan(time.Now())
	
	updatedAt := pgtype.Timestamp{}
	updatedAt.Scan(time.Now())
	
	return db.Friendship{
		ID:        id,
		FromUser:  fromUserID,
		ToUser:    toUserID,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func requireBodyMatchFriendshipResponse(t *testing.T, body *bytes.Buffer, friendship db.Friendship) {
    data, err := io.ReadAll(body)
    require.NoError(t, err)

    // Print the raw JSON for debugging
    t.Logf("Raw JSON response: %s", string(data))

    var gotResponse map[string]interface{}
    err = json.Unmarshal(data, &gotResponse)
    require.NoError(t, err)
    
    // Basic validations
    require.NotEmpty(t, gotResponse["ID"])
    require.NotEmpty(t, gotResponse["FromUser"])
    require.NotEmpty(t, gotResponse["ToUser"])
    
    // Compare ID and user IDs as strings
    require.Equal(t, friendship.ID.String(), gotResponse["ID"])
    require.Equal(t, friendship.FromUser.String(), gotResponse["FromUser"])
    require.Equal(t, friendship.ToUser.String(), gotResponse["ToUser"])
    
    // Get the Status map and extract the Status field
    statusMap, ok := gotResponse["Status"].(map[string]interface{})
    require.True(t, ok, "Status field is not a map")
    
    statusValue, exists := statusMap["Status"]
    require.True(t, exists, "Status.Status field not found")
    
    // Convert db.Status to string explicitly for comparison
    expectedStatusStr := string(friendship.Status.Status)
    actualStatusStr := statusValue.(string)
    
    require.Equal(t, expectedStatusStr, actualStatusStr, 
        "Status values don't match: expected %s, got %s", expectedStatusStr, actualStatusStr)
}

func requireBodyMatchFriendshipsResponse(t *testing.T, body *bytes.Buffer, friendships []db.Friendship) {
    data, err := io.ReadAll(body)
    require.NoError(t, err)
    
    var gotResponses []map[string]interface{}
    err = json.Unmarshal(data, &gotResponses)
    require.NoError(t, err)

    require.Equal(t, len(friendships), len(gotResponses))
    
    // Create a map of friendship IDs for easier lookup
    friendshipMap := make(map[string]db.Friendship)
    for _, f := range friendships {
        friendshipMap[f.ID.String()] = f
    }

    for _, resp := range gotResponses {
        id := resp["ID"].(string)
        original, exists := friendshipMap[id]
        require.True(t, exists)
        
        require.Equal(t, original.FromUser.String(), resp["FromUser"])
        require.Equal(t, original.ToUser.String(), resp["ToUser"])
        
        // Get the Status map and extract the Status field
        statusMap, ok := resp["Status"].(map[string]interface{})
        require.True(t, ok, "Status field is not a map")
        
        statusValue, exists := statusMap["Status"]
        require.True(t, exists, "Status.Status field not found")
        
        // Convert db.Status to string explicitly for comparison
        expectedStatusStr := string(original.Status.Status)
        actualStatusStr := statusValue.(string)
        
        require.Equal(t, expectedStatusStr, actualStatusStr,
            "Status values don't match: expected %s, got %s", expectedStatusStr, actualStatusStr)
    }
}