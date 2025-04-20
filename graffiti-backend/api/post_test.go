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

// TestCreatePostAPI tests the createPost handler
func TestCreatePostAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)
	post := randomPost(t, wall.ID, user.ID)

	const validPostType = "media"

	testCases := []struct {
		name          string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"wall_id":   wall.ID,
				"media_url": post.MediaUrl.String,
				"post_type": validPostType,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), wall.ID).
					Return(wall, nil)

				mockHub.EXPECT().
					CreatePost(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.CreatePostParams) (db.Post, error) {
						require.Equal(t, wall.ID.String(), params.WallID.String())
						require.Equal(t, user.ID.String(), params.Author.String())
						require.Equal(t, post.MediaUrl.String, params.MediaUrl.String)
						require.Equal(t, db.PostType(validPostType), params.PostType.PostType)
						return post, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchPostResponse(t, recorder.Body, post)
			},
		},
		{
			name: "BadRequest_MissingRequired",
			body: gin.H{
				"media_url": post.MediaUrl.String,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreatePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequest_InvalidWallID",
			body: gin.H{
				"wall_id":   "invalid-uuid",
				"media_url": post.MediaUrl.String,
				"post_type": validPostType,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreatePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequest_InvalidPostType",
			body: gin.H{
				"wall_id":   wall.ID.String(),
				"media_url": post.MediaUrl.String,
				"post_type": "invalid_type",
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreatePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"wall_id":   wall.ID.String(),
				"media_url": post.MediaUrl.String,
				"post_type": validPostType,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), wall.ID).
					Return(wall, nil)
				mockHub.EXPECT().
					CreatePost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Post{}, sql.ErrConnDone)
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

			server.router.POST("/test/posts", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.createPost(ctx)
			})

			recorder := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/test/posts", bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetPostAPI tests the getPost handler
func TestGetPostAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)
	post := randomPost(t, wall.ID, user.ID)

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
				var id pgtype.UUID
				id.Scan(post.ID.String())

				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(post, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPostResponse(t, recorder.Body, post)
			},
		},
		{
			name:   "PostNotFound",
			postID: uuid.New().String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Post{}, db.ErrRecordNotFound)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			postID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
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
					GetPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Post{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/api/v1/posts/%s", tc.postID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListPostsAPI tests the listPosts handler
func TestListPostsAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)

	n := 5
	posts := make([]db.Post, n)
	for i := 0; i < n; i++ {
		posts[i] = randomPost(t, wall.ID, user.ID)
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
					ListPosts(gomock.Any()).
					Times(1).
					Return(posts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPostsResponse(t, recorder.Body, posts)
			},
		},
		{
			name: "InternalError",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListPosts(gomock.Any()).
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

			request, err := http.NewRequest(http.MethodGet, "/api/v1/posts", nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListPostsByWallAPI tests the listPostsByWall handler
func TestListPostsByWallAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)

	n := 5
	posts := make([]db.Post, n)
	for i := 0; i < n; i++ {
		posts[i] = randomPost(t, wall.ID, user.ID)
	}

	testCases := []struct {
		name          string
		wallID        string
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			wallID: wall.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				var id pgtype.UUID
				id.Scan(wall.ID.String())

				mockHub.EXPECT().
					ListPostsByWall(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(posts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPostsResponse(t, recorder.Body, posts)
			},
		},
		{
			name:   "InvalidID",
			wallID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListPostsByWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			wallID: wall.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListPostsByWall(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/walls/%s/posts", tc.wallID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListPostsByWallWithAuthorsDetailsAPI tests the listPostsByWallWithAuthorsDetails handler
func TestListPostsByWallWithAuthorsDetailsAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)

	n := 5
	postsWithAuthors := make([]db.ListPostsByWallWithAuthorsDetailsRow, n)
	for i := 0; i < n; i++ {
		post := randomPost(t, wall.ID, user.ID)
		postsWithAuthors[i] = db.ListPostsByWallWithAuthorsDetailsRow{
			ID:             post.ID,
			WallID:         post.WallID,
			Author:         post.Author,
			MediaUrl:       post.MediaUrl,
			PostType:       post.PostType,
			IsHighlighted:  post.IsHighlighted,
			LikesCount:     post.LikesCount,
			IsDeleted:      post.IsDeleted,
			CreatedAt:      post.CreatedAt,
			Username:       user.Username,
			ProfilePicture: user.ProfilePicture,
			Fullname:       user.Fullname,
		}
	}

	testCases := []struct {
		name          string
		wallID        string
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			wallID: wall.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				var id pgtype.UUID
				id.Scan(wall.ID.String())

				mockHub.EXPECT().
					ListPostsByWallWithAuthorsDetails(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(postsWithAuthors, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var gotResponse []PostResponseWithAuthor
				err = json.Unmarshal(data, &gotResponse)
				require.NoError(t, err)
				require.Len(t, gotResponse, n)
			},
		},
		{
			name:   "InvalidID",
			wallID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListPostsByWallWithAuthorsDetails(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			wallID: wall.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListPostsByWallWithAuthorsDetails(gomock.Any(), gomock.Any()).
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

			server.router.GET("/test/walls/:id/posts", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.listPostsByWallWithAuthorsDetails(ctx)
			})

			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/test/walls/%s/posts", tc.wallID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetHighlightedPostsAPI tests the getHighlightedPosts handler
func TestGetHighlightedPostsAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)

	n := 5
	posts := make([]db.Post, n)
	for i := 0; i < n; i++ {
		post := randomPost(t, wall.ID, user.ID)
		post.IsHighlighted.Bool = true
		posts[i] = post
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
					GetHighlightedPosts(gomock.Any()).
					Times(1).
					Return(posts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPostsResponse(t, recorder.Body, posts)
			},
		},
		{
			name: "InternalError",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetHighlightedPosts(gomock.Any()).
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

			request, err := http.NewRequest(http.MethodGet, "/api/v1/posts/highlighted", nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetHighlightedPostsByWallAPI tests the getHighlightedPostsByWall handler
func TestGetHighlightedPostsByWallAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)

	n := 5
	posts := make([]db.Post, n)
	for i := 0; i < n; i++ {
		post := randomPost(t, wall.ID, user.ID)
		post.IsHighlighted.Bool = true
		posts[i] = post
	}

	testCases := []struct {
		name          string
		wallID        string
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			wallID: wall.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				var id pgtype.UUID
				id.Scan(wall.ID.String())

				mockHub.EXPECT().
					GetHighlightedPostsByWall(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(posts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPostsResponse(t, recorder.Body, posts)
			},
		},
		{
			name:   "InvalidID",
			wallID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetHighlightedPostsByWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			wallID: wall.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetHighlightedPostsByWall(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/walls/%s/posts/highlighted", tc.wallID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestUpdatePostAPI tests the updatePost handler
func TestUpdatePostAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)
	post := randomPost(t, wall.ID, user.ID)

	updatedPost := post
	newMediaURL := "https://updated-url.com/image.jpg"
	updatedPost.MediaUrl.String = newMediaURL

	testCases := []struct {
		name          string
		postID        string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			postID: post.ID.String(),
			body: gin.H{
				"media_url": newMediaURL,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(post, nil)

				mockHub.EXPECT().
					UpdatePost(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.UpdatePostParams) (db.Post, error) {
						require.Equal(t, post.ID.String(), params.ID.String())
						require.Equal(t, newMediaURL, params.MediaUrl.String)
						return updatedPost, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPostResponse(t, recorder.Body, updatedPost)
			},
		},
		{
			name:   "InvalidID",
			postID: "invalid-id",
			body: gin.H{
				"media_url": newMediaURL,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					UpdatePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "PostNotFound",
			postID: uuid.New().String(),
			body: gin.H{
				"media_url": newMediaURL,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Post{}, db.ErrRecordNotFound)

				mockHub.EXPECT().
					UpdatePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			postID: post.ID.String(),
			body: gin.H{
				"media_url": newMediaURL,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(post, nil)

				mockHub.EXPECT().
					UpdatePost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Post{}, sql.ErrConnDone)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/api/v1/posts/%s", tc.postID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestHighlightPostAPI tests the highlightPost handler
func TestHighlightPostAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)
	post := randomPost(t, wall.ID, user.ID)
	post.IsHighlighted.Bool = false

	highlightedPost := post
	highlightedPost.IsHighlighted.Bool = true

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
				var id pgtype.UUID
				id.Scan(post.ID.String())

				mockHub.EXPECT().
					HighlightPost(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(highlightedPost, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPostResponse(t, recorder.Body, highlightedPost)
			},
		},
		{
			name:   "InvalidID",
			postID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					HighlightPost(gomock.Any(), gomock.Any()).
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
					HighlightPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Post{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/api/v1/posts/%s/highlight", tc.postID)
			request, err := http.NewRequest(http.MethodPut, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestUnhighlightPostAPI tests the unhighlightPost handler
func TestUnhighlightPostAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)
	post := randomPost(t, wall.ID, user.ID)
	post.IsHighlighted.Bool = true

	unhighlightedPost := post
	unhighlightedPost.IsHighlighted.Bool = false

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
				var id pgtype.UUID
				id.Scan(post.ID.String())

				mockHub.EXPECT().
					UnhighlightPost(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(unhighlightedPost, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPostResponse(t, recorder.Body, unhighlightedPost)
			},
		},
		{
			name:   "InvalidID",
			postID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					UnhighlightPost(gomock.Any(), gomock.Any()).
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
					UnhighlightPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Post{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/api/v1/posts/%s/unhighlight", tc.postID)
			request, err := http.NewRequest(http.MethodPut, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestDeletePostAPI tests the deletePost handler
func TestDeletePostAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)
	post := randomPost(t, wall.ID, user.ID)

	otherUser, _ := randomUser(t)
	otherUserPost := randomPost(t, wall.ID, otherUser.ID)

	testCases := []struct {
		name          string
		postID        string
		currentUser   db.User
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK_PostOwner",
			postID:      post.ID.String(),
			currentUser: user,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(post, nil)

				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				mockHub.EXPECT().
					DeletePost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)

				mockHub.EXPECT().
					ListPostsByWall(gomock.Any(), gomock.Any()).
					AnyTimes().
					Return([]db.Post{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:        "OK_WallOwner",
			postID:      otherUserPost.ID.String(),
			currentUser: user, 
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(otherUserPost, nil)

				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				mockHub.EXPECT().
					DeletePost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)

				mockHub.EXPECT().
					ListPostsByWall(gomock.Any(), gomock.Any()).
					AnyTimes().
					Return([]db.Post{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:        "Unauthorized",
			postID:      otherUserPost.ID.String(),
			currentUser: otherUser, 
			setupMock: func(mockHub *mockdb.MockHub) {
				differentWall := wall
				differentWall.UserID = user.ID

				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(post, nil)

				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(differentWall, nil)

				mockHub.EXPECT().
					DeletePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:        "InvalidID",
			postID:      "invalid-id",
			currentUser: user,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					DeletePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:        "PostNotFound",
			postID:      uuid.New().String(),
			currentUser: user,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Post{}, db.ErrRecordNotFound)

				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					DeletePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:        "WallNotFound",
			postID:      post.ID.String(),
			currentUser: user,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(post, nil)

				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, db.ErrRecordNotFound)

				mockHub.EXPECT().
					DeletePost(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:        "DeleteError",
			postID:      post.ID.String(),
			currentUser: user,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetPost(gomock.Any(), gomock.Any()).
					Times(1).
					Return(post, nil)

				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				mockHub.EXPECT().
					DeletePost(gomock.Any(), gomock.Any()).
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
			server.router.DELETE("/test/posts/:id", func(ctx *gin.Context) {
				ctx.Set("currentUser", tc.currentUser)
				server.deletePost(ctx)
			})

			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/test/posts/%s", tc.postID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// Helper function to create a random post
func randomPost(t *testing.T, wallID pgtype.UUID, authorID pgtype.UUID) db.Post {
	id := pgtype.UUID{}
	id.Scan(uuid.New().String())

	mediaURL := pgtype.Text{}
	mediaURL.Scan(fmt.Sprintf("https://example.com/images/%s.jpg", util.RandomString(10)))

	postType := db.NullPostType{}
	postType.Scan(db.PostTypeMedia)

	isHighlighted := pgtype.Bool{}
	isHighlighted.Scan(false)

	likesCount := pgtype.Int4{}
	likesCount.Scan(0)

	isDeleted := pgtype.Bool{}
	isDeleted.Scan(false)

	createdAt := pgtype.Timestamp{}
	createdAt.Scan(time.Now())

	return db.Post{
		ID:            id,
		WallID:        wallID,
		Author:        authorID,
		MediaUrl:      mediaURL,
		PostType:      postType,
		IsHighlighted: isHighlighted,
		LikesCount:    likesCount,
		IsDeleted:     isDeleted,
		CreatedAt:     createdAt,
	}
}

func requireBodyMatchPostResponse(t *testing.T, body *bytes.Buffer, post db.Post) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotResponse postResponse
	err = json.Unmarshal(data, &gotResponse)
	require.NoError(t, err)

	require.Equal(t, post.ID.String(), gotResponse.ID)
	require.Equal(t, post.WallID.String(), gotResponse.WallID)
	require.Equal(t, post.Author.String(), gotResponse.Author)
	require.Equal(t, post.MediaUrl.String, gotResponse.MediaURL)
	require.Equal(t, string(post.PostType.PostType), gotResponse.PostType)
	require.Equal(t, post.IsHighlighted.Bool, gotResponse.IsHighlighted)
	require.Equal(t, post.LikesCount.Int32, gotResponse.LikesCount)
	require.Equal(t, post.IsDeleted.Bool, gotResponse.IsDeleted)
}

func requireBodyMatchPostsResponse(t *testing.T, body *bytes.Buffer, posts []db.Post) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotResponses []postResponse
	err = json.Unmarshal(data, &gotResponses)
	require.NoError(t, err)

	require.Equal(t, len(posts), len(gotResponses))

	postMap := make(map[string]db.Post)
	for _, post := range posts {
		postMap[post.ID.String()] = post
	}

	for _, resp := range gotResponses {
		originalPost, exists := postMap[resp.ID]
		require.True(t, exists)

		require.Equal(t, originalPost.WallID.String(), resp.WallID)
		require.Equal(t, originalPost.Author.String(), resp.Author)
		require.Equal(t, originalPost.MediaUrl.String, resp.MediaURL)
		require.Equal(t, string(originalPost.PostType.PostType), resp.PostType)
		require.Equal(t, originalPost.IsHighlighted.Bool, resp.IsHighlighted)
		require.Equal(t, originalPost.LikesCount.Int32, resp.LikesCount)
		require.Equal(t, originalPost.IsDeleted.Bool, resp.IsDeleted)
	}
}
