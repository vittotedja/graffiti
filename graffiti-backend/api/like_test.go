package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	mockdb "github.com/vittotedja/graffiti/graffiti-backend/db/mock"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
)

// TestUpdateLikeAPI tests the updateLike handler
func TestUpdateLikeAPI(t *testing.T) {
	user, _ := randomUser(t)
	post := randomPost(t, pgtype.UUID{}, pgtype.UUID{})

	testCases := []struct {
		name          string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK_Like",
			body: gin.H{
				"post_id": post.ID.String(),
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), post.ID).
					Return(post, nil)

				mockHub.EXPECT().
					CreateOrDeleteLikeTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(true, nil) // true for liked
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response["message"], "liked")
			},
		},
		{
			name: "OK_Unlike",
			body: gin.H{
				"post_id": post.ID.String(),
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateOrDeleteLikeTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(false, nil) // false for unliked
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response["message"], "unliked")
			},
		},
		{
			name: "BadRequest_MissingPostID",
			body: gin.H{},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateOrDeleteLikeTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequest_InvalidPostID",
			body: gin.H{
				"post_id": "invalid-uuid",
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateOrDeleteLikeTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"post_id": post.ID.String(),
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateOrDeleteLikeTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(false, sql.ErrConnDone)
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

			server.router.POST("/test/likes", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.updateLike(ctx)
			})

			recorder := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/test/likes", bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetLikeAPI tests the getLike handler
func TestGetLikeAPI(t *testing.T) {
	user, _ := randomUser(t)
	post := randomPost(t, pgtype.UUID{}, pgtype.UUID{})

	testCases := []struct {
		name          string
		postID        string
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK_Liked",
			postID: post.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetLike(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Like{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				var response map[string]bool
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.True(t, response["liked"])
			},
		},
		{
			name:   "OK_NotLiked",
			postID: post.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetLike(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Like{}, pgx.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				var response map[string]bool
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				require.False(t, response["liked"])
			},
		},
		{
			name:   "BadRequest_InvalidPostID",
			postID: "invalid-uuid",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetLike(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			postID: post.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetLike(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Like{}, sql.ErrConnDone)
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

			server.router.GET("/test/likes/:post_id", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.getLike(ctx)
			})

			recorder := httptest.NewRecorder()
			url := "/test/likes/" + tc.postID
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestDeleteLikeAPI tests the deleteLike handler
func TestDeleteLikeAPI(t *testing.T) {
	post := randomPost(t, pgtype.UUID{}, pgtype.UUID{})
	user, _ := randomUser(t)

	testCases := []struct {
		name          string
		postID        string
		userID        string
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			postID: post.ID.String(),
			userID: user.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					DeleteLike(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:   "BadRequest_InvalidPostID",
			postID: "invalid-uuid",
			userID: user.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					DeleteLike(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "BadRequest_InvalidUserID",
			postID: post.ID.String(),
			userID: "invalid-uuid",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					DeleteLike(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			postID: post.ID.String(),
			userID: user.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					DeleteLike(gomock.Any(), gomock.Any()).
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

			server.router.DELETE("/test/likes/:post_id/:user_id", func(ctx *gin.Context) {
				server.deleteLike(ctx)
			})

			recorder := httptest.NewRecorder()
			url := "/test/likes/" + tc.postID + "/" + tc.userID
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListLikesAPI tests the listLikes handler
func TestListLikesAPI(t *testing.T) {
	n := 5
	likes := make([]db.Like, n)
	for i := 0; i < n; i++ {
		likes[i] = randomLike(t)
	}

	testCases := []struct {
		name          string
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListLikes(gomock.Any()).
					Times(1).
					Return(likes, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchLikes(t, recorder.Body, likes)
			},
		},
		{
			name: "InternalError",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListLikes(gomock.Any()).
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

			server.router.GET("/test/likes", func(ctx *gin.Context) {
				server.listLikes(ctx)
			})

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, "/test/likes", nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListLikesByPostAPI tests the listLikesByPost handler
func TestListLikesByPostAPI(t *testing.T) {
	post := randomPost(t, pgtype.UUID{}, pgtype.UUID{})

	n := 5
	likes := make([]db.Like, n)
	for i := 0; i < n; i++ {
		likes[i] = randomLike(t)
		likes[i].PostID = post.ID // Same post ID for all likes
	}

	testCases := []struct {
		name          string
		postID        string
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			postID: post.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListLikesByPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(likes, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchLikes(t, recorder.Body, likes)
			},
		},
		{
			name:   "BadRequest_InvalidPostID",
			postID: "invalid-uuid",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListLikesByPost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			postID: post.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListLikesByPost(gomock.Any(), gomock.Any()).
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

			server.router.GET("/test/posts/:post_id/likes", func(ctx *gin.Context) {
				server.listLikesByPost(ctx)
			})

			recorder := httptest.NewRecorder()
			url := "/test/posts/" + tc.postID + "/likes"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListLikesByUserAPI tests the listLikesByUser handler
func TestListLikesByUserAPI(t *testing.T) {
	user, _ := randomUser(t)

	n := 5
	likes := make([]db.Like, n)
	for i := 0; i < n; i++ {
		likes[i] = randomLike(t)
		likes[i].UserID = user.ID // Same user ID for all likes
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
					ListLikesByUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(likes, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchLikes(t, recorder.Body, likes)
			},
		},
		{
			name:   "BadRequest_InvalidUserID",
			userID: "invalid-uuid",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListLikesByUser(gomock.Any(), gomock.Any()).
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
					ListLikesByUser(gomock.Any(), gomock.Any()).
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

			server.router.GET("/test/users/:user_id/likes", func(ctx *gin.Context) {
				server.listLikesByUser(ctx)
			})

			recorder := httptest.NewRecorder()
			url := "/test/users/" + tc.userID + "/likes"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// Helper function to create a random like
func randomLike(t *testing.T) db.Like {
	id := pgtype.UUID{}
	err := id.Scan(uuid.New().String())
	require.NoError(t, err)

	postID := pgtype.UUID{}
	err = postID.Scan(uuid.New().String())
	require.NoError(t, err)

	userID := pgtype.UUID{}
	err = userID.Scan(uuid.New().String())
	require.NoError(t, err)

	likedAt := pgtype.Timestamp{}
	err = likedAt.Scan(time.Now())
	require.NoError(t, err)

	return db.Like{
		ID:      id,
		PostID:  postID,
		UserID:  userID,
		LikedAt: likedAt,
	}
}

// Helper function to check if the response body matches the likes
func requireBodyMatchLikes(t *testing.T, body *bytes.Buffer, likes []db.Like) {
	data, err := json.Marshal(likes)
	require.NoError(t, err)

	var expected []interface{}
	err = json.Unmarshal(data, &expected)
	require.NoError(t, err)

	var got []interface{}
	err = json.Unmarshal(body.Bytes(), &got)
	require.NoError(t, err)

	require.Equal(t, len(expected), len(got))
}
