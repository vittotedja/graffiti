package api

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/vittotedja/graffiti/graffiti-backend/db/mock"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/token"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

// Helper function to add a token as a cookie
func addTokenAsCookie(
    t *testing.T,
    request *http.Request,
    tokenMaker token.Maker,
    username string,
    duration time.Duration,
) {
    token, _, err := tokenMaker.CreateToken(username, duration)
    require.NoError(t, err)

    cookie := &http.Cookie{
        Name:     "token",
        Value:    token,
        HttpOnly: true,
        Path:     "/",
    }
    request.AddCookie(cookie)
}

func TestAuthMiddleware(t *testing.T) {
    username := util.RandomUsername()

    testCases := []struct {
        name          string
        setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
        checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
    }{
        {
            name: "OK",
            setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
                addTokenAsCookie(t, request, tokenMaker, username, time.Minute)
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusOK, recorder.Code)
            },
        },
        {
            name: "NoToken",
            setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusUnauthorized, recorder.Code)
            },
        },
        {
            name: "InvalidToken",
            setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
                cookie := &http.Cookie{
                    Name:     "token",
                    Value:    "invalid-token",
                    HttpOnly: true,
                    Path:     "/",
                }
                request.AddCookie(cookie)
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusUnauthorized, recorder.Code)
            },
        },
        {
            name: "ExpiredToken",
            setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
                addTokenAsCookie(t, request, tokenMaker, username, -time.Minute)
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusUnauthorized, recorder.Code)
            },
        },
        {
            name: "UserNotFound",
            setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
                addTokenAsCookie(t, request, tokenMaker, "nonexistent-user", time.Minute)
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusInternalServerError, recorder.Code)
            },
        },
    }

    for i := range testCases {
        tc := testCases[i]
    
        t.Run(tc.name, func(t *testing.T) {
            server := newTestServer(t)
            
            mockHub := server.hub.(*mockdb.MockHub)
            
            if tc.name == "OK" {
                mockHub.EXPECT().
                    GetUserByUsername(gomock.Any(), username).
                    Return(db.User{Username: username}, nil).
                    AnyTimes()
            } else if tc.name == "UserNotFound" {
                mockHub.EXPECT().
                    GetUserByUsername(gomock.Any(), "nonexistent-user").
                    Return(db.User{}, sql.ErrNoRows).
                    AnyTimes()
            }
            
        })
    }
}