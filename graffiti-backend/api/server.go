package api

import (
	"log"
	"time"

	cron "github.com/vittotedja/graffiti/graffiti-backend/util/cron"

	"context"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vittotedja/graffiti/graffiti-backend/db/sqlc"
	"github.com/vittotedja/graffiti/graffiti-backend/token"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
	"github.com/vittotedja/graffiti/graffiti-backend/util/logger"
)

// Server serves HTTP requests
type Server struct {
	hub        *db.Hub
	db         *pgxpool.Pool
	config     util.Config
	router     *gin.Engine // helps us send each API request to the correct handler for processing
	tokenMaker token.Maker
	httpServer *http.Server
}

// NewServer creates a new HTTP server and sets up all routes
func NewServer(config util.Config) *Server {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		log.Fatal("cannot create token maker", err)
	}
	server := &Server{config: config, router: gin.Default(), tokenMaker: tokenMaker}
	server.router.Use(logger.Middleware())
	server.registerRoutes()

	return server
}

// Start runs the HTTP server on a specific address
// and returns an error if the server fails to start
func (s *Server) Start() error {
	// Init DB
	ctx := context.Background()

	connPool, err := pgxpool.New(ctx, s.config.DBSource)
	if err != nil {
		logger.Global().Error("Cannot connect to DB", err)
		return fmt.Errorf("cannot connect to db: %w", err)
	}
	s.db = connPool
	s.hub = db.NewHub(connPool)

	// Set up HTTP server
	s.httpServer = &http.Server{
		Addr:    s.config.ServerAddress,
		Handler: s.router,
	}

	cron.ScheduleMaterializedViewRefresh(s.db)

	logger.Global().Info("Server listening on %s", s.config.ServerAddress)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Global().Info("Attempting graceful shutdown...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Global().Error("HTTP server shutdown error", err)
		return fmt.Errorf("server shutdown failed: %v", err)
	}

	if s.db != nil {
		s.db.Close()
		logger.Global().Info("Database connection closed.")
	}

	logger.Global().Info("Server shut down successfully.")
	return nil
}

func (s *Server) registerRoutes() {

	frontendURL := s.config.FrontendURL
	// Apply CORS middleware
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL}, // frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ALB Health check endpoint
	s.router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Healthy")
	})

	s.router.POST("/api/v1/auth/register", s.Register)
	s.router.POST("/api/v1/auth/login", s.Login)
	s.router.POST("/api/v1/auth/logout", s.Logout)

	protected := s.router.Group("/api")
	protected.Use(s.AuthMiddleware())
	{
		// auth
		protected.POST("/v1/auth/me", s.Me)
		// users
		protected.GET("/v1/users/:id", s.getUser)
		protected.POST("/v2/users", s.updateUserNew) // no test

		// Protected Walls Endpoint
		protected.GET("/v1/walls/:id", s.getWall) // working
		protected.GET("/v2/walls", s.getOwnWall)
		protected.GET("/v1/users/:id/walls", s.listWallsByUser)
		protected.POST("/v2/walls", s.createNewWall)              //no test yet
		protected.PUT("/v1/walls/:id", s.updateWall)              // working
		protected.PUT("/v1/walls/:id/publicize", s.publicizeWall) // working
		protected.PUT("/v1/walls/:id/privatize", s.privatizeWall) // working
		protected.PUT("/v1/walls/:id/pin", s.pinWall)             // working
		protected.DELETE("/v1/walls/:id", s.deleteWall)           // working

		protected.GET("/v1/walls/archived", s.getArchivedWalls)
		protected.PUT("/v1/walls/:id/archive", s.archiveWall)
		protected.PUT("/v1/walls/:id/unarchive", s.unarchiveWall)

		// search
		protected.POST("/v1/users/search", s.searchUsers) //no test

		//uploads
		protected.POST("/v1/presign", s.presignHandler)

		//friends
		protected.POST("/v1/friend-requests", s.createFriendRequest) // working
		protected.POST("/v1/friendships", s.listFriendshipByUserPairs)
		protected.GET("/v1/friends", s.getFriendsByStatus)                 //status = friends, requested, sent
		protected.PUT("/v1/friend-requests/accept", s.acceptFriendRequest) // working
		protected.DELETE("/v1/friendships", s.deleteFriendship)

		//posts
		protected.GET("/v2/walls/:id/posts", s.listPostsByWallWithAuthorsDetails) // no test yet
		protected.DELETE("/v1/posts/:id", s.deletePost)                           // working
		protected.POST("/v1/posts", s.createPost)                                 // working

		//likes
		protected.POST("/v1/likes", s.updateLike) // add and delete likes in 1 endpoint in likes count
		protected.GET("/v1/likes/:post_id", s.getLike)

		//discover
		protected.POST("/v1/friends/discover", s.discoverFriendsByMutuals)
		protected.POST("/v1/friends/mutual", s.getMutualFriends)

	}

	s.router.GET("/api/v1/users", s.listUsers) // working
	s.router.DELETE("/api/v1/users/:id", s.deleteUser)
	s.router.PUT("/api/v1/users/:id/onboarding", s.finishOnboarding) // working

	s.router.GET("/api/v1/walls", s.listWalls) // working

	s.router.GET("/api/v1/posts/:id", s.getPost)                                     // working
	s.router.GET("/api/v1/posts", s.listPosts)                                       // working
	s.router.GET("/api/v1/walls/:id/posts", s.listPostsByWall)                       // working
	s.router.GET("/api/v1/posts/highlighted", s.getHighlightedPosts)                 // working
	s.router.GET("/api/v1/walls/:id/posts/highlighted", s.getHighlightedPostsByWall) // working
	s.router.PUT("/api/v1/posts/:id", s.updatePost)                                  // working
	s.router.PUT("/api/v1/posts/:id/highlight", s.highlightPost)                     // working
	s.router.PUT("/api/v1/posts/:id/unhighlight", s.unhighlightPost)                 // working

	// Friend Requests
	s.router.POST("/api/v1/friends/mutual/count", s.getNumberOfMutualFriends)

	// User Blocking
	s.router.PUT("/api/v1/users/block", s.blockUser)
	s.router.PUT("/api/v1/users/unblock", s.unblockUser)

	// Friends Retrieval
	s.router.GET("/api/v1/users/:id/accepted-friends", s.getFriends)                      // user_id; working
	s.router.GET("/api/v1/users/:id/friend-requests/pending", s.getPendingFriendRequests) // user_id; working
	s.router.GET("/api/v1/users/:id/friend-requests/sent", s.getSentFriendRequests)       // user_id; working

	// Existing friendship-related routes
	s.router.GET("/api/v1/users/:id/friendships", s.listFriendshipsByUserId)                            // working
	s.router.GET("/api/v1/users/:id/accepted-friends/count", s.getNumberOfFriends)                      // working
	s.router.GET("/api/v1/users/:id/friend-requests/pending/count", s.getNumberOfPendingFriendRequests) // working
	// router for listFriendshipByUserPairs

	// Set up routes for the like API
	s.router.GET("/api/v1/likes", s.listLikes)
	s.router.GET("/api/v1/posts/:id/likes", s.listLikesByPost)
	s.router.GET("/api/v1/users/:id/likes", s.listLikesByUser)
	s.router.DELETE("/api/v1/likes", s.deleteLike)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
