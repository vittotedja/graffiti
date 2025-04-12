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

// TestCreateWallAPI tests the createNewWall handler
func TestCreateWallAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)

	testCases := []struct {
		name          string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"title":            wall.Title,
				"description":      wall.Description.String,
				"background_image": wall.BackgroundImage.String,
				"is_public":        wall.IsPublic.Bool,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateTestWall(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.CreateTestWallParams) (db.Wall, error) {
						require.Equal(t, user.ID, params.UserID)
						require.Equal(t, wall.Title, params.Title)
						require.Equal(t, wall.Description.String, params.Description.String)
						require.Equal(t, wall.BackgroundImage.String, params.BackgroundImage.String)
						require.Equal(t, wall.IsPublic.Bool, params.IsPublic.Bool)
						return wall, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchWallResponse(t, recorder.Body, wall)
			},
		},
		{
			name: "BadRequest",
			body: gin.H{
				"description":      wall.Description.String,
				"background_image": wall.BackgroundImage.String,
				"is_public":        wall.IsPublic.Bool,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateTestWall(gomock.Any(), gomock.Any()).
					Times(0).MaxTimes(1).
					Return(db.Wall{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"title":            wall.Title,
				"description":      wall.Description.String,
				"background_image": wall.BackgroundImage.String,
				"is_public":        wall.IsPublic.Bool,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					CreateTestWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, sql.ErrConnDone)
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

			
			server.router.POST("/test/walls", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.createNewWall(ctx)
			})

			recorder := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			
			request, err := http.NewRequest(http.MethodPost, "/test/walls", bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetWallAPI tests the getWall handler
func TestGetWallAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)

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
					GetWall(gomock.Any(), gomock.Eq(id)).
					Times(1).
					Return(wall, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchWallResponse(t, recorder.Body, wall)
			},
		},
		{
			name:   "WallNotFound",
			wallID: uuid.New().String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, db.ErrRecordNotFound)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			wallID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
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
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, sql.ErrConnDone)
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

			
			server.router.GET("/test/walls/:id", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.getWall(ctx)
			})

			
			recorder := httptest.NewRecorder()

			
			url := fmt.Sprintf("/test/walls/%s", tc.wallID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListWallsAPI tests the listWalls handler
func TestListWallsAPI(t *testing.T) {
	user, _ := randomUser(t)
	n := 5
	walls := make([]db.Wall, n)
	for i := 0; i < n; i++ {
		walls[i] = randomWall(t, user.ID)
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
					ListWalls(gomock.Any()).
					Times(1).
					Return(walls, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchWallsResponse(t, recorder.Body, walls)
			},
		},
		{
			name: "InternalError",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListWalls(gomock.Any()).
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

			
			request, err := http.NewRequest(http.MethodGet, "/api/v1/walls", nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestUpdateWallAPI tests the updateWall handler
func TestUpdateWallAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)

	// Create updated wall
	updatedWall := wall
	newTitle := "Updated Title"
	newDescription := "Updated Description"
	updatedWall.Title = newTitle
	updatedWall.Description.String = newDescription

	testCases := []struct {
		name          string
		wallID        string
		body          gin.H
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			wallID: wall.ID.String(),
			body: gin.H{
				"title":       newTitle,
				"description": newDescription,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				// First expect GetWall to check ownership
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				// Then expect UpdateWall
				mockHub.EXPECT().
					UpdateWall(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.UpdateWallParams) (db.Wall, error) {
						require.Equal(t, wall.ID.String(), params.ID.String())
						require.Equal(t, newTitle, params.Title)
						require.Equal(t, newDescription, params.Description.String)
						return updatedWall, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchWallResponse(t, recorder.Body, updatedWall)
			},
		},
		{
			name:   "WallNotFound",
			wallID: uuid.New().String(),
			body: gin.H{
				"title":       newTitle,
				"description": newDescription,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, db.ErrRecordNotFound)

				// UpdateWall should not be called
				mockHub.EXPECT().
					UpdateWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "Unauthorized",
			wallID: wall.ID.String(),
			body: gin.H{
				"title":       newTitle,
				"description": newDescription,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				// Return a wall with different user ID
				differentUserWall := wall
				differentUserID := pgtype.UUID{}
				differentUserID.Scan(uuid.New().String())
				differentUserWall.UserID = differentUserID

				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(differentUserWall, nil)

				// UpdateWall should not be called
				mockHub.EXPECT().
					UpdateWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			wallID: "invalid-id",
			body: gin.H{
				"title":       newTitle,
				"description": newDescription,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					UpdateWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			wallID: wall.ID.String(),
			body: gin.H{
				"title":       newTitle,
				"description": newDescription,
			},
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				mockHub.EXPECT().
					UpdateWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, sql.ErrConnDone)
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

			
			server.router.PUT("/test/walls/:id", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.updateWall(ctx)
			})
			
			recorder := httptest.NewRecorder()
			
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/test/walls/%s", tc.wallID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestDeleteWallAPI tests the deleteWall handler
func TestDeleteWallAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)

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
				// First expect GetWall to check ownership
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				// Then expect DeleteWall
				mockHub.EXPECT().
					DeleteWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)

				// Mock ListPostsByWall for background goroutine
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
			name:   "WallNotFound",
			wallID: uuid.New().String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, db.ErrRecordNotFound)

				// DeleteWall should not be called
				mockHub.EXPECT().
					DeleteWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "Unauthorized",
			wallID: wall.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				differentUserWall := wall
				differentUserID := pgtype.UUID{}
				differentUserID.Scan(uuid.New().String())
				differentUserWall.UserID = differentUserID

				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(differentUserWall, nil)

				mockHub.EXPECT().
					DeleteWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			wallID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					DeleteWall(gomock.Any(), gomock.Any()).
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
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				mockHub.EXPECT().
					DeleteWall(gomock.Any(), gomock.Any()).
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

			server.router.DELETE("/test/walls/:id", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.deleteWall(ctx)
			})

			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/test/walls/%s", tc.wallID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestPublicizeWallAPI tests the publicizeWall handler
func TestPublicizeWallAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)

	// Create updated wall (publicized)
	publicizedWall := wall
	publicizedWall.IsPublic.Bool = true

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
				// First expect GetWall to check ownership
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				// Then expect PublicizeWall
				mockHub.EXPECT().
					PublicizeWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(publicizedWall, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchWallResponse(t, recorder.Body, publicizedWall)
			},
		},
		{
			name:   "WallNotFound",
			wallID: uuid.New().String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, db.ErrRecordNotFound)

				// PublicizeWall should not be called
				mockHub.EXPECT().
					PublicizeWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "Unauthorized",
			wallID: wall.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				// Return a wall with different user ID
				differentUserWall := wall
				differentUserID := pgtype.UUID{}
				differentUserID.Scan(uuid.New().String())
				differentUserWall.UserID = differentUserID

				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(differentUserWall, nil)

				// PublicizeWall should not be called
				mockHub.EXPECT().
					PublicizeWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			wallID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					PublicizeWall(gomock.Any(), gomock.Any()).
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
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				mockHub.EXPECT().
					PublicizeWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, sql.ErrConnDone)
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

			server.router.PUT("/test/walls/:id/publicize", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.publicizeWall(ctx)
			})

			
			recorder := httptest.NewRecorder()

			
			url := fmt.Sprintf("/test/walls/%s/publicize", tc.wallID)
			request, err := http.NewRequest(http.MethodPut, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestPrivatizeWallAPI tests the privatizeWall handler
func TestPrivatizeWallAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)
	wall.IsPublic.Bool = true // Make sure it's public initially

	// Create updated wall (privatized)
	privatizedWall := wall
	privatizedWall.IsPublic.Bool = false

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
				// First expect GetWall to check ownership
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				// Then expect PrivatizeWall
				mockHub.EXPECT().
					PrivatizeWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(privatizedWall, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchWallResponse(t, recorder.Body, privatizedWall)
			},
		},
		{
			name:   "WallNotFound",
			wallID: uuid.New().String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, db.ErrRecordNotFound)

				// PrivatizeWall should not be called
				mockHub.EXPECT().
					PrivatizeWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "Unauthorized",
			wallID: wall.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				// Return a wall with different user ID
				differentUserWall := wall
				differentUserID := pgtype.UUID{}
				differentUserID.Scan(uuid.New().String())
				differentUserWall.UserID = differentUserID

				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(differentUserWall, nil)

				// PrivatizeWall should not be called
				mockHub.EXPECT().
					PrivatizeWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			wallID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					PrivatizeWall(gomock.Any(), gomock.Any()).
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
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				mockHub.EXPECT().
					PrivatizeWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, sql.ErrConnDone)
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

			
			server.router.PUT("/test/walls/:id/privatize", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.privatizeWall(ctx)
			})

			
			recorder := httptest.NewRecorder()

			
			url := fmt.Sprintf("/test/walls/%s/privatize", tc.wallID)
			request, err := http.NewRequest(http.MethodPut, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestPinWallAPI tests the pinWall handler
func TestPinWallAPI(t *testing.T) {
	user, _ := randomUser(t)
	wall := randomWall(t, user.ID)
	wall.IsPinned.Bool = false // Make sure it's not pinned initially

	// Create updated wall (pinned)
	pinnedWall := wall
	pinnedWall.IsPinned.Bool = true

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
				// First expect GetWall to check ownership
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				// Then expect PinUnpinWall
				mockHub.EXPECT().
					PinUnpinWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(pinnedWall, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchWallResponse(t, recorder.Body, pinnedWall)
			},
		},
		{
			name:   "WallNotFound",
			wallID: uuid.New().String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, db.ErrRecordNotFound)

				// PinUnpinWall should not be called
				mockHub.EXPECT().
					PinUnpinWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "Unauthorized",
			wallID: wall.ID.String(),
			setupMock: func(mockHub *mockdb.MockHub) {
				// Return a wall with different user ID
				differentUserWall := wall
				differentUserID := pgtype.UUID{}
				differentUserID.Scan(uuid.New().String())
				differentUserWall.UserID = differentUserID

				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(differentUserWall, nil)

				// PinUnpinWall should not be called
				mockHub.EXPECT().
					PinUnpinWall(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			wallID: "invalid-id",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					GetWall(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					PinUnpinWall(gomock.Any(), gomock.Any()).
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
					GetWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(wall, nil)

				mockHub.EXPECT().
					PinUnpinWall(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Wall{}, sql.ErrConnDone)
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

			
			server.router.PUT("/test/walls/:id/pin", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.pinWall(ctx)
			})

			
			recorder := httptest.NewRecorder()

			
			url := fmt.Sprintf("/test/walls/%s/pin", tc.wallID)
			request, err := http.NewRequest(http.MethodPut, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestGetOwnWallAPI tests the getOwnWall handler
func TestGetOwnWallAPI(t *testing.T) {
	user, _ := randomUser(t)
	
	n := 5
	walls := make([]db.Wall, n)
	for i := 0; i < n; i++ {
		walls[i] = randomWall(t, user.ID)
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
					ListWallsByUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(walls, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchWallsResponse(t, recorder.Body, walls)
			},
		},
		{
			name: "InternalError",
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListWallsByUser(gomock.Any(), gomock.Any()).
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

			
			server.router.GET("/test/walls", func(ctx *gin.Context) {
				ctx.Set("currentUser", user)
				server.getOwnWall(ctx)
			})

			
			recorder := httptest.NewRecorder()

			
			request, err := http.NewRequest(http.MethodGet, "/test/walls", nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// TestListWallsByUserAPI tests the listWallsByUser handler
func TestListWallsByUserAPI(t *testing.T) {
	currentUser, _ := randomUser(t)
	otherUser, _ := randomUser(t)
	
	n := 5
	walls := make([]db.Wall, n)
	for i := 0; i < n; i++ {
		walls[i] = randomWall(t, otherUser.ID)
		if i % 2 == 0 {
			walls[i].IsPublic.Bool = false
		} else {
			walls[i].IsPublic.Bool = true
		}
	}

	// Create a friendship status
	friendshipStatus := db.Friendship{
		ID: pgtype.UUID{},
		FromUser: currentUser.ID,
		ToUser: otherUser.ID,
		Status: db.NullStatus{Status: "friends", Valid: true},
		CreatedAt: pgtype.Timestamp{},
		UpdatedAt: pgtype.Timestamp{},
	}

	testCases := []struct {
		name          string
		userID        string
		isFriend      bool
		setupMock     func(mockHub *mockdb.MockHub)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "Friend_ShowsAllWalls",
			userID:   otherUser.ID.String(),
			isFriend: true,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListWallsByUser(gomock.Any(), gomock.Eq(otherUser.ID)).
					Times(1).
					Return(walls, nil)

				mockHub.EXPECT().
					ListFriendshipByUserPairs(gomock.Any(), gomock.Any()).
					Times(1).
					Return(friendshipStatus, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)				
				requireBodyMatchWallsResponse(t, recorder.Body, walls)
			},
		},
		{
			name:     "NotFriend_ShowsOnlyPublicWalls",
			userID:   otherUser.ID.String(),
			isFriend: false,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListWallsByUser(gomock.Any(), gomock.Eq(otherUser.ID)).
					Times(1).
					Return(walls, nil)

				// Not friends
				mockHub.EXPECT().
					ListFriendshipByUserPairs(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Friendship{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			
				var publicWalls []db.Wall
				for _, wall := range walls {
					if wall.IsPublic.Bool {
						publicWalls = append(publicWalls, wall)
					}
				}
				requireBodyMatchWallsResponse(t, recorder.Body, publicWalls)
			},
		},
		{
			name:     "InvalidID",
			userID:   "invalid-id",
			isFriend: false,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListWallsByUser(gomock.Any(), gomock.Any()).
					Times(0)
				mockHub.EXPECT().
					ListFriendshipByUserPairs(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "InternalError",
			userID:   otherUser.ID.String(),
			isFriend: true,
			setupMock: func(mockHub *mockdb.MockHub) {
				mockHub.EXPECT().
					ListWallsByUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)

				mockHub.EXPECT().
					ListFriendshipByUserPairs(gomock.Any(), gomock.Any()).
					Times(0)
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

			
			server.router.GET("/test/users/:id/walls", func(ctx *gin.Context) {
				ctx.Set("currentUser", currentUser)
				server.listWallsByUser(ctx)
			})

			
			recorder := httptest.NewRecorder()

			
			url := fmt.Sprintf("/test/users/%s/walls", tc.userID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

// Helper function to create a random wall
func randomWall(t *testing.T, userID pgtype.UUID) db.Wall {
	rawUUID := uuid.New()
	uuidString := rawUUID.String()

    id := pgtype.UUID{}
	err := id.Scan(uuidString)
	require.NoError(t, err)

	description := pgtype.Text{}
	description.Scan(util.RandomString(20))

	backgroundImage := pgtype.Text{}
	backgroundImage.Scan(util.RandomProfilePictureURL())

	isPublic := pgtype.Bool{}
	isPublic.Scan(true)

	isArchived := pgtype.Bool{}
	isArchived.Scan(false)

	isDeleted := pgtype.Bool{}
	isDeleted.Scan(false)

	isPinned := pgtype.Bool{}
	isPinned.Scan(false)

	popularityScore := pgtype.Float8{}
	popularityScore.Scan(0.0)

	createdAt := pgtype.Timestamp{}
	createdAt.Scan(time.Now())

	updatedAt := pgtype.Timestamp{}
	updatedAt.Scan(time.Now())

	return db.Wall{
		ID:              id,
		UserID:          userID,
		Title:           util.RandomString(10),
		Description:     description,
		BackgroundImage: backgroundImage,
		IsPublic:        isPublic,
		IsArchived:      isArchived,
		IsDeleted:       isDeleted,
		IsPinned:        isPinned,
		PopularityScore: popularityScore,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}

// Helper function to match wall response
func requireBodyMatchWallResponse(t *testing.T, body *bytes.Buffer, wall db.Wall) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotResponse wallResponse
	err = json.Unmarshal(data, &gotResponse)
	require.NoError(t, err)

	require.Equal(t, wall.ID.String(), gotResponse.ID)
	require.Equal(t, wall.UserID.String(), gotResponse.UserID)
	require.Equal(t, wall.Title, gotResponse.Title)
	require.Equal(t, wall.Description.String, gotResponse.Description)
	require.Equal(t, wall.BackgroundImage.String, gotResponse.BackgroundImage)
	require.Equal(t, wall.IsPublic.Bool, gotResponse.IsPublic)
	require.Equal(t, wall.IsArchived.Bool, gotResponse.IsArchived)
	require.Equal(t, wall.IsDeleted.Bool, gotResponse.IsDeleted)
	require.Equal(t, wall.IsPinned.Bool, gotResponse.IsPinned)
	require.Equal(t, wall.PopularityScore.Float64, gotResponse.PopularityScore)
}

// Helper function to match multiple wall responses
func requireBodyMatchWallsResponse(t *testing.T, body *bytes.Buffer, walls []db.Wall) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotResponses []wallResponse
	err = json.Unmarshal(data, &gotResponses)
	require.NoError(t, err)

	require.Equal(t, len(walls), len(gotResponses))

	wallMap := make(map[string]db.Wall)
	for _, wall := range walls {
		wallMap[wall.ID.String()] = wall
	}

	for _, resp := range gotResponses {
		originalWall, exists := wallMap[resp.ID]
		require.True(t, exists)
		
		require.Equal(t, originalWall.UserID.String(), resp.UserID)
		require.Equal(t, originalWall.Title, resp.Title)
		require.Equal(t, originalWall.Description.String, resp.Description)
		require.Equal(t, originalWall.BackgroundImage.String, resp.BackgroundImage)
		require.Equal(t, originalWall.IsPublic.Bool, resp.IsPublic)
		require.Equal(t, originalWall.IsArchived.Bool, resp.IsArchived)
		require.Equal(t, originalWall.IsDeleted.Bool, resp.IsDeleted)
		require.Equal(t, originalWall.IsPinned.Bool, resp.IsPinned)
		require.Equal(t, originalWall.PopularityScore.Float64, resp.PopularityScore)
	}
}