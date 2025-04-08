package api

import (
	"testing"
	"time"
	"github.com/golang/mock/gomock"
	mockdb "github.com/vittotedja/graffiti/graffiti-backend/db/mock"

	"github.com/google/uuid"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

func TestGetUserAPI(t *testing.T) {
	
	user := randomUser()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hub := mockdb.NewMockStore(ctrl)
}

func TestGetUserAPI(t *testing.T) {
	user, _ := randomUser(t)
  
	testCases := []struct {
	  name          string
	  accountID     int64
	  setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	  buildStubs    func(store *mockdb.MockStore)
	  checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
	  {
		name:      "OK",
		accountID: account.ID,
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		  addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
		},
		buildStubs: func(store *mockdb.MockStore) {
		  store.EXPECT().
			GetAccount(gomock.Any(), gomock.Eq(account.ID)).
			Times(1).
			Return(account, nil)
		},
		checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusOK, recorder.Code)
		  requireBodyMatchAccount(t, recorder.Body, account)
		},
	  },
	  {
		name:      "UnauthorizedUser",
		accountID: account.ID,
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		  addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", util.DepositorRole, time.Minute)
		},
		buildStubs: func(store *mockdb.MockStore) {
		  store.EXPECT().
			GetAccount(gomock.Any(), gomock.Eq(account.ID)).
			Times(1).
			Return(account, nil)
		},
		checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusUnauthorized, recorder.Code)
		},
	  },
	  {
		name:      "NoAuthorization",
		accountID: account.ID,
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		},
		buildStubs: func(store *mockdb.MockStore) {
		  store.EXPECT().
			GetAccount(gomock.Any(), gomock.Any()).
			Times(0)
		},
		checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusUnauthorized, recorder.Code)
		},
	  },
	  {
		name:      "NotFound",
		accountID: account.ID,
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		  addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
		},
  
		buildStubs: func(store *mockdb.MockStore) {
		  store.EXPECT().
			GetAccount(gomock.Any(), gomock.Eq(account.ID)).
			Times(1).
			Return(db.Account{}, db.ErrRecordNotFound)
		},
		checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusNotFound, recorder.Code)
		},
	  },
	  {
		name:      "InternalError",
		accountID: account.ID,
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		  addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
		},
		buildStubs: func(store *mockdb.MockStore) {
		  store.EXPECT().
			GetAccount(gomock.Any(), gomock.Eq(account.ID)).
			Times(1).
			Return(db.Account{}, sql.ErrConnDone)
		},
		checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusInternalServerError, recorder.Code)
		},
	  },
	  {
		name:      "InvalidID",
		accountID: 0,
		setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		  addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
		},
		buildStubs: func(store *mockdb.MockStore) {
		  store.EXPECT().
			GetAccount(gomock.Any(), gomock.Any()).
			Times(0)
		},
		checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		  require.Equal(t, http.StatusBadRequest, recorder.Code)
		},
	  },
	}
  

func randomUser() db.User {
	return db.User{
		ID:              uuid.New(),
		Username:        util.RandomUsername(),
		Fullname:        util.RandomFullname(),
		Email:           util.RandomEmail(),
		HashedPassword:  util.HashPassword(util.RandomString(10)),
		ProfilePicture:  util.RandomProfilePictureURL(),
		Bio:             util.RandomBio(),
		HasOnboarded:    false,
		BackgroundImage: util.RandomProfilePictureURL(),
		OnboardingAt:    nil,
		CreatedAt:       time.now(),
		UpdatedAt:       nil
	}
}