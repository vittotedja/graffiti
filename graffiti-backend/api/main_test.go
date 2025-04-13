package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/vittotedja/graffiti/graffiti-backend/db/mock"
	"github.com/vittotedja/graffiti/graffiti-backend/token"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

func newTestServer(t *testing.T) *Server {
    config := util.Config{
        TokenSymmetricKey: util.RandomString(32), // Random symmetric key for JWT
    }

    tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
    require.NoError(t, err)

    router:= gin.Default()

    server := &Server{
        config:     config,
        router:     router,
        tokenMaker: tokenMaker,
    }
    
    mockCtrl := gomock.NewController(t)
    t.Cleanup(mockCtrl.Finish) 
    server.hub = mockdb.NewMockHub(mockCtrl)

    // server.router.Use(logger.Middleware())
    server.registerRoutes("unit-test")

    return server
}

func TestMain(m *testing.M) {
    gin.SetMode(gin.TestMode)

    os.Exit(m.Run())
}