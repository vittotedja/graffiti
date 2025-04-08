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
	"github.com/stretchr/testify/require"
	mockdb "github.com/vittotedja/graffiti/graffiti-backend/db/mock"
	"github.com/vittotedja/graffiti/graffiti-backend/token"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

func TestGetUserAPI(t *testing.T) {
	user := randomUser(t)
  
	testCases := []struct {
	  name          string
	  userID     	pgtype.UUID
	  body			gin.H
	  setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	  buildStubs    func(store *mockdb.MockHub)
	  checkResponse func(recoder *httptest.ResponseRecorder)
	}{
	  {
		name:      "OK",
		userID: user.ID,
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		  addTokenAsCookie(t, request, tokenMaker, user.Username, time.Minute)
		},
		buildStubs: func(store *mockdb.MockHub) {
		  store.EXPECT().
			GetUser(gomock.Any(), gomock.Eq(user.ID)).
			Times(1).
			Return(user, nil)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusOK, recorder.Code)
		  requireBodyMatchUser(t, recorder.Body, user)
		},
	  },
	  {
		name:      "UnauthorizedUser",
		userID: user.ID,
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			addTokenAsCookie(t, request, tokenMaker, "unauthorizedUser", time.Minute)
		},
		buildStubs: func(store *mockdb.MockHub) {
		  store.EXPECT().
			GetUser(gomock.Any(), gomock.Eq(user.ID)).
			Times(1).
			Return(user, nil)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusUnauthorized, recorder.Code)
		},
	  },
	  {
		name:      "NoAuthorization",
		userID: user.ID,
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		},
		buildStubs: func(store *mockdb.MockHub) {
		  store.EXPECT().
			GetUser(gomock.Any(), gomock.Any()).
			Times(0)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusUnauthorized, recorder.Code)
		},
	  },
	  {
		name:      "NotFound",
		userID: user.ID,
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			addTokenAsCookie(t, request, tokenMaker, user.Username, time.Minute)
		},
  
		buildStubs: func(store *mockdb.MockHub) {
		  store.EXPECT().
			GetUser(gomock.Any(), gomock.Eq(user.ID)).
			Times(1).
			Return(db.User{}, db.ErrRecordNotFound)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusNotFound, recorder.Code)
		},
	  },
	  {
		name:      "InternalError",
		userID: user.ID,
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			addTokenAsCookie(t, request, tokenMaker, user.Username, time.Minute)
		},
		buildStubs: func(store *mockdb.MockHub) {
		  store.EXPECT().
			GetUser(gomock.Any(), gomock.Eq(user.ID)).
			Times(1).
			Return(db.User{}, sql.ErrConnDone)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusInternalServerError, recorder.Code)
		},
	  },
	  {
		name:      "InvalidID",
		userID: func() pgtype.UUID {
			id := pgtype.UUID{}
			id.Scan(uuid.MustParse("00000000-0000-0000-0000-000000000000")) // Zero UUID
			return id
		}(),
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			addTokenAsCookie(t, request, tokenMaker, user.Username, time.Minute)
		},
		buildStubs: func(store *mockdb.MockHub) {
		  store.EXPECT().
			GetUser(gomock.Any(), gomock.Any()).
			Times(0)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusBadRequest, recorder.Code)
		},
	  },
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users/login" //????
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
} 

func randomUser(t *testing.T) db.User {
	id := pgtype.UUID{}
	id.Scan(uuid.New())
	
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
	
	return db.User{
		ID:              id,
		Username:        util.RandomUsername(),
		Fullname:        fullname,
		Email:           util.RandomEmail(),
		HashedPassword:  "",
		ProfilePicture:  profilePicture,
		Bio:             bio,
		HasOnboarded:    hasOnboarded,
		BackgroundImage: backgroundImage,
		OnboardingAt:    pgtype.Timestamp{}, // Empty timestamp
		CreatedAt:       createdAt,
		UpdatedAt:       pgtype.Timestamp{}, // Empty timestamp
	}
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.Fullname, gotUser.Fullname)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.HashedPassword)
}