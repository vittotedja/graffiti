package api

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/require"
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
        Name:     "token", // The cookie name expected by the middleware
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
                // Do not add any token
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusUnauthorized, recorder.Code)
            },
        },
        {
            name: "InvalidToken",
            setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
                // Add an invalid token
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
                // Add an expired token
                addTokenAsCookie(t, request, tokenMaker, username, -time.Minute)
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusUnauthorized, recorder.Code)
            },
        },
        {
            name: "UserNotFound",
            setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
                // Add a valid token but simulate user not found
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
            server := newTestServer(t, nil)
            authPath := "/auth"
            server.router.GET(
                authPath,
                server.AuthMiddleware(),
                func(ctx *gin.Context) {
                    ctx.JSON(http.StatusOK, gin.H{})
                },
            )

            recorder := httptest.NewRecorder()
            request, err := http.NewRequest(http.MethodGet, authPath, nil)
            require.NoError(t, err)

            tc.setupAuth(t, request, server.tokenMaker)
            server.router.ServeHTTP(recorder, request)
            tc.checkResponse(t, recorder)
        })
    }
}